package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"contextmemory/internal/providers"
	"contextmemory/internal/storage"
)

// Server holds our contextmemory implementation
type Server struct {
	provider providers.StorageProvider
}

// MemoryOutput defines a memory response
type MemoryOutput struct {
	ID        string            `json:"id" jsonschema:"Memory ID"`
	Name      string            `json:"name" jsonschema:"Memory name"`
	Content   string            `json:"content,omitempty" jsonschema:"Memory content"`
	Labels    map[string]string `json:"labels,omitempty" jsonschema:"Memory labels"`
	Namespace string            `json:"namespace" jsonschema:"Memory namespace"`
	CreatedAt string            `json:"createdAt" jsonschema:"Creation timestamp"`
	UpdatedAt string            `json:"updatedAt" jsonschema:"Last update timestamp"`
}

// GetMemoryOutput defines the unified response for get operations
type GetMemoryOutput struct {
	// For single memory retrieval
	Memory *MemoryOutput `json:"memory,omitempty" jsonschema:"Single memory details"`
	// For list operations
	Memories []MemoryOutput `json:"memories,omitempty" jsonschema:"List of memories"`
	Count    int            `json:"count,omitempty" jsonschema:"Number of memories returned"`
	// Formatted table output
	Table       string   `json:"table,omitempty" jsonschema:"Formatted table output"`
	Headers     []string `json:"headers,omitempty" jsonschema:"Table headers"`
	OutputType  string   `json:"outputType" jsonschema:"Type of output: single, list, table"`
	LabelFilter string   `json:"labelFilter,omitempty" jsonschema:"Applied label filter"`
}

// GetMemoryInput defines parameters for getting memories (unified list/get)
type GetMemoryInput struct {
	Namespace      string            `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	Name           string            `json:"name,omitempty" jsonschema:"Memory name (optional - if empty, lists all memories)"`
	LabelSelector  map[string]string `json:"labelSelector,omitempty" jsonschema:"Key-value pairs to filter memories"`
	OutputFormat   string            `json:"outputFormat,omitempty" jsonschema:"Output format: table (default), json, detailed"`
	IncludeContent bool              `json:"includeContent,omitempty" jsonschema:"Include memory content in response"`
	Limit          int               `json:"limit,omitempty" jsonschema:"Maximum number of results to return"`
}

// CreateMemoryInput defines parameters for creating a memory
type CreateMemoryInput struct {
	Namespace string            `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	Name      string            `json:"name" jsonschema:"Memory name"`
	Content   string            `json:"content" jsonschema:"Memory content"`
	Labels    map[string]string `json:"labels,omitempty" jsonschema:"Memory labels"`
}

