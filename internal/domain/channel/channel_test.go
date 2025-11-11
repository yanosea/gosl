package channel_test

import (
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/channel"
)

func TestNewChannel(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		chName   string
		chType   channel.Type
		expected channel.Channel
	}{
		{
			name:   "public channel",
			id:     "C123456",
			chName: "general",
			chType: channel.TypePublic,
			expected: channel.Channel{
				ID:          "C123456",
				Name:        "general",
				ChannelType: channel.TypePublic,
				UnreadCount: 0,
			},
		},
		{
			name:   "private channel",
			id:     "G123456",
			chName: "private-team",
			chType: channel.TypePrivate,
			expected: channel.Channel{
				ID:          "G123456",
				Name:        "private-team",
				ChannelType: channel.TypePrivate,
				UnreadCount: 0,
			},
		},
		{
			name:   "direct message",
			id:     "D123456",
			chName: "user1",
			chType: channel.TypeDM,
			expected: channel.Channel{
				ID:          "D123456",
				Name:        "user1",
				ChannelType: channel.TypeDM,
				UnreadCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := channel.NewChannel(tt.id, tt.chName, tt.chType)

			if ch.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", ch.ID, tt.expected.ID)
			}
			if ch.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", ch.Name, tt.expected.Name)
			}
			if ch.ChannelType != tt.expected.ChannelType {
				t.Errorf("ChannelType = %v, want %v", ch.ChannelType, tt.expected.ChannelType)
			}
			if ch.UnreadCount != tt.expected.UnreadCount {
				t.Errorf("UnreadCount = %v, want %v", ch.UnreadCount, tt.expected.UnreadCount)
			}
		})
	}
}

func TestChannel_UpdateUnreadCount(t *testing.T) {
	tests := []struct {
		name          string
		initialCount  int
		newCount      int
		expectedCount int
	}{
		{
			name:          "increase unread count",
			initialCount:  0,
			newCount:      5,
			expectedCount: 5,
		},
		{
			name:          "decrease unread count",
			initialCount:  10,
			newCount:      3,
			expectedCount: 3,
		},
		{
			name:          "reset to zero",
			initialCount:  5,
			newCount:      0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := channel.Channel{
				ID:          "C123456",
				Name:        "test",
				ChannelType: channel.TypePublic,
				UnreadCount: tt.initialCount,
			}

			ch.UpdateUnreadCount(tt.newCount)

			if ch.UnreadCount != tt.expectedCount {
				t.Errorf("UnreadCount = %v, want %v", ch.UnreadCount, tt.expectedCount)
			}
		})
	}
}

func TestChannel_UpdateLastMessageTime(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	tests := []struct {
		name         string
		initialTime  time.Time
		newTime      time.Time
		expectedTime time.Time
	}{
		{
			name:         "update to newer time",
			initialTime:  now,
			newTime:      later,
			expectedTime: later,
		},
		{
			name:         "update from zero time",
			initialTime:  time.Time{},
			newTime:      now,
			expectedTime: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := channel.Channel{
				ID:              "C123456",
				Name:            "test",
				ChannelType:     channel.TypePublic,
				LastMessageTime: tt.initialTime,
			}

			ch.UpdateLastMessageTime(tt.newTime)

			if !ch.LastMessageTime.Equal(tt.expectedTime) {
				t.Errorf("LastMessageTime = %v, want %v", ch.LastMessageTime, tt.expectedTime)
			}
		})
	}
}

func TestChannelType_String(t *testing.T) {
	tests := []struct {
		name     string
		chType   channel.Type
		expected string
	}{
		{
			name:     "public channel",
			chType:   channel.TypePublic,
			expected: "Public",
		},
		{
			name:     "private channel",
			chType:   channel.TypePrivate,
			expected: "Private",
		},
		{
			name:     "direct message",
			chType:   channel.TypeDM,
			expected: "DM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chType.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
