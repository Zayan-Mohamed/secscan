# SecScan - Enhanced Secret Scanner

<div align="center">

![Logo](/assets/logo.png)

![License](https://img.shields.io/badge/license-MIT-green.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg)
[![GitHub release](https://img.shields.io/github/release/Zayan-Mohamed/secscan.svg)](https://github.com/Zayan-Mohamed/secscan/releases)

**Fast, configurable, and intelligent secret detection for your source code**

[Quick Start](#-quick-start) • [Documentation](https://zayan-mohamed.github.io/secscan) • [Examples](EXAMPLES.md) • [Changelog](CHANGELOG.md)

</div>

---

## � Quick Install

### Linux & macOS

```bash
curl -fsSL https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-curl.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/Zayan-Mohamed/secscan/main/scripts/install-windows.ps1 | iex
```

### Using Go

```bash
go install github.com/Zayan-Mohamed/secscan/v2@latest
```

> 💡 **More installation options**: [See Installation Guide](INSTALL.md)

---

## 🚀 Features

- ✅ **Enhanced Detection** - 20+ built-in patterns for API keys, tokens, and secrets
- 🧠 **Smart Entropy Analysis** - Configurable Shannon entropy detection with reduced false positives
- 🎯 **Deduplication** - Automatically removes duplicate findings across commits
- 🚫 **Allowlist Support** - Filter out known false positives
- 📊 **Detailed Statistics** - Track scan performance and coverage
- 🎨 **Rich Output** - Color-coded severity levels and clean formatting
- 📜 **Git History Scanning** - Deep scan through your entire git history
- 🔧 **Configurable** - Custom rules via TOML configuration
- ⚡ **Fast** - Written in Go for maximum performance
- 📄 **JSON Export** - Machine-readable output for CI/CD integration
- 🙈 **Gitignore Support** - Automatically respects `.gitignore` patterns to skip irrelevant files
- 🌍 **Cross-Platform** - Works on Linux, macOS, and Windows

## 📦 Installation

### Quick Install by Platform

#### 🐧 Linux / 🍎 macOS

```bash
# Navigate to secscan directory
cd secscan

# Option 1: Using installation script (recommended)
./install.sh

# Option 2: Using Make
make install         # System-wide (requires sudo)
# OR
make install-local   # User-only (no sudo)
```

#### 🪟 Windows

```powershell
# Navigate to secscan directory
cd secscan

# Using PowerShell script (recommended)
.\install.ps1

# For system-wide installation (requires admin):
.\install.ps1 -Global
```

> **Note**: Windows doesn't come with `make` by default. Use the PowerShell script instead.

#### 🌍 Universal Installer (All Platforms)

Works on Linux, macOS, and Windows:

```bash
cd secscan
go run installer/install.go
```

### Verify Installation

```bash
secscan -version
which secscan
```

## 🎯 Quick Start

```bash
# Scan current directory
secscan

# Scan specific directory
secscan -root /path/to/project

# Scan without git history (faster)
secscan -history=false

# Respect .gitignore files (default behavior)
secscan -respect-gitignore=true

# Disable gitignore (scan all files including ignored ones)
secscan -respect-gitignore=false

# Adjust entropy threshold (useful range ~4.0-4.8; see Entropy Threshold below)
secscan -entropy 4.5

# Disable entropy detection entirely
secscan -no-entropy

# Export results to JSON
secscan -json report.json

# Verbose output (show all findings)
secscan -verbose

# Quiet mode (for CI/CD)
secscan -quiet
```

## 🔍 Detection Patterns

SecScan detects the following secret types out of the box:

- **Cloud Providers**: AWS keys, Google API keys
- **Payment**: Stripe keys (live & restricted)
- **Version Control**: GitHub tokens (PAT, OAuth, App)
- **Communication**: Slack tokens & webhooks
- **Email**: SendGrid, Mailgun API keys
- **Database**: Connection strings (PostgreSQL, MySQL, MongoDB, Redis)
- **Authentication**: JWT tokens, Supabase keys
- **Generic**: API keys, secrets, passwords
- **High Entropy**: Random-looking strings (configurable)

## ⚙️ Configuration

### Custom Rules File

Create a `.secscan.toml` file:

```toml
# Custom detection rules
custom_api = "mycompany_api_[0-9a-zA-Z]{32}"
internal_token = "int_tok_[A-Za-z0-9]{40}"
```

Use it:

```bash
secscan -config .secscan.toml
```

### Gitignore Support

**SecScan automatically respects `.gitignore` files** in your repository, helping to:

- ✅ Skip build artifacts, dependencies, and generated files
- ✅ Reduce false positives from vendor code
- ✅ Speed up scans by skipping irrelevant files
- ✅ Work seamlessly with your existing Git workflow

**How it works:**

- Automatically finds and loads all `.gitignore` files in the repository
- Supports nested `.gitignore` files in subdirectories
- Handles negation patterns (`!important.txt`)
- Supports directory-only patterns (`logs/`)
- Compatible with standard gitignore glob patterns

**Examples:**

```bash
# Default: gitignore is enabled
secscan -root .

# Explicitly enable gitignore (same as default)
secscan -respect-gitignore=true

# Disable gitignore to scan ALL files (useful for security audits)
secscan -respect-gitignore=false

# Verbose mode shows which files are being skipped
secscan -verbose -respect-gitignore=true
```

**When to disable gitignore:**

- Security audits where you need to scan everything
- Checking if secrets exist in build artifacts
- Debugging scan results

### Entropy Threshold

The entropy threshold controls how "random" a string must be to be flagged.

Entropy is only consulted on lines that name the value as a credential
(`api_key = "..."`, `token: ...`), and only on tokens 20-100 characters long.
On its own it is not a usable detector — set it high enough to reject ordinary
source and it rejects real credentials too; set it low enough to catch them and
every base64 fragment in the tree fires.

- **Default**: 4.1
- **Stricter**: 4.5 (fewer, higher-confidence findings)
- **Disabled**: Use `-no-entropy`

The useful range is narrow, and the ceiling is not a matter of taste. Shannon
entropy over a token's own characters cannot exceed `log2(len)`, so a 20
character secret tops out at 4.32 and a 32 character one at 5.00. Real
credentials land between roughly 4.7 and 4.8 — a GitHub PAT scores 4.77, a
Stripe key 4.75. A threshold of 5.0 or above is not "strict"; it is
unreachable, and silently detects nothing. Values above ~4.8 are not useful.

The default sits just above `log2(16) = 4.0`, the ceiling for hex, so git SHAs
and integrity hashes fall out by construction rather than by luck.

```bash
# Stricter - fewer, higher confidence findings
secscan -entropy 4.5

# Lenient mode - catch more potential secrets
secscan -entropy 4.0
```

## 📊 Output Format

### Human-Readable Output

```
🔍 SecScan v2.0.0 - Enhanced Secret Scanner
📂 Scanning: /path/to/project
⚙️  Entropy threshold: 4.1
📋 Rules loaded: 20
📜 Git history: enabled

🔍 Secret Scan Results
==================================================
Total findings: 5
  Critical (≥0.9): 2
  High (≥0.8):     1
  Medium (≥0.6):   2
  Low (<0.6):      0
==================================================

🔴 [CRITICAL] [AWS_ACCESS_KEY] src/config.js:42
  → AKIA****************ABCD (confidence: 0.90)

🟠 [HIGH] [GITHUB_PAT] .env:15
  → ghp_********************************WXYZ (confidence: 0.85)

📊 Scan Statistics
==================================================
Files scanned:    1,234
Commits scanned:  567
Total findings:   12,345
Unique findings:  5
Scan duration:    2.5s
==================================================
```

### JSON Output

```json
{
  "findings": [
    {
      "file": "src/config.js",
      "line": 42,
      "pattern": "aws_access_key",
      "excerpt": "AKIA****************ABCD",
      "confidence": 0.9,
      "verified": false,
      "hash": "a1b2c3d4e5f6g7h8"
    }
  ],
  "stats": {
    "files_scanned": 1234,
    "commits_scanned": 567,
    "findings_total": 12345,
    "findings_unique": 5,
    "scan_duration_ms": 2500
  },
  "version": "2.0.0"
}
```

## 🛡️ CI/CD Integration

### GitHub Actions

```yaml
name: Secret Scan

on: [push, pull_request]

jobs:
  secscan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0 # Full history for git scanning

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Install SecScan
        run: |
          cd secscan
          make install-local

      - name: Run SecScan
        run: secscan -quiet -json secscan-report.json

      - name: Upload Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: secscan-report
          path: secscan-report.json
```

### GitLab CI

```yaml
secret_scan:
  image: golang:1.21
  script:
    - cd secscan
    - make install-local
    - export PATH="$HOME/.local/bin:$PATH"
    - secscan -quiet -json report.json
  artifacts:
    reports:
      junit: report.json
    when: always
```

## 🎓 How It Works

### 1. Pattern Matching

SecScan uses regex patterns to detect known secret formats (AWS keys, GitHub tokens, etc.)

### 2. Entropy Analysis

Calculates Shannon entropy to find high-randomness strings that might be secrets:

```
Entropy = -Σ(p(x) * log2(p(x)))
```

Strings with entropy > threshold and diverse character sets are flagged.

### 3. Deduplication

Uses SHA-256 hashing to identify and remove duplicate findings across different files/commits.

### 4. Allowlisting

Filters out common false positives:

- All-caps constants
- Test/example values
- Boolean literals
- Masked secrets

## 🔧 Advanced Usage

### Skip Git History for Speed

```bash
secscan -history=false
```

### Scan Only Specific Patterns

Create a minimal config with only the rules you need:

```toml
# minimal-rules.toml
aws_access_key = "AKIA[0-9A-Z]{16}"
github_pat = "ghp_[0-9a-zA-Z]{36}"
```

```bash
secscan -config minimal-rules.toml
```

### Combine with Other Tools

```bash
# Find secrets and filter by pattern
secscan -json findings.json
jq '.findings[] | select(.pattern == "aws_access_key")' findings.json

# Count secrets by type
jq '.findings | group_by(.pattern) | map({pattern: .[0].pattern, count: length})' findings.json
```

## 📈 Improvements Over v1.0

| Feature              | v1.0                 | v2.0                 |
| -------------------- | -------------------- | -------------------- |
| Detection Patterns   | 4                    | 20+                  |
| False Positive Rate  | High (511K findings) | Low (~95% reduction) |
| Deduplication        | ❌                   | ✅                   |
| Allowlist Support    | ❌                   | ✅                   |
| Configurable Entropy | ❌ (fixed 4.0)       | ✅ (default 4.1)     |
| Skip Files/Dirs      | Limited              | Comprehensive        |
| Output Formatting    | Basic                | Rich with colors     |
| Statistics           | ❌                   | ✅                   |
| Performance          | Good                 | Excellent            |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

Inspired by:

- [TruffleHog](https://github.com/trufflesecurity/trufflehog)
- [GitLeaks](https://github.com/gitleaks/gitleaks)
- [Detect-Secrets](https://github.com/Yelp/detect-secrets)

## 📞 Support

- 🐛 [Report Issues](https://github.com/Zayan-Mohamed/secscan/issues)
- 💬 [Discussions](https://github.com/Zayan-Mohamed/secscan/discussions)
- 📧 Email: [Zayan Mohamed](mailto:itsm.zayan@gmail.com)

---

<div align="center">
Made with ❤️ by the SecScan team
</div>
