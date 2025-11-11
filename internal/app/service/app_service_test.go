package service

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/yanosea/gosl/internal/app/port/mock"
	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
)

// TestNewAppService tests creating a new AppService
func TestNewAppService(t *testing.T) {
	service := NewAppService(nil, nil, nil)

	if service == nil {
		t.Fatal("NewAppService returned nil")
	}
}

// TestAppService_GetChannels tests retrieving channels
func TestAppService_GetChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := mock.NewMockSlackRepository(ctrl)
	expectedChannels := []channel.Channel{
		{
			ID:          "C123",
			Name:        "general",
			ChannelType: channel.TypePublic,
		},
		{
			ID:          "C456",
			Name:        "random",
			ChannelType: channel.TypePublic,
		},
	}

	mockSlack.EXPECT().
		GetChannels(gomock.Any()).
		Return(expectedChannels, nil)

	service := NewAppService(nil, mockSlack, nil)
	ctx := context.Background()

	channels, err := service.GetChannels(ctx)
	if err != nil {
		t.Fatalf("GetChannels() error = %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("GetChannels() returned %d channels, want 2", len(channels))
	}

	if channels[0].Name != "general" {
		t.Errorf("First channel name = %v, want general", channels[0].Name)
	}
}

// TestAppService_GetMessages tests retrieving messages with caching
func TestAppService_GetMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := mock.NewMockSlackRepository(ctrl)
	expectedMessages := []message.Message{
		{
			ID:        "msg1",
			ChannelID: "C123",
			UserName:  "user1",
			Text:      "Hello",
			Timestamp: time.Now(),
		},
	}

	mockSlack.EXPECT().
		GetMessages(gomock.Any(), "C123", 5, "").
		Return(expectedMessages, "cursor123", nil)

	service := NewAppService(nil, mockSlack, nil)
	ctx := context.Background()

	// First call should fetch from Slack and cache
	messages, cursor, err := service.GetMessages(ctx, "C123", 5, "")
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("GetMessages() returned %d messages, want 1", len(messages))
	}

	if cursor != "cursor123" {
		t.Errorf("GetMessages() cursor = %v, want cursor123", cursor)
	}

	// Note: Cache functionality is tested separately
	// This test just verifies the service layer logic
}

// TestAppService_GetThreadReplies tests retrieving thread replies
func TestAppService_GetThreadReplies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := mock.NewMockSlackRepository(ctrl)
	parentMessage := message.Message{
		ID:        "msg1",
		ChannelID: "C123",
		ThreadTS:  "1234567890.123456",
		Text:      "Parent message",
	}
	replies := []message.Message{
		{
			ID:       "msg2",
			ThreadTS: "1234567890.123456",
			Text:     "Reply 1",
		},
	}

	mockSlack.EXPECT().
		GetThreadReplies(gomock.Any(), "C123", "1234567890.123456").
		Return(parentMessage, replies, nil)

	service := NewAppService(nil, mockSlack, nil)
	ctx := context.Background()

	parent, replyResults, err := service.GetThreadReplies(ctx, "C123", "1234567890.123456")
	if err != nil {
		t.Fatalf("GetThreadReplies() error = %v", err)
	}

	if parent.Text != "Parent message" {
		t.Errorf("Parent text = %v, want 'Parent message'", parent.Text)
	}

	if len(replyResults) != 1 {
		t.Errorf("GetThreadReplies() returned %d replies, want 1", len(replyResults))
	}
}

// TestAppService_SendMessage tests sending a message with validation
func TestAppService_SendMessage(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "valid message",
			text:    "Hello, world!",
			wantErr: false,
		},
		{
			name:    "empty message",
			text:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSlack := mock.NewMockSlackRepository(ctrl)

			if !tt.wantErr {
				mockSlack.EXPECT().
					PostMessage(gomock.Any(), "C123", tt.text).
					Return(nil)
			}

			service := NewAppService(nil, mockSlack, nil)
			ctx := context.Background()

			err := service.SendMessage(ctx, "C123", tt.text)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAppService_SendThreadReply tests sending a thread reply
func TestAppService_SendThreadReply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlack := mock.NewMockSlackRepository(ctrl)
	mockSlack.EXPECT().
		PostThreadReply(gomock.Any(), "C123", "1234567890.123456", "Reply text").
		Return(nil)

	service := NewAppService(nil, mockSlack, nil)
	ctx := context.Background()

	err := service.SendThreadReply(ctx, "C123", "1234567890.123456", "Reply text")
	if err != nil {
		t.Fatalf("SendThreadReply() error = %v", err)
	}
}
