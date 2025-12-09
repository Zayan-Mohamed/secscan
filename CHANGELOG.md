# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.2.0] - 2025-12-09

### Fixed

- 🔒 **Git History Scanning**: Fixed critical bug where lock files (package-lock.json, pnpm-lock.yaml, yarn.lock, etc.) were still being scanned in git history
  - Added file name extraction from git diff headers
  - Implemented skip file check before processing diff content
  - Now properly respects `shouldSkipFile()` function in git history scans
- 📋 **Enhanced Skip Patterns**: Improved file skip detection with comprehensive lock file patterns
  - Added support for Cargo.lock, Gemfile.lock, poetry.lock, composer.lock, pubspec.lock
  - Better coverage for package manager lock files across all ecosystems

### Changed

- **Git Diff Processing**: Modified to track current filename from diff headers and skip content accordingly
- **Performance**: Reduced false positives and improved scan speed by properly skipping lock files in git history

## [2.1.0] - 2025-12-08

### Added

- 🙈 **Gitignore Support**: Automatically respects `.gitignore` files when scanning
  - Finds and loads all `.gitignore` files in the repository hierarchy
  - Supports nested `.gitignore` files in subdirectories
  - Handles negation patterns (`!important.txt`)
  - Supports directory-only patterns (`logs/`)
  - Compatible with standard gitignore glob patterns including `**` wildcards
  - Reduces false positives from build artifacts and vendor code
  - Speeds up scans by skipping irrelevant files
- 🎚️ **Gitignore Control Flag**: New `-respect-gitignore` flag (default: `true`)
  - Enable/disable gitignore handling as needed
  - Useful for security audits where scanning all files is required
  - Shows number of loaded patterns in verbose output

### Changed

- **File Walking**: Modified to check gitignore patterns before scanning files and directories
- **Output**: Added gitignore status to scan initialization output
- **Performance**: Faster scans by skipping gitignored directories early in the walk process

### Fixed

- Binary files and build artifacts now properly excluded by default
- Reduced false positives from scanning compiled binaries when gitignore is present

## [2.0.0] - 2025-12-08

### Added

- 🎯 **Deduplication System**: Automatically removes duplicate findings using SHA-256 hashing
- 🚫 **Allowlist Support**: Built-in patterns to filter common false positives
- ⚙️ **Configurable Entropy**: Adjustable Shannon entropy threshold (default raised from 4.0 to 5.0)
- 📋 **20+ Detection Patterns**: Expanded from 4 to 20+ secret patterns including GitHub, Slack, SendGrid, etc.
- 📊 **Detailed Statistics**: Track files scanned, commits scanned, unique findings, and scan duration
- 🎨 **Rich Output Formatting**: Color-coded severity levels (Critical, High, Medium, Low)
- 🔧 **Custom Configuration**: Support for `.secscan.toml` configuration files
- 📄 **Enhanced JSON Export**: Includes statistics and metadata
- 🏃 **Performance Tracking**: Scan duration and throughput metrics
- 🎯 **Severity Classification**: Automatic confidence-based categorization
- 📝 **Verbose Mode**: Optional detailed output with `-verbose` flag
- 🔕 **Quiet Mode**: Minimal output for CI/CD integration
- 🚀 **Version Flag**: Display version information with `-version`
- 📑 **Better Documentation**: Comprehensive README with examples

### Changed

- **Entropy Threshold**: Raised default from 4.0 to 5.0 to reduce false positives
- **File Filtering**: Enhanced skip logic for lock files, minified files, and binary formats
- **Git History Entropy**: Uses higher threshold (default + 0.5) for historical scans
- **Comment Detection**: Automatically skips commented lines in source code
- **Output Format**: Improved readability with emojis and clear sections
- **Error Handling**: Better handling of file read errors and git failures

### Improved

- 📉 **95% Reduction in False Positives**: From 511K findings to manageable numbers
- ⚡ **Better Performance**: Optimized scanning with early exits and filtering
- 🎯 **Accuracy**: Context-aware detection with allowlist patterns
- 📚 **File Type Support**: Expanded to 30+ file extensions
- 🗂️ **Directory Skipping**: Comprehensive list of common build/cache directories
- 🔍 **Pattern Quality**: More precise regex patterns with fewer false matches

### Fixed

- False positives from constant names (e.g., `PAYMENT_METHOD_CARD`)
- Duplicate findings across git commits
- Excessive output size (reduced from 101MB to human-readable)
- Classification of test data as secrets
- Scanning of binary and generated files

## [1.0.0] - 2024-XX-XX

### Added

- Initial release
- Basic regex pattern matching (4 patterns)
- Git history scanning
- File tree scanning
- JSON output support
- Basic entropy detection (fixed 4.0 threshold)

### Known Issues (Fixed in 2.0.0)

- Very high false positive rate (511,125 findings on test project)
- No deduplication across commits
- Fixed entropy threshold too sensitive
- Limited file type filtering
- No allowlist support
- Excessive output size (101MB log files)

---

## Migration Guide: v1.0 → v2.0

### Breaking Changes

None! v2.0 is fully backward compatible.

### Recommended Updates

**Old Usage:**

```bash
secscan -root . -json report.json
```

**New Enhanced Usage:**

```bash
# Take advantage of new features
secscan -root . -json report.json -entropy 5.5 -verbose
```

### Configuration Migration

If you were using `rules.toml`, rename it to `.secscan.toml` for automatic loading.

### Output Changes

- JSON structure now includes `stats` section
- Findings include `hash` field for deduplication
- Exit codes remain the same (0 = no secrets, 1 = secrets found, 2 = error)
