#!/bin/bash
# Comprehensive prebuild script for GPU Pro
# Builds binaries for Windows, Linux, and macOS (amd64 and arm64)

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

echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   GPU Pro - Cross-Platform Prebuild Script${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Version:${NC} $VERSION"
echo -e "${BLUE}Build Time:${NC} $BUILD_TIME"
echo -e "${BLUE}Git Commit:${NC} $GIT_COMMIT"
echo ""

# Install dependencies
echo -e "${BLUE}Installing dependencies...${NC}"
go mod download
go mod tidy
echo ""

# Function to build for a specific platform
build_binary() {
    local OS=$1
    local ARCH=$2
    local BINARY_NAME="gpu-pro"

    # Windows binaries need .exe extension
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="gpu-pro.exe"
    fi

    # Output path
    local OUTPUT_PATH="$OUTPUT_DIR/gpu-pro-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        OUTPUT_PATH="${OUTPUT_PATH}.exe"
    fi

    # Build flags - always release mode
    local LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

    # Use CGO for better performance and GPU support when possible
    # Disable CGO for cross-compilation to avoid C toolchain issues
    local CGO_ENABLED=0

    # Enable CGO only for native builds
    if [ "$OS" = "$(uname -s | tr '[:upper:]' '[:lower:]')" ]; then
        local NATIVE_ARCH=$(uname -m)
        case "$NATIVE_ARCH" in
            x86_64) NATIVE_ARCH="amd64" ;;
            aarch64|arm64) NATIVE_ARCH="arm64" ;;
        esac

        if [ "$ARCH" = "$NATIVE_ARCH" ]; then
            CGO_ENABLED=1
        fi
    fi

    echo -e "${YELLOW}Building:${NC} $OS/$ARCH (CGO_ENABLED=$CGO_ENABLED)"

    # Build
    GOOS=$OS GOARCH=$ARCH CGO_ENABLED=$CGO_ENABLED go build \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_PATH" \
        2>&1 | grep -v "^#" || true

    if [ -f "$OUTPUT_PATH" ]; then
        local SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        echo -e "${GREEN}✓ Built:${NC} $OUTPUT_PATH (${SIZE})"

        # Compress release builds if UPX is available
        if command -v upx &> /dev/null; then
            echo -e "${BLUE}  Compressing with UPX...${NC}"
            upx --best --lzma "$OUTPUT_PATH" 2>&1 | grep -v "^Packing" || true
            SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
            echo -e "${GREEN}  Compressed:${NC} ${SIZE}"
        fi
    else
        echo -e "${RED}✗ Failed to build $OS/$ARCH${NC}"
        return 1
    fi
    echo ""
}

# Build matrix
echo -e "${BLUE}Building binaries for all platforms...${NC}"
echo ""

# Track build stats
TOTAL_BUILDS=0
SUCCESSFUL_BUILDS=0

# Linux builds
for ARCH in amd64 arm64; do
    ((TOTAL_BUILDS++))
    build_binary "linux" "$ARCH" && ((SUCCESSFUL_BUILDS++)) || true
done

# macOS builds (Darwin)
for ARCH in amd64 arm64; do
    ((TOTAL_BUILDS++))
    build_binary "darwin" "$ARCH" && ((SUCCESSFUL_BUILDS++)) || true
done

# Windows builds
for ARCH in amd64 arm64; do
    ((TOTAL_BUILDS++))
    build_binary "windows" "$ARCH" && ((SUCCESSFUL_BUILDS++)) || true
done

# Summary
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Build Summary${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "Total builds attempted: $TOTAL_BUILDS"
echo -e "Successful: ${GREEN}$SUCCESSFUL_BUILDS${NC}"
if [ $SUCCESSFUL_BUILDS -lt $TOTAL_BUILDS ]; then
    echo -e "Failed: ${RED}$((TOTAL_BUILDS - SUCCESSFUL_BUILDS))${NC}"
fi
echo ""

# List all generated binaries
echo -e "${BLUE}Generated binaries in ${OUTPUT_DIR}/:${NC}"
if [ -d "$OUTPUT_DIR" ]; then
    ls -lh "$OUTPUT_DIR"/ 2>/dev/null | grep gpu-pro | awk '{printf "  %-40s %10s\n", $9, $5}' || echo "  (none)"
fi
echo ""

# Create checksums
echo -e "${BLUE}Generating checksums...${NC}"
cd "$OUTPUT_DIR"
sha256sum gpu-pro-* > SHA256SUMS 2>/dev/null || shasum -a 256 gpu-pro-* > SHA256SUMS
cd ..
echo -e "${GREEN}✓ Checksums saved to ${OUTPUT_DIR}/SHA256SUMS${NC}"
echo ""

echo -e "${GREEN}Done!${NC}"
echo ""
echo -e "${BLUE}Usage examples:${NC}"
echo -e "  Linux (x64):     ./$OUTPUT_DIR/gpu-pro-linux-amd64"
echo -e "  Linux (ARM64):   ./$OUTPUT_DIR/gpu-pro-linux-arm64"
echo -e "  macOS (Intel):   ./$OUTPUT_DIR/gpu-pro-darwin-amd64"
echo -e "  macOS (Apple M): ./$OUTPUT_DIR/gpu-pro-darwin-arm64"
echo -e "  Windows (x64):   .\\$OUTPUT_DIR\\gpu-pro-windows-amd64.exe"
echo -e "  Windows (ARM64): .\\$OUTPUT_DIR\\gpu-pro-windows-arm64.exe"
echo ""
