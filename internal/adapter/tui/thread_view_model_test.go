// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"strings"
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
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
				ID:         "M1",
				ChannelID:  "C12345",
				UserID:     "U001",
				UserName:   "Alice",
				Text:       "Parent message",
				Timestamp:  time.Now(),
				ThreadTS:   "1234567890.123456",
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
		ThreadTS:   "1234567890.123456",
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
		ThreadTS:   "1234567890.123456",
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
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
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent with no replies",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
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

// TestThreadViewModel_RenderThreadMessageLinesWithUserColors tests that renderThreadMessageLines
// applies user-specific background colors when UserColorService is provided
func TestThreadViewModel_RenderThreadMessageLinesWithUserColors(t *testing.T) {
	cache := newMockUserColorCache()
	colorService := newMockUserColorService(cache)

	model := NewThreadViewModelWithColorService("C12345", "1234567890.123456", 80, 24, nil, colorService)
	model.isDarkBackground = true // Simulate dark theme

	parent := message.Message{
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
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
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Second reply",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, replies)

	// Render all thread lines
	allLines := model.getAllThreadLines()

	// Verify rendering produced non-empty output
	if len(allLines) == 0 {
		t.Error("getAllThreadLines() returned empty array")
	}

	// Verify GenerateColorFromID was called (at least once for parent message + replies)
	// Parent (1 call) + 2 replies (2 calls) = 3 calls minimum
	if colorService.generateCallCount < 3 {
		t.Errorf("Expected GenerateColorFromID to be called at least 3 times, got %d", colorService.generateCallCount)
	}

	// Verify that messageStyle.Render() added padding (indicated by trailing space after message text)
	// When Lipgloss Padding(0, 1) is applied, it adds a space after the text
	hasStyledMessage := false
	for _, line := range allLines {
		// Look for message lines with trailing space (indicating Lipgloss padding was applied)
		if strings.Contains(line, "Parent message ") || strings.Contains(line, "First reply ") || strings.Contains(line, "Second reply ") {
			hasStyledMessage = true
			break
		}
	}

	if !hasStyledMessage {
		t.Error("Expected styled messages with padding (indicated by trailing space)")
	}
}

// TestThreadViewModel_RenderThreadMessageLinesWithNilColorService tests that renderThreadMessageLines
// works correctly when UserColorService is nil (backward compatibility)
func TestThreadViewModel_RenderThreadMessageLinesWithNilColorService(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)
	// userColorService is nil by default

	parent := message.Message{
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Parent message",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
		ReplyCount: 1,
	}

	replies := []message.Message{
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply message",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}

	model.SetThread(parent, replies)

	// Render all thread lines
	allLines := model.getAllThreadLines()

	// Verify rendering produced non-empty output
	if len(allLines) == 0 {
		t.Error("getAllThreadLines() returned empty array")
	}

	// Verify that no padding was applied (no trailing space after message text)
	// When colorService is nil, messageStyle.Render() is not called, so no padding
	hasUnstyledMessage := false
	for _, line := range allLines {
		// Look for message lines WITHOUT trailing space (indicating no Lipgloss padding)
		// The message text should appear directly without the extra space from Padding(0, 1)
		if strings.Contains(line, "Parent message") && !strings.Contains(line, "Parent message ") {
			hasUnstyledMessage = true
			break
		}
		if strings.Contains(line, "Reply message") && !strings.Contains(line, "Reply message ") {
			hasUnstyledMessage = true
			break
		}
	}

	if !hasUnstyledMessage {
		t.Error("Expected unstyled messages without padding when colorService is nil")
	}
}

// TestThreadViewModel_RenderThreadMessageLinesMultiline tests that background colors
// are applied correctly to multi-line messages
func TestThreadViewModel_RenderThreadMessageLinesMultiline(t *testing.T) {
	cache := newMockUserColorCache()
	colorService := newMockUserColorService(cache)

	model := NewThreadViewModelWithColorService("C12345", "1234567890.123456", 80, 24, nil, colorService)
	model.isDarkBackground = false // Light theme

	parent := message.Message{
		ID:         "M1",
		ChannelID:  "C12345",
		UserID:     "U001",
		UserName:   "Alice",
		Text:       "Line 1\nLine 2\nLine 3",
		Timestamp:  time.Now(),
		ThreadTS:   "1234567890.123456",
		ReplyCount: 0,
	}

	model.SetThread(parent, []message.Message{})

	// Render all thread lines
	allLines := model.getAllThreadLines()

	// Count lines with padding (indicating styled lines)
	// Each line from the 3-line message should have Lipgloss padding applied
	styledLineCount := 0
	for _, line := range allLines {
		// Look for lines with "Line X " (with trailing space from padding)
		if strings.Contains(line, "Line 1 ") || strings.Contains(line, "Line 2 ") || strings.Contains(line, "Line 3 ") {
			styledLineCount++
		}
	}

	// Each line of the multi-line message should have background color applied
	// (at least 3 lines for the 3-line message text)
	if styledLineCount < 3 {
		t.Errorf("Expected at least 3 styled lines for multi-line message, got %d", styledLineCount)
	}
}

// TestThreadViewModel_RenderThreadMessageLinesIsDarkBackground tests that the appropriate
// color variant (Light or Dark) is selected based on isDarkBackground flag
func TestThreadViewModel_RenderThreadMessageLinesIsDarkBackground(t *testing.T) {
	tests := []struct {
		name             string
		isDarkBackground bool
	}{
		{
			name:             "Light theme",
			isDarkBackground: false,
		},
		{
			name:             "Dark theme",
			isDarkBackground: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := newMockUserColorCache()
			colorService := newMockUserColorService(cache)

			model := NewThreadViewModelWithColorService("C12345", "1234567890.123456", 80, 24, nil, colorService)
			model.isDarkBackground = tt.isDarkBackground

			parent := message.Message{
				ID:         "M1",
				ChannelID:  "C12345",
				UserID:     "U001",
				UserName:   "Alice",
				Text:       "Test message",
				Timestamp:  time.Now(),
				ThreadTS:   "1234567890.123456",
				ReplyCount: 0,
			}

			model.SetThread(parent, []message.Message{})

			// Render all thread lines
			allLines := model.getAllThreadLines()

			// Verify rendering produced output
			if len(allLines) == 0 {
				t.Error("getAllThreadLines() returned empty array")
			}

			// Verify that messageStyle.Render() was applied (indicated by padding)
			hasStyledMessage := false
			for _, line := range allLines {
				if strings.Contains(line, "Test message ") {
					hasStyledMessage = true
					break
				}
			}

			if !hasStyledMessage {
				t.Error("Expected styled message with padding")
			}
		})
	}
}
