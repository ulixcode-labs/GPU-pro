.PHONY: help build build-cli \
        docker-build-linux docker-build-all build-all-docker \
        clean run run-cli test deps

# Build variables
BINARY_NAME=gpu-pro
CLI_BINARY_NAME=gpu-pro-cli
DIST_DIR=dist
VERSION?=2.0.0

# Cross-platform build time detection
ifeq ($(OS),Windows_NT)
	BUILD_TIME=$(shell powershell -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'")
	BINARY_EXT=.exe
	RM=del /Q
	RMDIR=rmdir /S /Q
	MKDIR=if not exist $(DIST_DIR) mkdir $(DIST_DIR)
	EXEC_PREFIX=
else
	BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
	BINARY_EXT=
	RM=rm -f
	RMDIR=rm -rf
	MKDIR=mkdir -p $(DIST_DIR)
	EXEC_PREFIX=./
endif

LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
help:
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "                    GPU Pro - Build System                     "
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "📦 Main Commands:"
	@echo "  make build-all-docker   - Build all platforms (RECOMMENDED)"
	@echo "  make docker-build-linux - Build Linux with GPU support"
	@echo "  make docker-build-all   - Build all via Docker"
	@echo ""
	@echo "🔨 Development:"
	@echo "  make build              - Build web UI for current platform"
	@echo "  make build-cli          - Build CLI/TUI for current platform"
	@echo "  make run                - Build and run web UI"
	@echo "  make run-cli            - Build and run CLI/TUI"
	@echo "  make test               - Run tests"
	@echo ""
	@echo "🧹 Utilities:"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make deps               - Install dependencies"
	@echo ""
	@echo "📊 Output (8 binaries - 2 per OS/arch):"
	@echo "  Linux:  gpu-pro-linux-{amd64,arm64} + cli variants (with GPU ✅)"
	@echo "  macOS:  gpu-pro-darwin-{amd64,arm64} + cli variants"
	@echo ""
	@echo "📖 Documentation: QUICK_START.md"
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""

# ============================================================
# Development Builds (Current Platform)
# ============================================================

# Build web UI for current platform
build:
	@echo "Building web UI for current platform..."
	@go build $(LDFLAGS) -o $(BINARY_NAME)$(BINARY_EXT) .
	@echo "✓ Built: $(BINARY_NAME)$(BINARY_EXT)"

# Build CLI/TUI for current platform
build-cli:
	@echo "Building CLI/TUI for current platform..."
	@go build $(LDFLAGS) -o $(CLI_BINARY_NAME)$(BINARY_EXT) ./cmd/gpu-pro-cli
	@echo "✓ Built: $(CLI_BINARY_NAME)$(BINARY_EXT)"

# ============================================================
# Docker Builds (Production - Cross-platform with GPU)
# ============================================================

# Build Linux binaries with full NVML/GPU support using Docker
docker-build-linux:
	@echo "Building Linux binaries with Docker (includes NVML/GPU support)..."
	@./build-docker-linux.sh

# Build all platforms using Docker
docker-build-all:
	@echo "Building all platforms with Docker..."
	@docker build -f Dockerfile.builder -t gpu-pro-builder .
	@docker run --rm -v "$$(pwd):/workspace" gpu-pro-builder

# Build all platforms (Linux via Docker + macOS natively) - RECOMMENDED
build-all-docker:
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "Building all platforms (Linux via Docker, macOS natively)..."
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Step 1/2: Building Linux binaries via Docker with NVML..."
	@$(MAKE) docker-build-linux
	@echo ""
	@echo "Step 2/2: Building macOS binaries natively..."
	@$(MKDIR)
	@echo "Building macOS Intel binaries..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-amd64 ./cmd/gpu-pro-cli
	@echo "Building macOS Apple Silicon binaries..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags=darwin $(LDFLAGS) -o $(DIST_DIR)/$(CLI_BINARY_NAME)-darwin-arm64 ./cmd/gpu-pro-cli
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "✅ All platforms built successfully!"
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Linux binaries (with GPU support):"
	@ls -lh $(DIST_DIR)/*linux* 2>/dev/null | awk '{print "  " $$9, "(" $$5 ")"}'
	@echo ""
	@echo "macOS binaries:"
	@ls -lh $(DIST_DIR)/*darwin* 2>/dev/null | awk '{print "  " $$9, "(" $$5 ")"}'
	@echo ""

# ============================================================
# Utilities
# ============================================================

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	@if exist $(DIST_DIR) $(RMDIR) $(DIST_DIR)
	@if exist $(BINARY_NAME).exe $(RM) $(BINARY_NAME).exe
	@if exist $(CLI_BINARY_NAME).exe $(RM) $(CLI_BINARY_NAME).exe
	@if exist $(BINARY_NAME) $(RM) $(BINARY_NAME)
	@if exist $(CLI_BINARY_NAME) $(RM) $(CLI_BINARY_NAME)
else
	@$(RMDIR) $(DIST_DIR)/
	@$(RM) $(BINARY_NAME) $(BINARY_NAME).exe $(CLI_BINARY_NAME) $(CLI_BINARY_NAME).exe
endif
	@echo "✅ Done!"

# Run web UI locally
run: build
	@echo "Starting GPU Pro (web UI)..."
	@$(EXEC_PREFIX)$(BINARY_NAME)$(BINARY_EXT)

# Run CLI/TUI locally
run-cli: build-cli
	@echo "Starting GPU Pro (CLI/TUI)..."
	@$(EXEC_PREFIX)$(CLI_BINARY_NAME)$(BINARY_EXT)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Done!"
