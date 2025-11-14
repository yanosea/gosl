package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/usercolor"
)

// ThreadViewModel manages the thread view for displaying parent message and replies.
type ThreadViewModel struct {
	viewport          viewport.Model
	parentMessage     message.Message
	replies           []message.Message
	selectedIndex     int
	selectedMessageID string
	channelID         string
	threadTS          string
	width             int
	height            int
	viewportHeight    int
	scrollOffset      int
	inputArea         textarea.Model
	inputFocused      bool
	sender            MessageSender
	isDarkBackground  bool
	userColorService  usercolor.Service
}

// NewThreadViewModel creates a new ThreadViewModel instance.
func NewThreadViewModel(channelID, threadTS string, width, height int) ThreadViewModel {
	return NewThreadViewModelWithSender(channelID, threadTS, width, height, nil)
}

// NewThreadViewModelWithSender creates a new ThreadViewModel instance with a MessageSender.
func NewThreadViewModelWithSender(channelID, threadTS string, width, height int, sender MessageSender) ThreadViewModel {
	return NewThreadViewModelWithColorService(channelID, threadTS, width, height, sender, nil)
}

// NewThreadViewModelWithColorService creates a new ThreadViewModel with user color support
func NewThreadViewModelWithColorService(channelID, threadTS string, width, height int, sender MessageSender, colorService usercolor.Service) ThreadViewModel {
	// Calculate heights: header(2) + input area(5) + footer(2) = 9 lines reserved
	inputHeight := 3
	reservedHeight := 9
	viewportHeight := height - reservedHeight
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	vp := viewport.New(width, viewportHeight)
	vp.YPosition = 0

	// Create input textarea
	ta := textarea.New()
	ta.Placeholder = "Type your reply... (Ctrl+Enter to send, Esc to exit input)"
	ta.SetWidth(width - 4)
	ta.SetHeight(inputHeight)
	ta.ShowLineNumbers = false
	ta.Blur() // Start with viewport focused

	return ThreadViewModel{
		viewport:         vp,
		parentMessage:    message.Message{},
		replies:          []message.Message{},
		selectedIndex:    0,
		channelID:        channelID,
		threadTS:         threadTS,
		width:            width,
		height:           height,
		viewportHeight:   viewportHeight,
		scrollOffset:     0,
		inputArea:        ta,
		inputFocused:     false,
		sender:           sender,
		userColorService: colorService,
	}
}

// Init initializes the ThreadViewModel.
func (m ThreadViewModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.startRefreshTimer())
}

// startRefreshTimer returns a command that triggers a periodic refresh.
func (m ThreadViewModel) startRefreshTimer() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ThreadRefreshTickMsg{}
	})
}

