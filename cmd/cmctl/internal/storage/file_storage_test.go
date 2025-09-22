package storage

import (
	"path/filepath"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	tempDir := t.TempDir()

	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	if fs.storageDir != tempDir {
		t.Errorf("Expected storageDir %s, got %s", tempDir, fs.storageDir)
	}

	expectedMemoriesDir := filepath.Join(tempDir, "memories")
	if fs.memoriesDir != expectedMemoriesDir {
		t.Errorf("Expected memoriesDir %s, got %s", expectedMemoriesDir, fs.memoriesDir)
	}
}

func TestCreateMemory(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	req := CreateMemoryRequest{
		Name:    "Test Memory",
		Content: "Test content",
		Labels:  map[string]string{"test": "true"},
	}

	memory, err := fs.Create(req)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	if memory.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, memory.Name)
	}

	if memory.Content != req.Content {
		t.Errorf("Expected content %s, got %s", req.Content, memory.Content)
	}

	if memory.Labels["test"] != "true" {
		t.Errorf("Expected label test=true, got %v", memory.Labels)
	}

	if memory.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if memory.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

func TestGetMemory(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Create a memory first
	req := CreateMemoryRequest{
		Name:    "Test Memory",
		Content: "Test content",
	}

	created, err := fs.Create(req)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Get the memory
	retrieved, err := fs.Get(created.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
	}
}

func TestGetNonExistentMemory(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	_, err = fs.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent memory")
	}
}

func TestListMemories(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Create multiple memories
	memories := []CreateMemoryRequest{
		{Name: "Memory 1", Content: "Content 1"},
		{Name: "Memory 2", Content: "Content 2"},
	}

	for _, memReq := range memories {
		_, err := fs.Create(memReq)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}
	}

	// List memories
	listed, err := fs.List()
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(listed) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(listed))
	}
}

func TestDeleteMemory(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Create a memory
	req := CreateMemoryRequest{
		Name:    "Test Memory",
		Content: "Test content",
	}

	created, err := fs.Create(req)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Delete the memory
	err = fs.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify it's gone
	_, err = fs.Get(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted memory")
	}
}

func TestSearchMemories(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Create memories with different content
	searchMemories := []CreateMemoryRequest{
		{Name: "Go Tutorial", Content: "Learning Go programming language"},
		{Name: "Python Guide", Content: "Python best practices"},
		{Name: "Go Advanced", Content: "Advanced Go concepts"},
	}

	for _, memReq := range searchMemories {
		_, err := fs.Create(memReq)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}
	}

	// Search for "Go"
	searchReq := SearchRequest{
		Query: "Go",
	}

	response, err := fs.Search(searchReq)
	if err != nil {
		t.Fatalf("Failed to search memories: %v", err)
	}

	if len(response.Memories) != 2 {
		t.Errorf("Expected 2 results for 'Go' search, got %d", len(response.Memories))
	}
}

func TestMemoryLabels(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	labels := map[string]string{
		"project":  "test",
		"type":     "documentation",
		"priority": "high",
	}

	req := CreateMemoryRequest{
		Name:    "Test Memory",
		Content: "Test content",
		Labels:  labels,
	}

	memory, err := fs.Create(req)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	for key, expectedValue := range labels {
		if actualValue, exists := memory.Labels[key]; !exists {
			t.Errorf("Label %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Label %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestHealth(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	err = fs.Health()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestStorageInfo(t *testing.T) {
	tempDir := t.TempDir()
	fs, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create FileStorage: %v", err)
	}

	// Create a memory first
	req := CreateMemoryRequest{
		Name:    "Test Memory",
		Content: "Test content",
	}

	_, err = fs.Create(req)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	info, err := fs.GetStorageInfo()
	if err != nil {
		t.Fatalf("Failed to get storage info: %v", err)
	}

	if info.MemoriesCount != 1 {
		t.Errorf("Expected 1 total memory, got %d", info.MemoriesCount)
	}

	if info.StorageDir != tempDir {
		t.Errorf("Expected storage location %s, got %s", tempDir, info.StorageDir)
	}
}
