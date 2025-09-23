# Makefile Naming Conventions

## Overview

ContextMemory follows a hierarchical naming convention for make targets using the pattern `<functional-area>.<action>` to ensure consistency and discoverability.

## Naming Convention

### **Hierarchical Targets**

All specific functionality uses the dot notation pattern:

```
<functional-area>.<action>[.<qualifier>]
```

#### **Build System**
- `build.cli` - Build CLI for current platform
- `build.cli.wasm` - Build CLI for WebAssembly
- `build.cli.all` - Build CLI for all architectures
- `build.ui` - Build VS Code extension
- `build.core` - Build core TypeScript components
- `build.storage` - Build storage backend

#### **Testing System**
- `test.cli` - Test CLI functionality
- `test.cli.coverage` - Test CLI with coverage
- `test.cli.functional` - Run CLI functional tests
- `test.ui` - Test UI components
- `test.core` - Test core functionality
- `test.storage` - Test storage backend
- `test.integration` - Run integration tests
- `test.all` - Run all tests
- `test.all.coverage` - Run all tests with coverage

#### **Installation System**
- `install.cli` - Install CLI binary
- `install.ui` - Install VS Code extension

#### **Package System**
- `package.ui` - Create VS Code extension package

#### **Version Management**
- `version.sync` - Sync all versions from git tag
- `version.validate` - Check version consistency
- `version.bump-patch` - Bump patch version (0.7.0 → 0.7.1)
- `version.bump-minor` - Bump minor version (0.7.0 → 0.8.0)
- `version.bump-major` - Bump major version (0.7.0 → 1.0.0)

#### **Release System**
- `release.snapshot` - Create snapshot release
- `release.validate` - Run release validation

#### **Development System**
- `dev.start` - Start development environment
- `dev.stop` - Stop development environment
- `dev.iterate` - Full build and install cycle

#### **Information System**
- `info.status` - Show project status
- `info.structure` - Show project structure

### **Top-Level Convenience Targets**

Common workflow targets remain as simple names for ease of use:

- `help` - Show help (default target)
- `setup` - Initialize development environment
- `build` - Build all components
- `test` - Run all tests
- `clean` - Clean build artifacts
- `install` - Install all components
- `package` - Create packages (alias for package.ui)
- `release` - Create release
- `lint` - Check code quality

## Usage Examples

### **Development Workflow**
```bash
make clean                  # Clean environment
make build                  # Build everything
make install                # Install everything
make test                   # Test everything
```

### **Specific Component Work**
```bash
make build.cli              # Build only CLI
make test.cli.coverage      # Test CLI with coverage
make package.ui             # Package only extension
```

### **Version Management**
```bash
make version.validate       # Check version consistency
make version.sync           # Sync from git tags
make version.bump-patch     # Bump to next patch version
```

### **Release Process**
```bash
make release.validate       # Validate release readiness
make release.snapshot       # Create test release
make release                # Create production release
```

## Benefits

1. **Discoverability**: Tab completion works with dot notation
2. **Consistency**: Related functionality grouped under same prefix
3. **Scalability**: New targets can be added under existing functional areas
4. **Clarity**: Purpose is clear from the name
5. **Hierarchy**: Logical organization of related functionality

## Guidelines for New Targets

1. Use `<functional-area>.<action>` pattern for specific functionality
2. Maintain convenience aliases for common workflows at top level
3. Use dots to create sub-categories when needed (e.g., `test.cli.coverage`)
4. Use descriptive action names rather than abbreviations
5. Group related functionality under the same functional area prefix
