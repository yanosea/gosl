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

	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/textwrap"
	"github.com/yanosea/gosl/internal/domain/usercolor"
)

const (
	viewportHeightReserved = 4
	messageTextIndent      = 5
)

type MessageViewModel struct {
	viewport           viewport.Model
	messages           []message.Message
	selectedIndex      int
	selectedMessageID  string
	channelID          string
	nextCursor         string
	width              int
	height             int
	renderCache        *RenderCache
	stringBuilders     *StringBuilderPool
	messageLineHeights map[string]int // NEW: Line height cache for scroll position calculation
	isInitialized      bool
	inputArea          textarea.Model
	inputFocused       bool
	sender             MessageSender
	isDarkBackground   bool
	userColorService   usercolor.Service
	textWrapper        *textwrap.TextWrapper    // Text wrapping service
	textWrapConfig     *port.TextWrapConfig     // Text wrapping configuration
}

func NewMessageViewModel(channelID string, width, height int) MessageViewModel {
	return NewMessageViewModelWithSender(channelID, width, height, nil)
}

func NewMessageViewModelWithSender(channelID string, width, height int, sender MessageSender) MessageViewModel {
	return NewMessageViewModelWithColorService(channelID, width, height, sender, nil)
}

func NewMessageViewModelWithColorService(channelID string, width, height int, sender MessageSender, colorService usercolor.Service) MessageViewModel {
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
	ta.Placeholder = "Type your message... (Ctrl+Enter to send, Esc to exit input)"
	ta.SetWidth(width - 4)
	ta.SetHeight(inputHeight)
	ta.ShowLineNumbers = false
	ta.Blur() // Start with viewport focused

	// Initialize text wrapper with default configuration (wrapping disabled until config is set)
	defaultConfig := port.DefaultTextWrapConfig()

	return MessageViewModel{
		viewport:           vp,
		messages:           []message.Message{},
		selectedIndex:      0,
		channelID:          channelID,
		nextCursor:         "",
		width:              width,
		height:             height,
		renderCache:        NewRenderCache(),
		stringBuilders:     NewStringBuilderPool(),
		messageLineHeights: make(map[string]int), // Initialize line height cache
		inputArea:          ta,
		inputFocused:       false,
		sender:             sender,
		userColorService:   colorService,
		textWrapper:        textwrap.NewTextWrapper(),
		textWrapConfig:     &defaultConfig,
	}
}

// getCacheKey generates a cache key in the format "messageID-width"
// This ensures cache entries are invalidated when terminal width changes.
func getCacheKey(messageID string, width int) string {
	return fmt.Sprintf("%s-%d", messageID, width)
}

func (m MessageViewModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.startRefreshTimer())
}

// startRefreshTimer returns a command that triggers a periodic refresh.
func (m MessageViewModel) startRefreshTimer() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return MessageRefreshTickMsg{}
	})
}

func (m MessageViewModel) Update(msg tea.Msg) (MessageViewModel, tea.Cmd) {
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
				return m, m.sendMessage()
			}

			// Update input area for other keys
			m.inputArea, cmd = m.inputArea.Update(msg)
			return m, cmd
		}

		// Viewport is focused - handle viewport navigation keys
		switch msg.String() {
		case "esc":
			// Return to channel list (handled by AppModel)
			return m, nil

		case "i", "c":
			// Enter input mode - focus the input area
			m.inputFocused = true
			return m, m.inputArea.Focus()

		case "enter":
			// Open thread view (handled by AppModel)
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
			// Scroll up in viewport
			m.viewport.ViewUp()
			m.scrollToSelected()
			return m, nil

		case "pgdown", "ctrl+d":
			// Scroll down in viewport
			m.viewport.ViewDown()
			m.scrollToSelected()
			return m, nil
		}

	case BackgroundColorMsg:
		// Update background theme
		m.isDarkBackground = msg.IsDark
		return m, nil

	case tea.WindowSizeMsg:
		oldWidth := m.width
		m.width = msg.Width
		m.height = msg.Height

		// Recalculate layout
		inputHeight := 3
		reservedHeight := 9
		viewportHeight := msg.Height - reservedHeight
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight
		m.inputArea.SetWidth(msg.Width - 4)
		m.inputArea.SetHeight(inputHeight)

		// Clear cache if width changed (lipgloss rendering and text wrapping may differ)
		if oldWidth != m.width {
			m.renderCache.InvalidateAll()
			m.messageLineHeights = make(map[string]int)
		}

		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case NewMessageMsg:
		// Add new message if it belongs to this channel
		if msg.ChannelID == m.channelID {
			// Check if cursor was at the latest message before adding new message
			wasAtLatest := false
			if len(m.messages) > 0 {
				wasAtLatest = m.selectedIndex == len(m.messages)-1
			}

			m.AddNewMessage(msg.Message)

			// If cursor was at latest, move cursor to new latest message and scroll
			if wasAtLatest {
				m.selectedIndex = len(m.messages) - 1
				m.selectedMessageID = m.messages[m.selectedIndex].ID
			}

			m.viewport.SetContent(m.renderMessages())

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
			// Immediately refresh messages to show the new message
			return m, m.refreshMessages()
		}
		return m, nil

	case MessageRefreshTickMsg:
		// Refresh messages and restart timer
		return m, tea.Batch(
			m.refreshMessages(),
			m.startRefreshTimer(),
		)

	case MessagesLoadedMsg:
		// Update messages when loaded (from refresh or after sending)
		if msg.ChannelID == m.channelID {
			m.SetMessages(msg.Messages, msg.NextCursor)
			return m, nil
		}
		// If message doesn't match, fall through to default handling
	}

	// Update viewport or input based on focus
	if m.inputFocused {
		m.inputArea, cmd = m.inputArea.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MessageViewModel) View() string {
	var sb strings.Builder

	// Header: Channel indicator
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)
	sb.WriteString(headerStyle.Render(fmt.Sprintf("# %s", m.channelID)))
	sb.WriteString("\n\n")

	// Viewport content
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

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
		footer = "↑/k: Up | ↓/j: Down | g: Top | G: Bottom | Ctrl+U/D: Page | Enter: Thread | i/c: Input | Esc: Back | q: Quit"
	}
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

