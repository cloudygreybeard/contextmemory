package storage

import (
	"time"
)

// Memory represents a stored memory with content and metadata
type Memory struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Content   string            `json:"content"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
}

// CreateMemoryRequest represents a request to create a new memory
type CreateMemoryRequest struct {
	Name     string            `json:"name,omitempty"`
	Content  string            `json:"content"`
	Labels   map[string]string `json:"labels,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// UpdateMemoryRequest represents a request to update an existing memory
type UpdateMemoryRequest struct {
	ID       string            `json:"id"`
	Name     string            `json:"name,omitempty"`
	Content  string            `json:"content,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// SearchRequest represents a search query for memories
type SearchRequest struct {
	Query         string            `json:"query,omitempty"`
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	SortBy        string            `json:"sortBy,omitempty"`
	SortOrder     string            `json:"sortOrder,omitempty"`
}

// SearchResponse represents the result of a search operation
type SearchResponse struct {
	Memories []Memory `json:"memories"`
	Total    int      `json:"total"`
}

// StorageInfo provides information about the storage system
type StorageInfo struct {
	StorageDir    string `json:"storageDir"`
	MemoriesCount int    `json:"memoriesCount"`
	TotalSize     int64  `json:"totalSize"`
}
