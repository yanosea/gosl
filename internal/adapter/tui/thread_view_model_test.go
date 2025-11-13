// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/domain/message"
)

// TestThreadViewModel_Init tests the initialization of ThreadViewModel
func TestThreadViewModel_Init(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		threadTS  string
		width     int
		height    int
	}{
		{
			name:      "Initialize with valid dimensions",
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			width:     80,
			height:    24,
		},
		{
			name:      "Initialize with small dimensions",
			channelID: "C67890",
			threadTS:  "9876543210.654321",
			width:     40,
			height:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewThreadViewModel(tt.channelID, tt.threadTS, tt.width, tt.height)

			if model.channelID != tt.channelID {
				t.Errorf("expected channelID %s, got %s", tt.channelID, model.channelID)
			}

			if model.threadTS != tt.threadTS {
				t.Errorf("expected threadTS %s, got %s", tt.threadTS, model.threadTS)
			}

			if model.width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, model.width)
			}

			if model.height != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, model.height)
			}

			if len(model.replies) != 0 {
				t.Errorf("expected empty replies slice, got %d replies", len(model.replies))
			}

			if model.selectedIndex != 0 {
				t.Errorf("expected selectedIndex 0, got %d", model.selectedIndex)
			}
		})
	}
}

// TestThreadViewModel_SetThread tests setting thread data
func TestThreadViewModel_SetThread(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 2,
	}

	replies := []message.Message{
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "First reply",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U003",
			UserName:  "Charlie",
			Text:      "Second reply",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, replies)

	if model.parentMessage.ID != "M1" {
		t.Errorf("expected parent message ID 'M1', got '%s'", model.parentMessage.ID)
	}

	if len(model.replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(model.replies))
	}

	if model.replies[0].ID != "M2" {
		t.Errorf("expected first reply ID 'M2', got '%s'", model.replies[0].ID)
	}

	if model.replies[1].ID != "M3" {
		t.Errorf("expected second reply ID 'M3', got '%s'", model.replies[1].ID)
	}
}

// TestThreadViewModel_AddReply tests adding a new reply to the thread
func TestThreadViewModel_AddReply(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 1,
	}

	initialReplies := []message.Message{
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "First reply",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, initialReplies)

	// Add new reply
	newReply := message.Message{
		ID:        "M3",
		ChannelID: "C12345",
		UserID:    "U003",
		UserName:  "Charlie",
		Text:      "New reply",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	model.AddReply(newReply)

	if len(model.replies) != 2 {
		t.Errorf("expected 2 replies after adding, got %d", len(model.replies))
	}

	if model.replies[1].ID != "M3" {
		t.Errorf("expected last reply ID 'M3', got '%s'", model.replies[1].ID)
	}
}

// TestThreadViewModel_Update tests Update function with key events
func TestThreadViewModel_Update(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		expectedAction string
	}{
		{
			name:           "Press Esc to return to MessageView",
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedAction: "return",
		},
		{
			name:           "Press i to enter reply input",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
			expectedAction: "input",
		},
		{
			name:           "Press r to enter reply input",
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			expectedAction: "input",
		},
		{
			name:           "Press up to navigate",
			msg:            tea.KeyMsg{Type: tea.KeyUp},
			expectedAction: "navigate",
		},
		{
			name:           "Press down to navigate",
			msg:            tea.KeyMsg{Type: tea.KeyDown},
			expectedAction: "navigate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

			// Set up thread with parent and replies
			parent := message.Message{
				ID:        "M1",
				ChannelID: "C12345",
				UserID:    "U001",
				UserName:  "Alice",
				Text:      "Parent message",
				Timestamp: time.Now(),
				ThreadTS:  "1234567890.123456",
				ReplyCount: 1,
			}
			replies := []message.Message{
				{
					ID:        "M2",
					ChannelID: "C12345",
					UserID:    "U002",
					UserName:  "Bob",
					Text:      "Reply",
					Timestamp: time.Now(),
					ThreadTS:  "1234567890.123456",
				},
			}
			model.SetThread(parent, replies)

			_, cmd := model.Update(tt.msg)

			// For now, just verify that Update doesn't panic
			// Full command testing would require more complex setup
			_ = cmd
		})
	}
}

// TestThreadViewModel_View tests the View function
func TestThreadViewModel_View(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 1,
	}

	replies := []message.Message{
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply message",
			Timestamp: time.Date(2025, 1, 10, 14, 35, 0, 0, time.UTC),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, replies)

	view := model.View()

	// Verify that the view contains content
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Note: Full rendering test would require checking actual viewport content
	// This is a basic sanity check
}

// TestThreadViewModel_RenderThread tests thread rendering with indentation
func TestThreadViewModel_RenderThread(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 2,
	}

	replies := []message.Message{
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "First reply",
			Timestamp: time.Date(2025, 1, 10, 14, 35, 0, 0, time.UTC),
			ThreadTS:  "1234567890.123456",
		},
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U003",
			UserName:  "Charlie",
			Text:      "Second reply",
			Timestamp: time.Date(2025, 1, 10, 14, 40, 0, 0, time.UTC),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, replies)

	allLines := model.getAllThreadLines()

	if len(allLines) == 0 {
		t.Error("expected non-empty rendered content")
	}

	// Verify that the rendered content includes parent and replies
	// (Full verification would check for proper indentation and styling)
}

// TestThreadViewModel_Navigation tests navigation between messages
func TestThreadViewModel_Navigation(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 2,
	}

	replies := []message.Message{
		{ID: "M2", UserName: "Bob", Text: "Reply 1", ThreadTS: "1234567890.123456"},
		{ID: "M3", UserName: "Charlie", Text: "Reply 2", ThreadTS: "1234567890.123456"},
	}

	model.SetThread(parent, replies)

	// Initial position - should be at the last reply (index 2) on first load
	if model.selectedIndex != 2 {
		t.Errorf("expected initial selectedIndex 2 (last reply), got %d", model.selectedIndex)
	}

	// Move down at the end (should stay at 2)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.selectedIndex != 2 {
		t.Errorf("expected selectedIndex 2 (no change), got %d", model.selectedIndex)
	}

	// Move up
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1 after up, got %d", model.selectedIndex)
	}

	// Move up again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0 after second up, got %d", model.selectedIndex)
	}

	// Move up at the beginning (should stay at 0)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if model.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0 (no change), got %d", model.selectedIndex)
	}

	// Move down
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if model.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1 after down, got %d", model.selectedIndex)
	}
}

// TestThreadViewModel_EmptyThread tests behavior with no replies
func TestThreadViewModel_EmptyThread(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	parent := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent with no replies",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
		ReplyCount: 0,
	}

	model.SetThread(parent, []message.Message{})

	allLines := model.getAllThreadLines()

	if len(allLines) == 0 {
		t.Error("expected non-empty rendered content even with no replies")
	}

	// View should still render
	view := model.View()
	if view == "" {
		t.Error("expected non-empty view even with no replies")
	}
}
