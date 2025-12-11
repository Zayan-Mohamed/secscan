#!/usr/bin/env bash
# Release script for SecScan
# Automates the entire release process

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Functions for colored output
print_header() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if git repo is clean
check_git_status() {
    print_info "Checking git status..."
    
    if ! git diff-index --quiet HEAD --; then
        print_error "Git working directory is not clean"
        print_info "Please commit or stash your changes first"
        exit 1
    fi
    
    print_success "Git working directory is clean"
}

# Run tests
run_tests() {
    print_info "Running tests..."
    
    if [ -f "$PROJECT_ROOT/Makefile" ]; then
        make test || {
            print_error "Tests failed"
            exit 1
        }
    else
        go test ./... || {
            print_error "Tests failed"
            exit 1
        }
    fi
    
    print_success "All tests passed"
}

# Build for all platforms
build_all() {
    print_info "Building for all platforms..."
    
    cd "$PROJECT_ROOT"
    make build-all || {
        print_error "Build failed"
        exit 1
    }
    
    print_success "Built for all platforms"
}

# Create release notes
create_release_notes() {
    local version=$1
    local release_file="$PROJECT_ROOT/releases/RELEASE_v${version}.md"
    
    print_info "Creating release notes..."
    
    mkdir -p "$PROJECT_ROOT/releases"
    
    cat > "$release_file" << EOF
# SecScan v${version} Release Notes

**Release Date:** $(date +"%B %d, %Y")

## 🚀 What's New

<!-- Add your release notes here -->

### Features
- 

### Improvements
- 

### Bug Fixes
- 

### Documentation
- Updated installation documentation
- Added cross-platform installation support

---

## 📦 Installation

### Quick Install (Recommended)

**Linux/macOS:**
\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.sh | bash
\`\`\`

**Windows (PowerShell):**
\`\`\`powershell
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.ps1 | iex
\`\`\`

### Using Go
\`\`\`bash
go install github.com/Zayan-Mohamed/secscan@v${version}
\`\`\`

### Manual Download
Download the appropriate binary for your platform from the [releases page](https://github.com/Zayan-Mohamed/secscan/releases/tag/v${version}).

---

## 🔧 Supported Platforms

- ✅ Linux (amd64, arm64)
- ✅ macOS (amd64, arm64)
- ✅ Windows (amd64)

---

## 📊 Checksums

\`\`\`
<!-- Checksums will be added during GitHub release -->
\`\`\`

---

## 🐛 Known Issues

None at this time.

---

## 📝 Full Changelog

See [CHANGELOG.md](../CHANGELOG.md) for the complete changelog.

EOF

    print_success "Created release notes template: $release_file"
    print_info "Please edit the release notes before continuing"
    
    # Open in editor if available
    if command -v $EDITOR &> /dev/null; then
        $EDITOR "$release_file"
    elif command -v nano &> /dev/null; then
        nano "$release_file"
    elif command -v vim &> /dev/null; then
        vim "$release_file"
    fi
}

# Update changelog
update_changelog() {
    local version=$1
    local changelog="$PROJECT_ROOT/CHANGELOG.md"
    
    print_info "Updating CHANGELOG.md..."
    
    # Backup changelog
    cp "$changelog" "$changelog.bak"
    
    # Create temp file with new entry
    cat > "$changelog.tmp" << EOF
# Changelog

All notable changes to SecScan will be documented in this file.

## [${version}] - $(date +"%Y-%m-%d")

### Added
- 

### Changed
- 

### Fixed
- 

EOF
    
    # Append old changelog (skip header)
    tail -n +3 "$changelog.bak" >> "$changelog.tmp"
    mv "$changelog.tmp" "$changelog"
    rm "$changelog.bak"
    
    print_success "Updated CHANGELOG.md"
    print_warning "Please edit CHANGELOG.md to add details about this release"
}

# Create git tag
create_git_tag() {
    local version=$1
    
    print_info "Creating git tag v${version}..."
    
    git add -A
    git commit -m "Release v${version}" || print_warning "No changes to commit"
    git tag -a "v${version}" -m "Release v${version}"
    
    print_success "Created git tag v${version}"
}

# Push to GitHub
push_to_github() {
    local version=$1
    
    print_info "Pushing to GitHub..."
    
    git push origin main
    git push origin "v${version}"
    
    print_success "Pushed to GitHub"
}

# Create GitHub release
create_github_release() {
    local version=$1
    
    print_info "Creating GitHub release..."
    
    if ! command -v gh &> /dev/null; then
        print_warning "GitHub CLI (gh) not found"
        print_info "Please create the release manually at:"
        print_info "https://github.com/Zayan-Mohamed/secscan/releases/new?tag=v${version}"
        return
    fi
    
    # Create release with binaries
    gh release create "v${version}" \
        --title "SecScan v${version}" \
        --notes-file "releases/RELEASE_v${version}.md" \
        build/secscan-linux-amd64#secscan-linux-amd64 \
        build/secscan-darwin-amd64#secscan-darwin-amd64 \
        build/secscan-darwin-arm64#secscan-darwin-arm64 \
        build/secscan-windows-amd64.exe#secscan-windows-amd64.exe \
        || {
            print_error "Failed to create GitHub release"
            print_info "Please create it manually"
            return 1
        }
    
    print_success "Created GitHub release"
}

# Deploy documentation
deploy_docs() {
    print_info "Deploying documentation..."
    
    cd "$PROJECT_ROOT"
    
    if [ ! -f "mkdocs.yml" ]; then
        print_warning "mkdocs.yml not found, skipping documentation deployment"
        return
    fi
    
    # Install dependencies if needed
    if ! command -v mkdocs &> /dev/null; then
        print_info "Installing mkdocs..."
        pip install -r requirements-docs.txt || {
            print_warning "Failed to install mkdocs dependencies"
            return
        }
    fi
    
    # Deploy to GitHub Pages
    mkdocs gh-deploy --force || {
        print_warning "Failed to deploy documentation"
        return
    }
    
    print_success "Deployed documentation to GitHub Pages"
}

# Main release workflow
main() {
    cd "$PROJECT_ROOT"
    
    print_header "SecScan Release Process"
    
    # Get version
    local version="${1:-}"
    if [ -z "$version" ]; then
        print_error "No version specified"
        echo ""
        echo "Usage: $0 <version> [--skip-tests] [--skip-docs]"
        echo ""
        echo "Examples:"
        echo "  $0 2.3.0                    # Full release"
        echo "  $0 2.3.0 --skip-tests       # Skip tests"
        echo "  $0 2.3.0 --skip-docs        # Skip documentation deployment"
        exit 1
    fi
    
    # Parse flags
    local skip_tests=false
    local skip_docs=false
    
    for arg in "$@"; do
        case $arg in
            --skip-tests)
                skip_tests=true
                ;;
            --skip-docs)
                skip_docs=true
                ;;
        esac
    done
    
    print_info "Releasing version: $version"
    echo ""
    
    # Confirmation
    read -p "Continue with release? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Release cancelled"
        exit 0
    fi
    
    # Step 1: Check git status
    print_header "Step 1: Check Git Status"
    check_git_status
    
    # Step 2: Update version
    print_header "Step 2: Update Version"
    bash "$SCRIPT_DIR/version.sh" set "$version"
    
    # Step 3: Run tests
    if [ "$skip_tests" = false ]; then
        print_header "Step 3: Run Tests"
        run_tests
    else
        print_warning "Skipping tests (--skip-tests flag)"
    fi
    
    # Step 4: Build
    print_header "Step 4: Build All Platforms"
    build_all
    
    # Step 5: Update changelog
    print_header "Step 5: Update Changelog"
    update_changelog "$version"
    
    # Step 6: Create release notes
    print_header "Step 6: Create Release Notes"
    create_release_notes "$version"
    
    # Step 7: Create git tag
    print_header "Step 7: Create Git Tag"
    create_git_tag "$version"
    
    # Step 8: Push to GitHub
    print_header "Step 8: Push to GitHub"
    push_to_github "$version"
    
    # Step 9: Create GitHub release
    print_header "Step 9: Create GitHub Release"
    create_github_release "$version"
    
    # Step 10: Deploy documentation
    if [ "$skip_docs" = false ]; then
        print_header "Step 10: Deploy Documentation"
        deploy_docs
    else
        print_warning "Skipping documentation deployment (--skip-docs flag)"
    fi
    
    # Success!
    print_header "🎉 Release Complete!"
    echo ""
    print_success "SecScan v${version} has been released!"
    echo ""
    print_info "Next steps:"
    echo "  • Verify the release at: https://github.com/Zayan-Mohamed/secscan/releases/tag/v${version}"
    echo "  • Check documentation at: https://zayan-mohamed.github.io/secscan"
    echo "  • Test installation: curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/install.sh | bash"
    echo ""
}

main "$@"
