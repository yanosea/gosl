// Package tui provides TUI (Text User Interface) components using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
)

// TestPerformance_MessageRendering tests that message rendering completes within 500ms.
// Requirement 11.2: メッセージ一覧レンダリング500ms以内
func TestPerformance_MessageRendering(t *testing.T) {
	// Create MessageViewModel with many messages
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Generate 1000 messages
	messages := make([]message.Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserID:    fmt.Sprintf("user-%d", i%10),
			UserName:  fmt.Sprintf("User%d", i%10),
			Text:      fmt.Sprintf("This is test message number %d with some content", i),
			Timestamp: time.Now().Add(time.Duration(-i) * time.Second),
		}
	}

	viewModel.SetMessages(messages, "cursor")

	// Measure rendering time
	start := time.Now()
	rendered := viewModel.renderMessages()
	elapsed := time.Since(start)

	// Verify rendered content is not empty
	if len(rendered) == 0 {
		t.Error("Expected non-empty rendered content")
	}

	// Verify performance: should complete within 500ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("Message rendering took %v, expected < 500ms (Requirement 11.2)", elapsed)
	}

	t.Logf("Message rendering completed in %v (target: < 500ms)", elapsed)
}

// TestPerformance_KeyboardInputResponse tests keyboard input response time.
// Requirement 11.3: キーボード入力応答100ms以内
func TestPerformance_KeyboardInputResponse(t *testing.T) {
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Add some messages
	messages := make([]message.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserName:  "TestUser",
			Text:      "Test message",
			Timestamp: time.Now(),
		}
	}
	viewModel.SetMessages(messages, "")

	// Simulate key press (down arrow)
	type keyMsg struct {
		str string
	}
	msg := keyMsg{str: "down"}

	start := time.Now()
	viewModel.Update(msg)
	elapsed := time.Since(start)

	// Verify performance: should complete within 100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Keyboard input response took %v, expected < 100ms (Requirement 11.3)", elapsed)
	}

	t.Logf("Keyboard input response completed in %v (target: < 100ms)", elapsed)
}

// TestPerformance_LargeMessageScroll tests scrolling performance with many messages.
// Requirement 11.4: 1000件以上メッセージを含むチャンネルでもスムーズにスクロール
func TestPerformance_LargeMessageScroll(t *testing.T) {
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Generate 1500 messages
	messages := make([]message.Message, 1500)
	for i := 0; i < 1500; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserName:  fmt.Sprintf("User%d", i%10),
			Text:      strings.Repeat("a", 100), // 100 chars per message
			Timestamp: time.Now(),
		}
	}

	viewModel.SetMessages(messages, "")

	// Test multiple scroll operations
	start := time.Now()
	for i := 0; i < 100; i++ {
		type keyMsg struct {
			str string
		}
		viewModel.Update(keyMsg{str: "down"})
	}
	elapsed := time.Since(start)

	// Should handle 100 scroll operations quickly (< 500ms total)
	if elapsed > 500*time.Millisecond {
		t.Errorf("Scrolling 100 times took %v, expected < 500ms (Requirement 11.4)", elapsed)
	}

	t.Logf("Scrolling 100 messages completed in %v (target: < 500ms)", elapsed)
}

// TestPerformance_ChannelListDisplay tests channel list initial display time.
// Requirement 11.1: チャンネル一覧初期表示3秒以内
func TestPerformance_ChannelListDisplay(t *testing.T) {
	// Create ChannelListModel
	listModel := NewChannelListModel(80, 24)

	// Generate 100 channels
	channels := make([]channel.Channel, 100)
	for i := 0; i < 100; i++ {
		channels[i] = channel.Channel{
			ID:              fmt.Sprintf("ch-%d", i),
			Name:            fmt.Sprintf("Channel %d", i),
			ChannelType:     channel.TypePublic,
			UnreadCount:     i % 10,
			LastMessageTime: time.Now(),
		}
	}

	// Measure display time
	start := time.Now()
	listModel.SetChannels(channels)
	_ = listModel.View()
	elapsed := time.Since(start)

	// Verify performance: should complete within 3 seconds
	if elapsed > 3*time.Second {
		t.Errorf("Channel list display took %v, expected < 3s (Requirement 11.1)", elapsed)
	}

	t.Logf("Channel list display completed in %v (target: < 3s)", elapsed)
}

// BenchmarkMessageViewModel_RenderMessages benchmarks message rendering.
func BenchmarkMessageViewModel_RenderMessages(b *testing.B) {
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Generate 100 messages
	messages := make([]message.Message, 100)
	for i := 0; i < 100; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserName:  "TestUser",
			Text:      "Test message with some content",
			Timestamp: time.Now(),
		}
	}

	viewModel.SetMessages(messages, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viewModel.renderMessages()
	}
}

// BenchmarkMessageViewModel_RenderMessages_WithCache benchmarks cached rendering.
func BenchmarkMessageViewModel_RenderMessages_WithCache(b *testing.B) {
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Generate 100 messages
	messages := make([]message.Message, 100)
	for i := 0; i < 100; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserName:  "TestUser",
			Text:      "Test message with some content",
			Timestamp: time.Now(),
		}
	}

	viewModel.SetMessages(messages, "")

	// Pre-render to populate cache
	viewModel.renderMessages()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viewModel.renderMessages()
	}
}

// BenchmarkMessageViewModel_LargeMessages benchmarks rendering with large message content.
func BenchmarkMessageViewModel_LargeMessages(b *testing.B) {
	viewModel := NewMessageViewModel("test-channel", 80, 24)

	// Generate messages with large content
	messages := make([]message.Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = message.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			ChannelID: "test-channel",
			UserName:  "TestUser",
			Text:      strings.Repeat("a", 1000), // 1KB per message
			Timestamp: time.Now(),
		}
	}

	viewModel.SetMessages(messages, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viewModel.renderMessages()
	}
}
