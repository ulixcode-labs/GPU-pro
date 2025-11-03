#!/bin/bash
# GPU Pro - Launcher Script
# This script automatically detects your OS and architecture, downloads the correct binary
# to the current directory, and runs it.
#
# Usage:
#   ./install.sh

set -e

# Configuration
VERSION="${VERSION:-v2.0.0}"
GITHUB_REPO="${GITHUB_REPO:-ulixcode-labs/GPU-pro}"
# For GitHub releases, use: https://github.com/${GITHUB_REPO}/releases/download/${VERSION}
# For raw files from main branch:
BASE_URL="https://raw.githubusercontent.com/${GITHUB_REPO}/main/dist"

# For development/testing with local files
# Uncomment and set to your local dist directory for testing

# No flavor suffix - we only have production binaries now
FLAVOR=""

echo "================================================================"
echo "                    GPU Pro - Launcher"
echo "================================================================"
echo ""

# Function to print messages
print_info() {
    echo "[INFO] $1"
}

print_success() {
    echo "[SUCCESS] $1"
}

print_error() {
    echo "[ERROR] $1"
}

print_warning() {
    echo "[WARNING] $1"
}

print_step() {
    echo ">> $1"
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)     OS='linux' ;;
        darwin*)    OS='darwin' ;;
        msys*|mingw*|cygwin*|windows*) OS='windows' ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)   ARCH='amd64' ;;
        arm64|aarch64)  ARCH='arm64' ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
}

# Display platform information
show_platform_info() {
    print_step "Platform Detection"
    echo ""
    echo "  OS:           $OS"
    echo "  Architecture: $ARCH"
    echo ""

    case "$OS" in
        linux)
            echo "  GPU Support:  Full (NVML)"
            echo "                NVIDIA GPU monitoring via NVML"
            ;;
        windows)
            echo "  GPU Support:  Full (NVML)"
            echo "                NVIDIA GPU monitoring via NVML"
            ;;
        darwin)
            echo "  GPU Support:  Disabled"
            echo "                macOS does not support NVML"
            echo "                System metrics (CPU, RAM, Disk, Network) still available"
            ;;
    esac
    echo ""
}

# Ask user which mode to run
ask_mode() {
    print_step "Select Mode"
    echo ""
    echo "  1) Web UI    - Dashboard accessible via browser"
    echo "                 http://localhost:8889"
    echo "                 Full featured web interface"
    echo ""
    echo "  2) CLI/TUI   - Terminal-based interface"
    echo "                 Interactive text UI in your terminal"
    echo "                 Great for SSH sessions"
    echo ""

    while true; do
        echo -n "Enter choice [1, 2, or q to quit] (default: 1): "
        read MODE_CHOICE
        MODE_CHOICE=${MODE_CHOICE:-1}

        case "$MODE_CHOICE" in
            1)
                MODE="web"
                BINARY_TYPE=""
                print_info "Selected: Web UI mode"
                break
                ;;
            2)
                MODE="cli"
                BINARY_TYPE="-cli"
                print_info "Selected: CLI/TUI mode"
                break
                ;;
            q|Q|quit|QUIT)
                echo ""
                print_info "Installation cancelled. Goodbye!"
                echo ""
                exit 0
                ;;
            *)
                print_error "Invalid choice. Please enter 1, 2, or q."
                ;;
        esac
    done
    echo ""
}

# Construct binary name (no flavor suffix - we only have production binaries)
get_binary_name() {
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="gpu-pro${BINARY_TYPE}-${OS}-${ARCH}.exe"
    else
        BINARY_NAME="gpu-pro${BINARY_TYPE}-${OS}-${ARCH}"
    fi
    LOCAL_BINARY="gpu-pro${BINARY_TYPE}"
}

