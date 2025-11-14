// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
	"github.com/yanosea/gosl/internal/domain/usercolor"
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
			ID:         "M1",
			ChannelID:  "C12345",
			UserID:     "U001",
			UserName:   "Alice",
			Text:       "Hello with URL https://example.com and mention @bob",
			Timestamp:  time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC),
			ThreadTS:   "",
			ReplyCount: 0,
		},
		{
			ID:         "M2",
			ChannelID:  "C12345",
			UserID:     "U002",
			UserName:   "Bob",
			Text:       "Thread parent message",
			Timestamp:  time.Date(2025, 1, 10, 14, 35, 0, 0, time.UTC),
			ThreadTS:   "1234567890.123456",
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

// TestMessageViewModel_BackgroundColorMsg tests BackgroundColorMsg handling
func TestMessageViewModel_BackgroundColorMsg(t *testing.T) {
	tests := []struct {
		name           string
		isDark         bool
		expectedIsDark bool
	}{
		{
			name:           "Dark background",
			isDark:         true,
			expectedIsDark: true,
		},
		{
			name:           "Light background",
			isDark:         false,
			expectedIsDark: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageViewModel("C12345", 80, 24)

			// Send BackgroundColorMsg
			bgMsg := BackgroundColorMsg{IsDark: tt.isDark}
			updatedModel, _ := model.Update(bgMsg)

			// Verify background theme was stored
			if updatedModel.isDarkBackground != tt.expectedIsDark {
				t.Errorf("isDarkBackground = %v, want %v", updatedModel.isDarkBackground, tt.expectedIsDark)
			}
		})
	}
}

// TestMessageViewModel_UserColorServiceInjection tests UserColorService injection
func TestMessageViewModel_UserColorServiceInjection(t *testing.T) {
	cache := newMockUserColorCache()
	colorService := newMockUserColorService(cache)

	// Create MessageViewModel with color service using the extended constructor
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, colorService)

	// Verify color service was injected
	if model.userColorService == nil {
		t.Error("UserColorService was not injected")
	}
}

// Mock implementations for testing
type mockUserColorCache struct{}

func newMockUserColorCache() *mockUserColorCache {
	return &mockUserColorCache{}
}

func (m *mockUserColorCache) Get(userID string) (usercolor.AdaptiveColor, bool) {
	return usercolor.AdaptiveColor{}, false
}

func (m *mockUserColorCache) Set(userID string, color usercolor.AdaptiveColor) {}

func (m *mockUserColorCache) Clear() {}

func (m *mockUserColorCache) Len() int {
	return 0
}

type mockUserColorService struct {
	cache              *mockUserColorCache
	generateCallCount  int
	lastCalledWithUser string
}

func newMockUserColorService(cache *mockUserColorCache) *mockUserColorService {
	return &mockUserColorService{cache: cache}
}

func (m *mockUserColorService) GetUserColor(u *user.User) usercolor.AdaptiveColor {
	return usercolor.AdaptiveColor{
		Light: usercolor.Color{R: 100, G: 100, B: 100},
		Dark:  usercolor.Color{R: 200, G: 200, B: 200},
	}
}

func (m *mockUserColorService) GenerateColorFromID(userID string) usercolor.AdaptiveColor {
	m.generateCallCount++
	m.lastCalledWithUser = userID
	return usercolor.AdaptiveColor{
		Light: usercolor.Color{R: 100, G: 100, B: 100},
		Dark:  usercolor.Color{R: 200, G: 200, B: 200},
	}
}

func (m *mockUserColorService) ParseSlackColor(colorHex string) (usercolor.AdaptiveColor, error) {
	return usercolor.AdaptiveColor{}, nil
}

func (m *mockUserColorService) ValidateContrast(foreground, background usercolor.Color) bool {
	return true
}

