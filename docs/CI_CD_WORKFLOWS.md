# CI/CD Workflows

This document details our GitHub Actions workflows for automated testing, building, and releasing.

## Overview

ContextMemory uses a comprehensive CI/CD pipeline to ensure code quality and automate releases across multiple platforms and package formats.

## Workflows

### 1. Test Workflow (`test.yml`)

**Triggers:**
- Push to `main`, `feature/*`, `fix/*`, `hotfix/*`
- Pull requests to `main`

**Jobs:**

#### test-go
- **Matrix**: Go 1.20, 1.21
- **Steps**: Checkout, Go setup, test with coverage, golangci-lint
- **Coverage**: Reports to Codecov

#### test-multiarch-build  
- **Platforms**: Linux/macOS/Windows (AMD64/ARM64) + WebAssembly
- **Purpose**: Validate cross-platform compatibility

#### test-extension
- **Environment**: Node.js 18
- **Steps**: Build and package VS Code extension
- **Artifacts**: `.vsix` package

#### integration-test
- **Matrix**: ubuntu-latest, windows-latest, macos-latest
- **Purpose**: Full system testing with real CLI operations

### 2. Build Validation Workflow (`build-validation.yml`)

**Purpose**: Comprehensive build verification and version consistency

**Jobs:**

#### validate-build
- Clean build from scratch
- CLI version verification
- Extension packaging validation
- Integration test execution
- Release readiness validation

#### check-version-consistency
- Validates CLI, root, and UI versions match
- Prevents version mismatches in releases

### 3. Release Workflow (`release.yml`)

**Triggers:**
- Git tags matching `v*` pattern
- Manual workflow dispatch

**Jobs:**

#### goreleaser
- **Multi-platform builds**: GoReleaser handles CLI binaries
- **Extension packaging**: Creates VS Code `.vsix` packages
- **Release assets**: Uploads all artifacts to GitHub release

#### publish-extension
- **Marketplace**: Publishes to VS Code Marketplace
- **Dependency**: Runs after successful GoReleaser job

## Caching Strategy

All workflows use optimized caching:

- **Go modules**: `cache-dependency-path: cmd/cmctl/go.sum`
- **Node.js**: `cache: 'npm'` with multiple `package-lock.json` paths
- **QEMU**: For cross-architecture builds

## Quality Gates

### Required Checks
- All unit tests pass
- Integration tests pass across OS matrix
- Multi-architecture builds succeed
- Extension packages successfully
- Version consistency maintained
- Linting passes with no errors

### Coverage Requirements
- Go tests report coverage to Codecov
- Coverage trends tracked over time
- Prevents coverage regression

## Artifacts Produced

### On every push/PR:
- Test coverage reports
- Build logs and validation results

### On releases (tags):
- **CLI binaries**: Multiple platforms via GoReleaser
- **VS Code extension**: `.vsix` package
- **GitHub release**: With all assets attached
- **Marketplace**: Automatic VS Code Marketplace publishing

## Debugging Workflow Issues

1. **Check workflow status**: https://github.com/cloudygreybeard/contextmemory/actions
2. **Review specific job logs** for detailed error information
3. **Local reproduction**:
   ```bash
   make test.cli.coverage  # Reproduce Go test issues
   make test.integration   # Reproduce integration failures
   make build              # Reproduce build issues
   ```

## Security

- **Secrets management**: Uses GitHub secrets for marketplace publishing
- **Permissions**: Minimal required permissions per job
- **Dependency pinning**: Actions pinned to specific versions

## Maintenance

Workflows are automatically maintained through:
- Dependabot updates for actions
- Regular review of security advisories
- Performance optimization based on build times

---

For questions or issues with CI/CD workflows, please open an issue with the `ci/cd` label.
