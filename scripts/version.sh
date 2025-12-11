#!/usr/bin/env bash
# Version management script for SecScan
# Updates version across all files automatically

set -e

VERSION_FILE="VERSION"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to validate semantic version
validate_version() {
    if [[ ! $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        print_error "Invalid version format: $1"
        print_info "Expected format: MAJOR.MINOR.PATCH (e.g., 1.2.3 or 1.2.3-beta)"
        return 1
    fi
    return 0
}

# Function to get current version
get_current_version() {
    if [ -f "$PROJECT_ROOT/$VERSION_FILE" ]; then
        cat "$PROJECT_ROOT/$VERSION_FILE"
    else
        echo "0.0.0"
    fi
}

# Function to update version in files
update_version_in_files() {
    local new_version=$1
    local old_version=$(get_current_version)
    
    print_info "Updating version from $old_version to $new_version..."
    
    # Update VERSION file
    echo "$new_version" > "$PROJECT_ROOT/$VERSION_FILE"
    print_success "Updated VERSION file"
    
    # Update main.go
    if [ -f "$PROJECT_ROOT/main.go" ]; then
        sed -i "s/^\/\/ Version: .*/\/\/ Version: $new_version/" "$PROJECT_ROOT/main.go"
        sed -i "s/const version = \".*\"/const version = \"$new_version\"/" "$PROJECT_ROOT/main.go"
        print_success "Updated main.go"
    fi
    
    # Update mkdocs.yml
    if [ -f "$PROJECT_ROOT/mkdocs.yml" ]; then
        sed -i "s/site_name: SecScan Documentation.*/site_name: SecScan Documentation v$new_version/" "$PROJECT_ROOT/mkdocs.yml"
        print_success "Updated mkdocs.yml"
    fi
    
    # Update README.md
    if [ -f "$PROJECT_ROOT/README.md" ]; then
        sed -i "s/Version: .*/Version: $new_version/" "$PROJECT_ROOT/README.md"
        sed -i "s/secscan\/v[0-9.]*/secscan\/v$new_version/g" "$PROJECT_ROOT/README.md"
        print_success "Updated README.md"
    fi
    
    # Update install scripts
    if [ -f "$PROJECT_ROOT/install.sh" ]; then
        sed -i "s/VERSION=\".*\"/VERSION=\"$new_version\"/" "$PROJECT_ROOT/install.sh"
        print_success "Updated install.sh"
    fi
    
    if [ -f "$PROJECT_ROOT/install.ps1" ]; then
        sed -i "s/\$VERSION = \".*\"/\$VERSION = \"$new_version\"/" "$PROJECT_ROOT/install.ps1"
        print_success "Updated install.ps1"
    fi
    
    if [ -f "$PROJECT_ROOT/install.bat" ]; then
        sed -i "s/set VERSION=.*/set VERSION=$new_version/" "$PROJECT_ROOT/install.bat"
        print_success "Updated install.bat"
    fi
    
    # Update installer/install.go
    if [ -f "$PROJECT_ROOT/installer/install.go" ]; then
        sed -i "s/const version = \".*\"/const version = \"$new_version\"/" "$PROJECT_ROOT/installer/install.go"
        print_success "Updated installer/install.go"
    fi
    
    print_success "Version updated to $new_version in all files"
}

# Function to bump version
bump_version() {
    local current_version=$(get_current_version)
    local bump_type=$1
    
    IFS='.' read -r major minor patch <<< "$current_version"
    # Remove any pre-release suffix
    patch="${patch%%-*}"
    
    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid bump type: $bump_type"
            print_info "Valid types: major, minor, patch"
            return 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Main script logic
main() {
    cd "$PROJECT_ROOT"
    
    case "${1:-}" in
        get)
            # Get current version
            get_current_version
            ;;
        set)
            # Set specific version
            if [ -z "${2:-}" ]; then
                print_error "No version specified"
                print_info "Usage: $0 set <version>"
                exit 1
            fi
            
            if ! validate_version "$2"; then
                exit 1
            fi
            
            update_version_in_files "$2"
            ;;
        bump)
            # Bump version
            if [ -z "${2:-}" ]; then
                print_error "No bump type specified"
                print_info "Usage: $0 bump <major|minor|patch>"
                exit 1
            fi
            
            new_version=$(bump_version "$2")
            if [ $? -eq 0 ]; then
                update_version_in_files "$new_version"
            else
                exit 1
            fi
            ;;
        *)
            # Show help
            echo "SecScan Version Management"
            echo ""
            echo "Usage:"
            echo "  $0 get                    Get current version"
            echo "  $0 set <version>          Set specific version (e.g., 1.2.3)"
            echo "  $0 bump <type>            Bump version (major, minor, or patch)"
            echo ""
            echo "Examples:"
            echo "  $0 get                    # Show: 2.2.0"
            echo "  $0 set 3.0.0              # Set version to 3.0.0"
            echo "  $0 bump patch             # 2.2.0 -> 2.2.1"
            echo "  $0 bump minor             # 2.2.0 -> 2.3.0"
            echo "  $0 bump major             # 2.2.0 -> 3.0.0"
            exit 1
            ;;
    esac
}

main "$@"