func (m *MessageViewModel) SetMessages(messages []message.Message, cursor string) {
	prevSelectedMsgID := m.selectedMessageID
	isFirstLoad := m.selectedMessageID == ""
	wasAtLatest := false

	// Check if cursor was at the latest message before update
	if !isFirstLoad && len(m.messages) > 0 {
		wasAtLatest = m.selectedIndex == len(m.messages)-1
	}

	m.messages = messages
	m.nextCursor = cursor

	// Clear line heights cache (messages replaced)
	m.messageLineHeights = make(map[string]int)

	if isFirstLoad {
		if len(m.messages) > 0 {
			m.selectedIndex = len(m.messages) - 1
			m.selectedMessageID = m.messages[m.selectedIndex].ID
		} else {
			m.selectedIndex = 0
			m.selectedMessageID = ""
		}
	} else {
		// If was at latest, always move cursor to new latest message
		if wasAtLatest {
			if len(m.messages) > 0 {
				m.selectedIndex = len(m.messages) - 1
				m.selectedMessageID = m.messages[m.selectedIndex].ID
			} else {
				m.selectedIndex = 0
				m.selectedMessageID = ""
			}
		} else {
			// Try to find the previously selected message
			for i, msg := range m.messages {
				if msg.ID == prevSelectedMsgID {
					m.selectedIndex = i
					break
				}
			}
			// If not found, keep the current selectedIndex (cursor stays in place)
		}
	}

	m.renderCache.InvalidateAll()
	m.viewport.SetContent(m.renderMessages())

	// Scroll to bottom if first load OR was at latest
	if isFirstLoad || wasAtLatest {
		m.viewport.GotoBottom()
	}
}

func (m *MessageViewModel) AppendMessages(messages []message.Message, cursor string) {
	m.messages = append(messages, m.messages...)
	m.nextCursor = cursor
	m.viewport.SetContent(m.renderMessages())
}

