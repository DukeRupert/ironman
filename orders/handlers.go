package orders

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/dukerupert/ironman/views"
	"github.com/labstack/echo"
)

// Handler for the orders page
func (s *Service) HandleOrders(c echo.Context) error {
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

	// Convert for template
	pagePaginatedOrders := ConvertPaginatedOrdersToPagePaginatedOrders(*paginatedOrders)

	// Check if this is an HTMX request (for pagination)
	if c.Request().Header.Get("HX-Request") == "true" {
		slog.Info("HTMX request detected, returning table component only")
		// Return just the table component for HTMX swapping
		ordersTable := views.OrdersTableWithPagination(pagePaginatedOrders)
		return ordersTable.Render(c.Request().Context(), c.Response())
	}

	// Return full page for regular requests
	slog.Info("Regular request, returning full page")
	ordersPage := views.OrdersPage(pagePaginatedOrders)
	return ordersPage.Render(c.Request().Context(), c.Response())
}

// Handler for the order detail page
func (s *Service) HandleOrder(c echo.Context) error {
	slog.Info("Order details page requested", "path", c.Request().URL.Path)

	// Get pagination parameters from query string
	id := c.Param("id")
	if id == "" {
		slog.Error("Missing order id parameter")
		return c.NoContent(http.StatusBadRequest)
	}

	split := strings.Split(id, "_")
	orderType := split[0]
	orderID := split[1]
	if orderType == "" || orderID == "" {
		slog.Error("Malformed order id parameter")
		return c.NoContent(http.StatusBadRequest)
	}

	if orderType == "woo" {
		orderType = "woocommerce"
	}

	switch orderType {
	case "woo":
		orderType = "woocommerce"
	case "ord":
		orderType = "orderspace"
	}

	slog.Info("Fetching order details", "order_type", orderType, "order_id", orderID)

	
	// // Convert for template
	// pagePaginatedOrders := ConvertPaginatedOrdersToPagePaginatedOrders(*paginatedOrders)

	// // Check if this is an HTMX request (for pagination)
	// if c.Request().Header.Get("HX-Request") == "true" {
	// 	slog.Info("HTMX request detected, returning table component only")
	// 	// Return just the table component for HTMX swapping
	// 	ordersTable := views.OrdersTableWithPagination(pagePaginatedOrders)
	// 	return ordersTable.Render(c.Request().Context(), c.Response())
	// }

	// // Return full page for regular requests
	// slog.Info("Regular request, returning full page")
	// ordersPage := views.OrdersPage(pagePaginatedOrders)
	// return ordersPage.Render(c.Request().Context(), c.Response())

	return c.String(http.StatusOK, fmt.Sprintf("Order from: %s, #: %s", orderType, orderID))
}
