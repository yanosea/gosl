package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	VimUp    key.Binding
	VimDown  key.Binding
	VimLeft  key.Binding
	VimRight key.Binding

	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Refresh key.Binding

	NewMessage key.Binding
	Reply      key.Binding

	Search key.Binding

	PageUp   key.Binding
	PageDown key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Back}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.VimUp, k.VimDown, k.VimLeft, k.VimRight},
		{k.Enter, k.Back, k.Quit},
		{k.NewMessage, k.Reply, k.Search},
		{k.PageUp, k.PageDown, k.Refresh},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "上に移動 / Move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "下に移動 / Move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "左に移動 / Move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "右に移動 / Move right"),
	),

	VimUp: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "上に移動 / Move up"),
	),
	VimDown: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "下に移動 / Move down"),
	),
	VimLeft: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "左に移動 / Move left"),
	),
	VimRight: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "右に移動 / Move right"),
	),

	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "選択 / Select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "戻る / Back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "終了 / Quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?", "f1"),
		key.WithHelp("?/F1", "ヘルプ / Help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "更新 / Refresh"),
	),

	NewMessage: key.NewBinding(
		key.WithKeys("i", "c"),
		key.WithHelp("i/c", "新規メッセージ / New message"),
	),
	Reply: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "返信 / Reply"),
	),

	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "検索 / Search"),
	),

	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup/ctrl+u", "前ページ / Page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdown/ctrl+d", "次ページ / Page down"),
	),
}

type HelpModel struct {
	keys      keyMap
	help      help.Model
	viewport  viewport.Model
	width     int
	height    int
	shortMode bool
}

func NewHelpModel(width, height int) HelpModel {
	vp := viewport.New(width-4, height-6)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(1, 2)

	h := help.New()
	h.Width = width - 4

	model := HelpModel{
		keys:      keys,
		help:      h,
		viewport:  vp,
		width:     width,
		height:    height,
		shortMode: false,
	}

	model.viewport.SetContent(model.renderHelpContent())

	return model
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, nil

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp, tea.KeyDown:
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "k":
			m.viewport.LineUp(1)
			return m, nil
		case "j":
			m.viewport.LineDown(1)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		m.help.Width = msg.Width - 4

		m.viewport.SetContent(m.renderHelpContent())
		return m, nil
	}

	return m, nil
}

func (m HelpModel) View() string {
	if m.viewport.TotalLineCount() == 0 {
		m.viewport.SetContent(m.renderHelpContent())
	}

	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Padding(0, 2).
		MarginBottom(1)

	sb.WriteString(headerStyle.Render("📖 gosl - Slack TUI Client ヘルプ / Help"))
	sb.WriteString("\n\n")

	sb.WriteString(m.viewport.View())
	sb.WriteString("\n\n")

	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Padding(0, 2)

	footer := "↑/↓ または j/k: スクロール / Scroll | Esc: 戻る / Back"
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

func (m *HelpModel) renderHelpContent() string {
	var sb strings.Builder

	introStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Padding(0, 1).
		MarginBottom(1)

	sb.WriteString(introStyle.Render("Slackのメッセージング機能をターミナルで提供します。"))
	sb.WriteString("\n")
	sb.WriteString(introStyle.Render("Provides Slack messaging functionality in the terminal."))
	sb.WriteString("\n\n")

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3")).
		Padding(0, 1)

	sb.WriteString(sectionStyle.Render("🧭 ナビゲーション / Navigation"))
	sb.WriteString("\n")
	sb.WriteString(m.help.View(m.keys))
	sb.WriteString("\n\n")

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true).
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	bindings := []struct {
		key  string
		desc string
	}{
		{"↑/↓ または k/j", "上下に移動 / Move up/down"},
		{"←/→ または h/l", "左右に移動 / Move left/right"},
		{"Enter", "選択・決定 / Select/Confirm"},
		{"Esc", "戻る・キャンセル / Back/Cancel"},
		{"", ""},
		{"?/F1", "ヘルプ表示切替 / Toggle help"},
		{"q/Ctrl+C", "アプリケーション終了 / Quit application"},
		{"", ""},
		{"i/c", "新規メッセージ入力 / New message"},
		{"r", "スレッド返信 / Reply to thread"},
		{"Ctrl+Enter", "メッセージ送信 / Send message"},
		{"", ""},
		{"/", "チャンネル検索 / Search channels"},
		{"Ctrl+R", "手動更新 / Manual refresh"},
		{"", ""},
		{"Page Up/Ctrl+U", "前のメッセージを読込 / Load previous messages"},
		{"Page Down/Ctrl+D", "次ページへ移動 / Move to next page"},
	}

	for _, b := range bindings {
		if b.key == "" {
			sb.WriteString("\n")
			continue
		}
		sb.WriteString("  ")
		sb.WriteString(keyStyle.Render(b.key))
		sb.WriteString(" ")
		sb.WriteString(descStyle.Render(b.desc))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	sb.WriteString(sectionStyle.Render("🖥️  画面遷移 / Screen Navigation"))
	sb.WriteString("\n\n")

	screens := []struct {
		name string
		desc string
	}{
		{"起動画面 / Splash", "Socket Mode接続確認 / Socket Mode connection verification"},
		{"チャンネル一覧 / Channel List", "チャンネル選択・検索 / Channel selection and search"},
		{"メッセージ一覧 / Message View", "メッセージ表示・送信 / Message display and sending"},
		{"スレッド表示 / Thread View", "スレッド返信表示・送信 / Thread replies display and sending"},
		{"メッセージ入力 / Message Input", "複数行メッセージ入力 / Multi-line message input"},
	}

	for _, s := range screens {
		sb.WriteString("  • ")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render(s.name))
		sb.WriteString("\n    ")
		sb.WriteString(descStyle.Render(s.desc))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *HelpModel) SetShortMode(short bool) {
	m.shortMode = short
}
