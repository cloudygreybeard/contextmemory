package cursor

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	Name              string    `json:"name"` // The actual chat title!
	CreatedAt         int64     `json:"createdAt"`
	UnifiedMode       string    `json:"unifiedMode"`
	ForceMode         string    `json:"forceMode"`
	HasUnreadMessages bool      `json:"hasUnreadMessages"`
	Messages          []Message `json:"messages,omitempty"`
}

// AIServiceGeneration represents the richer aiService.generations data structure
type AIServiceGeneration struct {
	UnixMs          int64  `json:"unixMs"`
	GenerationUUID  string `json:"generationUUID"`
	Type            string `json:"type"`
	TextDescription string `json:"textDescription"`
	ConversationID  string `json:"conversationId,omitempty"`
	Role            string `json:"role,omitempty"`
}

// parseAIServicePromptsWithTitles converts aiService.prompts data to ChatTab format with composer titles
func (wr *WorkspaceReader) parseAIServicePromptsWithTitles(value string, composerTitles map[string]string) ([]ChatTab, error) {
	chatTab, err := wr.parseAIServicePromptsToSingleChat(value)
	if err != nil {
		return nil, err
	}

	// Fix nil pointer bug - check if chatTab is nil
	if chatTab == nil {
		return []ChatTab{}, nil
	}

	// Try to match with composer title
	// If there's only one composer title, use it for this chat
	if len(composerTitles) == 1 {
		for _, title := range composerTitles {
			chatTab.Title = title
			break
		}
	} else if len(composerTitles) > 1 {
		// Multiple titles available - for now, use the most recently created one
		// This is a reasonable assumption for the "latest" chat
		for _, title := range composerTitles {
			// We could add more sophisticated matching logic here
			// For now, prefer any non-empty title
			if chatTab.Title == "AI Service Chat" && title != "" {
				chatTab.Title = title
				break
			}
		}
	}

	return []ChatTab{*chatTab}, nil
}

// parseAIServicePromptsToSingleChat is the core logic for parsing aiService.prompts
func (wr *WorkspaceReader) parseAIServicePromptsToSingleChat(value string) (*ChatTab, error) {
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
		// Determine role - prioritize explicit role, then use smarter heuristics
		role := "user" // default
		if prompt.Role != "" {
			role = prompt.Role
		} else {
			// Improved role detection: look at content patterns
			content := prompt.Text
			if len(content) > 0 {
				// Look for assistant-like responses (longer, explanatory)
				if len(content) > 200 || containsAssistantMarkers(content) {
					role = "assistant"
				} else {
					// Shorter content typically from user
					role = "user"
				}
			}
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
	chatTab := &ChatTab{
		ID:        fmt.Sprintf("ai-service-%d", time.Now().Unix()),
		Title:     "AI Service Chat",
		Messages:  messages,
		Timestamp: time.Now().Unix() * 1000,
		CreatedAt: time.Now(),
	}

	return chatTab, nil
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

		// Use the actual chat name if available, otherwise generate from composer info
		title := composer.Name
		if title == "" {
			title = "Composer Chat"
			if composer.UnifiedMode != "" {
				title = fmt.Sprintf("%s Chat", cases.Title(language.English, cases.NoLower).String(composer.UnifiedMode))
			}
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

// parseAIServiceGenerations converts aiService.generations to ChatTab format (richer data source)
func (wr *WorkspaceReader) parseAIServiceGenerations(value string, composerTitles map[string]string) ([]ChatTab, error) {
	var generations []AIServiceGeneration
	if err := json.Unmarshal([]byte(value), &generations); err != nil {
		return nil, fmt.Errorf("failed to parse AI service generations: %w", err)
	}

	if len(generations) == 0 {
		return []ChatTab{}, nil
	}

	// Group generations into conversations and extract full content
	conversationMap := make(map[string][]AIServiceGeneration)
	for _, gen := range generations {
		if gen.Type == "composer" { // Only process AI chat pane interactions
			key := gen.ConversationID
			if key == "" {
				// Use generation UUID as fallback for grouping
				key = gen.GenerationUUID
			}
			conversationMap[key] = append(conversationMap[key], gen)
		}
	}

	var chatTabs []ChatTab
	for _, convGenerations := range conversationMap {
		if len(convGenerations) == 0 {
			continue
		}

		// Sort generations by timestamp
		sort.Slice(convGenerations, func(i, j int) bool {
			return convGenerations[i].UnixMs < convGenerations[j].UnixMs
		})

		// Extract full conversation from textDescription fields
		var messages []Message
		conversationContent := ""

		for i, gen := range convGenerations {
			if gen.TextDescription != "" {
				conversationContent += gen.TextDescription + "\n\n"

				// Create message from generation
				message := Message{
					ID:        gen.GenerationUUID,
					Role:      determineRoleFromContent(gen.TextDescription, i),
					Content:   gen.TextDescription,
					Timestamp: gen.UnixMs,
					CreatedAt: time.Unix(gen.UnixMs/1000, 0),
				}
				messages = append(messages, message)
			}
		}

		if len(messages) == 0 {
			continue
		}

		// Get title from composer titles or generate from content
		title := "AI Service Chat"
		if len(composerTitles) == 1 {
			for _, t := range composerTitles {
				title = t
				break
			}
		}

		// Create chat tab
		chatTab := ChatTab{
			ID:        fmt.Sprintf("generations-%d", convGenerations[0].UnixMs),
			Title:     title,
			Messages:  messages,
			Timestamp: convGenerations[len(convGenerations)-1].UnixMs,
			CreatedAt: time.Unix(convGenerations[0].UnixMs/1000, 0),
		}

		chatTabs = append(chatTabs, chatTab)
	}

	return chatTabs, nil
}

// determineRoleFromContent uses content analysis to determine if content is from user or assistant
func determineRoleFromContent(content string, index int) string {
	// Look for clear indicators of assistant responses
	if containsAssistantMarkers(content) {
		return "assistant"
	}

	// Look for user-like patterns (questions, short requests, file references)
	if containsUserMarkers(content) {
		return "user"
	}

	// Fallback: longer content is typically assistant, shorter is user
	if len(content) > 300 {
		return "assistant"
	}

	return "user"
}

// containsUserMarkers checks if content looks like user input
func containsUserMarkers(content string) bool {
	userMarkers := []string{
		"@", "?", "Can you", "How do I", "What is", "Please", "I want", "I need",
		"Let's", "Could you", "Would you", "Show me", "Help me", "I'm trying",
	}

	for _, marker := range userMarkers {
		if containsIgnoreCase(content, marker) {
			return true
		}
	}
	return false
}

// containsAssistantMarkers checks if content looks like assistant response
func containsAssistantMarkers(content string) bool {
	assistantMarkers := []string{
		"I'll", "I can", "Let me", "Here's", "You can", "This will",
		"```", "Here are", "To do this", "First,", "Next,", "Finally,",
		"## ", "### ", "**", "- [", "1. ", "2. ", "3. ",
	}

	for _, marker := range assistantMarkers {
		if containsIgnoreCase(content, marker) {
			return true
		}
	}
	return false
}
