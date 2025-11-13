package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
)

// MessageSender defines the interface for sending messages.
type MessageSender interface {
	SendMessage(ctx context.Context, channelID, text string) error
	SendThreadReply(ctx context.Context, channelID, threadTS, text string) error
	GetChannelMembers(ctx context.Context, channelID string) ([]user.User, error)
	GetThreadReplies(ctx context.Context, channelID, threadTS string) (parent message.Message, replies []message.Message, err error)
	GetMessages(ctx context.Context, channelID string, limit int, cursor string) (messages []message.Message, nextCursor string, err error)
}

// InputMode represents the mode of message input.
type InputMode int

const (
	// InputModeChannelMessage is for sending a message to a channel.
	InputModeChannelMessage InputMode = iota
	// InputModeThreadReply is for replying to a thread.
	InputModeThreadReply
)

// String returns the string representation of InputMode.
func (m InputMode) String() string {
	switch m {
	case InputModeChannelMessage:
		return "ChannelMessage"
	case InputModeThreadReply:
		return "ThreadReply"
	default:
		return "Unknown"
	}
}

// MessageInputModel manages the message input view.
type MessageInputModel struct {
	textarea           textarea.Model
	mode               InputMode
	channelID          string
	threadTS           string
	parentMessageText  string
	errorMessage       string
	hasError           bool
	width              int
	height             int
	sender             MessageSender
	mentionSuggestions []user.User
	showingSuggestions bool
	selectedSuggestion int
}

// NewMessageInputModel creates a new MessageInputModel instance without a sender.
// This is kept for backward compatibility; prefer NewMessageInputModelWithSender.
func NewMessageInputModel(mode InputMode, channelID, threadTS string, width, height int) MessageInputModel {
	return NewMessageInputModelWithSender(mode, channelID, threadTS, width, height, nil)
}

// NewMessageInputModelWithSender creates a new MessageInputModel instance with a MessageSender.
func NewMessageInputModelWithSender(mode InputMode, channelID, threadTS string, width, height int, sender MessageSender) MessageInputModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Ctrl+Enter to send, Esc to cancel)"
	ta.Focus()

	// Set textarea dimensions
	ta.SetWidth(width - 4)
	ta.SetHeight(height - 10) // Reserve space for header, footer, and error messages

	// Enable line numbers (optional)
	ta.ShowLineNumbers = false

	return MessageInputModel{
		textarea:           ta,
		mode:               mode,
		channelID:          channelID,
		threadTS:           threadTS,
		parentMessageText:  "",
		errorMessage:       "",
		hasError:           false,
		width:              width,
		height:             height,
		sender:             sender,
		mentionSuggestions: nil,
		showingSuggestions: false,
		selectedSuggestion: 0,
	}
}