func (m *MessageViewModel) AddNewMessage(msg message.Message) {
	// Check if cursor is at the last message before adding new message
	wasAtLatest := len(m.messages) > 0 && m.selectedIndex == len(m.messages)-1

	m.messages = append(m.messages, msg)

	// If cursor was at latest message, move it to the new latest message
	if wasAtLatest {
		m.selectedIndex = len(m.messages) - 1
		m.scrollToSelected()
	}
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
		cacheKey := getCacheKey(msg.ID, m.viewport.Width)
		if cached, found := m.renderCache.Get(cacheKey); found {
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

	// Apply text wrapping after highlighting, before styling
	wrappedText := m.wrapText(highlightedText)

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

	lines := strings.Split(wrappedText, "\n")
	for i, line := range lines {
		if i > 0 {
			msgBuilder.WriteString("\n")
		}
		msgBuilder.WriteString(strings.Repeat(" ", messageTextIndent))
		if m.userColorService != nil {
			msgBuilder.WriteString(messageStyle.Render(line))
		} else {
			msgBuilder.WriteString(line)
		}
	}

	if msg.ReplyCount > 0 {
		msgBuilder.WriteString(" ")
		threadIndicator := fmt.Sprintf("(%d replies)", msg.ReplyCount)
		msgBuilder.WriteString(threadStyle.Render(threadIndicator))
	}

	msgBuilder.WriteString("\n")

	rendered := msgBuilder.String()

	if !selected {
		cacheKey := getCacheKey(msg.ID, m.viewport.Width)
		m.renderCache.Set(cacheKey, rendered)

		// Update messageLineHeights cache
		lineCount := strings.Count(rendered, "\n")
		m.messageLineHeights[cacheKey] = lineCount
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

// wrapText wraps text according to terminal width and configuration.
// It calculates the available width by subtracting message text indent and user color padding
// from the viewport width.
func (m *MessageViewModel) wrapText(text string) string {
	// Return text as-is if text wrapper or config is not available
	if m.textWrapper == nil || m.textWrapConfig == nil {
		return text
	}

	// Return text as-is if wrapping is disabled
	if !m.textWrapConfig.Enabled {
		return text
	}

	// Calculate available width for text
	// Subtract messageTextIndent (5 spaces) and user color padding if applicable
	availableWidth := m.viewport.Width - messageTextIndent

	// If user color service is enabled, subtract additional padding (2 spaces for left+right padding)
	if m.userColorService != nil {
		availableWidth -= 2 // Padding(0, 1) adds 1 space on each side
	}

	// Ensure minimum width
	if availableWidth < 10 {
		// If width is too small, return text as-is to avoid errors
		return text
	}

	// Use MaxLineWidth from config if set, otherwise use available width
	wrapWidth := availableWidth
	if m.textWrapConfig.MaxLineWidth > 0 && m.textWrapConfig.MaxLineWidth < availableWidth {
		wrapWidth = m.textWrapConfig.MaxLineWidth
	}

	// Convert config to domain TextWrapOptions
	opts := m.textWrapConfig.ToOptions()

	// Wrap text using the TextWrapper
	wrapped, err := m.textWrapper.WrapText(text, wrapWidth, opts)
	if err != nil {
		// On error, log and return original text (graceful degradation)
		// TODO: Add proper error logging when logger is available
		return text
	}

	return wrapped
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
	// 1. Early return for empty messages
	if len(m.messages) == 0 {
		return
	}

	// 2. Get or calculate selected message height
	selectedCacheKey := getCacheKey(m.selectedMessageID, m.viewport.Width)
	selectedMessageHeight, ok := m.messageLineHeights[selectedCacheKey]
	if !ok {
		// Cache miss: render message and count lines
		sb := m.stringBuilders.Get()
		defer m.stringBuilders.Put(sb)
		m.renderMessage(sb, m.messages[m.selectedIndex], true)
		sb.WriteString("\n")
		rendered := sb.String()
		selectedMessageHeight = strings.Count(rendered, "\n")
		m.messageLineHeights[selectedCacheKey] = selectedMessageHeight
	}

	// 3. Calculate selected message line start position
	selectedLineStart := 0
	for i := 0; i < m.selectedIndex; i++ {
		msgCacheKey := getCacheKey(m.messages[i].ID, m.viewport.Width)
		msgHeight, ok := m.messageLineHeights[msgCacheKey]
		if !ok {
			// Cache miss: render and calculate line height
			msgSB := m.stringBuilders.Get()
			m.renderMessage(msgSB, m.messages[i], false)
			msgSB.WriteString("\n")
			rendered := msgSB.String()
			msgHeight = strings.Count(rendered, "\n")
			m.messageLineHeights[msgCacheKey] = msgHeight
			m.stringBuilders.Put(msgSB)
		}
		selectedLineStart += msgHeight
	}

	// 4. Check if cursor is already visible
	currentOffset := m.viewport.YOffset
	viewportHeight := m.viewport.Height

	if selectedLineStart >= currentOffset &&
		selectedLineStart+selectedMessageHeight <= currentOffset+viewportHeight {
		return // No scroll needed
	}

	// 5. Calculate desired scroll offset based on message height
	var desiredOffset int
	if selectedMessageHeight <= viewportHeight {
		// Message fits in viewport - ensure entire message is visible
		if selectedLineStart+selectedMessageHeight > currentOffset+viewportHeight {
			// Message bottom is below viewport bottom
			desiredOffset = selectedLineStart + selectedMessageHeight - viewportHeight
		} else {
			// Message top is above viewport top
			desiredOffset = selectedLineStart
		}
	} else {
		// Message is taller than viewport - align to top
		desiredOffset = selectedLineStart
	}

	// 6. Apply bounds checking
	if desiredOffset < 0 {
		desiredOffset = 0
	}

	// 7. Set scroll offset
	m.viewport.SetYOffset(desiredOffset)
}

// sendMessage sends a message to the channel.
func (m *MessageViewModel) sendMessage() tea.Cmd {
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

		// Send message via the sender
		ctx := context.Background()
		err := m.sender.SendMessage(ctx, m.channelID, text)
		if err != nil {
			return MessageSentMsg{
				Success: false,
				Error:   err.Error(),
			}
		}

		return MessageSentMsg{
			Success:   true,
			ChannelID: m.channelID,
			Text:      text,
		}
	}
}

// refreshMessages fetches the latest messages from the server.
func (m *MessageViewModel) refreshMessages() tea.Cmd {
	return func() tea.Msg {
		// If no sender is configured, skip refresh
		if m.sender == nil {
			return nil
		}

		// Fetch messages data
		ctx := context.Background()
		messages, nextCursor, err := m.sender.GetMessages(ctx, m.channelID, 50, "")
		if err != nil {
			// Silently ignore errors during refresh
			return nil
		}

		return MessagesLoadedMsg{
			ChannelID:  m.channelID,
			Messages:   messages,
			NextCursor: nextCursor,
		}
	}
}
