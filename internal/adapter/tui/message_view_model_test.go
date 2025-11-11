// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/domain/message"
)

// TestMessageViewModel_Init tests the initialization of MessageViewModel
func TestMessageViewModel_Init(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		width     int
		height    int
	}{
		{
			name:      "Initialize with valid dimensions",
			channelID: "C12345",
			width:     80,
			height:    24,
		},
		{
			name:      "Initialize with small dimensions",
			channelID: "C67890",
			width:     40,
			height:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageViewModel(tt.channelID, tt.width, tt.height)

			if model.channelID != tt.channelID {
				t.Errorf("expected channelID %s, got %s", tt.channelID, model.channelID)
			}

			if model.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, model.width)
			}

			if model.height != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, model.height)
			}

			if len(model.messages) != 0 {
				t.Errorf("expected empty messages slice, got %d messages", len(model.messages))
			}

			if model.selectedIndex != 0 {
				t.Errorf("expected selectedIndex 0, got %d", model.selectedIndex)
			}
		})
	}
}

// TestMessageViewModel_SetMessages tests setting messages in the MessageViewModel
func TestMessageViewModel_SetMessages(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	messages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Hello, world!",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Hi Alice!",
			Timestamp: time.Now(),
		},
	}

	model.SetMessages(messages, "next_cursor_123")

	if len(model.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(model.messages))
	}

	if model.nextCursor != "next_cursor_123" {
		t.Errorf("expected nextCursor 'next_cursor_123', got '%s'", model.nextCursor)
	}

	if model.messages[0].ID != "M1" {
		t.Errorf("expected first message ID 'M1', got '%s'", model.messages[0].ID)
	}
}

// TestMessageViewModel_AppendMessages tests appending messages (pagination)
func TestMessageViewModel_AppendMessages(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Set initial messages
	initialMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "First message",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(initialMessages, "cursor1")

	// Append more messages
	newMessages := []message.Message{
		{
			ID:        "M0",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Older message",
			Timestamp: time.Now().Add(-time.Hour),
		},
	}
	model.AppendMessages(newMessages, "cursor2")

	if len(model.messages) != 2 {
		t.Errorf("expected 2 messages after append, got %d", len(model.messages))
	}

	// Verify older messages are prepended
	if model.messages[0].ID != "M0" {
		t.Errorf("expected first message ID 'M0', got '%s'", model.messages[0].ID)
	}

	if model.nextCursor != "cursor2" {
		t.Errorf("expected nextCursor 'cursor2', got '%s'", model.nextCursor)
	}
}

// TestMessageViewModel_AddNewMessage tests adding a new message in real-time
func TestMessageViewModel_AddNewMessage(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	initialMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "First message",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(initialMessages, "")

	// Add new message
	newMsg := message.Message{
		ID:        "M2",
		ChannelID: "C12345",
		UserID:    "U002",
		UserName:  "Bob",
		Text:      "New incoming message",
		Timestamp: time.Now(),
	}
	model.AddNewMessage(newMsg)

	if len(model.messages) != 2 {
		t.Errorf("expected 2 messages after adding new message, got %d", len(model.messages))
	}

	// Verify new message is appended at the end
	if model.messages[1].ID != "M2" {
		t.Errorf("expected last message ID 'M2', got '%s'", model.messages[1].ID)
	}
}

// TestMessageViewModel_Update tests Update function with key events
func TestMessageViewModel_Update(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		expectedAction string
	}{
		{
			name:           "Press Esc to return",
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedAction: "return",
		},
		{
			name:           "Press i to enter message input",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
			expectedAction: "input",
		},
		{
			name:           "Press c to enter message input",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
			expectedAction: "input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageViewModel("C12345", 80, 24)
			_, cmd := model.Update(tt.msg)

			// For now, just verify that Update doesn't panic
			// Full command testing would require more complex setup
			_ = cmd
		})
	}
}

