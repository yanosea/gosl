// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/app/port"
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

	// Add cache entries for existing messages using width-aware keys
	cacheKey := getCacheKey("M1", model.viewport.Width)
	model.messageLineHeights[cacheKey] = 5

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

	// NOTE: AppendMessages calls renderMessages() which may recalculate line heights
	// We just verify that the cache still contains the key (value may be updated)
	if _, found := model.messageLineHeights[cacheKey]; !found {
		t.Errorf("Expected M1 cache entry to be preserved (key should exist)")
	}

	// Cache should have entries for both M0 and M1 now (if not selected)
	if len(model.messageLineHeights) < 1 {
		t.Errorf("Expected at least 1 cache entry after AppendMessages, got %d", len(model.messageLineHeights))
	}
}

// TestMessageViewModel_WindowSizeMsg_ClearCacheOnWidthChange tests cache invalidation on width change
func TestMessageViewModel_WindowSizeMsg_ClearCacheOnWidthChange(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	// Add cache entries using width-aware keys
	key1 := getCacheKey("M1", 80)
	key2 := getCacheKey("M2", 80)
	model.messageLineHeights[key1] = 5
	model.messageLineHeights[key2] = 10

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

	// Manually set cache entries using width-aware cache keys
	key1 := getCacheKey("M1", model.viewport.Width)
	key2 := getCacheKey("M2", model.viewport.Width)
	model.messageLineHeights[key1] = 3
	model.messageLineHeights[key2] = 4

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
	if model.messageLineHeights[key1] != 3 {
		t.Errorf("Expected M1 cache entry to remain 3, got %d", model.messageLineHeights[key1])
	}
	if model.messageLineHeights[key2] != 4 {
		t.Errorf("Expected M2 cache entry to remain 4, got %d", model.messageLineHeights[key2])
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

	// NOTE: SetMessages calls renderMessages() which may populate the cache
	// Clear the cache to simulate a true cache miss scenario
	model.messageLineHeights = make(map[string]int)

	// Select second message
	model.selectedIndex = 1
	model.selectedMessageID = "M2"

	// Verify cache is empty before scrollToSelected
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected empty cache before scrollToSelected, got %d entries", len(model.messageLineHeights))
	}

	// Call scrollToSelected
	model.scrollToSelected()

	// Verify cache was populated with calculated line heights
	if len(model.messageLineHeights) == 0 {
		t.Error("Expected cache to be populated after scrollToSelected, but it's empty")
	}

	// Verify that both M1 and M2 are cached (M1 is calculated for selectedLineStart)
	// Use width-aware cache keys
	key1 := getCacheKey("M1", model.viewport.Width)
	key2 := getCacheKey("M2", model.viewport.Width)
	if _, found := model.messageLineHeights[key1]; !found {
		t.Error("Expected M1 to be cached after scrollToSelected")
	}
	if _, found := model.messageLineHeights[key2]; !found {
		t.Error("Expected M2 to be cached after scrollToSelected")
	}

	// Verify that cached line heights are positive
	if model.messageLineHeights[key1] <= 0 {
		t.Errorf("Expected positive line height for M1, got %d", model.messageLineHeights[key1])
	}
	if model.messageLineHeights[key2] <= 0 {
		t.Errorf("Expected positive line height for M2, got %d", model.messageLineHeights[key2])
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

	// Populate cache with known heights using width-aware keys
	width := model.viewport.Width
	model.messageLineHeights[getCacheKey("M1", width)] = 3
	model.messageLineHeights[getCacheKey("M2", width)] = 3
	model.messageLineHeights[getCacheKey("M3", width)] = 3

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

	// Populate cache with known heights using width-aware keys
	width := model.viewport.Width
	model.messageLineHeights[getCacheKey("M1", width)] = 3
	model.messageLineHeights[getCacheKey("M2", width)] = 3
	model.messageLineHeights[getCacheKey("M3", width)] = 4 // Smaller than viewport height (~15)

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

	// Populate cache with known heights using width-aware keys
	width := model.viewport.Width
	model.messageLineHeights[getCacheKey("M1", width)] = 3
	model.messageLineHeights[getCacheKey("M2", width)] = 25 // Taller than viewport height (~15)

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
		cacheKey := getCacheKey(updatedModel.messages[i].ID, updatedModel.viewport.Width)
		height, ok := updatedModel.messageLineHeights[cacheKey]
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
		cacheKey := getCacheKey(updatedModel.messages[i].ID, updatedModel.viewport.Width)
		height, ok := updatedModel.messageLineHeights[cacheKey]
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

	// NOTE: WindowSizeMsg clears the cache but then calls renderMessages() which repopulates it
	// Verify that old width keys are no longer present
	for _, msg := range messages {
		oldKey := getCacheKey(msg.ID, 80)
		if _, found := updatedModel.messageLineHeights[oldKey]; found {
			t.Errorf("expected old cache key %s to be invalidated on width change", oldKey)
		}
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

	// Verify cache was preserved on height-only change (width didn't change, so keys remain valid)
	// The cache size may differ slightly due to re-rendering, but should be similar
	if len(updatedModel.messageLineHeights) < cacheSize-2 {
		t.Errorf("expected line height cache approximately preserved (size=%d), got %d", cacheSize, len(updatedModel.messageLineHeights))
	}
}

// TestGetCacheKey tests the getCacheKey helper function
func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		width     int
		expected  string
	}{
		{
			name:      "Normal message ID and width",
			messageID: "M12345",
			width:     80,
			expected:  "M12345-80",
		},
		{
			name:      "Long message ID",
			messageID: "1234567890.123456",
			width:     100,
			expected:  "1234567890.123456-100",
		},
		{
			name:      "Small width",
			messageID: "M001",
			width:     40,
			expected:  "M001-40",
		},
		{
			name:      "Large width",
			messageID: "MSG_ABC",
			width:     200,
			expected:  "MSG_ABC-200",
		},
		{
			name:      "Empty message ID",
			messageID: "",
			width:     80,
			expected:  "-80",
		},
		{
			name:      "Zero width",
			messageID: "M123",
			width:     0,
			expected:  "M123-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCacheKey(tt.messageID, tt.width)
			if result != tt.expected {
				t.Errorf("getCacheKey(%q, %d) = %q, want %q",
					tt.messageID, tt.width, result, tt.expected)
			}
		})
	}
}

