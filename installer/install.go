package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "2.2.2"

var (
	binaryName   = "secscan"
	buildDir     = "build"
	installPaths = map[string][]string{
		"linux":   {"/usr/local/bin", filepath.Join(os.Getenv("HOME"), ".local", "bin")},
		"darwin":  {"/usr/local/bin", filepath.Join(os.Getenv("HOME"), ".local", "bin")},
		"windows": {filepath.Join(os.Getenv("ProgramFiles"), "secscan"), filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin")},
	}
)

func main() {
	fmt.Println()
	fmt.Println("SecScan v" + version + " - Universal Installation Script")
	fmt.Println("=================================================")
	fmt.Println()
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// Check if Go is installed
	fmt.Println("Checking for Go installation...")
	if !checkGoInstalled() {
		fmt.Println("✗ Go is not installed or not in PATH")
		fmt.Println()
		fmt.Println("Please install Go 1.19 or later from: https://go.dev/doc/install")
		fmt.Println()
		os.Exit(1)
	}
	fmt.Println("✓ Go found")
	fmt.Println()

	// Build the binary
	binaryPath, err := buildBinary()
	if err != nil {
		fmt.Printf("✗ Build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Build complete: %s\n", binaryPath)
	fmt.Println()

	// Install the binary
	if err := installBinary(binaryPath); err != nil {
		fmt.Printf("✗ Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ Installation complete!")
	fmt.Println()
	fmt.Println("Verify installation:")
	fmt.Printf("  %s -version\n", binaryName)
	fmt.Printf("  %s --help\n", binaryName)
	fmt.Println()
	fmt.Println("Quick start:")
	if runtime.GOOS == "windows" {
		fmt.Printf("  %s -root C:\\path\\to\\project\n", binaryName)
	} else {
		fmt.Printf("  %s -root /path/to/project\n", binaryName)
	}
	fmt.Println()
}

func checkGoInstalled() bool {
	cmd := exec.Command("go", "version")
	return cmd.Run() == nil
}

func buildBinary() (string, error) {
	fmt.Println("Building binary...")

	// Create build directory
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	// Determine binary name with extension
	binaryFile := binaryName
	if runtime.GOOS == "windows" {
		binaryFile += ".exe"
	}

	outputPath := filepath.Join(buildDir, binaryFile)

	// Build command
	ldflags := fmt.Sprintf("-X main.Version=%s -s -w", version)
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build command failed: %w", err)
	}

	return outputPath, nil
}

func installBinary(binaryPath string) error {
	paths := installPaths[runtime.GOOS]
	if paths == nil {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	fmt.Println("Choose installation method:")
	for i, path := range paths {
		label := "Local"
		requiresSudo := false
		if i == 0 {
			label = "System-wide"
			if runtime.GOOS != "windows" {
				requiresSudo = true
			}
		}
		fmt.Printf("  %d) %s install to %s", i+1, label, path)
		if requiresSudo {
			fmt.Print(" (requires sudo/admin)")
		}
		fmt.Println()
	}
	fmt.Printf("  %d) Skip installation (binary will be in %s/)\n", len(paths)+1, buildDir)
	fmt.Println()
	fmt.Print("Enter choice [1-" + fmt.Sprintf("%d", len(paths)+1) + "]: ")

	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(paths)+1 {
		return fmt.Errorf("invalid choice: %d", choice)
	}

	// Skip installation
	if choice == len(paths)+1 {
		fmt.Println()
		fmt.Printf("Build complete. Binary is available at: %s\n", binaryPath)
		return nil
	}

	installPath := paths[choice-1]

	// Create install directory if it doesn't exist
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Determine target binary name
	targetBinary := filepath.Join(installPath, filepath.Base(binaryPath))

	// Copy binary
	fmt.Println()
	fmt.Printf("Installing to: %s\n", targetBinary)

	// For system-wide install on Unix, use sudo
	if choice == 1 && runtime.GOOS != "windows" {
		cmd := exec.Command("sudo", "cp", binaryPath, targetBinary)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy binary with sudo: %w", err)
		}

		cmd = exec.Command("sudo", "chmod", "+x", targetBinary)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	} else {
		// Local install or Windows
		input, err := os.ReadFile(binaryPath)
		if err != nil {
			return fmt.Errorf("failed to read binary: %w", err)
		}

		if err := os.WriteFile(targetBinary, input, 0755); err != nil {
			return fmt.Errorf("failed to write binary: %w", err)
		}
	}

	fmt.Printf("✓ Installed successfully to: %s\n", targetBinary)

	// Check if install path is in PATH
	checkPathEnvironment(installPath)

	return nil
}

func checkPathEnvironment(installPath string) {
	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	inPath := false
	for _, p := range paths {
		if strings.TrimSpace(p) == installPath {
			inPath = true
			break
		}
	}

	if !inPath {
		fmt.Println()
		fmt.Printf("⚠ %s is not in your PATH\n", installPath)
		fmt.Println()

		if runtime.GOOS == "windows" {
			fmt.Println("To add to PATH:")
			fmt.Println("  1. Search for 'Environment Variables' in Windows")
			fmt.Println("  2. Edit the 'Path' variable")
			fmt.Printf("  3. Add: %s\n", installPath)
		} else {
			shell := os.Getenv("SHELL")
			rcFile := ".bashrc"
			if strings.Contains(shell, "zsh") {
				rcFile = ".zshrc"
			}

			fmt.Printf("Add this line to your ~/%s:\n", rcFile)
			fmt.Printf("  export PATH=\"%s:$PATH\"\n", installPath)
			fmt.Println()
			fmt.Println("Then run:")
			fmt.Printf("  source ~/%s\n", rcFile)
		}
	} else {
		fmt.Println()
		fmt.Printf("✓ %s is already in your PATH\n", installPath)
	}
}
