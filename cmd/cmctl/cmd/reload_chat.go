package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	reloadSearch      string
	reloadLanguage    string
	reloadActivity    string
	reloadDate        string
	reloadLimit       int
	reloadFormat      string
	reloadInteractive bool
	reloadMemoryID    string
)

// reloadChatCmd represents the reload-chat command
var reloadChatCmd = &cobra.Command{
	Use:   "reload-chat [memory-id]",
	Short: "Reload previously captured chat conversations as context",
	Long: `Reload previously captured chat conversations to use as context in fresh AI conversations.

This command helps you search and retrieve stored chat memories, formatting them for easy
copy/paste into new AI conversations or automatic loading into Cursor's AI pane.

Output Formats:
  conversational    Full chat history with user/assistant labels (default)
  context-only      Clean context without chat formatting
  summary           Condensed version with key points
  raw              Original markdown format

Examples:
  # Interactive mode - search and select from available chats
  cmctl reload-chat --interactive

  # Search for specific topics
  cmctl reload-chat --search "authentication React"
  cmctl reload-chat --language javascript --activity debugging

  # Reload specific chat by ID
  cmctl reload-chat mem_abc123_def456

  # Different output formats
  cmctl reload-chat --search "React hooks" --format context-only
  cmctl reload-chat mem_abc123 --format summary`,
	RunE: runReloadChat,
}

func init() {
	rootCmd.AddCommand(reloadChatCmd)

	reloadChatCmd.Flags().StringVarP(&reloadSearch, "search", "s", "", "Search chat content and titles")
	reloadChatCmd.Flags().StringVarP(&reloadLanguage, "language", "l", "", "Filter by programming language")
	reloadChatCmd.Flags().StringVarP(&reloadActivity, "activity", "a", "", "Filter by activity type (debugging, implementation, learning, etc.)")
	reloadChatCmd.Flags().StringVarP(&reloadDate, "date", "d", "", "Filter by date (YYYY-MM-DD or relative like 'today', 'yesterday', 'week')")
	reloadChatCmd.Flags().IntVar(&reloadLimit, "limit", 10, "Limit number of results to show")
	reloadChatCmd.Flags().StringVarP(&reloadFormat, "format", "f", "conversational", "Output format: conversational|context-only|summary|raw")
	reloadChatCmd.Flags().BoolVarP(&reloadInteractive, "interactive", "i", false, "Interactive mode to browse and select chats")
	reloadChatCmd.Flags().StringVar(&reloadMemoryID, "memory-id", "", "Specific memory ID to reload (alternative to positional arg)")
}

func runReloadChat(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Handle specific memory ID
	if len(args) > 0 || reloadMemoryID != "" {
		memoryID := reloadMemoryID
		if len(args) > 0 {
			memoryID = args[0]
		}
		return reloadSpecificChat(fs, memoryID)
	}

	// Interactive mode
	if reloadInteractive {
		return runInteractiveReload(fs)
	}

	// Search and list mode
	return runSearchAndReload(fs)
}