// Update handles messages and updates the ThreadViewModel.
func (m ThreadViewModel) Update(msg tea.Msg) (ThreadViewModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If input is focused, handle input-specific keys
		if m.inputFocused {
			switch msg.String() {
			case "esc":
				// Exit input mode, return focus to viewport
				m.inputFocused = false
				m.inputArea.Blur()
				return m, nil

			case "ctrl+enter", "alt+enter", "ctrl+j":
				// Send message
				return m, m.sendThreadReply()
			}

			// Update input area for other keys
			m.inputArea, cmd = m.inputArea.Update(msg)
			return m, cmd
		}

		// Viewport is focused - handle viewport navigation keys
		switch msg.String() {
		case "esc":
			// Return to message view (handled by AppModel)
			return m, nil

		case "i", "r":
			// Enter input mode - focus the input area
			m.inputFocused = true
			return m, m.inputArea.Focus()

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				if m.selectedIndex == 0 {
					m.selectedMessageID = m.parentMessage.ID
				} else if m.selectedIndex-1 < len(m.replies) {
					m.selectedMessageID = m.replies[m.selectedIndex-1].ID
				}
				m.viewport.SetContent(m.renderThread())
				m.scrollToSelected()
			}
			return m, nil

		case "down", "j":
			// Total items = parent + replies
			maxIndex := len(m.replies)
			if m.selectedIndex < maxIndex {
				m.selectedIndex++
				if m.selectedIndex == 0 {
					m.selectedMessageID = m.parentMessage.ID
				} else if m.selectedIndex-1 < len(m.replies) {
					m.selectedMessageID = m.replies[m.selectedIndex-1].ID
				}
				m.viewport.SetContent(m.renderThread())
				m.scrollToSelected()
			}
			return m, nil

		case "g":
			// Vim-style: gg to go to top
			m.selectedIndex = 0
			m.selectedMessageID = m.parentMessage.ID
			m.viewport.SetContent(m.renderThread())
			m.viewport.GotoTop()
			return m, nil

		case "G":
			// Vim-style: G to go to bottom
			m.selectedIndex = len(m.replies)
			if len(m.replies) > 0 {
				m.selectedMessageID = m.replies[len(m.replies)-1].ID
			} else {
				m.selectedMessageID = m.parentMessage.ID
			}
			m.viewport.SetContent(m.renderThread())
			m.viewport.GotoBottom()
			return m, nil

		case "pgup", "ctrl+u":
			// Scroll up in viewport
			m.viewport.ViewUp()
			return m, nil

		case "pgdown", "ctrl+d":
			// Scroll down in viewport
			m.viewport.ViewDown()
			return m, nil
		}

	case BackgroundColorMsg:
		// Update background theme
		m.isDarkBackground = msg.IsDark
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Recalculate layout
		inputHeight := 3
		reservedHeight := 9
		viewportHeight := msg.Height - reservedHeight
		if viewportHeight < 5 {
			viewportHeight = 5
		}
		m.viewportHeight = viewportHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight
		m.inputArea.SetWidth(msg.Width - 4)
		m.inputArea.SetHeight(inputHeight)
		m.viewport.SetContent(m.renderThread())
		return m, nil

	case NewMessageMsg:
		// Add new reply to the thread if it belongs to this thread
		if msg.ChannelID == m.channelID && msg.Message.ThreadTS == m.threadTS {
			// Check if cursor was at the latest message before adding new reply
			wasAtLatest := false
			if len(m.replies) > 0 {
				wasAtLatest = m.selectedIndex == len(m.replies)
			} else {
				wasAtLatest = m.selectedIndex == 0
			}

			m.AddReply(msg.Message)

			// If cursor was at latest, move cursor to new latest reply and scroll
			if wasAtLatest {
				m.selectedIndex = len(m.replies)
				m.selectedMessageID = m.replies[len(m.replies)-1].ID
			}

			m.viewport.SetContent(m.renderThread())

			// Scroll to bottom if was at latest
			if wasAtLatest {
				m.viewport.GotoBottom()
			}

			return m, nil
		}
		// If message doesn't match, fall through to default handling

	case MessageSentMsg:
		if msg.Success {
			// Clear input area on successful send
			m.inputArea.Reset()
			// Keep input mode active for continuous messaging
			// inputFocused remains true, textarea stays focused
			// Immediately refresh thread to show the new message
			return m, m.refreshThreadData()
		}
		return m, nil

	case ThreadRefreshTickMsg:
		// Refresh thread data and restart timer
		return m, tea.Batch(
			m.refreshThreadData(),
			m.startRefreshTimer(),
		)

	case ThreadLoadedMsg:
		// Update thread data when loaded (from refresh or after sending)
		if msg.ChannelID == m.channelID && msg.ThreadTS == m.threadTS {
			m.SetThread(msg.ParentMessage, msg.Replies)
			return m, nil
		}
		// If message doesn't match, fall through to default handling
	}

	// Update input area or viewport based on focus
	if m.inputFocused {
		m.inputArea, cmd = m.inputArea.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the ThreadViewModel.
func (m ThreadViewModel) View() string {
	var sb strings.Builder

	// Header: Thread indicator
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)
	sb.WriteString(headerStyle.Render(fmt.Sprintf("🧵 Thread in #%s", m.channelID)))
	sb.WriteString("\n\n")

	// Render content directly without viewport
	allLines := m.getAllThreadLines()
	visibleLines := m.getVisibleLines(allLines)

	// Ensure we always render exactly viewportHeight lines
	for i := 0; i < m.viewportHeight; i++ {
		if i < len(visibleLines) {
			sb.WriteString(visibleLines[i])
		}
		sb.WriteString("\n")
	}

	// Input area separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
	separator := strings.Repeat("─", m.width)
	sb.WriteString(separatorStyle.Render(separator))
	sb.WriteString("\n")

	// Input area
	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 1)

	if m.inputFocused {
		inputBoxStyle = inputBoxStyle.BorderForeground(lipgloss.Color("2"))
	}

	sb.WriteString(inputBoxStyle.Render(m.inputArea.View()))
	sb.WriteString("\n")

	// Footer: Help text
	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Padding(0, 1)

	var footer string
	if m.inputFocused {
		footer = "Ctrl+Enter: Send | Esc: Exit input | q: Quit"
	} else {
		footer = "↑/k: Up | ↓/j: Down | g: Top | G: Bottom | Ctrl+U/D: Page | i/r: Input | Esc: Back | q: Quit"
	}
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

