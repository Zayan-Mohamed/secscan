# Example Usage

Comprehensive examples of SecScan in action.

## Quick Reference

```bash
# Basic scan
secscan

# Scan with JSON output
secscan -json results.json

# Quick scan (no history)
secscan -history=false

# High confidence only
secscan -entropy 6.0

# Pattern matching only
secscan -no-entropy

# Scan all files (ignore .gitignore)
secscan -respect-gitignore=false

# Custom config
secscan -config .secscan.toml

# Verbose output
secscan -verbose
```

## Real-World Scenarios

### Scenario 1: Initial Project Audit

You've inherited a codebase and want to check for secrets:

```bash
# Full deep scan with verbose output
secscan -verbose -json initial-audit.json

# Review high-confidence findings first
cat initial-audit.json | jq '.findings[] | select(.confidence >= 90)'
```

### Scenario 2: Pre-Commit Check

Quick scan before committing:

```bash
# Fast scan of working directory only
secscan -history=false

# If clean, commit
if [ $? -eq 0 ]; then
  git commit -m "Your commit message"
else
  echo "Fix secrets before committing!"
fi
```

### Scenario 3: CI/CD Pipeline

Automated scanning in your pipeline:

```bash
# Strict CI scan
secscan -config .secscan.ci.toml -json ci-results.json

# Parse results
FINDINGS=$(cat ci-results.json | jq '.statistics.total_findings')

if [ "$FINDINGS" -gt 0 ]; then
  echo "❌ Found $FINDINGS potential secrets"
  exit 1
else
  echo "✅ No secrets detected"
fi
```

### Scenario 4: Monorepo Scanning

Scan specific services in a monorepo:

```bash
# Scan each service
for service in services/*/; do
  echo "Scanning $service..."
  secscan -root "$service" -json "results-$(basename $service).json"
done

# Combine results
jq -s '.' results-*.json > combined-results.json
```

### Scenario 5: Git History Deep Dive

Find secrets that existed in the past:

```bash
# Full history scan
secscan -verbose -json history-scan.json

# Group by commit
cat history-scan.json | jq '.findings[] | .commit' | sort | uniq -c
```

## Development Workflows

### Daily Development

```bash
#!/bin/bash
# save as: scripts/check-secrets.sh

echo "🔍 Scanning for secrets..."
secscan -history=false -entropy 5.0 -json /tmp/daily-scan.json

if [ $? -eq 0 ]; then
  echo "✅ No secrets found - you're good to go!"
else
  echo "⚠️  Secrets detected! Review /tmp/daily-scan.json"
  cat /tmp/daily-scan.json | jq '.findings[]'
  exit 1
fi
```

### Weekly Audit

```bash
#!/bin/bash
# save as: scripts/weekly-audit.sh

DATE=$(date +%Y%m%d)
REPORT_DIR="security-reports"
mkdir -p "$REPORT_DIR"

echo "Running weekly security audit..."
secscan \
  -verbose \
  -config .secscan.strict.toml \
  -json "$REPORT_DIR/weekly-scan-$DATE.json"

# Generate summary
echo "\n📊 Audit Summary:"
jq '.statistics' "$REPORT_DIR/weekly-scan-$DATE.json"

# Archive old reports (keep last 4 weeks)
find "$REPORT_DIR" -name "weekly-scan-*.json" -mtime +28 -delete
```

### Release Preparation

```bash
#!/bin/bash
# save as: scripts/pre-release-check.sh

echo "🚀 Pre-release security check..."

# Deep scan with strict settings
secscan \
  -entropy 6.0 \
  -respect-gitignore=false \
  -verbose \
  -json release-scan.json

HIGH_CONF=$(cat release-scan.json | jq '[.findings[] | select(.confidence >= 90)] | length')

if [ "$HIGH_CONF" -gt 0 ]; then
  echo "❌ $HIGH_CONF high-confidence secrets found!"
  echo "   Fix these before releasing!"
  cat release-scan.json | jq '.findings[] | select(.confidence >= 90)'
  exit 1
else
  echo "✅ Release security check passed!"
fi
```

## Integration Examples

### Pre-commit Hook

