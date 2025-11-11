package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yanosea/gosl/internal/domain/channel"
)

type channelItem struct {
	channel channel.Channel
}

func (i channelItem) FilterValue() string {
	return i.channel.Name
}

func (i channelItem) Title() string {
	return i.channel.Name
}

func (i channelItem) Description() string {
	typeIcon := ""
	switch i.channel.ChannelType {
	case channel.TypePublic:
		typeIcon = "#"
	case channel.TypePrivate:
		typeIcon = "🔒"
	case channel.TypeDM:
		typeIcon = "@"
	}

	unreadStr := ""
	if i.channel.UnreadCount > 0 {
		unreadStr = fmt.Sprintf(" (%d unread)", i.channel.UnreadCount)
	}

	timeStr := ""
	if !i.channel.LastMessageTime.IsZero() {
		timeStr = fmt.Sprintf(" • %s", i.channel.LastMessageTime.Format("15:04"))
	}

	return fmt.Sprintf("%s%s%s", typeIcon, unreadStr, timeStr)
}

type ChannelListModel struct {
	list        list.Model
	channels    []channel.Channel
	searchMode  bool
	searchQuery string
	width       int
	height      int
}

func NewChannelListModel(width, height int) ChannelListModel {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("243")).
		BorderForeground(lipgloss.Color("170"))

	l := list.New([]list.Item{}, delegate, width, height)
	l.Title = "Channels"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		MarginLeft(2)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	return ChannelListModel{
		list:       l,
		channels:   []channel.Channel{},
		searchMode: false,
		width:      width,
		height:     height,
	}
}

func (m ChannelListModel) Init() tea.Cmd {
	return nil
}

func (m *ChannelListModel) SetChannels(channels []channel.Channel) {
	m.channels = channels
	m.updateListItems()
}

func (m *ChannelListModel) updateListItems() {
	sorted := m.sortChannels(m.channels)

	if m.searchMode && m.searchQuery != "" {
		sorted = m.filterChannels(m.searchQuery)
	}

	items := make([]list.Item, len(sorted))
	for i, ch := range sorted {
		items[i] = channelItem{channel: ch}
	}

	m.list.SetItems(items)
}

func (m ChannelListModel) filterChannels(query string) []channel.Channel {
	if query == "" {
		return m.channels
	}

	query = strings.ToLower(query)
	filtered := []channel.Channel{}

	for _, ch := range m.channels {
		if strings.Contains(strings.ToLower(ch.Name), query) {
			filtered = append(filtered, ch)
		}
	}

	return filtered
}

func (m ChannelListModel) sortChannels(channels []channel.Channel) []channel.Channel {
	sorted := make([]channel.Channel, len(channels))
	copy(sorted, channels)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].UnreadCount != sorted[j].UnreadCount {
			return sorted[i].UnreadCount > sorted[j].UnreadCount
		}
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

func (m ChannelListModel) Update(msg tea.Msg) (ChannelListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.searchMode = false
				m.searchQuery = ""
				m.updateListItems()
				return m, nil

			case tea.KeyEnter:
				m.searchMode = false
				return m, nil

			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.updateListItems()
				}
				return m, nil

			case tea.KeyRunes:
				m.searchQuery += string(msg.Runes)
				m.updateListItems()
				return m, nil
			}
		} else {
			switch msg.String() {
			case "/":
				m.searchMode = true
				m.searchQuery = ""
				return m, nil

			case "g":
				m.list.Select(0)
				return m, nil

			case "G":
				itemCount := len(m.list.Items())
				if itemCount > 0 {
					m.list.Select(itemCount - 1)
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ChannelListModel) View() string {
	var s string

	if m.searchMode {
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

		s += searchStyle.Render("Search: /") + m.searchQuery
		s += "\n"
		s += lipgloss.NewStyle().Faint(true).Render("Type to filter, Enter to apply, Esc to cancel")
		s += "\n\n"
	}

	s += m.list.View()

	return s
}

func (m ChannelListModel) GetSelectedChannel() *channel.Channel {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return nil
	}

	if item, ok := selectedItem.(channelItem); ok {
		return &item.channel
	}

	return nil
}
