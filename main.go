package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/dukerupert/ironman/dto"
	"github.com/dukerupert/ironman/orderspace"
	"github.com/dukerupert/ironman/woo"

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

// Example usage function
func main() {

	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize both clients
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

	// Fetch orders from both systems
	fmt.Println("Fetching orders from both systems...")
	fmt.Println(strings.Repeat("=", 60))

	// Get WooCommerce orders (last 10, sorted by date desc)
	wooOrders, err := wooClient.ListOrders(&woo.OrderListOptions{
		Page:    1,
		PerPage: 10,
		OrderBy: "date",
		Order:   "desc",
	})
	if err != nil {
		log.Printf("Error fetching WooCommerce orders: %v", err)
		wooOrders = &woo.OrdersResponse{Orders: []woo.Order{}}
	}

	// Get Orderspace orders (last 10)
	orderspaceOrders, err := orderspaceClient.GetLast10Orders()
	if err != nil {
		log.Printf("Error fetching Orderspace orders: %v", err)
		orderspaceOrders = &orderspace.OrdersResponse{Orders: []orderspace.Order{}}
	}

	// Convert all orders to unified format
	var unifiedOrders []dto.UnifiedOrder

	// Process WooCommerce orders
	for _, order := range wooOrders.Orders {
		unified, err := woo.ParseWooOrder(order)
		if err != nil {
			log.Printf("Error parsing WooCommerce order %d: %v", order.ID, err)
			continue
		}
		unifiedOrders = append(unifiedOrders, unified)
	}

	// Process Orderspace orders
	for _, order := range orderspaceOrders.Orders {
		unified, err := orderspace.ParseOrderspaceOrder(order)
		if err != nil {
			log.Printf("Error parsing Orderspace order %s: %v", order.ID, err)
			continue
		}
		unifiedOrders = append(unifiedOrders, unified)
	}

	// Sort all orders by date (most recent first)
	sort.Slice(unifiedOrders, func(i, j int) bool {
		return unifiedOrders[i].Date.After(unifiedOrders[j].Date)
	})

	// Print unified results
	fmt.Println("\nUnified Orders (sorted by date):")
	fmt.Println(strings.Repeat("=", 60))

	if len(unifiedOrders) == 0 {
		fmt.Println("No orders found from either system.")
		return
	}

	for _, order := range unifiedOrders {
		fmt.Printf("Order ID: %s, Date: %s, Total: %s, Origin: %s\n",
			order.ID,
			order.Date.Format("2006-01-02T15:04:05"),
			dto.FormatCurrency(order.Total, order.Currency),
			order.Origin)
	}

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	wooCount := len(wooOrders.Orders)
	orderspaceCount := len(orderspaceOrders.Orders)
	totalCount := len(unifiedOrders)

	fmt.Printf("Summary:\n")
	fmt.Printf("  WooCommerce orders: %d\n", wooCount)
	fmt.Printf("  Orderspace orders: %d\n", orderspaceCount)
	fmt.Printf("  Total unified orders: %d\n", totalCount)

	if totalCount > 0 {
		fmt.Printf("  Date range: %s to %s\n",
			unifiedOrders[len(unifiedOrders)-1].Date.Format("2006-01-02"),
			unifiedOrders[0].Date.Format("2006-01-02"))
	}

}