`.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running pre-commit secret scan..."
secscan -history=false -json /tmp/precommit-scan.json

if [ $? -ne 0 ]; then
  echo ""
  echo "❌ COMMIT BLOCKED: Secrets detected!"
  echo ""
  cat /tmp/precommit-scan.json | jq -r '.findings[] | "  • \(.file):\(.line) - \(.pattern)"'
  echo ""
  echo "Please remove secrets before committing."
  rm /tmp/precommit-scan.json
  exit 1
fi

rm /tmp/precommit-scan.json
echo "✅ No secrets detected"
exit 0
```

### Git Alias

Add to `.gitconfig`:

```ini
[alias]
  scan = "!secscan -history=false"
  scan-full = "!secscan -verbose -json git-scan.json"
  scan-strict = "!secscan -entropy 6.0 -respect-gitignore=false"
```

Usage:

```bash
git scan
git scan-full
git scan-strict
```

### Make Integration

Add to `Makefile`:

```makefile
.PHONY: security-scan
security-scan:
	@echo "Running security scan..."
	@secscan -json security-scan.json
	@echo "✅ Scan complete. Results in security-scan.json"

.PHONY: security-check
security-check:
	@secscan -history=false || (echo "❌ Security check failed" && exit 1)

.PHONY: pre-commit
pre-commit: security-check test lint
	@echo "✅ All pre-commit checks passed"
```

## Language-Specific Examples

### Go Projects

```bash
# Scan Go project, excluding vendor
secscan -root . -verbose
# vendor/ is usually in .gitignore
```

### Node.js Projects

```bash
# Fast scan excluding node_modules
secscan -root . -history=false
# node_modules/ is usually in .gitignore

# Include environment variable files
secscan -respect-gitignore=false | grep -E "\.env"
```

### Python Projects

```bash
# Scan Python project
secscan -root . -verbose

# Focus on config files
secscan -root . | grep -E "(settings|config|\.env)"
```

### Docker Projects

```bash
# Scan including Dockerfiles
secscan -root . | grep -E "(Dockerfile|docker-compose)"
```

## Advanced Filtering

### High Confidence Only

```bash
secscan -json all.json
cat all.json | jq '.findings[] | select(.confidence >= 90)'
```

### By File Pattern

```bash
secscan -json all.json
cat all.json | jq '.findings[] | select(.file | test("config|settings"))'
```

### By Pattern Type

```bash
secscan -json all.json
cat all.json | jq '.findings[] | select(.pattern == "AWS Access Key")'
```

### Group by Severity

```bash
secscan -json all.json

echo "High Confidence:"
cat all.json | jq '.findings[] | select(.confidence >= 90) | .file'

echo "\nMedium Confidence:"
cat all.json | jq '.findings[] | select(.confidence >= 70 and .confidence < 90) | .file'

echo "\nLow Confidence:"
cat all.json | jq '.findings[] | select(.confidence < 70) | .file'
```

## Comparison Scripts

### Before/After Comparison

```bash
#!/bin/bash
# Compare scans before and after changes

# Scan current state
secscan -json before.json

# Make your changes
# ... edit files ...

# Scan again
secscan -json after.json

# Compare
echo "Findings before: $(cat before.json | jq '.statistics.total_findings')"
echo "Findings after: $(cat after.json | jq '.statistics.total_findings')"

# Show difference
diff <(cat before.json | jq -S '.findings') <(cat after.json | jq -S '.findings')
```

## Tips and Tricks

### Quick Stats

```bash
secscan -json scan.json && cat scan.json | jq '.statistics'
```

### Export to CSV

```bash
secscan -json scan.json
cat scan.json | jq -r '.findings[] | [.file, .line, .pattern, .confidence] | @csv' > findings.csv
```

### Search Specific Pattern

```bash
secscan | grep -i "aws\|api\|token\|password"
```

### Scan Specific File Types

```bash
# Only scan .go files (requires custom filtering)
secscan -json all.json
cat all.json | jq '.findings[] | select(.file | endswith(".go"))'
```

## Next Steps

- 🔧 [Configuration Guide](configuration.md)
- 🚀 [Advanced Features](advanced-features.md)
- 🔄 [CI/CD Integration](ci-cd-integration.md)
- 📖 [CLI Reference](../reference/cli-options.md)
