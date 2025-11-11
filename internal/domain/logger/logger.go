package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

type Format int

const (
	FormatJSON Format = iota
	FormatText
)

type Config struct {
	Level      slog.Level
	OutputPath string
	Format     Format
	AddSource  bool
}

type Logger struct {
	slog   *slog.Logger
	writer io.WriteCloser
}

func NewLogger(config Config) (*Logger, error) {
	var writer io.WriteCloser
	var output io.Writer

	if config.OutputPath == "" {
		output = os.Stdout
		writer = nil
	} else {
		if err := ensureLogDirectory(config.OutputPath); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
		writer = file
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	if config.Format == FormatJSON {
		handler = slog.NewJSONHandler(output, handlerOpts)
	} else {
		handler = slog.NewTextHandler(output, handlerOpts)
	}

	return &Logger{
		slog:   slog.New(handler),
		writer: writer,
	}, nil
}

func (l *Logger) Close() error {
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)

	if l.slog.Enabled(ctx, slog.LevelError) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			allArgs = append(allArgs, "caller_file", file, "caller_line", line)
		}
	}

	l.slog.ErrorContext(ctx, msg, allArgs...)
}

func (l *Logger) WithContext(args ...any) *Logger {
	return &Logger{
		slog:   l.slog.With(args...),
		writer: l.writer,
	}
}

func GetLogFilePath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "gosl", "logs", "gosl.log")
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	return filepath.Join(dataHome, "gosl", "logs", "gosl.log")
}

func ensureLogDirectory(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0755)
}