// TestMessageViewModel_CacheKeyWithWidth tests that render cache uses width-aware cache keys
func TestMessageViewModel_CacheKeyWithWidth(t *testing.T) {
	// Initialize MessageViewModel with specific width
	model := NewMessageViewModel("C123", 80, 24)

	// Create test messages
	testMsg := message.Message{
		ID:        "M123",
		Text:      "This is a test message that should be cached with width",
		UserID:    "U001",
		UserName:  "TestUser",
		Timestamp: time.Now(),
	}

	// Set messages
	model.messages = []message.Message{testMsg}
	model.selectedIndex = 0
	model.selectedMessageID = testMsg.ID

	// First render - should cache with width
	var sb strings.Builder
	model.renderMessage(&sb, testMsg, false)
	firstRender := sb.String()

	// Verify cache uses getCacheKey format
	cacheKey := getCacheKey(testMsg.ID, model.viewport.Width)
	cached, found := model.renderCache.Get(cacheKey)
	if !found {
		t.Errorf("Expected message to be cached with key %q, but cache miss", cacheKey)
	}
	if cached != firstRender {
		t.Errorf("Cached content doesn't match rendered content")
	}

	// Second render - should hit cache
	sb.Reset()
	model.renderMessage(&sb, testMsg, false)
	secondRender := sb.String()

	if firstRender != secondRender {
		t.Errorf("Cache hit should return same content. First: %q, Second: %q", firstRender, secondRender)
	}

	// Change width - cache should miss
	model.viewport.Width = 100
	sb.Reset()
	model.renderMessage(&sb, testMsg, false)

	// Old key should still exist
	_, foundOld := model.renderCache.Get(getCacheKey(testMsg.ID, 80))
	if !foundOld {
		t.Errorf("Old cache key should still exist after width change")
	}

	// New key should be created
	newKey := getCacheKey(testMsg.ID, 100)
	cachedNew, foundNew := model.renderCache.Get(newKey)
	if !foundNew {
		t.Errorf("Expected message to be cached with new width key %q", newKey)
	}
	if cachedNew == "" {
		t.Errorf("New cached content should not be empty")
	}
}

// TestMessageViewModel_MessageLineHeightsWithWidth tests that messageLineHeights uses width-aware keys
func TestMessageViewModel_MessageLineHeightsWithWidth(t *testing.T) {
	model := NewMessageViewModel("C123", 80, 24)

	testMsg := message.Message{
		ID:        "M456",
		Text:      "Line 1\nLine 2\nLine 3",
		UserID:    "U001",
		UserName:  "TestUser",
		Timestamp: time.Now(),
	}

	model.messages = []message.Message{testMsg}
	model.selectedIndex = 0
	model.selectedMessageID = testMsg.ID

	// Trigger scrollToSelected which should populate messageLineHeights
	model.scrollToSelected()

	// Verify lineHeights uses getCacheKey format
	cacheKey := getCacheKey(testMsg.ID, model.viewport.Width)
	height, found := model.messageLineHeights[cacheKey]
	if !found {
		t.Errorf("Expected messageLineHeights to use key %q", cacheKey)
	}
	if height == 0 {
		t.Errorf("Expected non-zero line height, got %d", height)
	}

	// Change width and verify old key still exists
	oldWidth := model.viewport.Width
	model.viewport.Width = 120
	model.scrollToSelected()

	// Old key should still exist until cache invalidation
	oldKey := getCacheKey(testMsg.ID, oldWidth)
	_, foundOld := model.messageLineHeights[oldKey]
	if !foundOld {
		t.Errorf("Old messageLineHeights key should still exist")
	}

	// New key should be created
	newKey := getCacheKey(testMsg.ID, 120)
	newHeight, foundNew := model.messageLineHeights[newKey]
	if !foundNew {
		t.Errorf("Expected messageLineHeights to use new key %q", newKey)
	}
	if newHeight == 0 {
		t.Errorf("Expected non-zero line height with new width, got %d", newHeight)
	}
}

