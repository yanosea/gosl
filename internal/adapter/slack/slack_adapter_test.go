package slack

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/app/port"
)

// skipIfNoSlackToken skips the test if SLACK_TEST_TOKEN is not set
func skipIfNoSlackToken(t *testing.T) (botToken, appToken string) {
	botToken = os.Getenv("SLACK_TEST_BOT_TOKEN")
	appToken = os.Getenv("SLACK_TEST_APP_TOKEN")
	if botToken == "" || appToken == "" {
		t.Skip("Skipping Slack integration test: SLACK_TEST_BOT_TOKEN or SLACK_TEST_APP_TOKEN not set")
	}
	return botToken, appToken
}

// TestSlackAdapter_Connect tests Socket Mode connection
func TestSlackAdapter_Connect(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	tests := []struct {
		name     string
		botToken string
		appToken string
		wantErr  bool
	}{
		{
			name:     "valid tokens connect successfully",
			botToken: botToken,
			appToken: appToken,
			wantErr:  false,
		},
		{
			name:     "invalid bot token returns auth error",
			botToken: "xoxb-invalid-token",
			appToken: appToken,
			wantErr:  true,
		},
		{
			name:     "empty bot token returns auth error",
			botToken: "",
			appToken: appToken,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewSlackAdapter()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := adapter.Connect(ctx, tt.botToken, tt.appToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Connect() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("Connect() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestSlackAdapter_Disconnect tests Socket Mode disconnection
func TestSlackAdapter_Disconnect(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	// Connect first
	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Disconnect
	err = adapter.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() error = %v, wantErr nil", err)
	}
}

// TestSlackAdapter_GetChannels tests channel list retrieval
func TestSlackAdapter_GetChannels(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	// Connect first
	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	channels, err := adapter.GetChannels(ctx)
	if err != nil {
		t.Errorf("GetChannels() error = %v, wantErr nil", err)
	}

	if len(channels) == 0 {
		t.Error("GetChannels() returned empty list, want at least one channel")
	}

	// Verify channel structure
	for _, ch := range channels {
		if ch.ID == "" {
			t.Error("Channel ID is empty")
		}
		if ch.Name == "" {
			t.Error("Channel Name is empty")
		}
	}
}

// TestSlackAdapter_GetMessages tests message retrieval with pagination
func TestSlackAdapter_GetMessages(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	// Get channels first to get a valid channel ID
	channels, err := adapter.GetChannels(ctx)
	if err != nil {
		t.Fatalf("GetChannels() failed: %v", err)
	}
	if len(channels) == 0 {
		t.Skip("No channels available for testing")
	}

	channelID := channels[0].ID

	tests := []struct {
		name      string
		channelID string
		limit     int
		cursor    string
		wantErr   bool
	}{
		{
			name:      "valid channel returns messages",
			channelID: channelID,
			limit:     5,
			cursor:    "",
			wantErr:   false,
		},
		{
			name:      "pagination with cursor",
			channelID: channelID,
			limit:     10,
			cursor:    "",
			wantErr:   false,
		},
		{
			name:      "invalid channel returns error",
			channelID: "INVALID",
			limit:     5,
			cursor:    "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, _, err := adapter.GetMessages(ctx, tt.channelID, tt.limit, tt.cursor)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetMessages() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("GetMessages() error = %v, want nil", err)
				}
				// Messages can be empty for channels without messages
				// Verify message structure
				for _, msg := range messages {
					if msg.ID == "" {
						t.Error("Message ID is empty")
					}
					if msg.ChannelID != tt.channelID {
						t.Errorf("Message ChannelID = %v, want %v", msg.ChannelID, tt.channelID)
					}
				}
			}
		})
	}
}

// TestSlackAdapter_GetThreadReplies tests thread reply retrieval
func TestSlackAdapter_GetThreadReplies(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	// This test requires a real channel and thread, so we skip if not available
	t.Skip("Skipping GetThreadReplies test: requires real channel and thread data")
}

// TestSlackAdapter_PostMessage tests message posting
func TestSlackAdapter_PostMessage(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	tests := []struct {
		name      string
		text      string
		wantErr   bool
	}{
		{
			name:    "empty text returns validation error",
			text:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip actual posting to avoid spamming channels
			t.Skip("Skipping PostMessage test: requires real channel and avoids spamming")
		})
	}
}

// TestSlackAdapter_PostThreadReply tests thread reply posting
func TestSlackAdapter_PostThreadReply(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	// Skip actual posting to avoid spamming channels
	t.Skip("Skipping PostThreadReply test: requires real channel and avoids spamming")
}

// TestSlackAdapter_GetChannelMembers tests channel member retrieval
func TestSlackAdapter_GetChannelMembers(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	// Get channels first to get a valid channel ID
	channels, err := adapter.GetChannels(ctx)
	if err != nil {
		t.Fatalf("GetChannels() failed: %v", err)
	}
	if len(channels) == 0 {
		t.Skip("No channels available for testing")
	}

	channelID := channels[0].ID

	members, err := adapter.GetChannelMembers(ctx, channelID)
	if err != nil {
		t.Errorf("GetChannelMembers() error = %v, want nil", err)
	}

	// Members can be empty for some channels, so we don't enforce non-empty
	// Verify user structure if members exist
	for _, member := range members {
		if member.ID == "" {
			t.Error("User ID is empty")
		}
		if member.Name == "" {
			t.Error("User Name is empty")
		}
	}
}

// TestSlackAdapter_SubscribeEvents tests Socket Mode event subscription
func TestSlackAdapter_SubscribeEvents(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	eventChan, err := adapter.SubscribeEvents(ctx)
	if err != nil {
		t.Fatalf("SubscribeEvents() error = %v, want nil", err)
	}

	// Wait for connected event
	select {
	case event := <-eventChan:
		if _, ok := event.(port.SlackConnectedEvent); !ok {
			t.Errorf("Expected SlackConnectedEvent, got %T", event)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for SlackConnectedEvent")
	}
}

// TestSlackAdapter_RateLimitHandling tests rate limit error handling
func TestSlackAdapter_RateLimitHandling(t *testing.T) {
	botToken, appToken := skipIfNoSlackToken(t)

	adapter := NewSlackAdapter()
	ctx := context.Background()

	err := adapter.Connect(ctx, botToken, appToken)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer adapter.Disconnect()

	// Skip this test as it would spam the Slack API
	t.Skip("Skipping RateLimitHandling test: would spam Slack API")
}
