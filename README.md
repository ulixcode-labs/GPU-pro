<div align="center">

# GPU Pro

### Master Your AI Workflow

**Professional GPU & System Monitoring Platform**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![NVIDIA](https://img.shields.io/badge/NVIDIA-GPU-76B900?style=flat&logo=nvidia)](https://www.nvidia.com)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey?style=flat)]()

Real-time GPU monitoring with stunning web interface and terminal UI • Zero dependencies • Single binary • Cross-platform

[🚀 Quick Start](#-quick-start) • [✨ Features](#-features) • [📸 Screenshots](#-screenshots) • [🛠️ Installation](#️-installation) • [📚 Full Documentation](DOCUMENTATION.md)

</div>

---

## 🎯 Why GPU Pro?

GPU Pro is the **modern solution** for NVIDIA GPU monitoring, designed for AI engineers, ML researchers, and GPU cluster administrators. Get real-time insights into your GPU infrastructure with:

- 🎨 **Beautiful Web UI** - Sleek, responsive dashboard with real-time updates
- 💻 **Terminal UI** - Elegant TUI for SSH sessions and headless servers
- 📊 **Comprehensive Metrics** - GPU utilization, memory, temperature, power, processes, and more
- 🌐 **Network Monitoring** - Track connections, bandwidth, and geolocation
- 💾 **System Insights** - CPU, RAM, disk, and fan monitoring
- 🔌 **Hub Mode** - Aggregate multiple GPU nodes into one dashboard
- 🚀 **Zero Setup** - Single binary, no Python, no Node.js, no containers required
- ⚡ **Lightning Fast** - Built with Go for maximum performance

---

## ✨ Features

### 🖥️ GPU Monitoring
- **Real-time Metrics**: Utilization, memory usage, temperature, power consumption
- **Process Tracking**: See what's running on each GPU with detailed process information
- **Multi-GPU Support**: Monitor all GPUs simultaneously
- **Historical Charts**: Track performance over time with customizable time ranges
- **NVML & nvidia-smi**: Supports both native NVML library and nvidia-smi fallback

### 🌐 System Monitoring
- **Network I/O**: Real-time bandwidth tracking with historical charts
- **Disk I/O**: Monitor read/write operations and throughput
- **Network Connections**: Live connection tracking with geolocation on world map
- **Open Files**: Track open file descriptors and large files
- **System Resources**: CPU, RAM, disk usage, and fan speeds

### 🎨 User Interface
- **Web Dashboard**: Modern, responsive design with glassmorphism effects
- **Terminal UI (TUI)**: Beautiful colored terminal interface for SSH sessions
- **Dark Theme**: Professional tech-inspired design optimized for long sessions
- **Real-time Updates**: WebSocket-based live data streaming
- **Mobile Responsive**: Works perfectly on phones and tablets

### 🏗️ Deployment
- **Single Binary**: Everything embedded - templates, assets, and code
- **Cross-Platform**: Linux, Windows, macOS support
- **Multiple Modes**: Standalone or hub aggregation mode
- **Easy Deployment**: Systemd service, Docker, or bare metal
- **Zero Dependencies**: Just NVIDIA drivers required

---

## 📸 Screenshots

### Web UI Dashboard
> Stunning real-time GPU monitoring with modern design

<video src="https://youtu.be/vwBC54nXOPI" autoplay loop muted playsinline></video>

[![IMAGE ALT TEXT HERE]](https://youtu.be/vwBC54nXOPI)


https://youtu.be/vwBC54nXOPI

### Terminal UI
> Elegant TUI for SSH and headless servers

<video src="https://github.com/ulixcode-labs/GPU-pro/gpu-pro-TUI.webm" autoplay loop muted playsinline></video>


### Network Monitoring
> Live connection tracking with global geolocation

---

## 🚀 Quick Start

**The fastest way to get started:**

```bash
wget  https://raw.githubusercontent.com/ulixcode-labs/GPU-pro/refs/heads/main/install.sh && bash install.sh
```

That's it! The script will build the project and let you choose between Web UI or Terminal UI.

---

## 🛠️ Installation

### Prerequisites
- **NVIDIA GPU** with drivers installed
- **Go 1.24+** (for building from source)
- **Linux, Windows, or macOS**

### Option 1: Quick Start Script (Recommended)
```bash
chmod +x start.sh
./start.sh
```

### Option 2: Using Make
```bash
# Build and run Web UI
make run

# Build and run Terminal UI
make run-cli
```

### Option 3: Manual Build
```bash
# Build Web UI version
go build -o gpu-pro

# Build Terminal UI version
go build -o gpu-pro-cli ./cmd/gpu-pro-cli

# Run Web UI (access at http://localhost:1312)
./gpu-pro

# Run Terminal UI
./gpu-pro-cli
```

---

## 📚 Usage

### Web UI Mode (Default)

```bash
# Start with default settings (localhost:1312)
./gpu-pro

# Custom port
PORT=8080 ./gpu-pro

# Enable debug mode
DEBUG=true ./gpu-pro

# Custom update interval (seconds)
UPDATE_INTERVAL=1.0 ./gpu-pro
```

Access the dashboard at `http://localhost:1312` (or your custom port).

### Terminal UI Mode

```bash
# Launch TUI interface
./gpu-pro-cli
```

Perfect for SSH sessions and headless servers!

### Hub Mode (Multi-Node Monitoring)

```bash
# Aggregate multiple GPU nodes
GPU_PRO_MODE=hub \
NODE_URLS=http://node1:1312,http://node2:1312,http://node3:1312 \
./gpu-pro
```

---

## ⚙️ Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server bind address |
| `PORT` | `1312` | Server port |
| `DEBUG` | `false` | Enable debug logging |
| `UPDATE_INTERVAL` | `0.5` | GPU polling interval (seconds) |
| `NVIDIA_SMI_INTERVAL` | `2.0` | nvidia-smi fallback interval (seconds) |
| `NVIDIA_SMI` | `false` | Force nvidia-smi mode |
| `GPU_PRO_MODE` | `default` | Mode: `default` or `hub` |
| `NODE_NAME` | hostname | Node identifier |
| `NODE_URLS` | empty | Comma-separated node URLs (hub mode) |


## 🏗️ Building from Source

### Standard Build
```bash
make build-all-docker
```


## 🎯 Use Cases

- **AI/ML Training**: Monitor GPU utilization during model training
- **Research Labs**: Track multi-GPU workstations and servers
- **GPU Clusters**: Aggregate monitoring across multiple nodes
- **Cloud GPU Instances**: Monitor AWS, GCP, or Azure GPU VMs
- **Gaming Rigs**: Track GPU performance during gaming sessions
- **Crypto Mining**: Monitor mining rig performance and temperatures
- **Remote Work**: SSH-friendly TUI for remote GPU monitoring

---


## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. 🐛 **Report Bugs**: Open an issue with detailed reproduction steps
2. 💡 **Suggest Features**: Share your ideas in the issues section
3. 🔧 **Submit PRs**: Fork, create a feature branch, and submit a pull request
4. 📖 **Improve Docs**: Help us make the documentation better
5. ⭐ **Star the Project**: Show your support!

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🙏 Acknowledgments

- **NVIDIA** for the NVML library
- **Go NVML bindings** by NVIDIA
- **Charm** for beautiful TUI components (Bubble Tea, Lipgloss)
- **Fiber** for blazing-fast web framework
- **Chart.js** for stunning data visualizations

---