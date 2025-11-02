# GPU Pro - Quick Start Guide

## ✅ Recommended Build Command

```bash
make build-all-docker
```

This builds:
- **Linux binaries** (AMD64 + ARM64) with **full GPU/NVML support** via Docker
- **macOS binaries** (Intel + Apple Silicon) natively

## 📦 Available Commands

| Command | Description | Status |
|---------|-------------|--------|
| `make build-all-docker` | Build Linux (Docker) + macOS (native) | ✅ **USE THIS** |
| `make docker-build-linux` | Build only Linux binaries with GPU | ✅ Works |
| `make docker-build-all` | Build all platforms in Docker | ⚠️ Has exit code issues* |
| `make build-native` | Build for current platform only | ✅ Works |

\* `docker-build-all` builds binaries successfully but may return exit code 1. Use `build-all-docker` instead.

## 🚀 Quick Deployment

### 1. Build All Platforms
```bash
make build-all-docker
```

### 2. Deploy to Linux Server
```bash
# Copy binary
scp dist/gpu-pro-linux-amd64 user@server:/opt/gpu-pro/

# SSH and run
ssh user@server
chmod +x /opt/gpu-pro/gpu-pro-linux-amd64
./gpu-pro-linux-amd64
```

## 📊 Output Binaries

After running `make build-all-docker`, you'll have **exactly 8 binaries** (2 per OS/architecture):

**Linux (with GPU ✅):**
- `dist/gpu-pro-linux-amd64` (Web UI, 8.9MB)
- `dist/gpu-pro-cli-linux-amd64` (TUI, 7.4MB)
- `dist/gpu-pro-linux-arm64` (Web UI, 8.4MB)
- `dist/gpu-pro-cli-linux-arm64` (TUI, 6.9MB)

**macOS:**
- `dist/gpu-pro-darwin-amd64` (Web UI, 8.7MB)
- `dist/gpu-pro-cli-darwin-amd64` (TUI, 7.2MB)
- `dist/gpu-pro-darwin-arm64` (Web UI, 8.3MB)
- `dist/gpu-pro-cli-darwin-arm64` (TUI, 6.8MB)

**Note:** No debug/minimal/release variants - just clean production binaries!

## ❓ Troubleshooting

### "make: *** [docker-build-all] Error 1"

**Solution**: Use `make build-all-docker` instead.

The `docker-build-all` command tries to build everything including Windows binaries which may fail on some systems. The `build-all-docker` command is optimized to build only Linux (via Docker with GPU) and macOS (natively).

### "Docker is not running"

**Solution**: Start Docker Desktop and wait for it to initialize.

### Binaries don't work on Linux

**Requirements on target Linux system:**
- NVIDIA GPU
- NVIDIA drivers (450.80.02+)
- Linux kernel 3.10+

The binaries are statically linked and don't require additional libraries.

## 📖 More Documentation

- `BUILD_LINUX_WITH_DOCKER.md` - Comprehensive Docker build guide
- `DOCKER_BUILD_SUMMARY.md` - Technical details
- `make help` - All available make targets

## 🎉 You're Ready!

Simply run:
```bash
make build-all-docker
```

Your Linux binaries with full GPU support will be in the `dist/` directory! 🚀
