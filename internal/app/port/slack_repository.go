package port

import (
	"context"
	"errors"

	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
)

var (
	ErrInvalidToken      = errors.New("invalid Slack API token")
	ErrRateLimitExceeded = errors.New("Slack API rate limit exceeded")
	ErrNetworkError      = errors.New("network error connecting to Slack")
	ErrChannelNotFound   = errors.New("channel not found")
)

type SlackEvent interface {
	SlackEvent()
}

type SlackConnectedEvent struct{}

func (e SlackConnectedEvent) SlackEvent() {}

type SlackDisconnectedEvent struct {
	Reason string
}

func (e SlackDisconnectedEvent) SlackEvent() {}

type NewMessageEvent struct {
	ChannelID string
	Message   message.Message
}

func (e NewMessageEvent) SlackEvent() {}

type ChannelUpdateEvent struct {
	Channel channel.Channel
}

func (e ChannelUpdateEvent) SlackEvent() {}

type UserTypingEvent struct {
	ChannelID string
	UserID    string
}

func (e UserTypingEvent) SlackEvent() {}

type SlackRepository interface {
	Connect(ctx context.Context, botToken, appToken string) error
	Disconnect() error
	GetChannels(ctx context.Context) ([]channel.Channel, error)
	GetMessages(ctx context.Context, channelID string, limit int, cursor string) (messages []message.Message, nextCursor string, err error)
	GetThreadReplies(ctx context.Context, channelID, threadTS string) (parent message.Message, replies []message.Message, err error)
	PostMessage(ctx context.Context, channelID, text string) error
	PostThreadReply(ctx context.Context, channelID, threadTS, text string) error
	GetChannelMembers(ctx context.Context, channelID string) ([]user.User, error)
	SubscribeEvents(ctx context.Context) (<-chan SlackEvent, error)
}
