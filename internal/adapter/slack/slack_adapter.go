package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"

	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/logger"
	"github.com/yanosea/gosl/internal/domain/message"
	"github.com/yanosea/gosl/internal/domain/user"
)

var _ port.SlackRepository = (*SlackAdapter)(nil)

const (
	maxUserCacheSize      = 1000
	batchFetchConcurrency = 10
)

type SlackAdapter struct {
	client       *slack.Client
	socketClient *socketmode.Client
	connected    bool
	mu           sync.RWMutex
	eventChan    chan port.SlackEvent
	ctx          context.Context
	cancel       context.CancelFunc
	userCache    *lru.Cache[string, *slack.User]
	logger       *logger.Logger
	retryConfig  RetryConfig
}

func NewSlackAdapter() *SlackAdapter {
	cache, err := lru.New[string, *slack.User](maxUserCacheSize)
	if err != nil {
		panic(fmt.Sprintf("failed to create user cache: %v", err))
	}

	loggerInstance, err := logger.NewLogger(logger.Config{
		OutputPath: logger.GetLogFilePath(),
		Format:     logger.FormatJSON,
		AddSource:  false,
	})
	if err != nil {
		loggerInstance, _ = logger.NewLogger(logger.Config{
			OutputPath: "",
			Format:     logger.FormatText,
			AddSource:  false,
		})
	}

	return &SlackAdapter{
		userCache: cache,
		logger:    loggerInstance,
		retryConfig: RetryConfig{
			MaxRetries: DefaultMaxRetries,
			MaxBackoff: DefaultMaxBackoff,
		},
	}
}

func parseSlackTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}

	seconds, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Now()
	}

	sec := int64(seconds)
	nsec := int64((seconds - float64(sec)) * 1e9)

	return time.Unix(sec, nsec)
}

func (a *SlackAdapter) Connect(ctx context.Context, botToken, appToken string) error {
	if botToken == "" || appToken == "" {
		return port.ErrInvalidToken
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.client = slack.New(
		botToken,
		slack.OptionDebug(false),
		slack.OptionLog(nil),
		slack.OptionAppLevelToken(appToken),
	)

	a.socketClient = socketmode.New(
		a.client,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(nil),
	)

	authTest, err := a.client.AuthTestContext(ctx)
	if err != nil {
		if err.Error() == "invalid_auth" {
			return port.ErrInvalidToken
		}
		return fmt.Errorf("%w: %v", port.ErrNetworkError, err)
	}

	if authTest.UserID == "" {
		return port.ErrInvalidToken
	}

	a.connected = true

	a.ctx, a.cancel = context.WithCancel(ctx)

	return nil
}

func (a *SlackAdapter) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.connected {
		return nil
	}

	if a.cancel != nil {
		a.cancel()
	}

	if a.eventChan != nil {
		close(a.eventChan)
		a.eventChan = nil
	}

	a.connected = false
	a.client = nil
	a.socketClient = nil

	return nil
}

func (a *SlackAdapter) GetChannels(ctx context.Context) ([]channel.Channel, error) {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return nil, errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	var allChannels []channel.Channel
	cursor := ""

	for {
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Limit:           1000,
			Types:           []string{"public_channel", "private_channel", "im", "mpim"},
		}

		result, err := withRetry(ctx, a.retryConfig, func() (struct {
			conversations []slack.Channel
			nextCursor    string
		}, error) {
			conversations, nextCursor, err := client.GetConversationsContext(ctx, params)
			return struct {
				conversations []slack.Channel
				nextCursor    string
			}{conversations, nextCursor}, err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get conversations: %w", err)
		}

		for _, conv := range result.conversations {
			ch := convertToChannel(conv)
			allChannels = append(allChannels, ch)
		}

		cursor = result.nextCursor
		if cursor == "" {
			break
		}
	}

	return allChannels, nil
}

func (a *SlackAdapter) GetMessages(ctx context.Context, channelID string, limit int, cursor string) ([]message.Message, string, error) {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return nil, "", errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Cursor:    cursor,
		Limit:     limit,
	}

	history, err := withRetry(ctx, a.retryConfig, func() (*slack.GetConversationHistoryResponse, error) {
		return client.GetConversationHistoryContext(ctx, params)
	})
	if err != nil {
		if err.Error() == "channel_not_found" {
			return nil, "", port.ErrChannelNotFound
		}
		return nil, "", fmt.Errorf("failed to get messages: %w", err)
	}

	userIDSet := make(map[string]struct{}, len(history.Messages))
	for i := range history.Messages {
		if history.Messages[i].User != "" {
			userIDSet[history.Messages[i].User] = struct{}{}
		}
	}

	// Convert set to slice
	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	userInfoMap := a.batchGetUserInfo(ctx, userIDs)

	messages := make([]message.Message, len(history.Messages))
	for i := len(history.Messages) - 1; i >= 0; i-- {
		msg := history.Messages[i]
		userName := msg.User
		if userInfo, ok := userInfoMap[msg.User]; ok {
			userName = userInfo.Name
		}

		ts := parseSlackTimestamp(msg.Timestamp)

		replyCount := 0
		if msg.ThreadTimestamp != "" && msg.ReplyCount > 0 {
			replyCount = msg.ReplyCount
		}

		idx := len(history.Messages) - 1 - i
		messages[idx] = message.Message{
			ID:         msg.Timestamp,
			ChannelID:  channelID,
			UserID:     msg.User,
			UserName:   userName,
			Text:       msg.Text,
			Timestamp:  ts,
			ThreadTS:   msg.ThreadTimestamp,
			ReplyCount: replyCount,
		}
	}

	return messages, history.ResponseMetaData.NextCursor, nil
}