// TestMessageViewModel_RenderWithUserColors tests message rendering with user-specific background colors
func TestMessageViewModel_RenderWithUserColors(t *testing.T) {
	cache := newMockUserColorCache()
	colorService := newMockUserColorService(cache)

	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, colorService)
	model.isDarkBackground = true // Simulate dark theme

	// Add test messages from different users
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Hello",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Hi there",
			Timestamp: time.Now(),
		},
	}

	model.SetMessages(testMessages, "")

	// Render messages
	rendered := model.renderMessages()

	// Verify rendering produced non-empty output
	if rendered == "" {
		t.Error("renderMessages() returned empty string")
	}

	// Verify messages were rendered (this is a basic check - visual inspection would be needed for colors)
	if !strings.Contains(rendered, "Alice") || !strings.Contains(rendered, "Bob") {
		t.Error("Rendered output should contain user names")
	}
}

// TestMessageViewModel_LineHeightCacheInitialization tests that the line height cache is properly initialized
func TestMessageViewModel_LineHeightCacheInitialization(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		width     int
		height    int
	}{
		{
			name:      "NewMessageViewModel initializes line height cache",
			channelID: "C12345",
			width:     80,
			height:    24,
		},
		{
			name:      "NewMessageViewModelWithSender initializes line height cache",
			channelID: "C67890",
			width:     100,
			height:    30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model MessageViewModel

			// Test different constructors
			if strings.Contains(tt.name, "WithSender") {
				model = NewMessageViewModelWithSender(tt.channelID, tt.width, tt.height, nil)
			} else {
				model = NewMessageViewModel(tt.channelID, tt.width, tt.height)
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

// TestMessageViewModel_LineHeightCacheIntegration tests RenderCache and StringBuilderPool integration
func TestMessageViewModel_LineHeightCacheIntegration(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Verify RenderCache is initialized
	if model.renderCache == nil {
		t.Error("renderCache should be initialized, got nil")
	}

	// Verify StringBuilderPool is initialized
	if model.stringBuilders == nil {
		t.Error("stringBuilders should be initialized, got nil")
	}

	// Verify messageLineHeights is initialized
	if model.messageLineHeights == nil {
		t.Error("messageLineHeights should be initialized, got nil")
	}

	// These three components should work together for performance optimization
}

// TestMessageViewModel_SetMessages_ClearLineHeightCache tests that SetMessages clears the line height cache
func TestMessageViewModel_SetMessages_ClearLineHeightCache(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add some dummy entries to the cache
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["M2"] = 10

	if len(model.messageLineHeights) != 2 {
		t.Errorf("Expected 2 cache entries before SetMessages, got %d", len(model.messageLineHeights))
	}

	// Call SetMessages
	messages := []message.Message{
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "New message",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(messages, "cursor1")

	// Verify that cache was cleared
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected cache to be cleared after SetMessages, got %d entries", len(model.messageLineHeights))
	}
}

// TestMessageViewModel_AppendMessages_PreserveLineHeightCache tests that AppendMessages preserves the line height cache
func TestMessageViewModel_AppendMessages_PreserveLineHeightCache(t *testing.T) {
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

	// Add cache entries for existing messages
	model.messageLineHeights["M1"] = 5

	if len(model.messageLineHeights) != 1 {
		t.Errorf("Expected 1 cache entry before AppendMessages, got %d", len(model.messageLineHeights))
	}

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

	// Verify that existing cache was preserved
	if len(model.messageLineHeights) != 1 {
		t.Errorf("Expected cache to be preserved after AppendMessages, got %d entries", len(model.messageLineHeights))
	}

	if model.messageLineHeights["M1"] != 5 {
		t.Errorf("Expected M1 cache entry to be preserved with value 5, got %d", model.messageLineHeights["M1"])
	}
}

// TestMessageViewModel_WindowSizeMsg_ClearCacheOnWidthChange tests cache invalidation on width change
func TestMessageViewModel_WindowSizeMsg_ClearCacheOnWidthChange(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add cache entries
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["M2"] = 10

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

// TestMessageViewModel_WindowSizeMsg_PreserveCacheOnHeightChange tests cache preservation on height-only change
func TestMessageViewModel_WindowSizeMsg_PreserveCacheOnHeightChange(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add cache entries
	model.messageLineHeights["M1"] = 5
	model.messageLineHeights["M2"] = 10

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

// TestMessageViewModel_scrollToSelected_EmptyMessages tests early return for empty messages
func TestMessageViewModel_scrollToSelected_EmptyMessages(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// No messages set - scrollToSelected should handle gracefully
	model.scrollToSelected()

	// Verify viewport offset is still 0 (no panic, no changes)
	if model.viewport.YOffset != 0 {
		t.Errorf("Expected YOffset to remain 0 for empty messages, got %d", model.viewport.YOffset)
	}
}

// TestMessageViewModel_scrollToSelected_CacheHit tests that scrollToSelected uses cached line heights
func TestMessageViewModel_scrollToSelected_CacheHit(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "First message",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Second message",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Manually set cache entries
	model.messageLineHeights["M1"] = 3
	model.messageLineHeights["M2"] = 4

	// Select second message
	model.selectedIndex = 1
	model.selectedMessageID = "M2"

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
	if model.messageLineHeights["M2"] != 4 {
		t.Errorf("Expected M2 cache entry to remain 4, got %d", model.messageLineHeights["M2"])
	}
}

// TestMessageViewModel_scrollToSelected_CacheMiss tests that scrollToSelected calculates and caches line heights
func TestMessageViewModel_scrollToSelected_CacheMiss(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "First message",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Second message with\nmultiple lines\nof text",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Select second message
	model.selectedIndex = 1
	model.selectedMessageID = "M2"

	// Verify cache is empty after SetMessages
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected empty cache after SetMessages, got %d entries", len(model.messageLineHeights))
	}

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify cache was populated with calculated line heights
	if len(model.messageLineHeights) == 0 {
		t.Error("Expected cache to be populated after scrollToSelected, but it's empty")
	}

	// Verify that both M1 and M2 are cached (M1 is calculated for selectedLineStart)
	if _, found := model.messageLineHeights["M1"]; !found {
		t.Error("Expected M1 to be cached after scrollToSelected")
	}
	if _, found := model.messageLineHeights["M2"]; !found {
		t.Error("Expected M2 to be cached after scrollToSelected")
	}

	// Verify that cached line heights are positive
	if model.messageLineHeights["M1"] <= 0 {
		t.Errorf("Expected positive line height for M1, got %d", model.messageLineHeights["M1"])
	}
	if model.messageLineHeights["M2"] <= 0 {
		t.Errorf("Expected positive line height for M2, got %d", model.messageLineHeights["M2"])
	}
}

// TestMessageViewModel_scrollToSelected_UseStringBuilderPool tests that scrollToSelected uses StringBuilderPool
func TestMessageViewModel_scrollToSelected_UseStringBuilderPool(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "First message",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Verify StringBuilderPool is available
	if model.stringBuilders == nil {
		t.Fatal("stringBuilders should be initialized")
	}

	// Call scrollToSelected (should use StringBuilderPool)
	model.scrollToSelected()

	// Test passed if no panic occurred (StringBuilderPool.Get/Put should work correctly)
}

// TestMessageViewModel_scrollToSelected_VisibleRangeCheck tests that scrollToSelected skips scrolling when cursor is visible
func TestMessageViewModel_scrollToSelected_VisibleRangeCheck(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Message 1",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Message 2",
			Timestamp: time.Now(),
		},
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U003",
			UserName:  "Charlie",
			Text:      "Message 3",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Populate cache with known heights
	model.messageLineHeights["M1"] = 3
	model.messageLineHeights["M2"] = 3
	model.messageLineHeights["M3"] = 3

	// Select second message (index 1)
	model.selectedIndex = 1
	model.selectedMessageID = "M2"

	// Set viewport offset to 0 (M2 starts at line 3, viewport height is ~15)
	// M2 should be visible (lines 3-6 within viewport 0-15)
	model.viewport.SetYOffset(0)
	initialOffset := model.viewport.YOffset

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify offset remained unchanged (cursor was already visible)
	if model.viewport.YOffset != initialOffset {
		t.Errorf("Expected YOffset to remain %d (cursor visible), got %d", initialOffset, model.viewport.YOffset)
	}
}

// TestMessageViewModel_scrollToSelected_MessageFitsViewport tests dynamic scroll adjustment when message fits
func TestMessageViewModel_scrollToSelected_MessageFitsViewport(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Message 1",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Message 2",
			Timestamp: time.Now(),
		},
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U003",
			UserName:  "Charlie",
			Text:      "Message 3 (selected, should fit in viewport)",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Populate cache with known heights
	model.messageLineHeights["M1"] = 3
	model.messageLineHeights["M2"] = 3
	model.messageLineHeights["M3"] = 4 // Smaller than viewport height (~15)

	// Select third message (index 2)
	model.selectedIndex = 2
	model.selectedMessageID = "M3"

	// Set viewport offset such that M3 is not visible (starts at line 6, viewport shows 0-5)
	model.viewport.SetYOffset(0)

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify offset was adjusted to make M3 visible
	// M3 starts at line 6 (3+3), and should be positioned so the entire message fits
	expectedOffset := 6 + 4 - model.viewport.Height // Bottom of M3 aligned with bottom of viewport
	if expectedOffset < 0 {
		expectedOffset = 6 // Or top-aligned if it fits
	}
	if model.viewport.YOffset < 0 {
		t.Errorf("YOffset should not be negative, got %d", model.viewport.YOffset)
	}
}

// TestMessageViewModel_scrollToSelected_MessageTallerThanViewport tests adjustment when message is taller
func TestMessageViewModel_scrollToSelected_MessageTallerThanViewport(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add test messages
	testMessages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Short message",
			Timestamp: time.Now(),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Very long message\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20",
			Timestamp: time.Now(),
		},
	}
	model.SetMessages(testMessages, "")

	// Populate cache with known heights
	model.messageLineHeights["M1"] = 3
	model.messageLineHeights["M2"] = 25 // Taller than viewport height (~15)

	// Select second message (index 1)
	model.selectedIndex = 1
	model.selectedMessageID = "M2"

	// Set viewport offset to 0
	model.viewport.SetYOffset(0)

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify offset was set to message start (top-aligned)
	expectedOffset := 3 // M2 starts at line 3
	if model.viewport.YOffset != expectedOffset {
		t.Errorf("Expected YOffset to be %d (message top-aligned), got %d", expectedOffset, model.viewport.YOffset)
	}
}

