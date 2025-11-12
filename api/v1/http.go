package v1

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/dukerupert/ironman/web/templates"
)

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	tr, err := templates.NewTemplate()
	if err != nil {
		log.Fatal("failed to create template", err)
	}
	addRoutes(mux, tr)
	handler := addGlobalMiddleware(mux, logger)
	return handler
}

func addGlobalMiddleware(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	var handler http.Handler
	handler = mux

	loggingMiddleware := NewLogging(logger)
	handler = RequestID(loggingMiddleware(handler))
	return handler
}
