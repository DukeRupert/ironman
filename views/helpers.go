package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dukerupert/ironman/dto"
)

// Helper functions for templates
func getRowClass(origin string) string {
	if origin == "WooCommerce" {
		return "origin-woocommerce"
	}
	return "origin-orderspace"
}

func getOriginBadgeClass(origin string) string {
	base := "px-2 py-1 rounded-full text-xs font-medium "
	if origin == "WooCommerce" {
		return base + "bg-blue-100 text-blue-800"
	}
	return base + "bg-green-100 text-green-800"
}

func getOriginCircleClass(origin string) string {
	if origin == "WooCommerce" {
		return "bg-blue-400"
	}
	return "bg-green-400"
}

func countByOrigin(orders []dto.UnifiedOrder, origin string) int {
	count := 0
	for _, order := range orders {
		if order.Origin == origin {
			count++
		}
	}
	return count
}

func countByOriginString(orders []dto.UnifiedOrder, origin string) string {
	return strconv.Itoa(countByOrigin(orders, origin))
}

func lengthString(orders []dto.UnifiedOrder) string {
	return strconv.Itoa(len(orders))
}

// renderPageNumbers generates the page number links for pagination
func renderPageNumbers(pagination dto.PaginatedOrders) string {
	currentPage := pagination.CurrentPage
	totalPages := pagination.TotalPages
	
	var html strings.Builder
	
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
		html.WriteString(`<a href="?page=1" class="inline-flex items-center border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">1</a>`)
		
		if start > 2 {
			html.WriteString(`<span class="inline-flex items-center border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-500">...</span>`)
		}
	}
	
	// Show pages around current
	for i := start; i <= end; i++ {
		if i == currentPage {
			html.WriteString(fmt.Sprintf(`<a href="?page=%d" aria-current="page" class="inline-flex items-center border-t-2 border-indigo-500 px-4 pt-4 text-sm font-medium text-indigo-600">%d</a>`, i, i))
		} else {
			html.WriteString(fmt.Sprintf(`<a href="?page=%d" class="inline-flex items-center border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">%d</a>`, i, i))
		}
	}
	
	// Always show last page if not in range
	if end < totalPages {
		if end < totalPages-1 {
			html.WriteString(`<span class="inline-flex items-center border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-500">...</span>`)
		}
		
		html.WriteString(fmt.Sprintf(`<a href="?page=%d" class="inline-flex items-center border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">%d</a>`, totalPages, totalPages))
	}
	
	return html.String()
}

// getPageNumbersData generates page data for template iteration
func getPageNumbersData(currentPage, totalPages int) []dto.PageData {
	var pages []dto.PageData
	
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
		pages = append(pages, dto.PageData{Page: 1, IsCurrent: false, IsEllipsis: false})
		
		if start > 2 {
			pages = append(pages,dto.PageData{Page: 0, IsCurrent: false, IsEllipsis: true})
		}
	}
	
	// Show pages around current
	for i := start; i <= end; i++ {
		pages = append(pages, dto.PageData{Page: i, IsCurrent: i == currentPage, IsEllipsis: false})
	}
	
	// Always show last page if not in range
	if end < totalPages {
		if end < totalPages-1 {
			pages = append(pages, dto.PageData{Page: 0, IsCurrent: false, IsEllipsis: true})
		}
		
		pages = append(pages, dto.PageData{Page: totalPages, IsCurrent: false, IsEllipsis: false})
	}
	
	return pages
}