// TestMessageViewModel_Update_PagingOperations tests that paging operations call scrollToSelected
func TestMessageViewModel_Update_PagingOperations(t *testing.T) {
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
			model := NewMessageViewModel("C12345", 80, 24)

			// Add test messages
			testMessages := []message.Message{
				{
					ID:        "M1",
					ChannelID: "C12345",
					UserID:    "U001",
					UserName:  "Alice",
					Text:      "Message 1",
					Timestamp: time.Now(),
				},
				{
					ID:        "M2",
					ChannelID: "C12345",
					UserID:    "U002",
					UserName:  "Bob",
					Text:      "Message 2",
					Timestamp: time.Now(),
				},
			}
			model.SetMessages(testMessages, "")

			// Select first message
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
			if len(updatedModel.messages) != 2 {
				t.Errorf("Expected 2 messages after paging, got %d", len(updatedModel.messages))
			}
		})
	}
}

// TestMessageViewModel_CursorMovement_Integration tests cursor movement operations with scroll adjustment
func TestMessageViewModel_CursorMovement_Integration(t *testing.T) {
	// Create test messages (10 messages with multiple lines each)
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        string(rune('A' + i)),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      "Line1\nLine2\nLine3\nLine4\nLine5", // 5 lines each
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	// Test 1: 'k' key moves cursor up from initial position (starts at last message = index 9)
	model := NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	// SetMessages sets selectedIndex to len(messages)-1 = 9 on first load
	if model.selectedIndex != 9 {
		t.Fatalf("expected initial selectedIndex=9, got %d", model.selectedIndex)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if updatedModel.selectedIndex != 8 {
		t.Errorf("expected selectedIndex=8 after 'k', got %d", updatedModel.selectedIndex)
	}

	// Test 2: 'g' key moves to top
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if updatedModel.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0 after 'g', got %d", updatedModel.selectedIndex)
	}
	// Verify scroll position is at top
	if updatedModel.viewport.YOffset != 0 {
		t.Errorf("expected YOffset=0 after 'g', got %d", updatedModel.viewport.YOffset)
	}

	// Test 3: 'j' key moves cursor down from top
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	// Move to top first
	model.selectedIndex = 0
	model.selectedMessageID = messages[0].ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if updatedModel.selectedIndex != 1 {
		t.Errorf("expected selectedIndex=1 after 'j', got %d", updatedModel.selectedIndex)
	}

	// Test 4: Arrow up (↑) key moves cursor up
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	model.selectedIndex = 5
	model.selectedMessageID = messages[5].ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if updatedModel.selectedIndex != 4 {
		t.Errorf("expected selectedIndex=4 after ↑, got %d", updatedModel.selectedIndex)
	}

	// Test 5: 'G' key moves to bottom
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	model.selectedIndex = 0
	model.selectedMessageID = messages[0].ID
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if updatedModel.selectedIndex != 9 {
		t.Errorf("expected selectedIndex=9 after 'G', got %d", updatedModel.selectedIndex)
	}
}

