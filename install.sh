#!/bin/bash
# GPU Pro Installer Script
# Detects OS and architecture, downloads and installs the appropriate binary
# Usage: wget -q -O - https://raw.githubusercontent.com/USER/REPO/main/install.sh | bash
# Or: curl -fsSL https://raw.githubusercontent.com/USER/REPO/main/install.sh | bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# GitHub repository (update these with your actual repository details)
GITHUB_USER="${GITHUB_USER:-your-username}"
GITHUB_REPO="${GITHUB_REPO:-gpu-pro}"
GITHUB_BRANCH="${GITHUB_BRANCH:-main}"
VERSION="${VERSION:-latest}"

# Base URL for downloading binaries
BASE_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/${VERSION}"

# Installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}        GPU Pro Installer${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux*)
        OS="linux"
        ;;
    darwin*)
        OS="darwin"
        ;;
    msys*|mingw*|cygwin*|windows*)
        OS="windows"
        ;;
    *)
        echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l|armv6l)
        ARCH="arm"
        echo -e "${YELLOW}Warning: ARMv6/v7 detected. Using ARM64 binary, may not work.${NC}"
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}Detected OS:${NC} $OS"
echo -e "${BLUE}Detected Architecture:${NC} $ARCH"
echo ""

# Construct binary name
BINARY_NAME="gpu-pro-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

# Download URL
DOWNLOAD_URL="${BASE_URL}/${BINARY_NAME}"

echo -e "${BLUE}Download URL:${NC} $DOWNLOAD_URL"
echo ""

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download binary
echo -e "${YELLOW}Downloading GPU Pro...${NC}"
if command -v curl &> /dev/null; then
    curl -fsSL -o "$TMP_DIR/gpu-pro" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -q -O "$TMP_DIR/gpu-pro" "$DOWNLOAD_URL"
else
    echo -e "${RED}Error: Neither curl nor wget is available${NC}"
    exit 1
fi

if [ ! -f "$TMP_DIR/gpu-pro" ]; then
    echo -e "${RED}Error: Download failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Downloaded successfully${NC}"
echo ""

# Make binary executable
chmod +x "$TMP_DIR/gpu-pro"

# Verify binary works
echo -e "${YELLOW}Verifying binary...${NC}"
if ! "$TMP_DIR/gpu-pro" --version &> /dev/null; then
    echo -e "${YELLOW}Warning: Could not verify binary version${NC}"
    echo -e "${YELLOW}Binary may still work, continuing with installation...${NC}"
fi

# Install binary
echo -e "${YELLOW}Installing GPU Pro...${NC}"

# Check if we need sudo
if [ -w "$INSTALL_DIR" ]; then
    SUDO=""
else
    SUDO="sudo"
    echo -e "${YELLOW}Installation requires sudo privileges${NC}"
fi

# Install
if $SUDO mv "$TMP_DIR/gpu-pro" "$INSTALL_DIR/gpu-pro"; then
    echo -e "${GREEN}✓ Installed to $INSTALL_DIR/gpu-pro${NC}"
else
    # Fallback to user's local bin
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"

    if mv "$TMP_DIR/gpu-pro" "$INSTALL_DIR/gpu-pro"; then
        echo -e "${GREEN}✓ Installed to $INSTALL_DIR/gpu-pro${NC}"

        # Add to PATH if not already there
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            echo ""
            echo -e "${YELLOW}Note: $INSTALL_DIR is not in your PATH${NC}"
            echo -e "${YELLOW}Add the following to your ~/.bashrc or ~/.zshrc:${NC}"
            echo -e "${BLUE}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        fi
    else
        echo -e "${RED}Error: Installation failed${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Installation Complete!${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Run GPU Pro with:${NC}"
echo -e "  ${GREEN}gpu-pro${NC}                 - Start web dashboard (default port 8080)"
echo -e "  ${GREEN}gpu-pro --help${NC}          - Show help"
echo ""
echo -e "${BLUE}For CLI mode, use:${NC}"
echo -e "  ${GREEN}gpu-pro-cli${NC}             - Interactive terminal UI"
echo ""

# Check for GPU
echo -e "${BLUE}Checking for NVIDIA GPU...${NC}"
if command -v nvidia-smi &> /dev/null; then
    if nvidia-smi &> /dev/null; then
        echo -e "${GREEN}✓ NVIDIA GPU detected${NC}"
        GPU_COUNT=$(nvidia-smi --query-gpu=name --format=csv,noheader | wc -l)
        echo -e "${GREEN}  Found $GPU_COUNT GPU(s)${NC}"
    else
        echo -e "${YELLOW}⚠ nvidia-smi found but failed to run${NC}"
        echo -e "${YELLOW}  GPU monitoring may not be available${NC}"
        echo -e "${YELLOW}  System metrics will still work${NC}"
    fi
else
    echo -e "${YELLOW}⚠ No NVIDIA GPU detected${NC}"
    echo -e "${YELLOW}  GPU monitoring will not be available${NC}"
    echo -e "${GREEN}  System metrics (CPU, memory, disk, network) will still work${NC}"
fi

echo ""
echo -e "${BLUE}Thank you for installing GPU Pro!${NC}"
