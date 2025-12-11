# Universal Installer

This directory contains a cross-platform Go-based installer that works on **Linux**, **macOS**, and **Windows**.

## Usage

```bash
go run install.go
```

## Features

- **Cross-platform**: Works on Linux, macOS, and Windows
- **Auto-detection**: Automatically detects your operating system and architecture
- **Interactive**: Guides you through the installation process
- **PATH checking**: Verifies and helps configure your PATH environment variable
- **Multiple install options**:
  - System-wide installation
  - User-local installation
  - Skip installation (build only)

## How It Works

1. Checks if Go is installed
2. Builds the `secscan` binary for your platform
3. Offers installation options:
   - **Linux/macOS**: `/usr/local/bin` or `~/.local/bin`
   - **Windows**: `C:\Program Files\secscan` or `%USERPROFILE%\.local\bin`
4. Copies the binary to the chosen location
5. Verifies PATH configuration and provides guidance if needed

## Why Use This?

- **No platform-specific scripts needed**: One installer works everywhere
- **No dependencies**: Only requires Go (which you need anyway to build the tool)
- **Consistent experience**: Same installation process across all platforms
- **Safe**: Uses Go's standard library for reliable file operations

## Alternative Installation Methods

While this universal installer is convenient, you can also use:

- **Linux/macOS**: `../install.sh` or `make install`
- **Windows**: `../install.ps1`

See the main [INSTALL.md](../INSTALL.md) for more details.
