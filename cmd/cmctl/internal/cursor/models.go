package cursor

import (
	"time"
)

// CursorItem represents a key-value item in Cursor's state.vscdb
type CursorItem struct {
	Key   string `gorm:"column:key;primaryKey"`
	Value string `gorm:"column:value"`
}

// TableName specifies the table name for CursorItem
func (CursorItem) TableName() string {
	return "ItemTable"
}

// ChatData represents the complete chat data structure from Cursor
type ChatData struct {
	Tabs []ChatTab `json:"tabs"`
}

// ChatTab represents a single chat conversation in Cursor
type ChatTab struct {
	ID        string    `json:"id"`
	Title     string    `json:"title,omitempty"`
	Messages  []Message `json:"messages"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// Message represents a single message in the chat
type Message struct {
	ID        string    `json:"id,omitempty"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// GetDisplayTitle returns a human-readable title for the chat tab
func (ct *ChatTab) GetDisplayTitle() string {
	if ct.Title != "" {
		return ct.Title
	}

	// Generate title from first user message
	for _, msg := range ct.Messages {
		if msg.Role == "user" && len(msg.Content) > 0 {
			if len(msg.Content) > 50 {
				return msg.Content[:47] + "..."
			}
			return msg.Content
		}
	}

	return "Untitled Chat"
}

// GetContentPreview returns a preview of the chat content
func (ct *ChatTab) GetContentPreview(maxLength int) string {
	content := ""
	for _, msg := range ct.Messages {
		if msg.Role == "user" {
			content += "User: " + msg.Content + "\n"
		} else {
			content += "Assistant: " + msg.Content + "\n"
		}

		if len(content) > maxLength {
			return content[:maxLength-3] + "..."
		}
	}
	return content
}

// ToMarkdown converts the chat tab to markdown format
func (ct *ChatTab) ToMarkdown() string {
	md := "# " + ct.GetDisplayTitle() + "\n\n"

	if ct.CreatedAt.IsZero() && ct.Timestamp > 0 {
		ct.CreatedAt = time.Unix(ct.Timestamp/1000, 0)
	}

	if !ct.CreatedAt.IsZero() {
		md += "**Date**: " + ct.CreatedAt.Format("2006-01-02 15:04:05") + "\n\n"
	}

	for _, msg := range ct.Messages {
		switch msg.Role {
		case "user":
			md += "**User**: " + msg.Content + "\n\n"
		case "assistant":
			md += "**Assistant**: " + msg.Content + "\n\n"
		default:
			md += "**" + msg.Role + "**: " + msg.Content + "\n\n"
		}
	}

	return md
}

// ExtractTechnicalConcepts analyzes chat content for technical terms
func (ct *ChatTab) ExtractTechnicalConcepts() []string {
	technicalTerms := []string{
		"javascript", "typescript", "python", "java", "go", "rust", "cpp", "c++",
		"html", "css", "sql", "bash", "shell", "authentication", "authorization",
		"api", "database", "frontend", "backend", "microservices", "docker",
		"kubernetes", "deployment", "testing", "debugging", "performance",
		"optimization", "security", "encryption", "validation", "refactoring",
		"react", "vue", "angular", "nodejs", "express", "fastapi", "django",
		"flask", "spring", "laravel", "rails", "nextjs", "svelte",
	}

	found := make(map[string]bool)
	var concepts []string

	fullContent := ct.ToMarkdown()
	for _, term := range technicalTerms {
		if !found[term] && containsIgnoreCase(fullContent, term) {
			found[term] = true
			concepts = append(concepts, term)
		}
	}

	return concepts
}

// containsIgnoreCase checks if text contains substring case-insensitively
func containsIgnoreCase(text, substr string) bool {
	// Simple case-insensitive check
	return len(text) >= len(substr) &&
		(text == substr ||
			(len(text) > len(substr) &&
				findIgnoreCase(text, substr) >= 0))
}

// findIgnoreCase finds substring in text case-insensitively
func findIgnoreCase(text, substr string) int {
	textLower := toLower(text)
	substrLower := toLower(substr)

	for i := 0; i <= len(textLower)-len(substrLower); i++ {
		if textLower[i:i+len(substrLower)] == substrLower {
			return i
		}
	}
	return -1
}

// toLower converts string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, r := range []byte(s) {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