func reloadSpecificChat(fs *storage.FileStorage, memoryID string) error {
	memory, err := fs.Get(memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	// Verify it's a chat memory
	if memory.Labels["type"] != "chat" {
		return fmt.Errorf("memory %s is not a chat conversation (type=%s)", memoryID, memory.Labels["type"])
	}

	output := formatChatForReload(*memory, reloadFormat)
	fmt.Print(output)
	return nil
}

func runSearchAndReload(fs *storage.FileStorage) error {
	// Build search criteria
	req := storage.SearchRequest{
		LabelSelector:  map[string]string{"type": "chat"},
		Limit:          reloadLimit,
		UseIndex:       true,
		IncludeContent: false, // We'll load content only for matches
	}

	// Add filters
	if reloadLanguage != "" {
		req.LabelSelector["language"] = reloadLanguage
	}
	if reloadActivity != "" {
		req.LabelSelector["activity"] = reloadActivity
	}
	if reloadDate != "" {
		dateFilter, err := parseDateFilter(reloadDate)
		if err != nil {
			return fmt.Errorf("invalid date filter: %w", err)
		}
		req.LabelSelector["date"] = dateFilter
	}

	// Add text search if specified
	if reloadSearch != "" {
		req.Query = reloadSearch
		req.IncludeContent = true // Need content for text search
	}

	// Search for chat memories
	result, err := fs.Search(req)
	if err != nil {
		return fmt.Errorf("failed to search chat memories: %w", err)
	}

	if len(result.Memories) == 0 {
		fmt.Println("No chat memories found matching the criteria.")
		fmt.Println("\nTry:")
		fmt.Println("  cmctl reload-chat --interactive    # Browse all available chats")
		fmt.Println("  cmctl reload-chat --search 'topic' # Search for specific topics")
		fmt.Println("  cmctl list-cursor-chats           # Import new chats from Cursor")
		return nil
	}

	// If only one result, output it directly
	if len(result.Memories) == 1 {
		// Load full content if we don't have it
		if result.Memories[0].Content == "" {
			fullMemory, err := fs.Get(result.Memories[0].ID)
			if err != nil {
				return fmt.Errorf("failed to load memory content: %w", err)
			}
			result.Memories[0] = *fullMemory
		}

		output := formatChatForReload(result.Memories[0], reloadFormat)
		fmt.Print(output)
		return nil
	}

	// Multiple results - show selection list
	return showChatSelection(fs, result.Memories)
}

func runInteractiveReload(fs *storage.FileStorage) error {
	// Get all chat memories
	req := storage.SearchRequest{
		LabelSelector:  map[string]string{"type": "chat"},
		Limit:          100, // Show more for interactive mode
		UseIndex:       true,
		IncludeContent: false,
	}

	result, err := fs.Search(req)
	if err != nil {
		return fmt.Errorf("failed to search chat memories: %w", err)
	}

	if len(result.Memories) == 0 {
		fmt.Println("No chat memories found.")
		fmt.Println("\nTo import chats from Cursor:")
		fmt.Println("  cmctl import-cursor-chat --latest")
		fmt.Println("  cmctl list-cursor-chats")
		return nil
	}

	return showChatSelection(fs, result.Memories)
}

func showChatSelection(fs *storage.FileStorage, memories []storage.Memory) error {
	// Sort by creation date (newest first)
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].CreatedAt.After(memories[j].CreatedAt)
	})

	fmt.Printf("Found %d chat memories:\n\n", len(memories))

	for i, memory := range memories {
		fmt.Printf("%d. %s\n", i+1, memory.Name)
		fmt.Printf("   Date: %s", memory.CreatedAt.Format("2006-01-02 15:04"))

		if lang := memory.Labels["language"]; lang != "" {
			fmt.Printf(" | Language: %s", lang)
		}
		if activity := memory.Labels["activity"]; activity != "" {
			fmt.Printf(" | Activity: %s", activity)
		}
		fmt.Printf("\n")

		if len(memory.Content) > 0 {
			// Show preview if we have content
			preview := extractContentPreview(memory.Content, 100)
			fmt.Printf("   Preview: %s\n", preview)
		}
		fmt.Println()
	}

	fmt.Printf("Enter the number of the chat to reload (1-%d), or 0 to cancel: ", len(memories))
	var choice string
	if _, err := fmt.Scanln(&choice); err != nil {
		fmt.Println("Invalid input. Cancelled.")
		return nil
	}

	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 0 || choiceNum > len(memories) {
		fmt.Println("Invalid choice.")
		return nil
	}

	if choiceNum == 0 {
		fmt.Println("Cancelled.")
		return nil
	}

	selectedMemory := memories[choiceNum-1]

	// Load full content if needed
	if selectedMemory.Content == "" {
		fullMemory, err := fs.Get(selectedMemory.ID)
		if err != nil {
			return fmt.Errorf("failed to load memory content: %w", err)
		}
		selectedMemory = *fullMemory
	}

	fmt.Printf("\n--- Loading Chat: %s ---\n\n", selectedMemory.Name)
	output := formatChatForReload(selectedMemory, reloadFormat)
	fmt.Print(output)

	return nil
}

