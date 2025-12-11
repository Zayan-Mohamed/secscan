# SecScan Remote Installation Script for Windows
# Usage: irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-windows.ps1 | iex

param(
    [string]$Version = "latest",
    [string]$InstallPath = "$env:USERPROFILE\.local\bin",
    [switch]$Global = $false
)

$ErrorActionPreference = "Stop"

# Configuration
$REPO = "Zayan-Mohamed/secscan"
$BINARY_NAME = "secscan.exe"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White",
        [string]$Prefix = ""
    )
    
    if ($Prefix) {
        Write-Host "$Prefix " -NoNewline -ForegroundColor $Color
        Write-Host $Message
    } else {
        Write-Host $Message -ForegroundColor $Color
    }
}

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "====================================================" -ForegroundColor Cyan
    Write-Host "  $Message" -ForegroundColor Cyan
    Write-Host "====================================================" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput -Message $Message -Color Green -Prefix "[OK]"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput -Message $Message -Color Red -Prefix "[ERROR]"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput -Message $Message -Color Blue -Prefix "[INFO]"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput -Message $Message -Color Yellow -Prefix "[WARNING]"
}

function Get-LatestVersion {
    try {
        Write-Info "Fetching latest version from GitHub..."
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -ErrorAction Stop
        return $release.tag_name.TrimStart('v')
    } catch {
        Write-Error "Failed to fetch latest version: $_"
        Write-Info "Falling back to version 2.2.1"
        return "2.2.1"
    }
}

function Get-Architecture {
    $arch = [System.Environment]::Is64BitOperatingSystem
    if ($arch) {
        return "amd64"
    } else {
        Write-Error "32-bit Windows is not supported"
        Write-Info "Please download manually from: https://github.com/$REPO/releases"
        exit 1
    }
}

function Download-Binary {
    param(
        [string]$Version,
        [string]$Arch,
        [string]$TempDir
    )
    
    $binaryUrl = "https://github.com/$REPO/releases/download/v$Version/secscan-windows-$Arch.exe"
    $outputPath = Join-Path $TempDir $BINARY_NAME
    
    Write-Info "Downloading SecScan v$Version for Windows $Arch..."
    Write-Host "      URL: $binaryUrl" -ForegroundColor Gray
    
    try {
        # Use WebClient for better progress indication
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($binaryUrl, $outputPath)
        Write-Success "Download complete"
        return $outputPath
    } catch {
        Write-Error "Failed to download binary: $_"
        Write-Info "Please download manually from: https://github.com/$REPO/releases"
        exit 1
    }
}

function Install-Binary {
    param(
        [string]$SourcePath,
        [string]$InstallPath
    )
    
    # Create installation directory if it doesn't exist
    if (-not (Test-Path $InstallPath)) {
        Write-Info "Creating directory: $InstallPath"
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    $targetPath = Join-Path $InstallPath $BINARY_NAME
    
    # Copy binary
    try {
        Copy-Item $SourcePath $targetPath -Force
        Write-Success "Installed to: $targetPath"
        return $targetPath
    } catch {
        Write-Error "Failed to install binary: $_"
        exit 1
    }
}

function Add-ToPath {
    param(
        [string]$Path,
        [string]$Scope
    )
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", $Scope)
    
    if ($currentPath -notlike "*$Path*") {
        $newPath = "$currentPath;$Path"
        [Environment]::SetEnvironmentVariable("Path", $newPath, $Scope)
        
        # Update current session
        $env:Path = "$env:Path;$Path"
        
        Write-Success "Added $Path to PATH"
        Write-Warning "Please restart your terminal for PATH changes to take effect"
        return $true
    }
    return $false
}

# Main installation flow
try {
    Write-Header "SecScan Installation Script for Windows"
    
    # Detect architecture
    $arch = Get-Architecture
    Write-Info "Detected architecture: $arch"
    
    # Get version
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
    }
    Write-Info "Installing version: $Version"
    Write-Host ""
    
    # Determine installation path
    if ($Global) {
        # Check if running as administrator
        $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
        
        if (-not $isAdmin) {
            Write-Error "Global installation requires administrator privileges"
            Write-Info "Please run PowerShell as Administrator, or install locally without -Global flag"
            exit 1
        }
        
        $InstallPath = "$env:ProgramFiles\secscan"
        $pathScope = "Machine"
        Write-Info "Installing to system-wide location: $InstallPath"
    } else {
        $pathScope = "User"
        Write-Info "Installing to user directory: $InstallPath"
    }
    Write-Host ""
    
    # Create temporary directory
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) "secscan-install"
    if (Test-Path $tempDir) {
        Remove-Item $tempDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    # Download binary
    $binaryPath = Download-Binary -Version $Version -Arch $arch -TempDir $tempDir
    Write-Host ""
    
    # Install binary
    $installedPath = Install-Binary -SourcePath $binaryPath -InstallPath $InstallPath
    Write-Host ""
    
    # Check PATH and offer to add
    $currentPath = [Environment]::GetEnvironmentVariable("Path", $pathScope)
    $isInPath = $currentPath -split ";" | Where-Object { $_ -eq $InstallPath }
    
    if (-not $isInPath) {
        Write-Warning "$InstallPath is not in your PATH"
        Write-Host ""
        Write-Host "Do you want to add it to your PATH? (y/n): " -ForegroundColor Cyan -NoNewline
        $response = Read-Host
        Write-Host ""
        
        if ($response -eq 'y' -or $response -eq 'Y') {
            Add-ToPath -Path $InstallPath -Scope $pathScope
        } else {
            Write-Info "To add to PATH manually:"
            Write-Host "  1. Search for 'Environment Variables' in Windows" -ForegroundColor White
            Write-Host "  2. Edit the 'Path' variable" -ForegroundColor White
            Write-Host "  3. Add: $InstallPath" -ForegroundColor White
            Write-Host ""
            Write-Info "Or run from the full path: $installedPath"
        }
    } else {
        Write-Success "$InstallPath is already in your PATH"
    }
    
    # Cleanup
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    
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
    Write-Host "Documentation: https://github.com/$REPO" -ForegroundColor Gray
    Write-Host ""
    
} catch {
    Write-Host ""
    Write-Error "Installation failed: $_"
    Write-Info "Please report this issue at: https://github.com/$REPO/issues"
    exit 1
}
