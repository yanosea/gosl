package reconnect

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type State int

const (
	StateConnected State = iota
	StateDisconnected
	StateReconnecting
	StateError
)

func (s State) String() string {
	switch s {
	case StateConnected:
		return "Connected"
	case StateDisconnected:
		return "Disconnected"
	case StateReconnecting:
		return "Reconnecting"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type Config struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type Connector interface {
	Connect(ctx context.Context) error
	Disconnect() error
}

type Status struct {
	State          State
	RetryCount     int
	LastAttempt    time.Time
	LastError      error
	NextRetryDelay time.Duration
}

type ReconnectHandler struct {
	config Config
	mu     sync.RWMutex
	status Status
}

func NewReconnectHandler(config Config) *ReconnectHandler {
	return &ReconnectHandler{
		config: config,
		status: Status{
			State:      StateConnected,
			RetryCount: 0,
		},
	}
}

func (h *ReconnectHandler) HandleDisconnect(ctx context.Context, connector Connector) error {
	h.mu.Lock()
	h.status.State = StateReconnecting
	h.status.RetryCount = 0
	h.mu.Unlock()

	backoff := h.config.InitialBackoff

	for attempt := 0; attempt < h.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			h.status.State = StateError
			h.status.LastError = ctx.Err()
			h.mu.Unlock()
			return ctx.Err()
		default:
		}

		h.mu.Lock()
		h.status.RetryCount = attempt + 1
		h.status.LastAttempt = time.Now()
		h.status.NextRetryDelay = backoff
		h.mu.Unlock()

		err := connector.Connect(ctx)
		if err == nil {
			h.mu.Lock()
			h.status.State = StateConnected
			h.status.LastError = nil
			h.mu.Unlock()
			return nil
		}

		h.mu.Lock()
		h.status.LastError = err
		h.mu.Unlock()

		if attempt < h.config.MaxRetries-1 {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				h.mu.Lock()
				h.status.State = StateError
				h.status.LastError = ctx.Err()
				h.mu.Unlock()
				return ctx.Err()
			}

			backoff *= 2
			if backoff > h.config.MaxBackoff {
				backoff = h.config.MaxBackoff
			}
		}
	}

	h.mu.Lock()
	h.status.State = StateError
	h.mu.Unlock()

	return fmt.Errorf("reconnection failed after %d attempts: %w", h.config.MaxRetries, h.status.LastError)
}

func (h *ReconnectHandler) ManualReconnect(ctx context.Context, connector Connector) error {
	h.mu.Lock()
	h.status.State = StateReconnecting
	h.status.RetryCount = 0
	h.status.LastAttempt = time.Now()
	h.mu.Unlock()

	err := connector.Connect(ctx)

	h.mu.Lock()
	defer h.mu.Unlock()

	if err == nil {
		h.status.State = StateConnected
		h.status.LastError = nil
		return nil
	}

	h.status.State = StateError
	h.status.LastError = err
	return fmt.Errorf("manual reconnection failed: %w", err)
}

func (h *ReconnectHandler) GetStatus() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}
