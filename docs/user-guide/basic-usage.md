# Basic Usage

Learn how to use SecScan effectively in your daily workflow.

## Command Syntax

```bash
secscan [options]
```

## Common Usage Patterns

### 1. Scan Current Directory

```bash
secscan
```

### 2. Scan Specific Directory

```bash
secscan -root /path/to/project
```

### 3. Quick Scan (No History)

```bash
secscan -history=false
```

### 4. Export to JSON

```bash
secscan -json results.json
```

### 5. Custom Configuration

```bash
secscan -config .secscan.toml
```

## Common Options

### Directory and History

- `-root <path>` - Specify directory to scan (default: current directory)
- `-history=<true|false>` - Scan git history (default: true)
- `-respect-gitignore=<true|false>` - Honor .gitignore files (default: true)

### Detection Settings

- `-entropy <value>` - Set entropy threshold (default: 4.5)
- `-no-entropy` - Disable entropy-based detection
- `-config <file>` - Use custom configuration file

### Output Options

- `-json <file>` - Export results to JSON file
- `-verbose` - Show detailed scanning progress
- `-version` - Show version information

## Working with Different Project Types

### Go Projects

```bash
secscan -root . -config .secscan.toml
```

### JavaScript/Node.js

```bash
# Scan but ignore node_modules (automatic with .gitignore)
secscan -root .
```

### Python Projects

```bash
# Scan but ignore venv/virtualenv (automatic with .gitignore)
secscan -root .
```

### Monorepos

```bash
# Scan specific service
secscan -root ./services/api

# Or scan everything
secscan -root .
```

## Output Formats

### Terminal Output

Default color-coded output:

```
[HIGH] File: config/prod.env:12 (Pattern: AWS Access Key)
  AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
  Commit: a1b2c3d [2024-11-15] - "Add AWS credentials"

[MEDIUM] File: utils/crypto.go:42 (Pattern: High Entropy String)
  secret = "8f7a3b2c1d4e5f6a7b8c9d0e1f2a3b4c"
```

### JSON Output

Machine-readable format for integration:

```json
{
  "findings": [
    {
      "file": "config/prod.env",
      "line": 12,
      "commit": "a1b2c3d",
      "pattern": "AWS Access Key",
      "excerpt": "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
      "confidence": 95.5
    }
  ],
  "statistics": {
    "total_files": 342,
    "files_with_findings": 5,
    "total_findings": 12,
    "high_confidence": 4,
    "medium_confidence": 6,
    "low_confidence": 2
  }
}
```

## Confidence Levels

SecScan assigns confidence scores:

- **90-100%** (HIGH) - Almost certainly a real secret
- **70-89%** (MEDIUM) - Likely sensitive, review carefully
- **0-69%** (LOW) - May be false positive

## Best Practices

### 1. Regular Scanning

Run scans regularly:

```bash
# Before committing
secscan -history=false

# Weekly deep scan
secscan -json weekly-scan-$(date +%Y%m%d).json
```

### 2. Use Configuration Files

Create `.secscan.toml` in your project root:

```toml
[general]
entropy_threshold = 5.0
scan_history = true
respect_gitignore = true

[[allowlist]]
path = "test/fixtures"
reason = "Test data only"
```

### 3. CI/CD Integration

Add to your pipeline:

```bash
secscan -json scan-results.json
if [ -s scan-results.json ]; then
  echo "Secrets detected!"
  exit 1
fi
```

### 4. Incremental Scanning

For large repos:

```bash
# Daily: quick scan
secscan -history=false

# Weekly: full scan
secscan -verbose -json full-scan.json
```

## Performance Tips

### Speed Up Scans

1. **Skip git history**: `-history=false` (much faster)
2. **Use .gitignore**: Automatically skips node_modules, vendor, etc.
3. **Scan specific paths**: Target only changed directories

### Reduce False Positives

1. **Increase entropy threshold**: `-entropy 5.5`
2. **Use allowlists**: Configure known false positives
3. **Disable entropy detection**: `-no-entropy` for pattern-only

## Exit Codes

SecScan uses standard exit codes:

- `0` - No secrets found
- `1` - Secrets detected or error occurred

Use in scripts:

```bash
if secscan -root .; then
  echo "No secrets found"
else
  echo "Secrets detected!"
  exit 1
fi
```

## Next Steps

- 🔧 [Configuration Guide](configuration.md)
- 🚀 [Advanced Features](advanced-features.md)
- 💡 [More Examples](examples.md)
- 🔄 [CI/CD Integration](ci-cd-integration.md)
