#!/usr/bin/env bash
# Remote installation script for SecScan
# Usage: curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-curl.sh | bash

set -e

# Version to install (will be updated by release script)
VERSION="${SECSCAN_VERSION:-latest}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os="$(uname -s)"
    local arch="$(uname -m)"
    
    case "$os" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            print_info "Please download manually from: https://github.com/Zayan-Mohamed/secscan/releases"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            print_info "Please download manually from: https://github.com/Zayan-Mohamed/secscan/releases"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        print_info "Fetching latest version..."
        VERSION=$(curl -fsSL https://api.github.com/repos/Zayan-Mohamed/secscan/releases/latest | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
        
        if [ -z "$VERSION" ]; then
            print_error "Failed to fetch latest version"
            exit 1
        fi
        
        print_success "Latest version: $VERSION"
    fi
}

# Download binary
download_binary() {
    local binary_name="secscan-${PLATFORM}"
    local download_url="https://github.com/Zayan-Mohamed/secscan/releases/download/v${VERSION}/${binary_name}"
    local temp_file="/tmp/secscan-${PLATFORM}"
    
    print_info "Downloading SecScan v${VERSION} for ${PLATFORM}..."
    
    if command -v curl &> /dev/null; then
        curl -fsSL -o "$temp_file" "$download_url" || {
            print_error "Failed to download binary"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q -O "$temp_file" "$download_url" || {
            print_error "Failed to download binary"
            exit 1
        }
    else
        print_error "Neither curl nor wget found"
        print_info "Please install curl or wget and try again"
        exit 1
    fi
    
    print_success "Downloaded binary"
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file=$1
    local install_dir=""
    
    # Determine installation directory
    if [ -w "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
    elif [ -w "$HOME/.local/bin" ]; then
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
    else
        print_warning "No writable installation directory found"
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
        print_info "Installing to $install_dir (may require sudo)"
    fi
    
    print_info "Installing to $install_dir..."
    
    # Copy and make executable
    if [ -w "$install_dir" ]; then
        cp "$temp_file" "$install_dir/secscan"
        chmod +x "$install_dir/secscan"
    else
        sudo cp "$temp_file" "$install_dir/secscan"
        sudo chmod +x "$install_dir/secscan"
    fi
    
    # Cleanup
    rm -f "$temp_file"
    
    print_success "Installed to $install_dir/secscan"
    
    # Check if directory is in PATH
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        print_warning "$install_dir is not in your PATH"
        print_info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$install_dir\""
        echo ""
    fi
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    if command -v secscan &> /dev/null; then
        local installed_version=$(secscan --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -n1)
        print_success "SecScan v${installed_version} installed successfully!"
        return 0
    else
        print_error "Installation verification failed"
        print_info "Try running: hash -r"
        print_info "Or restart your terminal"
        return 1
    fi
}

# Main installation flow
main() {
    print_header "SecScan Installer"
    
    detect_platform
    print_info "Detected platform: $PLATFORM"
    
    get_latest_version
    
    temp_file=$(download_binary)
    
    install_binary "$temp_file"
    
    echo ""
    print_header "Installation Complete!"
    
    if verify_installation; then
        echo ""
        print_info "Get started with:"
        echo ""
        echo "    secscan -root .                    # Scan current directory"
        echo "    secscan -root . -git               # Scan with git history"
        echo "    secscan -root . -json report.json  # Generate JSON report"
        echo ""
        print_info "For more information, visit:"
        echo "    https://zayan-mohamed.github.io/secscan"
        echo ""
    fi
}

main "$@"
