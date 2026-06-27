// Package logger provides a thin, structured logging setup built on the
// standard library's log/slog. JSON output in production (machine-parseable
// for log aggregation), human-friendly text in development.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ctxKey is unexported so only this package can attach the request logger.
type ctxKey struct{}

// New builds the root logger. level is one of debug|info|warn|error;
// format is console|json.
func New(level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	if strings.EqualFold(format, "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext stores a logger (typically enriched with a request id) in ctx.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the request-scoped logger, or the default logger if none
// was attached. It never returns nil.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
