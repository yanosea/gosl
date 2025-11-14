// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"fmt"
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

// TestThreadViewModel_LineHeightCacheInitialization tests that line height cache is initialized
func TestThreadViewModel_LineHeightCacheInitialization(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		threadTS  string
		width     int
		height    int
	}{
		{
			name:      "NewThreadViewModel initializes line height cache",
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			width:     80,
			height:    24,
		},
		{
			name:      "NewThreadViewModelWithSender initializes line height cache",
			channelID: "C67890",
			threadTS:  "9876543210.654321",
			width:     100,
			height:    30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model ThreadViewModel

			// Test different constructors
			if strings.Contains(tt.name, "WithSender") {
				model = NewThreadViewModelWithSender(tt.channelID, tt.threadTS, tt.width, tt.height, nil)
			} else {
				model = NewThreadViewModel(tt.channelID, tt.threadTS, tt.width, tt.height)
			}

			// Verify that messageLineHeights is initialized (not nil)
			if model.messageLineHeights == nil {
				t.Error("messageLineHeights should be initialized, got nil")
			}

			// Verify that messageLineHeights is empty on initialization
			if len(model.messageLineHeights) != 0 {
				t.Errorf("messageLineHeights should be empty on initialization, got %d entries", len(model.messageLineHeights))
			}
		})
	}
}

// TestThreadViewModel_SetThread_ClearLineHeightCache tests that SetThread clears the line height cache
func TestThreadViewModel_SetThread_ClearLineHeightCache(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Add some dummy entries to the cache
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["R1"] = 10

	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cache entries before SetThread, got %d", len(model.messageLineHeights))
	}

	// Call SetThread
	parentMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	replies := []message.Message{
		{
			ID:        "R1",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply 1",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}
	model.SetThread(parentMsg, replies)

	// Verify that cache was cleared
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected cache to be cleared after SetThread, got %d entries", len(model.messageLineHeights))
	}
}

// TestThreadViewModel_AddReply_PreserveLineHeightCache tests that AddReply preserves the line height cache
func TestThreadViewModel_AddReply_PreserveLineHeightCache(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Set initial thread
	parentMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	replies := []message.Message{
		{
			ID:        "R1",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply 1",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}
	model.SetThread(parentMsg, replies)

	// Move cursor to parent message (not at latest reply)
	model.selectedIndex = 0
	model.selectedMessageID = parentMsg.ID

	// Add cache entries for existing messages
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["R1"] = 10

	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cache entries before AddReply, got %d", len(model.messageLineHeights))
	}

	// Add a new reply
	newReply := message.Message{
		ID:        "R2",
		ChannelID: "C12345",
		UserID:    "U003",
		UserName:  "Charlie",
		Text:      "Reply 2",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	model.AddReply(newReply)

	// Verify that existing cache was preserved
	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected cache to be preserved after AddReply, got %d entries", len(model.messageLineHeights))
	}

	if model.messageLineHeights["M1"] != 5 {
		t.Errorf("Expected M1 cache entry to be preserved with value 5, got %d", model.messageLineHeights["M1"])
	}

	if model.messageLineHeights["R1"] != 10 {
		t.Errorf("Expected R1 cache entry to be preserved with value 10, got %d", model.messageLineHeights["R1"])
	}
}

// TestThreadViewModel_WindowSizeMsg_ClearCacheOnWidthChange tests cache invalidation on width change
func TestThreadViewModel_WindowSizeMsg_ClearCacheOnWidthChange(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Add cache entries
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["R1"] = 10

	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cache entries before WindowSizeMsg, got %d", len(model.messageLineHeights))
	}

	// Send WindowSizeMsg with different width
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ := model.Update(msg)

	// Verify that cache was cleared due to width change
	if len(updatedModel.messageLineHeights) != 0 {
		t.Errorf("Expected cache to be cleared after width change, got %d entries", len(updatedModel.messageLineHeights))
	}
}

