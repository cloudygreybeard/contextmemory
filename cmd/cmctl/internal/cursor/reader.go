package cursor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WorkspaceReader provides access to Cursor's workspace storage
type WorkspaceReader struct {
	StoragePath string
}

// NewWorkspaceReader creates a new workspace reader
func NewWorkspaceReader() *WorkspaceReader {
	return &WorkspaceReader{
		StoragePath: getDefaultStoragePath(),
	}
}

// NewWorkspaceReaderWithPath creates a reader with custom storage path
func NewWorkspaceReaderWithPath(path string) *WorkspaceReader {
	return &WorkspaceReader{
		StoragePath: path,
	}
}

// getDefaultStoragePath returns the default Cursor workspace storage path
func getDefaultStoragePath() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Cursor", "User", "workspaceStorage")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "workspaceStorage")
	case "linux":
		return filepath.Join(homeDir, ".config", "Cursor", "User", "workspaceStorage")
	default:
		return filepath.Join(homeDir, ".cursor", "workspaceStorage")
	}
}

// FindWorkspaces returns all available workspace database paths
func (wr *WorkspaceReader) FindWorkspaces() ([]string, error) {
	entries, err := os.ReadDir(wr.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace storage: %w", err)
	}

	var workspaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			dbPath := filepath.Join(wr.StoragePath, entry.Name(), "state.vscdb")
			if _, err := os.Stat(dbPath); err == nil {
				workspaces = append(workspaces, dbPath)
			}
		}
	}

	return workspaces, nil
}