# Download binary
download_binary() {
    print_step "Downloading GPU Pro"
    echo ""

    # Check if running in local development mode
    if [ -n "$LOCAL_DIST" ] && [ -d "$LOCAL_DIST" ]; then
        print_warning "Development mode: Using local binary"
        print_info "Source: $LOCAL_DIST/$BINARY_NAME"

        if [ -f "$LOCAL_DIST/$BINARY_NAME" ]; then
            cp "$LOCAL_DIST/$BINARY_NAME" "./$LOCAL_BINARY"
            chmod +x "./$LOCAL_BINARY"
            print_success "Binary ready in current directory"
            return 0
        else
            print_error "Local binary not found: $LOCAL_DIST/$BINARY_NAME"
            echo ""
            echo "Available files:"
            ls -lh "$LOCAL_DIST/" 2>/dev/null || echo "  (directory not accessible)"
            exit 1
        fi
    fi

    # Download from GitHub releases
    DOWNLOAD_URL="${BASE_URL}/${BINARY_NAME}"
    print_info "URL: $DOWNLOAD_URL"
    echo ""

    # Try curl first, then wget
    if command -v curl &> /dev/null; then
        if curl -fSL --progress-bar "$DOWNLOAD_URL" -o "./$LOCAL_BINARY"; then
            chmod +x "./$LOCAL_BINARY"
            print_success "Download complete - binary saved to ./$LOCAL_BINARY"
        else
            print_error "Download failed"
            show_download_help
            exit 1
        fi
    elif command -v wget &> /dev/null; then
        if wget --show-progress -q "$DOWNLOAD_URL" -O "./$LOCAL_BINARY"; then
            chmod +x "./$LOCAL_BINARY"
            print_success "Download complete - binary saved to ./$LOCAL_BINARY"
        else
            print_error "Download failed"
            show_download_help
            exit 1
        fi
    else
        print_error "Neither curl nor wget found"
        echo ""
        echo "Please install one of them:"
        echo "  macOS:   brew install curl"
        echo "  Ubuntu:  sudo apt-get install curl"
        echo "  CentOS:  sudo yum install curl"
        exit 1
    fi
    echo ""
}

# Show download troubleshooting help
show_download_help() {
    echo ""
    print_warning "Download failed. Please check:"
    echo "  1. Internet connection is working"
    echo "  2. GitHub is accessible: https://github.com"
    echo "  3. Release exists: $DOWNLOAD_URL"
    echo ""
    echo "Manual download:"
    echo "  1. Visit: https://github.com/$GITHUB_REPO"
    echo "  2. Download: $BINARY_NAME"
    echo "  3. chmod +x $BINARY_NAME"
    echo "  4. ./$BINARY_NAME"
}

# Ask if user wants to install or just run
ask_install_or_run() {
    print_step "Installation Options"
    echo ""
    echo "  1) Install to $INSTALL_DIR"
    echo "     Binary will be available system-wide as 'gpu-pro${BINARY_TYPE}'"
    echo ""
    echo "  2) Run once without installing"
    echo "     Binary will run from /tmp and be removed after"
    echo ""

    while true; do
        echo -n "Enter choice [1, 2, or q to quit] (default: 1): "
        read INSTALL_CHOICE
        INSTALL_CHOICE=${INSTALL_CHOICE:-1}

        case "$INSTALL_CHOICE" in
            1)
                SHOULD_INSTALL=true
                print_info "Will install to: $INSTALL_DIR/${LOCAL_BINARY}"
                break
                ;;
            2)
                SHOULD_INSTALL=false
                print_info "Will run once from temporary directory"
                break
                ;;
            q|Q|quit|QUIT)
                echo ""
                print_info "Installation cancelled. Goodbye!"
                echo ""
                exit 0
                ;;
            *)
                print_error "Invalid choice. Please enter 1, 2, or q."
                ;;
        esac
    done
    echo ""
}

