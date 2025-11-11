package tui

import (
	"sync"
)

type RenderCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewRenderCache() *RenderCache {
	return &RenderCache{
		cache: make(map[string]string),
	}
}

func (rc *RenderCache) Get(messageID string) (string, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	content, found := rc.cache[messageID]
	return content, found
}

func (rc *RenderCache) Set(messageID string, content string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache[messageID] = content
}

func (rc *RenderCache) Invalidate(messageID string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	delete(rc.cache, messageID)
}

func (rc *RenderCache) InvalidateAll() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache = make(map[string]string)
}
