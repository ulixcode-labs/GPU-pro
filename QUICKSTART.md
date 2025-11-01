# Quick Start - Cross-Platform Builds

## TL;DR

```bash
# macOS (current platform)
GOOS=darwin CGO_ENABLED=0 go build -tags=darwin -o gpu-pro

# Linux (requires Linux machine or Docker)
GOOS=linux CGO_ENABLED=1 go build -tags=linux -o gpu-pro

# Windows (requires Windows machine or MinGW)
GOOS=windows CGO_ENABLED=1 go build -tags=windows -o gpu-pro.exe
```

## Build on Current Platform

```bash
# Automated build for current OS
./build-all.sh --mode single --flavors release

# Find binary in dist/
ls -lh dist/
```

## What Works Where

| Feature | Linux | Windows | macOS |
|---------|-------|---------|-------|
| GPU Monitoring | ✅ | ✅ | ❌ |
| System Metrics | ✅ | ✅ | ✅ |
| NVML Support | ✅ | ✅ | ❌ |
| CGO Required | ✅ | ✅ | ❌ |

## Expected Behavior

### macOS
```
🍎 Running on macOS - GPU monitoring is disabled
✓  System metrics will be available
```

### Linux/Windows
```
NVML initialized - Driver: XXX.XX
Detected N GPU(s)
GPU 0 (NVIDIA ...): Using NVML (utilization: X.X%)
```

## Need Help?

- Build issues: See `CROSS_PLATFORM_BUILD.md`
- Implementation details: See `PLATFORM_SUMMARY.md`
- Run the binary: `./gpu-pro` (will start on port 8889)

## File Structure

```
monitor/
├── monitor_linux.go      # Linux NVML implementation
├── monitor_windows.go    # Windows NVML implementation
├── monitor_darwin.go     # macOS stub (no GPU)
├── metrics_linux.go      # Linux metrics
├── metrics_windows.go    # Windows metrics
├── monitor.go.bak        # Backup of original
└── metrics.go.bak        # Backup of original
```