// Init initializes the MessageInputModel.
func (m MessageInputModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and updates the MessageInputModel.
func (m MessageInputModel) Update(msg tea.Msg) (MessageInputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle mention suggestion navigation
		if m.showingSuggestions {
			switch msg.Type {
			case tea.KeyUp:
				if m.selectedSuggestion > 0 {
					m.selectedSuggestion--
				}
				return m, nil
			case tea.KeyDown:
				if m.selectedSuggestion < len(m.mentionSuggestions)-1 {
					m.selectedSuggestion++
				}
				return m, nil
			case tea.KeyTab, tea.KeyEnter:
				// Select the highlighted suggestion
				m.selectMentionSuggestion()
				return m, nil
			case tea.KeyEsc:
				// Close suggestions
				m.showingSuggestions = false
				return m, nil
			}
		}

		// Check for Ctrl+Enter or Alt+Enter to send message
		// This must be checked BEFORE textarea.Update to prevent textarea from consuming the event
		// Note: Many terminals send Ctrl+J when Ctrl+Enter is pressed
		if msg.String() == "ctrl+enter" || msg.String() == "alt+enter" || msg.String() == "ctrl+j" {
			return m, m.sendMessage()
		}

		// Alternative check: Enter key with Ctrl modifier
		if msg.Type == tea.KeyEnter && msg.String() == "ctrl+enter" {
			return m, m.sendMessage()
		}

		// Check for @ to trigger mention autocomplete
		if msg.String() == "@" {
			// Let textarea handle the @ character first
			m.textarea, cmd = m.textarea.Update(msg)
			// Then fetch suggestions
			return m, tea.Batch(cmd, m.fetchMentionSuggestions())
		}

		switch msg.Type {
		case tea.KeyEsc:
			// Cancel and return to previous view (handled by AppModel)
			return m, nil

		case tea.KeyCtrlC:
			// Quit application (handled globally)
			return m, tea.Quit
		}

	case MentionSuggestionsMsg:
		if msg.Error != nil {
			// Silently ignore errors for mention suggestions
			return m, nil
		}
		if len(msg.Users) > 0 {
			m.mentionSuggestions = msg.Users
			m.showingSuggestions = true
			m.selectedSuggestion = 0
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)
		return m, nil

	case MessageSentMsg:
		if msg.Success {
			// Clear textarea and error on success
			m.textarea.Reset()
			m.ClearError()
		} else {
			// Set error message on failure
			m.SetError(msg.Error)
		}
		// Don't return early - let textarea handle the update below
	}

	// Update textarea for all message types
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the MessageInputModel.
func (m MessageInputModel) View() string {
	var sb strings.Builder

	// Header: Mode indicator
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)

	header := ""
	if m.mode == InputModeThreadReply {
		header = "📝 Reply to Thread"
		// Show parent message if set
		if m.parentMessageText != "" {
			parentStyle := lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color("8")).
				Padding(0, 1)
			parentPreview := m.parentMessageText
			if len(parentPreview) > 60 {
				parentPreview = parentPreview[:57] + "..."
			}
			sb.WriteString(parentStyle.Render("Replying to: " + parentPreview))
			sb.WriteString("\n")
		}
	} else {
		header = "📝 New Message"
	}
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n\n")

	// Error message (if any)
	if m.hasError {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true).
			Padding(0, 1)
		sb.WriteString(errorStyle.Render("❌ " + m.errorMessage))
		sb.WriteString("\n\n")
	}

	// Textarea
	sb.WriteString(m.textarea.View())
	sb.WriteString("\n")

	// Mention suggestions (if showing)
	if m.showingSuggestions && len(m.mentionSuggestions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(m.renderMentionSuggestions())
		sb.WriteString("\n")
	}

	// Footer: Help text
	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Padding(1, 1)
	footer := "Ctrl+Enter: Send | Enter: New line | Esc: Cancel"
	if m.showingSuggestions {
		footer = "↑/↓: Navigate | Tab/Enter: Select | Esc: Close"
	}
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

// renderMentionSuggestions renders the mention autocomplete suggestions.
func (m *MessageInputModel) renderMentionSuggestions() string {
	var sb strings.Builder

	suggestionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Padding(0, 1)

	sb.WriteString(headerStyle.Render("💬 Mention Suggestions:"))
	sb.WriteString("\n")

	// Show up to 10 suggestions
	maxShow := 10
	if len(m.mentionSuggestions) < maxShow {
		maxShow = len(m.mentionSuggestions)
	}

	for i := 0; i < maxShow; i++ {
		u := m.mentionSuggestions[i]
		displayText := fmt.Sprintf("@%s (%s)", u.Name, u.DisplayName)

		if i == m.selectedSuggestion {
			sb.WriteString(selectedStyle.Render(displayText))
		} else {
			sb.WriteString(suggestionStyle.Render(displayText))
		}
		sb.WriteString("\n")
	}

	if len(m.mentionSuggestions) > maxShow {
		moreStyle := lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)
		sb.WriteString(moreStyle.Render(fmt.Sprintf("... and %d more", len(m.mentionSuggestions)-maxShow)))
	}

	return sb.String()
}

// SetValue sets the textarea value.
func (m *MessageInputModel) SetValue(value string) {
	m.textarea.SetValue(value)
}

// GetValue returns the textarea value.
func (m *MessageInputModel) GetValue() string {
	return m.textarea.Value()
}

// Clear clears the textarea.
func (m *MessageInputModel) Clear() {
	m.textarea.Reset()
}

