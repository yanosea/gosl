package message_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/message"
)

func TestNewMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		id       string
		channelID string
		userID   string
		userName string
		text     string
		timestamp time.Time
		expected message.Message
	}{
		{
			name:      "basic message",
			id:        "1234567890.123456",
			channelID: "C123456",
			userID:    "U123456",
			userName:  "testuser",
			text:      "Hello, world!",
			timestamp: now,
			expected: message.Message{
				ID:        "1234567890.123456",
				ChannelID: "C123456",
				UserID:    "U123456",
				UserName:  "testuser",
				Text:      "Hello, world!",
				Timestamp: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.NewMessage(tt.id, tt.channelID, tt.userID, tt.userName, tt.text, tt.timestamp)

			if msg.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", msg.ID, tt.expected.ID)
			}
			if msg.ChannelID != tt.expected.ChannelID {
				t.Errorf("ChannelID = %v, want %v", msg.ChannelID, tt.expected.ChannelID)
			}
			if msg.UserID != tt.expected.UserID {
				t.Errorf("UserID = %v, want %v", msg.UserID, tt.expected.UserID)
			}
			if msg.UserName != tt.expected.UserName {
				t.Errorf("UserName = %v, want %v", msg.UserName, tt.expected.UserName)
			}
			if msg.Text != tt.expected.Text {
				t.Errorf("Text = %v, want %v", msg.Text, tt.expected.Text)
			}
			if !msg.Timestamp.Equal(tt.expected.Timestamp) {
				t.Errorf("Timestamp = %v, want %v", msg.Timestamp, tt.expected.Timestamp)
			}
		})
	}
}

func TestMessage_IsThread(t *testing.T) {
	tests := []struct {
		name     string
		threadTS string
		expected bool
	}{
		{
			name:     "message with thread",
			threadTS: "1234567890.123456",
			expected: true,
		},
		{
			name:     "message without thread",
			threadTS: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.Message{
				ID:       "1234567890.123456",
				ThreadTS: tt.threadTS,
			}

			result := msg.IsThread()
			if result != tt.expected {
				t.Errorf("IsThread() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessage_Validate(t *testing.T) {
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
		{
			name:    "message at max length",
			text:    strings.Repeat("a", 40000),
			wantErr: false,
		},
		{
			name:    "message exceeds max length",
			text:    strings.Repeat("a", 40001),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.Message{
				ID:   "1234567890.123456",
				Text: tt.text,
			}

			err := msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_SetThreadInfo(t *testing.T) {
	tests := []struct {
		name        string
		threadTS    string
		replyCount  int
		wantThreadTS string
		wantReplies  int
	}{
		{
			name:        "set thread info",
			threadTS:    "1234567890.123456",
			replyCount:  5,
			wantThreadTS: "1234567890.123456",
			wantReplies:  5,
		},
		{
			name:        "clear thread info",
			threadTS:    "",
			replyCount:  0,
			wantThreadTS: "",
			wantReplies:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.Message{
				ID: "1234567890.123456",
			}

			msg.SetThreadInfo(tt.threadTS, tt.replyCount)

			if msg.ThreadTS != tt.wantThreadTS {
				t.Errorf("ThreadTS = %v, want %v", msg.ThreadTS, tt.wantThreadTS)
			}
			if msg.ReplyCount != tt.wantReplies {
				t.Errorf("ReplyCount = %v, want %v", msg.ReplyCount, tt.wantReplies)
			}
		})
	}
}
