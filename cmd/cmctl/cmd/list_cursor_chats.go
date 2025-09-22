package cmd

import (
	"fmt"
	"time"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/cursor"
	"github.com/spf13/cobra"
)

var (
	listWorkspace string
	listSearch    string
	listLimit     int
)

// listCursorChatsCmd represents the list-cursor-chats command
var listCursorChatsCmd = &cobra.Command{
	Use:   "list-cursor-chats",
	Short: "List available chats from Cursor's AI pane",
	Long: `List chat conversations available in Cursor's workspace storage.

This command helps you discover what chats are available for import
from Cursor's AI pane across all workspaces.

Examples:
  # List all available chats
  cmctl list-cursor-chats

  # Search for chats containing specific text
  cmctl list-cursor-chats --search "authentication"

  # List chats from specific workspace
  cmctl list-cursor-chats --workspace /path/to/state.vscdb

  # Limit number of results
  cmctl list-cursor-chats --limit 5`,
	RunE: runListCursorChats,
}

func init() {
	rootCmd.AddCommand(listCursorChatsCmd)

	listCursorChatsCmd.Flags().StringVar(&listWorkspace, "workspace", "", "Path to specific workspace database")
	listCursorChatsCmd.Flags().StringVar(&listSearch, "search", "", "Search for chats containing text")
	listCursorChatsCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum number of chats to show")
}

func runListCursorChats(cmd *cobra.Command, args []string) error {
	// Initialize workspace reader
	var reader *cursor.WorkspaceReader
	if listWorkspace != "" {
		reader = cursor.NewWorkspaceReaderWithPath(listWorkspace)
	} else {
		reader = cursor.NewWorkspaceReader()
	}

	var chats []cursor.ChatTabWithWorkspace
	var err error

	if listSearch != "" {
		chats, err = reader.SearchChats(listSearch)
		if err != nil {
			return fmt.Errorf("failed to search chats: %w", err)
		}
	} else {
		chats, err = reader.ListAllChats()
		if err != nil {
			return fmt.Errorf("failed to list chats: %w", err)
		}
	}

	if len(chats) == 0 {
		if listSearch != "" {
			fmt.Printf("No chats found matching '%s'\n", listSearch)
		} else {
			fmt.Println("No chats found in Cursor workspaces")
		}
		return nil
	}

	// Apply limit
	if listLimit > 0 && len(chats) > listLimit {
		chats = chats[:listLimit]
	}

	// Display results
	if listSearch != "" {
		fmt.Printf("Found %d chat(s) matching '%s':\n\n", len(chats), listSearch)
	} else {
		fmt.Printf("Found %d chat(s) across workspaces:\n\n", len(chats))
	}

	for i, chat := range chats {
		fmt.Printf("Chat %d:\n", i+1)
		fmt.Printf("  ID: %s\n", chat.ID)
		fmt.Printf("  Title: %s\n", chat.GetDisplayTitle())
		fmt.Printf("  Workspace: %s\n", chat.WorkspaceName)
		fmt.Printf("  Messages: %d\n", len(chat.Messages))

		if chat.Timestamp > 0 {
			timestamp := time.Unix(chat.Timestamp/1000, 0)
			fmt.Printf("  Date: %s\n", timestamp.Format("2006-01-02 15:04:05"))
		}

		// Show technical concepts if found
		concepts := chat.ExtractTechnicalConcepts()
		if len(concepts) > 0 {
			conceptsStr := concepts[0]
			if len(concepts) > 1 {
				conceptsStr += fmt.Sprintf(" (+%d more)", len(concepts)-1)
			}
			fmt.Printf("  Concepts: %s\n", conceptsStr)
		}

		fmt.Printf("  Preview: %s\n", truncateString(chat.GetContentPreview(150), 150))
		fmt.Println()
	}

	if listLimit > 0 && len(chats) == listLimit {
		fmt.Printf("... (showing first %d results, use --limit to see more)\n", listLimit)
	}

	return nil
}