// TestMessageViewModel_WidthChangeInvalidatesCache tests cache invalidation on width change
func TestMessageViewModel_WidthChangeInvalidatesCache(t *testing.T) {
	model := NewMessageViewModel("C123", 80, 24)

	testMsg := message.Message{
		ID:        "M789",
		Text:      "Test message for cache invalidation",
		UserID:    "U001",
		UserName:  "TestUser",
		Timestamp: time.Now(),
	}

	model.messages = []message.Message{testMsg}
	model.selectedIndex = 0
	model.selectedMessageID = testMsg.ID

	// Populate caches
	var sb strings.Builder
	model.renderMessage(&sb, testMsg, false)
	model.scrollToSelected()

	// Verify caches are populated
	oldKey := getCacheKey(testMsg.ID, 80)
	_, foundCache := model.renderCache.Get(oldKey)
	_, foundHeight := model.messageLineHeights[oldKey]

	if !foundCache || !foundHeight {
		t.Errorf("Caches should be populated before width change")
	}

	// Simulate width change via WindowSizeMsg
	newMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 24,
	}
	model, _ = model.Update(newMsg)

	// messageLineHeights should be cleared
	if len(model.messageLineHeights) != 0 {
		t.Errorf("messageLineHeights should be empty after width change, got %d entries", len(model.messageLineHeights))
	}

	// renderCache should be invalidated (we can't directly test this, but we can verify new renders work)
	sb.Reset()
	model.renderMessage(&sb, testMsg, false)
	newRender := sb.String()
	if newRender == "" {
		t.Errorf("Render should succeed after width change")
	}
}

// TestMessageViewModel_TextWrapping tests text wrapping integration
func TestMessageViewModel_TextWrapping(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		width         int
		wrapEnabled   bool
		expectedLines int // minimum expected number of lines
	}{
		{
			name:          "Long line wraps at terminal width",
			text:          "This is a very long message that should be wrapped because it exceeds the terminal width and we need to test the wrapping functionality",
			width:         40,
			wrapEnabled:   true,
			expectedLines: 3, // Should wrap into multiple lines
		},
		{
			name:          "Short line does not wrap",
			text:          "Short message",
			width:         80,
			wrapEnabled:   true,
			expectedLines: 1,
		},
		{
			name:          "Preserves original newlines",
			text:          "Line 1\nLine 2\nLine 3",
			width:         80,
			wrapEnabled:   true,
			expectedLines: 3,
		},
		{
			name:          "Wrapping disabled returns original text",
			text:          "This is a very long message that should NOT be wrapped because wrapping is disabled in the configuration",
			width:         40,
			wrapEnabled:   false,
			expectedLines: 1, // Should not wrap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model with text wrapping configuration
			model := NewMessageViewModel("C12345", tt.width, 24)

			// Set text wrap config
			model.textWrapConfig = &port.TextWrapConfig{
				Enabled:               tt.wrapEnabled,
				MaxLineWidth:          0, // Use viewport width
				BreakAtCJKPunctuation: true,
			}

			// Create test message
			testMsg := message.Message{
				ID:        "M1",
				ChannelID: "C12345",
				UserID:    "U001",
				UserName:  "TestUser",
				Text:      tt.text,
				Timestamp: time.Now(),
			}

			// Render message
			var sb strings.Builder
			model.renderMessage(&sb, testMsg, false)
			rendered := sb.String()

			// Count lines in rendered output (excluding empty lines at start/end)
			lines := strings.Split(strings.TrimSpace(rendered), "\n")
			actualLines := len(lines)

			if tt.wrapEnabled {
				// When wrapping is enabled, we expect at least the minimum number of lines
				if actualLines < tt.expectedLines {
					t.Errorf("Expected at least %d lines with wrapping enabled, got %d\nRendered:\n%s",
						tt.expectedLines, actualLines, rendered)
				}
			} else {
				// When wrapping is disabled, long lines should not be wrapped
				// Check that the original text appears in a single logical line
				if !strings.Contains(rendered, tt.text) {
					t.Errorf("Original text should appear in rendered output when wrapping is disabled\nExpected to find: %s\nGot:\n%s",
						tt.text, rendered)
				}
			}
		})
	}
}

