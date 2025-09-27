package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/dukerupert/ironman/middleware"
)

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	tr, err := NewTemplate()
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

	loggingMiddleware := middleware.NewLogging(logger)
	handler = middleware.RequestID(loggingMiddleware(handler))
	return handler
}
