// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
)

// MockMessageSender is a mock implementation of MessageSender for testing
type MockMessageSender struct {
	sendMessageCalled      bool
	sendThreadReplyCalled  bool
	sendMessageError       error
	sendThreadReplyError   error
	lastChannelID          string
	lastThreadTS           string
	lastText               string
	getMembersCalled       bool
	getMembersResult       []user.User
	getMembersError        error
}

func (m *MockMessageSender) SendMessage(ctx context.Context, channelID, text string) error {
	m.sendMessageCalled = true
	m.lastChannelID = channelID
	m.lastText = text
	return m.sendMessageError
}

func (m *MockMessageSender) SendThreadReply(ctx context.Context, channelID, threadTS, text string) error {
	m.sendThreadReplyCalled = true
	m.lastChannelID = channelID
	m.lastThreadTS = threadTS
	m.lastText = text
	return m.sendThreadReplyError
}

func (m *MockMessageSender) GetChannelMembers(ctx context.Context, channelID string) ([]user.User, error) {
	m.getMembersCalled = true
	return m.getMembersResult, m.getMembersError
}

func (m *MockMessageSender) GetThreadReplies(ctx context.Context, channelID, threadTS string) (parent message.Message, replies []message.Message, err error) {
	// Mock implementation - not used in message input tests
	return message.Message{}, nil, nil
}

// TestMessageInputModel_Init tests the initialization of MessageInputModel
func TestMessageInputModel_Init(t *testing.T) {
	tests := []struct {
		name      string
		mode      InputMode
		channelID string
		threadTS  string
		width     int
		height    int
	}{
		{
			name:      "Initialize for channel message",
			mode:      InputModeChannelMessage,
			channelID: "C12345",
			threadTS:  "",
			width:     80,
			height:    24,
		},
		{
			name:      "Initialize for thread reply",
			mode:      InputModeThreadReply,
			channelID: "C12345",
			threadTS:  "1234567890.123456",
			width:     80,
			height:    24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageInputModel(tt.mode, tt.channelID, tt.threadTS, tt.width, tt.height)

			if model.mode != tt.mode {
				t.Errorf("expected mode %v, got %v", tt.mode, model.mode)
			}

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

			// Textarea should be initialized
			if model.textarea.Value() != "" {
				t.Errorf("expected empty textarea, got %s", model.textarea.Value())
			}
		})
	}
}

// TestInputMode_String tests the String method of InputMode
func TestInputMode_String(t *testing.T) {
	tests := []struct {
		mode     InputMode
		expected string
	}{
		{InputModeChannelMessage, "ChannelMessage"},
		{InputModeThreadReply, "ThreadReply"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.mode.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.mode.String())
			}
		})
	}
}

// TestMessageInputModel_Update_KeyBindings tests key bindings
func TestMessageInputModel_Update_KeyBindings(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		expectedAction string
	}{
		{
			name:           "Press Esc to cancel",
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedAction: "cancel",
		},
		// Note: In Bubble Tea v1.3, Ctrl+Enter is represented as a KeyMsg with String() == "ctrl+enter"
		// Testing this requires creating a proper KeyMsg, which is complex in unit tests
		// We'll test the string-based key handling separately
		{
			name:           "Press Enter for newline",
			msg:            tea.KeyMsg{Type: tea.KeyEnter},
			expectedAction: "newline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)
			_, cmd := model.Update(tt.msg)

			// For now, just verify that Update doesn't panic
			// Full command testing would require more complex setup
			_ = cmd
		})
	}
}

// TestMessageInputModel_SetValue tests setting the textarea value
func TestMessageInputModel_SetValue(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	text := "Test message content"
	model.SetValue(text)

	if model.textarea.Value() != text {
		t.Errorf("expected textarea value %s, got %s", text, model.textarea.Value())
	}
}

// TestMessageInputModel_GetValue tests getting the textarea value
func TestMessageInputModel_GetValue(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	text := "Test message"
	model.SetValue(text)

	value := model.GetValue()
	if value != text {
		t.Errorf("expected value %s, got %s", text, value)
	}
}

// TestMessageInputModel_Clear tests clearing the textarea
func TestMessageInputModel_Clear(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	model.SetValue("Some text")
	model.Clear()

	if model.textarea.Value() != "" {
		t.Errorf("expected empty textarea after clear, got %s", model.textarea.Value())
	}
}

// TestMessageInputModel_SetError tests setting an error message
func TestMessageInputModel_SetError(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	errorMsg := "Failed to send message"
	model.SetError(errorMsg)

	if model.errorMessage != errorMsg {
		t.Errorf("expected error message %s, got %s", errorMsg, model.errorMessage)
	}

	if !model.hasError {
		t.Error("expected hasError to be true")
	}
}

// TestMessageInputModel_ClearError tests clearing an error message
func TestMessageInputModel_ClearError(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	model.SetError("Some error")
	model.ClearError()

	if model.errorMessage != "" {
		t.Errorf("expected empty error message, got %s", model.errorMessage)
	}

	if model.hasError {
		t.Error("expected hasError to be false")
	}
}