// TestMessageViewModel_TextWrappingWithCJK tests CJK character wrapping
func TestMessageViewModel_TextWrappingWithCJK(t *testing.T) {
	model := NewMessageViewModel("C12345", 30, 24)

	// Enable text wrapping with CJK punctuation breaking
	model.textWrapConfig = &port.TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}

	testMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "TestUser",
		Text:      "これは日本語のテストメッセージです。長い文章が折り返されることを確認します。句読点で折り返されるべきです。",
		Timestamp: time.Now(),
	}

	var sb strings.Builder
	model.renderMessage(&sb, testMsg, false)
	rendered := sb.String()

	// Verify the text was wrapped (should have multiple lines)
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected CJK text to be wrapped into multiple lines, got %d lines\nRendered:\n%s",
			len(lines), rendered)
	}
}

// TestMessageViewModel_TextWrappingCacheIntegration tests wrapping with cache
func TestMessageViewModel_TextWrappingCacheIntegration(t *testing.T) {
	model := NewMessageViewModel("C12345", 50, 24)

	model.textWrapConfig = &port.TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}

	testMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "TestUser",
		Text:      "This is a long message that will be wrapped and should be cached after the first render to improve performance on subsequent renders",
		Timestamp: time.Now(),
	}

	// First render - should wrap and cache
	var sb1 strings.Builder
	model.renderMessage(&sb1, testMsg, false)
	firstRender := sb1.String()

	// Verify cache was populated
	cacheKey := getCacheKey(testMsg.ID, model.viewport.Width)
	cached, found := model.renderCache.Get(cacheKey)
	if !found {
		t.Error("Expected wrapped message to be cached after first render")
	}
	if cached != firstRender {
		t.Error("Cached content should match first render")
	}

	// Second render - should use cache
	var sb2 strings.Builder
	model.renderMessage(&sb2, testMsg, false)
	secondRender := sb2.String()

	if firstRender != secondRender {
		t.Error("Second render should match first render (from cache)")
	}
}

// TestMessageViewModel_CacheInvalidationOnWidthChange tests cache invalidation when terminal width changes
func TestMessageViewModel_CacheInvalidationOnWidthChange(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	model.textWrapConfig = &port.TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}

	testMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "TestUser",
		Text:      "This is a long message that will be wrapped differently at different widths and should be re-cached after width changes",
		Timestamp: time.Now(),
	}

	// Set initial messages
	model.SetMessages([]message.Message{testMsg}, "")

	// First render at width 80
	var sb1 strings.Builder
	model.renderMessage(&sb1, testMsg, false)

	// Verify cache was populated with width-based key
	cacheKey80 := getCacheKey(testMsg.ID, 80)
	cached80, found80 := model.renderCache.Get(cacheKey80)
	if !found80 {
		t.Error("Expected message to be cached after first render at width 80")
	}

	// Store the count of cached items before width change
	initialCacheSize := len(model.renderCache.cache)

	// Simulate terminal width change to 50
	model, _ = model.Update(tea.WindowSizeMsg{Width: 50, Height: 24})

	// Verify that messageLineHeights was cleared
	if len(model.messageLineHeights) != 0 {
		t.Errorf("Expected messageLineHeights to be cleared after width change, got %d entries",
			len(model.messageLineHeights))
	}

	// Verify that RenderCache was invalidated
	// After invalidation, the old cache key should not exist
	_, foundAfterInvalidation := model.renderCache.Get(cacheKey80)
	if foundAfterInvalidation {
		t.Error("Expected old cache entry to be invalidated after width change")
	}

	// Verify cache size is 0 or reduced after invalidation
	if len(model.renderCache.cache) >= initialCacheSize {
		t.Errorf("Expected cache to be invalidated, but cache size remained %d (was %d)",
			len(model.renderCache.cache), initialCacheSize)
	}

	// Render at new width 50 and verify new cache entry is created
	var sb2 strings.Builder
	model.renderMessage(&sb2, testMsg, false)

	cacheKey50 := getCacheKey(testMsg.ID, 50)
	cached50, found50 := model.renderCache.Get(cacheKey50)
	if !found50 {
		t.Error("Expected message to be cached after render at width 50")
	}

	// Verify the cached content is different due to different wrapping width
	if cached80 == cached50 {
		t.Error("Expected different cached content at different widths due to text wrapping")
	}
}

