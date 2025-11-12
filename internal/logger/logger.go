package logger

import (
	"io"
	"log/slog"
	"strings"
	"time"
)

func New(w io.Writer, level string, environment string) *slog.Logger {
	logLevel := parseLogLevel(level)
	isDev := environment == "dev" || environment == "development"
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: isDev, // Add file:line info in development
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	var handler slog.Handler
	if isDev {
		// Pretty printed for development
		handler = slog.NewTextHandler(w, opts)
	} else {
		// Structured JSON for production
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler)
}

func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(logLevel)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default fallback
	}
}