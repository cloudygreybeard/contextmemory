package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/utils"
)

// FileStorage implements file-based storage for memories
type FileStorage struct {
	storageDir  string
	memoriesDir string
	indexFile   string
	configFile  string
}

// Index represents the storage index for fast lookups
type Index struct {
	Memories    []IndexEntry `json:"memories"`
	LastUpdated time.Time    `json:"lastUpdated"`
}

// IndexEntry represents a memory entry in the index
type IndexEntry struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// NewFileStorage creates a new file-based storage instance
func NewFileStorage(storageDir string) (*FileStorage, error) {
	if storageDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		storageDir = filepath.Join(home, ".contextmemory-v2")
	}

	fs := &FileStorage{
		storageDir:  storageDir,
		memoriesDir: filepath.Join(storageDir, "memories"),
		indexFile:   filepath.Join(storageDir, "index.json"),
		configFile:  filepath.Join(storageDir, "config.json"),
	}

	if err := fs.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return fs, nil
}

// initialize sets up the storage directories and files
func (fs *FileStorage) initialize() error {
	// Create directories
	dirs := []string{fs.storageDir, fs.memoriesDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Initialize index if it doesn't exist
	if _, err := os.Stat(fs.indexFile); os.IsNotExist(err) {
		index := Index{
			Memories:    []IndexEntry{},
			LastUpdated: time.Now(),
		}
		if err := fs.writeIndex(index); err != nil {
			return fmt.Errorf("failed to initialize index: %w", err)
		}
	}

	// Initialize config if it doesn't exist
	if _, err := os.Stat(fs.configFile); os.IsNotExist(err) {
		config := map[string]any{
			"version": "2.0.0",
			"created": time.Now(),
			"storage": "file-based",
		}
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := os.WriteFile(fs.configFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	return nil
}

// Create creates a new memory
func (fs *FileStorage) Create(req CreateMemoryRequest) (*Memory, error) {
	memory := &Memory{
		ID:        utils.GenerateID(),
		Name:      req.Name,
		Content:   req.Content,
		Labels:    req.Labels,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  req.Metadata,
	}

	// Apply defaults
	if memory.Name == "" {
		memory.Name = fmt.Sprintf("Memory %s", time.Now().Format("2006-01-02"))
	}
	if memory.Labels == nil {
		memory.Labels = make(map[string]string)
	}
	if memory.Labels["type"] == "" {
		memory.Labels["type"] = "manual"
	}

	// Validate
	if err := fs.validateMemory(memory); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Write memory file
	if err := fs.writeMemory(memory); err != nil {
		return nil, fmt.Errorf("failed to write memory: %w", err)
	}

	// Update index
	if err := fs.updateIndex(memory, "create"); err != nil {
		// Log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to update index: %v\n", err)
	}

	return memory, nil
}

// Get retrieves a memory by ID
func (fs *FileStorage) Get(id string) (*Memory, error) {
	memoryFile := filepath.Join(fs.memoriesDir, id+".json")
	
	data, err := os.ReadFile(memoryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Memory not found
		}
		return nil, fmt.Errorf("failed to read memory file: %w", err)
	}

	var memory Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal memory: %w", err)
	}

	return &memory, nil
}

// Update updates an existing memory
func (fs *FileStorage) Update(req UpdateMemoryRequest) (*Memory, error) {
	existing, err := fs.Get(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing memory: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("memory not found: %s", req.ID)
	}

	// Update fields if provided
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Content != "" {
		existing.Content = req.Content
	}
	if req.Labels != nil {
		existing.Labels = req.Labels
	}
	if req.Metadata != nil {
		if existing.Metadata == nil {
			existing.Metadata = make(map[string]any)
		}
		for k, v := range req.Metadata {
			existing.Metadata[k] = v
		}
	}
	existing.UpdatedAt = time.Now()

	// Validate
	if err := fs.validateMemory(existing); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Write updated memory
	if err := fs.writeMemory(existing); err != nil {
		return nil, fmt.Errorf("failed to write memory: %w", err)
	}

	// Update index
	if err := fs.updateIndex(existing, "update"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update index: %v\n", err)
	}

	return existing, nil
}

// Delete removes a memory by ID
func (fs *FileStorage) Delete(id string) error {
	memoryFile := filepath.Join(fs.memoriesDir, id+".json")
	
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		return fmt.Errorf("memory not found: %s", id)
	}

	if err := os.Remove(memoryFile); err != nil {
		return fmt.Errorf("failed to delete memory file: %w", err)
	}

	// Update index
	if err := fs.updateIndex(&Memory{ID: id}, "delete"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update index: %v\n", err)
	}

	return nil
}

