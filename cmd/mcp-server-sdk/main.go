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

// Version information (set by build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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

// ConversationCheckpointInput defines parameters for creating conversation checkpoints
type ConversationCheckpointInput struct {
	Namespace       string            `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	ConversationID  string            `json:"conversationId,omitempty" jsonschema:"Unique conversation identifier (auto-generated if empty)"`
	Title           string            `json:"title,omitempty" jsonschema:"Conversation title/summary"`
	Phase           string            `json:"phase,omitempty" jsonschema:"Current phase of work (e.g., implementation, debugging, cleanup)"`
	Content         string            `json:"content" jsonschema:"Full conversation content (export format)"`
	KeyContext      []string          `json:"keyContext,omitempty" jsonschema:"Key points and decisions made"`
	NextSteps       []string          `json:"nextSteps,omitempty" jsonschema:"Planned next steps"`
	TruncationPoint string            `json:"truncationPoint,omitempty" jsonschema:"Point where context truncation occurred"`
	AutoGenerated   bool              `json:"autoGenerated,omitempty" jsonschema:"Whether this was an automatic checkpoint"`
	Labels          map[string]string `json:"labels,omitempty" jsonschema:"Additional labels for organization"`
}

// ConversationCheckpointOutput defines the response for creating conversation checkpoints
type ConversationCheckpointOutput struct {
	ConversationID string `json:"conversationId" jsonschema:"The conversation identifier used"`
	CheckpointID   string `json:"checkpointId" jsonschema:"The created memory ID"`
	Name           string `json:"name" jsonschema:"The created memory name"`
	Timestamp      string `json:"timestamp" jsonschema:"When the checkpoint was created"`
	Success        bool   `json:"success" jsonschema:"Whether the checkpoint was created successfully"`
	Message        string `json:"message" jsonschema:"Success or error message"`
}

// === PHASE 2: ENHANCED AUTOMATION TYPES ===

// IncrementalSnapshotInput defines parameters for creating incremental snapshots
type IncrementalSnapshotInput struct {
	Namespace      string            `json:"namespace,omitempty" jsonschema:"Memory namespace"`
	ConversationID string            `json:"conversationId" jsonschema:"Conversation identifier"`
	RecentContent  string            `json:"recentContent" jsonschema:"Recent conversation content since last checkpoint"`
	TokenCount     int               `json:"tokenCount" jsonschema:"Current estimated token count"`
	MessagesSince  int               `json:"messagesSince" jsonschema:"Number of messages since last checkpoint"`
	ContextSummary string            `json:"contextSummary,omitempty" jsonschema:"Summary of recent context"`
	AutoGenerated  bool              `json:"autoGenerated,omitempty" jsonschema:"Whether this was an automatic snapshot"`
	TriggerReason  string            `json:"triggerReason,omitempty" jsonschema:"What triggered this snapshot"`
	Labels         map[string]string `json:"labels,omitempty" jsonschema:"Additional labels for organization"`
}

// IncrementalSnapshotOutput defines the response for creating incremental snapshots
type IncrementalSnapshotOutput struct {
	ConversationID string `json:"conversationId" jsonschema:"The conversation identifier"`
	SnapshotID     string `json:"snapshotId" jsonschema:"The created snapshot memory ID"`
	Name           string `json:"name" jsonschema:"The created snapshot memory name"`
	Timestamp      string `json:"timestamp" jsonschema:"When the snapshot was created"`
	TokenCount     int    `json:"tokenCount" jsonschema:"Token count when snapshot was taken"`
	Success        bool   `json:"success" jsonschema:"Whether the snapshot was created successfully"`
	Message        string `json:"message" jsonschema:"Success or error message"`
}

// ConversationMonitorInput defines parameters for monitoring conversation state
type ConversationMonitorInput struct {
	ConversationID  string `json:"conversationId" jsonschema:"Conversation identifier to monitor"`
	CurrentContent  string `json:"currentContent,omitempty" jsonschema:"Current conversation content for token estimation"`
	MessageCount    int    `json:"messageCount,omitempty" jsonschema:"Current message count"`
	EstimatedTokens int    `json:"estimatedTokens,omitempty" jsonschema:"Estimated current token count"`
}

// ConversationMonitorOutput defines the response for conversation monitoring
type ConversationMonitorOutput struct {
	ConversationID          string  `json:"conversationId" jsonschema:"The conversation identifier"`
	CurrentTokens           int     `json:"currentTokens" jsonschema:"Current estimated token count"`
	MaxTokens               int     `json:"maxTokens" jsonschema:"Maximum token limit"`
	TokenUtilization        float64 `json:"tokenUtilization" jsonschema:"Current utilization percentage (0.0-1.0)"`
	WarningThreshold        float64 `json:"warningThreshold" jsonschema:"Warning threshold (default 0.8)"`
	CriticalThreshold       float64 `json:"criticalThreshold" jsonschema:"Critical threshold (default 0.95)"`
	Status                  string  `json:"status" jsonschema:"Status: healthy, warning, critical"`
	ShouldSnapshot          bool    `json:"shouldSnapshot" jsonschema:"Whether an incremental snapshot is recommended"`
	ShouldCheckpoint        bool    `json:"shouldCheckpoint" jsonschema:"Whether a full checkpoint is recommended"`
	LastCheckpoint          string  `json:"lastCheckpoint,omitempty" jsonschema:"Timestamp of last checkpoint"`
	LastSnapshot            string  `json:"lastSnapshot,omitempty" jsonschema:"Timestamp of last snapshot"`
	MessagesSinceCheckpoint int     `json:"messagesSinceCheckpoint" jsonschema:"Messages since last checkpoint"`
	RecommendedAction       string  `json:"recommendedAction" jsonschema:"Recommended next action"`
}

// ConversationThreadInput defines parameters for managing conversation threads
type ConversationThreadInput struct {
	ConversationID  string            `json:"conversationId" jsonschema:"Conversation identifier"`
	Title           string            `json:"title,omitempty" jsonschema:"Thread title"`
	Project         string            `json:"project,omitempty" jsonschema:"Project name"`
	Phase           string            `json:"phase,omitempty" jsonschema:"Current phase"`
	Status          string            `json:"status,omitempty" jsonschema:"Thread status: active, archived, completed"`
	ParentThread    string            `json:"parentThread,omitempty" jsonschema:"Parent conversation ID for threading"`
	Labels          map[string]string `json:"labels,omitempty" jsonschema:"Thread labels"`
	EstimatedTokens int               `json:"estimatedTokens,omitempty" jsonschema:"Current estimated token count"`
	MessageCount    int               `json:"messageCount,omitempty" jsonschema:"Current message count"`
}

// ConversationThreadOutput defines the response for thread management
type ConversationThreadOutput struct {
	ConversationID   string   `json:"conversationId" jsonschema:"The conversation identifier"`
	ThreadID         string   `json:"threadId" jsonschema:"The thread memory ID"`
	Name             string   `json:"name" jsonschema:"The thread memory name"`
	RelatedThreads   []string `json:"relatedThreads,omitempty" jsonschema:"Related conversation IDs"`
	TotalCheckpoints int      `json:"totalCheckpoints" jsonschema:"Number of checkpoints in this thread"`
	TotalSnapshots   int      `json:"totalSnapshots" jsonschema:"Number of snapshots in this thread"`
	Success          bool     `json:"success" jsonschema:"Whether the operation was successful"`
	Message          string   `json:"message" jsonschema:"Success or error message"`
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

// === PHASE 2: TOKEN COUNTING AND MONITORING UTILITIES ===

// estimateTokenCount provides a rough token count estimation
// Rule of thumb: ~4 characters per token for English text
func estimateTokenCount(content string) int {
	if content == "" {
		return 0
	}
	// Basic estimation: 4 characters ≈ 1 token
	charCount := len(content)
	return (charCount + 3) / 4 // Round up
}

// ConversationMonitorConfig holds monitoring thresholds
type ConversationMonitorConfig struct {
	MaxTokens          int     `json:"maxTokens"`
	WarningThreshold   float64 `json:"warningThreshold"`   // 0.8 = 80%
	CriticalThreshold  float64 `json:"criticalThreshold"`  // 0.95 = 95%
	SnapshotInterval   int     `json:"snapshotInterval"`   // Messages between snapshots
	CheckpointInterval int     `json:"checkpointInterval"` // Messages between full checkpoints
}

// getDefaultMonitorConfig returns sensible defaults
func getDefaultMonitorConfig() ConversationMonitorConfig {
	return ConversationMonitorConfig{
		MaxTokens:          100000, // Conservative estimate for context window
		WarningThreshold:   0.8,    // Warning at 80%
		CriticalThreshold:  0.95,   // Critical at 95%
		SnapshotInterval:   5,      // Snapshot every 5 messages
		CheckpointInterval: 20,     // Full checkpoint every 20 messages
	}
}

// analyzeConversationHealth determines conversation status and recommendations
func analyzeConversationHealth(tokenCount int, messageCount int, config ConversationMonitorConfig, lastCheckpoint, lastSnapshot string) ConversationMonitorOutput {
	utilization := float64(tokenCount) / float64(config.MaxTokens)

	var status string
	var shouldSnapshot, shouldCheckpoint bool
	var recommendedAction string

	// Determine status based on token utilization
	if utilization >= config.CriticalThreshold {
		status = "critical"
		shouldCheckpoint = true
		recommendedAction = "Create immediate full checkpoint - approaching context limit"
	} else if utilization >= config.WarningThreshold {
		status = "warning"
		shouldSnapshot = true
		recommendedAction = "Create incremental snapshot - approaching warning threshold"
	} else {
		status = "healthy"
		recommendedAction = "Continue conversation - context window healthy"
	}

	// Also consider message-based triggers
	if messageCount > 0 && messageCount%config.CheckpointInterval == 0 {
		shouldCheckpoint = true
		if status == "healthy" {
			recommendedAction = "Create scheduled checkpoint - message interval reached"
		}
	} else if messageCount > 0 && messageCount%config.SnapshotInterval == 0 {
		if !shouldSnapshot && status == "healthy" {
			shouldSnapshot = true
			recommendedAction = "Create scheduled snapshot - message interval reached"
		}
	}

	return ConversationMonitorOutput{
		CurrentTokens:           tokenCount,
		MaxTokens:               config.MaxTokens,
		TokenUtilization:        utilization,
		WarningThreshold:        config.WarningThreshold,
		CriticalThreshold:       config.CriticalThreshold,
		Status:                  status,
		ShouldSnapshot:          shouldSnapshot,
		ShouldCheckpoint:        shouldCheckpoint,
		LastCheckpoint:          lastCheckpoint,
		LastSnapshot:            lastSnapshot,
		MessagesSinceCheckpoint: messageCount, // Simplified for now
		RecommendedAction:       recommendedAction,
	}
}

// findLatestCheckpoint finds the most recent checkpoint for a conversation
func (s *Server) findLatestCheckpoint(conversationID string) (*storage.Memory, error) {
	searchReq := storage.SearchRequest{
		LabelSelector: map[string]string{
			"type":            "conversation-checkpoint",
			"conversation-id": conversationID,
		},
		Limit:    1,
		UseIndex: true,
	}

	result, err := s.provider.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if len(result.Memories) == 0 {
		return nil, nil
	}

	// Return the most recent (should be first due to storage sorting)
	return &result.Memories[0], nil
}

// findLatestSnapshot finds the most recent snapshot for a conversation
func (s *Server) findLatestSnapshot(conversationID string) (*storage.Memory, error) {
	searchReq := storage.SearchRequest{
		LabelSelector: map[string]string{
			"type":            "incremental-snapshot",
			"conversation-id": conversationID,
		},
		Limit:    1,
		UseIndex: true,
	}

	result, err := s.provider.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if len(result.Memories) == 0 {
		return nil, nil
	}

	return &result.Memories[0], nil
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

// CreateConversationCheckpoint implements the contextmemory_checkpoint tool
func (s *Server) CreateConversationCheckpoint(ctx context.Context, req *mcp.CallToolRequest, input ConversationCheckpointInput) (
	*mcp.CallToolResult,
	ConversationCheckpointOutput,
	error,
) {
	// Enhanced logging to detect automatic calls
	logFile, _ := os.OpenFile("/tmp/contextmemory-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()
	debugTimestamp := time.Now().Format("2006-01-02 15:04:05")
	logFile.WriteString(fmt.Sprintf("[%s] CreateConversationCheckpoint called with autoGenerated=%t, input: %+v\n", debugTimestamp, input.AutoGenerated, input))

	log.Printf("CreateConversationCheckpoint called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// Generate conversation ID if not provided
	conversationID := input.ConversationID
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%s", time.Now().Format("2006_01_02_15_04"))
	}

	// Generate checkpoint name
	timestamp := time.Now()
	checkpointName := fmt.Sprintf("%s_checkpoint_%s", conversationID, timestamp.Format("2006_01_02_15_04_05"))

	// Prepare checkpoint content in structured format
	var contentBuilder strings.Builder

	// Header with metadata
	contentBuilder.WriteString(fmt.Sprintf("# Conversation Checkpoint: %s\n", conversationID))
	contentBuilder.WriteString(fmt.Sprintf("_Created on %s_\n\n", timestamp.Format("02/01/2006 at 15:04:05 MST")))

	if input.Title != "" {
		contentBuilder.WriteString(fmt.Sprintf("## Title: %s\n\n", input.Title))
	}

	if input.Phase != "" {
		contentBuilder.WriteString(fmt.Sprintf("## Current Phase: %s\n\n", input.Phase))
	}

	// Key context section
	if len(input.KeyContext) > 0 {
		contentBuilder.WriteString("## Key Context\n\n")
		for _, context := range input.KeyContext {
			contentBuilder.WriteString(fmt.Sprintf("- %s\n", context))
		}
		contentBuilder.WriteString("\n")
	}

	// Next steps section
	if len(input.NextSteps) > 0 {
		contentBuilder.WriteString("## Next Steps\n\n")
		for i, step := range input.NextSteps {
			contentBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		contentBuilder.WriteString("\n")
	}

	// Truncation point if specified
	if input.TruncationPoint != "" {
		contentBuilder.WriteString(fmt.Sprintf("## Context Truncation Point\n\n%s\n\n", input.TruncationPoint))
	}

	// Full conversation content
	contentBuilder.WriteString("## Full Conversation\n\n")
	if input.Content != "" {
		contentBuilder.WriteString(input.Content)
	} else {
		contentBuilder.WriteString("_No conversation content provided_")
	}

	// Prepare labels for the checkpoint
	labels := make(map[string]string)
	if input.Labels != nil {
		for k, v := range input.Labels {
			labels[k] = v
		}
	}

	// Add standard checkpoint labels
	labels["type"] = "conversation-checkpoint"
	labels["conversation-id"] = conversationID
	labels["auto-generated"] = fmt.Sprintf("%t", input.AutoGenerated)
	if input.Phase != "" {
		labels["phase"] = input.Phase
	}

	// Create the memory
	createReq := storage.CreateMemoryRequest{
		Name:    checkpointName,
		Content: contentBuilder.String(),
		Labels:  labels,
		Metadata: map[string]any{
			"namespace":        namespace,
			"conversation-id":  conversationID,
			"checkpoint-type":  "full",
			"auto-generated":   input.AutoGenerated,
			"truncation-point": input.TruncationPoint,
		},
	}

	memory, err := s.provider.Create(createReq)
	if err != nil {
		return nil, ConversationCheckpointOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to create checkpoint: %v", err),
		}, nil
	}

	result := ConversationCheckpointOutput{
		ConversationID: conversationID,
		CheckpointID:   memory.ID,
		Name:           memory.Name,
		Timestamp:      timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Success:        true,
		Message:        fmt.Sprintf("Conversation checkpoint created successfully: %s", memory.Name),
	}

	return nil, result, nil
}

// === PHASE 2: ENHANCED AUTOMATION IMPLEMENTATIONS ===

// CreateIncrementalSnapshot implements the contextmemory_snapshot tool
func (s *Server) CreateIncrementalSnapshot(ctx context.Context, req *mcp.CallToolRequest, input IncrementalSnapshotInput) (
	*mcp.CallToolResult,
	IncrementalSnapshotOutput,
	error,
) {
	log.Printf("CreateIncrementalSnapshot called with input: %+v", input)

	// Set defaults
	namespace := input.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if input.ConversationID == "" {
		return nil, IncrementalSnapshotOutput{}, fmt.Errorf("conversation ID is required")
	}

	// Generate snapshot name
	timestamp := time.Now()
	snapshotName := fmt.Sprintf("%s_snapshot_%s", input.ConversationID, timestamp.Format("2006_01_02_15_04_05"))

	// Estimate token count if not provided
	tokenCount := input.TokenCount
	if tokenCount == 0 && input.RecentContent != "" {
		tokenCount = estimateTokenCount(input.RecentContent)
	}

	// Prepare snapshot content in structured format
	var contentBuilder strings.Builder

	// Header with metadata
	contentBuilder.WriteString(fmt.Sprintf("# Incremental Snapshot: %s\n", input.ConversationID))
	contentBuilder.WriteString(fmt.Sprintf("_Created on %s_\n\n", timestamp.Format("02/01/2006 at 15:04:05 MST")))

	// Snapshot metadata
	contentBuilder.WriteString("## Snapshot Metadata\n\n")
	contentBuilder.WriteString(fmt.Sprintf("- **Token Count**: %d\n", tokenCount))
	contentBuilder.WriteString(fmt.Sprintf("- **Messages Since Last Checkpoint**: %d\n", input.MessagesSince))
	contentBuilder.WriteString(fmt.Sprintf("- **Auto Generated**: %t\n", input.AutoGenerated))
	if input.TriggerReason != "" {
		contentBuilder.WriteString(fmt.Sprintf("- **Trigger Reason**: %s\n", input.TriggerReason))
	}
	contentBuilder.WriteString("\n")

	// Context summary if provided
	if input.ContextSummary != "" {
		contentBuilder.WriteString("## Context Summary\n\n")
		contentBuilder.WriteString(input.ContextSummary)
		contentBuilder.WriteString("\n\n")
	}

	// Recent conversation content
	contentBuilder.WriteString("## Recent Conversation Content\n\n")
	if input.RecentContent != "" {
		contentBuilder.WriteString(input.RecentContent)
	} else {
		contentBuilder.WriteString("_No recent content provided_")
	}

	// Prepare labels for the snapshot
	labels := make(map[string]string)
	if input.Labels != nil {
		for k, v := range input.Labels {
			labels[k] = v
		}
	}

	// Add standard snapshot labels
	labels["type"] = "incremental-snapshot"
	labels["conversation-id"] = input.ConversationID
	labels["auto-generated"] = fmt.Sprintf("%t", input.AutoGenerated)
	labels["token-count"] = fmt.Sprintf("%d", tokenCount)
	if input.TriggerReason != "" {
		labels["trigger-reason"] = input.TriggerReason
	}

	// Create the memory
	createReq := storage.CreateMemoryRequest{
		Name:    snapshotName,
		Content: contentBuilder.String(),
		Labels:  labels,
		Metadata: map[string]any{
			"namespace":       namespace,
			"conversation-id": input.ConversationID,
			"snapshot-type":   "incremental",
			"auto-generated":  input.AutoGenerated,
			"token-count":     tokenCount,
			"messages-since":  input.MessagesSince,
		},
	}

	memory, err := s.provider.Create(createReq)
	if err != nil {
		return nil, IncrementalSnapshotOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to create snapshot: %v", err),
		}, nil
	}

	result := IncrementalSnapshotOutput{
		ConversationID: input.ConversationID,
		SnapshotID:     memory.ID,
		Name:           memory.Name,
		Timestamp:      timestamp.Format("2006-01-02T15:04:05Z07:00"),
		TokenCount:     tokenCount,
		Success:        true,
		Message:        fmt.Sprintf("Incremental snapshot created successfully: %s", memory.Name),
	}

	return nil, result, nil
}

// MonitorConversation implements the contextmemory_monitor tool
func (s *Server) MonitorConversation(ctx context.Context, req *mcp.CallToolRequest, input ConversationMonitorInput) (
	*mcp.CallToolResult,
	ConversationMonitorOutput,
	error,
) {
	log.Printf("MonitorConversation called with input: %+v", input)

	if input.ConversationID == "" {
		return nil, ConversationMonitorOutput{}, fmt.Errorf("conversation ID is required")
	}

	// Get monitoring configuration
	config := getDefaultMonitorConfig()

	// Estimate token count if not provided
	tokenCount := input.EstimatedTokens
	if tokenCount == 0 && input.CurrentContent != "" {
		tokenCount = estimateTokenCount(input.CurrentContent)
	}

	// Find latest checkpoint and snapshot
	var lastCheckpointTime, lastSnapshotTime string

	latestCheckpoint, err := s.findLatestCheckpoint(input.ConversationID)
	if err != nil {
		log.Printf("Error finding latest checkpoint: %v", err)
	} else if latestCheckpoint != nil {
		lastCheckpointTime = latestCheckpoint.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	latestSnapshot, err := s.findLatestSnapshot(input.ConversationID)
	if err != nil {
		log.Printf("Error finding latest snapshot: %v", err)
	} else if latestSnapshot != nil {
		lastSnapshotTime = latestSnapshot.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	// Analyze conversation health
	health := analyzeConversationHealth(tokenCount, input.MessageCount, config, lastCheckpointTime, lastSnapshotTime)
	health.ConversationID = input.ConversationID

	return nil, health, nil
}

// CreateConversationThread implements the contextmemory_thread tool
func (s *Server) CreateConversationThread(ctx context.Context, req *mcp.CallToolRequest, input ConversationThreadInput) (
	*mcp.CallToolResult,
	ConversationThreadOutput,
	error,
) {
	log.Printf("CreateConversationThread called with input: %+v", input)

	// Set defaults
	namespace := "default"

	if input.ConversationID == "" {
		return nil, ConversationThreadOutput{}, fmt.Errorf("conversation ID is required")
	}

	// Generate thread name
	timestamp := time.Now()
	threadName := fmt.Sprintf("%s_thread", input.ConversationID)

	// Count existing checkpoints and snapshots for this conversation
	checkpointCount := 0
	snapshotCount := 0

	checkpoints, err := s.provider.Search(storage.SearchRequest{
		LabelSelector: map[string]string{
			"type":            "conversation-checkpoint",
			"conversation-id": input.ConversationID,
		},
		UseIndex: true,
	})
	if err == nil {
		checkpointCount = len(checkpoints.Memories)
	}

	snapshots, err := s.provider.Search(storage.SearchRequest{
		LabelSelector: map[string]string{
			"type":            "incremental-snapshot",
			"conversation-id": input.ConversationID,
		},
		UseIndex: true,
	})
	if err == nil {
		snapshotCount = len(snapshots.Memories)
	}

	// Find related threads (same project or parent)
	var relatedThreads []string
	if input.Project != "" || input.ParentThread != "" {
		relatedSearch := storage.SearchRequest{
			LabelSelector: map[string]string{
				"type": "conversation-thread",
			},
			UseIndex: true,
		}

		if input.Project != "" {
			relatedSearch.LabelSelector["project"] = input.Project
		}

		relatedResults, err := s.provider.Search(relatedSearch)
		if err == nil {
			for _, memory := range relatedResults.Memories {
				if convID, ok := memory.Labels["conversation-id"]; ok && convID != input.ConversationID {
					relatedThreads = append(relatedThreads, convID)
				}
			}
		}
	}

	// Prepare thread content
	var contentBuilder strings.Builder

	contentBuilder.WriteString(fmt.Sprintf("# Conversation Thread: %s\n", input.ConversationID))
	contentBuilder.WriteString(fmt.Sprintf("_Created on %s_\n\n", timestamp.Format("02/01/2006 at 15:04:05 MST")))

	if input.Title != "" {
		contentBuilder.WriteString(fmt.Sprintf("## Title: %s\n\n", input.Title))
	}

	contentBuilder.WriteString("## Thread Metadata\n\n")
	if input.Project != "" {
		contentBuilder.WriteString(fmt.Sprintf("- **Project**: %s\n", input.Project))
	}
	if input.Phase != "" {
		contentBuilder.WriteString(fmt.Sprintf("- **Phase**: %s\n", input.Phase))
	}
	if input.Status != "" {
		contentBuilder.WriteString(fmt.Sprintf("- **Status**: %s\n", input.Status))
	}
	if input.ParentThread != "" {
		contentBuilder.WriteString(fmt.Sprintf("- **Parent Thread**: %s\n", input.ParentThread))
	}
	contentBuilder.WriteString(fmt.Sprintf("- **Total Checkpoints**: %d\n", checkpointCount))
	contentBuilder.WriteString(fmt.Sprintf("- **Total Snapshots**: %d\n", snapshotCount))
	if input.EstimatedTokens > 0 {
		contentBuilder.WriteString(fmt.Sprintf("- **Estimated Tokens**: %d\n", input.EstimatedTokens))
	}
	if input.MessageCount > 0 {
		contentBuilder.WriteString(fmt.Sprintf("- **Message Count**: %d\n", input.MessageCount))
	}
	contentBuilder.WriteString("\n")

	if len(relatedThreads) > 0 {
		contentBuilder.WriteString("## Related Threads\n\n")
		for _, threadID := range relatedThreads {
			contentBuilder.WriteString(fmt.Sprintf("- %s\n", threadID))
		}
		contentBuilder.WriteString("\n")
	}

	// Prepare labels
	labels := make(map[string]string)
	if input.Labels != nil {
		for k, v := range input.Labels {
			labels[k] = v
		}
	}

	// Add standard thread labels
	labels["type"] = "conversation-thread"
	labels["conversation-id"] = input.ConversationID
	if input.Project != "" {
		labels["project"] = input.Project
	}
	if input.Phase != "" {
		labels["phase"] = input.Phase
	}
	if input.Status != "" {
		labels["status"] = input.Status
	}
	if input.ParentThread != "" {
		labels["parent-thread"] = input.ParentThread
	}

	// Create the thread memory
	createReq := storage.CreateMemoryRequest{
		Name:    threadName,
		Content: contentBuilder.String(),
		Labels:  labels,
		Metadata: map[string]any{
			"namespace":         namespace,
			"conversation-id":   input.ConversationID,
			"thread-type":       "conversation",
			"total-checkpoints": checkpointCount,
			"total-snapshots":   snapshotCount,
		},
	}

	memory, err := s.provider.Create(createReq)
	if err != nil {
		return nil, ConversationThreadOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to create thread: %v", err),
		}, nil
	}

	result := ConversationThreadOutput{
		ConversationID:   input.ConversationID,
		ThreadID:         memory.ID,
		Name:             memory.Name,
		RelatedThreads:   relatedThreads,
		TotalCheckpoints: checkpointCount,
		TotalSnapshots:   snapshotCount,
		Success:          true,
		Message:          fmt.Sprintf("Conversation thread created successfully: %s", memory.Name),
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
		Version: version,
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_checkpoint",
		Description: "Create a conversation checkpoint to preserve context before truncation",
	}, contextMemoryServer.CreateConversationCheckpoint)

	// Phase 2: Enhanced Automation Tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_snapshot",
		Description: "Create an incremental snapshot for lightweight context capture",
	}, contextMemoryServer.CreateIncrementalSnapshot)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_monitor",
		Description: "Monitor conversation health and get automated recommendations",
	}, contextMemoryServer.MonitorConversation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "contextmemory_thread",
		Description: "Create and manage conversation threads for relationship tracking",
	}, contextMemoryServer.CreateConversationThread)

	// Set up logging to stderr to avoid JSON-RPC interference
	log.SetOutput(os.Stderr)
	log.Println("Starting ContextMemory MCP server with official SDK...")

	// Run the server over stdin/stdout until client disconnects
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal("Server failed:", err)
	}
}
