# ContextMemory

> Clean, simple, file-based memory management for LLM development workflows

[![Version](https://img.shields.io/badge/version-0.6.1-blue.svg)](https://github.com/cloudygreybeard/contextmemory)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](https://github.com/cloudygreybeard/contextmemory/blob/main/LICENSE)

## Overview

ContextMemory transforms development conversations into a searchable knowledge base. Store session contexts, code snippets, and development notes with AI-assisted smart defaults.

**Key Features:**
- **Multiple storage options** - File-based (current), cloud storage (planned), remote APIs (planned)
- **Professional CLI** - `cmctl` with flexible output formats and verbosity controls
- **AI-assisted defaults** - Smart name and label generation  
- **Extensible architecture** - Provider system for different storage backends
- **Cross-platform** - Single Go binary for CLI operations
- **IDE integration** - VS Code extension for seamless workflow integration

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

# Advanced features
cmctl list --show-id              # Display memory IDs
cmctl list -o json                # JSON output for scripting
cmctl delete --labels "type=test" # Delete memories by criteria
cmctl search -q "auth" -o yaml    # Search with YAML output

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
├── ui/                # VS Code extension (TypeScript)
├── Makefile          # Development automation
└── README.md
```

## Architecture

**Production CLI (`cmctl`):**
- Go with Cobra framework for professional command-line experience
- Multiple output formats (table, JSON, YAML, JSONPath, Go templates)
- Flexible verbosity controls (`-v` flag) and configuration options
- Cross-platform single binary, fast startup
- Extensible provider architecture (file, cloud, remote)

**TypeScript Components:**
- Core memory operations and AI assistant
- VS Code extension with tree view and command integration
- Legacy CLI for development/testing

**Storage Providers:**
- **File Provider**: Local filesystem storage (current, fully functional)
- **S3 Provider**: AWS S3 backend (planned)
- **GCS Provider**: Google Cloud Storage (planned)  
- **Remote Provider**: HTTP API backend (planned)

## CLI Reference

### Core Operations

```bash
# Create memories
echo "content" | cmctl create --name "Memory Name" --labels "key=value,type=note"
cmctl create --name "Code Review" --file "./notes.md" --labels "type=review,lang=go"

# List and search
cmctl list                                    # Show all memories
cmctl list --show-id                         # Include memory IDs
cmctl list --labels "type=meeting"           # Filter by labels
cmctl search --query "authentication"        # Full-text search
cmctl search --labels "type=code,lang=go"    # Search with label filters

# Retrieve and manage
cmctl get <memory-id>                         # Get specific memory
cmctl get <memory-id> -o json                # JSON output
cmctl delete <memory-id>                     # Delete specific memory
cmctl delete --labels "type=test"           # Delete by criteria
cmctl delete --all                          # Delete all memories
cmctl health                                 # Check system health
cmctl info                                   # Show storage info
```

### Output Formats

```bash
# Multiple output formats for scripting and data extraction
cmctl list -o json                          # JSON format
cmctl list -o yaml                          # YAML format
cmctl list -o jsonpath='{.items[*].name}'   # Extract specific fields
cmctl get mem_123 -o go-template='{{.spec.content}}'  # Custom templates

# Advanced JSONPath examples
cmctl list -o jsonpath='{.items[?(@.labels.type=="test")].name}'  # Filter results
cmctl search -q "auth" -o jsonpath='{.items[*].id}'               # Get matching IDs
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

## VS Code Extension

The extension integrates ContextMemory into your development workflow:

### Installation

```bash
# Build and package extension
make build.ui && make package.ui

# Install in VS Code
code --install-extension ui/contextmemory-0.6.0.vsix

# Or install in Cursor  
cursor --install-extension ui/contextmemory-0.6.0.vsix
```

### Extension Features

**Command Palette:**
- `ContextMemory: Create Memory` - Create new memory
- `ContextMemory: Create Memory from Selection` - Create from selected code
- `ContextMemory: Create Memory from Current Chat` - Create from chat/markdown file
- `ContextMemory: Search Memories` - Search existing memories
- `ContextMemory: List All Memories` - Browse all memories
- `ContextMemory: Delete Memory` - Delete specific memory
- `ContextMemory: Delete Memories by Labels` - Bulk delete by criteria
- `ContextMemory: Delete All Memories` - Delete all memories (with confirmation)
- `ContextMemory: Open Configuration` - Manage extension settings
- `ContextMemory: Check Health` - Verify CLI connectivity

**Tree View:**
- Browse memories by category (Recent, Type, Language, Project)
- Click to open memory in editor
- Automatic categorization based on labels

**Context Menus:**
- Right-click selected code → "Create Memory from Selection"
- Right-click .md files → "Create Memory from Current Chat"
- Right-click memories in tree view → "Delete Memory"

**Configuration:**
```json
{
  "contextmemory.cliPath": "cmctl",
  "contextmemory.storageDir": "~/.contextmemory",
  "contextmemory.provider": "file", 
  "contextmemory.verbosity": 1,
  "contextmemory.autoSuggestLabels": true,
  "contextmemory.showMemoryIds": false,
  "contextmemory.defaultLabels": []
}
```

All settings can be configured through the VS Code settings UI or via the 
`ContextMemory: Open Configuration` command for an interactive editor.

## Installation & Development

### Installation

```bash
# Clone and build
git clone https://github.com/cloudygreybeard/contextmemory.git
cd contextmemory
make build

# Install CLI globally  
make install.cli    # Installs cmctl to /usr/local/bin/

# Build and install VS Code extension
make build.ui       # Compile TypeScript extension  
make package.ui     # Create .vsix package
make install.ui     # Install in VS Code/Cursor

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

# Extension development  
make build.ui       # Compile extension TypeScript
make package.ui     # Create installable .vsix package

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

✅ **Current (v0.6.0):**
- File-based storage with full CRUD operations
- Professional CLI (`cmctl`) with multiple output formats (JSON, YAML, JSONPath, Go templates)
- Memory deletion with flexible criteria (ID, labels, bulk operations)
- Optional memory ID display for advanced operations
- AI-assisted name and label generation
- Cross-platform Go binary (macOS, Linux, Windows)
- VS Code extension with comprehensive management capabilities
- Interactive configuration editor for all settings
- Extensible provider architecture foundation

🚧 **Planned:**
- Cloud storage providers (AWS S3, Google Cloud Storage)
- Remote HTTP API provider
- Enhanced VS Code extension features
- Advanced search and filtering capabilities
- Team collaboration features

---

**Version 0.6.1** - Enhanced with deletion capabilities, output formats, configuration editor, and improved documentation.