// TestThreadViewModel_WindowSizeMsg_PreserveCacheOnHeightChange tests cache preservation on height-only change
func TestThreadViewModel_WindowSizeMsg_PreserveCacheOnHeightChange(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Add cache entries
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["R1"] = 10

	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cache entries before WindowSizeMsg, got %d", len(model.messageLineHeights))
	}

	// Send WindowSizeMsg with same width but different height
	msg := tea.WindowSizeMsg{Width: 80, Height: 30}
	updatedModel, _ := model.Update(msg)

	// Verify that cache was preserved (height change doesn't affect rendering width)
	if len(updatedModel.messageLineHeights) != 2 {
		t.Errorf("Expected cache to be preserved after height-only change, got %d entries", len(updatedModel.messageLineHeights))
	}

	if updatedModel.messageLineHeights["M1"] != 5 {
		t.Errorf("Expected M1 cache entry to be preserved with value 5, got %d", updatedModel.messageLineHeights["M1"])
	}
}

// TestThreadViewModel_scrollToSelected_EmptyThread tests early return for empty thread
func TestThreadViewModel_scrollToSelected_EmptyThread(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// No parent or replies set - scrollToSelected should handle gracefully
	model.scrollToSelected()

	// Verify viewport offset is still 0 (no panic, no changes)
	if model.viewport.YOffset != 0 {
		t.Errorf("Expected YOffset to remain 0 for empty thread, got %d", model.viewport.YOffset)
	}
}

// TestThreadViewModel_scrollToSelected_ParentSelected tests parent message selection
func TestThreadViewModel_scrollToSelected_ParentSelected(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Set thread with parent and replies
	parentMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	replies := []message.Message{
		{
			ID:        "R1",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply 1",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}
	model.SetThread(parentMsg, replies)

	// Select parent message (selectedIndex == 0)
	model.selectedIndex = 0
	model.selectedMessageID = "M1"

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify offset is at top (parent should be visible at top)
	if model.viewport.YOffset != 0 {
		t.Errorf("Expected YOffset to be 0 for parent selection, got %d", model.viewport.YOffset)
	}
}

// TestThreadViewModel_scrollToSelected_CacheHit tests cache utilization
func TestThreadViewModel_scrollToSelected_CacheHit(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Set thread
	parentMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	replies := []message.Message{
		{
			ID:        "R1",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply 1",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}
	model.SetThread(parentMsg, replies)

	// Manually set cache entries
	model.messageLineHeights["M1"] = 3
	model.messageLineHeights["R1"] = 4

	// Select first reply
	model.selectedIndex = 1
	model.selectedMessageID = "R1"

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify cache was used (no new entries should be added)
	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cached entries after scrollToSelected, got %d", len(model.messageLineHeights))
	}

	// Verify cache values are unchanged
	if model.messageLineHeights["M1"] != 3 {
		t.Errorf("Expected M1 cache entry to remain 3, got %d", model.messageLineHeights["M1"])
	}
	if model.messageLineHeights["R1"] != 4 {
		t.Errorf("Expected R1 cache entry to remain 4, got %d", model.messageLineHeights["R1"])
	}
}

// TestThreadViewModel_scrollToSelected_CacheMiss tests cache population
func TestThreadViewModel_scrollToSelected_CacheMiss(t *testing.T) {
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

	// Set thread
	parentMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Alice",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}
	replies := []message.Message{
		{
			ID:        "R1",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Reply 1",
			Timestamp: time.Now(),
			ThreadTS:  "1234567890.123456",
		},
	}
	model.SetThread(parentMsg, replies)

	// Select first reply
	model.selectedIndex = 1
	model.selectedMessageID = "R1"

	// Verify cache is empty after SetThread
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected empty cache after SetThread, got %d entries", len(model.messageLineHeights))
	}

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify cache was populated
	if len(model.messageLineHeights) == 0 {
		t.Error("Expected cache to be populated after scrollToSelected, but it's empty")
	}

	// Verify that M1 and R1 are cached
	if _, found := model.messageLineHeights["M1"]; !found {
		t.Error("Expected M1 to be cached after scrollToSelected")
	}
	if _, found := model.messageLineHeights["R1"]; !found {
		t.Error("Expected R1 to be cached after scrollToSelected")
	}

	// Verify that cached line heights are positive
	if model.messageLineHeights["M1"] <= 0 {
		t.Errorf("Expected positive line height for M1, got %d", model.messageLineHeights["M1"])
	}
	if model.messageLineHeights["R1"] <= 0 {
		t.Errorf("Expected positive line height for R1, got %d", model.messageLineHeights["R1"])
	}
}

// TestThreadViewModel_Update_PagingOperations tests paging operations call scrollToSelected
func TestThreadViewModel_Update_PagingOperations(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"PageUp", "pgup"},
		{"PageDown", "pgdown"},
		{"Ctrl+U", "ctrl+u"},
		{"Ctrl+D", "ctrl+d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewThreadViewModel("C12345", "1234567890.123456", 80, 24)

			// Set thread
			parentMsg := message.Message{
				ID:        "M1",
				ChannelID: "C12345",
				UserID:    "U001",
				UserName:  "Alice",
				Text:      "Parent message",
				Timestamp: time.Now(),
				ThreadTS:  "1234567890.123456",
			}
			replies := []message.Message{
				{
					ID:        "R1",
					ChannelID: "C12345",
					UserID:    "U002",
					UserName:  "Bob",
					Text:      "Reply 1",
					Timestamp: time.Now(),
					ThreadTS:  "1234567890.123456",
				},
			}
			model.SetThread(parentMsg, replies)

			// Select parent
			model.selectedIndex = 0
			model.selectedMessageID = "M1"

			// Send paging key
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "pgup" {
				msg = tea.KeyMsg{Type: tea.KeyPgUp}
			} else if tt.key == "pgdown" {
				msg = tea.KeyMsg{Type: tea.KeyPgDown}
			}

			updatedModel, _ := model.Update(msg)

			// Verify that the model was updated (scrollToSelected was called)
			// Note: We can't directly verify scrollToSelected was called,
			// but we can ensure no panic occurred and model is valid
			if updatedModel.parentMessage.ID != "M1" {
				t.Errorf("Expected parent message M1 after paging, got %s", updatedModel.parentMessage.ID)
			}
		})
	}
}

