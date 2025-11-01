.PHONY: help build build-cli build-all build-quick build-linux build-macos build-windows \
        build-all-platforms build-native build-linux-only build-macos-only build-windows-only \
        linux-amd64 darwin-amd64 darwin-arm64 windows-amd64 \
        linux-cli darwin-cli darwin-cli-arm64 windows-cli \
        clean run run-cli test deps platform-info

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
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "                    GPU Pro - Build System                     "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "📦 Quick Start:"
	@echo "  make build-native     - Build for current platform (RECOMMENDED)"
	@echo "  make platform-info    - Show platform support details"
	@echo ""
	@echo "🔨 Development:"
	@echo "  make build            - Build web version for current platform"
	@echo "  make build-cli        - Build CLI/TUI version for current platform"
	@echo "  make run              - Build and run web version"
	@echo "  make run-cli          - Build and run CLI/TUI version"
	@echo "  make test             - Run tests"
	@echo ""
	@echo "🎯 Platform-Specific Builds (Native - Recommended):"
	@echo "  make build-linux-only   - Build Linux binaries (Web + CLI)"
	@echo "  make build-macos-only   - Build macOS binaries (Web + CLI)"
	@echo "  make build-windows-only - Build Windows binaries (Web + CLI)"
	@echo ""
	@echo "🌍 Individual Platform Targets:"
	@echo "  Web UI:"
	@echo "    make linux-amd64      - Linux AMD64 (GPU: ✅ NVML)"
	@echo "    make darwin-amd64     - macOS Intel (GPU: ❌ System metrics only)"
	@echo "    make darwin-arm64     - macOS Apple Silicon (GPU: ❌ System metrics only)"
	@echo "    make windows-amd64    - Windows AMD64 (GPU: ✅ NVML) ⚠️ Needs MinGW"
	@echo ""
	@echo "  CLI/TUI:"
	@echo "    make linux-cli        - Linux AMD64 CLI (GPU: ✅ NVML)"
	@echo "    make darwin-cli       - macOS Intel CLI (GPU: ❌ System metrics only)"
	@echo "    make darwin-cli-arm64 - macOS Apple Silicon CLI (GPU: ❌ System metrics only)"
	@echo "    make windows-cli      - Windows AMD64 CLI (GPU: ✅ NVML) ⚠️ Needs MinGW"
	@echo ""
	@echo "🔧 Legacy (uses build-all.sh script):"
	@echo "  make build-all        - Build all platforms and flavors"
	@echo "  make build-quick      - Build release binaries for all platforms"
	@echo ""
	@echo "🧹 Utilities:"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make deps             - Install dependencies"
	@echo ""
	@echo "ℹ️  Platform Support:"
	@echo "  Linux   - Full GPU support via NVML (requires CGO)"
	@echo "  Windows - Full GPU support via NVML (requires CGO + MinGW)"
	@echo "  macOS   - System metrics only, GPU disabled (no CGO needed)"
	@echo ""
	@echo "📖 Documentation: See CROSS_PLATFORM_BUILD.md for details"
	@echo "═══════════════════════════════════════════════════════════════"
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

# Build only Linux binaries (Web + CLI)
build-linux-only: linux-amd64 linux-cli
	@echo ""
	@echo "✅ Linux binaries built successfully!"
	@echo ""
	@ls -lh $(DIST_DIR)/ | grep linux || true

# Build only macOS binaries (Web + CLI)
build-macos-only: darwin-amd64 darwin-arm64 darwin-cli darwin-cli-arm64
	@echo ""
	@echo "✅ macOS binaries built successfully!"
	@echo ""
	@ls -lh $(DIST_DIR)/ | grep darwin || true

# Build only Windows binaries (Web + CLI)
build-windows-only: windows-amd64 windows-cli
	@echo ""
	@echo "✅ Windows binaries built successfully!"
	@echo ""
	@ls -lh $(DIST_DIR)/ | grep windows || true

# Create dist directory
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# ============================================================
# Web UI Binaries
# ============================================================

# Linux AMD64 (native build - uses NVML with CGO)
linux-amd64: $(DIST_DIR)
	@echo "Building Web UI for Linux (AMD64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags=linux $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64"

