// Data transfer objects (dto)
package dto

import (
	"fmt"
	"strings"
)

// UnifiedOrder represents an order from either system for display
type UnifiedOrder struct {
	OrderNumber string
	Customer    string
	OrderDate   string
	DeliverOn   string
	Total       string
	Status      string
	Origin      string
}

// FormatCurrency formats the currency amount based on the currency
func FormatCurrency(amount float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "GBP":
		return fmt.Sprintf("£%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}