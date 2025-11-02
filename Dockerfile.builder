# Multi-platform Go builder for GPU Pro
# Builds for Linux, Windows, macOS (AMD64 and ARM64)

FROM golang:1.23-bullseye

# Allow Go to auto-download the required version
ENV GOTOOLCHAIN=auto

# Install dependencies for cross-compilation and NVML
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    gcc-mingw-w64 \
    g++-mingw-w64 \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    gcc-x86-64-linux-gnu \
    g++-x86-64-linux-gnu \
    crossbuild-essential-amd64 \
    crossbuild-essential-arm64 \
    zip \
    wget \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and install NVML headers manually (lightweight approach)
# We only need the headers for compilation, not the full CUDA toolkit
RUN mkdir -p /usr/local/cuda/include /usr/local/cuda/lib64 && \
    cd /tmp && \
    wget -q https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-nvml-dev-11-8_11.8.89-1_amd64.deb && \
    dpkg -x cuda-nvml-dev-11-8_11.8.89-1_amd64.deb /tmp/nvml && \
    cp -r /tmp/nvml/usr/local/cuda-11.8/targets/x86_64-linux/include/* /usr/local/cuda/include/ && \
    cp -r /tmp/nvml/usr/local/cuda-11.8/targets/x86_64-linux/lib/stubs/* /usr/local/cuda/lib64/ 2>/dev/null || true && \
    rm -rf /tmp/nvml /tmp/*.deb

# Set environment variables for CGO to find NVML
ENV CGO_CFLAGS="-I/usr/local/cuda/include"
ENV CGO_LDFLAGS="-L/usr/local/cuda/lib64"

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
