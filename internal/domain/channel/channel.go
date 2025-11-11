package channel

import "time"

type Type int

const (
	TypePublic Type = iota
	TypePrivate
	TypeDM
)

func (t Type) String() string {
	switch t {
	case TypePublic:
		return "Public"
	case TypePrivate:
		return "Private"
	case TypeDM:
		return "DM"
	default:
		return "Unknown"
	}
}

type Channel struct {
	ID              string
	Name            string
	ChannelType     Type
	UnreadCount     int
	LastMessageTime time.Time
}

func NewChannel(id, name string, channelType Type) Channel {
	return Channel{
		ID:          id,
		Name:        name,
		ChannelType: channelType,
		UnreadCount: 0,
	}
}

func (c *Channel) UpdateUnreadCount(count int) {
	c.UnreadCount = count
}

func (c *Channel) UpdateLastMessageTime(t time.Time) {
	c.LastMessageTime = t
}
