# Building Linux Binaries with Full GPU Support using Docker

This guide explains how to build Linux binaries with full NVML/GPU support from macOS (or any platform) using Docker.

## Why Use Docker?

When cross-compiling from macOS to Linux with CGO enabled (required for NVML GPU support), you need:
- Linux headers and libraries
- Cross-compilation toolchain
- NVML development libraries

Docker provides a Linux environment with all these dependencies pre-installed, making it easy to build fully-featured Linux binaries from any platform.

## Prerequisites

- Docker Desktop installed and running
- At least 2GB of free disk space for the Docker image

## Quick Start

### Option 1: Using Make (Recommended)

```bash
# Build Linux binaries with full NVML/GPU support
make docker-build-linux
```

This will:
1. Build a Docker image with Go, GCC, and NVML headers
2. Compile Linux binaries (AMD64 and ARM64) inside the container
3. Copy the binaries to your local `dist/` directory

### Option 2: Using the Shell Script Directly

```bash
./build-docker-linux.sh
```

### Option 3: Manual Docker Build

```bash
# Build the Docker image (one time)
docker build -f Dockerfile.builder -t gpu-pro-builder .

# Run the build
docker run --rm -v "$(pwd):/workspace" gpu-pro-builder bash -c "
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
        go build -tags=linux -o dist/gpu-pro-linux-amd64 .

    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
        go build -tags=linux -o dist/gpu-pro-cli-linux-amd64 ./cmd/gpu-pro-cli
"
```

## Output

After the build completes, you'll find the following binaries in the `dist/` directory:

```
dist/
├── gpu-pro-linux-amd64        # Web UI for Linux AMD64 (with GPU support)
├── gpu-pro-linux-arm64        # Web UI for Linux ARM64 (with GPU support)
├── gpu-pro-cli-linux-amd64    # CLI/TUI for Linux AMD64 (with GPU support)
└── gpu-pro-cli-linux-arm64    # CLI/TUI for Linux ARM64 (with GPU support)
```

## Deploying to Linux Server

### Copy Binary to Server

```bash
# Copy Web UI version
scp dist/gpu-pro-linux-amd64 user@server:/opt/gpu-pro/

# Or copy CLI version
scp dist/gpu-pro-cli-linux-amd64 user@server:/opt/gpu-pro/
```

### Run on Linux Server

```bash
# SSH into the server
ssh user@server

# Make executable
chmod +x /opt/gpu-pro/gpu-pro-linux-amd64

# Run (requires NVIDIA GPU and drivers installed)
./gpu-pro-linux-amd64
```

## Features

### ✅ What's Included

- **Full NVML Support**: Real-time GPU monitoring via NVIDIA Management Library
- **CGO Enabled**: Native C library integration for maximum performance
- **Cross-Platform**: Build from macOS, Windows, or Linux
- **Multi-Architecture**: AMD64 and ARM64 support
- **Both Versions**: Web UI and CLI/TUI binaries

### 📝 Requirements on Target Linux System

The binary requires:
- Linux kernel 3.10+
- NVIDIA GPU with drivers installed
- NVIDIA driver version 450.80.02+ (for CUDA 11.0+)

The binary itself is statically linked and doesn't require any additional libraries to be installed.

## Docker Image Details

The `Dockerfile.builder` creates an image with:

- **Base**: Go 1.23 on Debian Bullseye
- **Compilers**: GCC, G++, MinGW (for Windows), ARM cross-compilers
- **NVIDIA**: CUDA NVML development headers and libraries
- **Size**: ~2-3GB (cached after first build)

The image is only built once and reused for subsequent builds.

## Troubleshooting

### Docker not running

```
Error: Docker is not running
Please start Docker and try again.
```

**Solution**: Start Docker Desktop and wait for it to fully initialize.

### Permission denied

```
permission denied while trying to connect to the Docker daemon
```

**Solution**: Make sure your user is in the `docker` group or run with `sudo` (not recommended).

### Build fails with "nvml.h: No such file"

This means the CUDA toolkit wasn't properly installed in the Docker image.

**Solution**: Rebuild the Docker image:
```bash
docker build -f Dockerfile.builder -t gpu-pro-builder . --no-cache
```

## Comparison: Docker Build vs Native Build

| Feature | Docker Build | Native macOS Build |
|---------|--------------|-------------------|
| GPU Support (Linux) | ✅ Full NVML | ❌ Disabled |
| CGO | ✅ Enabled | ❌ Not available |
| Speed | ~2-3 min (first), ~30s (cached) | ~10s |
| Setup Required | Docker only | None |
| Binary Size | ~10MB | ~8MB |
| Performance | Native (no overhead) | Native (no overhead) |

## Advanced Usage

### Build with Custom Version

```bash
VERSION=3.0.0 make docker-build-linux
```

### Build Only AMD64

```bash
docker run --rm -v "$(pwd):/workspace" gpu-pro-builder bash -c "
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
        go build -tags=linux -ldflags='-s -w' \
        -o dist/gpu-pro-linux-amd64 .
"
```

### Inspect the Docker Image

```bash
# Run interactive shell in the builder image
docker run --rm -it -v "$(pwd):/workspace" gpu-pro-builder bash

# Inside the container, you can:
# - Check Go version: go version
# - Check NVML headers: ls -la /usr/local/cuda/targets/x86_64-linux/include/nvml.h
# - Test build manually
```

## CI/CD Integration

The Docker build is ideal for CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Build Linux binaries
  run: |
    docker build -f Dockerfile.builder -t gpu-pro-builder .
    make docker-build-linux

- name: Upload artifacts
  uses: actions/upload-artifact@v3
  with:
    name: linux-binaries
    path: dist/gpu-pro-linux-*
```

## Summary

The Docker-based build system allows you to:
1. Build Linux binaries with full GPU support from any platform
2. Avoid complex cross-compilation toolchain setup
3. Ensure consistent builds across different development machines
4. Create production-ready binaries with NVML integration

Simply run `make docker-build-linux` and your binaries will be ready to deploy!
