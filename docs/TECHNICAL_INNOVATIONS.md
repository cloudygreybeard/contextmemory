# Technical Architecture

## Database Integration

ContextMemory uses [GORM](https://gorm.io/) with a pure Go SQLite driver for accessing Cursor's chat database. This approach provides several practical advantages:

### Database Access Implementation

GORM with `github.com/glebarez/sqlite` (pure Go implementation) enables structured database access:

```go
type CursorItem struct {
    Key   string `gorm:"column:key;primaryKey"`
    Value string `gorm:"column:value"`
}

// TableName specifies the table name for CursorItem
func (CursorItem) TableName() string {
    return "ItemTable"
}

// Query Cursor's database using GORM
db, err := gorm.Open(sqlite.Open(dbPath+"?mode=ro"), &gorm.Config{})
var item CursorItem
result := db.Where("key = ?", "aiService.prompts").First(&item)
```

This approach provides type-safe database queries while avoiding manual SQL construction.

### Pure Go SQLite Implementation

Using `github.com/glebarez/sqlite` (which utilizes `modernc.org/sqlite` internally) instead of the default GORM SQLite driver provides:

- No CGO dependencies for the CLI
- Simplified cross-platform builds
- Reduced external dependencies
- Excellent performance for database operations

*Note: While pure Go, the underlying `modernc.org/libc` dependency currently prevents WebAssembly compilation.*

## Current Architecture: Subprocess Integration

### VS Code Extension Architecture

The current implementation uses a decoupled approach with proven reliability:

```
Current Architecture:
├── cmctl (Go binary)           # Standalone CLI tool
│   ├── CRUD operations
│   ├── Cursor database access
│   └── Cross-platform builds
├── VS Code Extension (TypeScript)
│   ├── UI commands
│   ├── Tree view provider
│   └── Subprocess calls to cmctl
```

Benefits of this approach:
- ✅ **Reliability**: Proven subprocess communication patterns
- ✅ **Separation**: Clear boundaries between CLI and UI components  
- ✅ **Flexibility**: Independent development and deployment cycles
- ✅ **Simplicity**: No complex runtime integration

### WebAssembly Integration: Future Roadmap

Future exploration could investigate closer integration through WebAssembly:

**Potential Benefits:**
- Zero-installation experience (embedded CLI)
- Reduced subprocess overhead
- Single-package distribution
- Direct in-process function calls

**Current Blockers:**
- `modernc.org/libc` dependency excludes `js/wasm` targets
- SQLite integration complexity in WASM environment

**Possible Solutions for Future Investigation:**
1. **Alternative SQLite approach**: Research WASM-native SQLite implementations
2. **Hybrid architecture**: Core functions in WASM, SQLite operations via subprocess
3. **Dependency evolution**: Monitor `modernc.org/libc` for potential WASM support

**Timeline**: Post-v1.0 enhancement once core functionality is stable

## Cursor Integration

### Database Access

ContextMemory reads Cursor's workspace storage directly from local SQLite databases:

- **Location**: `~/Library/Application Support/Cursor/User/workspaceStorage/`
- **Format**: Each workspace contains a `state.vscdb` file
- **Access**: Read-only queries using GORM with pure Go SQLite

### Chat Data Parsing

The implementation handles multiple data formats found in Cursor's database:

```go
// Multiple possible chat data keys
chatKeys := []string{
    "aiService.prompts",           // AI service interactions
    "composer.composerData",       // Composer conversations  
    "workbench.panel.aichat.view.aichat.chatdata", // Legacy format
}
```

### Data Processing

Chat data is converted to structured markdown with automated labeling:

- **Smart naming**: Extracts topics from conversation content
- **Technology detection**: Identifies programming languages and frameworks
- **Activity classification**: Categorizes sessions (debugging, implementation, etc.)

## Build and Distribution

### Multi-Architecture Support

The build process targets multiple platforms for broad compatibility:

**Production Platforms:**
- Linux: x86_64, aarch64
- macOS: x86_64, aarch64 (Apple Silicon)
- Windows: x86_64

**Future Platforms:**
- WebAssembly: Planned for closer VS Code integration

### Continuous Integration

GoReleaser with GitHub Actions provides automated builds across the production target matrix, ensuring consistent releases for all supported platforms.
