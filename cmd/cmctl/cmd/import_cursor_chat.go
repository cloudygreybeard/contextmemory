package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/cursor"
	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	importLatest    bool
	importTabID     string
	importWorkspace string
	importPreview   bool
)

// importCursorChatCmd represents the import-cursor-chat command
var importCursorChatCmd = &cobra.Command{
	Use:   "import-cursor-chat",
	Short: "Import chat from Cursor's AI pane",
	Long: `Import chat conversations from Cursor's AI pane into ContextMemory.

This command accesses Cursor's local database to extract chat conversations
and create memory entries with intelligent naming and labeling.

Examples:
  # Import the most recent chat
  cmctl import-cursor-chat --latest

  # Import a specific chat by ID
  cmctl import-cursor-chat --tab-id abc123

  # Preview available chats before importing
  cmctl import-cursor-chat --preview

  # Import from specific workspace
  cmctl import-cursor-chat --latest --workspace /path/to/state.vscdb`,
	RunE: runImportCursorChat,
}

func init() {
	rootCmd.AddCommand(importCursorChatCmd)

	importCursorChatCmd.Flags().BoolVar(&importLatest, "latest", false, "Import the most recent chat")
	importCursorChatCmd.Flags().StringVar(&importTabID, "tab-id", "", "Import specific chat by tab ID")
	importCursorChatCmd.Flags().StringVar(&importWorkspace, "workspace", "", "Path to specific workspace database")
	importCursorChatCmd.Flags().BoolVar(&importPreview, "preview", false, "Preview available chats without importing")
}

func runImportCursorChat(cmd *cobra.Command, args []string) error {
	// Initialize workspace reader
	var reader *cursor.WorkspaceReader
	if importWorkspace != "" {
		reader = cursor.NewWorkspaceReaderWithPath(importWorkspace)
	} else {
		reader = cursor.NewWorkspaceReader()
	}

	if importPreview {
		return previewCursorChats(reader)
	}

	if !importLatest && importTabID == "" {
		return fmt.Errorf("must specify either --latest or --tab-id")
	}

	var chatTab *cursor.ChatTab
	var err error

	if importLatest {
		chatTab, err = reader.GetLatestChat()
		if err != nil {
			return fmt.Errorf("failed to get latest chat: %w", err)
		}
	} else {
		chatTab, _, err = reader.GetChatByID(importTabID)
		if err != nil {
			return fmt.Errorf("failed to get chat by ID: %w", err)
		}
	}

	// Convert chat to memory format
	memory := convertChatToMemory(chatTab)

	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	provider, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Create the memory
	createdMemory, err := provider.Create(memory)
	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	fmt.Printf("Successfully imported chat as memory:\n")
	fmt.Printf("ID: %s\n", createdMemory.ID)
	fmt.Printf("Name: %s\n", createdMemory.Name)
	fmt.Printf("Labels: %v\n", createdMemory.Labels)
	fmt.Printf("Content: %d characters\n", len(createdMemory.Content))

	return nil
}

