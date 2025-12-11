# Project Structure

## Directory Organization

```
secscan/
├── .github/                    # GitHub-specific files
│   └── workflows/             # GitHub Actions workflows
│       ├── ci.yml            # Continuous Integration
│       └── release.yml       # Release automation
│
├── assets/                    # Project assets
│   └── logo.png              # Project logo
│
├── build/                     # Compiled binaries (git-ignored)
│   ├── secscan-linux-amd64
│   ├── secscan-darwin-amd64
│   ├── secscan-darwin-arm64
│   └── secscan-windows-amd64.exe
│
├── docs/                      # Documentation source (MkDocs)
│   ├── index.md              # Documentation home
│   ├── release-guide.md      # Release process guide
│   ├── getting-started/      # Getting started guides
│   │   ├── installation.md
│   │   ├── quickstart.md
│   │   └── first-scan.md
│   ├── user-guide/           # User documentation
│   │   ├── basic-usage.md
│   │   ├── configuration.md
│   │   ├── ci-cd-integration.md
│   │   └── examples.md
│   └── reference/            # Reference documentation
│       └── cli-options.md
│
├── installer/                 # Universal Go-based installer
│   ├── install.go            # Cross-platform installer
│   └── README.md             # Installer documentation
│
├── releases/                  # Release notes (organized by version)
│   ├── RELEASE_v2.2.0.md
│   └── RELEASE_v2.3.0.md
│
├── scripts/                   # Automation scripts
│   ├── version.sh            # Version management
│   ├── release.sh            # Release automation
│   └── install-curl.sh       # Remote installation script
│
├── .gitignore                # Git ignore rules
├── .secscan.toml.example     # Example configuration
├── CHANGELOG.md              # Project changelog
├── CROSS_PLATFORM_INSTALLATION.md  # Cross-platform install guide
├── EXAMPLES.md               # Usage examples
├── go.mod                    # Go module definition
├── install.bat               # Windows batch installer
├── install.ps1               # Windows PowerShell installer
├── install.sh                # Unix-like installer
├── INSTALL.md                # Installation guide
├── INSTALLATION_SCRIPTS.md   # Installation scripts overview
├── LICENSE                   # Project license (MIT)
├── main.go                   # Main application source
├── Makefile                  # Build automation
├── mkdocs.yml                # MkDocs configuration
├── QUICK_INSTALL.md          # Quick installation reference
├── README.md                 # Project README
├── requirements-docs.txt     # Python deps for docs
├── RELEASE_NOTES.md          # General release notes
└── VERSION                   # Current version number
```

---

## File Descriptions

### Root Files

#### `VERSION`

Contains the current version number in semantic versioning format:

```
2.2.0
```

#### `main.go`

The main application source code containing:

- Version constant (auto-updated by scripts)
- All scanner logic
- Detection patterns
- CLI interface

#### `go.mod`

Go module definition. SecScan has **zero external dependencies** and uses only the Go standard library.

#### `Makefile`

Build automation with targets:

- `make build` - Build for current platform
- `make build-all` - Build for all platforms
- `make install` - Install locally
- `make test` - Run tests
- `make clean` - Clean build artifacts

#### `mkdocs.yml`

MkDocs configuration for documentation site at https://zayan-mohamed.github.io/secscan

---

### Installation Scripts

Located in project root for easy access:

#### `install.sh`

Universal installation script for **Linux and macOS**:

```bash
bash install.sh
```

#### `install.ps1`

PowerShell installation script for **Windows**:

```powershell
./install.ps1
```

#### `install.bat`

Batch file installer for **Windows**:

```cmd
install.bat
```

---

### Scripts Directory

#### `scripts/version.sh`

Version management automation:

```bash
./scripts/version.sh get          # Show current version
./scripts/version.sh set 2.3.0    # Set specific version
./scripts/version.sh bump patch   # Bump patch version
```

Updates version in all files:

- `VERSION`
- `main.go`
- `mkdocs.yml`
- `README.md`
- All installation scripts

#### `scripts/release.sh`

Complete release automation:

```bash
./scripts/release.sh 2.3.0
```

Automates:

1. Version update
2. Testing
3. Building
4. Changelog update
5. Git tagging
6. GitHub release
7. Documentation deployment

#### `scripts/install-curl.sh`

Remote installation script for curl-based installation:

```bash
curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-curl.sh | bash
```

---

### Installer Directory

#### `installer/install.go`

Universal Go-based installer that works on all platforms:

```bash
cd installer
go run install.go
```

Provides:

- Cross-platform installation
- Binary detection
- PATH management
- User-friendly interface

