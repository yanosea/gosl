package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/cache"
	"github.com/yanosea/gosl/internal/domain/message"
)

func TestNewMessageCache(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024) // 20 channels, 100MB

	if mc == nil {
		t.Fatal("NewMessageCache returned nil")
	}
}

func TestMessageCache_StoreAndGetMessages(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	channelID := "C123456"
	messages := []message.Message{
		message.NewMessage("1", channelID, "U1", "user1", "Hello", time.Now()),
		message.NewMessage("2", channelID, "U2", "user2", "World", time.Now()),
	}
	cursor := "next_cursor_123"

	// Store messages
	mc.StoreMessages(channelID, messages, cursor)

	// Get messages
	got, gotCursor, found := mc.GetMessages(channelID)

	if !found {
		t.Error("GetMessages() found = false, want true")
	}

	if len(got) != len(messages) {
		t.Errorf("GetMessages() returned %d messages, want %d", len(got), len(messages))
	}

	if gotCursor != cursor {
		t.Errorf("GetMessages() cursor = %v, want %v", gotCursor, cursor)
	}
}

func TestMessageCache_GetMessages_NotFound(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	_, _, found := mc.GetMessages("C999999")

	if found {
		t.Error("GetMessages() found = true, want false for non-existent channel")
	}
}

func TestMessageCache_AppendMessages(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	channelID := "C123456"
	initialMessages := []message.Message{
		message.NewMessage("1", channelID, "U1", "user1", "First", time.Now()),
	}
	mc.StoreMessages(channelID, initialMessages, "cursor1")

	// Append more messages
	additionalMessages := []message.Message{
		message.NewMessage("2", channelID, "U2", "user2", "Second", time.Now()),
		message.NewMessage("3", channelID, "U3", "user3", "Third", time.Now()),
	}
	mc.StoreMessages(channelID, additionalMessages, "cursor2")

	// Verify all messages are stored
	got, gotCursor, found := mc.GetMessages(channelID)

	if !found {
		t.Error("GetMessages() found = false, want true")
	}

	if len(got) != 3 {
		t.Errorf("GetMessages() returned %d messages, want 3", len(got))
	}

	if gotCursor != "cursor2" {
		t.Errorf("GetMessages() cursor = %v, want cursor2", gotCursor)
	}
}

func TestMessageCache_ConcurrentAccess(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent writes and reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			channelID := "C123456"

			for j := 0; j < numOperations; j++ {
				// Write
				messages := []message.Message{
					message.NewMessage("1", channelID, "U1", "user1", "Test", time.Now()),
				}
				mc.StoreMessages(channelID, messages, "cursor")

				// Read
				mc.GetMessages(channelID)
			}
		}(i)
	}

	wg.Wait()
}

func TestMessageCache_SetCurrentChannel(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	channelID := "C123456"
	mc.SetCurrentChannel(channelID)

	current := mc.GetCurrentChannel()
	if current != channelID {
		t.Errorf("GetCurrentChannel() = %v, want %v", current, channelID)
	}
}

func TestMessageCache_EstimateMemoryUsage(t *testing.T) {
	mc := cache.NewMessageCache(20, 100*1024*1024)

	channelID := "C123456"
	messages := []message.Message{
		message.NewMessage("1", channelID, "U1", "user1", "Hello", time.Now()),
		message.NewMessage("2", channelID, "U2", "user2", "World", time.Now()),
	}
	mc.StoreMessages(channelID, messages, "cursor")

	usage := mc.EstimateMemoryUsage()

	// Expect approximately 2 messages * 500 bytes = 1000 bytes
	if usage < 500 || usage > 2000 {
		t.Errorf("EstimateMemoryUsage() = %d, want approximately 1000", usage)
	}
}

func TestMessageCache_LRUEviction(t *testing.T) {
	// Create cache with very small memory limit to force eviction
	mc := cache.NewMessageCache(2, 1500) // Allow only ~3 messages (3 * 500 bytes)

	// Add messages to channel 1
	ch1 := "C111111"
	messages1 := []message.Message{
		message.NewMessage("1", ch1, "U1", "user1", "Message 1", time.Now()),
		message.NewMessage("2", ch1, "U1", "user1", "Message 2", time.Now()),
	}
	mc.StoreMessages(ch1, messages1, "cursor1")

	// Add messages to channel 2
	ch2 := "C222222"
	messages2 := []message.Message{
		message.NewMessage("3", ch2, "U2", "user2", "Message 3", time.Now()),
	}
	mc.StoreMessages(ch2, messages2, "cursor2")

	// Add messages to channel 3 - this should trigger eviction of ch1
	ch3 := "C333333"
	messages3 := []message.Message{
		message.NewMessage("4", ch3, "U3", "user3", "Message 4", time.Now()),
	}
	mc.StoreMessages(ch3, messages3, "cursor3")

	// Channel 1 should be evicted
	_, _, found := mc.GetMessages(ch1)
	if found {
		t.Error("Channel 1 should have been evicted but was still found")
	}

	// Channels 2 and 3 should still be present
	_, _, found2 := mc.GetMessages(ch2)
	if !found2 {
		t.Error("Channel 2 should still be in cache")
	}

	_, _, found3 := mc.GetMessages(ch3)
	if !found3 {
		t.Error("Channel 3 should still be in cache")
	}
}

func TestMessageCache_NoEvictCurrentChannel(t *testing.T) {
	// Create cache with very small memory limit
	mc := cache.NewMessageCache(2, 1500)

	// Set current channel
	ch1 := "C111111"
	mc.SetCurrentChannel(ch1)

	// Add messages to current channel
	messages1 := []message.Message{
		message.NewMessage("1", ch1, "U1", "user1", "Message 1", time.Now()),
		message.NewMessage("2", ch1, "U1", "user1", "Message 2", time.Now()),
	}
	mc.StoreMessages(ch1, messages1, "cursor1")

	// Add messages to other channels to trigger eviction
	ch2 := "C222222"
	messages2 := []message.Message{
		message.NewMessage("3", ch2, "U2", "user2", "Message 3", time.Now()),
	}
	mc.StoreMessages(ch2, messages2, "cursor2")

	ch3 := "C333333"
	messages3 := []message.Message{
		message.NewMessage("4", ch3, "U3", "user3", "Message 4", time.Now()),
	}
	mc.StoreMessages(ch3, messages3, "cursor3")

	// Current channel should NOT be evicted
	_, _, found := mc.GetMessages(ch1)
	if !found {
		t.Error("Current channel should not have been evicted")
	}
}