func previewCursorChats(reader *cursor.WorkspaceReader) error {
	chats, err := reader.ListAllChats()
	if err != nil {
		return fmt.Errorf("failed to list chats: %w", err)
	}

	if len(chats) == 0 {
		fmt.Println("No chats found in Cursor workspaces")
		return nil
	}

	fmt.Printf("Found %d chat(s) across workspaces:\n\n", len(chats))

	for i, chat := range chats {
		if i >= 10 { // Limit preview to 10 chats
			fmt.Printf("... and %d more\n", len(chats)-10)
			break
		}

		fmt.Printf("Chat %d:\n", i+1)
		fmt.Printf("  ID: %s\n", chat.ID)
		fmt.Printf("  Title: %s\n", chat.GetDisplayTitle())
		fmt.Printf("  Workspace: %s\n", chat.WorkspaceName)
		fmt.Printf("  Messages: %d\n", len(chat.Messages))
		if chat.Timestamp > 0 {
			timestamp := time.Unix(chat.Timestamp/1000, 0)
			fmt.Printf("  Date: %s\n", timestamp.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("  Preview: %s\n", truncateString(chat.GetContentPreview(100), 100))
		fmt.Println()
	}

	return nil
}

func convertChatToMemory(chatTab *cursor.ChatTab) storage.CreateMemoryRequest {
	// Generate intelligent name
	name := generateChatMemoryName(chatTab)

	// Generate labels based on chat analysis
	labels := generateChatLabels(chatTab)

	// Convert to markdown content
	content := chatTab.ToMarkdown()

	return storage.CreateMemoryRequest{
		Name:    name,
		Content: content,
		Labels:  labels,
	}
}

func generateChatMemoryName(chatTab *cursor.ChatTab) string {
	// Try to extract topic from title or first message
	title := chatTab.GetDisplayTitle()
	if title != "Untitled Chat" && title != "AI Service Chat" && len(title) > 0 {
		return cleanChatTitle(title)
	}

	// Generate from first meaningful user message
	for _, msg := range chatTab.Messages {
		if msg.Role == "user" && len(strings.TrimSpace(msg.Content)) > 0 {
			content := strings.TrimSpace(msg.Content)

			// Skip file references and common prefixes
			if strings.HasPrefix(content, "@") {
				continue
			}

			// Extract meaningful title from first sentence
			sentences := strings.Split(content, ".")
			if len(sentences) > 0 {
				firstSentence := strings.TrimSpace(sentences[0])
				if len(firstSentence) > 10 {
					// Limit length and clean up
					if len(firstSentence) > 60 {
						firstSentence = firstSentence[:57] + "..."
					}
					return strings.Title(strings.ToLower(firstSentence))
				}
			}
		}
	}

	// Analyze technical concepts
	concepts := chatTab.ExtractTechnicalConcepts()
	if len(concepts) > 0 {
		primaryConcept := concepts[0]
		if len(concepts) > 1 {
			return fmt.Sprintf("%s Development Discussion", strings.Title(primaryConcept))
		}
		return fmt.Sprintf("%s Chat", strings.Title(primaryConcept))
	}

	// Fallback to date-based naming
	if chatTab.Timestamp > 0 {
		timestamp := time.Unix(chatTab.Timestamp/1000, 0)
		return fmt.Sprintf("Development Session %s", timestamp.Format("2006-01-02"))
	}

	return "Cursor Chat Session"
}

func generateChatLabels(chatTab *cursor.ChatTab) map[string]string {
	labels := map[string]string{
		"type":   "chat",
		"source": "cursor-ai-pane",
	}

	// Add date
	if chatTab.Timestamp > 0 {
		timestamp := time.Unix(chatTab.Timestamp/1000, 0)
		labels["date"] = timestamp.Format("2006-01-02")
	}

	// Add technical concepts as labels
	concepts := chatTab.ExtractTechnicalConcepts()
	if len(concepts) > 0 {
		labels["language"] = concepts[0] // Primary language/concept
		if len(concepts) > 1 {
			labels["technologies"] = strings.Join(concepts[:3], ",") // Up to 3 technologies
		}
	}

	// Analyze activity type
	content := strings.ToLower(chatTab.ToMarkdown())
	activityPatterns := map[string]string{
		"debug":     "debugging",
		"error":     "debugging",
		"implement": "implementation",
		"create":    "implementation",
		"build":     "implementation",
		"review":    "code-review",
		"refactor":  "refactoring",
		"optimize":  "optimization",
		"test":      "testing",
		"explain":   "learning",
		"how":       "learning",
		"what":      "learning",
	}

	for pattern, activity := range activityPatterns {
		if strings.Contains(content, pattern) {
			labels["activity"] = activity
			break
		}
	}

	return labels
}

func cleanChatTitle(title string) string {
	// Remove common prefixes and clean up
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "Chat: ")
	title = strings.TrimPrefix(title, "Discussion: ")

	// Keep title verbatim - no capitalization changes

	return title
}
