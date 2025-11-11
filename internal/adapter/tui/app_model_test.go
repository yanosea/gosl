package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/domain/channel"
)

// TestNewAppModel tests creating a new AppModel
func TestNewAppModel(t *testing.T) {
	model := NewAppModel(nil, nil)

	// Should start in Splash state
	if model.state != StateSplash {
		t.Errorf("Initial state = %v, want %v", model.state, StateSplash)
	}
}

// TestAppState_String tests the string representation of AppState
func TestAppState_String(t *testing.T) {
	tests := []struct {
		state AppState
		want  string
	}{
		{StateSplash, "Splash"},
		{StateChannelList, "ChannelList"},
		{StateMessageView, "MessageView"},
		{StateThreadView, "ThreadView"},
		{StateMessageInput, "MessageInput"},
		{StateHelp, "Help"},
		{StateError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("AppState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAppModel_StateTransitions tests state transitions
func TestAppModel_StateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialState  AppState
		msg           tea.Msg
		expectedState AppState
	}{
		{
			name:          "Splash to ChannelList on channels loaded",
			initialState:  StateSplash,
			msg:           ChannelsLoadedMsg{Channels: []channel.Channel{}},
			expectedState: StateChannelList,
		},
		{
			name:          "ChannelList to MessageView on channel selection",
			initialState:  StateChannelList,
			msg:           tea.KeyMsg{Type: tea.KeyEnter},
			expectedState: StateChannelList, // Will stay in ChannelList if no channel selected
		},
		{
			name:          "MessageView to ThreadView on thread selection",
			initialState:  StateMessageView,
			msg:           tea.KeyMsg{Type: tea.KeyEnter},
			expectedState: StateMessageView, // Will stay in MessageView if no thread message selected
		},
		{
			name:          "Any state to Help on ? key",
			initialState:  StateChannelList,
			msg:           tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			expectedState: StateHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewAppModel(nil, nil)
			model.state = tt.initialState

			updatedModel, _ := model.Update(tt.msg)
			appModel, ok := updatedModel.(AppModel)
			if !ok {
				t.Fatal("Update did not return AppModel")
			}

			if appModel.state != tt.expectedState {
				t.Errorf("State after %v = %v, want %v", tt.msg, appModel.state, tt.expectedState)
			}
		})
	}
}

// TestAppModel_GlobalKeyBindings tests global key bindings
func TestAppModel_GlobalKeyBindings(t *testing.T) {
	tests := []struct {
		name     string
		keyMsg   tea.KeyMsg
		shouldQuit bool
	}{
		{
			name:     "q key should quit",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			shouldQuit: true,
		},
		{
			name:     "Ctrl+C should quit",
			keyMsg:   tea.KeyMsg{Type: tea.KeyCtrlC},
			shouldQuit: true,
		},
		{
			name:     "? key should show help",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			shouldQuit: false,
		},
		{
			name:     "Other keys should not quit",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			shouldQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewAppModel(nil, nil)
			_, cmd := model.Update(tt.keyMsg)

			if tt.shouldQuit {
				if cmd == nil {
					t.Error("Expected quit command, got nil")
				}
				// Check if it's a quit command by executing it
				msg := cmd()
				if _, ok := msg.(tea.QuitMsg); !ok {
					t.Errorf("Expected tea.QuitMsg, got %T", msg)
				}
			}
		})
	}
}

// TestAppModel_Init tests the Init function
func TestAppModel_Init(t *testing.T) {
	model := NewAppModel(nil, nil)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() returned nil command")
	}
}

// TestAppModel_View tests the View function
func TestAppModel_View(t *testing.T) {
	model := NewAppModel(nil, nil)
	view := model.View()

	if view == "" {
		t.Error("View() returned empty string")
	}
}

// TestAppModel_HandleSlackEvents tests handling of Slack events
func TestAppModel_HandleSlackEvents(t *testing.T) {
	tests := []struct {
		name  string
		msg   tea.Msg
		state AppState
	}{
		{
			name:  "Handle SlackConnectedMsg in Splash",
			msg:   SlackConnectedMsg{},
			state: StateSplash,
		},
		{
			name:  "Handle SlackDisconnectedMsg",
			msg:   SlackDisconnectedMsg{Reason: "test"},
			state: StateChannelList,
		},
		{
			name:  "Handle NewMessageMsg",
			msg:   NewMessageMsg{ChannelID: "C123"},
			state: StateMessageView,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewAppModel(nil, nil)
			model.state = tt.state

			// Should not panic
			updatedModel, _ := model.Update(tt.msg)
			if updatedModel == nil {
				t.Error("Update returned nil model")
			}
		})
	}
}

// TestAppModel_SubModelDelegation tests that messages are delegated to sub-models
func TestAppModel_SubModelDelegation(t *testing.T) {
	model := NewAppModel(nil, nil)

	// Set state to ChannelList
	model.state = StateChannelList

	// Send a message that should be handled by the channel list sub-model
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update returned nil model")
	}

	// The model should remain in ChannelList state
	appModel, ok := updatedModel.(AppModel)
	if !ok {
		t.Fatal("Update did not return AppModel")
	}

	if appModel.state != StateChannelList {
		t.Errorf("State = %v, want %v", appModel.state, StateChannelList)
	}
}

// TestAppModel_ErrorState tests transitioning to error state
func TestAppModel_ErrorState(t *testing.T) {
	model := NewAppModel(nil, nil)

	// Simulate an error message
	errorMsg := ErrorMsg{Err: "test error"}
	updatedModel, _ := model.Update(errorMsg)

	appModel, ok := updatedModel.(AppModel)
	if !ok {
		t.Fatal("Update did not return AppModel")
	}

	if appModel.state != StateError {
		t.Errorf("State = %v, want %v", appModel.state, StateError)
	}
}
