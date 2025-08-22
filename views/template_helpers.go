package views

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dukerupert/ironman/dto"
	"github.com/dukerupert/ironman/orderspace"
	"github.com/dukerupert/ironman/woo"
)

// PageData represents data for pagination page numbers
type PageData struct {
	Page       int
	IsCurrent  bool
	IsEllipsis bool
}

// TemplateData represents the data structure passed to templates
type TemplateData struct {
	Pagination dto.PagePaginatedOrders
	Order      any    // Can be orderspace.Order or woo.Order
	OrderType  string // "orderspace" or "woocommerce"
	// Add other fields as needed for different pages
}

// TemplateFuncs returns a map of template functions for html/template
func TemplateFuncs() template.FuncMap {
	// Create caser instances
	titleCaser := cases.Title(language.AmericanEnglish)
	lowerCaser := cases.Lower(language.AmericanEnglish)
	upperCaser := cases.Upper(language.AmericanEnglish)

	return template.FuncMap{
		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },

		// String functions
		"title": func(s string) string { return titleCaser.String(s) },
		"lower": func(s string) string { return lowerCaser.String(s) },
		"upper": func(s string) string { return upperCaser.String(s) },

		// Status badge class function
		"statusClass": func(status string) string {
			return lowerCaser.String(strings.ReplaceAll(status, " ", "-"))
		},

		// Origin badge functions (from your existing helpers.go)
		"originBadgeClass": func(origin string) string {
			base := "px-2 py-1 rounded-full text-xs font-medium "
			if origin == "WooCommerce" {
				return base + "bg-blue-100 text-blue-800"
			}
			return base + "bg-green-100 text-green-800"
		},

		"originCircleClass": func(origin string) string {
			if origin == "WooCommerce" {
				return "bg-blue-400"
			}
			return "bg-green-400"
		},

		// Count functions
		"countByOrigin": func(orders []dto.PageOrder, origin string) int {
			count := 0
			for _, order := range orders {
				if order.Origin == origin {
					count++
				}
			}
			return count
		},

		// Page numbers function
		"pageNumbers": func(currentPage, totalPages int) []PageData {
			return getPageNumbersData(currentPage, totalPages)
		},

		// Currency formatting
		"formatCurrency": formatCurrency,

		// Date formatting
		"formatOrderDate":    formatOrderDate,
		"formatDeliveryDate": formatDeliveryDate,

		// Status badge classes
		"getStatusBadgeClass":    getStatusBadgeClass,
		"getWooStatusBadgeClass": getWooStatusBadgeClass,

		// Address helpers
		"isAddressEmpty": isAddressEmpty,

		// Customer name helper for WooCommerce
		"getCustomerName": getCustomerName,
	}
}

// getPageNumbersData generates the page number data for pagination
// This replaces the templ version from your original helpers.go
func getPageNumbersData(currentPage, totalPages int) []PageData {
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
			pages = append(pages, PageData{IsEllipsis: true})
		}
	}

	// Show pages around current
	for i := start; i <= end; i++ {
		pages = append(pages, PageData{
			Page:       i,
			IsCurrent:  i == currentPage,
			IsEllipsis: false,
		})
	}

	// Always show last page if not in range
	if end < totalPages {
		if end < totalPages-1 {
			pages = append(pages, PageData{IsEllipsis: true})
		}
		pages = append(pages, PageData{Page: totalPages, IsCurrent: false, IsEllipsis: false})
	}

	return pages
}

// Helper functions for order detail templates

// formatCurrency formats currency amounts
func formatCurrency(amount float64, currency string) string {
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

// formatOrderDate formats ISO 8601 dates for display
func formatOrderDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse the ISO 8601 date
	t, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
	if err != nil {
		// Try alternative format
		t, err = time.Parse("2006-01-02T15:04:05-07:00", dateStr)
		if err != nil {
			return dateStr // Return original if parsing fails
		}
	}

	return t.Format("January 2, 2006")
}

// formatDeliveryDate formats delivery dates for display
func formatDeliveryDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse the date (assuming YYYY-MM-DD format)
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr // Return original if parsing fails
	}

	return t.Format("January 2, 2006")
}

// isAddressEmpty checks if an orderspace address is empty
func isAddressEmpty(address orderspace.OrderAddress) bool {
	return address.CompanyName == "" &&
		address.ContactName == "" &&
		address.Line1 == "" &&
		address.City == ""
}

// getStatusBadgeClass returns CSS classes for orderspace status badges
func getStatusBadgeClass(status string) string {
	lowerCaser := cases.Lower(language.AmericanEnglish)
	switch lowerCaser.String(status) {
	case "new":
		return "bg-purple-50 text-purple-600 ring-purple-600/20"
	case "confirmed":
		return "bg-blue-50 text-blue-600 ring-blue-600/20"
	case "processing":
		return "bg-yellow-50 text-yellow-600 ring-yellow-600/20"
	case "dispatched":
		return "bg-green-50 text-green-600 ring-green-600/20"
	case "delivered":
		return "bg-green-50 text-green-600 ring-green-600/20"
	case "cancelled":
		return "bg-red-50 text-red-600 ring-red-600/20"
	case "on hold":
		return "bg-gray-50 text-gray-600 ring-gray-600/20"
	default:
		return "bg-gray-50 text-gray-600 ring-gray-600/20"
	}
}

// getWooStatusBadgeClass returns CSS classes for WooCommerce status badges
func getWooStatusBadgeClass(status string) string {
	lowerCaser := cases.Lower(language.AmericanEnglish)
	switch lowerCaser.String(status) {
	case "pending":
		return "bg-yellow-50 text-yellow-600 ring-yellow-600/20"
	case "processing":
		return "bg-blue-50 text-blue-600 ring-blue-600/20"
	case "on-hold":
		return "bg-gray-50 text-gray-600 ring-gray-600/20"
	case "completed":
		return "bg-green-50 text-green-600 ring-green-600/20"
	case "cancelled":
		return "bg-red-50 text-red-600 ring-red-600/20"
	case "refunded":
		return "bg-red-50 text-red-600 ring-red-600/20"
	case "failed":
		return "bg-red-50 text-red-600 ring-red-600/20"
	default:
		return "bg-gray-50 text-gray-600 ring-gray-600/20"
	}
}

// getCustomerName extracts customer name from WooCommerce order
func getCustomerName(order woo.Order) string {
	if order.Billing.FirstName != "" || order.Billing.LastName != "" {
		return strings.TrimSpace(order.Billing.FirstName + " " + order.Billing.LastName)
	}
	if order.Billing.Company != "" {
		return order.Billing.Company
	}
	return "Guest Customer"
}
