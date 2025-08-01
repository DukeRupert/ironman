package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	// "github.com/dukerupert/ironman/orderspace"
	// "github.com/dukerupert/ironman/woo"
	"github.com/dukerupert/ironman/dto"
	"github.com/dukerupert/ironman/orderspace"
	"github.com/dukerupert/ironman/views"
	"github.com/dukerupert/ironman/woo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/spf13/viper"
)

type ClientConfig struct {
	OrderspaceBaseUrl      string
	OrderspaceClientID     string
	OrderspaceClientSecret string
	WooBaseUrl             string
	WooConsumerKey         string
	WooConsumerSecret      string
}

func LoadConfig() (*ClientConfig, error) {
	// Set config file name and path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Bind environment variables to config keys
	viper.BindEnv("orderspace_base_url", "ORDERSPACE_BASE_URL")
	viper.BindEnv("orderspace_client_id", "ORDERSPACE_CLIENT_ID")
	viper.BindEnv("orderspace_client_secret", "ORDERSPACE_CLIENT_SECRET")
	viper.BindEnv("woo_base_url", "WOO_BASE_URL")
	viper.BindEnv("woo_consumer_key", "WOO_CONSUMER_KEY")
	viper.BindEnv("woo_consumer_secret", "WOO_CONSUMER_SECRET")

	// Read config file (optional - env vars will override)
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, env vars can provide all values
		log.Printf("Config file not found or error reading: %v", err)
	}

	// Create and populate the config struct
	config := &ClientConfig{
		OrderspaceBaseUrl:      viper.GetString("orderspace_base_url"),
		OrderspaceClientID:     viper.GetString("orderspace_client_id"),
		OrderspaceClientSecret: viper.GetString("orderspace_client_secret"),
		WooBaseUrl:             viper.GetString("woo_base_url"),
		WooConsumerKey:         viper.GetString("woo_consumer_key"),
		WooConsumerSecret:      viper.GetString("woo_consumer_secret"),
	}

	return config, nil
}

// OrderService handles fetching and processing orders
type OrderService struct {
	wooClient        *woo.Client
	orderspaceClient *orderspace.Client
}

// NewOrderService creates a new order service
func NewOrderService(config ClientConfig) *OrderService {
	wooClient := woo.NewClient(
		config.WooBaseUrl,
		config.WooConsumerKey,
		config.WooConsumerSecret,
	)

	orderspaceClient := orderspace.NewClient(
		config.OrderspaceBaseUrl,
		config.OrderspaceClientID,
		config.OrderspaceClientSecret,
	)

	return &OrderService{
		wooClient:        wooClient,
		orderspaceClient: orderspaceClient,
	}
}

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

// ConvertWooOrder converts a WooCommerce order to UnifiedOrder
func (s *OrderService) ConvertWooOrder(order woo.Order) dto.UnifiedOrder {
	customer := strings.TrimSpace(order.Billing.FirstName + " " + order.Billing.LastName)
	if customer == "" {
		customer = order.Billing.Email
	}

	// Parse total
	total, err := strconv.ParseFloat(order.Total, 64)
	if err != nil {
		total = 0
	}

	// Format date
	orderDate := order.DateCreated
	if parsed, err := time.Parse("2006-01-02T15:04:05", order.DateCreated); err == nil {
		orderDate = parsed.Format("Jan 2, 2006")
	}

	return dto.UnifiedOrder{
		OrderNumber: "#" + strconv.Itoa(order.ID),
		Customer:    customer,
		OrderDate:   orderDate,
		DeliverOn:   "N/A",
		Total:       FormatCurrency(total, order.Currency),
		Origin:      "WooCommerce",
	}
}

// ConvertOrderspaceOrder converts an Orderspace order to UnifiedOrder
func (s *OrderService) ConvertOrderspaceOrder(order orderspace.Order) dto.UnifiedOrder {
	customer := order.CompanyName
	if customer == "" && order.BillingAddress.ContactName != "" {
		customer = order.BillingAddress.ContactName
	}

	// Format dates
	orderDate := order.Created
	if parsed, err := time.Parse("2006-01-02T15:04:05Z", order.Created); err == nil {
		orderDate = parsed.Format("Jan 2, 2006")
	}

	deliverOn := "N/A"
	if order.DeliveryDate != "" {
		if parsed, err := time.Parse("2006-01-02", order.DeliveryDate); err == nil {
			deliverOn = parsed.Format("Jan 2, 2006")
		} else {
			deliverOn = order.DeliveryDate
		}
	}

	return dto.UnifiedOrder{
		OrderNumber: "#" + strconv.Itoa(order.Number),
		Customer:    customer,
		OrderDate:   orderDate,
		DeliverOn:   deliverOn,
		Total:       FormatCurrency(order.GrossTotal, order.Currency),
		Status:      strings.Title(order.Status),
		Origin:      "Orderspace",
	}
}

