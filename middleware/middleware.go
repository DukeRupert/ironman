package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type contextKey string

const (
	LoggerKey    contextKey = "logger"
	StartTimeKey contextKey = "startTime"
	UserIDKey    contextKey = "userID"
	RequestIDKey contextKey = "requestID"
)

type Middleware func(http.Handler) http.Handler

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		replaceOptions := &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Replace the message key with event_type
				if a.Key == slog.MessageKey {
					return slog.String("event_type", a.Value.String())
				}
				return a
			},
		}
		l := slog.New(slog.NewJSONHandler(os.Stdout, replaceOptions))
		logger := l.With("requestID", "").With("userID", "").With("method", r.Method).With("path", r.URL.Path)

		// Set context
		ctx := context.WithValue(r.Context(), LoggerKey, logger)
		ctx = context.WithValue(ctx, StartTimeKey, start)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)
		logger.Info("api_request", "duration", elapsed.String())
	})
}
