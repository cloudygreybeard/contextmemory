# ContextMemory

A Model Context Protocol (MCP) server implementation for session context management, providing AI assistants with structured access to persistent memory storage.

## Overview

ContextMemory enables AI assistants to store, retrieve, and manage session context through a standardized JSON-RPC interface. This implementation provides both programmatic access via the MCP protocol and human administration through a command-line interface.

**Status**: Initial MVP exploring MCP integration patterns and cloud-native resource management.

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

### Protocol Operations

The MCP server implements seven core operations:

- `memories/v1/list` - List memories with optional filtering
- `memories/v1/get` - Retrieve specific memory by identifier
- `memories/v1/create` - Create new memory with metadata
- `memories/v1/update` - Replace memory content (PUT semantics)
- `memories/v1/patch` - Modify specific memory fields (PATCH semantics)
- `memories/v1/delete` - Remove memory by identifier
- `memories/v1/search` - Query memories by content and labels

## Features

### MCP Server
- JSON-RPC 2.0 protocol compliance
- Stdio-based communication for AI client integration
- Namespace support for resource organization
- Label-based filtering and search
- Index-optimized operations for performance
- Comprehensive error handling

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

### Build from Source

```bash
# Clone repository
git clone https://github.com/your-org/contextmemory.git
cd contextmemory

# Build MCP server
go build -o build/mcp-server ./cmd/mcp-server

# Build CLI tool
go build -o build/memctl ./cmd/memctl

# Install binaries (optional)
sudo cp build/mcp-server /usr/local/bin/
sudo cp build/memctl /usr/local/bin/
```

### Requirements
- Go 1.21 or later
- No external runtime dependencies

## Quick Start

### Using the MCP Server

The MCP server communicates via JSON-RPC over stdio:

```bash
# Start MCP server
./build/mcp-server

# Example request (from AI client)
echo '{"jsonrpc":"2.0","id":1,"method":"memories/v1/list","params":{}}' | ./build/mcp-server
```

### Using the CLI Tool

```bash
# List all memories
memctl get

# Create a new memory
echo "Meeting notes content" | memctl create --name "Team Meeting" --labels "type=notes,date=2025-01-01"

# Search memories
memctl search --query "meeting" --labels "type=notes"

# Get memory in YAML format
memctl get mem_abc123_def456 -o yaml

# Show help
memctl --help
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

### Advanced Queries

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

**Current State**: Functional MVP suitable for experimentation and development workflows.

**Planned Improvements**:
- Cloud storage provider implementations
- Enhanced query capabilities
- Metrics and observability
- Performance optimizations
- Extended MCP protocol features

This project explores patterns for AI memory management and MCP integration. Feedback and contributions are welcome as we develop these concepts further.
