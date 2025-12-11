# SecScan Installation Script for Windows
# PowerShell script to build and install secscan on Windows

param(
    [string]$InstallPath = "$env:USERPROFILE\.local\bin",
    [switch]$Global = $false,
    [switch]$Help = $false
)

$VERSION = "2.2.1"
$BINARY_NAME = "secscan.exe"
$BUILD_DIR = "build"

function Show-Help {
    Write-Host ""
    Write-Host "SecScan v$VERSION - Windows Installation Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\install.ps1                    # Install to user directory (default)"
    Write-Host "  .\install.ps1 -Global            # Install to system-wide location (requires admin)"
    Write-Host "  .\install.ps1 -InstallPath PATH  # Install to custom directory"
    Write-Host "  .\install.ps1 -Help              # Show this help"
    Write-Host ""
    Write-Host "Default install location: $InstallPath" -ForegroundColor Gray
    Write-Host ""
    exit 0
}

if ($Help) {
    Show-Help
}

Write-Host ""
Write-Host "SecScan v$VERSION - Windows Installation Script" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking for Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Go not found"
    }
    Write-Host "[OK] Go found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go 1.19 or later from: https://go.dev/doc/install" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host ""

# Build the binary
Write-Host "Building $BINARY_NAME..." -ForegroundColor Yellow
try {
    # Create build directory
    if (-not (Test-Path $BUILD_DIR)) {
        New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null
    }

    # Build with version info
    $ldflags = "-X main.Version=$VERSION -s -w"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    & go build -ldflags $ldflags -o "$BUILD_DIR\$BINARY_NAME" main.go
    
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }
    
    Write-Host "[OK] Build complete: $BUILD_DIR\$BINARY_NAME" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Build failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Determine installation path
if ($Global) {
    # Check if running as administrator
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    
    if (-not $isAdmin) {
        Write-Host "[ERROR] Global installation requires administrator privileges" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please run PowerShell as Administrator, or install locally without -Global flag" -ForegroundColor Yellow
        Write-Host ""
        exit 1
    }
    
    $InstallPath = "$env:ProgramFiles\secscan"
    Write-Host "Installing to system-wide location: $InstallPath" -ForegroundColor Yellow
} else {
    Write-Host "Installing to user directory: $InstallPath" -ForegroundColor Yellow
}

# Create installation directory if it doesn't exist
if (-not (Test-Path $InstallPath)) {
    Write-Host "Creating directory: $InstallPath" -ForegroundColor Gray
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
}

# Copy binary to installation path
try {
    Copy-Item "$BUILD_DIR\$BINARY_NAME" "$InstallPath\$BINARY_NAME" -Force
    Write-Host "[OK] Installed successfully to: $InstallPath\$BINARY_NAME" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to copy binary: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Check if installation path is in PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
$isInPath = $currentPath -split ";" | Where-Object { $_ -eq $InstallPath }

if (-not $isInPath) {
    Write-Host "WARNING: $InstallPath is not in your PATH" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Do you want to add it to your PATH? (y/n): " -ForegroundColor Cyan -NoNewline
    $response = Read-Host
    
    if ($response -eq 'y' -or $response -eq 'Y') {
        try {
            $scope = if ($Global) { "Machine" } else { "User" }
            $currentPath = [Environment]::GetEnvironmentVariable("Path", $scope)
            
            if ($currentPath -notlike "*$InstallPath*") {
                $newPath = "$currentPath;$InstallPath"
                [Environment]::SetEnvironmentVariable("Path", $newPath, $scope)
                
                # Update current session
                $env:Path = "$env:Path;$InstallPath"
                
                Write-Host "[OK] Added $InstallPath to PATH" -ForegroundColor Green
                Write-Host ""
                Write-Host "WARNING: Please restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
            }
        } catch {
            Write-Host "[ERROR] Failed to update PATH: $_" -ForegroundColor Red
            Write-Host ""
            Write-Host "You can manually add the following to your PATH:" -ForegroundColor Yellow
            Write-Host "  $InstallPath" -ForegroundColor White
        }
    } else {
        Write-Host ""
        Write-Host "To add to PATH manually:" -ForegroundColor Yellow
        Write-Host "  1. Search for 'Environment Variables' in Windows" -ForegroundColor White
        Write-Host "  2. Edit the 'Path' variable" -ForegroundColor White
        Write-Host "  3. Add: $InstallPath" -ForegroundColor White
        Write-Host ""
        Write-Host "Or run from the full path:" -ForegroundColor Yellow
        Write-Host "  $InstallPath\$BINARY_NAME" -ForegroundColor White
    }
} else {
    Write-Host "[OK] $InstallPath is already in your PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Verify installation:" -ForegroundColor Cyan
Write-Host "  secscan -version" -ForegroundColor White
Write-Host "  secscan --help" -ForegroundColor White
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Cyan
Write-Host "  secscan -root C:\path\to\project" -ForegroundColor White
Write-Host ""
