# GPU Pro - Universal Installer

## Quick Start for End Users

```bash
# One command to install and run:
curl -fsSL https://your-domain.com/install.sh | bash
```

That's it! The script will:
1. Detect your OS and architecture
2. Ask you 2 simple questions
3. Download and install GPU Pro
4. Launch it automatically

## What Gets Installed

The installer automatically detects your platform and installs the correct version:

| Your Platform | What You Get | GPU Support |
|---------------|--------------|-------------|
| macOS (Intel or Apple Silicon) | System metrics dashboard | ❌ No GPU (macOS limitation) |
| Linux (AMD64 or ARM64) | Full GPU monitoring | ✅ NVIDIA GPUs via NVML |
| Windows (AMD64) | Full GPU monitoring | ✅ NVIDIA GPUs via NVML |

## Interactive Prompts

### Question 1: Which interface?

```
Select Mode

  1) Web UI    - Dashboard accessible via browser
                 http://localhost:8889
                 Full featured web interface

  2) CLI/TUI   - Terminal-based interface
                 Interactive text UI in your terminal
                 Great for SSH sessions

Enter choice [1 or 2] (default: 1):
```

**Recommendation**: Choose **1** for most users (Web UI)

### Question 2: Install or run once?

```
Installation Options

  1) Install to ~/.local/bin
     Binary will be available system-wide as 'gpu-pro'

  2) Run once without installing
     Binary will run from /tmp and be removed after

Enter choice [1 or 2] (default: 1):
```

**Recommendation**: Choose **1** to install permanently

## After Installation

### Web UI Mode
After installation, open your browser to:
```
http://localhost:8889
```

You'll see:
- Real-time GPU metrics (Linux/Windows)
- System metrics (CPU, RAM, Disk, Network) on all platforms
- Process monitoring
- Network connections

### CLI/TUI Mode
An interactive terminal interface will launch immediately.
Use arrow keys to navigate, press `q` to quit.

## Platform-Specific Information

### macOS Users
- **GPU Monitoring**: Not available (Apple GPUs use Metal, not NVML)
- **What Works**: All system metrics (CPU, Memory, Disk, Network)
- **Message**: You'll see "🍎 Running on macOS - GPU monitoring is disabled"

### Linux Users
- **GPU Monitoring**: Full support for NVIDIA GPUs
- **Requirements**: NVIDIA drivers installed
- **Detection**: Script automatically checks for `nvidia-smi`

### Windows Users
- **GPU Monitoring**: Full support for NVIDIA GPUs
- **Requirements**: NVIDIA drivers installed
- **Detection**: Script automatically checks for GPU

## Manual Installation

If the script doesn't work, you can install manually:

1. **Download the binary** for your platform from releases:
   - macOS Intel: `gpu-pro-darwin-amd64`
   - macOS Apple Silicon: `gpu-pro-darwin-arm64`
   - Linux: `gpu-pro-linux-amd64`
   - Windows: `gpu-pro-windows-amd64.exe`

2. **Make it executable** (Linux/macOS):
   ```bash
   chmod +x gpu-pro-*
   ```

3. **Run it**:
   ```bash
   ./gpu-pro-darwin-arm64  # Example for macOS Apple Silicon
   ```

## Uninstallation

```bash
# Remove the binary
rm ~/.local/bin/gpu-pro
rm ~/.local/bin/gpu-pro-cli  # If CLI was also installed

# Or if installed to /usr/local/bin
sudo rm /usr/local/bin/gpu-pro
```

## Advanced Usage

### Install Specific Version
```bash
VERSION=v1.5.0 curl -fsSL https://your-domain.com/install.sh | bash
```

### Install to Custom Directory
```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://your-domain.com/install.sh | bash
```

### Non-Interactive Installation
```bash
# Automatically choose: Web UI + Install
./install.sh <<EOF
1
1
EOF
```

## Troubleshooting

### "Download failed"
**Cause**: Can't reach GitHub or release doesn't exist
**Solution**:
- Check internet connection
- Try manual download (see above)

### "Binary not in PATH"
**Cause**: `~/.local/bin` not in your PATH
**Solution**: Add to `~/.bashrc` or `~/.zshrc`:
```bash
export PATH="$PATH:$HOME/.local/bin"
```
Then: `source ~/.bashrc`

### "No NVIDIA GPU detected"
**Cause**: No NVIDIA GPU or drivers not installed
**Solution**:
- This is normal on macOS (not supported)
- On Linux/Windows: Install NVIDIA drivers
- System metrics will still work without GPU

### Permission Denied
**Cause**: Can't write to install directory
**Solution**:
```bash
# Use sudo
sudo ./install.sh

# Or install to user directory (default)
INSTALL_DIR=~/.local/bin ./install.sh
```

## Security Note

When piping from the internet, you can review the script first:
```bash
curl -fsSL https://your-domain.com/install.sh > install.sh
less install.sh  # Review the script
chmod +x install.sh
./install.sh
```

## Support

- **Documentation**: See `INSTALL_SCRIPT.md` for detailed info
- **Issues**: https://github.com/YOUR_USERNAME/gpu-pro/issues
- **Questions**: Check existing documentation first

## What the Script Does

The installer is completely transparent and:
1. ✅ Detects your OS and architecture
2. ✅ Shows what will be installed
3. ✅ Downloads only the specific binary you need
4. ✅ Asks permission before installing
5. ✅ Shows GPU detection status
6. ✅ Provides helpful error messages
7. ✅ Cleans up after itself

**No sudo required** unless you choose to install to `/usr/local/bin`

## Examples

### Example 1: Quick Install (Default Options)
```bash
$ ./install.sh
# Press Enter twice (accepts defaults)
# - Installs Web UI
# - Installs to ~/.local/bin
# - Launches automatically
```

### Example 2: CLI Mode, Run Once
```bash
$ ./install.sh
# Choose: 2 (CLI/TUI)
# Choose: 2 (Run once)
# - Runs CLI immediately
# - No permanent installation
```

### Example 3: Web UI, Install Permanently
```bash
$ ./install.sh
# Choose: 1 (Web UI)
# Choose: 1 (Install)
# - Installs to ~/.local/bin
# - Can run 'gpu-pro' anytime
```

## Features

- ✅ **Single file**: Only need `install.sh`
- ✅ **No dependencies**: Pure bash script
- ✅ **Cross-platform**: Works on Linux, macOS, Windows
- ✅ **Interactive**: Friendly prompts, no guessing
- ✅ **Smart**: Auto-detects everything
- ✅ **Safe**: Shows what it will do before doing it
- ✅ **Colorful**: Easy to read output
- ✅ **Helpful**: Clear error messages

---

**Made with ❤️ for easy GPU and system monitoring**