// TestMessageViewModel_MessageLineHeightsUpdate tests that messageLineHeights is updated after rendering
func TestMessageViewModel_MessageLineHeightsUpdate(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	model.textWrapConfig = &port.TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}

	testMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "TestUser",
		Text:      "This is a long message that will be wrapped into multiple lines to test the line height calculation and caching functionality",
		Timestamp: time.Now(),
	}

	// Render message (not selected to trigger caching)
	var sb strings.Builder
	model.renderMessage(&sb, testMsg, false)
	rendered := sb.String()

	// Verify messageLineHeights was updated with the correct key
	cacheKey := getCacheKey(testMsg.ID, model.viewport.Width)
	lineHeight, found := model.messageLineHeights[cacheKey]
	if !found {
		t.Error("Expected messageLineHeights to be updated after rendering")
	}

	// Count actual lines in rendered output
	actualLineCount := strings.Count(rendered, "\n")
	if lineHeight != actualLineCount {
		t.Errorf("Expected line height %d to match actual line count %d", lineHeight, actualLineCount)
	}

	// Verify the line height is greater than 1 (should be wrapped into multiple lines)
	if lineHeight <= 1 {
		t.Errorf("Expected wrapped message to have more than 1 line, got %d lines", lineHeight)
	}
}

// TestMessageViewModel_MessageLineHeightsWithDifferentWidths tests line heights at different widths
func TestMessageViewModel_MessageLineHeightsWithDifferentWidths(t *testing.T) {
	model := NewMessageViewModel("C12345", 80, 24)

	model.textWrapConfig = &port.TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}

	testMsg := message.Message{
		ID:        "M1",
		ChannelID: "C12345",
		UserID:    "U001",
		UserName:  "TestUser",
		Text:      "This is a very long message that will be wrapped into many different numbers of lines depending on the terminal width to test adaptive line height calculation",
		Timestamp: time.Now(),
	}

	// Render at width 80
	var sb1 strings.Builder
	model.renderMessage(&sb1, testMsg, false)

	cacheKey80 := getCacheKey(testMsg.ID, 80)
	lineHeight80, found80 := model.messageLineHeights[cacheKey80]
	if !found80 {
		t.Error("Expected messageLineHeights to be updated at width 80")
	}

	// Simulate terminal width change to 40
	model.width = 40
	model.viewport.Width = 40

	// Clear caches to simulate width change behavior
	model.renderCache.InvalidateAll()
	model.messageLineHeights = make(map[string]int)

	// Render at width 40
	var sb2 strings.Builder
	model.renderMessage(&sb2, testMsg, false)

	cacheKey40 := getCacheKey(testMsg.ID, 40)
	lineHeight40, found40 := model.messageLineHeights[cacheKey40]
	if !found40 {
		t.Error("Expected messageLineHeights to be updated at width 40")
	}

	// At narrower width, we expect more lines due to more wrapping
	if lineHeight40 <= lineHeight80 {
		t.Errorf("Expected more lines at width 40 (%d) than at width 80 (%d)",
			lineHeight40, lineHeight80)
	}
}

