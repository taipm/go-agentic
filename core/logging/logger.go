package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	// Global logger instance (structured JSON by default)
	loggerMu sync.RWMutex
	logger   *slog.Logger
)

func init() {
	// Default: JSON logger to stdout
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
}

// SetLogger allows custom logger configuration
func SetLogger(l *slog.Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger = l
}

// GetLogger returns the global logger
func GetLogger() *slog.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return logger
}

// SetOutput changes the output destination (for testing)
func SetOutput(w io.Writer) {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger = slog.New(handler)
}

// SetLevel changes the log level
func SetLevel(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger = slog.New(handler)
}

// WithTraceID adds trace_id to context for log correlation
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID extracts trace_id from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

type contextKey string

const traceIDKey contextKey = "trace_id"