// TestMessageInputModel_View tests the View function
func TestMessageInputModel_View(t *testing.T) {
	tests := []struct {
		name string
		mode InputMode
	}{
		{
			name: "Channel message mode view",
			mode: InputModeChannelMessage,
		},
		{
			name: "Thread reply mode view",
			mode: InputModeThreadReply,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewMessageInputModel(tt.mode, "C12345", "1234567890.123456", 80, 24)

			view := model.View()

			if view == "" {
				t.Error("expected non-empty view")
			}
		})
	}
}

// TestMessageInputModel_Focus tests focusing the textarea
func TestMessageInputModel_Focus(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	model.Focus()

	if !model.textarea.Focused() {
		t.Error("expected textarea to be focused")
	}
}

// TestMessageInputModel_Blur tests blurring the textarea
func TestMessageInputModel_Blur(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	model.Focus()
	model.Blur()

	if model.textarea.Focused() {
		t.Error("expected textarea to be blurred")
	}
}

// TestMessageInputModel_SetParentMessageText tests setting parent message text for thread replies
func TestMessageInputModel_SetParentMessageText(t *testing.T) {
	model := NewMessageInputModel(InputModeThreadReply, "C12345", "1234567890.123456", 80, 24)

	parentText := "This is the parent message"
	model.SetParentMessageText(parentText)

	if model.parentMessageText != parentText {
		t.Errorf("expected parentMessageText %s, got %s", parentText, model.parentMessageText)
	}
}

// TestMessageInputModel_WindowSize tests window size updates
func TestMessageInputModel_WindowSize(t *testing.T) {
	model := NewMessageInputModel(InputModeChannelMessage, "C12345", "", 80, 24)

	newWidth := 100
	newHeight := 30
	msg := tea.WindowSizeMsg{Width: newWidth, Height: newHeight}

	updatedModel, _ := model.Update(msg)

	if updatedModel.width != newWidth {
		t.Errorf("expected width %d, got %d", newWidth, updatedModel.width)
	}

	if updatedModel.height != newHeight {
		t.Errorf("expected height %d, got %d", newHeight, updatedModel.height)
	}
}

// TestMessageInputModel_SendMessage tests sending a channel message
func TestMessageInputModel_SendMessage(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		mode          InputMode
		mockError     error
		expectSuccess bool
	}{
		{
			name:          "Send channel message successfully",
			text:          "Hello, channel!",
			mode:          InputModeChannelMessage,
			mockError:     nil,
			expectSuccess: true,
		},
		{
			name:          "Send empty message fails validation",
			text:          "",
			mode:          InputModeChannelMessage,
			mockError:     nil,
			expectSuccess: false,
		},
		{
			name:          "Send message with service error",
			text:          "Test message",
			mode:          InputModeChannelMessage,
			mockError:     errors.New("network error"),
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMessageSender{sendMessageError: tt.mockError}
			model := NewMessageInputModelWithSender(tt.mode, "C12345", "", 80, 24, mock)
			model.SetValue(tt.text)

			// Trigger send by calling sendMessage directly
			cmd := model.sendMessage()
			if cmd == nil {
				t.Fatal("expected non-nil command")
			}

			// Execute the command to get the result message
			msg := cmd()
			sentMsg, ok := msg.(MessageSentMsg)
			if !ok {
				t.Fatalf("expected MessageSentMsg, got %T", msg)
			}

			if sentMsg.Success != tt.expectSuccess {
				t.Errorf("expected success=%v, got success=%v", tt.expectSuccess, sentMsg.Success)
			}

			// Verify that the mock was called appropriately for successful sends
			if tt.expectSuccess && tt.text != "" {
				if !mock.sendMessageCalled {
					t.Error("expected SendMessage to be called")
				}
				if mock.lastText != tt.text {
					t.Errorf("expected text %s, got %s", tt.text, mock.lastText)
				}
			}
		})
	}
}

// TestMessageInputModel_SendThreadReply tests sending a thread reply
func TestMessageInputModel_SendThreadReply(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		threadTS      string
		mockError     error
		expectSuccess bool
	}{
		{
			name:          "Send thread reply successfully",
			text:          "Reply to thread",
			threadTS:      "1234567890.123456",
			mockError:     nil,
			expectSuccess: true,
		},
		{
			name:          "Send empty reply fails validation",
			text:          "",
			threadTS:      "1234567890.123456",
			mockError:     nil,
			expectSuccess: false,
		},
		{
			name:          "Send reply with service error",
			text:          "Test reply",
			threadTS:      "1234567890.123456",
			mockError:     errors.New("network error"),
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMessageSender{sendThreadReplyError: tt.mockError}
			model := NewMessageInputModelWithSender(InputModeThreadReply, "C12345", tt.threadTS, 80, 24, mock)
			model.SetValue(tt.text)

			// Trigger send
			cmd := model.sendMessage()
			if cmd == nil {
				t.Fatal("expected non-nil command")
			}

			// Execute the command
			msg := cmd()
			sentMsg, ok := msg.(MessageSentMsg)
			if !ok {
				t.Fatalf("expected MessageSentMsg, got %T", msg)
			}

			if sentMsg.Success != tt.expectSuccess {
				t.Errorf("expected success=%v, got success=%v", tt.expectSuccess, sentMsg.Success)
			}

			// Verify mock was called for successful sends
			if tt.expectSuccess && tt.text != "" {
				if !mock.sendThreadReplyCalled {
					t.Error("expected SendThreadReply to be called")
				}
				if mock.lastThreadTS != tt.threadTS {
					t.Errorf("expected threadTS %s, got %s", tt.threadTS, mock.lastThreadTS)
				}
			}
		})
	}
}

