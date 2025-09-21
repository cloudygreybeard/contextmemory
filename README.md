# ContextMemory

> Clean, simple, file-based memory management for LLM development workflows

## Vision

ContextMemory focuses on simplicity and core value delivery:

- **File-based storage** - No servers, no databases, no complexity
- **CRUD fundamentals** - Create, Read, Update, Delete memories 
- **Multiple interfaces** - VS Code panel, commands, CLI
- **AI-assisted defaults** - Smart naming and labeling
- **Kubernetes-style** - Name + labels, minimal metadata
- **Stay invisible** - Enhance LLM workflows without friction
- **Modular architecture** - Extensible for future backends

## Core Principles

1. **Simplicity First** - Remove all non-essential complexity
2. **File-Based** - Local storage, easy backup/sync
3. **Automation-Driven** - Makefile-based development workflow
4. **User Agency** - Smart defaults but user control
5. **LLM-Friendly** - Designed for AI-assisted development
6. **Extensible** - Clean abstractions for future growth

## Architecture

```
contextmemory/
├── core/              # Core memory operations (TypeScript)
├── storage/           # File-based storage backend (TypeScript)
├── cmd/cm/            # Go CLI with Cobra framework
│   ├── main.go        # CLI entry point  
│   ├── cmd/           # Cobra commands
│   └── internal/      # Storage and utilities
├── ui/                # VS Code extension UI (TypeScript)
├── docs/              # Documentation
├── Makefile          # Automation workflows
└── README.md
```

### Hybrid Implementation

**TypeScript Components:**
- VS Code extension (required by VS Code API)
- Core interfaces and AI assistant 
- Legacy CLI (being phased out)

**Go Components:**
- Primary CLI tool (`cm`) with Cobra framework
- Single binary distribution
- Fast startup and cross-platform support

## Getting Started

```bash
make setup        # Initialize development environment
make build        # Build all components (Go CLI + TypeScript)
make install.cli  # Install Go CLI to /usr/local/bin/cmctl
make test         # Run tests
make install.ui   # Install VS Code extension locally
```

### CLI Usage

```bash
# Fast Go CLI (recommended)
cmctl create --name "Session Notes" --content "..." --labels "type=session"
cmctl list
cmctl search --query "authentication" --labels "type=code"
cmctl get mem_abc123_def456
cmctl health

# TypeScript CLI (legacy)
node dist/cli/cm.js create --name "test"
```

## Development Status

🚧 **In Development** - v0.6.0

**Current Features:**
- ✅ File-based storage with full CRUD operations
- ✅ Professional CLI (`cmctl`) with kubectl-style output
- ✅ AI-assisted name and label generation  
- ✅ Cross-platform Go binary
- ✅ Extensible provider architecture (file, s3, gcs, remote)

**Next Steps:**
- VS Code extension UI
- Cloud storage providers (S3, GCS)
- Remote API provider

---

*Evolved from earlier prototypes, focused on core value delivery and production readiness.*