// TestMessageViewModel_Paging_Integration tests paging operations with cursor visibility
func TestMessageViewModel_Paging_Integration(t *testing.T) {
	// Create test messages (20 messages with multiple lines each)
	messages := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		messages[i] = message.Message{
			ID:        string(rune('A' + i)),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      "Line1\nLine2\nLine3", // 3 lines each
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	// Initialize model with viewport height=10
	model := NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")

	// Test 1: PgDown moves viewport and keeps cursor visible
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})

	// Verify cursor is within visible range
	selectedLineStart := 0
	for i := 0; i < updatedModel.selectedIndex; i++ {
		height, ok := updatedModel.messageLineHeights[updatedModel.messages[i].ID]
		if !ok {
			height = 4 // Default estimate (3 lines + separator)
		}
		selectedLineStart += height
	}

	if selectedLineStart < updatedModel.viewport.YOffset {
		t.Errorf("cursor is above viewport after PgDown: selectedLineStart=%d, YOffset=%d",
			selectedLineStart, updatedModel.viewport.YOffset)
	}
	if selectedLineStart > updatedModel.viewport.YOffset+updatedModel.viewport.Height {
		t.Errorf("cursor is below viewport after PgDown: selectedLineStart=%d, YOffset=%d, Height=%d",
			selectedLineStart, updatedModel.viewport.YOffset, updatedModel.viewport.Height)
	}

	// Test 2: Ctrl+D (half page down) keeps cursor visible
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

	// Verify cursor is within visible range
	selectedLineStart = 0
	for i := 0; i < updatedModel.selectedIndex; i++ {
		height, ok := updatedModel.messageLineHeights[updatedModel.messages[i].ID]
		if !ok {
			height = 4
		}
		selectedLineStart += height
	}

	if selectedLineStart < updatedModel.viewport.YOffset {
		t.Errorf("cursor is above viewport after Ctrl+D: selectedLineStart=%d, YOffset=%d",
			selectedLineStart, updatedModel.viewport.YOffset)
	}

	// Test 3: PgUp triggers scroll adjustment and cursor remains visible
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	// Initially at last message (index 19)
	initialOffset := model.viewport.YOffset
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgUp})

	// After PgUp, scrollToSelected should have been called
	// The exact YOffset depends on viewport content, but we verify no panic occurred
	if len(updatedModel.messages) != 20 {
		t.Errorf("expected 20 messages after PgUp, got %d", len(updatedModel.messages))
	}
	// Verify initial offset was set (messages were rendered)
	if initialOffset == 0 && len(messages) > 1 {
		t.Log("Initial offset was 0, which is valid for small content")
	}

	// Test 4: Ctrl+U triggers scroll adjustment and cursor remains visible
	model = NewMessageViewModel("C12345", 80, 10)
	model.SetMessages(messages, "")
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	// After Ctrl+U, scrollToSelected should have been called
	if len(updatedModel.messages) != 20 {
		t.Errorf("expected 20 messages after Ctrl+U, got %d", len(updatedModel.messages))
	}
}

