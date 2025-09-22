package utils

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()

	if id == "" {
		t.Error("Generated ID should not be empty")
	}

	// Check format: mem_<hex>_<hex>
	if !strings.HasPrefix(id, "mem_") {
		t.Errorf("ID should start with 'mem_', got: %s", id)
	}

	parts := strings.Split(id, "_")
	if len(parts) != 3 {
		t.Errorf("ID should have 3 parts separated by '_', got: %s", id)
	}

	// Check that hex parts are the right length (8 characters each)
	if len(parts[1]) != 8 {
		t.Errorf("First hex part should be 8 characters, got %d: %s", len(parts[1]), parts[1])
	}

	if len(parts[2]) != 6 {
		t.Errorf("Second hex part should be 6 characters, got %d: %s", len(parts[2]), parts[2])
	}

	// Check uniqueness by generating multiple IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		newID := GenerateID()
		if ids[newID] {
			t.Errorf("Generated duplicate ID: %s", newID)
		}
		ids[newID] = true
	}
}

func TestGenerateIDFormat(t *testing.T) {
	id := GenerateID()

	// Should match pattern: mem_[a-f0-9]{8}_[a-f0-9]{6}
	expectedLen := len("mem_") + 8 + len("_") + 6
	if len(id) != expectedLen {
		t.Errorf("Expected ID length %d, got %d: %s", expectedLen, len(id), id)
	}

	// Verify hex characters
	parts := strings.Split(id, "_")
	for _, char := range parts[1] + parts[2] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Invalid hex character '%c' in ID: %s", char, id)
		}
	}
}