// SetThread sets the parent message and replies for the thread view.
func (m *ThreadViewModel) SetThread(parent message.Message, replies []message.Message) {
	prevSelectedMsgID := m.selectedMessageID
	isFirstLoad := m.selectedMessageID == ""
	wasAtLatest := false

	// Check if cursor was at the latest message before update
	if !isFirstLoad {
		if len(m.replies) > 0 {
			wasAtLatest = m.selectedIndex == len(m.replies)
		} else {
			wasAtLatest = m.selectedIndex == 0
		}
	}

	m.parentMessage = parent
	m.replies = replies

	if isFirstLoad {
		if len(m.replies) > 0 {
			m.selectedIndex = len(m.replies)
			m.selectedMessageID = m.replies[len(m.replies)-1].ID
		} else {
			m.selectedIndex = 0
			m.selectedMessageID = parent.ID
		}
	} else {
		// If was at latest, always move cursor to new latest message
		if wasAtLatest {
			if len(m.replies) > 0 {
				m.selectedIndex = len(m.replies)
				m.selectedMessageID = m.replies[len(m.replies)-1].ID
			} else {
				m.selectedIndex = 0
				m.selectedMessageID = parent.ID
			}
		} else {
			// Try to find the previously selected message
			if prevSelectedMsgID == m.parentMessage.ID {
				m.selectedIndex = 0
			} else {
				for i, reply := range m.replies {
					if reply.ID == prevSelectedMsgID {
						m.selectedIndex = i + 1
						break
					}
				}
			}
			// If not found, keep the current selectedIndex (cursor stays in place)
		}
	}

	m.viewport.SetContent(m.renderThread())

	// Scroll to bottom if first load OR was at latest
	if isFirstLoad || wasAtLatest {
		m.viewport.GotoBottom()
	}
}

// AddReply adds a new reply to the thread (real-time update).
func (m *ThreadViewModel) AddReply(reply message.Message) {
	m.replies = append(m.replies, reply)
}

// renderThreadMessage renders a single message to the StringBuilder.
// indent specifies the number of spaces to indent (0 for parent, 2+ for replies).
// index is the position in the thread (0 for parent, 1+ for replies).
func (m *ThreadViewModel) renderThreadMessage(sb *strings.Builder, msg message.Message, indent int, index int) {
	// User name style
	userStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2"))

	// Timestamp style
	timestampStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("8"))

	// Thread parent indicator style
	parentStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5"))

	// Reply indicator style
	replyStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("6"))

	// Selection indicator
	isSelected := m.selectedIndex == index
	if isSelected {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("> "))
	} else {
		sb.WriteString("  ")
	}

	// Indentation
	for i := 0; i < indent; i++ {
		sb.WriteString(" ")
	}

	// Thread indicator
	if index == 0 {
		sb.WriteString(parentStyle.Render("┌─ "))
	} else {
		sb.WriteString(replyStyle.Render("└─ "))
	}

	// User name and timestamp
	sb.WriteString(userStyle.Render(msg.UserName))
	sb.WriteString(" ")
	sb.WriteString(timestampStyle.Render(msg.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString("\n")

	// Message text (with indentation)
	highlightedText := m.highlightText(msg.Text)

	// Calculate base indentation for message text
	baseIndent := "  " + strings.Repeat(" ", indent) + "   "

	// Apply user-specific background color if colorService is available
	var messageStyle lipgloss.Style
	if m.userColorService != nil {
		// Generate color from UserID
		adaptiveColor := m.userColorService.GenerateColorFromID(msg.UserID)

		// Select appropriate color based on terminal theme
		var bgColor usercolor.Color
		if m.isDarkBackground {
			bgColor = adaptiveColor.Dark
		} else {
			bgColor = adaptiveColor.Light
		}

		// Create style with background color and contrasting foreground
		messageStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor.ToHex())).
			Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
			Padding(0, 1)
	} else {
		// No color service - use default style
		messageStyle = lipgloss.NewStyle()
	}

	// Handle multi-line messages by adding proper indentation to each line
	textLines := strings.Split(highlightedText, "\n")
	for i, line := range textLines {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(baseIndent)
		if m.userColorService != nil {
			sb.WriteString(messageStyle.Render(line))
		} else {
			sb.WriteString(line)
		}
	}

	sb.WriteString("\n")
}

