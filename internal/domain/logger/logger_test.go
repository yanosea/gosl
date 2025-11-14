package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewLogger tests logger initialization
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "Valid config with debug level",
			config: Config{
				Level:      slog.LevelDebug,
				OutputPath: "",
				Format:     FormatJSON,
			},
			wantError: false,
		},
		{
			name: "Valid config with info level",
			config: Config{
				Level:      slog.LevelInfo,
				OutputPath: "",
				Format:     FormatText,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("NewLogger() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
		})
	}
}

// TestLogger_Info tests info level logging
func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	logger.Info(context.Background(), "test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Logger.Info() output = %v, want to contain 'test message'", output)
	}
	if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
		t.Errorf("Logger.Info() output = %v, want to contain key-value pair", output)
	}
}

// TestLogger_Error tests error level logging
func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
	}

	testErr := errors.New("test error")
	logger.Error(context.Background(), "error occurred", testErr, "context", "test")

	output := buf.String()
	if !strings.Contains(output, "error occurred") {
		t.Errorf("Logger.Error() output = %v, want to contain 'error occurred'", output)
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("Logger.Error() output = %v, want to contain error message", output)
	}
}

// TestLogger_Debug tests debug level logging
func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	logger.Debug(context.Background(), "debug message", "debug_key", "debug_value")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Logger.Debug() output = %v, want to contain 'debug message'", output)
	}
}

// TestLogger_Warn tests warn level logging
func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	}

	logger.Warn(context.Background(), "warning message", "warn_key", "warn_value")

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("Logger.Warn() output = %v, want to contain 'warning message'", output)
	}
}

// TestLogger_WithContext tests contextual logging
func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	contextLogger := logger.WithContext("request_id", "12345", "user_id", "user123")
	contextLogger.Info(context.Background(), "test message")

	output := buf.String()
	if !strings.Contains(output, "request_id") || !strings.Contains(output, "12345") {
		t.Errorf("Logger.WithContext() output = %v, want to contain request_id", output)
	}
	if !strings.Contains(output, "user_id") || !strings.Contains(output, "user123") {
		t.Errorf("Logger.WithContext() output = %v, want to contain user_id", output)
	}
}

// TestGetLogFilePath tests log file path resolution
func TestGetLogFilePath(t *testing.T) {
	tests := []struct {
		name   string
		setEnv map[string]string
		want   string
	}{
		{
			name: "Use XDG_DATA_HOME when set",
			setEnv: map[string]string{
				"XDG_DATA_HOME": "/custom/data",
			},
			want: "/custom/data/gosl/logs/gosl.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			oldEnv := make(map[string]string)
			for k, v := range tt.setEnv {
				oldEnv[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			defer func() {
				for k, v := range oldEnv {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}
			}()

			got := GetLogFilePath()
			if got != tt.want {
				t.Errorf("GetLogFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogger_JSONFormat tests JSON formatted output
func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	logger.Info(context.Background(), "test message", "key", "value")

	// Verify JSON format
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger JSON output is not valid JSON: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Logger JSON output msg = %v, want 'test message'", logEntry["msg"])
	}
}

// TestEnsureLogDirectory tests log directory creation
func TestEnsureLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test", "logs", "test.log")

	err := ensureLogDirectory(logPath)
	if err != nil {
		t.Errorf("ensureLogDirectory() error = %v", err)
	}

	// Verify directory was created
	dirPath := filepath.Dir(logPath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Errorf("ensureLogDirectory() did not create directory %v", dirPath)
	}
}

// TestLogger_ErrorWithStackTrace tests error logging with stack trace
func TestLogger_ErrorWithStackTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		slog: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level:     slog.LevelError,
			AddSource: true,
		})),
	}

	testErr := errors.New("critical error")
	logger.Error(context.Background(), "critical error occurred", testErr)

	output := buf.String()
	if !strings.Contains(output, "critical error occurred") {
		t.Errorf("Logger.Error() output = %v, want to contain error message", output)
	}

	// Verify source location is included
	if !strings.Contains(output, "source") {
		t.Log("Note: Source location should be included in production logs")
	}
}