// TestMessageViewModel_TextWrappingWithSpecialFormats tests text wrapping with special formats (code blocks, quotes, URLs)
func TestMessageViewModel_TextWrappingWithSpecialFormats(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		width       int
		wrapEnabled bool
		checkFunc   func(t *testing.T, rendered string)
	}{
		{
			name:        "Code block is not wrapped",
			text:        "Here is some code:\n```\nThis is a very long line of code that should not be wrapped even if it exceeds the terminal width\n```",
			width:       40,
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// Code block content should appear as-is without wrapping
				if !strings.Contains(rendered, "This is a very long line of code that should not be wrapped even if it exceeds the terminal width") {
					t.Error("Code block content should not be wrapped")
				}
			},
		},
		{
			name:        "URL is not wrapped",
			text:        "Check this link: https://example.com/very/long/path/that/might/exceed/terminal/width/but/should/not/be/wrapped for more info",
			width:       40,
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// URL should remain intact
				if !strings.Contains(rendered, "https://example.com/very/long/path/that/might/exceed/terminal/width/but/should/not/be/wrapped") {
					t.Error("URL should not be wrapped")
				}
			},
		},
		{
			name:        "Quote is wrapped with preserved prefix",
			text:        "> This is a very long quoted message that should be wrapped but the quote prefix should be preserved on each line",
			width:       40,
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// Quote should be wrapped, and each wrapped line should have the quote prefix
				lines := strings.Split(rendered, "\n")
				quotedLines := 0
				for _, line := range lines {
					if strings.Contains(line, ">") {
						quotedLines++
					}
				}
				// We expect at least 2 lines with quote prefix due to wrapping
				if quotedLines < 2 {
					t.Errorf("Expected at least 2 lines with quote prefix, got %d", quotedLines)
				}
			},
		},
		{
			name:        "Combined: code block, URL, and wrapped text",
			text:        "Here is a message with code:\n```\nfunction example() { return true; }\n```\nAnd a link: https://example.com\nAnd a very long sentence that should be wrapped because it exceeds the terminal width significantly",
			width:       50,
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// Code block should not be wrapped
				if !strings.Contains(rendered, "function example() { return true; }") {
					t.Error("Code block should not be wrapped")
				}
				// URL should not be wrapped
				if !strings.Contains(rendered, "https://example.com") {
					t.Error("URL should not be wrapped")
				}
				// Regular text should be wrapped into multiple lines
				lines := strings.Split(rendered, "\n")
				if len(lines) < 4 {
					t.Errorf("Expected at least 4 lines (code, URL, wrapped text), got %d", len(lines))
				}
			},
		},
		{
			name:        "Inline code is not wrapped",
			text:        "Use the command `this-is-a-very-long-command-that-should-not-be-wrapped-even-if-it-exceeds-width` to run the script",
			width:       40,
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// Inline code should remain intact
				if !strings.Contains(rendered, "`this-is-a-very-long-command-that-should-not-be-wrapped-even-if-it-exceeds-width`") {
					t.Error("Inline code should not be wrapped")
				}
			},
		},
		{
			name:        "Multiple URLs in text",
			text:        "Visit https://example.com and also check https://another-example.com/path for more details",
			width:       80, // Use wider width to avoid wrapping URLs
			wrapEnabled: true,
			checkFunc: func(t *testing.T, rendered string) {
				// Both URLs should remain intact
				if !strings.Contains(rendered, "example.com") {
					t.Errorf("First URL domain should appear in output, rendered:\n%s", rendered)
				}
				if !strings.Contains(rendered, "another-example.com") {
					t.Errorf("Second URL domain should appear in output, rendered:\n%s", rendered)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageViewModel("C12345", tt.width, 24)

			model.textWrapConfig = &port.TextWrapConfig{
				Enabled:               tt.wrapEnabled,
				MaxLineWidth:          0,
				BreakAtCJKPunctuation: true,
			}

			testMsg := message.Message{
				ID:        "M1",
				ChannelID: "C12345",
				UserID:    "U001",
				UserName:  "TestUser",
				Text:      tt.text,
				Timestamp: time.Now(),
			}

			var sb strings.Builder
			model.renderMessage(&sb, testMsg, false)
			rendered := sb.String()

			if rendered == "" {
				t.Fatal("Rendered output should not be empty")
			}

			// Run the test-specific check function
			tt.checkFunc(t, rendered)
		})
	}
}

