# Homebrew Formula Setup

This document details the setup and maintenance of the ContextMemory Homebrew formula.

## Overview

ContextMemory uses GoReleaser to automatically generate and maintain a Homebrew formula for the CLI tool (`cmctl`). The formula is published to a custom tap at `cloudygreybeard/homebrew-contextmemory`.

## Setup Requirements

### 1. GitHub Repository

Create the Homebrew tap repository:
```bash
# Create new repository on GitHub
# Repository name: homebrew-contextmemory
# Owner: cloudygreybeard
# Public repository
```

### 2. GitHub Token Configuration

Generate a Personal Access Token with the following permissions:
- `repo` (Full control of private repositories)
- `public_repo` (Access public repositories)

Add the token to GitHub Secrets:
- Repository: `cloudygreybeard/contextmemory`
- Secret name: `HOMEBREW_TAP_GITHUB_TOKEN`
- Secret value: `<generated-token>`

### 3. GoReleaser Configuration

The `.goreleaser.yml` file includes the `brews` section for automatic formula generation:

```yaml
brews:
  - name: cmctl
    ids:
      - standard
    homepage: https://github.com/cloudygreybeard/contextmemory
    description: "Session context management CLI for LLM development workflows"
    license: "Apache-2.0"
    
    tap:
      owner: cloudygreybeard
      name: homebrew-contextmemory
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    
    folder: Formula
    commit_author:
      name: contextmemory-bot
      email: bot@contextmemory.dev
    commit_msg_template: "Brew formula update for {{ .ProjectName }} version {{ .Tag }}"
    
    dependencies:
      - name: git
    
    test: |
      system "#{bin}/cmctl", "--version"
    
    caveats: |
      To complete your ContextMemory setup, install the VS Code extension:
        cursor --install-extension cloudygreybeard.contextmemory
      
      Or download from VS Code Marketplace:
        https://marketplace.visualstudio.com/items?itemName=cloudygreybeard.contextmemory
```

## User Installation Workflow

Once the tap is set up, users can install ContextMemory via Homebrew:

```bash
# Add the tap
brew tap cloudygreybeard/contextmemory

# Install the CLI
brew install cmctl

# Install the VS Code extension (manual step)
cursor --install-extension cloudygreybeard.contextmemory
```

## Automatic Updates

The formula is automatically updated on each tagged release:

1. Developer creates a new git tag (e.g., `v0.7.0`)
2. GitHub Actions triggers the release workflow
3. GoReleaser builds the CLI binaries
4. GoReleaser generates the Homebrew formula
5. Formula is committed to `homebrew-contextmemory` repository
6. Users receive updates via `brew upgrade`

## Testing

### Local Testing

Test the GoReleaser configuration without publishing:

```bash
# Dry run (requires GoReleaser installed)
goreleaser release --snapshot --clean

# Check generated formula
ls dist/homebrew/
```

### Formula Testing

Test the generated formula:

```bash
# Install from local tap (after formula generation)
brew install --build-from-source cloudygreybeard/contextmemory/cmctl

# Test formula functionality
cmctl --version
cmctl health
```

## Maintenance

### Formula Updates

Formula updates are automatic, but manual intervention may be needed for:
- Dependency changes
- Installation path modifications
- Test script updates

### Troubleshooting

Common issues and solutions:

1. **Token Permission Errors**
   - Verify `HOMEBREW_TAP_GITHUB_TOKEN` has correct permissions
   - Check token expiration date

2. **Formula Generation Failures**
   - Validate `.goreleaser.yml` syntax
   - Ensure `homebrew-contextmemory` repository exists and is accessible

3. **Installation Failures**
   - Check binary compatibility
   - Verify archive format and contents

## Directory Structure

The tap repository structure:
```
homebrew-contextmemory/
├── Formula/
│   └── cmctl.rb          # Generated formula
├── README.md             # Tap documentation
└── .github/
    └── workflows/        # Optional CI for formula testing
```

## Related Documentation

- [GoReleaser Homebrew Documentation](https://goreleaser.com/customization/homebrew/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Creating Homebrew Taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