// renderThreadMessageLines renders a single message and returns it as a slice of lines.
// indent specifies the number of spaces to indent (0 for parent, 2+ for replies).
// index is the position in the thread (0 for parent, 1+ for replies).
func (m *ThreadViewModel) renderThreadMessageLines(msg message.Message, indent int, index int) []string {
	var lines []string
	// User name style
	userStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2"))

	// Timestamp style
	timestampStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("8"))

	// Thread parent indicator style
	parentStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5"))

	// Reply indicator style
	replyStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("6"))

	// Build header line with explicit width
	var headerLine strings.Builder

	// Selection indicator
	isSelected := m.selectedIndex == index
	if isSelected {
		headerLine.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("> "))
	} else {
		headerLine.WriteString("  ")
	}

	// Indentation
	for i := 0; i < indent; i++ {
		headerLine.WriteString(" ")
	}

	// Thread indicator
	if index == 0 {
		headerLine.WriteString(parentStyle.Render("┌─ "))
	} else {
		headerLine.WriteString(replyStyle.Render("└─ "))
	}

	// User name and timestamp
	headerLine.WriteString(userStyle.Render(msg.UserName))
	headerLine.WriteString(" ")
	headerLine.WriteString(timestampStyle.Render(msg.Timestamp.Format("2006-01-02 15:04:05")))

	lines = append(lines, headerLine.String())

	// Message text (with indentation)
	// Apply highlighting (reuse from MessageViewModel)
	highlightedText := m.highlightText(msg.Text)

	// Calculate base indentation for message text
	baseIndent := "  " + strings.Repeat(" ", indent) + "   "

	// Apply user-specific background color if colorService is available
	var messageStyle lipgloss.Style
	if m.userColorService != nil {
		// Generate color from UserID
		adaptiveColor := m.userColorService.GenerateColorFromID(msg.UserID)

		// Select appropriate color based on terminal theme
		var bgColor usercolor.Color
		if m.isDarkBackground {
			bgColor = adaptiveColor.Dark
		} else {
			bgColor = adaptiveColor.Light
		}

		// Create style with background color and contrasting foreground
		messageStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor.ToHex())).
			Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
			Padding(0, 1)
	} else {
		// No color service - use default style
		messageStyle = lipgloss.NewStyle()
	}

	// Handle multi-line messages by adding proper indentation to each line
	textLines := strings.Split(highlightedText, "\n")
	for _, line := range textLines {
		if m.userColorService != nil {
			lines = append(lines, baseIndent+messageStyle.Render(line))
		} else {
			lines = append(lines, baseIndent+line)
		}
	}

	// Add blank line after message
	lines = append(lines, "")

	return lines
}

// highlightText applies syntax highlighting to message text.
// This includes URL highlighting and mention highlighting.
// (Same implementation as MessageViewModel)
func (m *ThreadViewModel) highlightText(text string) string {
	// First, highlight URLs
	text = m.highlightURLs(text)

	// Then, highlight mentions
	text = m.highlightMentions(text)

	return text
}

