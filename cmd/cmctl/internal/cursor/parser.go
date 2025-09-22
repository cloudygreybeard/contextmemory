package cursor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AIServicePrompt represents the structure of aiService.prompts data
type AIServicePrompt struct {
	Text      string    `json:"text"`
	Timestamp int64     `json:"timestamp,omitempty"`
	ID        string    `json:"id,omitempty"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// ComposerData represents the structure of composer.composerData
type ComposerData struct {
	AllComposers []ComposerEntry `json:"allComposers"`
}

type ComposerEntry struct {
	Type              string    `json:"type"`
	ComposerID        string    `json:"composerId"`
	CreatedAt         int64     `json:"createdAt"`
	UnifiedMode       string    `json:"unifiedMode"`
	ForceMode         string    `json:"forceMode"`
	HasUnreadMessages bool      `json:"hasUnreadMessages"`
	Messages          []Message `json:"messages,omitempty"`
}

// parseAIServicePrompts converts aiService.prompts data to ChatTab format
func (wr *WorkspaceReader) parseAIServicePrompts(value string) ([]ChatTab, error) {
	// Try to parse as array of prompts
	var prompts []AIServicePrompt
	if err := json.Unmarshal([]byte(value), &prompts); err != nil {
		// Try to parse as single prompt
		var singlePrompt AIServicePrompt
		if err := json.Unmarshal([]byte(value), &singlePrompt); err != nil {
			return nil, fmt.Errorf("failed to parse AI service prompts: %w", err)
		}
		prompts = []AIServicePrompt{singlePrompt}
	}

	if len(prompts) == 0 {
		return nil, nil
	}

	// Convert to ChatTab format
	var messages []Message
	for i, prompt := range prompts {
		// Determine role - alternate between user and assistant
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		if prompt.Role != "" {
			role = prompt.Role
		}

		timestamp := prompt.Timestamp
		if timestamp == 0 && !prompt.CreatedAt.IsZero() {
			timestamp = prompt.CreatedAt.Unix() * 1000
		}
		if timestamp == 0 {
			timestamp = time.Now().Unix() * 1000
		}

		message := Message{
			ID:        fmt.Sprintf("prompt-%d", i),
			Role:      role,
			Content:   prompt.Text,
			Timestamp: timestamp,
		}
		messages = append(messages, message)
	}

	// Create a single chat tab from all prompts
	chatTab := ChatTab{
		ID:        fmt.Sprintf("ai-service-%d", time.Now().Unix()),
		Title:     "AI Service Chat",
		Messages:  messages,
		Timestamp: time.Now().Unix() * 1000,
		CreatedAt: time.Now(),
	}

	return []ChatTab{chatTab}, nil
}

// parseComposerData converts composer.composerData to ChatTab format
func (wr *WorkspaceReader) parseComposerData(value string) ([]ChatTab, error) {
	var composerData ComposerData
	if err := json.Unmarshal([]byte(value), &composerData); err != nil {
		return nil, fmt.Errorf("failed to parse composer data: %w", err)
	}

	var chatTabs []ChatTab
	for _, composer := range composerData.AllComposers {
		if composer.Type != "head" {
			continue // Skip non-head composers for now
		}

		// Generate title from composer info
		title := "Composer Chat"
		if composer.UnifiedMode != "" {
			title = fmt.Sprintf("%s Chat", strings.Title(composer.UnifiedMode))
		}

		chatTab := ChatTab{
			ID:        composer.ComposerID,
			Title:     title,
			Messages:  composer.Messages, // May be empty, that's ok
			Timestamp: composer.CreatedAt,
			CreatedAt: time.Unix(composer.CreatedAt/1000, 0),
		}

		// If no messages but we have composer data, create a placeholder
		if len(chatTab.Messages) == 0 {
			chatTab.Messages = []Message{
				{
					ID:        "composer-info",
					Role:      "system",
					Content:   fmt.Sprintf("Composer session: %s mode, created at %s", composer.UnifiedMode, chatTab.CreatedAt.Format("2006-01-02 15:04:05")),
					Timestamp: composer.CreatedAt,
				},
			}
		}

		chatTabs = append(chatTabs, chatTab)
	}

	return chatTabs, nil
}
