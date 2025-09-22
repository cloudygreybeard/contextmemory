# ContextMemory Architecture

## Overview

ContextMemory uses a hybrid architecture combining a Go CLI backend with TypeScript frontend components, designed for both standalone CLI usage and VS Code integration.

## Components

### CLI Backend (`cmctl`)
- **Language**: Go with Cobra framework
- **Database**: GORM with `github.com/glebarez/sqlite` (pure Go SQLite)
- **Features**: CRUD operations, search, multiple output formats
- **Distribution**: Single binary, cross-platform

### VS Code Extension
- **Language**: TypeScript
- **Integration**: Command palette, tree view, status bar
- **CLI Interface**: Subprocess calls to `cmctl` binary
- **Optional**: WebAssembly embedding of CLI functionality

### Storage Layer
- **Primary**: File-based JSON storage
- **Index**: SQLite-based indexing for search performance
- **Provider Pattern**: Extensible for future cloud storage backends

## Database Integration

### Cursor Workspace Access
ContextMemory reads Cursor's workspace data directly:

```
~/Library/Application Support/Cursor/User/workspaceStorage/
├── <workspace-hash>/
│   └── state.vscdb          # SQLite database
└── <workspace-hash>/
    └── state.vscdb
```

### Data Schema
Cursor uses a simple key-value schema:
```sql
CREATE TABLE ItemTable (
    key TEXT UNIQUE ON CONFLICT REPLACE, 
    value BLOB
);
```

### GORM Implementation
Type-safe database access using GORM:

```go
type CursorItem struct {
    Key   string `gorm:"column:key;primaryKey"`
    Value string `gorm:"column:value"`
}

func (CursorItem) TableName() string {
    return "ItemTable"
}
```

## WebAssembly Integration

### Compilation
The CLI compiles to WebAssembly using pure Go dependencies:

```bash
GOOS=js GOARCH=wasm go build -o cmctl.wasm
```

This works because `github.com/glebarez/sqlite` has no CGO dependencies.

### Extension Integration
The VS Code extension can optionally embed the WASM binary:

```
ui/src/wasm/
├── cmctl.wasm       # 13MB compiled CLI
├── wasm_exec.js     # 17KB Go runtime
└── wasmService.ts   # TypeScript interface
```

### Usage Patterns
- **Subprocess mode**: Extension calls external `cmctl` binary
- **Embedded mode**: Extension uses embedded WASM for specific operations
- **Hybrid approach**: Combination based on operation type and performance requirements

## Build Pipeline

### Multi-Architecture Support
GoReleaser configuration targets:
- Linux: x86_64, aarch64  
- macOS: x86_64, aarch64
- Windows: x86_64
- WebAssembly: js/wasm

### Dependencies
Key external dependencies:
- `gorm.io/gorm` - ORM framework
- `github.com/glebarez/sqlite` - Pure Go SQLite driver
- `github.com/spf13/cobra` - CLI framework
- `modernc.org/sqlite` - Underlying SQLite implementation (via glebarez)

## Data Flow

### Memory Creation
1. User input (CLI or extension)
2. Content processing and validation
3. AI-assisted naming and labeling (when configured)
4. JSON serialization to file storage
5. Index update for search functionality

### Cursor Import
1. Workspace discovery in Cursor storage directory
2. SQLite database connection (read-only)
3. Query for chat data keys (`aiService.prompts`, `composer.composerData`)
4. JSON parsing and conversation reconstruction
5. Conversion to ContextMemory format with automated labeling

### Search Operations
1. Query parsing and label filtering
2. Full-text search across memory content
3. Index-based result ranking
4. Output formatting (table, JSON, YAML, etc.)

This architecture balances simplicity with extensibility, providing both a robust CLI tool and seamless editor integration.
