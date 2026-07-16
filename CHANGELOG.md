# Changelog

All notable changes to SecScan will be documented in this file.

## [2.2.3] - 2026-07-17

Measured across a 37-repo corpus (1,874 files, 1,362 commits of history) on
default settings, false positives drop from 475 to 20 with no loss of recall.

### Breaking

- **The module path is now `github.com/Zayan-Mohamed/secscan/v2`.** Go requires
  a major-version suffix for v2+, and without it every tag from v2.0.0 to
  v2.2.2 was uninstallable — `go install ...@v2.2.2` failed outright, and
  `@latest` silently resolved to a pseudo-version of `main`'s tip. Installing a
  pinned release now works for the first time:

  ```bash
  go install github.com/Zayan-Mohamed/secscan/v2@latest
  ```

  The old path without `/v2` no longer resolves.

- `Finding.hash` is now derived from the secret rather than its location, so
  hashes from earlier versions will not match. Any stored baseline must be
  regenerated.
- A history scan that fails now exits 2 instead of 0.
- The default `-entropy` is 4.1, down from 5.0.

### Fixed

- **`-root` was ignored by the history scanner.** The git helpers ran with no
  working directory set, so they inherited the process CWD. `secscan -root
  /repo` — the obvious CI invocation — reported `commits_scanned: 0` and exited
  0, indistinguishable from a clean scan. Git history scanning, the headline
  feature, silently did nothing for every CI user.
- **Entropy detection could not fire on a real credential.** Shannon entropy
  over a token's own characters is bounded by `log2(len)`, so the 5.0 default
  silently required a 32+ character token with almost no repeated characters.
  Real credentials score below it: an AWS key id is 3.68, a GitHub PAT 4.77, a
  Stripe key 4.75. The detector only ever fired on long high-alphabet blobs —
  exactly the things that are not secrets. Entropy is now consulted only on
  lines that name the value as a credential, and candidate tokens are extracted
  from the base64url alphabet instead of `\S{20,}`, which had been handing whole
  URLs and HTML attributes to the check as single tokens.
- **Deduplication across commits could never fire.** The hash mixed in the
  location — the file path, or the commit SHA in history — so one leaked
  credential produced a fresh "unique" finding for every commit that touched it.
  One secret across four commits reported as 4; it now reports as 1, with an
  occurrence count.
- **The allowlist was dead code against every `generic_` rule.** Patterns were
  tested against the whole regex match (`password="testpass123"`) rather than
  the captured value, so anchored entries like `^test` could never match and
  placeholders were reported as credentials.
- **`heroku_api` was a bare UUID regex.** Heroku keys are UUIDs, so the shape
  alone carried no signal — it matched every seeded uuid in every fixture and
  SQL migration (99 findings, none real). It now requires the value to be named
  as a Heroku credential.
- **SecScan flagged its own test fixtures and documentation.** The allowlist was
  anchored, so it never matched AWS's canonical `AKIAIOSFODNN7EXAMPLE`.
- **Supabase anon keys are no longer reported.** They are designed to ship to
  browsers and are guarded by row-level security; only `service_role` bypasses
  it. Tokens are now judged by decoding them rather than by which rule matched.
- **`.min.js` and `.min.css` were never skipped.** `filepath.Ext("app.min.js")`
  returns `.js`, so the multi-part suffixes never matched and minified bundles
  were scanned as ordinary source.
- Three different version strings shipped in one binary: the constant said
  2.2.2, stdout said v2.1.0, and the JSON report said 2.2.0.
- History findings now name the file and line they came from. They previously
  reported `(git-history)` with an offset into the diff, which made 398 of the
  original 475 findings unactionable.
- Generated documentation sites are skipped. `git rev-list --all` reaches
  `gh-pages`, so every published MkDocs site was scanned; one file,
  `search/search_index.json`, produced 1,376 findings on its own.

### Added

- `-min-confidence` gates reporting by confidence, so CI can fail on severity
  instead of on any finding at all.
- `Finding.occurrences` counts how many times the same secret was seen.
- `stats.history_scanned` reports whether history coverage actually happened, so
  a scan that skipped it cannot be mistaken for a clean one.
- `Finding.verified` is now populated — previously it was written `false` on
  every finding and never revisited.

## [2.2.2] - 2025-12-12

### Added
- 

### Changed
- 

### Fixed
- 

All notable changes to SecScan will be documented in this file.

## [2.2.1] - 2025-12-11

### Added
- 

### Changed
- 

### Fixed
- 

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
