package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/domain/channel"
)

// TestNewChannelListModel tests creating a new ChannelListModel
func TestNewChannelListModel(t *testing.T) {
	model := NewChannelListModel(80, 24)

	if model.searchMode {
		t.Error("Initial searchMode should be false")
	}

	if model.width != 80 {
		t.Errorf("Initial width = %v, want 80", model.width)
	}

	if model.height != 24 {
		t.Errorf("Initial height = %v, want 24", model.height)
	}
}

// TestChannelListModel_Init tests initialization
func TestChannelListModel_Init(t *testing.T) {
	model := NewChannelListModel(80, 24)
	cmd := model.Init()

	// List model init may or may not return a command
	_ = cmd
}

// TestChannelListModel_SetChannels tests setting channels
func TestChannelListModel_SetChannels(t *testing.T) {
	model := NewChannelListModel(80, 24)

	channels := []channel.Channel{
		channel.NewChannel("C001", "general", channel.TypePublic),
		channel.NewChannel("C002", "random", channel.TypePublic),
		channel.NewChannel("D001", "user1", channel.TypeDM),
	}

	model.SetChannels(channels)

	if len(model.channels) != 3 {
		t.Errorf("SetChannels() stored %d channels, want 3", len(model.channels))
	}
}

// TestChannelListModel_Update_SearchMode tests entering search mode
func TestChannelListModel_Update_SearchMode(t *testing.T) {
	model := NewChannelListModel(80, 24)

	// Send '/' key to enter search mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ := model.Update(keyMsg)

	if !updatedModel.searchMode {
		t.Error("'/' key should enable search mode")
	}
}

// TestChannelListModel_Update_ExitSearchMode tests exiting search mode
func TestChannelListModel_Update_ExitSearchMode(t *testing.T) {
	model := NewChannelListModel(80, 24)
	model.searchMode = true
	model.searchQuery = "test"

	// Send Esc key to exit search mode
	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(keyMsg)

	if updatedModel.searchMode {
		t.Error("Esc key should disable search mode")
	}

	if updatedModel.searchQuery != "" {
		t.Error("Esc key should clear search query")
	}
}

// TestChannelListModel_Update_SearchQuery tests updating search query
func TestChannelListModel_Update_SearchQuery(t *testing.T) {
	model := NewChannelListModel(80, 24)
	model.searchMode = true

	// Type 'gen' in search mode
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	model, _ = model.Update(keyMsg)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	model, _ = model.Update(keyMsg)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	model, _ = model.Update(keyMsg)

	if model.searchQuery != "gen" {
		t.Errorf("Search query = %v, want 'gen'", model.searchQuery)
	}
}

// TestChannelListModel_Update_Backspace tests backspace in search mode
func TestChannelListModel_Update_Backspace(t *testing.T) {
	model := NewChannelListModel(80, 24)
	model.searchMode = true
	model.searchQuery = "test"

	// Send backspace
	keyMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.Update(keyMsg)

	if updatedModel.searchQuery != "tes" {
		t.Errorf("After backspace, query = %v, want 'tes'", updatedModel.searchQuery)
	}
}

// TestChannelListModel_FilterChannels tests channel filtering
func TestChannelListModel_FilterChannels(t *testing.T) {
	model := NewChannelListModel(80, 24)

	channels := []channel.Channel{
		channel.NewChannel("C001", "general", channel.TypePublic),
		channel.NewChannel("C002", "random", channel.TypePublic),
		channel.NewChannel("C003", "development", channel.TypePublic),
		channel.NewChannel("D001", "user1", channel.TypeDM),
	}
	model.SetChannels(channels)

	// Filter with "gen"
	filtered := model.filterChannels("gen")

	if len(filtered) != 1 {
		t.Errorf("Filter 'gen' returned %d channels, want 1", len(filtered))
	}

	if len(filtered) > 0 && filtered[0].Name != "general" {
		t.Errorf("Filtered channel name = %v, want 'general'", filtered[0].Name)
	}
}

// TestChannelListModel_FilterChannels_EmptyQuery tests filtering with empty query
func TestChannelListModel_FilterChannels_EmptyQuery(t *testing.T) {
	model := NewChannelListModel(80, 24)

	channels := []channel.Channel{
		channel.NewChannel("C001", "general", channel.TypePublic),
		channel.NewChannel("C002", "random", channel.TypePublic),
	}
	model.SetChannels(channels)

	// Filter with empty query should return all
	filtered := model.filterChannels("")

	if len(filtered) != 2 {
		t.Errorf("Empty filter returned %d channels, want 2", len(filtered))
	}
}


// TestChannelListModel_View tests rendering
func TestChannelListModel_View(t *testing.T) {
	model := NewChannelListModel(80, 24)

	channels := []channel.Channel{
		{
			ID:              "C001",
			Name:            "general",
			ChannelType:     channel.TypePublic,
			UnreadCount:     5,
			LastMessageTime: time.Now(),
		},
		{
			ID:              "D001",
			Name:            "user1",
			ChannelType:     channel.TypeDM,
			UnreadCount:     0,
			LastMessageTime: time.Now(),
		},
	}
	model.SetChannels(channels)

	view := model.View()

	// View should contain channel names
	if !strings.Contains(view, "general") {
		t.Error("View should contain 'general' channel")
	}

	// View should show unread count
	if !strings.Contains(view, "5") {
		t.Error("View should show unread count")
	}
}

// TestChannelListModel_View_SearchMode tests rendering in search mode
func TestChannelListModel_View_SearchMode(t *testing.T) {
	model := NewChannelListModel(80, 24)
	model.searchMode = true
	model.searchQuery = "test"

	view := model.View()

	// View should show search query
	if !strings.Contains(view, "test") {
		t.Error("View should contain search query")
	}

	// View should indicate search mode
	if !strings.Contains(view, "Search") || !strings.Contains(view, "/") {
		t.Error("View should indicate search mode")
	}
}

// TestChannelListModel_GetSelectedChannel tests getting selected channel
func TestChannelListModel_GetSelectedChannel(t *testing.T) {
	model := NewChannelListModel(80, 24)

	channels := []channel.Channel{
		channel.NewChannel("C001", "general", channel.TypePublic),
		channel.NewChannel("C002", "random", channel.TypePublic),
	}
	model.SetChannels(channels)

	selected := model.GetSelectedChannel()

	if selected == nil {
		t.Error("GetSelectedChannel() should return a channel")
		return
	}

	if selected.ID != "C001" {
		t.Errorf("Selected channel ID = %v, want 'C001'", selected.ID)
	}
}

// TestChannelListModel_GetSelectedChannel_Empty tests getting selected with no channels
func TestChannelListModel_GetSelectedChannel_Empty(t *testing.T) {
	model := NewChannelListModel(80, 24)

	selected := model.GetSelectedChannel()

	if selected != nil {
		t.Error("GetSelectedChannel() should return nil when no channels")
	}
}

// TestChannelListModel_WindowSize tests window size updates
func TestChannelListModel_WindowSize(t *testing.T) {
	model := NewChannelListModel(80, 24)

	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ := model.Update(windowMsg)

	if updatedModel.width != 100 {
		t.Errorf("Width = %d, want 100", updatedModel.width)
	}

	if updatedModel.height != 30 {
		t.Errorf("Height = %d, want 30", updatedModel.height)
	}
}