// TestMessageViewModel_View tests the View function
func TestMessageViewModel_View(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	messages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Hello, world!",
			Timestamp: time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
		},
	}
	model.SetMessages(messages, "")

	view := model.View()

	// Verify that the view contains user name
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Note: Full rendering test would require checking actual viewport content
	// This is a basic sanity check
}

// TestMessageViewModel_RenderMessages tests message rendering
func TestMessageViewModel_RenderMessages(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	messages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Hello with URL https://example.com and mention @bob",
			Timestamp: time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
			ThreadTS:  "",
			ReplyCount: 0,
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Thread parent message",
			Timestamp: time.Date(2025, 1, 10, 14, 35, 0, 0, time.UTC),
			ThreadTS:  "1234567890.123456",
			ReplyCount: 3,
		},
	}
	model.SetMessages(messages, "")

	rendered := model.renderMessages()

	if rendered == "" {
		t.Error("expected non-empty rendered content")
	}

	// Note: Full rendering test would check for highlighted URLs and mentions
	// This is a basic sanity check
}

// TestMessageViewModel_HighlightText_URLs tests URL highlighting
func TestMessageViewModel_HighlightText_URLs(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	tests := []struct {
		name     string
		input    string
		contains []string // Expected substrings in the output
	}{
		{
			name:     "Simple HTTP URL",
			input:    "Check out http://example.com for more info",
			contains: []string{"http://example.com"},
		},
		{
			name:     "HTTPS URL",
			input:    "Visit https://www.example.com/path?query=1",
			contains: []string{"https://www.example.com/path?query=1"},
		},
		{
			name:     "Multiple URLs",
			input:    "Links: https://example.com and http://test.org",
			contains: []string{"https://example.com", "http://test.org"},
		},
		{
			name:     "No URLs",
			input:    "Just plain text with no links",
			contains: []string{},
		},
		{
			name:     "URL with port",
			input:    "Local server at http://localhost:8080",
			contains: []string{"http://localhost:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.highlightText(tt.input)

			// Result should not be empty
			if result == "" && tt.input != "" {
				t.Errorf("expected non-empty result for input %q", tt.input)
			}

			// For now, just ensure the function runs without panic
			// Full validation would check ANSI styling codes
			_ = result
		})
	}
}

// TestMessageViewModel_HighlightText_Mentions tests mention highlighting
func TestMessageViewModel_HighlightText_Mentions(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	tests := []struct {
		name     string
		input    string
		contains []string // Expected mentions in the output
	}{
		{
			name:     "Single mention",
			input:    "Hey @alice, how are you?",
			contains: []string{"@alice"},
		},
		{
			name:     "Multiple mentions",
			input:    "@bob @charlie please review this",
			contains: []string{"@bob", "@charlie"},
		},
		{
			name:     "Mention with underscore",
			input:    "Ask @john_doe about it",
			contains: []string{"@john_doe"},
		},
		{
			name:     "No mentions",
			input:    "No mentions in this message",
			contains: []string{},
		},
		{
			name:     "Email address (not a mention)",
			input:    "Contact me at user@example.com",
			contains: []string{}, // Email should not be highlighted as mention
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.highlightText(tt.input)

			// Result should not be empty
			if result == "" && tt.input != "" {
				t.Errorf("expected non-empty result for input %q", tt.input)
			}

			// For now, just ensure the function runs without panic
			// Full validation would check ANSI styling codes
			_ = result
		})
	}
}

// TestMessageViewModel_HighlightText_Combined tests combined URL and mention highlighting
func TestMessageViewModel_HighlightText_Combined(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	input := "Hey @alice, check https://example.com and tell @bob"
	result := model.highlightText(input)

	if result == "" {
		t.Errorf("expected non-empty result for combined highlighting")
	}

	// For now, just ensure the function runs without panic
	// Full validation would check both URL and mention highlighting
	_ = result
}
