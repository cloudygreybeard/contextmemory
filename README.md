# ContextMemory

> Clean, simple, file-based memory management for LLM development workflows

[![Version](https://img.shields.io/badge/version-0.6.0-blue.svg)](https://github.com/cloudygreybeard/contextmemory)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](https://github.com/cloudygreybeard/contextmemory/blob/main/LICENSE)

## Overview

ContextMemory is a production-ready tool that transforms development conversations into a searchable knowledge base. Store session contexts, code snippets, and development notes with AI-assisted smart defaults.

**Key Features:**
- **File-based storage** - No servers, databases, or external dependencies
- **Professional CLI** - `cmctl` with kubectl-style output and verbosity controls
- **AI-assisted defaults** - Smart name and label generation  
- **Extensible architecture** - Provider system for file, cloud, and remote storage
- **Cross-platform** - Single Go binary, works everywhere
- **LLM-optimized** - Designed for AI-assisted development workflows

## Quick Start

```bash
# Install
git clone https://github.com/cloudygreybeard/contextmemory.git
cd contextmemory
make build && make install.cli

# Basic usage
echo "Meeting notes..." | cmctl create --name "Sprint Planning" --labels "type=meeting,team=eng"
cmctl list
cmctl search --query "authentication" --labels "type=code"
cmctl get <memory-id>

# Provider options
cmctl --provider file health      # Local file storage (default)
cmctl --provider s3 health        # AWS S3 (coming soon)
cmctl --provider gcs health       # Google Cloud (coming soon)
```

## Architecture

```
contextmemory/
├── cmd/cmctl/         # Go CLI with Cobra framework (primary interface)
│   ├── main.go        # CLI entry point  
│   ├── cmd/           # Cobra commands (create, list, search, etc.)
│   └── internal/      # Storage providers and utilities
├── core/              # Core memory operations (TypeScript)
├── storage/           # File-based storage backend (TypeScript)  
├── ui/                # VS Code extension UI (TypeScript, future)
├── Makefile          # Development automation
└── README.md
```

## Architecture

**Production CLI (`cmctl`):**
- Go with Cobra framework for professional command-line experience
- kubectl-style output with verbosity controls (`-v` flag)
- Cross-platform single binary, fast startup
- Extensible provider architecture (file, cloud, remote)

**TypeScript Components:**
- Core memory operations and AI assistant
- VS Code extension (planned)
- Legacy CLI for development/testing

**Storage Providers:**
- **File Provider**: Local filesystem storage (current, production-ready)
- **S3 Provider**: AWS S3 backend (skeleton implemented)
- **GCS Provider**: Google Cloud Storage (skeleton implemented)  
- **Remote Provider**: HTTP API backend (skeleton implemented)

## CLI Reference

### Core Operations

```bash
# Create memories
echo "content" | cmctl create --name "Memory Name" --labels "key=value,type=note"
cmctl create --name "Code Review" --file "./notes.md" --labels "type=review,lang=go"

# List and search
cmctl list                                    # Show all memories
cmctl list --labels "type=meeting"           # Filter by labels
cmctl search --query "authentication"        # Full-text search
cmctl search --labels "type=code,lang=go"    # Search with label filters

# Retrieve and manage
cmctl get <memory-id>                         # Get specific memory
cmctl get <memory-id> --output json          # JSON output
cmctl health                                  # Check system health
cmctl info                                    # Show storage info
```

### Verbosity Controls

```bash
cmctl -v=0 list          # Quiet mode (essential output only)
cmctl -v=1 list          # Normal mode (default)
cmctl -v=2 list          # Verbose mode (debug info)
```

### Provider Selection

```bash
cmctl --provider file health      # Local file storage (default)
cmctl --provider s3 health        # AWS S3 (requires configuration)
cmctl --provider gcs health       # Google Cloud Storage
cmctl --provider remote health    # HTTP API backend
```

## Installation & Development

### Installation

```bash
# Clone and build
git clone https://github.com/cloudygreybeard/contextmemory.git
cd contextmemory
make build

# Install CLI globally  
make install.cli    # Installs cmctl to /usr/local/bin/

# Verify installation
cmctl --version     # Should show: cmctl version 0.6.0
```

### Development Workflow

```bash
# Setup development environment
make setup          # Initialize dependencies and structure
make build          # Build Go CLI + TypeScript components  
make install.cli    # Install CLI locally
make info.status    # Check project status

# Development iteration
make dev.iterate    # Quick build and install cycle
```

### Configuration

ContextMemory uses `~/.contextmemory/` for storage and configuration:

```bash
~/.contextmemory/
├── config.yaml     # Provider and CLI configuration
├── memories/       # JSON files for each memory
└── index.json      # Search index and metadata
```

## Features

✅ **Production Ready:**
- File-based storage with full CRUD operations
- Professional CLI (`cmctl`) with kubectl-style output  
- AI-assisted name and label generation
- Cross-platform Go binary (macOS, Linux, Windows)
- Extensible provider architecture
- Comprehensive error handling and validation

🚧 **Planned:**
- VS Code extension with integrated UI
- AWS S3 and Google Cloud Storage providers
- Remote HTTP API provider
- Advanced search and filtering
- Team collaboration features

---

**Version 0.6.0** - Production-ready file-based memory management for LLM workflows.
