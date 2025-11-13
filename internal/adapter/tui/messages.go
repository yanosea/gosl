package tui

import (
	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
)

type SlackEventMsg interface {
	slackEventMsg()
}

type SlackConnectedMsg struct{}

func (m SlackConnectedMsg) slackEventMsg() {}

type SlackDisconnectedMsg struct{
	Reason string
}

func (m SlackDisconnectedMsg) slackEventMsg() {}

type NewMessageMsg struct {
	ChannelID string
	Message   message.Message
}

func (m NewMessageMsg) slackEventMsg() {}

type ChannelUpdateMsg struct {
	Channel channel.Channel
}

func (m ChannelUpdateMsg) slackEventMsg() {}

type UserTypingMsg struct {
	ChannelID string
	UserID    string
}

func (m UserTypingMsg) slackEventMsg() {}

type ChannelsLoadedMsg struct {
	Channels []channel.Channel
}

type MessagesLoadedMsg struct {
	ChannelID  string
	Messages   []message.Message
	NextCursor string
}

type ThreadLoadedMsg struct {
	ChannelID     string
	ThreadTS      string
	ParentMessage message.Message
	Replies       []message.Message
}

type ThreadRefreshTickMsg struct{}

type MessageRefreshTickMsg struct{}
