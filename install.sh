#!/bin/bash
# SecScan Installation Script for Linux/macOS
# This script builds and installs secscan as a global command

set -e

VERSION="2.2.1"
BINARY_NAME="secscan"
BUILD_DIR="build"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM="Linux";;
    Darwin*)    PLATFORM="macOS";;
    *)          
        echo "Unsupported operating system: ${OS}"
        echo "This script is for Linux/macOS only."
        echo ""
        echo "For Windows, please use: install.ps1"
        echo "Or use the universal installer: go run installer/install.go"
        exit 1
        ;;
esac

echo "SecScan v${VERSION} - Installation Script"
echo "Platform: ${PLATFORM}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.19 or later."
    echo "   Visit: https://go.dev/doc/install"
    exit 1
fi

echo "Go found: $(go version)"
echo ""

# Build the binary
echo "Building ${BINARY_NAME}..."
mkdir -p "${BUILD_DIR}"
go build -ldflags "-X main.Version=${VERSION} -s -w" -o "${BUILD_DIR}/${BINARY_NAME}" main.go
echo "Build complete"
echo ""

# Ask user for installation preference
echo "Choose installation method:"
echo "  1) System-wide install to /usr/local/bin (requires sudo) - RECOMMENDED"
echo "  2) Local install to ~/.local/bin (no sudo required)"
echo "  3) Skip installation (binary will be in build/ directory)"
echo ""
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        echo ""
        echo "Installing to /usr/local/bin (requires sudo)..."
        
        # Check for existing installation in ~/.local/bin
        if [ -f ~/.local/bin/${BINARY_NAME} ]; then
            echo "Warning: ${BINARY_NAME} found in ~/.local/bin"
            echo "   Removing old version from ~/.local/bin..."
            rm -f ~/.local/bin/${BINARY_NAME}
        fi
        
        sudo cp "${BUILD_DIR}/${BINARY_NAME}" /usr/local/bin/
        sudo chmod +x /usr/local/bin/${BINARY_NAME}
        echo "Installed successfully to /usr/local/bin/${BINARY_NAME}"
        echo ""
        echo "Verify installation:"
        echo "  ${BINARY_NAME} -version"
        echo "  which ${BINARY_NAME}"
        ;;
    2)
        echo ""
        echo "Installing to ~/.local/bin..."
        
        # Check for existing installation in /usr/local/bin
        if [ -f /usr/local/bin/${BINARY_NAME} ]; then
            echo "Warning: ${BINARY_NAME} found in /usr/local/bin"
            echo "   You may want to remove it with: sudo rm /usr/local/bin/${BINARY_NAME}"
            echo "   Continuing with local installation..."
            echo ""
        fi
        
        mkdir -p ~/.local/bin
        cp "${BUILD_DIR}/${BINARY_NAME}" ~/.local/bin/
        chmod +x ~/.local/bin/${BINARY_NAME}
        echo "Installed successfully to ~/.local/bin/${BINARY_NAME}"
        echo ""
        
        # Check if ~/.local/bin is in PATH
        if echo "$PATH" | grep -q "$HOME/.local/bin"; then
            echo "~/.local/bin is already in your PATH"
            echo ""
            echo "Verify installation:"
            echo "  ${BINARY_NAME} -version"
        else
            echo "~/.local/bin is NOT in your PATH"
            echo ""
            echo "Add this line to your ~/.bashrc or ~/.zshrc:"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
            echo ""
            echo "Then run:"
            echo "  source ~/.bashrc  # or ~/.zshrc"
            echo ""
            echo "After that, verify installation:"
            echo "  ${BINARY_NAME} -version"
        fi
        ;;
    3)
        echo ""
        echo "Build complete. Binary is available at:"
        echo "  ${BUILD_DIR}/${BINARY_NAME}"
        echo ""
        echo "Run with:"
        echo "  ./${BUILD_DIR}/${BINARY_NAME}"
        ;;
    *)
        echo ""
        echo "Invalid choice. Binary is available at: ${BUILD_DIR}/${BINARY_NAME}"
        exit 1
        ;;
esac

echo ""
echo "Installation complete!"
echo ""
echo "Quick start:"
echo "  ${BINARY_NAME} -root /path/to/project"
echo "  ${BINARY_NAME} --help"
echo ""
