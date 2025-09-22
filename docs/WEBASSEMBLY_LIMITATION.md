# WebAssembly Compilation Limitation

## Issue Summary

We have confirmed that the `modernc.org/libc` package (a dependency of `github.com/glebarez/sqlite`) does not support the `js/wasm` target, preventing WebAssembly compilation of our CLI while maintaining full SQLite functionality.

## Validation Results

### ✅ Standard Compilation & Runtime
```bash
$ ./cmctl import-cursor-chat --latest
Successfully imported chat as memory:
ID: mem_68d0a873_0d4dde
Name: AI Service Chat
Labels: map[activity:debugging date:2025-09-22 ...]
```

### ❌ WebAssembly Compilation
```bash
$ GOOS=js GOARCH=wasm go build -o cmctl.wasm .
build constraints exclude all Go files in modernc.org/libc@v1.22.5/errno
build constraints exclude all Go files in modernc.org/libc@v1.22.5/pthread
# ... (additional libc subpackages)
```

## Root Cause Analysis

The dependency chain is:
- `gorm.io/gorm` (ORM framework)
- `github.com/glebarez/sqlite` (Pure Go SQLite driver for GORM)
- `modernc.org/sqlite` (Pure Go SQLite implementation) 
- `modernc.org/libc` (C library emulation for Go) **← Build constraints exclude js/wasm**

## Attempted Solutions

### 1. **`github.com/ncruces/go-sqlite3`** 
- **Approach**: Uses WASM-based SQLite with `wazero` runtime
- **Result**: ✅ Compiles to WASM but ❌ Runtime fails with "no SQLite binary embed/set/loaded"
- **Issue**: Requires explicit SQLite binary embedding and complex initialization

### 2. **Standard `glebarez/sqlite`**
- **Approach**: Pure Go SQLite implementation via `modernc.org/sqlite`
- **Result**: ✅ Perfect runtime functionality but ❌ WASM compilation blocked
- **Issue**: `modernc.org/libc` fundamentally doesn't support js/wasm targets

## Impact Assessment

### ❌ WebAssembly Limitations
- Cannot embed CLI as WASM in VS Code extension
- No zero-installation deployment option
- Cannot leverage WASM performance optimizations

### ✅ Standard Build Capabilities  
- Full cross-platform support (Linux, Darwin, Windows, x86_64, aarch64)
- Complete CLI functionality including Cursor SQLite database access
- Excellent runtime performance with pure Go SQLite
- Reliable subprocess integration with VS Code extension

## Current Architecture Decision

The VS Code extension uses **subprocess calls** to the external `cmctl` binary. This approach:

- ✅ **Reliability**: Works consistently across all platforms
- ✅ **Separation**: Clear boundaries between CLI and UI components  
- ✅ **Simplicity**: Avoids WASM complexity and binary management
- ✅ **Independence**: CLI and UI can evolve separately
- ✅ **Performance**: Native binary performance vs WASM overhead

## Alternative Approaches Evaluated

1. **Dual-build Strategy**: WASM build without SQLite, standard build with full features
   - ❌ Complexity: Maintaining two feature sets
   - ❌ User confusion: Different capabilities per platform

2. **Remove SQLite Dependency**: Eliminate Cursor chat integration  
   - ❌ Value loss: Core differentiator depends on this feature
   - ❌ User impact: Breaks programmatic chat export

3. **Custom SQLite Binding**: Build WASM-compatible SQLite interface
   - ❌ Maintenance: Significant ongoing development overhead
   - ❌ Risk: Recreating what `modernc.org/sqlite` already provides

## Conclusion

The **subprocess architecture** represents the optimal balance of:
- **Functionality**: Full feature set including Cursor integration
- **Maintainability**: Single codebase, proven patterns
- **Reliability**: No WASM runtime complexities
- **Performance**: Native execution speed
- **Cross-platform**: Universal compatibility

While WebAssembly would have been an elegant solution, the current architecture delivers all core value propositions without compromise.