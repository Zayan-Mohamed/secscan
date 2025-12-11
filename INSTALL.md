# Installation Guide

This guide covers different methods to install SecScan on **Linux**, **macOS**, and **Windows**.

> **⚠️ Platform-Specific Notes:**
>
> - **Linux/macOS**: Use `install.sh` or `make` commands
> - **Windows**: Use `install.ps1` (PowerShell) or the universal Go installer
> - **All Platforms**: Use the universal Go installer (`go run installer/install.go`)

---

## 📦 Quick Install by Platform

### 🐧 Linux

**Recommended Method:**

```bash
cd secscan
./install.sh
```

Or using Make:

```bash
cd secscan
make install        # System-wide (requires sudo)
# OR
make install-local  # User-only (no sudo)
```

---

### 🍎 macOS

**Recommended Method:**

```bash
cd secscan
./install.sh
```

Or using Make:

```bash
cd secscan
make install        # System-wide (requires sudo)
# OR
make install-local  # User-only (no sudo)
```

**Note:** On macOS, you might need to allow the binary in System Preferences → Security & Privacy if prompted.

---

### 🪟 Windows

**Method 1: PowerShell Script (Recommended)**

Open PowerShell (as Administrator for system-wide install) and run:

```powershell
cd secscan
.\install.ps1
```

**For system-wide installation:**

```powershell
.\install.ps1 -Global
```

**For custom installation path:**

```powershell
.\install.ps1 -InstallPath "C:\your\custom\path"
```

**Method 2: Manual Build**

```powershell
# Build the binary
go build -o secscan.exe main.go

# Move to a directory in your PATH, for example:
# System-wide: C:\Program Files\secscan\secscan.exe
# User: %USERPROFILE%\.local\bin\secscan.exe
```

**Note:** Windows doesn't come with `make` by default. Use PowerShell script or manual build instead.

---

### 🌍 Universal Installer (All Platforms)

The universal Go-based installer works on **Linux**, **macOS**, and **Windows**:

```bash
cd secscan
go run installer/install.go
```

This installer will:

1. Auto-detect your platform
2. Build the appropriate binary
3. Offer installation options specific to your OS
4. Guide you through PATH configuration

---

## 📝 Detailed Installation Methods

### Method 1: Installation Scripts

#### Linux/macOS: `install.sh`

```bash
cd secscan
./install.sh
```

The script will:

1. Build the binary
2. Ask you to choose between system-wide or local installation
3. Set up everything automatically
4. Verify your PATH configuration

#### Windows: `install.ps1`

```powershell
cd secscan
.\install.ps1
```

Options:

- Default: Installs to `%USERPROFILE%\.local\bin`
- `-Global`: Installs to `C:\Program Files\secscan` (requires admin)
- `-InstallPath PATH`: Custom installation directory

### Method 2: Using Make (Linux/macOS Only)

> **⚠️ Windows Note:** `make` is not standard on Windows. Use `install.ps1` instead.

**System-wide Installation:**

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

**Local Installation:**

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

### Method 3: Manual Installation

#### Linux/macOS

**System-wide:**

```bash
# Build the binary
cd secscan
go build -o build/secscan main.go

# Copy to /usr/local/bin (requires sudo)
sudo cp build/secscan /usr/local/bin/
sudo chmod +x /usr/local/bin/secscan
```

**Local:**

```bash
# Build the binary
cd secscan
go build -o build/secscan main.go

# Copy to local bin directory
mkdir -p ~/.local/bin
cp build/secscan ~/.local/bin/
chmod +x ~/.local/bin/secscan

# Add to PATH (if not already there)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Windows

**Using Command Prompt or PowerShell:**

```powershell
# Build the binary
cd secscan
go build -o build\secscan.exe main.go

# Copy to a directory in your PATH
# For example, create a local bin directory:
mkdir %USERPROFILE%\.local\bin
copy build\secscan.exe %USERPROFILE%\.local\bin\

# Add to PATH (PowerShell as Administrator):
$oldPath = [Environment]::GetEnvironmentVariable("Path", "User")
$newPath = "$oldPath;$env:USERPROFILE\.local\bin"
[Environment]::SetEnvironmentVariable("Path", $newPath, "User")

# Restart your terminal to apply changes
```

### Method 4: Using Go Install

If you have Go installed and the repository is public:

```bash
go install github.com/Zayan-Mohamed/secscan@latest
```

This will install to:

- **Linux/macOS**: `$GOPATH/bin` (usually `~/go/bin`)
- **Windows**: `%GOPATH%\bin` (usually `%USERPROFILE%\go\bin`)

Ensure the Go bin directory is in your PATH:

**Linux/macOS:**

```bash
export PATH="$HOME/go/bin:$PATH"
```

**Windows:**

```powershell
$env:Path += ";$env:USERPROFILE\go\bin"
```

```bash
export PATH="$HOME/go/bin:$PATH"
```

---

## 🔧 Platform-Specific Notes

### Linux

- **Recommended**: `./install.sh` or `make install`
- **Paths**: `/usr/local/bin` (system) or `~/.local/bin` (user)
- Most distributions come with `make` pre-installed

### macOS

- **Recommended**: `./install.sh` or `make install`
- **Paths**: `/usr/local/bin` (system) or `~/.local/bin` (user)
- **Security Note**: On first run, you may see a security prompt:
  1. Go to System Preferences → Security & Privacy
  2. Click "Allow Anyway" if prompted
  3. Run `secscan` again

### Windows

- **Recommended**: `.\install.ps1` (PowerShell)
- **Paths**:
  - User: `%USERPROFILE%\.local\bin` (default)
  - System: `C:\Program Files\secscan` (requires admin)
- **Important**: Windows doesn't include `make` by default
  - Use PowerShell script: `.\install.ps1`
  - Or universal installer: `go run installer/install.go`
  - Or manual build (see Method 3)

### Windows WSL (Windows Subsystem for Linux)

If using WSL, follow the Linux instructions:

```bash
./install.sh
```

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