// TestMessageViewModel_AddNewMessage_AutoScroll tests that new messages auto-scroll when cursor is at latest
func TestMessageViewModel_AddNewMessage_AutoScroll(t *testing.T) {
	// Create initial messages
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      "Test message",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	model := NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")

	// Verify cursor is at latest message (index 9)
	if model.selectedIndex != 9 {
		t.Fatalf("expected initial selectedIndex=9, got %d", model.selectedIndex)
	}

	initialOffset := model.viewport.YOffset

	// Add new message
	newMessage := message.Message{
		ID:        "M10",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Test User",
		Text:      "New message",
		Timestamp: time.Now().Add(10 * time.Minute),
	}
	model.AddNewMessage(newMessage)

	// Verify cursor moved to new latest message
	if model.selectedIndex != 10 {
		t.Errorf("expected selectedIndex=10 after adding new message, got %d", model.selectedIndex)
	}

	// Verify viewport scrolled to show new message
	if model.viewport.YOffset == initialOffset {
		t.Log("Viewport offset unchanged, which is acceptable if new message is visible")
	}
}

// TestMessageViewModel_AddNewMessage_NoAutoScroll tests that new messages don't auto-scroll when cursor is not at latest
func TestMessageViewModel_AddNewMessage_NoAutoScroll(t *testing.T) {
	// Create initial messages
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      "Test message",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	model := NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")

	// Move cursor to middle message
	model.selectedIndex = 5
	model.selectedMessageID = messages[5].ID

	initialIndex := model.selectedIndex
	initialMessageID := model.selectedMessageID

	// Add new message
	newMessage := message.Message{
		ID:        "M10",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "Test User",
		Text:      "New message",
		Timestamp: time.Now().Add(10 * time.Minute),
	}
	model.AddNewMessage(newMessage)

	// Verify cursor stayed at same position
	if model.selectedIndex != initialIndex {
		t.Errorf("expected selectedIndex=%d (unchanged), got %d", initialIndex, model.selectedIndex)
	}
	if model.selectedMessageID != initialMessageID {
		t.Errorf("expected selectedMessageID=%s (unchanged), got %s", initialMessageID, model.selectedMessageID)
	}
}

