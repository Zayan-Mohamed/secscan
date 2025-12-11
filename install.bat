@echo off
REM SecScan Installation Script for Windows (Batch)
REM This is a simple alternative to install.ps1 for systems where PowerShell scripts are restricted
REM For full features, use install.ps1 instead

echo.
echo SecScan v2.2.0 - Windows Installation Script (Batch)
echo ====================================================
echo.
echo Note: For more options, use install.ps1 (PowerShell)
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Go is not installed or not in PATH
    echo.
    echo Please install Go 1.19 or later from: https://go.dev/doc/install
    echo.
    pause
    exit /b 1
)

echo Go found: 
go version
echo.

REM Build the binary
echo Building secscan.exe...
if not exist build mkdir build
go build -ldflags "-X main.Version=2.2.0 -s -w" -o build\secscan.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Error: Build failed
    pause
    exit /b 1
)

echo.
echo Build complete: build\secscan.exe
echo.

REM Create local bin directory
set "INSTALL_DIR=%USERPROFILE%\.local\bin"
echo Installing to: %INSTALL_DIR%
echo.

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM Copy binary
copy /Y build\secscan.exe "%INSTALL_DIR%\secscan.exe" >nul

if %ERRORLEVEL% NEQ 0 (
    echo Error: Failed to copy binary
    pause
    exit /b 1
)

echo Installation complete!
echo.
echo Installed to: %INSTALL_DIR%\secscan.exe
echo.
echo IMPORTANT: Make sure %INSTALL_DIR% is in your PATH
echo.
echo To add to PATH:
echo   1. Search for "Environment Variables" in Windows
echo   2. Edit the "Path" variable
echo   3. Add: %INSTALL_DIR%
echo   4. Restart your terminal
echo.
echo Or use the PowerShell script for automatic PATH configuration:
echo   .\install.ps1
echo.
echo Verify installation (after adding to PATH):
echo   secscan -version
echo   secscan --help
echo.
pause
