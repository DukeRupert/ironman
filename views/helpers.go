package views

import (
	"strconv"
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