# Install binary to system
install_binary() {
    print_step "Installing GPU Pro"
    echo ""

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Check if we can write to install directory
    if [ -w "$INSTALL_DIR" ]; then
        if mv "/tmp/$LOCAL_BINARY" "$INSTALL_DIR/$LOCAL_BINARY"; then
            print_success "Installed to: $INSTALL_DIR/$LOCAL_BINARY"
        else
            print_error "Installation failed"
            exit 1
        fi
    else
        print_warning "Need elevated permissions for installation"
        if sudo mv "/tmp/$LOCAL_BINARY" "$INSTALL_DIR/$LOCAL_BINARY"; then
            print_success "Installed to: $INSTALL_DIR/$LOCAL_BINARY"
        else
            print_error "Installation failed even with sudo"
            exit 1
        fi
    fi

    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        print_warning "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "  Add this line to your ~/.bashrc or ~/.zshrc:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "  Then run: source ~/.bashrc  (or source ~/.zshrc)"
        echo ""
    fi
    echo ""
}

# Show GPU detection info
show_gpu_info() {
    if [ "$OS" = "darwin" ]; then
        return  # Skip GPU check on macOS
    fi

    print_step "GPU Detection"
    echo ""

    if command -v nvidia-smi &> /dev/null; then
        if nvidia-smi &> /dev/null; then
            GPU_COUNT=$(nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | wc -l | tr -d ' ')
            print_success "Found $GPU_COUNT NVIDIA GPU(s)"

            # Show GPU names
            nvidia-smi --query-gpu=name,driver_version --format=csv,noheader 2>/dev/null | while IFS=',' read -r name driver; do
                echo "  • $name (Driver: $driver)"
            done
        else
            print_warning "nvidia-smi found but not working"
            echo "  System metrics will still be available"
        fi
    else
        print_warning "No NVIDIA GPU detected"
        echo "  System metrics (CPU, RAM, Disk, Network) will still work"
    fi
    echo ""
}

# Run the binary
run_binary() {
    show_gpu_info

    print_step "Starting GPU Pro"
    echo ""

    # Determine which binary to run
    if [ -f "./$LOCAL_BINARY" ]; then
        BINARY_PATH="./$LOCAL_BINARY"
    else
        print_error "Binary not found!"
        exit 1
    fi

    # Display startup information based on mode
    if [ "$MODE" = "web" ]; then
        echo "============================================================="
        echo "  GPU Pro Web UI"
        echo "============================================================="
        echo ""
        echo "  Dashboard: http://localhost:8889"
        echo ""
        echo "  Press Ctrl+\\ to stop the server"
        echo "  Press Ctrl+C to interrupt"
        echo ""
        echo "============================================================="
        echo ""

        # Run the binary directly
        exec "$BINARY_PATH"
    else
        echo "============================================================="
        echo "  GPU Pro CLI/TUI"
        echo "============================================================="
        echo ""
        echo "  Press Ctrl+C to exit"
        echo ""
        echo "============================================================="
        echo ""

        # Run the binary directly for CLI mode
        exec "$BINARY_PATH"
    fi
}

# Clean up temporary files (not needed anymore as we keep binary in current dir)
cleanup() {
    :  # No-op - we keep the binary in current directory for user to run
}

# Show usage
show_usage() {
    cat << EOF
GPU Pro Launcher

Usage:
  ./install.sh

This script will:
  1. Detect your OS and architecture
  2. Ask which mode you want (Web UI or CLI/TUI)
  3. Download the appropriate binary to the current directory
  4. Run it directly

Environment Variables:
  VERSION        Release version to download (default: v2.0.0)
  LOCAL_DIST     Use local binaries for testing (no download)

Examples:
  # Download and run specific version
  VERSION=v1.5.0 ./install.sh

  # Development mode with local binaries
  LOCAL_DIST=./dist ./install.sh

Platform Support:
  Linux   - Full GPU support via NVML
  Windows - Full GPU support via NVML
  macOS   - System metrics only (GPU disabled)

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo ""
            show_usage
            exit 1
            ;;
    esac
done

# Main installation flow
main() {
    # Detect platform
    detect_os
    detect_arch
    show_platform_info

    # Ask user preferences
    ask_mode

    # Get binary name (production binaries only, no flavor variants)
    get_binary_name

    # Download binary to current directory
    download_binary

    # Run the binary
    run_binary
}

# Handle Ctrl+C gracefully
trap cleanup EXIT

# Run main if script is executed (not sourced)
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main
fi
