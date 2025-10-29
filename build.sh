#!/bin/bash
# Build script for GPU Pro Go version

set -e

echo "Building GPU Pro..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build info
VERSION=${VERSION:-"2.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Detect OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

echo -e "${BLUE}Target:${NC} $OS/$ARCH"
echo -e "${BLUE}Version:${NC} $VERSION"
echo -e "${BLUE}Build Time:${NC} $BUILD_TIME"
echo -e "${BLUE}Git Commit:${NC} $GIT_COMMIT"

# Install dependencies
echo -e "\n${BLUE}Installing dependencies...${NC}"
go mod download
go mod tidy

# Build flags
LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

# Build binary
echo -e "\n${BLUE}Building binary...${NC}"
CGO_ENABLED=1 go build -ldflags="$LDFLAGS" -o gpu-pro

# Get binary size
SIZE=$(du -h gpu-pro | cut -f1)
echo -e "\n${GREEN}Build successful!${NC}"
echo -e "${BLUE}Binary:${NC} ./gpu-pro"
echo -e "${BLUE}Size:${NC} $SIZE"

# Optional: Compress with UPX if available
if command -v upx &> /dev/null; then
    echo -e "\n${BLUE}UPX detected. Compress binary? [y/N]${NC}"
    read -r -n 1 compress
    echo
    if [[ $compress =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}Compressing with UPX...${NC}"
        upx --best --lzma gpu-pro
        SIZE=$(du -h gpu-pro | cut -f1)
        echo -e "${GREEN}Compressed size:${NC} $SIZE"
    fi
fi

echo -e "\n${GREEN}Done! Run with:${NC} ./gpu-pro"