// GetUnifiedOrders fetches and combines orders from both systems
func (s *OrderService) GetUnifiedOrders() ([]dto.UnifiedOrder, error) {
	slog.Info("Starting to fetch unified orders")
	var unifiedOrders []dto.UnifiedOrder

	// Fetch WooCommerce orders
	slog.Info("Fetching WooCommerce orders", "limit", 10)
	wooOrders, err := s.wooClient.ListOrders(&woo.OrderListOptions{
		Page:    1,
		PerPage: 10,
		OrderBy: "date",
		Order:   "desc",
	})
	if err != nil {
		slog.Error("Error fetching WooCommerce orders", "error", err)
	} else {
		slog.Info("WooCommerce orders fetched successfully",
			"count", len(wooOrders.Orders),
			"total_available", wooOrders.Pagination.Total)

		for i, order := range wooOrders.Orders {
			slog.Debug("Processing WooCommerce order",
				"index", i,
				"order_id", order.ID,
				"status", order.Status,
				"total", order.Total,
				"date", order.DateCreated)

			unified := s.ConvertWooOrder(order)
			unifiedOrders = append(unifiedOrders, unified)

			slog.Debug("Converted WooCommerce order",
				"unified_order", unified)
		}
	}

	// Fetch Orderspace orders
	slog.Info("Fetching Orderspace orders", "limit", 10)
	orderspaceOrders, err := s.orderspaceClient.GetLast10Orders()
	if err != nil {
		slog.Error("Error fetching Orderspace orders", "error", err)
	} else {
		slog.Info("Orderspace orders fetched successfully",
			"count", len(orderspaceOrders.Orders))

		for i, order := range orderspaceOrders.Orders {
			slog.Debug("Processing Orderspace order",
				"index", i,
				"order_id", order.ID,
				"order_number", order.Number,
				"status", order.Status,
				"total", order.GrossTotal,
				"date", order.Created)

			unified := s.ConvertOrderspaceOrder(order)
			unifiedOrders = append(unifiedOrders, unified)

			slog.Debug("Converted Orderspace order",
				"unified_order", unified)
		}
	}

	slog.Info("Unified orders processing complete",
		"total_unified_orders", len(unifiedOrders),
		"woo_orders", len(wooOrders.Orders),
		"orderspace_orders", len(orderspaceOrders.Orders))

	// Log the final unified orders for debugging
	if len(unifiedOrders) == 0 {
		slog.Warn("No unified orders found - both systems returned empty or failed")
	} else {
		slog.Info("Sample unified orders for verification")
		for i, order := range unifiedOrders {
			if i >= 3 { // Only log first 3 for brevity
				break
			}
			slog.Info("Sample unified order",
				"index", i,
				"order_number", order.OrderNumber,
				"customer", order.Customer,
				"total", order.Total,
				"origin", order.Origin)
		}
	}

	return unifiedOrders, nil
}

// Handler for the orders page
func (s *OrderService) HandleOrders(c echo.Context) error {
	slog.Info("Orders page requested", "path", c.Request().URL.Path)
	
	orders, err := s.GetUnifiedOrders()
	if err != nil {
		slog.Error("Error in HandleOrders", "error", err)
		return c.String(http.StatusInternalServerError, "Error fetching orders: "+err.Error())
	}

	slog.Info("Rendering orders page", "order_count", len(orders))
	page := views.OrdersPage(orders)
	return page.Render(c.Request().Context(), c.Response())
}

// Example usage function
func main() {
	// Set up structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set to Debug to see all logs
	})))

	slog.Info("Starting Unified Orders Dashboard")

	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize services
	slog.Info("Initializing order service")
	orderService := NewOrderService(*config)

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

	// Start server
	slog.Info("Server starting", "port", 8080)
	e.Logger.Fatal(e.Start(":8080"))
}