# Linux ARM64 (native build on ARM - uses NVML with CGO)
linux-arm64: $(DIST_DIR)
	@echo "Building Web UI for Linux (ARM64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -tags=linux $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64"

# macOS Intel (no GPU support, CGO disabled)
darwin-amd64: $(DIST_DIR)
	@echo "Building Web UI for macOS (Intel) - GPU support disabled..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64"

# macOS Apple Silicon (no GPU support, CGO disabled)
darwin-arm64: $(DIST_DIR)
	@echo "Building Web UI for macOS (Apple Silicon) - GPU support disabled..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64"

# Windows AMD64 (uses NVML with CGO, requires mingw-w64 cross-compiler)
windows-amd64: $(DIST_DIR)
	@echo "Building Web UI for Windows (AMD64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -tags=windows $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# ============================================================
# CLI Binaries
# ============================================================

# Linux CLI (uses NVML with CGO)
linux-cli: $(DIST_DIR)
	@echo "Building CLI for Linux (AMD64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags=linux $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-amd64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-amd64"

# Linux CLI ARM64 (uses NVML with CGO)
linux-cli-arm64: $(DIST_DIR)
	@echo "Building CLI for Linux (ARM64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -tags=linux $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-arm64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-linux-arm64"

# macOS Intel CLI (no GPU support, CGO disabled)
darwin-cli: $(DIST_DIR)
	@echo "Building CLI for macOS (Intel) - GPU support disabled..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-amd64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-amd64"

# macOS Apple Silicon CLI (no GPU support, CGO disabled)
darwin-cli-arm64: $(DIST_DIR)
	@echo "Building CLI for macOS (Apple Silicon) - GPU support disabled..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-arm64 ./cmd/gpu-pro-cli
	@echo "✓ Built: $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-arm64"

# Windows CLI (uses NVML with CGO, requires mingw-w64 cross-compiler)
windows-cli: $(DIST_DIR)
	@echo "Building CLI for Windows (AMD64) - GPU support enabled..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -tags=windows $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-windows-amd64.exe ./cmd/gpu-pro-cli
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

# Show platform support information
platform-info:
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "               GPU Pro - Platform Support Matrix               "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Platform  | GPU Support | NVML | CGO | Build Command"
	@echo "----------|-------------|------|-----|---------------------------"
	@echo "Linux     | ✅ Full     | ✅   | ✅  | make linux-amd64"
	@echo "Windows   | ✅ Full     | ✅   | ✅  | make windows-amd64 (MinGW)"
	@echo "macOS     | ❌ Disabled | ❌   | ❌  | make darwin-amd64"
	@echo ""
	@echo "Current Platform:"
	@uname -s | tr '[:upper:]' '[:lower:]' | sed 's/^/  OS:   /'
	@uname -m | sed 's/^/  Arch: /' | sed 's/x86_64/amd64 (Intel)/' | sed 's/arm64/arm64 (Apple Silicon)/' | sed 's/aarch64/arm64/'
	@echo ""
	@echo "Quick Commands:"
	@echo "  Build for current platform:  make build-native"
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	if [ "$$OS" = "linux" ]; then \
		echo "  Build Linux only:            make build-linux-only"; \
	elif [ "$$OS" = "darwin" ]; then \
		echo "  Build macOS only:            make build-macos-only"; \
	fi
	@echo "  View all options:            make help"
	@echo ""
	@echo "Implementation Details:"
	@echo "  Source Files:"
	@echo "    - monitor/monitor_linux.go   (Linux GPU via NVML)"
	@echo "    - monitor/monitor_windows.go (Windows GPU via NVML)"
	@echo "    - monitor/monitor_darwin.go  (macOS stub, no GPU)"
	@echo ""
	@echo "  Build Tags:"
	@echo "    - Linux:   -tags=linux   CGO_ENABLED=1"
	@echo "    - Windows: -tags=windows CGO_ENABLED=1"
	@echo "    - macOS:   -tags=darwin  CGO_ENABLED=0"
	@echo ""
	@echo "Documentation: CROSS_PLATFORM_BUILD.md"
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
