package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dukerupert/ironman/dto"
	"github.com/dukerupert/ironman/orderspace"
	"github.com/dukerupert/ironman/views"
	"github.com/dukerupert/ironman/woo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	titleCaser       cases.Caser
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

	// Create a title caser for English
	titleCaser := cases.Title(language.English)

	return &OrderService{
		wooClient:        wooClient,
		orderspaceClient: orderspaceClient,
		titleCaser:       titleCaser,
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

	// Parse date for sorting
	sortDate, err := time.Parse("2006-01-02T15:04:05", order.DateCreated)
	if err != nil {
		slog.Warn("Failed to parse WooCommerce date for sorting", "date", order.DateCreated, "error", err)
		sortDate = time.Now() // Fallback to current time
	}

	// Format date for display
	orderDate := order.DateCreated
	if err == nil {
		orderDate = sortDate.Format("Jan 2, 2006")
	}

	return dto.UnifiedOrder{
		OrderNumber: "#" + strconv.Itoa(order.ID),
		Customer:    customer,
		OrderDate:   orderDate,
		DeliverOn:   "N/A",
		Total:       FormatCurrency(total, order.Currency),
		Status:      s.titleCaser.String(order.Status),
		Origin:      "WooCommerce",
		SortDate:    sortDate,
	}
}

// ConvertOrderspaceOrder converts an Orderspace order to UnifiedOrder
func (s *OrderService) ConvertOrderspaceOrder(order orderspace.Order) dto.UnifiedOrder {
	customer := order.CompanyName
	if customer == "" && order.BillingAddress.ContactName != "" {
		customer = order.BillingAddress.ContactName
	}

	// Parse date for sorting
	sortDate, err := time.Parse("2006-01-02T15:04:05Z", order.Created)
	if err != nil {
		slog.Warn("Failed to parse Orderspace date for sorting", "date", order.Created, "error", err)
		sortDate = time.Now() // Fallback to current time
	}

	// Format date for display
	orderDate := order.Created
	if err == nil {
		orderDate = sortDate.Format("Jan 2, 2006")
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
		Status:      s.titleCaser.String(order.Status),
		Origin:      "Orderspace",
		SortDate:    sortDate,
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

	// Sort all orders by date (most recent first)
	slog.Info("Sorting unified orders by date (most recent first)")
	sort.Slice(unifiedOrders, func(i, j int) bool {
		return unifiedOrders[i].SortDate.After(unifiedOrders[j].SortDate)
	})

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

// GetUnifiedOrdersPaginated fetches and combines orders from both systems with pagination
func (s *OrderService) GetUnifiedOrdersPaginated(page, perPage int) (*dto.PaginatedOrders, error) {
	slog.Info("Starting to fetch paginated unified orders", "page", page, "per_page", perPage)
	
	// For now, we'll fetch more orders than needed and paginate in memory
	// In a production system, you'd want to implement proper API pagination
	fetchLimit := perPage * 5 // Fetch more to ensure we have enough after merging

	var unifiedOrders []dto.UnifiedOrder

	// Fetch WooCommerce orders
	slog.Info("Fetching WooCommerce orders", "limit", fetchLimit)
	wooOrders, err := s.wooClient.ListOrders(&woo.OrderListOptions{
		Page:    1,
		PerPage: fetchLimit,
		OrderBy: "date",
		Order:   "desc",
	})
	if err != nil {
		slog.Error("Error fetching WooCommerce orders", "error", err)
	} else {
		slog.Info("WooCommerce orders fetched successfully", 
			"count", len(wooOrders.Orders),
			"total_available", wooOrders.Pagination.Total)
		
		for _, order := range wooOrders.Orders {
			unified := s.ConvertWooOrder(order)
			unifiedOrders = append(unifiedOrders, unified)
		}
	}

	// Fetch Orderspace orders
	slog.Info("Fetching Orderspace orders", "limit", fetchLimit)
	orderspaceOrders, err := s.orderspaceClient.GetAllOrders(fetchLimit, "")
	if err != nil {
		slog.Error("Error fetching Orderspace orders", "error", err)
	} else {
		slog.Info("Orderspace orders fetched successfully", 
			"count", len(orderspaceOrders.Orders))
		
		for _, order := range orderspaceOrders.Orders {
			unified := s.ConvertOrderspaceOrder(order)
			unifiedOrders = append(unifiedOrders, unified)
		}
	}

	// Sort all orders by date (most recent first)
	slog.Info("Sorting unified orders by date (most recent first)")
	sort.Slice(unifiedOrders, func(i, j int) bool {
		return unifiedOrders[i].SortDate.After(unifiedOrders[j].SortDate)
	})

	totalOrders := len(unifiedOrders)
	totalPages := (totalOrders + perPage - 1) / perPage
	
	// Calculate pagination bounds
	start := (page - 1) * perPage
	end := start + perPage
	
	if start >= totalOrders {
		start = totalOrders
	}
	if end > totalOrders {
		end = totalOrders
	}
	
	// Slice the orders for this page
	var pageOrders []dto.UnifiedOrder
	if start < end {
		pageOrders = unifiedOrders[start:end]
	}

	result := &dto.PaginatedOrders{
		Orders:      pageOrders,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalOrders: totalOrders,
		PerPage:     perPage,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}

	slog.Info("Pagination complete", 
		"page", page,
		"total_pages", totalPages,
		"total_orders", totalOrders,
		"orders_on_page", len(pageOrders),
		"has_prev", result.HasPrev,
		"has_next", result.HasNext)

	return result, nil
}

// Handler for the orders page
func (s *OrderService) HandleOrders(c echo.Context) error {
	slog.Info("Orders page requested", "path", c.Request().URL.Path)
	
	// Get pagination parameters from query string
	pageParam := c.QueryParam("page")
	page := 1
	if pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	
	perPageParam := c.QueryParam("per_page")
	perPage := 10 // Default per page
	if perPageParam != "" {
		if pp, err := strconv.Atoi(perPageParam); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}
	
	slog.Info("Pagination parameters", "page", page, "per_page", perPage)
	
	paginatedOrders, err := s.GetUnifiedOrdersPaginated(page, perPage)
	if err != nil {
		slog.Error("Error in HandleOrders", "error", err)
		return c.String(http.StatusInternalServerError, "Error fetching orders: "+err.Error())
	}

	slog.Info("Rendering orders page", "order_count", len(paginatedOrders.Orders))

	ordersPage := views.OrdersPage(*paginatedOrders)
	return ordersPage.Render(c.Request().Context(), c.Response())
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
