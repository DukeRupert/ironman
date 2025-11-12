package v1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const (
	LoggerKey    contextKey = "logger"
	StartTimeKey contextKey = "startTime"
	UserIDKey    contextKey = "userID"
	RequestIDKey contextKey = "requestID"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func NewLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Context().Value(RequestIDKey).(string)

			// Create request-scoped logger with context
			requestLogger := logger.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Add logger to request context
			ctx := context.WithValue(r.Context(), LoggerKey, requestLogger)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w}

			requestLogger.Info("request started")

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			requestLogger.Info("request completed",
				"status", wrapped.status,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrapped.size,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func generateRequestID() string {
	bytes := make([]byte, 8) // 16 character hex string
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
