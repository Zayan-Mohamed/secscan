# Release Guide for SecScan

This document provides a comprehensive guide for releasing new versions of SecScan.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Release](#quick-release)
- [Manual Release Process](#manual-release-process)
- [Version Management](#version-management)
- [Testing Before Release](#testing-before-release)
- [Post-Release Checklist](#post-release-checklist)

---

## Prerequisites

### Required Tools

- **Git**: For version control and tagging
- **Go 1.19+**: For building binaries
- **Make**: For running build tasks
- **GitHub CLI (gh)**: Optional, for automated releases
- **Python 3.x**: For documentation deployment
- **MkDocs**: For documentation (install via `pip install -r requirements-docs.txt`)

### Access Required

- Push access to the main branch
- Permission to create releases on GitHub
- Access to deploy GitHub Pages (if deploying docs)

---

## Quick Release

The fastest way to release a new version:

```bash
# Automated release (recommended)
./scripts/release.sh 2.3.0

# Skip tests if already run
./scripts/release.sh 2.3.0 --skip-tests

# Skip documentation deployment
./scripts/release.sh 2.3.0 --skip-docs
```

This script will:

1. ✅ Check git status
2. ✅ Update version across all files
3. ✅ Run tests
4. ✅ Build for all platforms
5. ✅ Update changelog
6. ✅ Create release notes
7. ✅ Create git tag
8. ✅ Push to GitHub
9. ✅ Create GitHub release with binaries
10. ✅ Deploy documentation

---

## Manual Release Process

If you prefer more control, follow these steps:

### 1. Prepare for Release

```bash
# Ensure your working directory is clean
git status

# Make sure you're on the main branch
git checkout main
git pull origin main

# Run tests
make test

# Build for all platforms
make build-all
```

### 2. Update Version

```bash
# Update version to 2.3.0
./scripts/version.sh set 2.3.0

# Or bump version automatically
./scripts/version.sh bump patch   # 2.2.0 -> 2.2.1
./scripts/version.sh bump minor   # 2.2.0 -> 2.3.0
./scripts/version.sh bump major   # 2.2.0 -> 3.0.0
```

This updates version in:

- `VERSION` file
- `main.go`
- `mkdocs.yml`
- `README.md`
- All installation scripts

### 3. Update Documentation

```bash
# Edit CHANGELOG.md
nano CHANGELOG.md

# Create release notes
mkdir -p releases
nano releases/RELEASE_v2.3.0.md
```

### 4. Commit and Tag

```bash
# Stage all changes
git add -A

# Commit
git commit -m "Release v2.3.0"

# Create annotated tag
git tag -a v2.3.0 -m "Release v2.3.0"

# Push changes and tags
git push origin main
git push origin v2.3.0
```

### 5. Create GitHub Release

**Option A: Using GitHub CLI**

```bash
gh release create v2.3.0 \
  --title "SecScan v2.3.0" \
  --notes-file releases/RELEASE_v2.3.0.md \
  build/secscan-linux-amd64#secscan-linux-amd64 \
  build/secscan-darwin-amd64#secscan-darwin-amd64 \
  build/secscan-darwin-arm64#secscan-darwin-arm64 \
  build/secscan-windows-amd64.exe#secscan-windows-amd64.exe
```

**Option B: Using GitHub Web Interface**

1. Go to https://github.com/Zayan-Mohamed/secscan/releases/new
2. Select tag: `v2.3.0`
3. Set title: `SecScan v2.3.0`
4. Copy-paste release notes from `releases/RELEASE_v2.3.0.md`
5. Upload binaries from `build/` directory
6. Click "Publish release"

### 6. Deploy Documentation

```bash
# Install dependencies if not already installed
pip install -r requirements-docs.txt

# Deploy to GitHub Pages
mkdocs gh-deploy --force
```

Docs will be available at: https://zayan-mohamed.github.io/secscan

---

## Version Management

### Version File Structure

The `VERSION` file contains the current version in semantic versioning format:

```
2.2.0
```

### Version Script Usage

```bash
# Get current version
./scripts/version.sh get

# Set specific version
./scripts/version.sh set 3.0.0

# Bump version
./scripts/version.sh bump major    # X.0.0
./scripts/version.sh bump minor    # x.X.0
./scripts/version.sh bump patch    # x.x.X
```

### Semantic Versioning Guidelines

- **Major version (X.0.0)**: Breaking changes, incompatible API changes
- **Minor version (x.X.0)**: New features, backward-compatible
- **Patch version (x.x.X)**: Bug fixes, backward-compatible

**Examples:**

- `2.2.0 → 2.2.1`: Bug fix
- `2.2.0 → 2.3.0`: New feature
- `2.2.0 → 3.0.0`: Breaking change

---

## Testing Before Release

### Run Comprehensive Tests

```bash
# Unit tests
go test -v ./...

# Integration tests
make test

# Test on multiple platforms (if available)
GOOS=linux GOARCH=amd64 go build -o test-linux
GOOS=darwin GOARCH=amd64 go build -o test-darwin
GOOS=windows GOARCH=amd64 go build -o test-windows.exe
```

### Test Installation Scripts

```bash
# Test local installation script
bash install.sh

# Test PowerShell script (on Windows)
pwsh install.ps1

# Test universal Go installer
cd installer
go run install.go
```

### Verify Binaries

```bash
# Check binary works
./build/secscan-linux-amd64 --version

# Run quick scan
./build/secscan-linux-amd64 -root . -verbose
```

---

## Post-Release Checklist

After releasing, verify the following:

### Immediate Verification

- [ ] GitHub release created successfully
- [ ] All binaries uploaded and downloadable
- [ ] Release notes are correct and complete
- [ ] Documentation deployed to GitHub Pages
- [ ] Checksums generated and included

### Test Installation Methods

```bash
# Test curl installation (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-curl.sh | bash

# Verify version
secscan --version

# Test basic functionality
secscan -root . -quiet
```

### Update External Resources

- [ ] Update README badges if needed
- [ ] Update examples in documentation
- [ ] Announce on social media/blog (optional)
- [ ] Update package managers (if applicable)

### Monitor for Issues

- [ ] Watch GitHub issues for new bug reports
- [ ] Check GitHub Actions for any failures
- [ ] Verify documentation is rendering correctly
- [ ] Test installation on different platforms

---

## Automated Release with GitHub Actions

SecScan uses GitHub Actions for automated releases. When you push a tag:

1. **CI Workflow** runs tests on multiple platforms
2. **Release Workflow** triggers on tag push:
   - Builds binaries for all platforms
   - Generates checksums
   - Creates GitHub release
   - Deploys documentation

### Trigger Automated Release

```bash
# Create and push tag
git tag -a v2.3.0 -m "Release v2.3.0"
git push origin v2.3.0

# GitHub Actions will automatically:
# - Build binaries
# - Create release
# - Deploy docs
```

Monitor progress at: https://github.com/Zayan-Mohamed/secscan/actions

---

## Troubleshooting

### Build Failures

```bash
# Clean build directory
rm -rf build/*

# Rebuild
make build-all

# Check for errors
go build -v ./...
```

### Tag Already Exists

```bash
# Delete local tag
git tag -d v2.3.0

# Delete remote tag
git push origin :refs/tags/v2.3.0

# Recreate tag
git tag -a v2.3.0 -m "Release v2.3.0"
git push origin v2.3.0
```

### Documentation Deployment Fails

```bash
# Reinstall dependencies
pip install --upgrade -r requirements-docs.txt

# Test locally
mkdocs serve

# Deploy manually
mkdocs gh-deploy --force
```

### GitHub CLI Issues

```bash
# Login to GitHub CLI
gh auth login

# Check authentication
gh auth status

# Create release manually via web interface if gh fails
```

---

## Quick Reference

### Common Commands

```bash
# Get version
./scripts/version.sh get

# Bump patch version
./scripts/version.sh bump patch

# Full release
./scripts/release.sh 2.3.0

# Skip tests
./scripts/release.sh 2.3.0 --skip-tests

# Build all platforms
make build-all

# Deploy docs only
mkdocs gh-deploy --force
```

### Important URLs

- **Releases**: https://github.com/Zayan-Mohamed/secscan/releases
- **Documentation**: https://zayan-mohamed.github.io/secscan
- **Issues**: https://github.com/Zayan-Mohamed/secscan/issues
- **Actions**: https://github.com/Zayan-Mohamed/secscan/actions

---

## Need Help?

If you encounter issues during the release process:

1. Check this guide thoroughly
2. Review GitHub Actions logs
3. Check [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines
4. Open an issue on GitHub
5. Contact: itsm.zayan@gmail.com

---

**Last Updated**: December 2025
**Maintainer**: Zayan Mohamed
