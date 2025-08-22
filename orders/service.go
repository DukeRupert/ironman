// Package orders provides order management and caching functionality
package orders

import (
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dukerupert/ironman/orderspace"
	"github.com/dukerupert/ironman/woo"
	"github.com/spf13/viper"
)

// Service handles fetching and processing orders with caching
type Service struct {
	wooClient        *woo.Client
	orderspaceClient *orderspace.Client
	titleCaser       cases.Caser
	OrderSpaceClient *orderspace.Client

	// Cache fields
	cachedOrders  []UnifiedOrder
	lastFetched   time.Time
	cacheDuration time.Duration
	mutex         sync.RWMutex
	refreshing    bool
}

// UnifiedOrder represents an order from either system for display
type UnifiedOrder struct {
	ID          string
	OrderNumber int
	Customer    string
	OrderDate   string
	DeliverOn   string
	Total       string
	Status      string
	Origin      string
	SortDate    time.Time // Added for sorting purposes
}

// PaginatedOrders represents paginated order results
type PaginatedOrders struct {
	Orders      []UnifiedOrder
	CurrentPage int
	TotalPages  int
	TotalOrders int
	PerPage     int
	HasPrev     bool
	HasNext     bool
}

// PageData represents a page link or ellipsis in pagination
type PageData struct {
	Page       int
	IsCurrent  bool
	IsEllipsis bool
}

type ClientConfig struct {
	OrderspaceBaseURL      string
	OrderspaceClientID     string
	OrderspaceClientSecret string
	WooBaseURL             string
	WooConsumerKey         string
	WooConsumerSecret      string
}

func loadConfig() (*ClientConfig, error) {
	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Set the config file name (without extension) and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env") // Specify "env" for .env files

	// Add the path where the .env file is located (e.g., current directory)
	viper.AddConfigPath(".")

	// Bind environment variables to config keys
	viper.BindEnv("orderspace_base_url", "ORDERSPACE_BASE_URL")
	viper.BindEnv("orderspace_client_id", "ORDERSPACE_CLIENT_ID")
	viper.BindEnv("orderspace_client_secret", "ORDERSPACE_CLIENT_SECRET")
	viper.BindEnv("woo_base_url", "WOO_BASE_URL")
	viper.BindEnv("woo_consumer_key", "WOO_CONSUMER_KEY")
	viper.BindEnv("woo_consumer_secret", "WOO_CONSUMER_SECRET")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if .env is optional
			log.Println("No .env file found, relying on environment variables or defaults.")
		} else {
			// Handle other errors reading the config file
			log.Fatalf("Error reading config file: %s", err)
		}
	}

	// Create and populate the config struct
	config := &ClientConfig{
		OrderspaceBaseURL:      viper.GetString("orderspace_base_url"),
		OrderspaceClientID:     viper.GetString("orderspace_client_id"),
		OrderspaceClientSecret: viper.GetString("orderspace_client_secret"),
		WooBaseURL:             viper.GetString("woo_base_url"),
		WooConsumerKey:         viper.GetString("woo_consumer_key"),
		WooConsumerSecret:      viper.GetString("woo_consumer_secret"),
	}

	return config, nil
}

// NewService creates a new order service with caching
func NewService() *Service {
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	wooClient := woo.NewClient(
		config.WooBaseURL,
		config.WooConsumerKey,
		config.WooConsumerSecret,
	)

	orderspaceClient := orderspace.NewClient(
		config.OrderspaceBaseURL,
		config.OrderspaceClientID,
		config.OrderspaceClientSecret,
	)

	// Create a title caser for English
	titleCaser := cases.Title(language.English)

	service := &Service{
		wooClient:        wooClient,
		orderspaceClient: orderspaceClient,
		OrderSpaceClient: orderspaceClient,
		titleCaser:       titleCaser,
		cacheDuration:    5 * time.Minute, // Cache for 5 minutes
	}

	// Start background refresh routine
	service.startBackgroundRefresh()

	slog.Info("Order service initialized")
	return service
}

