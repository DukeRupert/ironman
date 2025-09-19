package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	e.Static("/static", "public/static")
	e.GET("/", handleGetLandingPage)
	e.GET("/login", handleGetLoginPage)
	e.GET("/signup", handleGetSignupPage)
	e.GET("/forgot-password", handleForgotPasswordPage)
	e.POST("/forgot-password", handleForgotPassword)
	e.GET("/reset-password", handleResetPasswordPage)
	e.POST("/reset-password", handleResetPassword)
	e.GET("/hello", Hello)
	e.GET("/upload", upload)
	e.POST("/upload", handleUpload)
}

func handleGetLandingPage(c echo.Context) error {
	return c.Render(http.StatusOK, "landing", nil)
}

func handleGetLoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func handleGetSignupPage(c echo.Context) error {
	return c.Render(http.StatusOK, "signup", nil)
}

func handleForgotPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "forgot-password", nil)
}

func handleForgotPassword(c echo.Context) error {
	return c.String(http.StatusOK, "FIXME: Send email with reset link")
}

func handleResetPasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "reset-password", nil)
}

func handleResetPassword(c echo.Context) error {
	return c.String(http.StatusOK, "FIXME: Update password")
}

func GlobalMiddleware(e *echo.Echo, logger *slog.Logger) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
}
