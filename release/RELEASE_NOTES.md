# SecScan v2.1.0 Release Notes

**Release Date:** December 8, 2025

## 🎉 What's New

### 🙈 Gitignore Support (Major Feature)

SecScan now automatically respects `.gitignore` files in your repository! This major enhancement brings:

- **Automatic Detection**: Finds and loads all `.gitignore` files in your repository hierarchy
- **Smart Filtering**: Skips build artifacts, dependencies, and generated files during scans
- **Reduced False Positives**: No more noise from binary files and vendor code
- **Faster Scans**: Skips irrelevant directories early in the walk process
- **Full Compatibility**: Supports standard gitignore patterns including:
  - Negation patterns (`!important.txt`)
  - Directory-only patterns (`logs/`)
  - Glob patterns with `**` wildcards
  - Nested `.gitignore` files

### ⚙️ New Command-Line Flag

```bash
-respect-gitignore=true|false    # Enable/disable gitignore support (default: true)
```

**Examples:**

```bash
# Default behavior - gitignore enabled
secscan -root .

# Disable for security audits
secscan -respect-gitignore=false

# Verbose mode shows what's being skipped
secscan -verbose -respect-gitignore=true
```

## 📊 Benefits

### Performance Improvements
- Faster scans by skipping gitignored directories
- Reduced memory usage by not loading irrelevant files

### Accuracy Improvements  
- Fewer false positives from binary files
- Cleaner output focused on actual source code
- Better integration with existing Git workflows

### Use Cases

**When to keep gitignore enabled (default):**
- Regular development scans
- CI/CD pipeline integration
- Quick security checks
- Code review automation

**When to disable gitignore:**
- Comprehensive security audits
- Checking if secrets exist in build artifacts
- Forensic analysis
- Debugging scan results

## 🔄 Important Behavior Note

**Git history scanning is NOT affected by gitignore!**

- **Working tree scan**: Respects `.gitignore` (skips ignored files)
- **Git history scan**: Ignores `.gitignore` (scans all commits)

This ensures you find secrets that were committed before files were added to `.gitignore`.

## 📦 Installation

### Quick Install

```bash
# Build and install globally
cd secscan
make install

# Verify installation
secscan -version
```

### Update Existing Installation

```bash
cd secscan
git pull origin main
make install
```

## 🔧 Technical Details

### Pattern Matching Algorithm
- Recursive `.gitignore` file discovery
- Proper handling of relative paths
- Pattern priority (later patterns override earlier ones)
- Support for all standard gitignore features

### Implementation
- ~370 lines of new code
- Zero external dependencies (pure Go stdlib)
- Minimal performance overhead
- Backward compatible

## 🆕 Changes Since v2.0.0

### Added
- Gitignore pattern parser
- Gitignore file collector
- Pattern matching engine with glob support
- `-respect-gitignore` command-line flag
- Gitignore status in scan output

### Changed
- `walkFiles` function now accepts `Config` parameter
- File walking logic checks gitignore before scanning
- Output shows number of loaded gitignore patterns
- Version bumped to 2.1.0

### Fixed
- Binary files properly excluded when gitignore is present
- Reduced false positives from build artifacts

## 📝 Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete version history.

## 🙏 Acknowledgments

Thanks to the community for requesting this feature!

## 📞 Support

- 🐛 [Report Issues](https://github.com/Zayan-Mohamed/secscan/issues)
- 💬 [Discussions](https://github.com/Zayan-Mohamed/secscan/discussions)
- 📧 Email: itsm.zayan@gmail.com

---

**Upgrade today and enjoy cleaner, faster secret scans!** 🚀
