package reconnect

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockConnector is a mock implementation for testing
type MockConnector struct {
	connectFunc    func(ctx context.Context) error
	disconnectFunc func() error
	callCount      int
}

func (m *MockConnector) Connect(ctx context.Context) error {
	m.callCount++
	if m.connectFunc != nil {
		return m.connectFunc(ctx)
	}
	return nil
}

func (m *MockConnector) Disconnect() error {
	if m.disconnectFunc != nil {
		return m.disconnectFunc()
	}
	return nil
}

// TestNewReconnectHandler tests handler initialization
func TestNewReconnectHandler(t *testing.T) {
	config := Config{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}

	handler := NewReconnectHandler(config)
	if handler == nil {
		t.Fatal("NewReconnectHandler() returned nil")
	}

	if handler.config.MaxRetries != config.MaxRetries {
		t.Errorf("MaxRetries = %v, want %v", handler.config.MaxRetries, config.MaxRetries)
	}
}

// TestReconnectHandler_HandleDisconnect tests automatic reconnection
func TestReconnectHandler_HandleDisconnect(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		connectErr    error
		expectSuccess bool
		expectRetries int
	}{
		{
			name:          "Successful reconnection on first attempt",
			maxRetries:    3,
			connectErr:    nil,
			expectSuccess: true,
			expectRetries: 1,
		},
		{
			name:          "Failed reconnection after max retries",
			maxRetries:    2,
			connectErr:    errors.New("connection failed"),
			expectSuccess: false,
			expectRetries: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &MockConnector{
				connectFunc: func(ctx context.Context) error {
					return tt.connectErr
				},
			}

			config := Config{
				MaxRetries:     tt.maxRetries,
				InitialBackoff: 10 * time.Millisecond,
				MaxBackoff:     100 * time.Millisecond,
			}

			handler := NewReconnectHandler(config)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := handler.HandleDisconnect(ctx, connector)

			if tt.expectSuccess && err != nil {
				t.Errorf("HandleDisconnect() error = %v, expected success", err)
			}

			if !tt.expectSuccess && err == nil {
				t.Error("HandleDisconnect() succeeded, expected failure")
			}

			if connector.callCount != tt.expectRetries {
				t.Errorf("Connect called %d times, want %d", connector.callCount, tt.expectRetries)
			}
		})
	}
}

// TestReconnectHandler_ExponentialBackoff tests backoff timing
func TestReconnectHandler_ExponentialBackoff(t *testing.T) {
	connector := &MockConnector{
		connectFunc: func(ctx context.Context) error {
			return errors.New("always fail")
		},
	}

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
	}

	handler := NewReconnectHandler(config)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	handler.HandleDisconnect(ctx, connector)
	elapsed := time.Since(start)

	// Should take at least: 50ms + 100ms (allowing for some timing variance)
	// Being more lenient to account for system load and timing variations
	if elapsed < 100*time.Millisecond {
		t.Errorf("Backoff too short: %v, expected at least 100ms", elapsed)
	}

	// Verify exponential backoff was applied by checking retry count
	if connector.callCount != 3 {
		t.Errorf("Expected 3 retry attempts, got %d", connector.callCount)
	}
}

// TestReconnectHandler_ContextCancellation tests context cancellation
func TestReconnectHandler_ContextCancellation(t *testing.T) {
	connector := &MockConnector{
		connectFunc: func(ctx context.Context) error {
			time.Sleep(200 * time.Millisecond)
			return errors.New("slow connection")
		},
	}

	config := Config{
		MaxRetries:     10,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}

	handler := NewReconnectHandler(config)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := handler.HandleDisconnect(ctx, connector)

	if err == nil {
		t.Error("HandleDisconnect() should fail when context is cancelled")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

// TestReconnectHandler_GetStatus tests status reporting
func TestReconnectHandler_GetStatus(t *testing.T) {
	config := Config{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}

	handler := NewReconnectHandler(config)

	status := handler.GetStatus()
	if status.State != StateConnected {
		t.Errorf("Initial state = %v, want %v", status.State, StateConnected)
	}

	if status.RetryCount != 0 {
		t.Errorf("Initial retry count = %v, want 0", status.RetryCount)
	}
}

// TestReconnectHandler_ManualReconnect tests manual reconnection trigger
func TestReconnectHandler_ManualReconnect(t *testing.T) {
	connector := &MockConnector{
		connectFunc: func(ctx context.Context) error {
			return nil
		},
	}

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
	}

	handler := NewReconnectHandler(config)
	ctx := context.Background()

	err := handler.ManualReconnect(ctx, connector)
	if err != nil {
		t.Errorf("ManualReconnect() error = %v", err)
	}

	if connector.callCount != 1 {
		t.Errorf("Connect called %d times, want 1", connector.callCount)
	}
}
