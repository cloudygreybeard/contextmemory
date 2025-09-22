# Contributing to ContextMemory v2

## Git Workflow

We use a **fork-pull request workflow** to maintain code quality and enable proper review:

### Branch Strategy

- `main` - Protected production branch
- `feature/*` - Feature development branches  
- `fix/*` - Bug fix branches
- `hotfix/*` - Urgent production fixes
- `docs/*` - Documentation updates

### Development Process

1. **Create Feature Branch**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/your-feature-name
   ```

2. **Development Iteration**
   ```bash
   make dev.iterate  # Build and test
   git add .
   git commit -m "descriptive commit message"
   ```

3. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   # Create PR via GitHub UI
   ```

4. **After Review Approval**
   ```bash
   # Merge via GitHub UI (squash merge preferred)
   git checkout main
   git pull origin main
   git branch -d feature/your-feature-name
   ```

## Commit Messages

Use conventional commit format:
- `feat:` - New features
- `fix:` - Bug fixes  
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Test updates
- `chore:` - Maintenance

## Development Commands

```bash
make setup      # Initialize environment
make build      # Build all components  
make test       # Run tests
make test.cli.coverage  # Run Go tests with coverage
make test.integration   # Run full integration tests
make lint       # Check code quality
make dev.iterate # Full build and test cycle
make validate.release   # Validate release readiness
```

## CI/CD Automated Testing

Our GitHub Actions workflows automatically test all changes:

### Workflows That Run on Your Branch

**🔄 On Push to `feature/*`, `fix/*`, `hotfix/*`:**
- **Test Workflow**: Go unit tests, multi-arch builds, extension testing
- **Build Validation**: Full clean builds, version consistency, integration tests

**📝 On Pull Request to `main`:**
- All test workflows run again
- Comprehensive integration testing across OS matrix
- Coverage reporting and quality checks

### What Gets Tested

- **Go CLI**: Unit tests across Go 1.20 & 1.21, coverage reporting
- **Multi-Architecture**: Linux/macOS/Windows (AMD64/ARM64) + WebAssembly
- **VS Code Extension**: TypeScript compilation, packaging, functionality
- **Integration**: Full system testing with real CLI operations
- **Version Consistency**: Ensure CLI, UI, and root versions match
- **Release Readiness**: Validate that changes are ready for production

### Workflow Status

Check workflow status at: https://github.com/cloudygreybeard/contextmemory/actions

All workflows must pass ✅ before PR can be merged.

## Code Quality

- All code must be TypeScript (UI) or Go (CLI)
- Follow existing patterns and naming conventions
- Add tests for new functionality
- Update documentation as needed
- Ensure `make build` and `make test` pass before PR

## Review Process

1. **Automated checks must pass** ✅
   - All GitHub Actions workflows
   - Coverage requirements met
   - No linting errors
2. **At least one approval required**
3. **Squash merge to main**
4. **Delete feature branch after merge**

---

**Remember**: Never commit directly to `main` - always use feature branches and PRs!
