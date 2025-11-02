.PHONY: help build build-cli \
        docker-build-linux docker-build-all build-all-docker \
        clean run run-cli test deps

# Build variables
BINARY_NAME=gpu-pro
CLI_BINARY_NAME=gpu-pro-cli
DIST_DIR=dist
VERSION?=2.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
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
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Built: $(BINARY_NAME)"

# Build CLI/TUI for current platform
build-cli:
	@echo "Building CLI/TUI for current platform..."
	@go build $(LDFLAGS) -o $(CLI_BINARY_NAME) ./cmd/gpu-pro-cli
	@echo "✓ Built: $(CLI_BINARY_NAME)"

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
	@mkdir -p $(DIST_DIR)
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
	@rm -rf $(DIST_DIR)/
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(CLI_BINARY_NAME) $(CLI_BINARY_NAME).exe
	@echo "✅ Done!"

# Run web UI locally
run: build
	@echo "Starting GPU Pro (web UI)..."
	@./$(BINARY_NAME)

# Run CLI/TUI locally
run-cli: build-cli
	@echo "Starting GPU Pro (CLI/TUI)..."
	@./$(CLI_BINARY_NAME)

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
