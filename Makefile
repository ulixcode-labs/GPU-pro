.PHONY: help build build-cli build-all build-quick build-linux build-macos build-windows \
        build-all-platforms build-native build-linux-only \
        linux-amd64 darwin-amd64 darwin-arm64 windows-amd64 \
        linux-cli darwin-cli darwin-cli-arm64 windows-cli \
        clean run run-cli test deps

# Build variables
BINARY_NAME=gpu-pro
CLI_BINARY_NAME=gpu-pro-cli
DIST_DIR=dist
VERSION?=2.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# CGO settings for different platforms
# Note: NVML requires CGO to be enabled
CGO_ENABLED=1

# Default target
help:
	@echo "GPU Pro - Build Targets"
	@echo ""
	@echo "Development:"
	@echo "  make build        - Build web version for current platform"
	@echo "  make build-cli    - Build CLI/TUI version for current platform"
	@echo "  make run          - Build and run web version"
	@echo "  make run-cli      - Build and run CLI/TUI version"
	@echo "  make test         - Run tests"
	@echo ""
	@echo "Native Builds (Recommended):"
	@echo "  make build-native        - Build for current platform (auto-detect)"
	@echo "  make build-linux-only    - Build Linux binaries (if on Linux)"
	@echo ""
	@echo "Cross-Platform Compilation (requires toolchains):"
	@echo "  make build-all-platforms - Build for all platforms (needs cross-compilers)"
	@echo "  make linux-amd64         - Build web UI for Linux (AMD64)"
	@echo "  make darwin-amd64        - Build web UI for macOS (Intel) ⚠️"
	@echo "  make darwin-arm64        - Build web UI for macOS (Apple Silicon) ⚠️"
	@echo "  make windows-amd64       - Build web UI for Windows (AMD64) ⚠️"
	@echo "  make linux-cli           - Build CLI for Linux (AMD64)"
	@echo "  make darwin-cli          - Build CLI for macOS (Intel) ⚠️"
	@echo "  make darwin-cli-arm64    - Build CLI for macOS (Apple Silicon) ⚠️"
	@echo "  make windows-cli         - Build CLI for Windows (AMD64) ⚠️"
	@echo ""
	@echo "  ⚠️  = Requires cross-compilation toolchain (see BUILD_GUIDE.md)"
	@echo ""
	@echo "Legacy (uses build scripts):"
	@echo "  make build-all    - Build all platforms and flavors (web + CLI)"
	@echo "  make build-quick  - Build release binaries for all platforms"
	@echo "  make build-linux  - Build for Linux (all flavors)"
	@echo "  make build-macos  - Build for macOS (all flavors)"
	@echo "  make build-windows- Build for Windows (all flavors)"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Install dependencies"
	@echo ""

# Simple build for current platform (web version)
build:
	@echo "Building web version for current platform..."
	@./build.sh

# Build CLI/TUI version
build-cli:
	@echo "Building CLI version..."
	@go build -o gpu-pro-cli ./cmd/gpu-pro-cli

# ============================================================
# Cross-Platform Compilation Targets
# ============================================================

# Build for all platforms (Web UI + CLI)
# Note: This will fail if cross-compilation tools are not installed
build-all-platforms: linux-amd64 darwin-amd64 darwin-arm64 windows-amd64 linux-cli darwin-cli darwin-cli-arm64 windows-cli
	@echo ""
	@echo "✅ All platforms built successfully!"
	@echo ""
	@echo "Build artifacts in $(DIST_DIR):"
	@ls -lh $(DIST_DIR)/

# Build only for current platform (native build, always works)
build-native:
	@echo "Detecting current platform..."
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	ARCH=$$(uname -m); \
	if [ "$$ARCH" = "x86_64" ]; then ARCH="amd64"; fi; \
	if [ "$$ARCH" = "aarch64" ]; then ARCH="arm64"; fi; \
	if [ "$$OS" = "linux" ]; then \
		echo "Building for Linux ($$ARCH)..."; \
		$(MAKE) linux-$$ARCH linux-cli; \
	elif [ "$$OS" = "darwin" ]; then \
		echo "Building for macOS ($$ARCH)..."; \
		$(MAKE) darwin-$$ARCH darwin-cli-$$ARCH; \
	else \
		echo "Unsupported platform: $$OS"; \
		exit 1; \
	fi
	@echo ""
	@echo "✅ Native platform built successfully!"
	@echo ""
	@ls -lh $(DIST_DIR)/

# Build only Linux binaries (works on Linux)
build-linux-only: linux-amd64 linux-cli
	@echo "✅ Linux binaries built!"
	@ls -lh $(DIST_DIR)/gpu-pro-linux* 2>/dev/null || true

# Create dist directory
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# ============================================================
# Web UI Binaries
# ============================================================

# Linux AMD64 (native build)
linux-amd64: $(DIST_DIR)
	@echo "Building Web UI for Linux (AMD64)..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64"

# Linux ARM64 (native build on ARM)
linux-arm64: $(DIST_DIR)
	@echo "Building Web UI for Linux (ARM64)..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64"

# macOS Intel (requires macOS or cross-compiler)
darwin-amd64: $(DIST_DIR)
	@echo "Building Web UI for macOS (Intel)..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64"

# macOS Apple Silicon (requires macOS or cross-compiler)
darwin-arm64: $(DIST_DIR)
	@echo "Building Web UI for macOS (Apple Silicon)..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64"

# Windows AMD64 (requires mingw-w64 cross-compiler on Linux)
windows-amd64: $(DIST_DIR)
	@echo "Building Web UI for Windows (AMD64)..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# ============================================================
# CLI Binaries
# ============================================================

# Linux CLI
linux-cli: $(DIST_DIR)
	@echo "Building CLI for Linux (AMD64)..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-amd64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-amd64"

# Linux CLI ARM64
linux-cli-arm64: $(DIST_DIR)
	@echo "Building CLI for Linux (ARM64)..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-arm64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-arm64"

# macOS Intel CLI
darwin-cli: $(DIST_DIR)
	@echo "Building CLI for macOS (Intel)..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-amd64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-amd64"

# macOS Apple Silicon CLI
darwin-cli-arm64: $(DIST_DIR)
	@echo "Building CLI for macOS (Apple Silicon)..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-arm64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-arm64"

# Windows CLI
windows-cli: $(DIST_DIR)
	@echo "Building CLI for Windows (AMD64)..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-windows-amd64.exe ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-windows-amd64.exe"

# ============================================================
# Legacy Build Script Targets
# ============================================================

# Build all platforms and flavors
build-all:
	@echo "Building all platforms and flavors..."
	@./build-all.sh --mode all

# Quick build (release only)
build-quick:
	@echo "Building release binaries..."
	@./build-all.sh --mode quick

# Platform-specific builds
build-linux:
	@echo "Building for Linux..."
	@./build-all.sh --platforms linux --flavors release,debug,minimal

build-macos:
	@echo "Building for macOS..."
	@./build-all.sh --platforms darwin --flavors release,debug,minimal

build-windows:
	@echo "Building for Windows..."
	@./build-all.sh --platforms windows --flavors release,debug,minimal

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist/
	@rm -f gpu-pro gpu-pro.exe gpu-pro-cli gpu-pro-cli.exe
	@echo "Done!"

# Run web version locally
run: build
	@echo "Starting GPU Pro (web version)..."
	@./gpu-pro

# Run CLI version locally
run-cli: build-cli
	@echo "Starting GPU Pro (CLI version)..."
	@./gpu-pro-cli

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Done!"
