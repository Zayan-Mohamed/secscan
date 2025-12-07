# Installation Guide

This guide covers different methods to install SecScan as a proper command on your system.

## Quick Install (Recommended)

### Method 1: Using the Installation Script

The easiest way to install:

```bash
cd secscan
./install.sh
```

The script will:

1. Build the binary
2. Ask you to choose between system-wide or local installation
3. Set up everything automatically

### Method 2: Using Make (System-wide)

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

### Method 3: Using Make (Local - No sudo)

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

## Manual Installation

### System-wide Installation

```bash
# Build the binary
cd secscan
make build

# Copy to /usr/local/bin (requires sudo)
sudo cp build/secscan /usr/local/bin/
sudo chmod +x /usr/local/bin/secscan
```

### Local Installation

```bash
# Build the binary
cd secscan
make build

# Copy to local bin directory
mkdir -p ~/.local/bin
cp build/secscan ~/.local/bin/
chmod +x ~/.local/bin/secscan

# Add to PATH (if not already there)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Using Go Install

If you have Go installed and the repository is public:

```bash
go install github.com/Zayan-Mohamed/secscan@latest
```

This will install to `$GOPATH/bin` (usually `~/go/bin`).

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Platform-Specific Instructions

### Linux

All methods above work on Linux. Recommended: `make install`

### macOS

All methods above work on macOS. Recommended: `make install`

On macOS, you might need to allow the binary to run:

1. Run `secscan` once
2. Go to System Preferences → Security & Privacy
3. Click "Allow Anyway" if prompted

### Windows (WSL)

Use the Linux instructions within WSL.

### Windows (Native)

Build on Windows:

```powershell
go build -o secscan.exe main.go
```

Then move `secscan.exe` to a directory in your PATH, or add the current directory to PATH.

## Verification

After installation, verify that `secscan` is working:

```bash
# Check version
secscan -version

# Find installation location
which secscan

# Run a quick test
secscan --help
```

## Troubleshooting

### Command not found

If you get "command not found", check your PATH:

```bash
echo $PATH
```

Make sure the installation directory is included.

### Permission denied

If you get "permission denied", make the binary executable:

```bash
chmod +x /path/to/secscan
```

### Cannot execute binary file

If the binary won't execute, you may need to rebuild for your platform:

```bash
cd secscan
make build
```

## Uninstallation

### If installed with make install

```bash
cd secscan
make uninstall
```

### Manual uninstall

```bash
# System-wide
sudo rm /usr/local/bin/secscan

# Local
rm ~/.local/bin/secscan
```

## Next Steps

Once installed, check out:

- [Quick Start Guide](QUICKSTART.md) - Get started in 2 minutes
- [Examples](EXAMPLES.md) - Real-world usage examples
- [README](README.md) - Full documentation

Run your first scan:

```bash
secscan -root /path/to/your/project
```
