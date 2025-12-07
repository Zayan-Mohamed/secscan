# SecScan - Enhanced Secret Scanner

<div align="center">

![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg)

**Fast, configurable, and intelligent secret detection for your source code**

[Quick Start](QUICKSTART.md) • [Installation Guide](INSTALL.md) • [Examples](EXAMPLES.md) • [Changelog](CHANGELOG.md)

</div>

> 💡 **New to SecScan?** Run `./install.sh` for easy setup, or use `make install` to install globally. Then use `secscan` from anywhere!

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

## 📦 Installation

### Quick Install (Recommended)

```bash
# Clone or navigate to the secscan directory
cd secscan

# Build and install system-wide (requires sudo)
make install
```

This will install `secscan` to `/usr/local/bin/` making it available as a global command.

### Manual Installation

```bash
# Build the binary
make build

# Install to /usr/local/bin (requires sudo)
sudo cp build/secscan /usr/local/bin/
sudo chmod +x /usr/local/bin/secscan
```

### Alternative: Local Installation (No sudo required)

```bash
# Build the binary
make build

# Copy to local bin directory (ensure ~/.local/bin is in your PATH)
mkdir -p ~/.local/bin
cp build/secscan ~/.local/bin/
chmod +x ~/.local/bin/secscan

# Add to PATH if not already (add this to your ~/.bashrc or ~/.zshrc)
export PATH="$HOME/.local/bin:$PATH"
```

### Using Go Install

```bash
go install github.com/Zayan-Mohamed/secscan@latest
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

# Adjust entropy threshold (higher = fewer false positives)
secscan -entropy 6.0

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

### Entropy Threshold

The entropy threshold controls how "random" a string must be to be flagged:

- **Default**: 5.0 (balanced)
- **Strict**: 6.0+ (fewer false positives)
- **Lenient**: 4.0-4.5 (more sensitive)
- **Disabled**: Use `-no-entropy`

```bash
# Strict mode - very high confidence
secscan -entropy 6.5

# Lenient mode - catch more potential secrets
secscan -entropy 4.0
```

## 📊 Output Format

### Human-Readable Output

```
🔍 SecScan v2.0.0 - Enhanced Secret Scanner
📂 Scanning: /path/to/project
⚙️  Entropy threshold: 5.0
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
| Configurable Entropy | ❌ (fixed 4.0)       | ✅ (default 5.0)     |
| Skip Files/Dirs      | Limited              | Comprehensive        |
| Output Formatting    | Basic                | Rich with colors     |
| Statistics           | ❌                   | ✅                   |
| Performance          | Good                 | Excellent            |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

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
