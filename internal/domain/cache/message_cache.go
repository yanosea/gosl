package cache

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/domain/message"
)

var _ port.CacheRepository = (*MessageCache)(nil)

const (
	AverageMessageSize = 500
)

type MessageCache struct {
	mu             sync.RWMutex
	messages       map[string][]message.Message
	cursors        map[string]string
	lru            *lru.Cache[string, bool]
	maxChannels    int
	maxMemoryBytes int64
	currentChannel string
}

func NewMessageCache(maxChannels int, maxMemoryBytes int64) *MessageCache {
	mc := &MessageCache{
		messages:       make(map[string][]message.Message),
		cursors:        make(map[string]string),
		maxChannels:    maxChannels,
		maxMemoryBytes: maxMemoryBytes,
	}

	lruCache, _ := lru.NewWithEvict[string, bool](maxChannels, func(key string, value bool) {
		if key != mc.currentChannel {
			delete(mc.messages, key)
			delete(mc.cursors, key)
		}
	})

	mc.lru = lruCache
	return mc
}

func (c *MessageCache) GetMessages(channelID string) ([]message.Message, string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	msgs, ok := c.messages[channelID]
	if !ok {
		return nil, "", false
	}

	cursor := c.cursors[channelID]
	c.lru.Get(channelID)

	return msgs, cursor, true
}

func (c *MessageCache) StoreMessages(channelID string, msgs []message.Message, cursor string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages[channelID] = append(c.messages[channelID], msgs...)
	c.cursors[channelID] = cursor
	c.lru.Add(channelID, true)

	for c.estimateMemoryUsage() > c.maxMemoryBytes {
		if !c.evictLRU() {
			break
		}
	}
}

func (c *MessageCache) SetCurrentChannel(channelID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentChannel = channelID
}

func (c *MessageCache) GetCurrentChannel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentChannel
}

func (c *MessageCache) EstimateMemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.estimateMemoryUsage()
}

func (c *MessageCache) estimateMemoryUsage() int64 {
	totalMessages := 0
	for _, msgs := range c.messages {
		totalMessages += len(msgs)
	}
	return int64(totalMessages * AverageMessageSize)
}

func (c *MessageCache) evictLRU() bool {
	keys := c.lru.Keys()
	if len(keys) == 0 {
		return false
	}

	for _, key := range keys {
		if key != c.currentChannel {
			delete(c.messages, key)
			delete(c.cursors, key)
			c.lru.Remove(key)
			return true
		}
	}

	return false
}