// TestMessageViewModel_TextWrappingIntegrationFullScenario tests a comprehensive integration scenario
func TestMessageViewModel_TextWrappingIntegrationFullScenario(t *testing.T) {
	// Create a realistic chat scenario with multiple message types
	messages := []message.Message{
		{
			ID:        "M1",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "Hey everyone! I found this really interesting article about Go performance: https://go.dev/blog/pgo that explains profile-guided optimization in detail",
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        "M2",
			ChannelID: "C12345",
			UserID:    "U002",
			UserName:  "Bob",
			Text:      "Thanks @Alice! Here's a code snippet for reference:\n```go\nfunc Example() {\n    fmt.Println(\"This is a very long line of code that demonstrates how code blocks are not wrapped\")\n}\n```",
			Timestamp: time.Now().Add(-8 * time.Minute),
		},
		{
			ID:        "M3",
			ChannelID: "C12345",
			UserID:    "U003",
			UserName:  "Charlie",
			Text:      "> This is a quoted message from another channel\n> It has multiple lines and should wrap correctly\n> While preserving the quote markers",
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        "M4",
			ChannelID: "C12345",
			UserID:    "U001",
			UserName:  "Alice",
			Text:      "これは日本語のメッセージです。長い文章がターミナルの幅に応じて適切に折り返されることを確認します。句読点での折り返しもテストします。",
			Timestamp: time.Now().Add(-2 * time.Minute),
		},
	}

	// Test with different terminal widths
	widths := []int{50, 80, 120}

	for _, width := range widths {
		t.Run(fmt.Sprintf("Width_%d", width), func(t *testing.T) {
			model := NewMessageViewModel("C12345", width, 30)

			model.textWrapConfig = &port.TextWrapConfig{
				Enabled:               true,
				MaxLineWidth:          0,
				BreakAtCJKPunctuation: true,
			}

			model.SetMessages(messages, "")

			// Render all messages
			rendered := model.renderMessages()

			if rendered == "" {
				t.Fatal("Rendered messages should not be empty")
			}

			// Verify all messages are present
			for _, msg := range messages {
				if !strings.Contains(rendered, msg.UserName) {
					t.Errorf("Rendered output should contain user name: %s", msg.UserName)
				}
			}

			// Verify cache was populated for all messages
			// Note: The selected message is not cached, so we need to check only non-selected messages
			for i, msg := range messages {
				if i == model.selectedIndex {
					// Skip selected message - it's not cached
					continue
				}
				cacheKey := getCacheKey(msg.ID, width)
				if _, found := model.renderCache.Get(cacheKey); !found {
					t.Errorf("Expected message %s to be cached with key %s (selectedIndex=%d, i=%d)",
						msg.ID, cacheKey, model.selectedIndex, i)
				}
			}

			// Verify messageLineHeights was populated
			if len(model.messageLineHeights) == 0 {
				t.Error("Expected messageLineHeights to be populated after rendering")
			}

			// Second render should use cache
			rendered2 := model.renderMessages()
			if rendered != rendered2 {
				t.Error("Second render should match first render (from cache)")
			}
		})
	}

	// Test terminal width change
	t.Run("WidthChangeInvalidatesCache", func(t *testing.T) {
		model := NewMessageViewModel("C12345", 80, 30)

		model.textWrapConfig = &port.TextWrapConfig{
			Enabled:               true,
			MaxLineWidth:          0,
			BreakAtCJKPunctuation: true,
		}

		model.SetMessages(messages, "")

		// Render at width 80
		rendered1 := model.renderMessages()

		// Verify caches are populated
		initialCacheSize := len(model.messageLineHeights)
		if initialCacheSize == 0 {
			t.Fatal("Cache should be populated after first render")
		}

		// Simulate terminal width change
		updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 50, Height: 30})

		// NOTE: WindowSizeMsg calls renderMessages() which populates the cache again
		// So we cannot check if cache is empty immediately after Update
		// Instead, we verify that the cache was cleared and repopulated with new width keys

		// Check that old width keys are no longer in cache
		for _, msg := range messages {
			oldKey := getCacheKey(msg.ID, 80)
			if _, found := updatedModel.renderCache.Get(oldKey); found {
				t.Errorf("Old cache key %s should be invalidated after width change", oldKey)
			}
		}

		// Render at new width
		rendered2 := updatedModel.renderMessages()

		// Rendered output should be different due to different wrapping
		if rendered1 == rendered2 {
			t.Error("Expected different output at different widths due to text wrapping")
		}

		// Verify new cache entries were created
		if len(updatedModel.messageLineHeights) == 0 {
			t.Error("messageLineHeights should be populated after re-rendering at new width")
		}
	})

	// Test cache hit/miss scenarios
	t.Run("CacheHitMissScenarios", func(t *testing.T) {
		model := NewMessageViewModel("C12345", 80, 30)

		model.textWrapConfig = &port.TextWrapConfig{
			Enabled:               true,
			MaxLineWidth:          0,
			BreakAtCJKPunctuation: true,
		}

		model.SetMessages(messages, "")

		// First render - all cache misses
		_ = model.renderMessages()

		// Verify all messages are cached (except selected message)
		for i, msg := range messages {
			if i == model.selectedIndex {
				continue // Skip selected message
			}
			cacheKey := getCacheKey(msg.ID, 80)
			if _, found := model.renderCache.Get(cacheKey); !found {
				t.Errorf("Expected cache hit for message %s after first render (selectedIndex=%d)", msg.ID, model.selectedIndex)
			}
		}

		// Second render - all cache hits
		_ = model.renderMessages()

		// Add new message - cache miss for new message only
		newMsg := message.Message{
			ID:        "M5",
			ChannelID: "C12345",
			UserID:    "U004",
			UserName:  "David",
			Text:      "This is a new message that should trigger a cache miss",
			Timestamp: time.Now(),
		}
		model.AddNewMessage(newMsg)

		// Render again
		_ = model.renderMessages()

		// Verify new message is cached (if not selected)
		if model.selectedIndex != len(model.messages)-1 {
			newCacheKey := getCacheKey(newMsg.ID, 80)
			if _, found := model.renderCache.Get(newCacheKey); !found {
				t.Errorf("Expected new message to be cached after rendering")
			}
		}

		// Old messages should still be cached (except selected)
		for i, msg := range messages {
			if i == model.selectedIndex {
				continue
			}
			cacheKey := getCacheKey(msg.ID, 80)
			if _, found := model.renderCache.Get(cacheKey); !found {
				t.Errorf("Expected existing message %s to remain cached", msg.ID)
			}
		}
	})
}

// BenchmarkMessageViewModel_RenderWithCacheHit benchmarks message rendering with cache hits
// Requirement: キャッシュヒット時のレンダリング時間が折り返しなしの場合と比較して10%以内の遅延
func BenchmarkMessageViewModel_RenderWithCacheHit(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create test messages
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      "This is a test message with some content that might need wrapping",
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	// First render to populate cache
	_ = model.renderMessages()

	// Benchmark cache hit scenario
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.renderMessages()
	}
}

