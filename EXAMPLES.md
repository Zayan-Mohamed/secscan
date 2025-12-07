# SecScan Example Usage Guide

This document provides practical examples of using SecScan in various scenarios.

## Basic Usage

### Scan Current Directory

```bash
secscan
```

### Scan Specific Project

```bash
secscan -root /path/to/your/project
```

## Common Scenarios

### 1. Quick Scan (Skip Git History)

For faster scans when you only care about current files:

```bash
secscan -history=false
```

### 2. High Confidence Only

Reduce false positives by increasing entropy threshold:

```bash
secscan -entropy 6.0
```

### 3. Pattern Matching Only

Disable entropy detection entirely:

```bash
secscan -no-entropy
```

### 4. Export Results for Analysis

```bash
secscan -json report.json
```

Then analyze with jq:

```bash
# List all unique secret types found
jq '.findings | group_by(.pattern) | map({type: .[0].pattern, count: length})' report.json

# Show only critical findings
jq '.findings[] | select(.confidence >= 0.9)' report.json

# Count findings by file
jq '.findings | group_by(.file) | map({file: .[0].file, secrets: length}) | sort_by(-.secrets)' report.json
```

### 5. CI/CD Integration

Run in quiet mode and fail build if secrets found:

```bash
secscan -quiet -json secscan-report.json
if [ $? -eq 1 ]; then
  echo "❌ Secrets detected! Check secscan-report.json"
  exit 1
fi
```

## Custom Configuration

### Create Custom Rules

Create `.secscan.toml`:

```toml
# Your company's API key format
company_api_key = "cmp_[0-9a-zA-Z]{40}"

# Internal service tokens
internal_token = "int_[A-Za-z0-9_-]{32}"

# Custom JWT pattern
custom_jwt = "eyJ[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+"
```

Use it:

```bash
secscan -config .secscan.toml
```

## Real-World Examples

### Example 1: Monorepo Scan

```bash
# Scan entire monorepo with custom settings
secscan -root ~/projects/monorepo \
  -entropy 5.5 \
  -verbose \
  -json monorepo-scan.json
```

### Example 2: Pre-Commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "🔍 Running secret scan..."
secscan -root . -history=false -quiet -json /tmp/secscan.json

if [ $? -eq 1 ]; then
  echo ""
  echo "❌ SECRET DETECTED! Commit blocked."
  echo "Review findings:"
  jq -r '.findings[] | "  - \(.file):\(.line) [\(.pattern)]"' /tmp/secscan.json
  echo ""
  exit 1
fi

echo "✅ No secrets detected"
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

### Example 3: GitHub Action

`.github/workflows/secret-scan.yml`:

```yaml
name: Secret Scan

on: [push, pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Install SecScan
        run: |
          cd secscan
          make install-local

      - name: Run Scan
        run: secscan -json results.json -verbose
        continue-on-error: true

      - name: Comment PR with Results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const results = JSON.parse(fs.readFileSync('results.json'));
            const count = results.findings.length;

            if (count > 0) {
              const comment = `## 🔍 Secret Scan Results\n\n` +
                `⚠️ Found ${count} potential secret(s)\n\n` +
                `Please review the findings and ensure no sensitive data is committed.`;
              
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: comment
              });
            }
```

### Example 4: Docker Integration

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -ldflags "-s -w" -o secscan main.go

FROM alpine:latest
RUN apk --no-cache add git
COPY --from=builder /app/secscan /usr/local/bin/
ENTRYPOINT ["secscan"]
```

Build and run:

```bash
docker build -t secscan .
docker run -v $(pwd):/scan secscan -root /scan
```

### Example 5: Scheduled Scanning

Create a cron job to scan regularly:

```bash
# Run every day at 2 AM
0 2 * * * /usr/local/bin/secscan -root /path/to/project -json /var/log/secscan/scan-$(date +\%Y\%m\%d).json
```

## Interpreting Results

### Understanding Confidence Scores

- **0.9+** (Critical): Very likely a real secret

  - AWS keys, GitHub tokens, API keys with known formats
  - **Action**: Rotate immediately!

- **0.8-0.89** (High): Probably a secret

  - Database connection strings, JWT tokens
  - **Action**: Review and rotate if confirmed

- **0.6-0.79** (Medium): Could be a secret

  - Generic API keys, high-entropy strings
  - **Action**: Investigate context

- **<0.6** (Low): Possibly a false positive
  - Very high entropy strings without clear pattern
  - **Action**: Review but likely safe

### Example Output Interpretation

```
🔴 [CRITICAL] [AWS_ACCESS_KEY] src/config.js:42
  → AKIA****************ABCD (confidence: 0.90)
```

**This is critical!** An AWS access key was found. You should:

1. Rotate the key immediately in AWS Console
2. Update your code to use environment variables
3. Add `.env` to `.gitignore`
4. Consider using AWS Secrets Manager

```
🟡 [MEDIUM] [HIGH_ENTROPY] tests/fixtures/data.json:15
  → test***************************data (confidence: 0.65)
```

**Review needed.** This is in a test file with lower confidence. Check if it's:

- Test data (safe)
- A real secret that shouldn't be in tests (unsafe)

## Troubleshooting

### Too Many False Positives?

Increase entropy threshold:

```bash
secscan -entropy 6.0
```

Or disable entropy completely:

```bash
secscan -no-entropy
```

### Scan Taking Too Long?

Skip git history:

```bash
secscan -history=false
```

### Want More Details?

Use verbose mode:

```bash
secscan -verbose
```

## Best Practices

1. **Run Locally First**: Test before adding to CI/CD
2. **Tune Threshold**: Adjust `-entropy` based on your codebase
3. **Review Output**: Don't blindly trust confidence scores
4. **Regular Scans**: Schedule periodic scans of important repos
5. **Combine Tools**: Use with other security tools for defense in depth

## Advanced Filtering

### Using jq for Analysis

Filter by pattern type:

```bash
jq '.findings[] | select(.pattern == "aws_access_key")' report.json
```

Group by file:

```bash
jq -r '.findings | group_by(.file) | .[] | "\(.[0].file): \(length) finding(s)"' report.json
```

Show only git history findings:

```bash
jq '.findings[] | select(.commit != null)' report.json
```

Get statistics:

```bash
jq '.stats' report.json
```
