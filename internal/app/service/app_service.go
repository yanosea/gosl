package service

import (
	"context"

	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
)

type AppService struct {
	configRepo port.ConfigRepository
	slackRepo  port.SlackRepository
	cache      port.CacheRepository
}

func NewAppService(
	configRepo port.ConfigRepository,
	slackRepo port.SlackRepository,
	cache port.CacheRepository,
) *AppService {
	return &AppService{
		configRepo: configRepo,
		slackRepo:  slackRepo,
		cache:      cache,
	}
}

func (s *AppService) GetChannels(ctx context.Context) ([]channel.Channel, error) {
	return s.slackRepo.GetChannels(ctx)
}

func (s *AppService) GetMessages(ctx context.Context, channelID string, limit int, cursor string) ([]message.Message, string, error) {
	if cursor == "" && s.cache != nil {
		if cachedMsgs, cachedCursor, ok := s.cache.GetMessages(channelID); ok {
			return cachedMsgs, cachedCursor, nil
		}
	}

	messages, nextCursor, err := s.slackRepo.GetMessages(ctx, channelID, limit, cursor)
	if err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		s.cache.StoreMessages(channelID, messages, nextCursor)
	}

	return messages, nextCursor, nil
}

func (s *AppService) GetThreadReplies(ctx context.Context, channelID, threadTS string) (message.Message, []message.Message, error) {
	return s.slackRepo.GetThreadReplies(ctx, channelID, threadTS)
}

func (s *AppService) SendMessage(ctx context.Context, channelID, text string) error {
	msg := message.Message{Text: text}
	if err := msg.Validate(); err != nil {
		return err
	}

	return s.slackRepo.PostMessage(ctx, channelID, text)
}

func (s *AppService) SendThreadReply(ctx context.Context, channelID, threadTS, text string) error {
	msg := message.Message{Text: text}
	if err := msg.Validate(); err != nil {
		return err
	}

	return s.slackRepo.PostThreadReply(ctx, channelID, threadTS, text)
}

func (s *AppService) GetChannelMembers(ctx context.Context, channelID string) ([]user.User, error) {
	return s.slackRepo.GetChannelMembers(ctx, channelID)
}

func (s *AppService) SetCurrentChannel(channelID string) {
	if s.cache != nil {
		s.cache.SetCurrentChannel(channelID)
	}
}
