#!/bin/bash

# GPU Pro - Quick Start Script
# This script makes it easy to build and run GPU Pro with a single command

set -e

# Banner
print_banner() {
    echo ""
    echo "=================================================="
    echo "           GPU Pro - Quick Start"
    echo "        Master Your AI Workflow"
    echo "=================================================="
    echo ""
}

# Print messages
info() {
    echo "[INFO] $1"
}

success() {
    echo "[SUCCESS] $1"
}

error() {
    echo "[ERROR] $1"
}

warning() {
    echo "[WARNING] $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        error "Go is not installed!"
        echo ""
        echo "Please install Go 1.24+ from: https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    success "Go ${GO_VERSION} detected"
}

# Build Web UI
build_webui() {
    info "Building Web UI version..."
    if go build -ldflags="-s -w" -o gpu-pro; then
        success "Web UI build successful: gpu-pro"
        return 0
    else
        error "Web UI build failed"
        return 1
    fi
}

# Build Terminal UI
build_tui() {
    info "Building Terminal UI version..."
    if go build -ldflags="-s -w" -o gpu-pro-cli ./cmd/gpu-pro-cli; then
        success "Terminal UI build successful: gpu-pro-cli"
        return 0
    else
        error "Terminal UI build failed"
        return 1
    fi
}

# Check if binaries exist
check_binaries() {
    WEBUI_EXISTS=false
    TUI_EXISTS=false

    if [ -f "gpu-pro" ]; then
        WEBUI_EXISTS=true
    fi

    if [ -f "gpu-pro-cli" ]; then
        TUI_EXISTS=true
    fi
}

# Run Web UI
run_webui() {
    echo ""
    info "Starting GPU Pro Web UI..."
    echo ""
    echo "=================================================="
    echo "  Web Dashboard: http://localhost:8889"
    echo "=================================================="
    echo ""
    info "Press Ctrl+C to stop the server"
    echo ""

    ./gpu-pro
}

# Run Terminal UI
run_tui() {
    echo ""
    info "Starting GPU Pro Terminal UI..."
    echo ""

    ./gpu-pro-cli
}

# Show menu
show_menu() {
    echo ""
    echo "Choose an option:"
    echo ""
    echo "  1) Web UI      - Beautiful web dashboard (http://localhost:8889)"
    echo "  2) Terminal UI - Elegant TUI for SSH/terminal sessions"
    echo "  3) Build both  - Just build, don't run"
    echo "  4) Clean       - Remove built binaries"
    echo "  q) Quit"
    echo ""
    echo -n "Enter your choice [1-4/q]: "
}

# Clean binaries
clean_binaries() {
    info "Cleaning built binaries..."
    rm -f gpu-pro gpu-pro-cli
    success "Clean complete"
}

# Main function
main() {
    clear
    print_banner

    # Check Go installation
    check_go
    echo ""

    # Parse command line arguments for direct execution
    if [ "$1" = "web" ] || [ "$1" = "webui" ] || [ "$1" = "1" ]; then
        check_binaries
        if [ "$WEBUI_EXISTS" = false ]; then
            build_webui || exit 1
        else
            success "Using existing binary: gpu-pro"
        fi
        run_webui
        exit 0
    elif [ "$1" = "tui" ] || [ "$1" = "cli" ] || [ "$1" = "2" ]; then
        check_binaries
        if [ "$TUI_EXISTS" = false ]; then
            build_tui || exit 1
        else
            success "Using existing binary: gpu-pro-cli"
        fi
        run_tui
        exit 0
    elif [ "$1" = "build" ] || [ "$1" = "3" ]; then
        build_webui
        build_tui
        echo ""
        success "Build complete! Use './start.sh' to run."
        exit 0
    elif [ "$1" = "clean" ] || [ "$1" = "4" ]; then
        clean_binaries
        exit 0
    elif [ "$1" = "quit" ] || [ "$1" = "q" ] || [ "$1" = "Q" ]; then
        echo ""
        info "Goodbye!"
        echo ""
        exit 0
    fi

    # Check existing binaries
    check_binaries

    if [ "$WEBUI_EXISTS" = true ]; then
        success "Found existing Web UI binary"
    fi

    if [ "$TUI_EXISTS" = true ]; then
        success "Found existing Terminal UI binary"
    fi

    # Interactive menu
    while true; do
        show_menu
        read -r choice

        case $choice in
            1)
                if [ "$WEBUI_EXISTS" = false ]; then
                    echo ""
                    build_webui || continue
                fi
                run_webui
                break
                ;;
            2)
                if [ "$TUI_EXISTS" = false ]; then
                    echo ""
                    build_tui || continue
                fi
                run_tui
                break
                ;;
            3)
                echo ""
                build_webui
                echo ""
                build_tui
                echo ""
                success "Build complete! Run './start.sh' again to start."
                echo ""
                exit 0
                ;;
            4)
                echo ""
                clean_binaries
                echo ""
                exit 0
                ;;
            q|Q)
                echo ""
                info "Goodbye!"
                echo ""
                exit 0
                ;;
            *)
                echo ""
                error "Invalid choice. Please enter 1, 2, 3, 4, or q"
                ;;
        esac
    done
}

# Handle Ctrl+C gracefully
trap 'echo ""; echo "[WARNING] Interrupted"; echo ""; exit 130' INT

# Run main function
main "$@"
