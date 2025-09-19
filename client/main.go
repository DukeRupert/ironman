package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	// "github.com/dukerupert/go-claude"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func Hello(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}

func upload(c echo.Context) error {
	return c.Render(http.StatusOK, "upload", nil)
}

func handleUpload(c echo.Context) error {
	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")

	//-----------
	// Read file
	//-----------

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully with fields name=%s and email=%s.</p>", file.Filename, name, email))
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// establishd database connection
	ctx := context.Background()
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
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

	e := NewEchoServer(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
