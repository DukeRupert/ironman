package views

import (
	"html/template"
	"strings"

	"github.com/dukerupert/ironman/dto"
)

// PageData represents data for pagination page numbers
type PageData struct {
	Page       int
	IsCurrent  bool
	IsEllipsis bool
}

// TemplateFuncs returns a map of template functions for html/template
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },

		// Status badge class function
		"statusClass": func(status string) string {
			return strings.ToLower(strings.ReplaceAll(status, " ", "-"))
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
			Page:      i,
			IsCurrent: i == currentPage,
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

// TemplateData represents the data structure passed to templates
type TemplateData struct {
	Pagination dto.PagePaginatedOrders
	// Add other fields as needed for different pages
}