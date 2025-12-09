# README for Documentation Development

This README explains how to build and deploy the SecScan documentation.

## Prerequisites

- Python 3.8+
- pip

## Installation

Install MkDocs and required plugins:

```bash
pip install mkdocs mkdocs-material pymdown-extensions
```

## Local Development

### Start Development Server

```bash
mkdocs serve
```

The documentation will be available at `http://127.0.0.1:8000`

The server will auto-reload when you make changes to the documentation files.

### Build Documentation

Build static HTML files:

```bash
mkdocs build
```

This creates a `site/` directory with the static website.

## Project Structure

```
docs/
├── index.md                    # Homepage
├── getting-started/
│   ├── quickstart.md          # Quick start guide
│   ├── installation.md        # Installation instructions
│   └── first-scan.md          # First scan tutorial
├── user-guide/
│   ├── basic-usage.md         # Basic usage guide
│   ├── configuration.md       # Configuration reference
│   ├── advanced-features.md   # Advanced features
│   ├── ci-cd-integration.md   # CI/CD integration
│   └── examples.md            # Usage examples
├── reference/
│   ├── cli-options.md         # CLI reference
│   ├── config-file.md         # Config file reference
│   ├── patterns.md            # Detection patterns
│   └── output-formats.md      # Output format specs
├── deployment/
│   ├── github-actions.md      # GitHub Actions guide
│   ├── gitlab-ci.md           # GitLab CI guide
│   ├── jenkins.md             # Jenkins guide
│   └── docker.md              # Docker guide
└── about/
    ├── changelog.md           # Changelog
    ├── release-notes.md       # Release notes
    ├── contributing.md        # Contributing guide
    └── license.md             # License
```

## Deployment

### GitHub Pages

1. Build the documentation:

   ```bash
   mkdocs gh-deploy
   ```

2. This will build and push to the `gh-pages` branch

3. Enable GitHub Pages in repository settings:
   - Go to Settings → Pages
   - Source: Deploy from branch
   - Branch: `gh-pages` → `/root`

### Manual Deployment

1. Build the documentation:

   ```bash
   mkdocs build
   ```

2. Deploy the `site/` directory to your hosting provider

## Writing Documentation

### Markdown Files

All documentation is written in Markdown with additional features from PyMdown Extensions.

### Code Blocks

````markdown
```bash
secscan -root /path/to/project
```
````

````

### Admonitions

```markdown
!!! tip
    This is a helpful tip!

!!! warning
    This is a warning!

!!! info
    This is informational!
````

### Tabs

````markdown
=== "Linux"
`bash
    make install
    `

=== "macOS"
`bash
    brew install secscan
    `
````

## Customization

### Theme

Edit `mkdocs.yml` to customize the theme:

```yaml
theme:
  name: material
  palette:
    primary: blue
    accent: indigo
```

### Navigation

Update the `nav` section in `mkdocs.yml` to change the navigation structure.

## Best Practices

1. **Keep it Simple**: Use clear, concise language
2. **Use Examples**: Include practical code examples
3. **Cross-Reference**: Link related pages
4. **Test Commands**: Verify all code examples work
5. **Update Regularly**: Keep docs in sync with code changes

## Useful Commands

```bash
# Serve docs locally
mkdocs serve

# Build static site
mkdocs build

# Deploy to GitHub Pages
mkdocs gh-deploy

# Clean build directory
rm -rf site/
```

## Troubleshooting

### Port Already in Use

```bash
mkdocs serve -a 127.0.0.1:8001
```

### Build Errors

Clean and rebuild:

```bash
rm -rf site/
mkdocs build
```

## Contributing

When contributing to documentation:

1. Edit files in the `docs/` directory
2. Test locally with `mkdocs serve`
3. Ensure all links work
4. Commit changes
5. Open a pull request

## Resources

- [MkDocs Documentation](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [PyMdown Extensions](https://facelessuser.github.io/pymdown-extensions/)