---

### Documentation

#### `docs/`

MkDocs-based documentation source:

**Getting Started:**

- `getting-started/installation.md` - Installation instructions
- `getting-started/quickstart.md` - Quick start guide
- `getting-started/first-scan.md` - First scan tutorial

**User Guide:**

- `user-guide/basic-usage.md` - Basic usage guide
- `user-guide/configuration.md` - Configuration reference
- `user-guide/ci-cd-integration.md` - CI/CD integration
- `user-guide/examples.md` - Usage examples

**Reference:**

- `reference/cli-options.md` - CLI options reference

**Development:**

- `release-guide.md` - Release process guide

---

### GitHub Actions

#### `.github/workflows/ci.yml`

Continuous Integration workflow:

- Runs on push and pull requests
- Tests on Linux, macOS, Windows
- Tests with Go 1.19, 1.20, 1.21
- Runs linters
- Uploads coverage reports

#### `.github/workflows/release.yml`

Release automation workflow:

- Triggers on tag push
- Builds for all platforms
- Generates checksums
- Creates GitHub release
- Deploys documentation

---

### Releases Directory

#### `releases/`

Contains detailed release notes for each version:

- `RELEASE_v2.2.0.md`
- `RELEASE_v2.3.0.md`
- etc.

Each release note includes:

- What's new
- Bug fixes
- Installation instructions
- Breaking changes
- Migration guide

---

### Build Directory

#### `build/`

**Git-ignored** directory containing compiled binaries:

```
build/
├── secscan-linux-amd64
├── secscan-darwin-amd64
├── secscan-darwin-arm64
└── secscan-windows-amd64.exe
```

Generated by `make build-all`.

---

## File Organization Best Practices

### What Goes Where?

**Root Directory:**

- Core project files (main.go, go.mod, etc.)
- Installation scripts (for easy access)
- Documentation markdown files
- Configuration files

**`scripts/` Directory:**

- Development and release automation
- Version management
- Remote installation scripts

**`docs/` Directory:**

- All MkDocs documentation
- Organized by topic
- Rendered to GitHub Pages

**`releases/` Directory:**

- Detailed release notes
- One file per version
- Used by release automation

**`installer/` Directory:**

- Universal installer code
- Installer-specific documentation

**`.github/` Directory:**

- GitHub-specific files only
- Workflows, issue templates, etc.

**`build/` Directory:**

- Compiled binaries (git-ignored)
- Generated by build process

---

## Maintenance Guidelines

### Adding New Files

1. **Scripts:** Add to `scripts/` directory
2. **Documentation:** Add to appropriate `docs/` subdirectory
3. **Build artifacts:** Ensure added to `.gitignore`
4. **Installation scripts:** Keep in root for accessibility

### Updating Documentation

1. Edit files in `docs/` directory
2. Test locally: `mkdocs serve`
3. Preview at http://127.0.0.1:8000
4. Deploy: `mkdocs gh-deploy`

### Managing Releases

1. Use `scripts/version.sh` to update version
2. Create release notes in `releases/` directory
3. Use `scripts/release.sh` for full automation
4. Or push tag to trigger GitHub Actions

---

## Version Control

### .gitignore Highlights

```gitignore
# Binaries
build/
*.exe
secscan

# Go build cache
*.test
*.out

# IDE files
.vscode/
.idea/
*.swp

# OS files
.DS_Store
Thumbs.db

# Documentation build
site/
```

### What to Commit

✅ **DO commit:**

- Source code (main.go)
- Scripts (scripts/\*)
- Documentation source (docs/\*)
- Configuration examples
- Installation scripts
- VERSION file
- Makefile

❌ **DON'T commit:**

- Compiled binaries
- Build artifacts
- Documentation build output (site/)
- IDE-specific files
- OS-specific files

---

## Quick Reference

### Build Commands

```bash
make build          # Build for current platform
make build-all      # Build all platforms
make install        # Install locally
make test           # Run tests
make clean          # Clean artifacts
```

### Documentation Commands

```bash
mkdocs serve        # Serve docs locally
mkdocs build        # Build docs
mkdocs gh-deploy    # Deploy to GitHub Pages
```

### Version Commands

```bash
./scripts/version.sh get         # Get version
./scripts/version.sh set 2.3.0   # Set version
./scripts/version.sh bump patch  # Bump version
```

### Release Commands

```bash
./scripts/release.sh 2.3.0                # Full release
./scripts/release.sh 2.3.0 --skip-tests   # Skip tests
./scripts/release.sh 2.3.0 --skip-docs    # Skip docs
```

---

**Last Updated**: December 2025
