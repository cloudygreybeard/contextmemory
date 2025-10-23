# ContextMemory

A Model Context Protocol (MCP) server implementation for conversation context management, providing AI assistants with automated context preservation and persistent conversation memory.

## Overview

ContextMemory enables AI assistants to preserve and manage conversation context beyond standard token limits through automation and persistent memory storage. The system provides conversation monitoring, automated context preservation, and cross-session conversation threading.

**Status**: Functional Phase 2 automation system with context preservation capabilities.

## Key Features

- **Context Preservation**: Automatically preserve conversation history before context truncation
- **Conversation Monitoring**: Monitor conversation health with token utilization and recommendations  
- **Incremental Snapshots**: Lightweight context capture between full checkpoints
- **Conversation Threading**: Cross-session relationship mapping and conversation organization
- **Automated Triggers**: Token and message-based triggers for context management
- **Context Retention**: Preserve conversation history beyond summarization
- **Installation**: Single-command installation with automatic IDE configuration
- **Organization**: Label-based memory organization and search capabilities

## Documentation

- **[Installation Guide](docs/INSTALLATION_GUIDE.md)** - Complete setup instructions for all platforms
- **[Automation System](docs/AUTOMATION_SYSTEM.md)** - Intelligent automation and monitoring features  
- **[Distribution Strategy](docs/DISTRIBUTION_STRATEGY.md)** - Binary, SaaS, and hybrid deployment options
- **[Automation Triggers](docs/AUTOMATION_TRIGGERS.md)** - Manual, polling, and event-driven trigger mechanisms
- **[Conversation Persistence Design](docs/CONVERSATION_PERSISTENCE_DESIGN.md)** - Complete data architecture and retrieval patterns

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Documentation](#documentation)
- [Architecture](#architecture)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Examples](#examples)
- [Development](#development)
- [Contributing](#contributing)
- [Status and Roadmap](#status-and-roadmap)
- [License](#license)

## Architecture

### Components

- **MCP Server** (`mcp-server`): JSON-RPC 2.0 server implementing the Model Context Protocol over stdio
- **CLI Tool** (`memctl`): Human-friendly command-line interface for memory administration
- **Storage Layer**: File-based storage with JSON metadata and indexing
- **Provider System**: Pluggable storage backend architecture (file provider implemented)

### Resource Model

Resources follow standard cloud-native patterns with `apiVersion`, `kind`, `metadata`, and `spec` fields:

```yaml
apiVersion: contextmemory.io/v1
kind: Memory
metadata:
  name: session-notes
  namespace: default
  uid: mem_abc123_def456
  labels:
    type: session
    project: api
spec:
  content: "Session context content..."
```

### MCP Tools

The MCP server implements comprehensive memory and automation tools:

**Core Memory Management:**
- `contextmemory_get` - List all memories or get specific memory (kubectl-style interface)
- `contextmemory_create` - Create new memory with metadata and labels
- `contextmemory_update` - Update existing memory content
- `contextmemory_delete` - Delete memory by name
- `contextmemory_search` - Search memories by content or labels

**Phase 2 Automation:**
- `contextmemory_checkpoint` - Create conversation checkpoint to preserve context before truncation
- `contextmemory_snapshot` - Create incremental snapshot for lightweight context capture
- `contextmemory_monitor` - Monitor conversation health and get automated recommendations
- `contextmemory_thread` - Create and manage conversation threads for relationship tracking

## Features

### Core Memory Management
- JSON-RPC 2.0 protocol compliance via Model Context Protocol
- Stdio-based communication for AI client integration
- Namespace support for resource organization and project isolation
- Label-based filtering and search capabilities
- Index-optimized operations for memory access

### Phase 2 Automation
- Conversation monitoring with token count analysis
- Health assessment (healthy/warning/critical status)
- Automated context preservation with threshold-based triggers
- Incremental snapshots for context capture between checkpoints
- Full conversation checkpoints preserving conversation history
- Conversation threading enabling cross-session relationship mapping
- Recommendations for preservation strategies

### CLI Tool (`memctl`)
- Command structure inspired by container orchestration tools
- Multiple output formats: table, JSON, YAML, JSONPath, Go templates
- Label selector queries
- Namespace-aware operations
- Auto-discovery of MCP server binary

### Storage
- File-based persistence with JSON metadata
- Fast metadata-only operations via indexing
- Configurable storage directory
- Health check capabilities

## Installation

**Quick Start:**
```bash
# Single-command installation
curl -sSL https://install.contextmemory.dev | bash

# Configure Cursor IDE
contextmemory config cursor

# Usage: @contextmemory monitor
```

**See the [Installation Guide](docs/INSTALLATION_GUIDE.md)** for:
- Manual installation instructions
- Package manager options (Homebrew, APT, Chocolatey, etc.)
- Multi-platform support (macOS, Linux, Windows)
- IDE configuration for Cursor, VS Code, and others
- Troubleshooting and verification steps

### Build from Source

```bash
# Clone repository
git clone https://github.com/your-org/contextmemory.git
cd contextmemory

# Build enhanced MCP server with automation
go build -o build/mcp-server-sdk ./cmd/mcp-server-sdk

# Install to system path
sudo cp build/mcp-server-sdk /usr/local/bin/contextmemory

# Initialize configuration
contextmemory init
```

### Requirements
- Go 1.24 or later
- No external runtime dependencies

## Quick Start

### Using with Cursor IDE

After installation, use ContextMemory in Cursor conversations:

```bash
# Monitor conversation health
@contextmemory monitor --conversation-id "my_project_2025" --message-count 15 --estimated-tokens 12000

# Result: Status: healthy, Token Utilization: 12%, Recommendation: Continue conversation

# Create checkpoint to preserve full conversation
@contextmemory checkpoint --conversation-id "my_project_2025" --title "Project Planning Session" --phase "planning"

# Create incremental snapshot for recent context
@contextmemory snapshot --conversation-id "my_project_2025" --recent-content "Recent discussion about API design..." --messages-since 5

# List conversation history
@contextmemory get --output-format table --limit 10

# Search across conversations
@contextmemory search --query "database design" --labels "project=my_project"
```

### Automation Examples

```bash
# Create conversation thread with relationships
@contextmemory thread --conversation-id "implementation_session" --project "contextmemory" --parent-thread "planning_session"

# Monitor multiple conversation aspects
@contextmemory monitor --conversation-id "current" --message-count 25 --estimated-tokens 85000
# Result: Status: warning, Recommendation: Create incremental snapshot - approaching warning threshold

# Create snapshot following recommendation
@contextmemory snapshot --conversation-id "current" --trigger-reason "threshold-warning" --auto-generated true
```

### Using the CLI Tool (Alternative)

For direct command-line usage without IDE integration:

```bash
# List all memories in table format
contextmemory get --output-format table

# Create memory from command line
contextmemory create --name "Architecture Notes" --content "System design decisions..." --labels "type=docs,project=api"

# Search with filters
contextmemory search --query "authentication" --labels "type=docs" --include-content
```

## Configuration

### Storage Directory

Default storage location: `~/.contextmemory`

Override with:
- Environment: `CONTEXTMEMORY_STORAGE_DIR`
- CLI flag: `--storage-dir`
- Config file: `~/.contextmemory/config.yaml`

### Namespace Configuration

```bash
# Use specific namespace
memctl get --namespace production

# Set default namespace
export CONTEXTMEMORY_NAMESPACE=development
```

### Output Formats

```bash
# Table format (default)
memctl get

# JSON output
memctl get -o json

# Extract specific fields with JSONPath
memctl get -o jsonpath='{.items[*].metadata.name}'

# YAML format
memctl get -o yaml
```

## Examples

### Basic Memory Management

```bash
# Create memory from file
memctl create --name "API Documentation" < api-notes.md --labels "type=docs,component=api"

# Update memory content
memctl patch mem_123 --content "Updated content"

# Delete memory
memctl delete mem_123

# List memories with IDs shown
memctl get --show-id
```

### Query Examples

```bash
# Find memories by type
memctl get --labels "type=session"

# Search content with metadata-only results
memctl search --query "authentication" --no-content

# Combine search criteria
memctl search --query "API" --labels "type=docs" --limit 5
```

### Integration Example

```python
import subprocess
import json

def query_memories(query):
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "memories/v1/search",
        "params": {"query": query, "includeContent": True}
    }
    
    result = subprocess.run(
        ["mcp-server"],
        input=json.dumps(request),
        text=True,
        capture_output=True
    )
    
    return json.loads(result.stdout)
```

## Development

### Project Structure

```
contextmemory/
├── cmd/
│   ├── mcp-server/          # MCP server implementation
│   └── memctl/              # CLI tool
├── internal/
│   ├── client/              # MCP client library
│   ├── output/              # Output formatting
│   ├── providers/           # Storage providers
│   ├── storage/             # Storage implementation
│   └── utils/               # Shared utilities
├── build/                   # Build artifacts
└── go.mod
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

### Adding Storage Providers

The provider system supports pluggable storage backends:

```go
type StorageProvider interface {
    Create(req storage.CreateMemoryRequest) (*storage.Memory, error)
    Get(id string) (*storage.Memory, error)
    Update(req storage.UpdateMemoryRequest) (*storage.Memory, error)
    Delete(id string) error
    List() ([]storage.Memory, error)
    Search(req storage.SearchRequest) (*storage.SearchResponse, error)
}
```

Current implementations:
- File provider (default)
- Cloud providers (planned)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with appropriate tests
4. Ensure all tests pass
5. Submit a pull request

### Code Standards
- Follow Go conventions and formatting
- Use descriptive variable and function names
- Include tests for new functionality
- Update documentation for user-facing changes

## License

Licensed under the Apache License, Version 2.0. See LICENSE file for details.

## Status and Roadmap

**Current State**: Functional Phase 2 automation system with context preservation, conversation monitoring, and automated memory management.

### Phase 2 Complete (Automation)
- Conversation health monitoring and analysis
- Token and message-based preservation triggers  
- Context preservation via checkpoints and snapshots
- Cross-session conversation threading and organization
- Automated preservation recommendations

### Phase 2.1 (Planned Enhancements)
- Threshold-based notifications
- Background monitoring options
- Extended search capabilities
- Configuration management

### Phase 3 (Future Development)
- Event-driven IDE integration
- Pattern-based automation
- Usage analytics and insights
- Team collaboration features

### Distribution Options
- Binary distribution (current) - Local installation and storage
- SaaS Platform - Cloud-based deployment option
- Hybrid Architecture - Combined local and cloud functionality

See the [Distribution Strategy](docs/DISTRIBUTION_STRATEGY.md) for implementation details.
