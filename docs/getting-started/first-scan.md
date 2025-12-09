# Your First Scan

Let's walk through running your first comprehensive scan with SecScan.

## Basic Scan

The simplest way to scan a project:

```bash
# Navigate to your project
cd /path/to/your/project

# Run scan
secscan
```

This will scan:

- All files in the current directory and subdirectories
- Git history (if in a git repository)
- Respect `.gitignore` patterns

## Understanding the Output

### Scan Summary

```
Scanning directory: /home/user/myproject
Respecting .gitignore patterns
Scanning git history...
Processing 156 commits...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                 SECURITY SCAN RESULTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Finding Details

Each finding shows:

- **Severity Level**: HIGH (🔴), MEDIUM (🟡), or LOW (🟢)
- **File Path**: Where the secret was found
- **Line Number**: Exact location in the file
- **Pattern Type**: What kind of secret was detected
- **Code Excerpt**: Context around the finding

Example:

```
[HIGH] File: config/database.go:42 (Pattern: PostgreSQL Connection String)
  db_url = "postgresql://admin:p4ssw0rd@localhost/prod"
  Commit: a1b2c3d [2024-11-15] - "Update database config"
```

### Statistics

At the end, you'll see:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                    STATISTICS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total files scanned:    342
Files with findings:    5
Total findings:         12
High confidence:        4
Medium confidence:      6
Low confidence:         2
Scan duration:          2.34s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Customizing Your Scan

### Skip Git History (Faster)

For large repositories, scanning git history can be slow:

```bash
secscan -history=false
```

### Adjust Sensitivity

Reduce false positives by increasing the entropy threshold:

```bash
# Default is 4.5, higher values = fewer but more confident findings
secscan -entropy 5.5
```

### Disable Entropy Detection

Only use pattern matching:

```bash
secscan -no-entropy
```

### Scan Ignored Files

Include files normally ignored by `.gitignore`:

```bash
secscan -respect-gitignore=false
```

### Export Results

Save findings to a JSON file:

```bash
secscan -json scan-results.json
```

### Verbose Output

See detailed scanning progress:

```bash
secscan -verbose
```

## Recommended First Scan

For your first scan, we recommend:

```bash
secscan -verbose -json initial-scan.json
```

This will:

- Show you what's being scanned
- Save results for later review
- Use default settings (good balance)

## Interpreting Results

### What to Fix First

1. **HIGH confidence findings** - These are almost certainly real secrets

   - Change these credentials immediately
   - Rotate any exposed API keys
   - Update configuration files

2. **MEDIUM confidence findings** - Review these carefully

   - May be test data or false positives
   - Verify if they're sensitive

3. **LOW confidence findings** - Usually safe to ignore
   - Often hash values or non-sensitive strings
   - Review if in sensitive files

### Common False Positives

SecScan may flag:

- Test fixtures and mock data
- Example configurations in documentation
- Non-sensitive hash values
- Public API keys for testing

To suppress these, use an allowlist (see [Configuration Guide](../user-guide/configuration.md)).

## Next Steps

- 📖 Learn about [Basic Usage](../user-guide/basic-usage.md)
- 🔧 Set up [Configuration](../user-guide/configuration.md)
- 💡 Explore more [Examples](../user-guide/examples.md)
- 🚀 Integrate with [CI/CD](../user-guide/ci-cd-integration.md)

## Tips

!!! tip "Start with History Disabled"
For very large repositories, start with `-history=false` to get a quick overview before scanning the full git history.

!!! warning "Don't Commit Findings"
Never commit the JSON output file to your repository - it contains the actual secrets!

!!! info "Incremental Scanning"
For regular scans, consider using `-history=false` and only scan git history periodically.
