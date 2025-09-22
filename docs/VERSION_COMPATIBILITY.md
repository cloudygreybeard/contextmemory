# Version Compatibility Policy

ContextMemory follows a **Minor Version Compatibility** policy to ensure reliable operation between the VS Code extension and the `cmctl` CLI.

## Policy Overview

**Compatibility Rule**: Extension v`X.Y.Z` requires CLI v`X.Y.*` (same major and minor version, any patch version)

### Examples

| Extension Version | Compatible CLI Versions | Incompatible CLI Versions |
|------------------|------------------------|--------------------------|
| v0.6.1          | v0.6.0, v0.6.1, v0.6.2, v0.6.10 | v0.5.x, v0.7.x, v1.x.x |
| v0.7.0          | v0.7.0, v0.7.1, v0.7.5 | v0.6.x, v0.8.x, v1.x.x |
| v1.0.0          | v1.0.0, v1.0.1, v1.0.9 | v0.x.x, v1.1.x, v2.x.x |

## Rationale

This policy balances flexibility with reliability:

- **Patch versions** (0.6.0 → 0.6.1) are backward compatible by definition
- **Minor versions** (0.6.x → 0.7.x) may introduce API changes requiring coordination
- **Major versions** (0.x → 1.x) can introduce breaking changes

## Version Checking Behavior

### Extension Activation
- Automatic version compatibility check on extension startup
- Clear user notification for version mismatches with actionable options:
  - **Update CLI**: Direct link to installation guide
  - **Check Docs**: Link to documentation
  - **Continue Anyway**: Proceed with warning about potential issues

### Health Command
The `ContextMemory: Check Health` command displays:
- CLI version
- Extension version  
- Compatibility status
- Storage information
- Guidance for version mismatches

## Implementation

### Version Detection
```typescript
// CLI version detection
const version = await cmctlService.getVersion(); // "0.6.1"

// Compatibility check
const check = await cmctlService.checkVersionCompatibility();
// Returns: { compatible: boolean, reason?: string, cliVersion: string, extensionVersion: string }
```

### Compatibility Logic
```typescript
function isCompatible(cliVersion: string, extensionVersion: string): boolean {
    const cli = parseVersion(cliVersion);
    const ext = parseVersion(extensionVersion);
    
    return cli.major === ext.major && cli.minor === ext.minor;
}
```

## Testing

Run the compatibility test suite:
```bash
node test-version-compatibility.js
```

This validates the compatibility logic against various version combinations.

## Alternative Policies Considered

1. **Exact Version Match**: Too restrictive for independent patch updates
2. **Kubernetes-style Skew**: ±1 minor version tolerance - adds complexity
3. **Range-based**: Configurable ranges - too complex to maintain

The chosen **Minor Version Compatibility** policy follows semantic versioning best practices and provides the right balance for this project.
