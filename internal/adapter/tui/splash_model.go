package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConnectionStatus int

const (
	ConnectionStatusConnecting ConnectionStatus = iota
	ConnectionStatusConnected
	ConnectionStatusFailed
)

func (c ConnectionStatus) String() string {
	switch c {
	case ConnectionStatusConnecting:
		return "Connecting"
	case ConnectionStatusConnected:
		return "Connected"
	case ConnectionStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

type SplashModel struct {
	connectionStatus ConnectionStatus
	errorMessage     string
	spinner          spinner.Model
	width            int
	height           int
}

func NewSplashModel() SplashModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return SplashModel{
		connectionStatus: ConnectionStatusConnecting,
		spinner:          s,
	}
}

func (m SplashModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SlackConnectedMsg:
		m.connectionStatus = ConnectionStatusConnected
		return m, nil

	case SlackDisconnectedMsg:
		m.connectionStatus = ConnectionStatusFailed
		m.errorMessage = msg.Reason
		return m, nil

	case ErrorMsg:
		m.connectionStatus = ConnectionStatusFailed
		m.errorMessage = msg.Err
		return m, nil

	default:
		// Handle spinner tick
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m SplashModel) View() string {
	var s string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginTop(2).
		MarginBottom(2)

	s += titleStyle.Render("gosl") + "\n"
	s += lipgloss.NewStyle().Faint(true).Render("Slack TUI Client") + "\n\n"

	switch m.connectionStatus {
	case ConnectionStatusConnecting:
		s += fmt.Sprintf("%s %s\n", m.spinner.View(), "Connecting to Slack...")

	case ConnectionStatusConnected:
		checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s += checkStyle.Render("✓") + " Connected\n"
		s += "\n"
		s += lipgloss.NewStyle().Faint(true).Render("Loading channels...") + "\n"

	case ConnectionStatusFailed:
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		s += errorStyle.Render("✗") + " Failed\n"
		s += "\n"
		if m.errorMessage != "" {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("Error: "+m.errorMessage) + "\n"
		}
		s += "\n"
		s += lipgloss.NewStyle().Faint(true).Render("Press 'q' or Ctrl+C to quit") + "\n"
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		s,
	)
}