// UpdateMemoryInput defines parameters for updating a memory
type UpdateMemoryInput struct {
	Namespace string            `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	Name      string            `json:"name" jsonschema:"Memory name"`
	Content   string            `json:"content" jsonschema:"Memory content"`
	Labels    map[string]string `json:"labels,omitempty" jsonschema:"Memory labels"`
}

// DeleteMemoryInput defines parameters for deleting a memory
type DeleteMemoryInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	Name      string `json:"name" jsonschema:"Memory name"`
}

// SearchMemoriesInput defines parameters for searching memories
type SearchMemoriesInput struct {
	Query          string            `json:"query,omitempty" jsonschema:"Search query"`
	Namespace      string            `json:"namespace,omitempty" jsonschema:"Namespace to filter by"`
	LabelSelector  map[string]string `json:"labelSelector,omitempty" jsonschema:"Key-value pairs to filter memories"`
	IncludeContent bool              `json:"includeContent,omitempty" jsonschema:"Include memory content in response"`
	Limit          int               `json:"limit,omitempty" jsonschema:"Maximum number of results to return"`
}

// SearchMemoriesOutput defines the response for searching memories
type SearchMemoriesOutput struct {
	Memories []MemoryOutput `json:"memories" jsonschema:"List of memories matching the search"`
	Count    int            `json:"count" jsonschema:"Number of memories returned"`
	Query    string         `json:"query,omitempty" jsonschema:"The search query used"`
}

// DeleteMemoryOutput defines the response for deleting a memory
type DeleteMemoryOutput struct {
	Success bool   `json:"success" jsonschema:"Whether the deletion was successful"`
	Message string `json:"message" jsonschema:"Success or error message"`
}

// formatTable creates a kubectl-style table output
func formatTable(memories []MemoryOutput, includeLongFormat bool) (string, []string) {
	headers := []string{"NAME", "NAMESPACE", "AGE", "LABELS"}
	if includeLongFormat {
		headers = append(headers, "CONTENT-SIZE")
	}

	var rows []string
	for _, memory := range memories {
		// Calculate age
		createdTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", memory.CreatedAt)
		age := time.Since(createdTime)
		ageStr := formatAge(age)

		// Format labels
		labelsStr := formatLabels(memory.Labels)
		if labelsStr == "" {
			labelsStr = "<none>"
		}

		row := fmt.Sprintf("%-30s %-12s %-8s %s",
			truncate(memory.Name, 30),
			memory.Namespace,
			ageStr,
			labelsStr)

		if includeLongFormat {
			contentSize := fmt.Sprintf("%d", len(memory.Content))
			row += fmt.Sprintf(" %-12s", contentSize)
		}

		rows = append(rows, row)
	}

	// Create header row
	headerRow := ""
	for i, header := range headers {
		if i == 0 {
			headerRow += fmt.Sprintf("%-30s", header)
		} else if i == 1 {
			headerRow += fmt.Sprintf(" %-12s", header)
		} else if i == 2 {
			headerRow += fmt.Sprintf(" %-8s", header)
		} else if i == 3 {
			headerRow += fmt.Sprintf(" %s", header)
		} else if i == 4 {
			headerRow += fmt.Sprintf(" %-12s", header)
		}
	}

	table := headerRow + "\n" + strings.Join(rows, "\n")
	return table, headers
}

// formatAge converts a duration to a human-readable age string
func formatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// formatLabels converts label map to a string
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// GetMemory implements the unified contextmemory_get tool (replaces both list and get)
func (s *Server) GetMemory(ctx context.Context, req *mcp.CallToolRequest, input GetMemoryInput) (
	*mcp.CallToolResult,
	GetMemoryOutput,
	error,
) {
	log.Printf("GetMemory called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}
	outputFormat := input.OutputFormat
	if outputFormat == "" {
		outputFormat = "table"
	}

	// If name is specified, get single memory
	if input.Name != "" {
		memory, err := s.provider.Get(input.Name)
		if err != nil {
			return nil, GetMemoryOutput{}, fmt.Errorf("failed to get memory: %w", err)
		}

		// Extract namespace from memory metadata
		memNamespace := namespace
		if memory.Metadata != nil {
			if ns, ok := memory.Metadata["namespace"].(string); ok && ns != "" {
				memNamespace = ns
			}
		}

		memoryOutput := MemoryOutput{
			ID:        memory.ID,
			Name:      memory.Name,
			Content:   memory.Content,
			Labels:    memory.Labels,
			Namespace: memNamespace,
			CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: memory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		result := GetMemoryOutput{
			Memory:     &memoryOutput,
			OutputType: "single",
		}

		return nil, result, nil
	}

	// List memories (when no name specified)
	searchReq := storage.SearchRequest{
		LabelSelector:  input.LabelSelector,
		Limit:          input.Limit,
		IncludeContent: input.IncludeContent,
		UseIndex:       true,
	}

	searchResult, err := s.provider.Search(searchReq)
	if err != nil {
		return nil, GetMemoryOutput{}, fmt.Errorf("failed to list memories: %w", err)
	}

	// Convert storage memories to output format
	memories := make([]MemoryOutput, 0, len(searchResult.Memories))
	for _, memory := range searchResult.Memories {
		// Extract namespace from memory metadata
		memNamespace := namespace
		if memory.Metadata != nil {
			if ns, ok := memory.Metadata["namespace"].(string); ok && ns != "" {
				memNamespace = ns
			}
		}

		content := ""
		if input.IncludeContent {
			content = memory.Content
		}

		memoryOutput := MemoryOutput{
			ID:        memory.ID,
			Name:      memory.Name,
			Content:   content,
			Labels:    memory.Labels,
			Namespace: memNamespace,
			CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: memory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		memories = append(memories, memoryOutput)
	}

	result := GetMemoryOutput{
		Memories: memories,
		Count:    len(memories),
	}

	// Format output based on requested format
	switch outputFormat {
	case "table":
		table, headers := formatTable(memories, input.IncludeContent)
		result.Table = table
		result.Headers = headers
		result.OutputType = "table"
	case "json", "detailed":
		result.OutputType = "list"
	default:
		// Default to table format
		table, headers := formatTable(memories, input.IncludeContent)
		result.Table = table
		result.Headers = headers
		result.OutputType = "table"
	}

	// Add label filter info if used
	if len(input.LabelSelector) > 0 {
		result.LabelFilter = formatLabels(input.LabelSelector)
	}

	return nil, result, nil
}

// CreateMemory implements the contextmemory_create tool
func (s *Server) CreateMemory(ctx context.Context, req *mcp.CallToolRequest, input CreateMemoryInput) (
	*mcp.CallToolResult,
	MemoryOutput,
	error,
) {
	log.Printf("CreateMemory called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if input.Name == "" {
		return nil, MemoryOutput{}, fmt.Errorf("memory name is required")
	}
	if input.Content == "" {
		return nil, MemoryOutput{}, fmt.Errorf("memory content is required")
	}

	// Create memory request
	createReq := storage.CreateMemoryRequest{
		Name:    input.Name,
		Content: input.Content,
		Labels:  input.Labels,
		Metadata: map[string]any{
			"namespace": namespace,
		},
	}

	memory, err := s.provider.Create(createReq)
	if err != nil {
		return nil, MemoryOutput{}, fmt.Errorf("failed to create memory: %w", err)
	}

	result := MemoryOutput{
		ID:        memory.ID,
		Name:      memory.Name,
		Content:   memory.Content,
		Labels:    memory.Labels,
		Namespace: namespace,
		CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: memory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return nil, result, nil
}

// UpdateMemory implements the contextmemory_update tool
func (s *Server) UpdateMemory(ctx context.Context, req *mcp.CallToolRequest, input UpdateMemoryInput) (
	*mcp.CallToolResult,
	MemoryOutput,
	error,
) {
	log.Printf("UpdateMemory called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if input.Name == "" {
		return nil, MemoryOutput{}, fmt.Errorf("memory name is required")
	}
	if input.Content == "" {
		return nil, MemoryOutput{}, fmt.Errorf("memory content is required")
	}

	// Create update request
	updateReq := storage.UpdateMemoryRequest{
		ID:      input.Name, // Use name as ID for now
		Name:    input.Name,
		Content: input.Content,
		Labels:  input.Labels,
		Metadata: map[string]any{
			"namespace": namespace,
		},
	}

	memory, err := s.provider.Update(updateReq)
	if err != nil {
		return nil, MemoryOutput{}, fmt.Errorf("failed to update memory: %w", err)
	}

	result := MemoryOutput{
		ID:        memory.ID,
		Name:      memory.Name,
		Content:   memory.Content,
		Labels:    memory.Labels,
		Namespace: namespace,
		CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: memory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return nil, result, nil
}

// DeleteMemory implements the contextmemory_delete tool
func (s *Server) DeleteMemory(ctx context.Context, req *mcp.CallToolRequest, input DeleteMemoryInput) (
	*mcp.CallToolResult,
	DeleteMemoryOutput,
	error,
) {
	log.Printf("DeleteMemory called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if input.Name == "" {
		return nil, DeleteMemoryOutput{}, fmt.Errorf("memory name is required")
	}

	err := s.provider.Delete(input.Name)
	if err != nil {
		return nil, DeleteMemoryOutput{Success: false, Message: fmt.Sprintf("Failed to delete memory: %v", err)}, nil
	}

	result := DeleteMemoryOutput{
		Success: true,
		Message: fmt.Sprintf("Memory '%s' deleted successfully", input.Name),
	}

	return nil, result, nil
}

// SearchMemories implements the contextmemory_search tool
func (s *Server) SearchMemories(ctx context.Context, req *mcp.CallToolRequest, input SearchMemoriesInput) (
	*mcp.CallToolResult,
	SearchMemoriesOutput,
	error,
) {
	log.Printf("SearchMemories called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// Create search request for the storage layer
	searchReq := storage.SearchRequest{
		Query:          input.Query,
		LabelSelector:  input.LabelSelector,
		Limit:          input.Limit,
		IncludeContent: input.IncludeContent,
		UseIndex:       true,
	}

	searchResult, err := s.provider.Search(searchReq)
	if err != nil {
		return nil, SearchMemoriesOutput{}, fmt.Errorf("search failed: %w", err)
	}

	// Convert storage memories to output format
	memories := make([]MemoryOutput, 0, len(searchResult.Memories))
	for _, memory := range searchResult.Memories {
		// Extract namespace from memory metadata
		memNamespace := namespace
		if memory.Metadata != nil {
			if ns, ok := memory.Metadata["namespace"].(string); ok && ns != "" {
				memNamespace = ns
			}
		}

		content := ""
		if input.IncludeContent {
			content = memory.Content
		}

		memoryOutput := MemoryOutput{
			ID:        memory.ID,
			Name:      memory.Name,
			Content:   content,
			Labels:    memory.Labels,
			Namespace: memNamespace,
			CreatedAt: memory.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: memory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		memories = append(memories, memoryOutput)
	}

	result := SearchMemoriesOutput{
		Memories: memories,
		Count:    len(memories),
		Query:    input.Query,
	}

	return nil, result, nil
}

func main() {
	// Initialize file storage provider with default configuration
	config := providers.ProviderConfig{
		Type:       providers.FileProvider,
		StorageDir: "", // Uses ~/.contextmemory by default
		Timeout:    30,
	}

	factory := providers.NewProviderFactory()
	providerInterface, err := factory.CreateProvider(config)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	provider := providerInterface.(providers.StorageProvider)
	if err := provider.ValidateConfig(); err != nil {
		log.Fatalf("Storage provider validation failed: %v", err)
	}

	// Create our server instance
	contextMemoryServer := &Server{provider: provider}

	// Create MCP server with our implementation info
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "contextmemory",
		Version: "0.7.0",
	}, nil)

	// Add contextmemory tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_get",
		Description: "Get memories - list all memories or get a specific memory by name (kubectl-style interface)",
	}, contextMemoryServer.GetMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_create",
		Description: "Create a new memory",
	}, contextMemoryServer.CreateMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_update",
		Description: "Update an existing memory",
	}, contextMemoryServer.UpdateMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_delete",
		Description: "Delete a memory by name",
	}, contextMemoryServer.DeleteMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_search",
		Description: "Search memories by content or labels",
	}, contextMemoryServer.SearchMemories)

	// Set up logging to stderr to avoid JSON-RPC interference
	log.SetOutput(os.Stderr)
	log.Println("Starting ContextMemory MCP server with official SDK...")

	// Run the server over stdin/stdout until client disconnects
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal("Server failed:", err)
	}
}