func (a *SlackAdapter) GetThreadReplies(ctx context.Context, channelID, threadTS string) (message.Message, []message.Message, error) {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return message.Message{}, nil, errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	result, err := withRetry(ctx, a.retryConfig, func() (struct {
		msgs []slack.Message
	}, error) {
		msgs, _, _, err := client.GetConversationRepliesContext(ctx, params)
		return struct{ msgs []slack.Message }{msgs}, err
	})
	if err != nil {
		return message.Message{}, nil, fmt.Errorf("failed to get thread replies: %w", err)
	}

	msgs := result.msgs

	if len(msgs) == 0 {
		return message.Message{}, nil, errors.New("thread not found")
	}

	userIDSet := make(map[string]struct{}, len(msgs))
	for i := range msgs {
		if msgs[i].User != "" {
			userIDSet[msgs[i].User] = struct{}{}
		}
	}

	// Convert set to slice
	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	userInfoMap := a.batchGetUserInfo(ctx, userIDs)

	parentSlackMsg := msgs[0]
	userName := parentSlackMsg.User
	if userInfo, ok := userInfoMap[parentSlackMsg.User]; ok {
		userName = userInfo.Name
	}

	ts := parseSlackTimestamp(parentSlackMsg.Timestamp)

	parent := message.Message{
		ID:         parentSlackMsg.Timestamp,
		ChannelID:  channelID,
		UserID:     parentSlackMsg.User,
		UserName:   userName,
		Text:       parentSlackMsg.Text,
		Timestamp:  ts,
		ThreadTS:   threadTS,
		ReplyCount: len(msgs) - 1,
	}

	replies := make([]message.Message, len(msgs)-1)
	for i := 1; i < len(msgs); i++ {
		replyMsg := msgs[i]
		replyUserName := replyMsg.User
		if userInfo, ok := userInfoMap[replyMsg.User]; ok {
			replyUserName = userInfo.Name
		}

		replyTS := parseSlackTimestamp(replyMsg.Timestamp)

		replies[i-1] = message.Message{
			ID:        replyMsg.Timestamp,
			ChannelID: channelID,
			UserID:    replyMsg.User,
			UserName:  replyUserName,
			Text:      replyMsg.Text,
			Timestamp: replyTS,
			ThreadTS:  threadTS,
		}
	}

	return parent, replies, nil
}

func (a *SlackAdapter) PostMessage(ctx context.Context, channelID, text string) error {
	msg := message.Message{Text: text}
	if err := msg.Validate(); err != nil {
		return err
	}

	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	_, err := withRetry(ctx, a.retryConfig, func() (struct{}, error) {
		_, _, err := client.PostMessageContext(ctx, channelID, slack.MsgOptionText(text, false))
		return struct{}{}, err
	})
	if err != nil {
		if err.Error() == "channel_not_found" {
			return port.ErrChannelNotFound
		}
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

func (a *SlackAdapter) PostThreadReply(ctx context.Context, channelID, threadTS, text string) error {
	msg := message.Message{Text: text}
	if err := msg.Validate(); err != nil {
		return err
	}

	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	_, err := withRetry(ctx, a.retryConfig, func() (struct{}, error) {
		_, _, err := client.PostMessageContext(
			ctx,
			channelID,
			slack.MsgOptionText(text, false),
			slack.MsgOptionTS(threadTS),
		)
		return struct{}{}, err
	})
	if err != nil {
		return fmt.Errorf("failed to post thread reply: %w", err)
	}

	return nil
}

func (a *SlackAdapter) GetChannelMembers(ctx context.Context, channelID string) ([]user.User, error) {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return nil, errors.New("not connected to Slack")
	}
	client := a.client
	a.mu.RUnlock()

	var allMembers []user.User
	cursor := ""

	for {
		params := &slack.GetUsersInConversationParameters{
			ChannelID: channelID,
			Cursor:    cursor,
			Limit:     1000,
		}

		result, err := withRetry(ctx, a.retryConfig, func() (struct {
			memberIDs  []string
			nextCursor string
		}, error) {
			memberIDs, nextCursor, err := client.GetUsersInConversationContext(ctx, params)
			return struct {
				memberIDs  []string
				nextCursor string
			}{memberIDs, nextCursor}, err
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get channel members: %w", err)
		}

		for _, userID := range result.memberIDs {
			userInfo, err := a.getUserInfo(ctx, userID)
			if err != nil {
				continue
			}

			displayName := userInfo.Profile.DisplayName
			if displayName == "" {
				displayName = userInfo.RealName
			}
			if displayName == "" {
				displayName = userInfo.Name
			}

			u := user.User{
				ID:          userInfo.ID,
				Name:        userInfo.Name,
				DisplayName: displayName,
			}

			allMembers = append(allMembers, u)
		}

		cursor = result.nextCursor
		if cursor == "" {
			break
		}
	}

	return allMembers, nil
}

func (a *SlackAdapter) SubscribeEvents(ctx context.Context) (<-chan port.SlackEvent, error) {
	a.mu.Lock()
	if !a.connected {
		a.mu.Unlock()
		return nil, errors.New("not connected to Slack")
	}

	if a.eventChan != nil {
		a.mu.Unlock()
		return a.eventChan, nil
	}

	a.eventChan = make(chan port.SlackEvent, 100)
	socketClient := a.socketClient
	a.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if a.logger != nil {
					a.logger.Error(context.Background(), "Socket Mode event loop panic", fmt.Errorf("%v", r))
				}
			}
		}()

		a.eventChan <- port.SlackConnectedEvent{}

		go func() {
			if err := socketClient.RunContext(a.ctx); err != nil {
				if a.logger != nil {
					a.logger.Error(context.Background(), "Socket Mode error", err)
				}
			}
		}()

		for {
			select {
			case <-a.ctx.Done():
				a.eventChan <- port.SlackDisconnectedEvent{Reason: "context cancelled"}
				return

			case evt := <-socketClient.Events:
				a.handleSocketModeEvent(evt)
			}
		}
	}()

	return a.eventChan, nil
}

