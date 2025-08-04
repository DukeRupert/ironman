package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/dukerupert/ironman/orders"
	"github.com/dukerupert/ironman/views"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// FormatCurrency formats the currency amount
func FormatCurrency(amount float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "USD":
		return "$" + strconv.FormatFloat(amount, 'f', 2, 64)
	case "GBP":
		return "£" + strconv.FormatFloat(amount, 'f', 2, 64)
	case "EUR":
		return "€" + strconv.FormatFloat(amount, 'f', 2, 64)
	default:
		return strconv.FormatFloat(amount, 'f', 2, 64) + " " + currency
	}
}

// Example usage function
func main() {
	// Set up structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set to Debug to see all logs
	})))

	slog.Info("Starting Unified Orders Dashboard")

	// Initialize services
	slog.Info("Initializing order service")
	orderService := orders.NewService()

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/static", "static")

	e.GET("/", func(c echo.Context) error {
		page := views.HomePage()
		return page.Render(context.Background(), c.Response())
	})

	e.GET("/orders", orderService.HandleOrders)
	e.GET("/orders/:id", orderService.HandleOrder)
	e.POST("/refresh-orders", func(c echo.Context) error {
		slog.Info("Manual refresh requested")
		if err := orderService.RefreshCache(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"status": "Cache refreshed successfully",
		})
	})

	e.GET("/test-orderspace", func(c echo.Context) error {
		slog.Info("Testing Orderspace API connection")

		// Test listing orders first
		ordersResponse, err := orderService.OrderSpaceClient.GetAllOrders(5, "")
		if err != nil {
			slog.Error("Failed to list orders", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}

		slog.Info("Successfully listed orders", "count", len(ordersResponse.Orders))

		// If we have orders, try to get the details of the first one
		if len(ordersResponse.Orders) > 0 {
			firstOrder := ordersResponse.Orders[0]
			slog.Info("Trying to get details for first order",
				"order_id", firstOrder.ID,
				"order_number", firstOrder.Number)

			detailedOrder, err := orderService.OrderSpaceClient.GetOrder(firstOrder.ID)
			if err != nil {
				slog.Error("Failed to get order details", "order_id", firstOrder.ID, "error", err)
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"list_success":   true,
					"orders_count":   len(ordersResponse.Orders),
					"detail_error":   err.Error(),
					"first_order_id": firstOrder.ID,
				})
			}

			slog.Info("Successfully got order details",
				"order_id", detailedOrder.ID,
				"order_number", detailedOrder.Number,
				"company_name", detailedOrder.CompanyName)

			return c.JSON(http.StatusOK, map[string]interface{}{
				"list_success":   true,
				"detail_success": true,
				"orders_count":   len(ordersResponse.Orders),
				"first_order": map[string]interface{}{
					"id":           detailedOrder.ID,
					"number":       detailedOrder.Number,
					"company_name": detailedOrder.CompanyName,
					"status":       detailedOrder.Status,
				},
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"list_success": true,
			"orders_count": 0,
			"message":      "No orders found to test detail endpoint",
		})
	})

	// Start server
	slog.Info("Server starting", "port", 8080)
	e.Logger.Fatal(e.Start(":8080"))
}
