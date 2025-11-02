#!/bin/bash
# Quick script to build Linux binaries with full GPU/NVML support using Docker

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   GPU Pro - Docker Linux Builder${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${YELLOW}Error: Docker is not running${NC}"
    echo "Please start Docker and try again."
    exit 1
fi

# Build the Docker image
echo -e "${YELLOW}Step 1/2: Building Docker image...${NC}"
docker build -f Dockerfile.builder -t gpu-pro-builder .

echo ""
echo -e "${YELLOW}Step 2/2: Building Linux binaries with NVML support...${NC}"

# Run the build inside Docker and copy binaries out
docker run --rm \
    -v "$(pwd):/workspace" \
    -e VERSION="${VERSION:-2.0.0}" \
    gpu-pro-builder \
    bash -c "
        # Only build Linux binaries
        export VERSION=${VERSION:-2.0.0}
        export BUILD_TIME=\$(date -u '+%Y-%m-%d_%H:%M:%S')
        export GIT_COMMIT=\$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

        mkdir -p dist

        echo 'Building Linux AMD64 Web UI with NVML...'
        # Detect host architecture and use appropriate compiler
        HOST_ARCH=\$(uname -m)
        if [ \"\$HOST_ARCH\" = \"aarch64\" ] || [ \"\$HOST_ARCH\" = \"arm64\" ]; then
            # Cross-compile from ARM64 to AMD64
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ \
                go build -tags=linux \
                -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
                -o dist/gpu-pro-linux-amd64 .
        else
            # Native AMD64 build
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc CXX=g++ \
                go build -tags=linux \
                -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
                -o dist/gpu-pro-linux-amd64 .
        fi

        echo 'Building Linux AMD64 CLI with NVML...'
        if [ \"\$HOST_ARCH\" = \"aarch64\" ] || [ \"\$HOST_ARCH\" = \"arm64\" ]; then
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ \
                go build -tags=linux \
                -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
                -o dist/gpu-pro-cli-linux-amd64 ./cmd/gpu-pro-cli
        else
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc CXX=g++ \
                go build -tags=linux \
                -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
                -o dist/gpu-pro-cli-linux-amd64 ./cmd/gpu-pro-cli
        fi

        echo 'Building Linux ARM64 Web UI with NVML...'
        CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
            go build -tags=linux \
            -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
            -o dist/gpu-pro-linux-arm64 .

        echo 'Building Linux ARM64 CLI with NVML...'
        CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
            go build -tags=linux \
            -ldflags='-w -s -X main.Version='\$VERSION' -X main.BuildTime='\$BUILD_TIME' -X main.GitCommit='\$GIT_COMMIT \
            -o dist/gpu-pro-cli-linux-arm64 ./cmd/gpu-pro-cli

        echo ''
        echo 'Build complete! Generated binaries:'
        ls -lh dist/gpu-pro*linux* 2>/dev/null || true
    "

echo ""
echo -e "${GREEN}✅ Success!${NC}"
echo ""
echo -e "${BLUE}Generated Linux binaries with full NVML/GPU support:${NC}"
ls -lh dist/gpu-pro*linux* 2>/dev/null || echo "No binaries found in dist/"
echo ""
echo -e "${BLUE}Usage:${NC}"
echo -e "  Copy to Linux server:  scp dist/gpu-pro-linux-amd64 user@server:/path/"
echo -e "  Run on Linux:          ./gpu-pro-linux-amd64"
echo ""
