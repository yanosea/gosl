package port

import "github.com/yanosea/gosl/internal/domain/message"

type CacheRepository interface {
	GetMessages(channelID string) ([]message.Message, string, bool)
	StoreMessages(channelID string, msgs []message.Message, cursor string)
	InvalidateMessages(channelID string)
	SetCurrentChannel(channelID string)
	GetCurrentChannel() string
	EstimateMemoryUsage() int64
}
