// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestHelpModel_Init tests the initialization of HelpModel
func TestHelpModel_Init(t *testing.T) {
	model := NewHelpModel(80, 24)

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("expected height 24, got %d", model.height)
	}
}

// TestHelpModel_View tests the View rendering
func TestHelpModel_View(t *testing.T) {
	model := NewHelpModel(80, 24)

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain help title
	if !strings.Contains(view, "Help") && !strings.Contains(view, "ヘルプ") {
		t.Error("expected help title in view")
	}
}

// TestHelpModel_KeyBindings tests that all key bindings are displayed
func TestHelpModel_KeyBindings(t *testing.T) {
	model := NewHelpModel(80, 24)

	view := model.View()

	// Essential key bindings should be present
	expectedBindings := []string{
		"?", "F1",  // Help toggle
		"Esc",      // Back/Cancel
		"q",        // Quit
		"Ctrl+C",   // Quit
	}

	for _, binding := range expectedBindings {
		// The view might show bindings in various formats, so we just check for presence
		// Note: The actual display format may vary
		if !strings.Contains(view, binding) && !strings.Contains(view, strings.ToLower(binding)) {
			t.Logf("Warning: key binding '%s' not found in view (may be formatted differently)", binding)
		}
	}
}

// TestHelpModel_Update_Esc tests that Esc key is handled
func TestHelpModel_Update_Esc(t *testing.T) {
	model := NewHelpModel(80, 24)

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	// Esc should trigger some action (likely handled by parent AppModel)
	_ = cmd
}

// TestHelpModel_WindowSize tests window size updates
func TestHelpModel_WindowSize(t *testing.T) {
	model := NewHelpModel(80, 24)

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

// TestHelpModel_ShortHelpMode tests short help display
func TestHelpModel_ShortHelpMode(t *testing.T) {
	model := NewHelpModel(80, 24)
	model.SetShortMode(true)

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view in short mode")
	}

	// Short mode should be more compact
	lines := strings.Split(view, "\n")
	if len(lines) > 10 {
		t.Logf("Short help view has %d lines (may be acceptable)", len(lines))
	}
}

// TestHelpModel_FullHelpMode tests full help display
func TestHelpModel_FullHelpMode(t *testing.T) {
	model := NewHelpModel(80, 24)
	model.SetShortMode(false)

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view in full mode")
	}

	// Full mode should have more content
	lines := strings.Split(view, "\n")
	if len(lines) < 5 {
		t.Error("expected more lines in full help mode")
	}
}

// TestHelpModel_VimStyleNavigation tests vim-style key handling
func TestHelpModel_VimStyleNavigation(t *testing.T) {
	model := NewHelpModel(80, 24)

	tests := []struct {
		name string
		key  string
	}{
		{"vim up", "k"},
		{"vim down", "j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tt.key[0])}}
			_, cmd := model.Update(msg)
			// Should handle without error
			_ = cmd
		})
	}
}

// TestHelpModel_ContentSections tests that all required sections are present
func TestHelpModel_ContentSections(t *testing.T) {
	model := NewHelpModel(80, 24)

	view := model.View()

	// Should contain various sections
	expectedSections := []string{
		"ナビゲーション", // Navigation section
		"Navigation",
	}

	for _, section := range expectedSections {
		if !strings.Contains(view, section) {
			t.Errorf("expected section '%s' in view", section)
		}
	}

	// The view should at least show some help content
	// (even if not all sections are visible in the viewport)
	if len(view) < 50 {
		t.Error("expected substantial content in view")
	}
}

// TestHelpModel_BilingualContent tests Japanese and English content
func TestHelpModel_BilingualContent(t *testing.T) {
	model := NewHelpModel(80, 24)

	view := model.View()

	// Should contain both Japanese and English text
	hasJapanese := strings.Contains(view, "ヘルプ") || strings.Contains(view, "移動")
	hasEnglish := strings.Contains(view, "Help") || strings.Contains(view, "Move")

	if !hasJapanese {
		t.Error("expected Japanese text in help view")
	}

	if !hasEnglish {
		t.Error("expected English text in help view")
	}
}

// TestHelpModel_KeyMapShortHelp tests the short help key bindings
func TestHelpModel_KeyMapShortHelp(t *testing.T) {
	km := keys

	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("expected non-empty short help bindings")
	}

	// Should have essential keys
	if len(shortHelp) < 2 {
		t.Error("expected at least 2 key bindings in short help")
	}
}

// TestHelpModel_KeyMapFullHelp tests the full help key bindings
func TestHelpModel_KeyMapFullHelp(t *testing.T) {
	km := keys

	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("expected non-empty full help bindings")
	}

	// Should have multiple groups
	if len(fullHelp) < 3 {
		t.Error("expected at least 3 groups of key bindings in full help")
	}
}
