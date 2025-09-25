package main

import (
	"log/slog"
	"net/http"

	"github.com/labstack/gommon/log"
)

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	tr, err := NewTemplate()
	if err != nil {
		log.Fatal("failed to create template", err)
	}
	addRoutes(logger, mux, tr)
	var handler http.Handler = mux
	// Middleware here
	Logging := NewLoggingMiddleware(logger)
	handler = Logging(handler)
	return handler
}

// func NewEchoServer(logger *slog.Logger) *echo.Echo {
// 	e := echo.New()
// 	t, err := NewTemplate()
// 	if err != nil {
// 		e.Logger.Fatal("failed to create template", err)
// 	}
// 	e.Renderer = t
// 	e.Logger.SetLevel(log.INFO)
// 	// global middleware
// 	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
// 		LogStatus:   true,
// 		LogURI:      true,
// 		LogError:    true,
// 		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
// 		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
// 			if v.Error == nil {
// 				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
// 					slog.String("uri", v.URI),
// 					slog.Int("status", v.Status),
// 				)
// 			} else {
// 				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
// 					slog.String("uri", v.URI),
// 					slog.Int("status", v.Status),
// 					slog.String("err", v.Error.Error()),
// 				)
// 			}
// 			return nil
// 		},
// 	}))
// 	RegisterRoutes(e)
// 	return e
// }
