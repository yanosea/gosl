// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/message"
)

func TestRenderCache_Get(t *testing.T) {
	cache := NewRenderCache()
	msg := message.Message{
		ID:        "msg1",
		ChannelID: "ch1",
		UserID:    "user1",
		UserName:  "Alice",
		Text:      "Hello World",
		Timestamp: time.Now(),
	}

	// Initially no cache
	_, found := cache.Get(msg.ID)
	if found {
		t.Error("Expected no cache entry initially")
	}

	// Set cache
	rendered := "rendered content"
	cache.Set(msg.ID, rendered)

	// Should retrieve cached value
	cachedValue, found := cache.Get(msg.ID)
	if !found {
		t.Error("Expected cache entry to be found")
	}
	if cachedValue != rendered {
		t.Errorf("Expected cached value %q, got %q", rendered, cachedValue)
	}
}

func TestRenderCache_Invalidate(t *testing.T) {
	cache := NewRenderCache()
	msg := message.Message{
		ID:        "msg1",
		ChannelID: "ch1",
		UserID:    "user1",
		UserName:  "Alice",
		Text:      "Hello World",
		Timestamp: time.Now(),
	}

	// Set cache
	cache.Set(msg.ID, "rendered content")

	// Verify cached
	_, found := cache.Get(msg.ID)
	if !found {
		t.Error("Expected cache entry to be found")
	}

	// Invalidate
	cache.Invalidate(msg.ID)

	// Should no longer be cached
	_, found = cache.Get(msg.ID)
	if found {
		t.Error("Expected cache entry to be invalidated")
	}
}

func TestRenderCache_InvalidateAll(t *testing.T) {
	cache := NewRenderCache()

	// Set multiple cache entries
	cache.Set("msg1", "content1")
	cache.Set("msg2", "content2")
	cache.Set("msg3", "content3")

	// Verify all cached
	for _, id := range []string{"msg1", "msg2", "msg3"} {
		_, found := cache.Get(id)
		if !found {
			t.Errorf("Expected cache entry %s to be found", id)
		}
	}

	// Invalidate all
	cache.InvalidateAll()

	// All should be cleared
	for _, id := range []string{"msg1", "msg2", "msg3"} {
		_, found := cache.Get(id)
		if found {
			t.Errorf("Expected cache entry %s to be invalidated", id)
		}
	}
}

func TestRenderCache_ConcurrentAccess(t *testing.T) {
	cache := NewRenderCache()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			msgID := string(rune('a' + id))
			cache.Set(msgID, "content")
			cache.Get(msgID)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkRenderCache_SetAndGet(b *testing.B) {
	cache := NewRenderCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgID := "msg"
		content := "rendered content for benchmark"
		cache.Set(msgID, content)
		cache.Get(msgID)
	}
}
