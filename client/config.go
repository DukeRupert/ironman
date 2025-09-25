package main

import (
	"io"
	"log/slog"
	"strings"
	"time"
)

type Config struct {
	APP_HOST          string
	APP_PORT          string
	DB_HOST           string
	DB_PORT           string
	DB_USER           string
	DB_PASSWORD       string
	DB_NAME           string
	LOG_LEVEL         string // debug, info, warn, error
	ENVIRONMENT       string // prod, dev
	ANTHROPIC_API_KEY string
}

// Order of precedence from least to greatest is
// default -> environment -> flag
func GetConfig(environ []string, args []string) Config {
	// default config
	config := Config{
		APP_HOST:          "localhost",
		APP_PORT:          "8080",
		DB_HOST:           "localhost",
		DB_PORT:           "5432",
		DB_USER:           "postgres",
		DB_PASSWORD:       "",
		DB_NAME:           "postgres",
		LOG_LEVEL:         "info",
		ANTHROPIC_API_KEY: "",
	}

	if appHost := getEnv(environ, "APP_HOST"); appHost != "" {
		config.APP_HOST = appHost
	}

	if dbUser := getEnv(environ, "DB_USER"); dbUser != "" {
		config.DB_USER = dbUser
	}

	if dbPassword := getEnv(environ, "DB_PASSWORD"); dbPassword != "" {
		config.DB_PASSWORD = dbPassword
	}

	if dbName := getEnv(environ, "DB_NAME"); dbName != "" {
		config.DB_NAME = dbName
	}

	if logLevel := getEnv(environ, "LOG_LEVEL"); logLevel != "" {
		config.LOG_LEVEL = logLevel
	}

	if environment := getEnv(environ, "ENVIRONMENT"); environment != "" {
		config.ENVIRONMENT = environment
	}

	if anthropicApiKey := getEnv(environ, "ANTHROPIC_API_KEY"); anthropicApiKey != "" {
		config.ANTHROPIC_API_KEY = anthropicApiKey
	}

	// Flags
	if appHost := getFlag(args, "app_host"); appHost != "" {
		config.APP_HOST = appHost
	}

	if dbUser := getFlag(args, "db_user"); dbUser != "" {
		config.DB_USER = dbUser
	}

	if dbPassword := getFlag(args, "db_password"); dbPassword != "" {
		config.DB_PASSWORD = dbPassword
	}

	if dbName := getFlag(args, "db_name"); dbName != "" {
		config.DB_NAME = dbName
	}

	if environment := getFlag(args, "environment"); environment != "" {
		config.ENVIRONMENT = environment
	}

	if logLevel := getFlag(args, "log_level"); logLevel != "" {
		config.LOG_LEVEL = logLevel
	}

	if anthropicApiKey := getFlag(args, "anthropic_api_key"); anthropicApiKey != "" {
		config.ANTHROPIC_API_KEY = anthropicApiKey
	}


	return config
}

func SetupLogger(w io.Writer, level string, environment string) *slog.Logger {
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

// helpers
func getFlag(args []string, flag string) string {
	// Look for --flag=value format
	prefix := "--" + flag + "="
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}

	// Look for --flag value format (separate arguments)
	flagArg := "--" + flag
	for i, arg := range args {
		if arg == flagArg && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

func getEnv(environ []string, key string) string {
	prefix := key + "="
	for _, env := range environ {
		if strings.HasPrefix(env, prefix) {
			return env[len(prefix):]
		}
	}
	return ""
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