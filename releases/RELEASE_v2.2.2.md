# SecScan v2.2.2 Release Notes

**Release Date:** December 12, 2025

# Installation Script Updates - Summary

## Issue

The Windows installation command was failing when run remotely:

```powershell
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.ps1 | iex
```

**Error:** `no required module provides package main.go: go.mod file not found`

## Root Cause

The root `install.ps1` script was designed for **local installation** (building from source), but was being used for **remote installation** (downloading from GitHub). When executed remotely, the script runs without access to the source code files (`main.go`, `go.mod`).

## Solution

### 1. Created New Remote Installation Script

**File:** `scripts/install-windows.ps1`

- Downloads pre-built binaries from GitHub releases
- Does not require Go or source code
- Matches the functionality of `scripts/install-curl.sh` for Linux/macOS
- Supports version selection: `latest` or specific version
- Includes proper error handling and user feedback

### 2. Updated Root Installation Script

**File:** `install.ps1`

- Added clear documentation header explaining it's for **local installation**
- Fixed Unicode character issues (replaced emoji with ASCII text)
- Added reference to remote installation script
- Now displays: `[OK]`, `[ERROR]`, `WARNING:` instead of emoji

### 3. Updated Documentation

Updated the following files to use the correct remote installation path:

- `README.md` - Main installation instructions
- `INSTALL.md` - Comprehensive installation guide
- `releases/RELEASE_v2.2.1.md` - Release notes

## Usage

### Remote Installation (No Source Required)

```powershell
# Downloads pre-built binary from GitHub releases
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-windows.ps1 | iex
```

### Local Installation (Requires Source Code)

```powershell
# Clone repository first
git clone https://github.com/Zayan-Mohamed/secscan.git
cd secscan

# Build and install from source
.\install.ps1
```

## Features of New Remote Script

1. **Binary Download**: Fetches pre-built executables from GitHub releases
2. **Version Support**: Install latest or specific version
3. **Architecture Detection**: Automatically detects Windows architecture
4. **PATH Management**: Offers to add installation directory to PATH
5. **Admin Support**: Supports both user and system-wide installation
6. **Error Handling**: Comprehensive error messages and fallback options
7. **Progress Feedback**: Clear status messages throughout installation

## Files Changed

1. **Created**: `scripts/install-windows.ps1` (new remote installer)
2. **Updated**: `install.ps1` (fixed Unicode issues, added docs)
3. **Updated**: `README.md` (corrected installation URL)
4. **Updated**: `INSTALL.md` (added remote installation section)
5. **Updated**: `releases/RELEASE_v2.2.1.md` (corrected installation URL)

## Testing

To test the new remote installation:

```powershell
# Test remote installation
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-windows.ps1 | iex

# Verify installation
secscan -version
secscan --help
```

## Benefits

1. Users can install without cloning the repository
2. No Go compiler required for end users
3. Faster installation (downloads binary vs. building)
4. Consistent with Linux/macOS installation experience
5. Proper separation of local vs. remote installation methods
6. Fixed PowerShell parsing errors

---

## 📦 Installation

### Quick Install (Recommended)

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.ps1 | iex
```

### Using Go
```bash
go install github.com/Zayan-Mohamed/secscan@v2.2.2
```

### Manual Download
Download the appropriate binary for your platform from the [releases page](https://github.com/Zayan-Mohamed/secscan/releases/tag/v2.2.2).

---

## 🔧 Supported Platforms

- ✅ Linux (amd64, arm64)
- ✅ macOS (amd64, arm64)
- ✅ Windows (amd64)

---

## 📊 Checksums

```
<!-- Checksums will be added during GitHub release -->
```

---

## 🐛 Known Issues

None at this time.

---

## 📝 Full Changelog

See [CHANGELOG.md](../CHANGELOG.md) for the complete changelog.