func formatChatForReload(memory storage.Memory, format string) string {
	switch format {
	case "context-only":
		return formatAsContext(memory)
	case "summary":
		return formatAsSummary(memory)
	case "raw":
		return memory.Content
	default: // "conversational"
		return formatAsConversational(memory)
	}
}

func formatAsConversational(memory storage.Memory) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# Previous Conversation: %s\n\n", memory.Name))
	output.WriteString(fmt.Sprintf("*Captured on %s*\n\n", memory.CreatedAt.Format("2006-01-02 15:04")))

	if lang := memory.Labels["language"]; lang != "" {
		output.WriteString(fmt.Sprintf("*Language/Technology: %s*\n", lang))
	}
	if activity := memory.Labels["activity"]; activity != "" {
		output.WriteString(fmt.Sprintf("*Activity: %s*\n", activity))
	}

	output.WriteString("\n---\n\n")
	output.WriteString(memory.Content)
	output.WriteString("\n\n---\n\n")
	output.WriteString("*Please continue this conversation or use the above context for reference.*\n")

	return output.String()
}

func formatAsContext(memory storage.Memory) string {
	// Strip out the conversational markers and just provide clean context
	content := memory.Content

	// Remove markdown headers
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Skip markdown headers and date lines
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "**Date**:") {
			continue
		}

		// Convert user/assistant markers to simple context
		if strings.HasPrefix(line, "**User**: ") {
			line = "Question: " + strings.TrimPrefix(line, "**User**: ")
		} else if strings.HasPrefix(line, "**Assistant**: ") {
			line = "Response: " + strings.TrimPrefix(line, "**Assistant**: ")
		}

		cleanLines = append(cleanLines, line)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Context from previous session (%s):\n\n", memory.CreatedAt.Format("2006-01-02")))
	output.WriteString(strings.Join(cleanLines, "\n"))

	return output.String()
}

func formatAsSummary(memory storage.Memory) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("## Summary: %s\n\n", memory.Name))
	output.WriteString(fmt.Sprintf("*Session from %s*\n\n", memory.CreatedAt.Format("2006-01-02")))

	// Extract key topics and concepts
	if techs := memory.Labels["technologies"]; techs != "" {
		output.WriteString(fmt.Sprintf("**Technologies discussed**: %s\n", techs))
	}
	if lang := memory.Labels["language"]; lang != "" {
		output.WriteString(fmt.Sprintf("**Primary language**: %s\n", lang))
	}
	if activity := memory.Labels["activity"]; activity != "" {
		output.WriteString(fmt.Sprintf("**Activity type**: %s\n", activity))
	}

	output.WriteString("\n**Key points from conversation**:\n")

	// Extract key exchanges (simplified)
	content := memory.Content
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "**User**: ") {
			userQ := strings.TrimPrefix(line, "**User**: ")
			if len(userQ) > 100 {
				userQ = userQ[:97] + "..."
			}
			output.WriteString(fmt.Sprintf("- Asked about: %s\n", userQ))
		}
	}

	output.WriteString(fmt.Sprintf("\n*Full conversation available in memory: %s*\n", memory.ID))

	return output.String()
}

func extractContentPreview(content string, maxLength int) string {
	// Extract a meaningful preview from the chat content
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "**User**: ") {
			userContent := strings.TrimPrefix(line, "**User**: ")
			if len(userContent) > maxLength {
				return userContent[:maxLength-3] + "..."
			}
			return userContent
		}
	}

	// Fallback to first non-empty line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "**Date**:") {
			if len(line) > maxLength {
				return line[:maxLength-3] + "..."
			}
			return line
		}
	}

	return "No preview available"
}

func parseDateFilter(dateStr string) (string, error) {
	// Handle relative dates
	now := time.Now()

	switch strings.ToLower(dateStr) {
	case "today":
		return now.Format("2006-01-02"), nil
	case "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "week":
		return now.AddDate(0, 0, -7).Format("2006-01-02"), nil
	}

	// Try to parse as exact date
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return "", fmt.Errorf("invalid date format, use YYYY-MM-DD or 'today', 'yesterday', 'week'")
	}

	return dateStr, nil
}
