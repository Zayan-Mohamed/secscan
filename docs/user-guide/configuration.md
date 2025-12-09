# Configuration

SecScan can be configured using command-line flags or a TOML configuration file.

## Configuration File

Create `.secscan.toml` in your project root:

```toml
[general]
entropy_threshold = 5.0
scan_history = true
respect_gitignore = true
verbose = false

[[allowlist]]
path = "test/fixtures/"
reason = "Test data and mock credentials"

[[allowlist]]
path = "docs/examples/"
reason = "Documentation examples"

[[allowlist]]
value = "EXAMPLE_API_KEY_1234567890"
reason = "Placeholder in documentation"

[[custom_patterns]]
name = "Custom API Key Format"
pattern = "MYAPP-[A-Z0-9]{32}"
description = "Our custom API key format"
```

## Configuration Options

### General Settings

| Option              | Type  | Default  | Description                         |
| ------------------- | ----- | -------- | ----------------------------------- |
| `entropy_threshold` | float | 4.5      | Minimum entropy for detection (0-8) |
| `scan_history`      | bool  | true     | Scan git history                    |
| `respect_gitignore` | bool  | true     | Honor .gitignore patterns           |
| `verbose`           | bool  | false    | Show detailed output                |
| `max_file_size`     | int   | 10485760 | Max file size in bytes (10MB)       |

### Allowlist Configuration

Suppress false positives using allowlists:

#### By Path

```toml
[[allowlist]]
path = "test/"
reason = "Test files only"
```

#### By Value

```toml
[[allowlist]]
value = "sk-test-1234567890"
reason = "Test API key"
```

#### By Pattern

```toml
[[allowlist]]
pattern = "^test_.*_key$"
reason = "Test variables"
```

### Custom Patterns

Add your own detection patterns:

```toml
[[custom_patterns]]
name = "Internal Token Format"
pattern = "INT-[A-Z]{3}-[0-9]{16}"
description = "Internal service tokens"
severity = "high"
```

## Command-Line Flags

Command-line flags override configuration file settings.

### Basic Flags

```bash
# Specify config file
secscan -config /path/to/config.toml

# Override entropy threshold
secscan -entropy 5.5

# Disable git history scanning
secscan -history=false

# Disable gitignore
secscan -respect-gitignore=false

# Enable verbose mode
secscan -verbose
```

### Output Flags

```bash
# Export to JSON
secscan -json results.json

# Show version
secscan -version
```

### Detection Flags

```bash
# Disable entropy detection
secscan -no-entropy

# Scan specific directory
secscan -root /path/to/project
```

## Configuration Examples

### Strict Configuration

For maximum security:

```toml
[general]
entropy_threshold = 6.0
scan_history = true
respect_gitignore = false
verbose = true

# No allowlists - catch everything
```

### Balanced Configuration

Good for most projects:

```toml
[general]
entropy_threshold = 5.0
scan_history = true
respect_gitignore = true
verbose = false

[[allowlist]]
path = "test/"
reason = "Test files"

[[allowlist]]
path = "docs/"
reason = "Documentation"
```

### Fast Configuration

For quick scans:

```toml
[general]
entropy_threshold = 5.5
scan_history = false
respect_gitignore = true
verbose = false
```

## Environment-Specific Configuration

### Development

`.secscan.dev.toml`:

```toml
[general]
entropy_threshold = 4.5
scan_history = false
respect_gitignore = true

[[allowlist]]
path = "test/"
reason = "Test data"
```

### CI/CD

`.secscan.ci.toml`:

```toml
[general]
entropy_threshold = 5.5
scan_history = true
respect_gitignore = true
verbose = true

# Stricter - fewer allowlists
```

Usage:

```bash
# Development
secscan -config .secscan.dev.toml

# CI/CD
secscan -config .secscan.ci.toml
```

## Gitignore Integration

SecScan automatically respects `.gitignore` patterns:

```gitignore
# Automatically skipped when respect_gitignore = true
node_modules/
vendor/
*.log
.env
venv/
```

To scan ignored files:

```bash
secscan -respect-gitignore=false
```

## Best Practices

### 1. Version Control

Commit your configuration:

```bash
git add .secscan.toml
git commit -m "Add SecScan configuration"
```

### 2. Document Allowlists

Always add a `reason` to allowlist entries:

```toml
[[allowlist]]
path = "fixtures/test-data.json"
reason = "Mock API responses for testing"
```

### 3. Different Configs for Different Environments

- `.secscan.toml` - Default/development
- `.secscan.ci.toml` - CI/CD pipeline
- `.secscan.strict.toml` - Pre-production audit

### 4. Regular Reviews

Periodically review your configuration:

```bash
# Test with strict settings
secscan -config .secscan.strict.toml
```

## Troubleshooting

### Too Many False Positives

1. Increase entropy threshold:

   ```toml
   entropy_threshold = 5.5
   ```

2. Add allowlists for test directories

3. Disable entropy detection:
   ```bash
   secscan -no-entropy
   ```

### Missing Secrets

1. Lower entropy threshold:

   ```toml
   entropy_threshold = 4.0
   ```

2. Enable git history:

   ```toml
   scan_history = true
   ```

3. Disable gitignore:
   ```bash
   secscan -respect-gitignore=false
   ```

## Next Steps

- 🚀 [Advanced Features](advanced-features.md)
- 💡 [Examples](examples.md)
- 🔄 [CI/CD Integration](ci-cd-integration.md)
- 📖 [CLI Reference](../reference/cli-options.md)