// TestThreadViewModel_CursorMovement_Integration tests cursor movement operations with scroll adjustment
func TestThreadViewModel_CursorMovement_Integration(t *testing.T) {
	// Create parent message
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      "Parent line1\nParent line2\nParent line3",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	// Create 10 reply messages with multiple lines each
	replies := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		replies[i] = message.Message{
			ID:        string(rune('R') + rune(i)),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      "Reply line1\nReply line2\nReply line3",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	// Test 1: 'k' key moves cursor up from initial position (starts at last reply)
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	// SetThread sets selectedIndex to len(replies) = 10 on first load (last reply)
	if model.selectedIndex != 10 {
		t.Fatalf("expected initial selectedIndex=10 (last reply), got %d", model.selectedIndex)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if updatedModel.selectedIndex != 9 {
		t.Errorf("expected selectedIndex=9 after 'k', got %d", updatedModel.selectedIndex)
	}

	// Test 2: 'g' key moves to top (parent)
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if updatedModel.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0 (parent) after 'g', got %d", updatedModel.selectedIndex)
	}
	// Verify scroll position is at top when parent is selected
	if updatedModel.viewport.YOffset != 0 {
		t.Errorf("expected YOffset=0 when parent is selected, got %d", updatedModel.viewport.YOffset)
	}

	// Test 3: 'j' key moves cursor from parent to first reply
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	// Move to parent first
	model.selectedIndex = 0
	model.selectedMessageID = parent.ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if updatedModel.selectedIndex != 1 {
		t.Errorf("expected selectedIndex=1 (first reply) after 'j', got %d", updatedModel.selectedIndex)
	}

	// Test 4: Arrow down (↓) moves between replies
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	model.selectedIndex = 2
	model.selectedMessageID = replies[1].ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	if updatedModel.selectedIndex != 3 {
		t.Errorf("expected selectedIndex=3 after ↓, got %d", updatedModel.selectedIndex)
	}

	// Test 5: 'G' key moves to bottom (last reply)
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	model.selectedIndex = 0
	model.selectedMessageID = parent.ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	expectedLastIndex := len(replies) // parent(0) + replies
	if updatedModel.selectedIndex != expectedLastIndex {
		t.Errorf("expected selectedIndex=%d after 'G', got %d", expectedLastIndex, updatedModel.selectedIndex)
	}
}

// TestThreadViewModel_Paging_Integration tests paging operations with cursor visibility
func TestThreadViewModel_Paging_Integration(t *testing.T) {
	// Create parent message with many lines
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      "Parent line1\nParent line2\nParent line3\nParent line4\nParent line5",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	// Create 20 reply messages
	replies := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		replies[i] = message.Message{
			ID:        string(rune('R') + rune(i)),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      "Reply line1\nReply line2\nReply line3",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	// Test 1: PgDown triggers scroll adjustment
	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 15)
	model.SetThread(parent, replies)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})

	// Verify no panic occurred and model is valid
	if updatedModel.parentMessage.ID != "P1" {
		t.Errorf("expected parent message P1 after PgDown, got %s", updatedModel.parentMessage.ID)
	}
	if len(updatedModel.replies) != 20 {
		t.Errorf("expected 20 replies after PgDown, got %d", len(updatedModel.replies))
	}

	// Test 2: Ctrl+D (half page down) triggers scroll adjustment
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 15)
	model.SetThread(parent, replies)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

	// Verify no panic occurred
	if len(updatedModel.replies) != 20 {
		t.Errorf("expected 20 replies after Ctrl+D, got %d", len(updatedModel.replies))
	}

	// Test 3: PgUp triggers scroll adjustment
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 15)
	model.SetThread(parent, replies)
	model.selectedIndex = 10
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgUp})

	// Verify no panic occurred
	if updatedModel.parentMessage.ID != "P1" {
		t.Errorf("expected parent message P1 after PgUp, got %s", updatedModel.parentMessage.ID)
	}

	// Test 4: Ctrl+U (half page up) triggers scroll adjustment
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 15)
	model.SetThread(parent, replies)
	model.selectedIndex = 10
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	// Verify no panic occurred
	if updatedModel.parentMessage.ID != "P1" {
		t.Errorf("expected parent message P1 after Ctrl+U, got %s", updatedModel.parentMessage.ID)
	}
}

