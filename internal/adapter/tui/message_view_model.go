package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yanosea/gosl/internal/domain/message"
)

const (
	viewportHeightReserved = 4
	messageTextIndent      = 5
)

type MessageViewModel struct {
	viewport         viewport.Model
	messages         []message.Message
	selectedIndex    int
	selectedMessageID string
	channelID        string
	nextCursor       string
	width            int
	height           int
	renderCache      *RenderCache
	stringBuilders   *StringBuilderPool
	isInitialized    bool
}

func NewMessageViewModel(channelID string, width, height int) MessageViewModel {
	vp := viewport.New(width, height-viewportHeightReserved)
	vp.YPosition = 0

	return MessageViewModel{
		viewport:       vp,
		messages:       []message.Message{},
		selectedIndex:  0,
		channelID:      channelID,
		nextCursor:     "",
		width:          width,
		height:         height,
		renderCache:    NewRenderCache(),
		stringBuilders: NewStringBuilderPool(),
	}
}

func (m MessageViewModel) Init() tea.Cmd {
	return nil
}

func (m MessageViewModel) Update(msg tea.Msg) (MessageViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil

		case "i", "c":
			return m, nil

		case "enter":
			if m.selectedIndex < len(m.messages) {
				selectedMsg := m.messages[m.selectedIndex]
				if selectedMsg.ReplyCount > 0 {
					return m, nil
				}
			}
			return m, nil

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.messages) {
					m.selectedMessageID = m.messages[m.selectedIndex].ID
				}
				m.viewport.SetContent(m.renderMessages())
				m.scrollToSelected()
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.messages)-1 {
				m.selectedIndex++
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.messages) {
					m.selectedMessageID = m.messages[m.selectedIndex].ID
				}
				m.viewport.SetContent(m.renderMessages())
				m.scrollToSelected()
			}
			return m, nil

		case "g":
			if len(m.messages) > 0 {
				m.selectedIndex = 0
				m.selectedMessageID = m.messages[m.selectedIndex].ID
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoTop()
			}
			return m, nil

		case "G":
			if len(m.messages) > 0 {
				m.selectedIndex = len(m.messages) - 1
				m.selectedMessageID = m.messages[m.selectedIndex].ID
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
			}
			return m, nil

		case "pgup", "ctrl+u":
			return m, nil

		case "pgdown", "ctrl+d":
			m.viewport.ViewDown()
			return m, nil

		case "ctrl+r":
			return m, nil
		}

	case tea.WindowSizeMsg:
		prevYOffset := m.viewport.YOffset
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - viewportHeightReserved
		m.viewport.SetContent(m.renderMessages())
		m.viewport.SetYOffset(prevYOffset)
		return m, nil

	case NewMessageMsg:
		if msg.ChannelID == m.channelID {
			currentYOffset := m.viewport.YOffset
			m.AddNewMessage(msg.Message)
			m.viewport.SetContent(m.renderMessages())
			m.viewport.SetYOffset(currentYOffset)
		}
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m MessageViewModel) View() string {
	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)
	sb.WriteString(headerStyle.Render(fmt.Sprintf("# %s", m.channelID)))
	sb.WriteString("\n\n")

	sb.WriteString(m.viewport.View())

	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Padding(1, 1)
	footer := "↑/k: Up | ↓/j: Down | g: Top | G: Bottom | Ctrl+U/D: Page | Enter: Thread | i/c: Reply | Esc: Back | q: Quit"
	sb.WriteString("\n")
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

func (m *MessageViewModel) SetMessages(messages []message.Message, cursor string) {
	isFirstLoad := !m.isInitialized
	currentYOffset := m.viewport.YOffset

	m.messages = messages
	m.nextCursor = cursor

	if isFirstLoad {
		if len(m.messages) > 0 {
			m.selectedIndex = len(m.messages) - 1
			m.selectedMessageID = m.messages[m.selectedIndex].ID
		} else {
			m.selectedIndex = 0
			m.selectedMessageID = ""
		}
	} else {
		found := false
		for i, msg := range m.messages {
			if msg.ID == m.selectedMessageID {
				m.selectedIndex = i
				found = true
				break
			}
		}

		if !found {
			if len(m.messages) > 0 {
				m.selectedIndex = len(m.messages) - 1
				m.selectedMessageID = m.messages[m.selectedIndex].ID
			} else {
				m.selectedIndex = 0
				m.selectedMessageID = ""
			}
		}
	}

	m.renderCache.InvalidateAll()
	m.viewport.SetContent(m.renderMessages())

	if isFirstLoad {
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetYOffset(currentYOffset)
	}
	m.isInitialized = true
}

func (m *MessageViewModel) AppendMessages(messages []message.Message, cursor string) {
	m.messages = append(messages, m.messages...)
	m.nextCursor = cursor
	m.viewport.SetContent(m.renderMessages())
}

func (m *MessageViewModel) AddNewMessage(msg message.Message) {
	m.messages = append(m.messages, msg)
}

