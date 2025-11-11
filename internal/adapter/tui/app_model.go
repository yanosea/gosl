package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/app/service"
)

type AppState int

const (
	StateSplash AppState = iota
	StateChannelList
	StateMessageView
	StateThreadView
	StateMessageInput
	StateHelp
	StateError
)

func (s AppState) String() string {
	switch s {
	case StateSplash:
		return "Splash"
	case StateChannelList:
		return "ChannelList"
	case StateMessageView:
		return "MessageView"
	case StateThreadView:
		return "ThreadView"
	case StateMessageInput:
		return "MessageInput"
	case StateHelp:
		return "Help"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type AppModel struct {
	state          AppState
	appService     *service.AppService
	config         *port.Config
	width          int
	height         int
	autoNavigated  bool

	splash       SplashModel
	channelList  ChannelListModel
	messageView  MessageViewModel
	threadView   ThreadViewModel
	messageInput MessageInputModel
	helpView     HelpModel
	errorMessage string
}

func NewAppModel(appService *service.AppService, config *port.Config) AppModel {
	return AppModel{
		state:        StateSplash,
		appService:   appService,
		config:       config,
		splash:       NewSplashModel(),
		channelList:  NewChannelListModel(80, 24),
		messageView:  NewMessageViewModel("", 80, 24),
		threadView:   NewThreadViewModel("", "", 80, 24),
		messageInput: NewMessageInputModelWithSender(InputModeChannelMessage, "", "", 80, 24, appService),
		helpView:     NewHelpModel(80, 24),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.splash.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.state = StateHelp
			return m, nil
		}
	}

	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowMsg.Width
		m.height = windowMsg.Height
		switch m.state {
		case StateSplash:
			m.splash.width = windowMsg.Width
			m.splash.height = windowMsg.Height
		case StateChannelList:
			m.channelList.width = windowMsg.Width
			m.channelList.height = windowMsg.Height
			m.channelList.list.SetSize(windowMsg.Width, windowMsg.Height)
		case StateMessageView:
			m.messageView, _ = m.messageView.Update(windowMsg)
		case StateThreadView:
			m.threadView, _ = m.threadView.Update(windowMsg)
		case StateMessageInput:
			m.messageInput, _ = m.messageInput.Update(windowMsg)
		case StateHelp:
			m.helpView, _ = m.helpView.Update(windowMsg)
		}
	}

	switch evt := msg.(type) {
	case SlackConnectedMsg:
		if m.state == StateSplash {
			return m, loadChannelsCmd(m.appService)
		}
		return m, nil

	case ChannelsLoadedMsg:
		m.channelList.SetChannels(evt.Channels)
		m.state = StateChannelList

		if !m.autoNavigated && m.config != nil && m.config.DefaultChannel != "" {
			m.autoNavigated = true
			for _, ch := range evt.Channels {
				if ch.Name == m.config.DefaultChannel {
					m.messageView = NewMessageViewModel(ch.ID, m.width, m.height)
					m.state = StateMessageView
					return m, loadMessagesCmd(m.appService, ch.ID, 50, "")
				}
			}
		}
		return m, nil

	case MessagesLoadedMsg:
		if m.state == StateMessageView && evt.ChannelID == m.messageView.channelID {
			m.messageView.SetMessages(evt.Messages, evt.NextCursor)
		}
		return m, nil

	case ThreadLoadedMsg:
		break

	case SlackDisconnectedMsg:
		m.state = StateError
		m.errorMessage = "Disconnected: " + evt.Reason
		return m, nil

	case ErrorMsg:
		m.state = StateError
		m.errorMessage = evt.Err
		return m, nil
	}

	switch m.state {
	case StateSplash:
		return m.updateSplash(msg)
	case StateChannelList:
		return m.updateChannelList(msg)
	case StateMessageView:
		return m.updateMessageView(msg)
	case StateThreadView:
		return m.updateThreadView(msg)
	case StateMessageInput:
		return m.updateMessageInput(msg)
	case StateHelp:
		return m.updateHelp(msg)
	case StateError:
		return m.updateError(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.state {
	case StateSplash:
		return m.viewSplash()
	case StateChannelList:
		return m.viewChannelList()
	case StateMessageView:
		return m.viewMessageView()
	case StateThreadView:
		return m.viewThreadView()
	case StateMessageInput:
		return m.viewMessageInput()
	case StateHelp:
		return m.viewHelp()
	case StateError:
		return m.viewError()
	}

	return "Unknown state"
}

func (m AppModel) updateSplash(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.splash, cmd = m.splash.Update(msg)
	return m, cmd
}

func (m AppModel) updateChannelList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEnter && !m.channelList.searchMode {
			selectedChannel := m.channelList.GetSelectedChannel()
			if selectedChannel != nil {
				m.messageView = NewMessageViewModel(selectedChannel.ID, m.width, m.height)
				m.state = StateMessageView
				return m, loadMessagesCmd(m.appService, selectedChannel.ID, 5, "")
			}
		}
	}

	var cmd tea.Cmd
	m.channelList, cmd = m.channelList.Update(msg)
	return m, cmd
}