// TestThreadViewModel_AddReply_AutoScroll tests that new replies auto-scroll when cursor is at latest
func TestThreadViewModel_AddReply_AutoScroll(t *testing.T) {
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	replies := make([]message.Message, 5)
	for i := 0; i < 5; i++ {
		replies[i] = message.Message{
			ID:        fmt.Sprintf("R%d", i),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      "Reply message",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)

	// Verify cursor is at latest reply (index 5 = parent + 5 replies)
	if model.selectedIndex != 5 {
		t.Fatalf("expected initial selectedIndex=5, got %d", model.selectedIndex)
	}

	// Add new reply
	newReply := message.Message{
		ID:        "R5",
		ChannelID: "C12345",
		UserID:    "U002",
		UserName:  "Reply User",
		Text:      "New reply",
		Timestamp: time.Now().Add(6 * time.Minute),
		ThreadTS:  "1234567890.123456",
	}
	model.AddReply(newReply)

	// Verify cursor moved to new latest reply
	if model.selectedIndex != 6 {
		t.Errorf("expected selectedIndex=6 after adding new reply, got %d", model.selectedIndex)
	}
}

// TestThreadViewModel_AddReply_NoAutoScroll tests that new replies don't auto-scroll when cursor is not at latest
func TestThreadViewModel_AddReply_NoAutoScroll(t *testing.T) {
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      "Parent message",
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	replies := make([]message.Message, 5)
	for i := 0; i < 5; i++ {
		replies[i] = message.Message{
			ID:        fmt.Sprintf("R%d", i),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      "Reply message",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)

	// Move cursor to parent
	model.selectedIndex = 0
	model.selectedMessageID = parent.ID

	initialIndex := model.selectedIndex
	initialMessageID := model.selectedMessageID

	// Add new reply
	newReply := message.Message{
		ID:        "R5",
		ChannelID: "C12345",
		UserID:    "U002",
		UserName:  "Reply User",
		Text:      "New reply",
		Timestamp: time.Now().Add(6 * time.Minute),
		ThreadTS:  "1234567890.123456",
	}
	model.AddReply(newReply)

	// Verify cursor stayed at parent
	if model.selectedIndex != initialIndex {
		t.Errorf("expected selectedIndex=%d (unchanged), got %d", initialIndex, model.selectedIndex)
	}
	if model.selectedMessageID != initialMessageID {
		t.Errorf("expected selectedMessageID=%s (unchanged), got %s", initialMessageID, model.selectedMessageID)
	}
}

// TestThreadViewModel_InputMode_ScrollPosition tests that scroll position is maintained when entering input mode
func TestThreadViewModel_InputMode_ScrollPosition(t *testing.T) {
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      strings.Repeat("Parent line\n", 5),
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	replies := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		replies[i] = message.Message{
			ID:        fmt.Sprintf("R%d", i),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      strings.Repeat("Reply line\n", 3),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)

	// Set specific scroll position
	model.selectedIndex = 10
	model.selectedMessageID = replies[9].ID
	model.viewport.SetYOffset(50)

	initialOffset := model.viewport.YOffset
	initialIndex := model.selectedIndex

	// Enter input mode with 'i' key
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Verify scroll position maintained
	if updatedModel.viewport.YOffset != initialOffset {
		t.Errorf("expected YOffset=%d (unchanged), got %d", initialOffset, updatedModel.viewport.YOffset)
	}
	if updatedModel.selectedIndex != initialIndex {
		t.Errorf("expected selectedIndex=%d (unchanged), got %d", initialIndex, updatedModel.selectedIndex)
	}
	if !updatedModel.inputFocused {
		t.Error("expected inputFocused=true after 'i' key")
	}

	// Test with 'r' key
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	model.selectedIndex = 10
	model.selectedMessageID = replies[9].ID
	model.viewport.SetYOffset(50)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if updatedModel.viewport.YOffset != 50 {
		t.Errorf("expected YOffset=50 (unchanged) after 'r', got %d", updatedModel.viewport.YOffset)
	}
	if !updatedModel.inputFocused {
		t.Error("expected inputFocused=true after 'r' key")
	}
}

// TestThreadViewModel_WindowResize_ScrollAdjustment tests scroll position recalculation on window resize
func TestThreadViewModel_WindowResize_ScrollAdjustment(t *testing.T) {
	parent := message.Message{
		ID:        "P1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Parent User",
		Text:      strings.Repeat("Parent line\n", 5),
		Timestamp: time.Now(),
		ThreadTS:  "1234567890.123456",
	}

	replies := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		replies[i] = message.Message{
			ID:        fmt.Sprintf("R%d", i),
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Reply User",
			Text:      strings.Repeat("Reply line\n", 3),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			ThreadTS:  "1234567890.123456",
		}
	}

	model := NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)

	// Set specific position
	model.selectedIndex = 10
	model.selectedMessageID = replies[9].ID

	// Simulate window resize (width change)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})

	// Verify cache was cleared on width change
	if len(updatedModel.messageLineHeights) > 0 {
		t.Error("expected line height cache to be cleared on width change")
	}

	// Verify cursor position maintained
	if updatedModel.selectedIndex != 10 {
		t.Errorf("expected selectedIndex=10 (unchanged), got %d", updatedModel.selectedIndex)
	}

	// Simulate window resize (height change only)
	model = NewThreadViewModel("C12345", "1234567890.123456", 80, 20)
	model.SetThread(parent, replies)
	model.selectedIndex = 10
	model.selectedMessageID = replies[9].ID
	// Populate some cache
	model.scrollToSelected()
	cacheSize := len(model.messageLineHeights)

	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})

	// Verify cache was preserved on height-only change
	if len(updatedModel.messageLineHeights) != cacheSize {
		t.Errorf("expected line height cache preserved (size=%d), got %d", cacheSize, len(updatedModel.messageLineHeights))
	}
}
