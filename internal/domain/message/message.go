package message

import (
	"errors"
	"time"
)

const (
	MaxMessageLength = 40000
)

var (
	ErrEmptyMessage   = errors.New("message text cannot be empty")
	ErrMessageTooLong = errors.New("message text exceeds maximum length of 40,000 characters")
)

type Message struct {
	ID         string
	ChannelID  string
	UserID     string
	UserName   string
	Text       string
	Timestamp  time.Time
	ThreadTS   string
	ReplyCount int
}

func NewMessage(id, channelID, userID, userName, text string, timestamp time.Time) Message {
	return Message{
		ID:        id,
		ChannelID: channelID,
		UserID:    userID,
		UserName:  userName,
		Text:      text,
		Timestamp: timestamp,
	}
}

func (m *Message) IsThread() bool {
	return m.ThreadTS != ""
}

func (m *Message) Validate() error {
	if m.Text == "" {
		return ErrEmptyMessage
	}
	if len(m.Text) > MaxMessageLength {
		return ErrMessageTooLong
	}
	return nil
}

func (m *Message) SetThreadInfo(threadTS string, replyCount int) {
	m.ThreadTS = threadTS
	m.ReplyCount = replyCount
}
