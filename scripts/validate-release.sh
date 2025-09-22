#!/bin/bash
set -euo pipefail

# Release Validation Script
# Comprehensive validation before release

echo "🚀 ContextMemory Release Validation"
echo "=================================="
echo

# Check prerequisites
echo "📋 Checking prerequisites..."
command -v git >/dev/null || { echo "❌ git not found"; exit 1; }
command -v make >/dev/null || { echo "❌ make not found"; exit 1; }
command -v go >/dev/null || { echo "❌ go not found"; exit 1; }
command -v npm >/dev/null || { echo "❌ npm not found"; exit 1; }
command -v jq >/dev/null || { echo "❌ jq not found"; exit 1; }
echo "✅ All prerequisites found"
echo

# Validate git state
echo "🔍 Validating git state..."
if [[ $(git status --porcelain | wc -l) -gt 0 ]]; then
    echo "⚠️  Uncommitted changes found"
    git status --short
    echo
else
    echo "✅ Git working directory clean"
fi

# Get current version
CURRENT_VERSION=$(jq -r '.version' package.json)
echo "📦 Current version: $CURRENT_VERSION"
echo

# Run comprehensive tests
echo "🧪 Running comprehensive test suite..."
if make test.all; then
    echo "✅ All tests passed"
else
    echo "❌ Tests failed - aborting release"
    exit 1
fi

# Validate documentation
echo "📚 Validating documentation..."
REQUIRED_DOCS=(
    "README.md"
    "docs/ROADMAP.md"
    "docs/TECHNICAL_INNOVATIONS.md"
    "docs/WEBASSEMBLY_LIMITATION.md"
    "docs/VERSION_COMPATIBILITY.md"
)

for doc in "${REQUIRED_DOCS[@]}"; do
    if [[ -f "$doc" ]]; then
        echo "  ✅ $doc"
    else
        echo "  ❌ $doc missing"
        exit 1
    fi
done

# Check for version consistency
echo "🔢 Checking version consistency..."
CLI_VERSION=$(./cmd/cmctl/cmctl --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
UI_VERSION=$(jq -r '.version' ui/package.json)
ROOT_VERSION=$(jq -r '.version' package.json)

if [[ "$CLI_VERSION" == "$UI_VERSION" && "$UI_VERSION" == "$ROOT_VERSION" ]]; then
    echo "✅ Version consistency validated: $CURRENT_VERSION"
else
    echo "❌ Version mismatch detected:"
    echo "  CLI:  $CLI_VERSION"
    echo "  UI:   $UI_VERSION"
    echo "  Root: $ROOT_VERSION"
    exit 1
fi

# Validate build artifacts
echo "🏗️  Validating build artifacts..."
if [[ -f "cmd/cmctl/cmctl" ]]; then
    echo "✅ CLI binary present"
else
    echo "❌ CLI binary missing"
    exit 1
fi

if [[ -f "ui/contextmemory-$CURRENT_VERSION.vsix" ]]; then
    VSIX_SIZE=$(stat -f%z "ui/contextmemory-$CURRENT_VERSION.vsix" 2>/dev/null || echo "0")
    echo "✅ Extension package present ($VSIX_SIZE bytes)"
else
    echo "❌ Extension package missing"
    exit 1
fi

# Check changelog
echo "📝 Checking changelog..."
if [[ -f "CHANGELOG.md" ]] && grep -q "$CURRENT_VERSION" CHANGELOG.md; then
    echo "✅ Changelog updated for $CURRENT_VERSION"
else
    echo "⚠️  Changelog may need updating for $CURRENT_VERSION"
fi

echo
echo "🎉 Release validation completed successfully!"
echo "✅ Version $CURRENT_VERSION is ready for release"
echo
echo "Next steps:"
echo "1. Review changes: git diff --staged"
echo "2. Commit changes: git commit -m 'feat: release v$CURRENT_VERSION'"
echo "3. Create tag: git tag v$CURRENT_VERSION"
echo "4. Push changes: git push origin main --tags"