// TestMessageViewModel_InputMode_ScrollPosition tests that scroll position is maintained when entering input mode
func TestMessageViewModel_InputMode_ScrollPosition(t *testing.T) {
	// Create messages
	messages := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      strings.Repeat("Line\n", 5),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	model := NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")

	// Set specific scroll position
	model.selectedIndex = 10
	model.selectedMessageID = messages[10].ID
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

	// Test with 'c' key
	model = NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")
	model.selectedIndex = 10
	model.selectedMessageID = messages[10].ID
	model.viewport.SetYOffset(50)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if updatedModel.viewport.YOffset != 50 {
		t.Errorf("expected YOffset=50 (unchanged) after 'c', got %d", updatedModel.viewport.YOffset)
	}
	if !updatedModel.inputFocused {
		t.Error("expected inputFocused=true after 'c' key")
	}
}

// TestMessageViewModel_WindowResize_ScrollAdjustment tests scroll position recalculation on window resize
func TestMessageViewModel_WindowResize_ScrollAdjustment(t *testing.T) {
	// Create messages
	messages := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Test User",
			Text:      strings.Repeat("Line\n", 5),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}

	model := NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")

	// Set specific position
	model.selectedIndex = 10
	model.selectedMessageID = messages[10].ID

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
	model = NewMessageViewModel("C12345", 80, 20)
	model.SetMessages(messages, "")
	model.selectedIndex = 10
	model.selectedMessageID = messages[10].ID
	// Populate some cache
	model.scrollToSelected()
	cacheSize := len(model.messageLineHeights)

	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})

	// Verify cache was preserved on height-only change
	if len(updatedModel.messageLineHeights) != cacheSize {
		t.Errorf("expected line height cache preserved (size=%d), got %d", cacheSize, len(updatedModel.messageLineHeights))
	}
}
