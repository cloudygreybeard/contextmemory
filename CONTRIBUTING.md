# Contributing to ContextMemory v2

## Git Workflow

We use a **fork-pull request workflow** to maintain code quality and enable proper review:

### Branch Strategy

- `main` - Protected production branch
- `feature/*` - Feature development branches  
- `fix/*` - Bug fix branches
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
make dev.iterate # Full build and test cycle
```

## Code Quality

- All code must be TypeScript
- Follow existing patterns and naming conventions
- Add tests for new functionality
- Update documentation as needed
- Ensure `make build` passes before PR

## Review Process

1. Automated checks must pass
2. At least one approval required
3. Squash merge to main
4. Delete feature branch after merge

---

**Remember**: Never commit directly to `main` - always use feature branches and PRs!
