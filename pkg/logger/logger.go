package logger

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}))
	slog.SetDefault(defaultLogger)
}

// Info logs an informational message with optional key-value pairs
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}