// SetError sets an error message.
func (m *MessageInputModel) SetError(err string) {
	m.errorMessage = err
	m.hasError = true
}

// ClearError clears the error message.
func (m *MessageInputModel) ClearError() {
	m.errorMessage = ""
	m.hasError = false
}

// Focus focuses the textarea and returns the focus command for cursor blinking.
func (m *MessageInputModel) Focus() tea.Cmd {
	return m.textarea.Focus()
}

// Blur blurs the textarea.
func (m *MessageInputModel) Blur() {
	m.textarea.Blur()
}

// SetParentMessageText sets the parent message text (for thread replies).
func (m *MessageInputModel) SetParentMessageText(text string) {
	m.parentMessageText = text
}

// sendMessage creates a command to send the message.
func (m *MessageInputModel) sendMessage() tea.Cmd {
	text := strings.TrimSpace(m.textarea.Value())

	// Validate message
	if text == "" {
		return func() tea.Msg {
			return MessageSentMsg{
				Success: false,
				Error:   "Message cannot be empty",
			}
		}
	}

	// Create appropriate send command based on mode
	if m.mode == InputModeThreadReply {
		return m.sendThreadReply(text)
	}
	return m.sendChannelMessage(text)
}

// sendChannelMessage sends a message to the channel.
func (m *MessageInputModel) sendChannelMessage(text string) tea.Cmd {
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

// sendThreadReply sends a reply to a thread.
func (m *MessageInputModel) sendThreadReply(text string) tea.Cmd {
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

// MessageSentMsg is sent when a message is sent (success or failure).
type MessageSentMsg struct {
	Success   bool
	Error     string
	ChannelID string
	ThreadTS  string
	Text      string
}

func (m MessageSentMsg) String() string {
	if m.Success {
		return fmt.Sprintf("Message sent successfully to channel %s", m.ChannelID)
	}
	return fmt.Sprintf("Failed to send message: %s", m.Error)
}

// MentionSuggestionsMsg is sent when mention suggestions are fetched.
type MentionSuggestionsMsg struct {
	Users []user.User
	Error error
}

// fetchMentionSuggestions fetches channel members for mention autocomplete.
func (m *MessageInputModel) fetchMentionSuggestions() tea.Cmd {
	return func() tea.Msg {
		if m.sender == nil {
			return MentionSuggestionsMsg{
				Users: nil,
				Error: errors.New("sender not configured"),
			}
		}

		ctx := context.Background()
		users, err := m.sender.GetChannelMembers(ctx, m.channelID)
		if err != nil {
			return MentionSuggestionsMsg{
				Users: nil,
				Error: err,
			}
		}

		// Limit to 50 suggestions
		if len(users) > 50 {
			users = users[:50]
		}

		return MentionSuggestionsMsg{
			Users: users,
			Error: nil,
		}
	}
}

// filterMentionSuggestions filters mention suggestions based on query.
func (m *MessageInputModel) filterMentionSuggestions(query string) []user.User {
	if query == "" {
		return m.mentionSuggestions
	}

	query = strings.ToLower(query)
	filtered := make([]user.User, 0)

	for _, u := range m.mentionSuggestions {
		nameLower := strings.ToLower(u.Name)
		displayNameLower := strings.ToLower(u.DisplayName)
		
		if strings.Contains(nameLower, query) || strings.Contains(displayNameLower, query) {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

// selectMentionSuggestion inserts the selected mention into the textarea.
func (m *MessageInputModel) selectMentionSuggestion() {
	if m.selectedSuggestion < 0 || m.selectedSuggestion >= len(m.mentionSuggestions) {
		return
	}

	selected := m.mentionSuggestions[m.selectedSuggestion]
	currentValue := m.textarea.Value()

	// Find the last @ symbol
	lastAtIndex := strings.LastIndex(currentValue, "@")
	if lastAtIndex == -1 {
		return
	}

	// Replace from @ to cursor position with the selected username
	newValue := currentValue[:lastAtIndex] + "@" + selected.Name + " "
	m.textarea.SetValue(newValue)

	// Clear suggestions
	m.showingSuggestions = false
	m.mentionSuggestions = nil
	m.selectedSuggestion = 0
}
