# Multi-platform Go builder for GPU Pro
# Builds for Linux, Windows, macOS (AMD64 and ARM64)

FROM golang:1.24-bullseye

# Install dependencies for cross-compilation
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    gcc-mingw-w64 \
    g++-mingw-w64 \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    gcc-x86-64-linux-gnu \
    g++-x86-64-linux-gnu \
    zip \
    && rm -rf /var/lib/apt/lists/*

# Install OSXCross for macOS builds (optional, can skip if not needed)
# Note: This requires macOS SDK which has licensing restrictions
# For now, we'll use GOOS/GOARCH without CGO for macOS

# Set up Go environment
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Create workspace
WORKDIR /workspace

# Entry point for builds
COPY docker-build.sh /usr/local/bin/build
RUN chmod +x /usr/local/bin/build

CMD ["/usr/local/bin/build"]