// highlightURLs highlights URLs in the text.
func (m *ThreadViewModel) highlightURLs(text string) string {
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

// highlightMentions highlights @mentions in the text.
func (m *ThreadViewModel) highlightMentions(text string) string {
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

// renderThread renders all thread messages as a single string for the viewport.
func (m *ThreadViewModel) renderThread() string {
	var sb strings.Builder

	// Render parent message
	m.renderThreadMessage(&sb, m.parentMessage, 0, 0)
	sb.WriteString("\n")

	// Render replies
	for i, reply := range m.replies {
		m.renderThreadMessage(&sb, reply, 2, i+1)
		sb.WriteString("\n")
	}

	if len(m.replies) == 0 {
		sb.WriteString(lipgloss.NewStyle().Faint(true).Render("  No replies yet"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// getAllThreadLines returns all rendered lines for the thread.
func (m *ThreadViewModel) getAllThreadLines() []string {
	var lines []string

	// Render parent message
	lines = append(lines, m.renderThreadMessageLines(m.parentMessage, 0, 0)...)

	// Render replies
	for i, reply := range m.replies {
		lines = append(lines, m.renderThreadMessageLines(reply, 2, i+1)...)
	}

	if len(m.replies) == 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("  No replies yet"))
	}

	return lines
}

// scrollToSelected scrolls the viewport to ensure the selected message is visible.
func (m *ThreadViewModel) scrollToSelected() {
	selectedLineStart := 0

	var sb strings.Builder

	// Count lines for parent message (index 0)
	if m.selectedIndex > 0 {
		m.renderThreadMessage(&sb, m.parentMessage, 0, 0)
		sb.WriteString("\n")
		selectedLineStart += strings.Count(sb.String(), "\n")
		sb.Reset()
	}

	// Count lines for replies before selected
	for i := 0; i < m.selectedIndex-1 && i < len(m.replies); i++ {
		m.renderThreadMessage(&sb, m.replies[i], 2, i+1)
		sb.WriteString("\n")
		selectedLineStart += strings.Count(sb.String(), "\n")
		sb.Reset()
	}

	viewportHeight := m.viewport.Height
	desiredOffset := selectedLineStart - (viewportHeight / 3)

	if desiredOffset < 0 {
		desiredOffset = 0
	}

	m.viewport.SetYOffset(desiredOffset)
}

// getVisibleLines returns the subset of lines that should be visible based on scroll offset.
func (m *ThreadViewModel) getVisibleLines(allLines []string) []string {
	// Calculate scroll offset to center selected message
	selectedLineStart := 0
	parentLines := m.renderThreadMessageLines(m.parentMessage, 0, 0)

	if m.selectedIndex == 0 {
		selectedLineStart = 0
	} else {
		selectedLineStart = len(parentLines)
		for i := 0; i < m.selectedIndex-1 && i < len(m.replies); i++ {
			replyLines := m.renderThreadMessageLines(m.replies[i], 2, i+1)
			selectedLineStart += len(replyLines)
		}
	}

	// Center the selected message
	offset := selectedLineStart - (m.viewportHeight / 3)
	if offset < 0 {
		offset = 0
	}
	if offset > len(allLines)-m.viewportHeight {
		offset = len(allLines) - m.viewportHeight
		if offset < 0 {
			offset = 0
		}
	}

	// Extract visible range
	endIdx := offset + m.viewportHeight
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}

	return allLines[offset:endIdx]
}

// sendThreadReply sends a reply to the thread.
func (m *ThreadViewModel) sendThreadReply() tea.Cmd {
	text := strings.TrimSpace(m.inputArea.Value())

	// Validate message
	if text == "" {
		return func() tea.Msg {
			return MessageSentMsg{
				Success: false,
				Error:   "Message cannot be empty",
			}
		}
	}

	return func() tea.Msg {
		// If no sender is configured, return an error
		if m.sender == nil {
			return MessageSentMsg{
				Success: false,
				Error:   "Message sender not configured",
			}
		}

		// Send thread reply via the sender
		ctx := context.Background()
		err := m.sender.SendThreadReply(ctx, m.channelID, m.threadTS, text)
		if err != nil {
			return MessageSentMsg{
				Success: false,
				Error:   err.Error(),
			}
		}

		return MessageSentMsg{
			Success:   true,
			ChannelID: m.channelID,
			ThreadTS:  m.threadTS,
			Text:      text,
		}
	}
}

// refreshThreadData fetches the latest thread data from the server.
func (m *ThreadViewModel) refreshThreadData() tea.Cmd {
	return func() tea.Msg {
		// If no sender is configured, skip refresh
		if m.sender == nil {
			return nil
		}

		// Fetch thread data
		ctx := context.Background()
		parent, replies, err := m.sender.GetThreadReplies(ctx, m.channelID, m.threadTS)
		if err != nil {
			// Silently ignore errors during refresh
			return nil
		}

		return ThreadLoadedMsg{
			ChannelID:     m.channelID,
			ThreadTS:      m.threadTS,
			ParentMessage: parent,
			Replies:       replies,
		}
	}
}
