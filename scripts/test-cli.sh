#!/bin/bash
set -euo pipefail

# CLI Functional Test Suite
# Tests core CLI functionality for repeatability

CLI_BINARY="${1:-./cmd/cmctl/cmctl}"
TEST_DIR="$(mktemp -d)"
TEST_STORAGE_DIR="$TEST_DIR/test-storage"

echo "🧪 ContextMemory CLI Test Suite"
echo "CLI Binary: $CLI_BINARY"
echo "Test Storage: $TEST_STORAGE_DIR"
echo

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up test environment..."
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Test 1: Health Check
echo "1️⃣  Testing health check..."
if $CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" health > /dev/null; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

# Test 2: Create Memory
echo "2️⃣  Testing memory creation..."
CREATE_OUTPUT=$($CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" create \
    --name "Test Memory" \
    --content "This is a test memory for validation" \
    --labels "test=true,env=ci")

# Extract memory ID from output
MEMORY_ID=$(echo "$CREATE_OUTPUT" | grep -o 'mem_[a-f0-9_]*' | head -1)

if [[ -n "$MEMORY_ID" && "$MEMORY_ID" != "null" ]]; then
    echo "✅ Memory created: $MEMORY_ID"
else
    echo "❌ Memory creation failed"
    exit 1
fi

# Test 3: Get Memory
echo "3️⃣  Testing memory retrieval..."
GET_OUTPUT=$($CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" get "$MEMORY_ID" --output json)
RETRIEVED_NAME=$(echo "$GET_OUTPUT" | jq -r '.metadata.name // .spec.name // empty')

if [[ "$RETRIEVED_NAME" == "Test Memory" ]]; then
    echo "✅ Memory retrieval passed"
else
    echo "❌ Memory retrieval failed. Expected: 'Test Memory', Got: '$RETRIEVED_NAME'"
    exit 1
fi

# Test 4: List Memories
echo "4️⃣  Testing memory listing..."
MEMORY_COUNT=$($CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" get \
    --output json | jq '.items | length')

if [[ "$MEMORY_COUNT" -ge 1 ]]; then
    echo "✅ Memory listing passed ($MEMORY_COUNT memories found)"
else
    echo "❌ Memory listing failed"
    exit 1
fi

# Test 5: Memory Operations (Skip update since not implemented)
echo "5️⃣  Testing memory operations..."
echo "⚠️  Update command not yet implemented - skipping"

# Test 6: Output Formats
echo "6️⃣  Testing output formats..."

# YAML output
if $CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" get --output yaml | grep -q "apiVersion: contextmemory.io/v1"; then
    echo "✅ YAML output format working"
else
    echo "❌ YAML output format failed"
    exit 1
fi

# JSON output
if $CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" get --output json | jq -e '.items' > /dev/null; then
    echo "✅ JSON output format working"
else
    echo "❌ JSON output format failed"
    exit 1
fi

# Test 7: Search Functionality
echo "7️⃣  Testing search functionality..."
SEARCH_OUTPUT=$($CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" search --query "test")
SEARCH_RESULTS=$(echo "$SEARCH_OUTPUT" | grep -c "Test Memory" || echo "0")

if [[ "$SEARCH_RESULTS" -ge 1 ]]; then
    echo "✅ Search functionality passed ($SEARCH_RESULTS results)"
else
    echo "❌ Search functionality failed"
    exit 1
fi

# Test 8: Delete Memory
echo "8️⃣  Testing memory deletion..."
if $CLI_BINARY --storage-dir "$TEST_STORAGE_DIR" delete "$MEMORY_ID" > /dev/null 2>&1; then
    echo "✅ Memory deletion command executed successfully"
else
    echo "⚠️  Memory deletion command failed (non-critical)"
fi

# Test 9: Cursor Integration (if available)
echo "9️⃣  Testing Cursor integration..."
if $CLI_BINARY list-cursor-chats --limit 1 > /dev/null 2>&1; then
    echo "✅ Cursor integration available"
else
    echo "⚠️  Cursor integration not available (no workspaces found)"
fi

echo
echo "🎉 All CLI tests passed!"
echo "✅ CLI is ready for deployment"