// GetLatestWorkspace returns the most recently modified workspace
func (wr *WorkspaceReader) GetLatestWorkspace() (string, error) {
	workspaces, err := wr.FindWorkspaces()
	if err != nil {
		return "", err
	}

	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces found in %s", wr.StoragePath)
	}

	// Sort by modification time
	sort.Slice(workspaces, func(i, j int) bool {
		infoI, errI := os.Stat(workspaces[i])
		infoJ, errJ := os.Stat(workspaces[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return workspaces[0], nil
}

// OpenWorkspaceDB opens a GORM connection to a workspace database
func (wr *WorkspaceReader) OpenWorkspaceDB(dbPath string) (*gorm.DB, error) {
	// Configure GORM with pure Go SQLite driver
	db, err := gorm.Open(sqlite.Open(dbPath+"?mode=ro"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open workspace database: %w", err)
	}

	return db, nil
}

// GetChatData retrieves and parses chat data from workspace
func (wr *WorkspaceReader) GetChatData(dbPath string) (*ChatData, error) {
	db, err := wr.OpenWorkspaceDB(dbPath)
	if err != nil {
		return nil, err
	}

	chatData := &ChatData{Tabs: []ChatTab{}}
	
	// First, get composer data to extract titles
	composerTitles := make(map[string]string) // map[composerID]title
	var composerItem CursorItem
	if result := db.Where("key = ?", "composer.composerData").First(&composerItem); result.Error == nil {
		var composerData ComposerData
		if err := json.Unmarshal([]byte(composerItem.Value), &composerData); err == nil {
			for _, composer := range composerData.AllComposers {
				if composer.Name != "" {
					composerTitles[composer.ComposerID] = composer.Name
				}
			}
		}
	}

	// Try different possible chat data keys (different Cursor versions)
	chatKeys := []string{
		"workbench.panel.aichat.view.aichat.chatdata", // Newer format with actual titles
		"aiService.prompts",                            // Legacy format
		"composer.composerData",                        // Composer chats
	}

	for _, key := range chatKeys {
		var item CursorItem
		result := db.Where("key = ?", key).First(&item)
		if result.Error != nil {
			continue // Try next key
		}

		// Parse based on key type
		if key == "workbench.panel.aichat.view.aichat.chatdata" {
			// Modern Cursor format with proper titles
			var tempData ChatData
			if err := json.Unmarshal([]byte(item.Value), &tempData); err == nil {
				chatData.Tabs = append(chatData.Tabs, tempData.Tabs...)
			}
		} else if key == "aiService.prompts" {
			tabs, err := wr.parseAIServicePromptsWithTitles(item.Value, composerTitles)
			if err == nil && len(tabs) > 0 {
				chatData.Tabs = append(chatData.Tabs, tabs...)
			}
		} else if key == "composer.composerData" {
			tabs, err := wr.parseComposerData(item.Value)
			if err == nil && len(tabs) > 0 {
				chatData.Tabs = append(chatData.Tabs, tabs...)
			}
		} else {
			// Fallback format
			var tempData ChatData
			if err := json.Unmarshal([]byte(item.Value), &tempData); err == nil {
				chatData.Tabs = append(chatData.Tabs, tempData.Tabs...)
			}
		}
	}

	return chatData, nil
}

// GetLatestChat returns the most recent chat from the latest workspace
func (wr *WorkspaceReader) GetLatestChat() (*ChatTab, error) {
	latestWorkspace, err := wr.GetLatestWorkspace()
	if err != nil {
		return nil, err
	}

	chatData, err := wr.GetChatData(latestWorkspace)
	if err != nil {
		return nil, err
	}

	if len(chatData.Tabs) == 0 {
		return nil, fmt.Errorf("no chats found in latest workspace")
	}

	// Sort by timestamp to get latest
	sort.Slice(chatData.Tabs, func(i, j int) bool {
		return chatData.Tabs[i].Timestamp > chatData.Tabs[j].Timestamp
	})

	return &chatData.Tabs[0], nil
}

// GetChatByID retrieves a specific chat by its ID
func (wr *WorkspaceReader) GetChatByID(chatID string) (*ChatTab, string, error) {
	workspaces, err := wr.FindWorkspaces()
	if err != nil {
		return nil, "", err
	}

	// Search through all workspaces for the chat ID
	for _, workspacePath := range workspaces {
		chatData, err := wr.GetChatData(workspacePath)
		if err != nil {
			continue // Skip errored workspaces
		}

		for _, tab := range chatData.Tabs {
			if tab.ID == chatID {
				return &tab, workspacePath, nil
			}
		}
	}

	return nil, "", fmt.Errorf("chat with ID %s not found", chatID)
}

// ListAllChats returns all chats from all workspaces with workspace info
func (wr *WorkspaceReader) ListAllChats() ([]ChatTabWithWorkspace, error) {
	workspaces, err := wr.FindWorkspaces()
	if err != nil {
		return nil, err
	}

	var allChats []ChatTabWithWorkspace

	for _, workspacePath := range workspaces {
		chatData, err := wr.GetChatData(workspacePath)
		if err != nil {
			continue // Skip errored workspaces
		}

		workspaceName := filepath.Base(filepath.Dir(workspacePath))

		for _, tab := range chatData.Tabs {
			allChats = append(allChats, ChatTabWithWorkspace{
				ChatTab:       tab,
				WorkspacePath: workspacePath,
				WorkspaceName: workspaceName,
			})
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(allChats, func(i, j int) bool {
		return allChats[i].Timestamp > allChats[j].Timestamp
	})

	return allChats, nil
}

// ChatTabWithWorkspace extends ChatTab with workspace information
type ChatTabWithWorkspace struct {
	ChatTab
	WorkspacePath string `json:"workspacePath"`
	WorkspaceName string `json:"workspaceName"`
}

// SearchChats searches for chats containing specific text
func (wr *WorkspaceReader) SearchChats(query string) ([]ChatTabWithWorkspace, error) {
	allChats, err := wr.ListAllChats()
	if err != nil {
		return nil, err
	}

	var matches []ChatTabWithWorkspace

	for _, chat := range allChats {
		// Search in title and content
		if containsIgnoreCase(chat.GetDisplayTitle(), query) ||
			containsIgnoreCase(chat.ToMarkdown(), query) {
			matches = append(matches, chat)
		}
	}

	return matches, nil
}
