// Package orders - template helper functions
package orders

import (
	"strconv"

	"github.com/dukerupert/ironman/dto"
)

// GetRowClass returns CSS class for table rows based on origin
func GetRowClass(origin string) string {
	if origin == "WooCommerce" {
		return "origin-woocommerce"
	}
	return "origin-orderspace"
}

// GetOriginBadgeClass returns CSS classes for origin badges
func GetOriginBadgeClass(origin string) string {
	base := "px-2 py-1 rounded-full text-xs font-medium "
	if origin == "WooCommerce" {
		return base + "bg-blue-100 text-blue-800"
	}
	return base + "bg-green-100 text-green-800"
}

// GetOriginCircleClass returns CSS classes for origin circle indicators
func GetOriginCircleClass(origin string) string {
	if origin == "WooCommerce" {
		return "bg-blue-400"
	}
	return "bg-green-400"
}

// CountByOrigin counts orders by origin
func CountByOrigin(orders []UnifiedOrder, origin string) int {
	count := 0
	for _, order := range orders {
		if order.Origin == origin {
			count++
		}
	}
	return count
}

// CountByOriginString returns count as string
func CountByOriginString(orders []UnifiedOrder, origin string) string {
	return strconv.Itoa(CountByOrigin(orders, origin))
}

// LengthString returns slice length as string
func LengthString(orders []UnifiedOrder) string {
	return strconv.Itoa(len(orders))
}

// ConvertUnifiedOrderToPageOrder converts a single UnifiedOrder to PageOrder
func ConvertUnifiedOrderToPageOrder(order UnifiedOrder) dto.PageOrder {
	return dto.PageOrder{
		OrderNumber: order.OrderNumber,
		Customer:    order.Customer,
		OrderDate:   order.OrderDate,
		DeliverOn:   order.DeliverOn,
		Total:       order.Total,
		Status:      order.Status,
		Origin:      order.Origin,
		SortDate:    order.SortDate,
	}
}

// ConvertPaginatedOrdersToPagePaginatedOrders converts PaginatedOrders to PagePaginatedOrders
func ConvertPaginatedOrdersToPagePaginatedOrders(paginatedOrders PaginatedOrders) dto.PagePaginatedOrders {
	pageOrders := make([]dto.PageOrder, len(paginatedOrders.Orders))
	
	for i, order := range paginatedOrders.Orders {
		pageOrders[i] = ConvertUnifiedOrderToPageOrder(order)
	}
	
	return dto.PagePaginatedOrders{
		Orders:      pageOrders,
		CurrentPage: paginatedOrders.CurrentPage,
		TotalPages:  paginatedOrders.TotalPages,
		TotalOrders: paginatedOrders.TotalOrders,
		PerPage:     paginatedOrders.PerPage,
		HasPrev:     paginatedOrders.HasPrev,
		HasNext:     paginatedOrders.HasNext,
	}
}