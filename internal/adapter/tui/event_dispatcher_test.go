package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/app/port"
	"github.com/yanosea/gosl/internal/domain/channel"
	"github.com/yanosea/gosl/internal/domain/message"
)

// TestEventDispatcher_ConvertSlackEvent tests conversion of Slack events to Bubble Tea messages
func TestEventDispatcher_ConvertSlackEvent(t *testing.T) {
	tests := []struct {
		name      string
		slackEvt  port.SlackEvent
		wantType  string
		wantError bool
	}{
		{
			name:     "SlackConnectedEvent converts to SlackConnectedMsg",
			slackEvt: port.SlackConnectedEvent{},
			wantType: "SlackConnectedMsg",
		},
		{
			name: "SlackDisconnectedEvent converts to SlackDisconnectedMsg",
			slackEvt: port.SlackDisconnectedEvent{
				Reason: "connection lost",
			},
			wantType: "SlackDisconnectedMsg",
		},
		{
			name: "NewMessageEvent converts to NewMessageMsg",
			slackEvt: port.NewMessageEvent{
				ChannelID: "C123",
				Message: message.Message{
					ID:        "msg1",
					ChannelID: "C123",
					UserID:    "U456",
					UserName:  "testuser",
					Text:      "Hello, world!",
					Timestamp: time.Now(),
				},
			},
			wantType: "NewMessageMsg",
		},
		{
			name: "ChannelUpdateEvent converts to ChannelUpdateMsg",
			slackEvt: port.ChannelUpdateEvent{
				Channel: channel.Channel{
					ID:          "C123",
					Name:        "general",
					ChannelType: channel.TypePublic,
				},
			},
			wantType: "ChannelUpdateMsg",
		},
		{
			name: "UserTypingEvent converts to UserTypingMsg",
			slackEvt: port.UserTypingEvent{
				ChannelID: "C123",
				UserID:    "U456",
			},
			wantType: "UserTypingMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := NewEventDispatcher()
			msg := dispatcher.convertSlackEvent(tt.slackEvt)

			// Check if the message type matches expected
			switch tt.wantType {
			case "SlackConnectedMsg":
				if _, ok := msg.(SlackConnectedMsg); !ok {
					t.Errorf("Expected SlackConnectedMsg, got %T", msg)
				}
			case "SlackDisconnectedMsg":
				if _, ok := msg.(SlackDisconnectedMsg); !ok {
					t.Errorf("Expected SlackDisconnectedMsg, got %T", msg)
				}
			case "NewMessageMsg":
				if _, ok := msg.(NewMessageMsg); !ok {
					t.Errorf("Expected NewMessageMsg, got %T", msg)
				}
			case "ChannelUpdateMsg":
				if _, ok := msg.(ChannelUpdateMsg); !ok {
					t.Errorf("Expected ChannelUpdateMsg, got %T", msg)
				}
			case "UserTypingMsg":
				if _, ok := msg.(UserTypingMsg); !ok {
					t.Errorf("Expected UserTypingMsg, got %T", msg)
				}
			}
		})
	}
}

// TestEventDispatcher_Start tests the event dispatcher lifecycle
func TestEventDispatcher_Start(t *testing.T) {
	dispatcher := NewEventDispatcher()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock event channel
	eventChan := make(chan port.SlackEvent, 10)

	// Create a test program (we'll use a minimal model)
	testModel := &testBubbleTeaModel{
		receivedMsgs: make([]tea.Msg, 0),
	}
	program := tea.NewProgram(testModel)

	// Start the dispatcher
	err := dispatcher.Start(ctx, eventChan, program)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Send a test event
	testEvent := port.SlackConnectedEvent{}
	eventChan <- testEvent

	// Give the dispatcher time to process
	time.Sleep(100 * time.Millisecond)

	// Verify the event was converted and sent
	// Note: In a real test, we'd need a way to intercept Program.Send()
	// For now, we're just testing that Start() doesn't error
}

// TestEventDispatcher_Stop tests graceful shutdown
func TestEventDispatcher_Stop(t *testing.T) {
	dispatcher := NewEventDispatcher()
	ctx, cancel := context.WithCancel(context.Background())

	eventChan := make(chan port.SlackEvent, 10)
	testModel := &testBubbleTeaModel{
		receivedMsgs: make([]tea.Msg, 0),
	}
	program := tea.NewProgram(testModel)

	err := dispatcher.Start(ctx, eventChan, program)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Cancel context to trigger shutdown
	cancel()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Dispatcher should have stopped gracefully
	// In a real implementation, we'd verify no goroutines are leaked
}

// UnsupportedEvent is a custom event type for testing unsupported events
type UnsupportedEvent struct{}

func (e UnsupportedEvent) SlackEvent() {}

// TestEventDispatcher_UnsupportedEventType tests handling of unknown events
func TestEventDispatcher_UnsupportedEventType(t *testing.T) {
	dispatcher := NewEventDispatcher()

	unsupportedEvt := UnsupportedEvent{}
	msg := dispatcher.convertSlackEvent(unsupportedEvt)

	// Should return nil or a specific "unknown event" message
	if msg != nil {
		t.Errorf("Expected nil for unsupported event, got %T", msg)
	}
}

// TestEventDispatcher_ConcurrentEvents tests handling multiple events concurrently
func TestEventDispatcher_ConcurrentEvents(t *testing.T) {
	dispatcher := NewEventDispatcher()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventChan := make(chan port.SlackEvent, 100)
	testModel := &testBubbleTeaModel{
		receivedMsgs: make([]tea.Msg, 0),
	}
	program := tea.NewProgram(testModel)

	err := dispatcher.Start(ctx, eventChan, program)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Send multiple events concurrently
	numEvents := 50
	for i := 0; i < numEvents; i++ {
		eventChan <- port.SlackConnectedEvent{}
	}

	// Give time to process
	time.Sleep(500 * time.Millisecond)

	// All events should be processed without errors
	// In a real test, we'd count received messages
}

// testBubbleTeaModel is a minimal Bubble Tea model for testing
type testBubbleTeaModel struct {
	receivedMsgs []tea.Msg
}

func (m *testBubbleTeaModel) Init() tea.Cmd {
	return nil
}

func (m *testBubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.receivedMsgs = append(m.receivedMsgs, msg)
	return m, nil
}

func (m *testBubbleTeaModel) View() string {
	return ""
}