// Search searches for memories based on the given criteria
func (fs *FileStorage) Search(req SearchRequest) (*SearchResponse, error) {
	memories, err := fs.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	// Apply filters
	filtered := fs.applyFilters(memories, req)

	// Apply sorting
	fs.applySorting(filtered, req)

	// Apply limit
	if req.Limit > 0 && len(filtered) > req.Limit {
		filtered = filtered[:req.Limit]
	}

	return &SearchResponse{
		Memories: filtered,
		Total:    len(memories),
	}, nil
}

// List returns all memories
func (fs *FileStorage) List() ([]Memory, error) {
	files, err := filepath.Glob(filepath.Join(fs.memoriesDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob memory files: %w", err)
	}

	var memories []Memory
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping corrupted file %s: %v\n", file, err)
			continue
		}

		var memory Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping corrupted file %s: %v\n", file, err)
			continue
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// Health checks if the storage is accessible and healthy
func (fs *FileStorage) Health() error {
	// Check if storage directory is accessible
	if _, err := os.Stat(fs.storageDir); err != nil {
		return fmt.Errorf("storage directory not accessible: %w", err)
	}

	// Try to write a test file
	testFile := filepath.Join(fs.storageDir, ".health-check")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// GetStorageInfo returns information about the storage
func (fs *FileStorage) GetStorageInfo() (*StorageInfo, error) {
	files, err := filepath.Glob(filepath.Join(fs.memoriesDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob memory files: %w", err)
	}

	var totalSize int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}

	return &StorageInfo{
		StorageDir:    fs.storageDir,
		MemoriesCount: len(files),
		TotalSize:     totalSize,
	}, nil
}

// Helper methods

func (fs *FileStorage) validateMemory(memory *Memory) error {
	if memory.Name == "" {
		return fmt.Errorf("memory name cannot be empty")
	}
	if len(memory.Name) > 200 {
		return fmt.Errorf("memory name too long (max 200 characters)")
	}
	if memory.Labels != nil {
		for k, v := range memory.Labels {
			if len(k) > 63 || len(v) > 63 {
				return fmt.Errorf("label key/value too long (max 63 characters)")
			}
		}
	}
	return nil
}

func (fs *FileStorage) writeMemory(memory *Memory) error {
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	memoryFile := filepath.Join(fs.memoriesDir, memory.ID+".json")
	if err := os.WriteFile(memoryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return nil
}

func (fs *FileStorage) applyFilters(memories []Memory, req SearchRequest) []Memory {
	var filtered []Memory

	for _, memory := range memories {
		// Text search
		if req.Query != "" {
			query := strings.ToLower(req.Query)
			if !strings.Contains(strings.ToLower(memory.Name), query) &&
				!strings.Contains(strings.ToLower(memory.Content), query) {
				continue
			}
		}

		// Label selector
		if req.LabelSelector != nil {
			match := true
			for k, v := range req.LabelSelector {
				if memory.Labels[k] != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		filtered = append(filtered, memory)
	}

	return filtered
}

func (fs *FileStorage) applySorting(memories []Memory, req SearchRequest) {
	// Simple sorting implementation
	// TODO: Implement proper sorting based on req.SortBy and req.SortOrder
}

func (fs *FileStorage) updateIndex(memory *Memory, operation string) error {
	index, err := fs.readIndex()
	if err != nil {
		return err
	}

	switch operation {
	case "create":
		entry := IndexEntry{
			ID:        memory.ID,
			Name:      memory.Name,
			Labels:    memory.Labels,
			CreatedAt: memory.CreatedAt,
			UpdatedAt: memory.UpdatedAt,
		}
		index.Memories = append(index.Memories, entry)
	case "update":
		for i, entry := range index.Memories {
			if entry.ID == memory.ID {
				index.Memories[i] = IndexEntry{
					ID:        memory.ID,
					Name:      memory.Name,
					Labels:    memory.Labels,
					CreatedAt: entry.CreatedAt, // Preserve original creation time
					UpdatedAt: memory.UpdatedAt,
				}
				break
			}
		}
	case "delete":
		for i, entry := range index.Memories {
			if entry.ID == memory.ID {
				index.Memories = append(index.Memories[:i], index.Memories[i+1:]...)
				break
			}
		}
	}

	index.LastUpdated = time.Now()
	return fs.writeIndex(index)
}

func (fs *FileStorage) readIndex() (Index, error) {
	var index Index
	
	data, err := os.ReadFile(fs.indexFile)
	if err != nil {
		return index, err
	}

	err = json.Unmarshal(data, &index)
	return index, err
}

func (fs *FileStorage) writeIndex(index Index) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.indexFile, data, 0644)
}
