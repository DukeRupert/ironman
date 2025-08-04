// Data dto contains data transfer objects
package dto

import (
	"fmt"
	"strings"
	"time"
)

type PageOrder struct {
	ID          string
	OrderNumber string
	Customer    string
	OrderDate   string
	DeliverOn   string
	Total       string
	Status      string
	Origin      string
	SortDate    time.Time // Added for sorting purposes
}

// PaginatedOrders represents paginated order results
type PagePaginatedOrders struct {
	Orders      []PageOrder
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