func (m *MessageViewModel) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().Faint(true).Render("No messages in this channel")
	}

	sb := m.stringBuilders.Get()
	defer m.stringBuilders.Put(sb)

	for i, msg := range m.messages {
		isSelected := i == m.selectedIndex
		m.renderMessage(sb, msg, isSelected)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *MessageViewModel) renderMessage(sb *strings.Builder, msg message.Message, selected bool) {
	if !selected {
		if cached, found := m.renderCache.Get(msg.ID); found {
			sb.WriteString(cached)
			return
		}
	}

	msgBuilder := m.stringBuilders.Get()
	defer m.stringBuilders.Put(msgBuilder)

	userStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2"))

	timestampStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("8"))

	threadStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("4"))

	if selected {
		msgBuilder.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("> "))
	} else {
		msgBuilder.WriteString("  ")
	}

	msgBuilder.WriteString(userStyle.Render(msg.UserName))
	msgBuilder.WriteString(" ")
	msgBuilder.WriteString(timestampStyle.Render(msg.Timestamp.Format("2006-01-02 15:04:05")))
	msgBuilder.WriteString("\n")

	highlightedText := m.highlightText(msg.Text)

	lines := strings.Split(highlightedText, "\n")
	for i, line := range lines {
		if i > 0 {
			msgBuilder.WriteString("\n")
		}
		msgBuilder.WriteString(strings.Repeat(" ", messageTextIndent))
		msgBuilder.WriteString(line)
	}

	if msg.ReplyCount > 0 {
		msgBuilder.WriteString(" ")
		threadIndicator := fmt.Sprintf("(%d replies)", msg.ReplyCount)
		msgBuilder.WriteString(threadStyle.Render(threadIndicator))
	}

	msgBuilder.WriteString("\n")

	rendered := msgBuilder.String()

	if !selected {
		m.renderCache.Set(msg.ID, rendered)
	}

	sb.WriteString(rendered)
}

func (m *MessageViewModel) highlightText(text string) string {
	text = m.highlightURLs(text)
	text = m.highlightMentions(text)
	return text
}

func (m *MessageViewModel) highlightURLs(text string) string {
	urlPattern := `https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+`

	urlMatches := findAllMatches(text, urlPattern)

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Underline(true)

	for _, match := range urlMatches {
		styledURL := urlStyle.Render(match)
		text = strings.Replace(text, match, styledURL, 1)
	}

	return text
}

func (m *MessageViewModel) highlightMentions(text string) string {
	mentionPattern := `(^|[^a-zA-Z0-9_.])@([a-zA-Z0-9_-]+)`

	mentionMatches := findAllSubmatchesWithGroups(text, mentionPattern)

	mentionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Bold(true)

	for _, match := range mentionMatches {
		if len(match) >= 3 {
			prefix := match[1]
			mention := "@" + match[2]
			styledMention := prefix + mentionStyle.Render(mention)
			originalText := prefix + mention
			text = strings.Replace(text, originalText, styledMention, 1)
		}
	}

	return text
}

func findAllMatches(text, pattern string) []string {
	var matches []string

	if pattern == `https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+` {
		words := strings.Fields(text)
		for _, word := range words {
			if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
				url := word
				for len(url) > 0 {
					lastChar := url[len(url)-1]
					if lastChar == '.' || lastChar == ',' || lastChar == ';' || lastChar == ')' {
						if lastChar == ')' && strings.Contains(url, "(") {
							break
						}
						url = url[:len(url)-1]
					} else {
						break
					}
				}
				if url != "" {
					matches = append(matches, url)
				}
			}
		}
	}

	return matches
}

func findAllSubmatchesWithGroups(text, pattern string) [][]string {
	var matches [][]string

	if pattern == `(^|[^a-zA-Z0-9_.])@([a-zA-Z0-9_-]+)` {
		words := strings.Split(text, " ")
		for i, word := range words {
			atIndex := strings.Index(word, "@")
			if atIndex >= 0 {
				if atIndex > 0 {
					charBefore := word[atIndex-1]
					if isAlphanumeric(charBefore) || charBefore == '.' || charBefore == '_' {
						continue
					}
				}

				mention := word[atIndex+1:]
				username := ""
				for _, char := range mention {
					if isAlphanumeric(byte(char)) || char == '_' || char == '-' {
						username += string(char)
					} else {
						break
					}
				}

				if username != "" {
					prefix := ""
					if i == 0 && atIndex == 0 {
						prefix = ""
					} else if atIndex > 0 {
						prefix = word[:atIndex]
					} else {
						prefix = ""
					}
					matches = append(matches, []string{prefix + "@" + username, prefix, username})
				}
			}
		}
	}

	return matches
}

func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func (m *MessageViewModel) scrollToSelected() {
	if len(m.messages) == 0 {
		return
	}

	selectedLineStart := 0

	sb := m.stringBuilders.Get()
	defer m.stringBuilders.Put(sb)

	for i, msg := range m.messages {
		msgSB := m.stringBuilders.Get()
		isSelected := i == m.selectedIndex
		m.renderMessage(msgSB, msg, isSelected)
		msgSB.WriteString("\n")
		rendered := msgSB.String()
		m.stringBuilders.Put(msgSB)

		lineCount := strings.Count(rendered, "\n")

		if i < m.selectedIndex {
			selectedLineStart += lineCount
		} else if i == m.selectedIndex {
			break
		}
	}

	viewportHeight := m.viewport.Height

	desiredOffset := selectedLineStart - (viewportHeight / 3)

	if desiredOffset < 0 {
		desiredOffset = 0
	}

	m.viewport.SetYOffset(desiredOffset)
}
