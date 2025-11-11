package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewSplashModel tests creating a new SplashModel
func TestNewSplashModel(t *testing.T) {
	model := NewSplashModel()

	if model.connectionStatus != ConnectionStatusConnecting {
		t.Errorf("Initial status = %v, want %v", model.connectionStatus, ConnectionStatusConnecting)
	}
}

// TestSplashModel_ConnectionStatus tests the connection status enum
func TestSplashModel_ConnectionStatus(t *testing.T) {
	tests := []struct {
		status ConnectionStatus
		want   string
	}{
		{ConnectionStatusConnecting, "Connecting"},
		{ConnectionStatusConnected, "Connected"},
		{ConnectionStatusFailed, "Failed"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("ConnectionStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSplashModel_Init tests initialization
func TestSplashModel_Init(t *testing.T) {
	model := NewSplashModel()
	cmd := model.Init()

	// Should return a spinner tick command
	if cmd == nil {
		t.Error("Init() returned nil command")
	}
}

// TestSplashModel_Update_Connected tests transition on successful connection
func TestSplashModel_Update_Connected(t *testing.T) {
	model := NewSplashModel()

	// Send connected message
	updatedModel, _ := model.Update(SlackConnectedMsg{})

	if updatedModel.connectionStatus != ConnectionStatusConnected {
		t.Errorf("Status = %v, want %v", updatedModel.connectionStatus, ConnectionStatusConnected)
	}
}

// TestSplashModel_Update_Disconnected tests handling connection failures
func TestSplashModel_Update_Disconnected(t *testing.T) {
	model := NewSplashModel()

	// Send disconnected message
	errorMsg := SlackDisconnectedMsg{Reason: "auth error"}
	updatedModel, _ := model.Update(errorMsg)

	if updatedModel.connectionStatus != ConnectionStatusFailed {
		t.Errorf("Status = %v, want %v", updatedModel.connectionStatus, ConnectionStatusFailed)
	}

	if updatedModel.errorMessage != "auth error" {
		t.Errorf("Error message = %v, want 'auth error'", updatedModel.errorMessage)
	}
}

// TestSplashModel_Update_Error tests handling generic errors
func TestSplashModel_Update_Error(t *testing.T) {
	model := NewSplashModel()

	// Send error message
	errorMsg := ErrorMsg{Err: "network error"}
	updatedModel, _ := model.Update(errorMsg)

	if updatedModel.connectionStatus != ConnectionStatusFailed {
		t.Errorf("Status = %v, want %v", updatedModel.connectionStatus, ConnectionStatusFailed)
	}

	if updatedModel.errorMessage != "network error" {
		t.Errorf("Error message = %v, want 'network error'", updatedModel.errorMessage)
	}
}

// TestSplashModel_View tests rendering
func TestSplashModel_View(t *testing.T) {
	tests := []struct {
		name   string
		status ConnectionStatus
		error  string
		want   []string // strings that should appear in the view
	}{
		{
			name:   "connecting state",
			status: ConnectionStatusConnecting,
			error:  "",
			want:   []string{"gosl", "Connecting"},
		},
		{
			name:   "connected state",
			status: ConnectionStatusConnected,
			error:  "",
			want:   []string{"gosl", "Connected"},
		},
		{
			name:   "failed state",
			status: ConnectionStatusFailed,
			error:  "test error",
			want:   []string{"gosl", "Failed", "test error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewSplashModel()
			model.connectionStatus = tt.status
			model.errorMessage = tt.error

			view := model.View()

			for _, expected := range tt.want {
				if !strings.Contains(view, expected) {
					t.Errorf("View() missing '%s'\nGot: %s", expected, view)
				}
			}
		})
	}
}

// TestSplashModel_Update_WindowSize tests window size handling
func TestSplashModel_Update_WindowSize(t *testing.T) {
	model := NewSplashModel()

	// Send window size message
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(windowMsg)

	if updatedModel.width != 80 {
		t.Errorf("Width = %v, want 80", updatedModel.width)
	}

	if updatedModel.height != 24 {
		t.Errorf("Height = %v, want 24", updatedModel.height)
	}
}

// TestSplashModel_Update_Spinner tests spinner updates
func TestSplashModel_Update_Spinner(t *testing.T) {
	model := NewSplashModel()

	// The spinner should handle its own tick messages
	// We just verify that Update doesn't crash with spinner messages
	updatedModel, cmd := model.Update(model.spinner.Tick())

	if cmd == nil {
		t.Error("Expected spinner tick command, got nil")
	}

	// Model should still be valid
	if updatedModel.connectionStatus != ConnectionStatusConnecting {
		t.Error("Spinner update changed connection status unexpectedly")
	}
}

// TestSplashModel_AutoTransition tests that connected state can trigger transition
func TestSplashModel_AutoTransition(t *testing.T) {
	model := NewSplashModel()

	// Simulate connection success
	model, _ = model.Update(SlackConnectedMsg{})

	if model.connectionStatus != ConnectionStatusConnected {
		t.Errorf("Status = %v, want %v", model.connectionStatus, ConnectionStatusConnected)
	}

	// The AppModel should handle the actual transition
	// SplashModel just updates its state
}

// TestSplashModel_KeyboardQuit tests quit key handling
func TestSplashModel_KeyboardQuit(t *testing.T) {
	model := NewSplashModel()

	// Send quit key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(keyMsg)

	// SplashModel should pass through key messages
	// The AppModel handles global quit keys
	// So we just verify it doesn't crash
	_ = cmd
}
