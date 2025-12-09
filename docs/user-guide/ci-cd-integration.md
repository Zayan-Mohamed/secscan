# CI/CD Integration

Integrate SecScan into your continuous integration and deployment pipelines.

## GitHub Actions

Create `.github/workflows/secscan.yml`:

```yaml
name: Secret Scanning

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  scan:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        with:
          fetch-depth: 0 # Full history for git scanning

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Install SecScan
        run: |
          git clone https://github.com/Zayan-Mohamed/secscan.git
          cd secscan
          make build
          sudo cp build/secscan /usr/local/bin/

      - name: Run SecScan
        run: |
          secscan -json scan-results.json

      - name: Upload scan results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: secscan-results
          path: scan-results.json

      - name: Check for secrets
        run: |
          if [ -s scan-results.json ]; then
            echo "❌ Secrets detected!"
            cat scan-results.json
            exit 1
          else
            echo "✅ No secrets found"
          fi
```

## GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - security

secret_scan:
  stage: security
  image: golang:1.21

  before_script:
    - git clone https://github.com/Zayan-Mohamed/secscan.git
    - cd secscan
    - make build
    - cp build/secscan /usr/local/bin/
    - cd ..

  script:
    - secscan -json scan-results.json || true
    - |
      if [ -s scan-results.json ]; then
        echo "❌ Secrets detected!"
        cat scan-results.json
        exit 1
      else
        echo "✅ No secrets found"
      fi

  artifacts:
    when: always
    paths:
      - scan-results.json
    expire_in: 30 days

  only:
    - branches
    - merge_requests
```

## Jenkins

Create `Jenkinsfile`:

```groovy
pipeline {
    agent any

    stages {
        stage('Install SecScan') {
            steps {
                sh '''
                    git clone https://github.com/Zayan-Mohamed/secscan.git
                    cd secscan
                    make build
                    cp build/secscan ${WORKSPACE}/secscan
                '''
            }
        }

        stage('Secret Scan') {
            steps {
                sh '''
                    ./secscan -json scan-results.json || true
                '''
            }
        }

        stage('Process Results') {
            steps {
                script {
                    def results = readJSON file: 'scan-results.json'
                    if (results.findings && results.findings.size() > 0) {
                        error "❌ Secrets detected: ${results.findings.size()} findings"
                    } else {
                        echo "✅ No secrets found"
                    }
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'scan-results.json', allowEmptyArchive: true
        }
    }
}
```

## CircleCI

Create `.circleci/config.yml`:

```yaml
version: 2.1

jobs:
  secret-scan:
    docker:
      - image: cimg/go:1.21

    steps:
      - checkout

      - run:
          name: Install SecScan
          command: |
            git clone https://github.com/Zayan-Mohamed/secscan.git
            cd secscan
            make build
            sudo cp build/secscan /usr/local/bin/

      - run:
          name: Run SecScan
          command: |
            secscan -json scan-results.json || true

      - run:
          name: Check Results
          command: |
            if [ -s scan-results.json ]; then
              echo "❌ Secrets detected!"
              cat scan-results.json
              exit 1
            else
              echo "✅ No secrets found"
            fi

      - store_artifacts:
          path: scan-results.json

workflows:
  version: 2
  scan:
    jobs:
      - secret-scan
```

## Docker Integration

### Dockerfile for SecScan

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o secscan main.go

FROM alpine:latest
RUN apk --no-cache add git
COPY --from=builder /build/secscan /usr/local/bin/secscan
ENTRYPOINT ["secscan"]
```

### Using in CI

```yaml
# GitHub Actions
- name: Run SecScan in Docker
  run: |
    docker build -t secscan .
    docker run -v $(pwd):/scan secscan -root /scan -json /scan/results.json
```

## Pre-commit Hook

Scan before committing:

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running SecScan..."
secscan -history=false -json /tmp/secscan-results.json

if [ -s /tmp/secscan-results.json ]; then
    echo "❌ Secrets detected! Commit aborted."
    cat /tmp/secscan-results.json
    rm /tmp/secscan-results.json
    exit 1
fi

echo "✅ No secrets found"
rm /tmp/secscan-results.json
exit 0
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Advanced CI/CD Patterns

### Scan Only Changed Files

```bash
# Get changed files
CHANGED_FILES=$(git diff --name-only HEAD~1)

# Scan only changed files (pseudo-code - adapt as needed)
secscan -history=false | grep -F "$CHANGED_FILES"
```

### Differential Scanning

```bash
# Scan current branch
secscan -json current.json

# Scan main branch
git checkout main
secscan -json main.json

# Compare results
diff current.json main.json
```

### Scheduled Deep Scans

```yaml
# GitHub Actions - weekly full scan
on:
  schedule:
    - cron: "0 0 * * 0" # Every Sunday

jobs:
  deep-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - name: Deep scan with history
        run: secscan -verbose -json weekly-scan.json
```

## Best Practices

### 1. Fail Fast

Configure CI to fail immediately on secrets:

```bash
secscan -json results.json
if [ -s results.json ]; then
  exit 1
fi
```

### 2. Archive Results

Always save scan results:

```yaml
artifacts:
  paths:
    - scan-results.json
  expire_in: 30 days
```

### 3. Different Configs for CI

Use stricter settings in CI:

```bash
secscan -config .secscan.ci.toml -entropy 5.5
```

### 4. Notifications

Send alerts on findings:

```bash
if [ -s scan-results.json ]; then
  curl -X POST $SLACK_WEBHOOK \
    -d '{"text":"⚠️ Secrets detected in build!"}'
fi
```

### 5. Branch Protection

Require SecScan to pass before merging:

- GitHub: Settings → Branches → Branch protection rules
- GitLab: Settings → Repository → Protected branches

## Performance Tips

### Skip History in PR Checks

```bash
# Fast scan for PRs
secscan -history=false -json results.json
```

### Cache SecScan Binary

```yaml
# GitHub Actions
- name: Cache SecScan
  uses: actions/cache@v3
  with:
    path: /usr/local/bin/secscan
    key: secscan-v2.1.0
```

## Troubleshooting

### CI Timeout

Reduce scan scope:

```bash
secscan -history=false -entropy 5.5
```

### False Positives Failing Builds

Use CI-specific allowlist:

```toml
# .secscan.ci.toml
[[allowlist]]
path = "test/"
reason = "Test fixtures"
```

## Next Steps

- 🐳 [Docker Deployment](../deployment/docker.md)
- 🔧 [Configuration Guide](configuration.md)
- 💡 [Examples](examples.md)
- 📖 [CLI Reference](../reference/cli-options.md)
