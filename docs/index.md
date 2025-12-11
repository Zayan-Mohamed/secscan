# Welcome to SecScan

<p align="center">

  ![GitHub release](https://img.shields.io/github/release/Zayan-Mohamed/secscan.svg)
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg" alt="Go Version">
</p>

<p align="center">
  <strong>Fast, configurable, and intelligent secret detection for your source code</strong>
</p>

---

## What is SecScan?

SecScan is a powerful command-line tool designed to detect secrets, API keys, tokens, and other sensitive information in your source code. Built with Go for maximum performance, it provides comprehensive scanning capabilities for both current files and git history.

## Key Features

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
- 🙈 **Gitignore Support** - Automatically respects `.gitignore` patterns

## Quick Example

```bash
# Install SecScan
make install

# Scan your project
secscan -root /path/to/project

# Export results to JSON
secscan -root /path/to/project -json report.json
```

## Why SecScan?

Unlike basic regex-based scanners, SecScan combines:

- **Pattern matching** for known secret formats
- **Entropy analysis** to detect high-randomness strings
- **Git history scanning** to find leaked secrets in commits
- **Smart filtering** to reduce false positives
- **Fast performance** suitable for CI/CD pipelines

## Get Started

Ready to secure your code? Check out our [Quick Start Guide](getting-started/quickstart.md) or jump straight to [Installation](getting-started/installation.md).

## Need Help?

- 📖 Browse the [User Guide](user-guide/basic-usage.md)
- 💡 See [Examples](user-guide/examples.md)
- 🐛 Report issues on [GitHub](https://github.com/Zayan-Mohamed/secscan/issues)
- 📧 Contact: [itsm.zayan@gmail.com](mailto:itsm.zayan@gmail.com)
