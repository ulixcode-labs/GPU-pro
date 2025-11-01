#!/bin/bash
# Docker-based cross-platform build script for GPU Pro

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

VERSION=${VERSION:-"2.0.0"}
BUILD_TIME=$(date -u +"%Y-%m-%d_%H:%M:%S")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   GPU Pro - Docker Cross-Platform Builder${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Version:${NC} $VERSION"
echo -e "${BLUE}Build Time:${NC} $BUILD_TIME"
echo -e "${BLUE}Git Commit:${NC} $GIT_COMMIT"
echo

# Create output directory
mkdir -p dist

# Build configurations
# Format: "GOOS GOARCH EXTENSION CC CXX CGO_ENABLED"
BUILDS=(
    # Linux builds (CGO enabled for NVML support)
    "linux amd64 NOEXT x86_64-linux-gnu-gcc x86_64-linux-gnu-g++ 1"
    "linux arm64 NOEXT aarch64-linux-gnu-gcc aarch64-linux-gnu-g++ 1"

    # Windows builds (CGO enabled)
    "windows amd64 .exe x86_64-w64-mingw32-gcc x86_64-w64-mingw32-g++ 1"
    "windows arm64 .exe aarch64-w64-mingw32-gcc aarch64-w64-mingw32-g++ 0"

    # macOS builds (CGO disabled - simpler, works across platforms)
    "darwin amd64 NOEXT NOCC NOCXX 0"
    "darwin arm64 NOEXT NOCC NOCXX 0"
)

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=$3
    local CC=$4
    local CXX=$5
    local CGO=$6

    local PLATFORM="${GOOS}-${GOARCH}"

    # Handle extension
    local ACTUAL_EXT=""
    if [ "$EXT" != "NOEXT" ]; then
        ACTUAL_EXT="$EXT"
    fi

    local OUTPUT="dist/gpu-pro-${PLATFORM}${ACTUAL_EXT}"

    echo -e "${YELLOW}Building:${NC} ${PLATFORM}"

    # Set environment
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=$CGO

    if [ "$CC" != "NOCC" ] && [ -n "$CC" ]; then
        export CC=$CC
    fi

    if [ "$CXX" != "NOCXX" ] && [ -n "$CXX" ]; then
        export CXX=$CXX
    fi

    # Build flags
    LDFLAGS="-w -s"
    LDFLAGS="$LDFLAGS -X main.version=$VERSION"
    LDFLAGS="$LDFLAGS -X main.buildTime=$BUILD_TIME"
    LDFLAGS="$LDFLAGS -X main.gitCommit=$GIT_COMMIT"

    # Build
    if go build -ldflags="$LDFLAGS" -o "$OUTPUT" main.go 2>&1 > /tmp/build.log; then
        # Get file size
        SIZE=$(du -h "$OUTPUT" | cut -f1)
        echo -e "${GREEN}✓${NC} Built: ${OUTPUT} (${SIZE})"
        return 0
    else
        echo -e "${RED}✗ Failed to build ${PLATFORM}${NC}"
        tail -10 /tmp/build.log
        return 1
    fi
}

# Build all platforms
SUCCESS=0
FAILED=0

for BUILD_CONFIG in "${BUILDS[@]}"; do
    read -r GOOS GOARCH EXT CC CXX CGO <<< "$BUILD_CONFIG"

    if build_platform "$GOOS" "$GOARCH" "$EXT" "$CC" "$CXX" "$CGO"; then
        ((SUCCESS++))
    else
        ((FAILED++))
    fi

    echo
done

# Create archives
echo -e "${BLUE}Creating release archives...${NC}"

cd dist
for file in gpu-pro-*; do
    if [ -f "$file" ]; then
        # Determine archive type
        if [[ $file == *.exe ]]; then
            # Windows: zip
            zip "${file%.exe}.zip" "$file" > /dev/null 2>&1
            echo -e "${GREEN}✓${NC} Created: ${file%.exe}.zip"
        else
            # Linux/macOS: tar.gz
            tar -czf "${file}.tar.gz" "$file" > /dev/null 2>&1
            echo -e "${GREEN}✓${NC} Created: ${file}.tar.gz"
        fi
    fi
done
cd ..

echo
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "Success: ${GREEN}${SUCCESS}${NC} | Failed: ${RED}${FAILED}${NC}"
echo
echo -e "${BLUE}Output directory:${NC} dist/"
ls -lh dist/
echo