func (m AppModel) updateMessageView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			m.state = StateChannelList
			return m, nil
		}
		if keyMsg.Type == tea.KeyEnter {
			if m.messageView.selectedIndex < len(m.messageView.messages) {
				selectedMsg := m.messageView.messages[m.messageView.selectedIndex]
				if selectedMsg.ReplyCount > 0 || selectedMsg.ThreadTS != "" {
					threadTS := selectedMsg.ThreadTS
					if threadTS == "" {
						threadTS = selectedMsg.ID
					}
					m.threadView = NewThreadViewModelWithSender(m.messageView.channelID, threadTS, m.width, m.height, m.appService)
					m.state = StateThreadView
					initCmd := m.threadView.Init()
					loadCmd := loadThreadCmd(m.appService, m.messageView.channelID, threadTS)
					return m, tea.Batch(initCmd, loadCmd)
				}
			}
			return m, nil
		}
		if keyMsg.String() == "i" || keyMsg.String() == "c" {
			m.messageInput = NewMessageInputModelWithSender(InputModeChannelMessage, m.messageView.channelID, "", m.width, m.height, m.appService)
			m.state = StateMessageInput
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.messageView, cmd = m.messageView.Update(msg)
	return m, cmd
}

func (m AppModel) updateThreadView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc && !m.threadView.inputFocused {
			m.state = StateMessageView
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.threadView, cmd = m.threadView.Update(msg)
	return m, cmd
}

func (m AppModel) updateMessageInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			if m.messageInput.mode == InputModeThreadReply {
				m.state = StateThreadView
				return m, loadThreadCmd(m.appService, m.threadView.channelID, m.threadView.threadTS)
			} else {
				m.state = StateMessageView
				return m, loadMessagesCmd(m.appService, m.messageView.channelID, 50, "")
			}
		}
	}

	if sentMsg, ok := msg.(MessageSentMsg); ok {
		if sentMsg.Success {
		}
	}

	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

func (m AppModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			m.state = StateChannelList
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.helpView, cmd = m.helpView.Update(msg)
	return m, cmd
}

func (m AppModel) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			m.state = StateChannelList
			return m, nil
		}
	}
	return m, nil
}

func (m AppModel) viewSplash() string {
	return m.splash.View()
}

func (m AppModel) viewChannelList() string {
	return m.channelList.View()
}

func (m AppModel) viewMessageView() string {
	return m.messageView.View()
}

func (m AppModel) viewThreadView() string {
	return m.threadView.View()
}

func (m AppModel) viewMessageInput() string {
	return m.messageInput.View()
}

func (m AppModel) viewHelp() string {
	return m.helpView.View()
}

func (m AppModel) viewError() string {
	return "Error: " + m.errorMessage + "\nPress Esc to return"
}

type ErrorMsg struct {
	Err string
}

func loadChannelsCmd(appService *service.AppService) tea.Cmd {
	return func() tea.Msg {
		channels, err := appService.GetChannels(context.Background())
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("Failed to load channels: %v", err)}
		}
		return ChannelsLoadedMsg{Channels: channels}
	}
}

func loadMessagesCmd(appService *service.AppService, channelID string, limit int, cursor string) tea.Cmd {
	return func() tea.Msg {
		messages, nextCursor, err := appService.GetMessages(context.Background(), channelID, limit, cursor)
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("Failed to load messages: %v", err)}
		}
		return MessagesLoadedMsg{
			ChannelID:  channelID,
			Messages:   messages,
			NextCursor: nextCursor,
		}
	}
}

func loadThreadCmd(appService *service.AppService, channelID, threadTS string) tea.Cmd {
	return func() tea.Msg {
		parent, replies, err := appService.GetThreadReplies(context.Background(), channelID, threadTS)
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("Failed to load thread: %v", err)}
		}
		return ThreadLoadedMsg{
			ChannelID:     channelID,
			ThreadTS:      threadTS,
			ParentMessage: parent,
			Replies:       replies,
		}
	}
}
