#!/bin/bash
# Enhanced build script for GPU Pro with multiple flavors
# Supports: Linux, macOS, Windows | amd64, arm64 | debug, release, minimal

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Build info
VERSION=${VERSION:-"2.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Output directory
OUTPUT_DIR="dist"
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   GPU Pro - Multi-Platform Build Script${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Version:${NC} $VERSION"
echo -e "${BLUE}Build Time:${NC} $BUILD_TIME"
echo -e "${BLUE}Git Commit:${NC} $GIT_COMMIT"
echo ""

# Function to build for a specific platform
build_binary() {
    local OS=$1
    local ARCH=$2
    local FLAVOR=$3
    local BINARY_NAME="gpu-pro"

    # Windows binaries need .exe extension
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="gpu-pro.exe"
    fi

    # Output path
    local OUTPUT_PATH="$OUTPUT_DIR/gpu-pro-${OS}-${ARCH}-${FLAVOR}"
    if [ "$OS" = "windows" ]; then
        OUTPUT_PATH="${OUTPUT_PATH}.exe"
    fi

    # Build flags based on flavor
    local LDFLAGS=""
    local BUILD_TAGS=""
    local CGO_ENABLED=1

    case "$FLAVOR" in
        release)
            LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
            ;;
        debug)
            LDFLAGS="-X main.Version=$VERSION-debug -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
            ;;
        minimal)
            LDFLAGS="-s -w -X main.Version=$VERSION-minimal -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
            CGO_ENABLED=0  # Static build, no CGO
            ;;
    esac

    echo -e "${YELLOW}Building:${NC} $OS/$ARCH ($FLAVOR)"

    # Build
    GOOS=$OS GOARCH=$ARCH CGO_ENABLED=$CGO_ENABLED go build \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_PATH" \
        2>&1 | grep -v "^#" || true

    if [ -f "$OUTPUT_PATH" ]; then
        local SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        echo -e "${GREEN}✓ Built:${NC} $OUTPUT_PATH (${SIZE})"

        # Compress release builds if UPX is available
        if [ "$FLAVOR" = "release" ] && command -v upx &> /dev/null; then
            echo -e "${BLUE}  Compressing with UPX...${NC}"
            upx --best --lzma "$OUTPUT_PATH" 2>&1 | grep -v "^Packing" || true
            SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
            echo -e "${GREEN}  Compressed:${NC} ${SIZE}"
        fi
    else
        echo -e "${RED}✗ Failed to build $OS/$ARCH${NC}"
    fi
    echo ""
}

# Install dependencies
echo -e "${BLUE}Installing dependencies...${NC}"
go mod download
go mod tidy
echo ""

# Parse command line arguments
BUILD_MODE="all"  # all, single, quick
PLATFORMS=""
FLAVORS=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        --platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        --flavors)
            FLAVORS="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --mode MODE         Build mode: all, quick, or single (default: all)"
            echo "  --platforms LIST    Comma-separated list of platforms (e.g., linux,darwin,windows)"
            echo "  --flavors LIST      Comma-separated list of flavors (e.g., release,debug,minimal)"
            echo "  --help              Show this help message"
            echo ""
            echo "Build Modes:"
            echo "  all      Build for all platforms, architectures, and flavors"
            echo "  quick    Build release binaries for all platforms/architectures"
            echo "  single   Build for current platform only (all flavors)"
            echo ""
            echo "Flavors:"
            echo "  release  Optimized production build (stripped, compressed)"
            echo "  debug    Debug build with symbols"
            echo "  minimal  Minimal static build without CGO"
            echo ""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Determine build targets based on mode
case "$BUILD_MODE" in
    all)
        echo -e "${YELLOW}Build Mode:${NC} All platforms, all flavors"
        PLATFORMS="linux,darwin,windows"
        FLAVORS="release,debug,minimal"
        ;;
    quick)
        echo -e "${YELLOW}Build Mode:${NC} Quick release builds"
        PLATFORMS="linux,darwin,windows"
        FLAVORS="release"
        ;;
    single)
        echo -e "${YELLOW}Build Mode:${NC} Current platform only"
        CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        PLATFORMS="$CURRENT_OS"
        FLAVORS="release,debug,minimal"
        ;;
esac

# Convert comma-separated lists to arrays
IFS=',' read -ra PLATFORM_ARRAY <<< "$PLATFORMS"
IFS=',' read -ra FLAVOR_ARRAY <<< "$FLAVORS"

echo ""
echo -e "${BLUE}Building for:${NC}"
echo -e "  Platforms: ${PLATFORM_ARRAY[*]}"
echo -e "  Flavors: ${FLAVOR_ARRAY[*]}"
echo ""

# Build matrix
TOTAL_BUILDS=0
SUCCESSFUL_BUILDS=0

for OS in "${PLATFORM_ARRAY[@]}"; do
    for FLAVOR in "${FLAVOR_ARRAY[@]}"; do
        # Build for amd64
        ((TOTAL_BUILDS++))
        build_binary "$OS" "amd64" "$FLAVOR" && ((SUCCESSFUL_BUILDS++)) || true

        # Build for arm64 (except Windows minimal builds which may have issues)
        if [ "$OS" != "windows" ] || [ "$FLAVOR" != "minimal" ]; then
            ((TOTAL_BUILDS++))
            build_binary "$OS" "arm64" "$FLAVOR" && ((SUCCESSFUL_BUILDS++)) || true
        fi
    done
done

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Build Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "Total builds: $TOTAL_BUILDS"
echo -e "Successful: ${GREEN}$SUCCESSFUL_BUILDS${NC}"
if [ $SUCCESSFUL_BUILDS -lt $TOTAL_BUILDS ]; then
    echo -e "Failed: ${RED}$((TOTAL_BUILDS - SUCCESSFUL_BUILDS))${NC}"
fi
echo ""

# List all generated binaries
echo -e "${BLUE}Generated binaries in ${OUTPUT_DIR}/:${NC}"
ls -lh "$OUTPUT_DIR"/ | grep gpu-pro | awk '{printf "  %s %s %s\n", $9, "("$5")"}'
echo ""

echo -e "${GREEN}Done!${NC}"
echo ""
echo -e "${BLUE}Usage examples:${NC}"
echo -e "  Linux:   ./$OUTPUT_DIR/gpu-pro-linux-amd64-release"
echo -e "  macOS:   ./$OUTPUT_DIR/gpu-pro-darwin-amd64-release"
echo -e "  Windows: ./$OUTPUT_DIR/gpu-pro-windows-amd64-release.exe"
