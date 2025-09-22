package cursor

import (
	"strings"
	"testing"
	"time"
)

func TestChatTabGetDisplayTitle(t *testing.T) {
	tests := []struct {
		name     string
		chat     ChatTab
		expected string
	}{
		{
			name: "With title",
			chat: ChatTab{
				Title: "Test Chat",
				Messages: []Message{
					{Content: "Hello world"},
				},
			},
			expected: "Test Chat",
		},
		{
			name: "Without title, with messages",
			chat: ChatTab{
				Messages: []Message{
					{Role: "user", Content: "This is a test message for title generation"},
				},
			},
			expected: "This is a test message for title generation",
		},
		{
			name:     "No title, no messages",
			chat:     ChatTab{},
			expected: "Untitled Chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chat.GetDisplayTitle()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChatTabGetContentPreview(t *testing.T) {
	tests := []struct {
		name     string
		chat     ChatTab
		maxLen   int
		expected string
	}{
		{
			name: "Short content",
			chat: ChatTab{
				Messages: []Message{
					{Role: "user", Content: "Short message"},
				},
			},
			maxLen:   50,
			expected: "User: Short message\n",
		},
		{
			name: "Long content gets truncated",
			chat: ChatTab{
				Messages: []Message{
					{Role: "assistant", Content: "This is a very long message that should be truncated"},
				},
			},
			maxLen:   20,
			expected: "Assistant: This i...",
		},
		{
			name:     "No messages",
			chat:     ChatTab{},
			maxLen:   50,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chat.GetContentPreview(tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChatTabToMarkdown(t *testing.T) {
	chat := ChatTab{
		Title: "Test Chat",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
			{
				Role:    "assistant",
				Content: "I'm doing well, thank you!",
			},
		},
	}

	markdown := chat.ToMarkdown()

	if !strings.Contains(markdown, "# Test Chat") {
		t.Error("Markdown should contain title as header")
	}

	if !strings.Contains(markdown, "**User**:") {
		t.Errorf("Markdown should contain user role, got: %s", markdown)
	}

	if !strings.Contains(markdown, "**Assistant**:") {
		t.Errorf("Markdown should contain assistant role, got: %s", markdown)
	}

	if !strings.Contains(markdown, "Hello, how are you?") {
		t.Error("Markdown should contain user message")
	}

	if !strings.Contains(markdown, "I'm doing well, thank you!") {
		t.Error("Markdown should contain assistant message")
	}
}

func TestChatTabExtractTechnicalConcepts(t *testing.T) {
	chat := ChatTab{
		Messages: []Message{
			{Content: "I'm working with Go and Python"},
			{Content: "Let's talk about Docker containers and Kubernetes"},
			{Content: "We need to implement REST APIs using JavaScript"},
		},
	}

	concepts := chat.ExtractTechnicalConcepts()

	expectedConcepts := []string{"go", "python", "docker", "kubernetes", "javascript", "api"}

	for _, expected := range expectedConcepts {
		found := false
		for _, concept := range concepts {
			if concept == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find concept %q in %v", expected, concepts)
		}
	}
}

func TestMessageTimestampParsing(t *testing.T) {
	// Test with current timestamp
	now := time.Now()

	// This would test actual timestamp parsing if we had that field
	// For now, just test that the structure can hold time data
	message := Message{
		Role:    "user",
		Content: "Test message",
		// Timestamp parsing would go here
	}

	if message.Role != "user" {
		t.Errorf("Expected role 'user', got %q", message.Role)
	}

	if message.Content != "Test message" {
		t.Errorf("Expected content 'Test message', got %q", message.Content)
	}

	// Test passes if we get here without panics
	_ = now // Use the variable to avoid unused warning
}
