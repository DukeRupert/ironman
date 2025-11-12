package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	// "github.com/dukerupert/go-claude"
	"github.com/dukerupert/ironman/api/v1"
	"github.com/dukerupert/ironman/internal/config"
	"github.com/dukerupert/ironman/internal/logger"
	"github.com/jackc/pgx/v5"
)



func run(ctx context.Context, w io.Writer, environ []string, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config := config.GetConfig(environ, args)

	logger := logger.New(w, config.LOG_LEVEL, config.ENVIRONMENT)

	// Debug environment
	logger.Debug("config", "environ", config, "args", args)
	
	// establish database connection
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.DB_USER, config.DB_PASSWORD, config.DB_HOST, config.DB_PORT, config.DB_NAME)
	db, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)

	err = db.Ping(ctx)
	if err != nil {
		panic(err)
	}

	logger.Info("database connection established...")

	srv := v1.NewServer(logger)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.APP_HOST, config.APP_PORT),
		Handler: srv,
	}
	var wg sync.WaitGroup

	// Start the HTTP server
	wg.Go(func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	})

	// Handle graceful shutdown
	wg.Go(func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})

	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Environ(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// Create client with API key from environment
// anthropicApiKey := os.Getenv("ANTHROPIC_API_KEY")
// if anthropicApiKey == "" {
// 	log.Fatal("Missing anthropic api key")
// }
// client := anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

// // Simple message
// resp, err := client.SimpleMessage(
// 	context.Background(),
// 	"claude-3-opus-20240229",
// 	"Hello, Claude!",
// 	4096,
// )
// if err != nil {
// 	log.Fatal(err)
// }

// fmt.Printf("Response: %s\n", resp.GetText())
// fmt.Printf("Tokens used: %d input, %d output\n",
// 	resp.Usage.InputTokens, resp.Usage.OutputTokens)
