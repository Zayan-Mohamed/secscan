# Quick Start

Get up and running with SecScan in under 5 minutes!

## Installation

The fastest way to install SecScan:

=== "Linux/macOS"

    ```bash
    # Clone the repository
    git clone https://github.com/Zayan-Mohamed/secscan.git
    cd secscan

    # Build and install
    make install
    ```

=== "Go Install"

    ```bash
    go install github.com/Zayan-Mohamed/secscan@latest
    ```

=== "Manual Build"

    ```bash
    # Clone and build
    git clone https://github.com/Zayan-Mohamed/secscan.git
    cd secscan
    make build

    # Binary will be in build/secscan
    ./build/secscan -version
    ```

## Verify Installation

```bash
secscan -version
```

Expected output:

```
SecScan v2.1.0
```

## Your First Scan

### Scan Current Directory

```bash
secscan
```

This will:

- Scan all files in the current directory
- Include git history if in a git repository
- Respect `.gitignore` patterns
- Display findings with color-coded severity

### Scan a Specific Project

```bash
secscan -root /path/to/your/project
```

### Quick Scan (Skip Git History)

For faster scans:

```bash
secscan -history=false
```

## Understanding Results

SecScan categorizes findings by confidence level:

- 🔴 **HIGH** (90-100%) - Very likely a real secret
- 🟡 **MEDIUM** (70-89%) - Potentially sensitive
- 🟢 **LOW** (<70%) - May be a false positive

Example output:

```
[HIGH] File: config/database.go:42 (Pattern: PostgreSQL Connection String)
  db_url = "postgresql://admin:p4ssw0rd@localhost/prod"

[MEDIUM] File: utils/crypto.go:15 (Pattern: High Entropy String)
  secret_key = "a8f5f167f44f4964e6c998dee827110c"
```

## Next Steps

- 📖 Learn more about [Installation Options](installation.md)
- 🎯 Run your [First Detailed Scan](first-scan.md)
- 🔧 Explore [Configuration Options](../user-guide/configuration.md)
- 💡 See more [Examples](../user-guide/examples.md)

## Common Issues

!!! warning "Permission Denied"
If you get "permission denied" when running `make install`:
`bash
    # Use local installation instead
    make install-local
    `

!!! info "Command Not Found"
If `secscan` is not found after installation:
`bash
    # Add to PATH (add to ~/.bashrc or ~/.zshrc)
    export PATH="$HOME/.local/bin:$PATH"
    source ~/.bashrc
    `
