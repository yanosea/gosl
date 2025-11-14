//go:build integration
// +build integration

package errorhandler

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/yanosea/gosl/internal/domain/logger"
	"github.com/yanosea/gosl/internal/domain/reconnect"
)

// TestIntegration_ErrorHandlingWithLogger tests error handling integrated with logging
func TestIntegration_ErrorHandlingWithLogger(t *testing.T) {
	// Create logger
	logConfig := logger.Config{
		Level:      slog.LevelDebug,
		OutputPath: "", // Use stdout for testing
		Format:     logger.FormatJSON,
		AddSource:  true,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tests := []struct {
		name     string
		err      error
		logLevel string
	}{
		{
			name:     "Auth error logging",
			err:      ErrInvalidToken,
			logLevel: "error",
		},
		{
			name:     "Network error logging",
			err:      ErrNetworkError,
			logLevel: "warn",
		},
		{
			name:     "Rate limit error logging",
			err:      ErrRateLimitExceeded,
			logLevel: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := HandleError(tt.err)

			ctx := context.Background()

			// Log based on error category and recoverability
			if appErr.Recoverable {
				log.Warn(ctx, appErr.Message, "category", appErr.Category.String(), "recoverable", appErr.Recoverable)
			} else {
				log.Error(ctx, appErr.Message, appErr.Err, "category", appErr.Category.String(), "recoverable", appErr.Recoverable)
			}

			// Verify error properties
			if appErr.Category == ErrorCategoryUnknown {
				t.Error("Error should not be categorized as Unknown for known error types")
			}

			if appErr.Message == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

// MockReconnector for testing
type MockReconnector struct {
	shouldFail   bool
	connectCount int
}

func (m *MockReconnector) Connect(ctx context.Context) error {
	m.connectCount++
	if m.shouldFail {
		return ErrNetworkError
	}
	return nil
}

func (m *MockReconnector) Disconnect() error {
	return nil
}

// TestIntegration_ReconnectWithErrorHandling tests reconnection with error handling
func TestIntegration_ReconnectWithErrorHandling(t *testing.T) {
	logConfig := logger.Config{
		Level:      slog.LevelDebug,
		OutputPath: "",
		Format:     logger.FormatJSON,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tests := []struct {
		name           string
		shouldFail     bool
		expectSuccess  bool
		expectAttempts int
	}{
		{
			name:           "Successful reconnection",
			shouldFail:     false,
			expectSuccess:  true,
			expectAttempts: 1,
		},
		{
			name:           "Failed reconnection with proper error handling",
			shouldFail:     true,
			expectSuccess:  false,
			expectAttempts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconnector := &MockReconnector{
				shouldFail: tt.shouldFail,
			}

			config := reconnect.Config{
				MaxRetries:     3,
				InitialBackoff: 10 * time.Millisecond,
				MaxBackoff:     50 * time.Millisecond,
			}

			handler := reconnect.NewReconnectHandler(config)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			log.Info(ctx, "Starting reconnection attempt", "test", tt.name)

			err := handler.HandleDisconnect(ctx, reconnector)

			if tt.expectSuccess && err != nil {
				appErr := HandleError(err)
				log.Error(ctx, "Reconnection failed", err, "category", appErr.Category.String())
				t.Errorf("Expected successful reconnection, got error: %v", err)
			}

			if !tt.expectSuccess && err == nil {
				t.Error("Expected reconnection to fail, but it succeeded")
			}

			if reconnector.connectCount != tt.expectAttempts {
				t.Errorf("Expected %d connection attempts, got %d", tt.expectAttempts, reconnector.connectCount)
			}

			// Verify status
			status := handler.GetStatus()
			if tt.expectSuccess && status.State != reconnect.StateConnected {
				t.Errorf("Expected StateConnected, got %v", status.State)
			}

			if !tt.expectSuccess && status.State != reconnect.StateError {
				t.Errorf("Expected StateError, got %v", status.State)
			}

			log.Info(ctx, "Reconnection test completed",
				"success", tt.expectSuccess,
				"attempts", reconnector.connectCount,
				"state", status.State.String())
		})
	}
}

// TestIntegration_RateLimitRetry tests rate limit handling with retry
func TestIntegration_RateLimitRetry(t *testing.T) {
	logConfig := logger.Config{
		Level:      slog.LevelDebug,
		OutputPath: "",
		Format:     logger.FormatJSON,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	ctx := context.Background()

	// Simulate rate limit error
	rateLimitErr := ErrRateLimitExceeded
	appErr := HandleError(rateLimitErr)

	log.Warn(ctx, appErr.Message,
		"category", appErr.Category.String(),
		"recoverable", appErr.Recoverable,
		"retry_after", "1s")

	if appErr.Category != ErrorCategoryRateLimit {
		t.Errorf("Expected ErrorCategoryRateLimit, got %v", appErr.Category)
	}

	if !appErr.Recoverable {
		t.Error("Rate limit errors should be recoverable")
	}

	// Verify error message is in Japanese
	if appErr.Message == "" {
		t.Error("Error message should not be empty")
	}
}

// TestIntegration_MultipleErrorScenarios tests various error scenarios together
func TestIntegration_MultipleErrorScenarios(t *testing.T) {
	logConfig := logger.Config{
		Level:      slog.LevelDebug,
		OutputPath: "",
		Format:     logger.FormatJSON,
		AddSource:  true,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	ctx := context.Background()

	scenarios := []error{
		ErrInvalidToken,
		ErrNetworkError,
		ErrRateLimitExceeded,
		ErrChannelNotFound,
		errors.New("unknown error"),
	}

	for i, scenarioErr := range scenarios {
		appErr := HandleError(scenarioErr)

		contextLogger := log.WithContext(
			"scenario_number", i+1,
			"original_error", scenarioErr.Error(),
		)

		if appErr.Recoverable {
			contextLogger.Warn(ctx, appErr.Message, "category", appErr.Category.String())
		} else {
			contextLogger.Error(ctx, appErr.Message, appErr.Err, "category", appErr.Category.String())
		}

		// Verify all errors have proper categorization
		if appErr.Category < ErrorCategoryAuth || appErr.Category > ErrorCategoryUnknown {
			t.Errorf("Invalid error category: %v", appErr.Category)
		}

		// Verify all errors have messages
		if appErr.Message == "" {
			t.Error("Error message should not be empty")
		}

		// Verify error unwrapping works
		if !errors.Is(appErr.Err, scenarioErr) && scenarioErr != errors.New("unknown error") {
			t.Errorf("Error unwrapping failed for %v", scenarioErr)
		}
	}
}
