package tui

import (
	"context"
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yanosea/gosl/internal/app/port"
)

type EventDispatcher struct{
	program *tea.Program
	mu      sync.RWMutex
	running bool
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

func (d *EventDispatcher) Start(ctx context.Context, eventChan <-chan port.SlackEvent, program *tea.Program) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("event dispatcher is already running")
	}
	d.program = program
	d.running = true
	d.mu.Unlock()

	go d.eventLoop(ctx, eventChan)

	return nil
}

func (d *EventDispatcher) eventLoop(ctx context.Context, eventChan <-chan port.SlackEvent) {
	defer func() {
		d.mu.Lock()
		d.running = false
		d.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case slackEvent, ok := <-eventChan:
			if !ok {
				return
			}

			msg := d.convertSlackEvent(slackEvent)
			if msg != nil {
				d.mu.RLock()
				program := d.program
				d.mu.RUnlock()

				if program != nil {
					program.Send(msg)
				}
			}
		}
	}
}

func (d *EventDispatcher) convertSlackEvent(evt port.SlackEvent) tea.Msg {
	switch e := evt.(type) {
	case port.SlackConnectedEvent:
		return SlackConnectedMsg{}

	case port.SlackDisconnectedEvent:
		return SlackDisconnectedMsg{
			Reason: e.Reason,
		}

	case port.NewMessageEvent:
		return NewMessageMsg{
			ChannelID: e.ChannelID,
			Message:   e.Message,
		}

	case port.ChannelUpdateEvent:
		return ChannelUpdateMsg{
			Channel: e.Channel,
		}

	case port.UserTypingEvent:
		return UserTypingMsg{
			ChannelID: e.ChannelID,
			UserID:    e.UserID,
		}

	default:
		return nil
	}
}

func (d *EventDispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.running = false
	d.program = nil
}

func (d *EventDispatcher) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}
