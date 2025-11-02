#!/bin/bash
# Docker-based cross-platform build script for GPU Pro

# Don't exit on error - we want to collect all build results
set +e

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

# Detect host architecture to determine correct compiler
HOST_ARCH=$(uname -m)

# Build configurations
# Format: "GOOS GOARCH EXTENSION CC CXX CGO_ENABLED BUILD_TAG"
if [ "$HOST_ARCH" = "aarch64" ] || [ "$HOST_ARCH" = "arm64" ]; then
    # Running on ARM64 (e.g., Apple Silicon in Docker)
    BUILDS=(
        # Linux builds (CGO enabled for NVML support)
        "linux amd64 NOEXT x86_64-linux-gnu-gcc x86_64-linux-gnu-g++ 1 linux"
        "linux arm64 NOEXT aarch64-linux-gnu-gcc aarch64-linux-gnu-g++ 1 linux"

        # macOS builds (CGO disabled - no GPU support)
        "darwin amd64 NOEXT NOCC NOCXX 0 darwin"
        "darwin arm64 NOEXT NOCC NOCXX 0 darwin"
    )
else
    # Running on AMD64
    BUILDS=(
        # Linux builds (CGO enabled for NVML support)
        "linux amd64 NOEXT gcc g++ 1 linux"
        "linux arm64 NOEXT aarch64-linux-gnu-gcc aarch64-linux-gnu-g++ 1 linux"

        # macOS builds (CGO disabled - no GPU support)
        "darwin amd64 NOEXT NOCC NOCXX 0 darwin"
        "darwin arm64 NOEXT NOCC NOCXX 0 darwin"
    )
fi

# Function to build for a specific platform and binary type
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=$3
    local CC=$4
    local CXX=$5
    local CGO=$6
    local BUILD_TAG=$7
    local BUILD_TYPE=$8  # "web" or "cli"

    local PLATFORM="${GOOS}-${GOARCH}"

    # Determine binary name and source path
    local BINARY_NAME="gpu-pro"
    local SOURCE_PATH="."

    if [ "$BUILD_TYPE" = "cli" ]; then
        BINARY_NAME="gpu-pro-cli"
        SOURCE_PATH="./cmd/gpu-pro-cli"
    fi

    # Handle extension
    local ACTUAL_EXT=""
    if [ "$EXT" != "NOEXT" ]; then
        ACTUAL_EXT="$EXT"
    fi

    local OUTPUT="dist/${BINARY_NAME}-${PLATFORM}${ACTUAL_EXT}"

    local TYPE_LABEL=""
    if [ "$BUILD_TYPE" = "cli" ]; then
        TYPE_LABEL=" [CLI]"
    else
        TYPE_LABEL=" [Web UI]"
    fi

    echo -e "${YELLOW}Building:${NC} ${PLATFORM}${TYPE_LABEL}"

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
    LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
    LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_TIME"
    LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"

    # Build
    if go build -tags="$BUILD_TAG" -ldflags="$LDFLAGS" -o "$OUTPUT" "$SOURCE_PATH" 2>&1 > /tmp/build-${BUILD_TYPE}-${PLATFORM}.log; then
        # Get file size
        SIZE=$(du -h "$OUTPUT" | cut -f1)
        echo -e "${GREEN}✓${NC} Built: ${OUTPUT} (${SIZE})"
        return 0
    else
        echo -e "${RED}✗ Failed to build ${PLATFORM}${TYPE_LABEL}${NC}"
        echo -e "${RED}Build log:${NC}"
        tail -20 /tmp/build-${BUILD_TYPE}-${PLATFORM}.log
        return 1
    fi
}

# Build all platforms
SUCCESS=0
FAILED=0

for BUILD_CONFIG in "${BUILDS[@]}"; do
    read -r GOOS GOARCH EXT CC CXX CGO BUILD_TAG <<< "$BUILD_CONFIG"

    # Build both web UI and CLI for each platform
    for BUILD_TYPE in "web" "cli"; do
        echo -e "${BLUE}[DEBUG] Starting build: $GOOS/$GOARCH $BUILD_TYPE${NC}" >&2
        if build_platform "$GOOS" "$GOARCH" "$EXT" "$CC" "$CXX" "$CGO" "$BUILD_TAG" "$BUILD_TYPE"; then
            ((SUCCESS++))
            echo -e "${GREEN}[DEBUG] Success count: $SUCCESS${NC}" >&2
        else
            ((FAILED++))
            echo -e "${RED}[DEBUG] Failed count: $FAILED${NC}" >&2
        fi
        echo
    done
done

# Skip archive creation - binaries are ready to use
# Users can create their own archives if needed

echo
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "Success: ${GREEN}${SUCCESS}${NC} | Failed: ${RED}${FAILED}${NC}"
echo
echo -e "${BLUE}Output directory:${NC} dist/"
ls -lh dist/
echo

# Exit with error if any builds failed
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some builds failed. Please check the logs above.${NC}"
    exit 1
fi

exit 0