func (a *SlackAdapter) handleSocketModeEvent(evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		var payload slackevents.EventsAPIEvent
		if err := json.Unmarshal(evt.Request.Payload, &payload); err != nil {
			if a.logger != nil {
				a.logger.Error(context.Background(), "Failed to unmarshal EventsAPI payload", err)
			}
			return
		}

		a.socketClient.Ack(*evt.Request)

		switch ev := payload.InnerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			a.handleMessageEventV2(ev)
		}

	case socketmode.EventTypeConnectionError:
		a.eventChan <- port.SlackDisconnectedEvent{Reason: "connection error"}

	case socketmode.EventTypeDisconnect:
		a.eventChan <- port.SlackDisconnectedEvent{Reason: "disconnected"}
	}
}

func (a *SlackAdapter) handleMessageEventV2(ev *slackevents.MessageEvent) {
	if ev.BotID != "" || ev.SubType != "" {
		return
	}

	userName := ev.User
	if userInfo, err := a.getUserInfo(context.Background(), ev.User); err == nil {
		userName = userInfo.Name
	}

	ts := parseSlackTimestamp(ev.TimeStamp)

	msg := message.Message{
		ID:        ev.TimeStamp,
		ChannelID: ev.Channel,
		UserID:    ev.User,
		UserName:  userName,
		Text:      ev.Text,
		Timestamp: ts,
		ThreadTS:  ev.ThreadTimeStamp,
	}

	a.eventChan <- port.NewMessageEvent{
		ChannelID: ev.Channel,
		Message:   msg,
	}
}

func (a *SlackAdapter) getUserInfo(ctx context.Context, userID string) (*slack.User, error) {
	if cached, ok := a.userCache.Get(userID); ok {
		return cached, nil
	}

	userInfo, err := withRetry(ctx, a.retryConfig, func() (*slack.User, error) {
		return a.client.GetUserInfoContext(ctx, userID)
	})
	if err != nil {
		return nil, err
	}

	a.userCache.Add(userID, userInfo)

	return userInfo, nil
}

func (a *SlackAdapter) batchGetUserInfo(ctx context.Context, userIDs []string) map[string]*slack.User {
	result := make(map[string]*slack.User, len(userIDs))
	resultMu := sync.Mutex{}

	var uncachedIDs []string
	for _, userID := range userIDs {
		if cached, ok := a.userCache.Get(userID); ok {
			result[userID] = cached
		} else {
			uncachedIDs = append(uncachedIDs, userID)
		}
	}

	if len(uncachedIDs) == 0 {
		return result
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(batchFetchConcurrency)

	for _, userID := range uncachedIDs {
		userID := userID
		g.Go(func() error {
			userInfo, err := withRetry(ctx, a.retryConfig, func() (*slack.User, error) {
				return a.client.GetUserInfoContext(ctx, userID)
			})
			if err != nil {
				return nil
			}

			a.userCache.Add(userID, userInfo)

			resultMu.Lock()
			result[userID] = userInfo
			resultMu.Unlock()

			return nil
		})
	}

	_ = g.Wait()

	return result
}

func convertToChannel(conv slack.Channel) channel.Channel {
	var channelType channel.Type
	if conv.IsIM {
		channelType = channel.TypeDM
	} else if conv.IsPrivate {
		channelType = channel.TypePrivate
	} else {
		channelType = channel.TypePublic
	}

	return channel.Channel{
		ID:              conv.ID,
		Name:            conv.Name,
		ChannelType:     channelType,
		UnreadCount:     0,
		LastMessageTime: time.Time{},
	}
}
