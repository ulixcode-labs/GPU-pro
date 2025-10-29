<div align="center">

# GPU Pro - Complete Documentation

**Professional GPU & System Monitoring Platform**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org) [![NVIDIA](https://img.shields.io/badge/NVIDIA-GPU-76B900?style=flat&logo=nvidia)](https://www.nvidia.com) [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE) [![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey)]()

Real-time GPU monitoring with stunning web interface and terminal UI • Zero dependencies • Single binary • Cross-platform

</div>

---

## 📑 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Features](#-features)
- [Building from Source](#-building-from-source)
- [Usage](#-usage)
- [Alert System](#-alert-system)
- [Network Access](#-network-access)
- [Hub Mode](#-hub-mode)
- [CLI Reference](#-cli-reference)
- [Configuration](#-configuration)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)

---

## 🚀 Quick Start

### Install (One Command)

**Linux/macOS:**
```bash
curl -fsSL https://install.gpu-pro.dev | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://install.gpu-pro.dev/windows | iex
```

### Run

```bash
# Web dashboard (http://localhost:8080)
gpu-pro

# Custom port
PORT=3000 gpu-pro

# Terminal UI
gpu-pro-cli

# Hub mode (aggregate multiple nodes)
MODE=hub NODE_URLS=http://node1:8080,http://node2:8080 gpu-pro
```

**No GPU?** GPU Pro works perfectly fine! You'll get full system monitoring (CPU, RAM, disk, network) without NVIDIA hardware.

---

## 📦 Installation

### Method 1: Pre-built Binaries (Recommended)

**Download from releases:**
```bash
# Linux x64
wget https://github.com/YOUR_REPO/gpu-pro/releases/latest/download/gpu-pro-linux-amd64
chmod +x gpu-pro-linux-amd64
sudo mv gpu-pro-linux-amd64 /usr/local/bin/gpu-pro

# macOS (Apple Silicon)
wget https://github.com/YOUR_REPO/gpu-pro/releases/latest/download/gpu-pro-darwin-arm64
chmod +x gpu-pro-darwin-arm64
sudo mv gpu-pro-darwin-arm64 /usr/local/bin/gpu-pro
```

### Method 2: Build from Source

**Prerequisites:**
- Go 1.24+
- NVIDIA drivers (optional, for GPU monitoring)

**Build:**
```bash
git clone https://github.com/YOUR_REPO/gpu-pro.git
cd gpu-pro
make build        # Current platform
make build-all    # All platforms
```

**Quick builds:**
```bash
./build.sh              # Release build (current platform)
./build-all.sh          # All platforms, all flavors
./build-all.sh --quick  # Release only, faster
```

### Method 3: Go Install

```bash
go install github.com/YOUR_REPO/gpu-pro/cmd/gpu-pro@latest
go install github.com/YOUR_REPO/gpu-pro/cmd/gpu-pro-cli@latest
```

### Method 4: Docker

```bash
docker run -p 8080:8080 --gpus all gpu-pro/gpu-pro:latest
```

---

## ✨ Features

### 🖥️ GPU Monitoring
- ✅ Real-time metrics (utilization, memory, temperature, power)
- ✅ Process tracking with command lines and resource usage
- ✅ Multi-GPU support
- ✅ Historical charts (1min, 5min, 15min views)
- ✅ GPU clock speeds and performance states
- ✅ Fan speed monitoring
- ✅ NVML native support + nvidia-smi fallback

### 🌐 System Monitoring
- ✅ **Network I/O**: Real-time bandwidth with charts
- ✅ **Disk I/O**: Read/write operations tracking
- ✅ **Network Connections**: Live connection list with geolocation world map
- ✅ **Open Files**: Track file descriptors
- ✅ **Large Files**: Find largest files in any directory
- ✅ **CPU & Memory**: Host system metrics
- ✅ **Fan Speeds**: System fan monitoring (Linux)

### 🔔 Alert System

**Web Interface:**
- Real-time threshold monitoring (Temperature, Memory, Power)
- Visual alerts with banner and side panel
- Browser desktop notifications
- Snooze alerts (5 min / 30 min)
- Acknowledge to dismiss
- Configurable thresholds via sliders
- Alert history with timestamps

**CLI Interface:**
- Interactive alert view (press `a`)
- Navigation with arrow keys
- Snooze and acknowledge actions
- Auto-resolve when metrics normalize
- TTL-based expiration (active: 1hr, resolved: 30s, ack: 1min)

**Default Thresholds:**
| Metric | Warning | Critical |
|--------|---------|----------|
| Temperature | 75°C | 85°C |
| Memory | 85% | 95% |
| Power | 90% | 98% |

### 🎨 User Interfaces

**Web Dashboard:**
- Modern glassmorphism design
- Real-time updates (500ms)
- Responsive mobile layout
- Chart.js visualizations
- Dark theme optimized

**Terminal UI (CLI):**
- Elegant TUI with keyboard navigation
- Color-coded metrics
- Live process monitoring
- Interactive alert management
- Works over SSH

### 🔌 Hub Mode
Aggregate multiple GPU nodes into one dashboard:
```bash
MODE=hub NODE_URLS=http://192.168.1.10:8080,http://192.168.1.11:8080 gpu-pro
```

---

## 🔨 Building from Source

### Simple Build

```bash
# Clone repository
git clone https://github.com/YOUR_REPO/gpu-pro.git
cd gpu-pro

# Build for current platform
go build -o gpu-pro cmd/gpu-pro/main.go
go build -o gpu-pro-cli cmd/gpu-pro-cli/main.go
```

### Advanced Build

```bash
# Using Makefile
make build              # Current platform
make build-linux        # Linux only
make build-macos        # macOS only
make build-windows      # Windows only
make build-all          # All platforms, all flavors

# Using build script
./build-all.sh                                    # All platforms, all flavors
./build-all.sh --quick                            # Release only (faster)
./build-all.sh --platforms linux,darwin           # Specific platforms
./build-all.sh --flavors release,debug            # Specific flavors
```

### Build Flavors

| Flavor | Optimization | Size | Debug Info | Use Case |
|--------|-------------|------|------------|----------|
| **release** | ✅ Full | Small | ❌ Stripped | Production |
| **debug** | ❌ None | Large | ✅ Full | Development |
| **minimal** | ✅ Full | Smallest | ❌ Stripped | Embedded |

### Build Output

```
dist/
├── gpu-pro-linux-amd64-release
├── gpu-pro-linux-arm64-release
├── gpu-pro-darwin-amd64-release
├── gpu-pro-darwin-arm64-release
├── gpu-pro-windows-amd64-release.exe
└── gpu-pro-cli-linux-amd64-release
```

---

## 💡 Usage

### Web Dashboard

```bash
# Start server
gpu-pro

# Custom port
PORT=3000 gpu-pro

# Custom host
HOST=0.0.0.0 PORT=8080 gpu-pro

# Hub mode
MODE=hub NODE_URLS=http://node1:8080,http://node2:8080 gpu-pro
```

**Access:** Open http://localhost:8080 in browser

**Features:**
- Overview: All GPUs at a glance
- Per-GPU tabs: Detailed metrics and charts
- System Metrics: CPU, RAM, disk, network
- Alerts: Click 🔔 icon to view/configure
- Time Range: Switch between 1min, 5min, 15min views

### Terminal UI (CLI)

```bash
gpu-pro-cli
```

**Keyboard Shortcuts:**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate GPUs |
| `a` | Toggle alert view |
| `s` | Snooze alert (5 min) |
| `S` | Snooze alert (30 min) |
| `A` | Acknowledge alert |
| `q` | Quit |

**View Modes:**
- Default: GPU metrics with charts
- Alert View: Press `a` to view/manage alerts

### Command-Line Options

```bash
# View alert history
gpu-pro-cli --view-alerts

# Configure thresholds
gpu-pro-cli --config-thresholds

# Specify config file
gpu-pro-cli --config gpu-thresholds.json
```

---

## 🔔 Alert System

### Web Alerts

**Accessing:**
1. Click 🔔 bell icon in header
2. View active alerts in banner
3. Manage alerts in side panel

**Managing Alerts:**

**Snooze:**
```
Click 💤 5m  → Snooze for 5 minutes
Click 💤 30m → Snooze for 30 minutes
```

**Acknowledge:**
```
Click ✓ → Permanently dismiss alert
```

**Configure Thresholds:**
1. Open alert panel (🔔)
2. Click "Configuration" tab
3. Drag sliders to adjust thresholds
4. Click "💾 Save Configuration"

**Alert States:**
- 🔴 **Active**: Needs attention
- 💤 **Snoozed**: Temporarily hidden (shows remaining time)
- ✓ **Acknowledged**: Permanently dismissed
- ✓ **Resolved**: Auto-resolved when metric returns to normal

### CLI Alerts

**View Alerts:**
```bash
# Within CLI
Press 'a' key

# From command line
gpu-pro-cli --view-alerts
```

**Interactive Alert Management:**
```
→ 🔴 14:32:45 [CRITICAL] GPU 0 - Temperature: 87.0°C
  🟡 14:32:30 [WARNING] GPU 0 - Memory: 86.5% [💤 4m]
  🔴 14:32:15 [CRITICAL] GPU 1 - Power: 99.2% [✓ ACK]

↑/↓: Navigate | s: Snooze 5m | S: Snooze 30m | A: Acknowledge
```

**Lifecycle:**
```
Active (1 hour) → Snoozed → Back to Active
              ↘ Acknowledged (1 min TTL)
              ↘ Resolved (30s TTL) → Auto-expire
```

### Alert Files

```bash
gpu-thresholds.json  # Threshold configuration
gpu-alerts.log       # Alert history (persistent)
```

**Threshold Configuration:**
```json
{
  "temp_warning": 75,
  "temp_critical": 85,
  "memory_warning": 85,
  "memory_critical": 95,
  "power_warning": 90,
  "power_critical": 98
}
```

---

## 🌐 Network Access

### Local Network Access

**Allow external connections:**
```bash
HOST=0.0.0.0 PORT=8080 gpu-pro
```

**Firewall (Linux):**
```bash
sudo ufw allow 8080/tcp
```

**Firewall (Windows):**
```powershell
netsh advfirewall firewall add rule name="GPU Pro" dir=in action=allow protocol=TCP localport=8080
```

### Remote Access via SSH Tunnel

**From remote machine:**
```bash
ssh -L 8080:localhost:8080 user@gpu-server
```

**Then access:** http://localhost:8080

### Reverse Proxy (nginx)

```nginx
server {
    listen 80;
    server_name gpu.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

### Security Considerations

⚠️ **GPU Pro has no built-in authentication**

**Options:**
1. **VPN**: Access only through VPN
2. **SSH Tunnel**: Recommended for remote access
3. **Reverse Proxy**: Add authentication (nginx + basic auth, OAuth)
4. **Firewall**: Restrict to specific IPs

**Production Setup:**
```bash
# Only listen on localhost (use SSH tunnel)
HOST=127.0.0.1 PORT=8080 gpu-pro

# Or restrict by IP in firewall
sudo ufw allow from 192.168.1.0/24 to any port 8080
```

---

## 🔌 Hub Mode

Monitor multiple GPU nodes from a single dashboard.

### Setup

**Node Servers (on each GPU machine):**
```bash
# Node 1 (192.168.1.10)
HOST=0.0.0.0 PORT=8080 gpu-pro

# Node 2 (192.168.1.11)
HOST=0.0.0.0 PORT=8080 gpu-pro
```

**Hub Server (aggregator):**
```bash
MODE=hub NODE_URLS=http://192.168.1.10:8080,http://192.168.1.11:8080 gpu-pro
```

**Access:** http://hub-server:8080

### Hub Features
- Aggregated view of all nodes
- Per-node grouping in UI
- Individual node health status
- Automatic reconnection on failures
- All metrics from all nodes

---

## 📚 CLI Reference

### Web Server Commands

```bash
# Basic
gpu-pro                                    # Start web server (port 8080)
PORT=3000 gpu-pro                          # Custom port
HOST=0.0.0.0 PORT=8080 gpu-pro            # Listen on all interfaces

# Hub mode
MODE=hub NODE_URLS=http://n1:8080,http://n2:8080 gpu-pro

# Environment variables
PORT=8080                 # Server port (default: 8080)
HOST=127.0.0.1           # Listen address (default: localhost)
MODE=monitor             # Mode: monitor, hub (default: monitor)
NODE_NAME="GPU-Node-1"   # Node identifier (hub mode)
NODE_URLS="url1,url2"    # Comma-separated node URLs (hub mode)
```

### Terminal UI Commands

```bash
# Basic
gpu-pro-cli                               # Launch TUI

# Options
gpu-pro-cli --view-alerts                 # View full alert log
gpu-pro-cli --config-thresholds           # Configure thresholds
gpu-pro-cli --config gpu-config.json      # Use specific config
```

### Build Commands

```bash
make build                # Build current platform
make build-linux          # Linux binaries
make build-macos          # macOS binaries
make build-windows        # Windows binaries
make build-all            # All platforms

make clean                # Remove dist/ folder
make test                 # Run tests
make install              # Install to $GOPATH/bin
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Web server port | `8080` |
| `HOST` | Listen address | `localhost` |
| `MODE` | Operating mode | `monitor` |
| `NODE_NAME` | Node identifier | hostname |
| `NODE_URLS` | Hub node URLs | - |

### Configuration Files

**Alert Thresholds** (`gpu-thresholds.json`):
```json
{
  "temp_warning": 75,
  "temp_critical": 85,
  "memory_warning": 85,
  "memory_critical": 95,
  "power_warning": 90,
  "power_critical": 98
}
```

**Alert Log** (`gpu-alerts.log`):
```
[2025-10-28 14:27:39] GPU 0 - warning Memory: 85.0 (threshold: 85.0)
[2025-10-28 14:27:40] GPU 0 - critical Power: 98.2 (threshold: 98.0)
```

### Browser Notifications

**Enable:**
1. GPU Pro will request permission on first load
2. Grant permission in browser
3. Desktop notifications will appear for alerts

**Disable:**
- Browser settings → Notifications → GPU Pro → Block

---

## 🐛 Troubleshooting

### No GPUs Detected

**Check NVIDIA drivers:**
```bash
nvidia-smi
```

**If error:**
```bash
# Ubuntu/Debian
sudo apt install nvidia-driver-xxx

# Verify
nvidia-smi
```

**Still not working?**
- Ensure NVML library is installed
- Check driver version compatibility
- GPU Pro will fallback to nvidia-smi if NVML fails

### Web UI Not Loading

**Check server is running:**
```bash
curl http://localhost:8080
```

**Check firewall:**
```bash
sudo ufw status
sudo ufw allow 8080/tcp
```

**Check port availability:**
```bash
sudo lsof -i :8080
```

### Alerts Not Triggering

**Check thresholds:**
```bash
cat gpu-thresholds.json
```

**Too high?** Lower thresholds in web UI (🔔 → Configuration)

**Check browser console:**
- Press F12 → Console tab
- Look for alert checking logs
- Should see: `Checking alerts for GPU: {...}`

### High CPU Usage

**Reduce update frequency:**
- Modify `socket-handlers.js` → Increase `DOM_UPDATE_INTERVAL`
- Default: 1000ms (1 second)

**Reduce chart history:**
- Use 1-minute view instead of 15-minute
- Fewer data points = less rendering

### Permission Denied

**Linux:**
```bash
# Make binary executable
chmod +x gpu-pro

# Install to system
sudo cp gpu-pro /usr/local/bin/
```

**File access:**
```bash
# Ensure write permissions for alert files
chmod 644 gpu-thresholds.json
chmod 644 gpu-alerts.log
```

---

## 🤝 Contributing

We welcome contributions! Here's how:

### Development Setup

```bash
# Clone repo
git clone https://github.com/YOUR_REPO/gpu-pro.git
cd gpu-pro

# Install dependencies
go mod download

# Run in development
go run cmd/gpu-pro/main.go
```

### Code Structure

```
gpu-pro/
├── cmd/
│   ├── gpu-pro/           # Web server
│   └── gpu-pro-cli/       # Terminal UI
├── monitor/               # GPU monitoring logic
├── handlers/              # Web API handlers
├── config/               # Configuration
├── static/
│   ├── js/               # Frontend JavaScript
│   └── css/              # Styles
├── templates/            # HTML templates
└── dist/                 # Build output
```

### Making Changes

1. **Fork** the repository
2. **Create branch**: `git checkout -b feature/amazing-feature`
3. **Make changes** and test thoroughly
4. **Commit**: `git commit -m "Add amazing feature"`
5. **Push**: `git push origin feature/amazing-feature`
6. **Open Pull Request** with description

### Coding Standards

- **Go**: Follow standard Go conventions (`gofmt`, `golint`)
- **JavaScript**: ES6+ syntax, clear variable names
- **CSS**: Use existing CSS variables for theming
- **Comments**: Document complex logic

### Testing

```bash
# Run tests
go test ./...

# Test build
make build

# Test all platforms
make build-all
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

## 🙏 Acknowledgments

- **Go Fiber** - Web framework
- **Chart.js** - Data visualization
- **gopsutil** - System metrics
- **NVML** - NVIDIA GPU monitoring
- **Leaflet.js** - Network connection map

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/YOUR_REPO/gpu-pro/issues)
- **Discussions**: [GitHub Discussions](https://github.com/YOUR_REPO/gpu-pro/discussions)
- **Documentation**: This file

---

<div align="center">

**Made with ❤️ by the GPU Pro Team**

[⬆ Back to Top](#gpu-pro---complete-documentation)

</div>