// TestMessageInputModel_MentionSuggestions tests mention autocomplete functionality
func TestMessageInputModel_MentionSuggestions(t *testing.T) {
	mockUsers := []user.User{
		{ID: "U1", Name: "alice", DisplayName: "Alice"},
		{ID: "U2", Name: "bob", DisplayName: "Bob"},
		{ID: "U3", Name: "charlie", DisplayName: "Charlie"},
	}

	tests := []struct {
		name              string
		initialText       string
		mockUsers         []user.User
		mockError         error
		expectSuggestions bool
		expectedCount     int
	}{
		{
			name:              "Fetch suggestions for channel",
			initialText:       "@",
			mockUsers:         mockUsers,
			mockError:         nil,
			expectSuggestions: true,
			expectedCount:     3,
		},
		{
			name:              "No suggestions when error occurs",
			initialText:       "@",
			mockUsers:         nil,
			mockError:         errors.New("network error"),
			expectSuggestions: false,
			expectedCount:     0,
		},
		{
			name:              "Empty suggestions when no users",
			initialText:       "@",
			mockUsers:         []user.User{},
			mockError:         nil,
			expectSuggestions: false,
			expectedCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMessageSender{
				getMembersResult: tt.mockUsers,
				getMembersError:  tt.mockError,
			}
			model := NewMessageInputModelWithSender(InputModeChannelMessage, "C12345", "", 80, 24, mock)

			// Trigger mention fetch
			cmd := model.fetchMentionSuggestions()
			if cmd == nil {
				t.Fatal("expected non-nil command")
			}

			// Execute the command
			msg := cmd()
			suggestionsMsg, ok := msg.(MentionSuggestionsMsg)
			if !ok {
				t.Fatalf("expected MentionSuggestionsMsg, got %T", msg)
			}

			// Update model with suggestions
			updatedModel, _ := model.Update(suggestionsMsg)

			if tt.expectSuggestions {
				if !updatedModel.showingSuggestions {
					t.Error("expected showingSuggestions to be true")
				}
				if len(updatedModel.mentionSuggestions) != tt.expectedCount {
					t.Errorf("expected %d suggestions, got %d", tt.expectedCount, len(updatedModel.mentionSuggestions))
				}
			} else {
				if updatedModel.showingSuggestions {
					t.Error("expected showingSuggestions to be false")
				}
			}
		})
	}
}

// TestMessageInputModel_FilterMentionSuggestions tests filtering of mention suggestions
func TestMessageInputModel_FilterMentionSuggestions(t *testing.T) {
	mockUsers := []user.User{
		{ID: "U1", Name: "alice", DisplayName: "Alice"},
		{ID: "U2", Name: "alicia", DisplayName: "Alicia"},
		{ID: "U3", Name: "bob", DisplayName: "Bob"},
	}

	model := NewMessageInputModelWithSender(InputModeChannelMessage, "C12345", "", 80, 24, nil)
	model.mentionSuggestions = mockUsers
	model.showingSuggestions = true

	// Filter with "ali"
	filtered := model.filterMentionSuggestions("ali")

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered suggestions, got %d", len(filtered))
	}

	// Verify filtering is case-insensitive
	filteredUpper := model.filterMentionSuggestions("ALI")
	if len(filteredUpper) != 2 {
		t.Errorf("expected case-insensitive filtering, got %d results", len(filteredUpper))
	}
}

// TestMessageInputModel_SelectMentionSuggestion tests selecting a mention suggestion
func TestMessageInputModel_SelectMentionSuggestion(t *testing.T) {
	mockUsers := []user.User{
		{ID: "U1", Name: "alice", DisplayName: "Alice"},
		{ID: "U2", Name: "bob", DisplayName: "Bob"},
	}

	model := NewMessageInputModelWithSender(InputModeChannelMessage, "C12345", "", 80, 24, nil)
	model.SetValue("Hello @ali")
	model.mentionSuggestions = mockUsers
	model.showingSuggestions = true
	model.selectedSuggestion = 0

	// Select the first suggestion
	model.selectMentionSuggestion()

	// Verify that @alice was inserted
	value := model.GetValue()
	if !strings.Contains(value, "@alice") {
		t.Errorf("expected '@alice' in value, got: %s", value)
	}

	if model.showingSuggestions {
		t.Error("expected showingSuggestions to be false after selection")
	}
}