// FormatCurrency formats the currency amount based on the currency
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
func (s *Service) ConvertWooOrder(order woo.Order) UnifiedOrder {
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

	return UnifiedOrder{
		ID:          strconv.Itoa(order.ID),
		OrderNumber: order.ID,
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
func (s *Service) ConvertOrderspaceOrder(order orderspace.Order) UnifiedOrder {
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

	return UnifiedOrder{
		ID:          order.ID,
		OrderNumber: order.Number,
		Customer:    customer,
		OrderDate:   orderDate,
		DeliverOn:   deliverOn,
		Total:       FormatCurrency(order.GrossTotal, order.Currency),
		Status:      s.titleCaser.String(order.Status),
		Origin:      "Orderspace",
		SortDate:    sortDate,
	}
}

// GetUnifiedOrdersPaginated fetches orders from cache or refreshes if needed
func (s *Service) GetUnifiedOrdersPaginated(page, perPage int) (*PaginatedOrders, error) {
	slog.Info("Getting paginated orders", "page", page, "per_page", perPage)

	// Check if cache needs refresh
	s.mutex.RLock()
	needsRefresh := time.Since(s.lastFetched) > s.cacheDuration || len(s.cachedOrders) == 0
	cacheSize := len(s.cachedOrders)
	lastFetched := s.lastFetched
	isRefreshing := s.refreshing
	s.mutex.RUnlock()

	slog.Info("Cache status",
		"needs_refresh", needsRefresh,
		"cache_size", cacheSize,
		"last_fetched", lastFetched.Format("15:04:05"),
		"cache_age_seconds", int(time.Since(lastFetched).Seconds()),
		"is_refreshing", isRefreshing)

	// Refresh cache if needed (but don't refresh if already refreshing)
	if needsRefresh && !isRefreshing {
		slog.Info("Cache refresh needed, fetching fresh data")
		if err := s.refreshCache(); err != nil {
			slog.Error("Failed to refresh cache", "error", err)
			// Continue with stale cache if available
			if cacheSize == 0 {
				return nil, fmt.Errorf("no cached data available and refresh failed: %w", err)
			}
		}
	} else if needsRefresh && isRefreshing {
		slog.Info("Cache refresh already in progress, using existing cache")
	} else {
		slog.Info("Using cached data", "cache_age_seconds", int(time.Since(lastFetched).Seconds()))
	}

	// Paginate from cache
	return s.paginateFromCache(page, perPage), nil
}

// refreshCache fetches fresh data from both APIs and updates the cache
func (s *Service) refreshCache() error {
	s.mutex.Lock()
	s.refreshing = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.refreshing = false
		s.mutex.Unlock()
	}()

	slog.Info("Starting cache refresh")
	start := time.Now()

	var unifiedOrders []UnifiedOrder
	fetchLimit := 50 // Fetch more orders for better cache

	// Fetch WooCommerce orders
	slog.Info("Fetching WooCommerce orders for cache", "limit", fetchLimit)
	wooOrders, err := s.wooClient.ListOrders(&woo.OrderListOptions{
		Page:    1,
		PerPage: fetchLimit,
		OrderBy: "date",
		Order:   "desc",
	})
	if err != nil {
		slog.Error("Error fetching WooCommerce orders for cache", "error", err)
	} else {
		slog.Info("WooCommerce orders fetched for cache", "count", len(wooOrders.Orders))
		for _, order := range wooOrders.Orders {
			unified := s.ConvertWooOrder(order)
			unifiedOrders = append(unifiedOrders, unified)
		}
	}

	// Fetch Orderspace orders
	slog.Info("Fetching Orderspace orders for cache", "limit", fetchLimit)
	orderspaceOrders, err := s.orderspaceClient.GetAllOrders(fetchLimit, "")
	if err != nil {
		slog.Error("Error fetching Orderspace orders for cache", "error", err)
	} else {
		slog.Info("Orderspace orders fetched for cache", "count", len(orderspaceOrders.Orders))
		for _, order := range orderspaceOrders.Orders {
			unified := s.ConvertOrderspaceOrder(order)
			unifiedOrders = append(unifiedOrders, unified)
		}
	}

	// Sort all orders by date (most recent first)
	sort.Slice(unifiedOrders, func(i, j int) bool {
		return unifiedOrders[i].SortDate.After(unifiedOrders[j].SortDate)
	})

	// Update cache
	s.mutex.Lock()
	s.cachedOrders = unifiedOrders
	s.lastFetched = time.Now()
	s.mutex.Unlock()

	duration := time.Since(start)
	slog.Info("Cache refresh completed",
		"total_orders", len(unifiedOrders),
		"duration_ms", duration.Milliseconds(),
		"next_refresh", time.Now().Add(s.cacheDuration).Format("15:04:05"))

	return nil
}

// paginateFromCache paginates the cached orders
func (s *Service) paginateFromCache(page, perPage int) *PaginatedOrders {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	totalOrders := len(s.cachedOrders)
	totalPages := (totalOrders + perPage - 1) / perPage

	if totalPages == 0 {
		totalPages = 1
	}

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
	var pageOrders []UnifiedOrder
	if start < end && start < len(s.cachedOrders) {
		pageOrders = s.cachedOrders[start:end]
	}

	return &PaginatedOrders{
		Orders:      pageOrders,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalOrders: totalOrders,
		PerPage:     perPage,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}
}

// startBackgroundRefresh starts a background goroutine to refresh cache periodically
func (s *Service) startBackgroundRefresh() {
	go func() {
		// Initial fetch
		slog.Info("Starting initial cache load")
		if err := s.refreshCache(); err != nil {
			slog.Error("Initial cache load failed", "error", err)
		}

		// Set up periodic refresh
		ticker := time.NewTicker(s.cacheDuration)
		defer ticker.Stop()

		for range ticker.C {
			slog.Info("Background cache refresh triggered")
			if err := s.refreshCache(); err != nil {
				slog.Error("Background cache refresh failed", "error", err)
			}
		}
	}()
}

// RefreshCache manually refreshes the cache (for manual refresh button)
func (s *Service) RefreshCache() error {
	slog.Info("Manual cache refresh requested")
	return s.refreshCache()
}

// GetCacheInfo returns information about the cache state
func (s *Service) GetCacheInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"cached_orders":     len(s.cachedOrders),
		"last_fetched":      s.lastFetched.Format("2006-01-02 15:04:05"),
		"cache_age_seconds": int(time.Since(s.lastFetched).Seconds()),
		"next_refresh":      s.lastFetched.Add(s.cacheDuration).Format("15:04:05"),
		"is_refreshing":     s.refreshing,
	}
}

// GetPageNumbersData generates page data for template iteration
func GetPageNumbersData(currentPage, totalPages int) []PageData {
	var pages []PageData

	// Show pages around current page
	start := currentPage - 2
	end := currentPage + 2

	if start < 1 {
		start = 1
	}
	if end > totalPages {
		end = totalPages
	}

	// Always show first page if not in range
	if start > 1 {
		pages = append(pages, PageData{Page: 1, IsCurrent: false, IsEllipsis: false})

		if start > 2 {
			pages = append(pages, PageData{Page: 0, IsCurrent: false, IsEllipsis: true})
		}
	}

	// Show pages around current
	for i := start; i <= end; i++ {
		pages = append(pages, PageData{Page: i, IsCurrent: i == currentPage, IsEllipsis: false})
	}

	// Always show last page if not in range
	if end < totalPages {
		if end < totalPages-1 {
			pages = append(pages, PageData{Page: 0, IsCurrent: false, IsEllipsis: true})
		}

		pages = append(pages, PageData{Page: totalPages, IsCurrent: false, IsEllipsis: false})
	}

	return pages
}
