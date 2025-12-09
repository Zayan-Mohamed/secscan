# Command-Line Options Reference

Complete reference for all SecScan command-line options.

## Synopsis

```bash
secscan [options]
```

## Options

### General Options

#### `-root <path>`

Specify the directory to scan.

- **Type**: String
- **Default**: Current directory (`.`)
- **Example**: `secscan -root /path/to/project`

#### `-config <file>`

Path to configuration file.

- **Type**: String
- **Default**: `.secscan.toml` (if exists)
- **Example**: `secscan -config custom-config.toml`

#### `-version`

Display version information and exit.

- **Type**: Flag
- **Example**: `secscan -version`

#### `-verbose`

Enable verbose output showing detailed scanning progress.

- **Type**: Flag
- **Default**: `false`
- **Example**: `secscan -verbose`

### Detection Options

#### `-entropy <value>`

Set the minimum Shannon entropy threshold for detection.

- **Type**: Float (0.0 - 8.0)
- **Default**: `4.5`
- **Example**: `secscan -entropy 5.5`
- **Notes**: Higher values reduce false positives but may miss some secrets

#### `-no-entropy`

Disable entropy-based detection entirely.

- **Type**: Flag
- **Default**: `false`
- **Example**: `secscan -no-entropy`
- **Notes**: Only use pattern matching

### Git Options

#### `-history=<bool>`

Enable or disable git history scanning.

- **Type**: Boolean
- **Default**: `true`
- **Example**: `secscan -history=false`
- **Notes**: Disabling speeds up scans significantly

#### `-respect-gitignore=<bool>`

Honor `.gitignore` patterns when scanning.

- **Type**: Boolean
- **Default**: `true`
- **Example**: `secscan -respect-gitignore=false`
- **Notes**: Set to `false` to scan all files including ignored ones

### Output Options

#### `-json <file>`

Export findings to JSON file.

- **Type**: String
- **Default**: None (terminal output only)
- **Example**: `secscan -json results.json`

## Exit Codes

| Code | Meaning                            |
| ---- | ---------------------------------- |
| `0`  | Success - no secrets found         |
| `1`  | Secrets detected or error occurred |

## Usage Examples

### Basic Scans

```bash
# Scan current directory
secscan

# Scan specific project
secscan -root ~/projects/myapp

# Quick scan without history
secscan -history=false
```

### Detection Tuning

```bash
# High confidence only
secscan -entropy 6.0

# Pattern matching only
secscan -no-entropy

# Use custom config
secscan -config .secscan.strict.toml
```

### Output Control

```bash
# Verbose mode
secscan -verbose

# Export to JSON
secscan -json scan-results.json

# Both verbose and JSON
secscan -verbose -json results.json
```

### Advanced Usage

```bash
# Strict scan: no history, high threshold, all files
secscan -history=false -entropy 6.0 -respect-gitignore=false

# CI/CD mode: custom config, JSON output
secscan -config .secscan.ci.toml -json ci-results.json

# Development mode: fast scan with verbose output
secscan -history=false -verbose
```

## Environment Variables

SecScan does not currently use environment variables, but you can use them in your shell:

```bash
# Set default scan path
export SECSCAN_ROOT="/path/to/project"
secscan -root "$SECSCAN_ROOT"

# Set default config
export SECSCAN_CONFIG=".secscan.prod.toml"
secscan -config "$SECSCAN_CONFIG"
```

## Configuration File vs Command-Line

Command-line options take precedence over configuration file settings:

```bash
# Config file says entropy_threshold = 4.5
# This overrides it to 6.0
secscan -config .secscan.toml -entropy 6.0
```

## Common Option Combinations

### Fast Daily Scan

```bash
secscan -history=false -entropy 5.5
```

### Weekly Deep Scan

```bash
secscan -verbose -json weekly-$(date +%Y%m%d).json
```

### CI/CD Pipeline

```bash
secscan -config .secscan.ci.toml -json results.json
```

### Pre-commit Hook

```bash
secscan -history=false -json /tmp/precommit-results.json
```

### Audit Mode

```bash
secscan -respect-gitignore=false -entropy 4.0 -json audit.json
```

## Next Steps

- 📖 [Configuration File Reference](config-file.md)
- 🔍 [Detection Patterns](patterns.md)
- 📄 [Output Formats](output-formats.md)
- 💡 [Usage Examples](../user-guide/examples.md)
