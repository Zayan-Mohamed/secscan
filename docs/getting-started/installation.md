# Installation Guide

This guide covers all methods to install SecScan on your system.

## Prerequisites

- Go 1.19 or higher (for building from source)
- Git (for cloning the repository)
- Linux, macOS, or WSL2 on Windows

## Installation Methods

### Method 1: Quick Install (Recommended)

Use the installation script for an interactive setup:

```bash
cd secscan
./install.sh
```

The script will:

1. Build the binary
2. Ask you to choose between system-wide or local installation
3. Set up everything automatically
4. Verify the installation

### Method 2: Make Install (System-wide)

Install globally to `/usr/local/bin` (requires sudo):

```bash
cd secscan
make install
```

Verify installation:

```bash
secscan -version
which secscan
```

### Method 3: Make Install (Local - No sudo)

Install to `~/.local/bin` without sudo:

```bash
cd secscan
make install-local
```

If `~/.local/bin` is not in your PATH, add this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc  # or ~/.zshrc for zsh users
```

### Method 4: Go Install

If the repository is public and you have Go installed:

```bash
go install github.com/Zayan-Mohamed/secscan@latest
```

This installs to `$GOPATH/bin` (usually `~/go/bin`).

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$GOPATH/bin:$PATH"
```

### Method 5: Manual Build

For complete control over the build process:

```bash
# Clone the repository
git clone https://github.com/Zayan-Mohamed/secscan.git
cd secscan

# Build the binary
make build

# Binary will be in build/secscan
./build/secscan -version
```

#### System-wide Manual Installation

```bash
# After building
sudo cp build/secscan /usr/local/bin/
sudo chmod +x /usr/local/bin/secscan
```

#### Local Manual Installation

```bash
# After building
mkdir -p ~/.local/bin
cp build/secscan ~/.local/bin/
chmod +x ~/.local/bin/secscan
```

## Verify Installation

After installation, verify that SecScan is properly installed:

```bash
# Check version
secscan -version

# Check location
which secscan

# Run a test scan
secscan -root /tmp -history=false
```

## Uninstalling

### System-wide Installation

```bash
sudo rm /usr/local/bin/secscan
```

### Local Installation

```bash
rm ~/.local/bin/secscan
```

### Go Install

```bash
rm $GOPATH/bin/secscan
```

## Troubleshooting

### Permission Denied

If you get "permission denied" when trying to install system-wide:

- Use `make install-local` instead
- Or use `sudo make install`

### Command Not Found

If `secscan` is not found after installation:

1. Check if the binary exists:

   ```bash
   ls -l ~/.local/bin/secscan
   # or
   ls -l /usr/local/bin/secscan
   ```

2. Ensure the directory is in your PATH:

   ```bash
   echo $PATH
   ```

3. Add to PATH if needed:
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```

### Build Fails

If the build fails:

1. Check Go version:
   ```bash
   go version
   ```
2. Ensure Go 1.19 or higher is installed

3. Try cleaning and rebuilding:
   ```bash
   make clean
   make build
   ```

## Next Steps

- 🚀 [Quick Start Guide](quickstart.md)
- 🎯 [Run Your First Scan](first-scan.md)
- 🔧 [Configuration Options](../user-guide/configuration.md)