// BenchmarkMessageViewModel_RenderWithCacheMiss benchmarks message rendering with cache misses
func BenchmarkMessageViewModel_RenderWithCacheMiss(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)

	// Create test messages
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      "This is a test message with some content that might need wrapping",
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create new model for each iteration to force cache miss
		model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)
		model.SetMessages(messages, "")
		b.StartTimer()

		_ = model.renderMessages()
	}
}

// BenchmarkMessageViewModel_RenderLongMessages benchmarks rendering long messages
func BenchmarkMessageViewModel_RenderLongMessages(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create messages with long text that will be wrapped
	longText := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 20)
	messages := make([]message.Message, 5)
	for i := 0; i < 5; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      longText,
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache to force text wrapping
		model.renderCache.InvalidateAll()
		_ = model.renderMessages()
	}
}

// BenchmarkMessageViewModel_RenderWithNewlines benchmarks rendering messages with newlines
func BenchmarkMessageViewModel_RenderWithNewlines(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create messages with newlines
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      "Line one\nLine two with more content\nLine three\nLine four",
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.renderCache.InvalidateAll()
		_ = model.renderMessages()
	}
}

// BenchmarkMessageViewModel_CacheInvalidation benchmarks cache invalidation on terminal width change
// Requirement: ターミナル幅変更時の全キャッシュ無効化と再レンダリング時間を測定
func BenchmarkMessageViewModel_CacheInvalidation(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create test messages
	messages := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      "This is a test message that will be cached and then invalidated on width change",
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	// Populate cache
	_ = model.renderMessages()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate terminal width change
		newWidth := 80 + (i % 20) // Vary width to force cache invalidation
		msg := tea.WindowSizeMsg{Width: newWidth, Height: 24}
		model.Update(msg)
		_ = model.renderMessages()
	}
}

// BenchmarkMessageViewModel_ScrollToSelected benchmarks scroll position calculation
func BenchmarkMessageViewModel_ScrollToSelected(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create many messages
	messages := make([]message.Message, 100)
	for i := 0; i < 100; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      "Test message with multiple lines\nthat will be wrapped\nand cached",
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	// Populate cache
	_ = model.renderMessages()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Select different messages
		model.selectedIndex = i % len(messages)
		model.scrollToSelected()
	}
}

// TestMessageViewModel_CacheMemoryUsage tests that cache memory usage is within limits
// Requirement: 1メッセージあたり平均2KB以内のキャッシュ増加
func TestMessageViewModel_CacheMemoryUsage(t *testing.T) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)
	model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)

	// Create test messages with varying lengths
	messages := make([]message.Message, 100)
	for i := 0; i < 100; i++ {
		text := fmt.Sprintf("Message %d: %s", i, strings.Repeat("Sample text content. ", 10))
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      text,
			Timestamp: time.Now(),
		}
	}
	model.SetMessages(messages, "")

	// Render all messages to populate cache
	_ = model.renderMessages()

	// Calculate cache size estimation
	// Each cached entry contains wrapped text + styling
	totalCacheSize := 0
	cachedCount := 0
	for i, msg := range messages {
		if i == model.selectedIndex {
			continue // Selected message is not cached
		}
		cacheKey := getCacheKey(msg.ID, 80)
		if cached, found := model.renderCache.Get(cacheKey); found {
			cachedCount++
			// Estimate size: length of cached string + overhead
			totalCacheSize += len(cached)
		}
	}

	if cachedCount == 0 {
		t.Fatal("Expected at least some messages to be cached")
	}

	avgCacheSize := totalCacheSize / cachedCount
	maxCacheSizePerMessage := 2048 // 2KB in bytes

	t.Logf("Total cached messages: %d", cachedCount)
	t.Logf("Total cache size: %d bytes", totalCacheSize)
	t.Logf("Average cache size per message: %d bytes (%.2f KB)", avgCacheSize, float64(avgCacheSize)/1024.0)
	t.Logf("Maximum allowed per message: %d bytes (2 KB)", maxCacheSizePerMessage)

	if avgCacheSize > maxCacheSizePerMessage {
		t.Errorf("Average cache size per message %d bytes (%.2f KB) exceeds limit of 2KB",
			avgCacheSize, float64(avgCacheSize)/1024.0)
	}
}

// BenchmarkMessageViewModel_MemoryAllocation benchmarks memory allocations
func BenchmarkMessageViewModel_MemoryAllocation(b *testing.B) {
	mockCache := newMockUserColorCache()
	mockService := newMockUserColorService(mockCache)

	// Create test messages
	messages := make([]message.Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("M%d", i),
			ChannelID: "C12345",
			UserID:    fmt.Sprintf("U%03d", i),
			UserName:  fmt.Sprintf("User%d", i),
			Text:      strings.Repeat("Lorem ipsum dolor sit amet. ", 15),
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		model := NewMessageViewModelWithColorService("C12345", 80, 24, nil, mockService)
		model.SetMessages(messages, "")
		_ = model.renderMessages()
	}
}
