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
// In your orders package
func (s *Service) HandleOrders(c echo.Context) error {
	// Get page parameter
	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Get orders (your existing logic)
	paginatedOrders, err := s.GetUnifiedOrdersPaginated(page, 20)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch orders",
		})
	}

	// Convert for template
	pagePaginatedOrders := ConvertPaginatedOrdersToPagePaginatedOrders(*paginatedOrders)

	// Check if this is an HTMX request (for pagination)
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return just the table component for HTMX swapping
		return c.Render(http.StatusOK, "orders_table_with_pagination", pagePaginatedOrders)
	}

	// Prepare template data
	data := views.TemplateData{
		Pagination: pagePaginatedOrders,
	}

	// Return full page for regular requests
	return c.Render(http.StatusOK, "orders_page", data)
}

// HandleOrder for the /order/:id route returns detail page
func (s *Service) HandleOrder(c echo.Context) error {
	slog.Info("Order details page requested", "path", c.Request().URL.Path)

	// Get pagination parameters from query string
	id := c.Param("id")
	if id == "" {
		slog.Error("Missing order id parameter")
		return c.NoContent(http.StatusBadRequest)
	}

	orderType := "woocommerce"
	// Check if id begins with "or_" which indicates an OrderSpace order
	if strings.Contains(id, "or_") {
		orderType = "orderspace"
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

	return c.String(http.StatusOK, fmt.Sprintf("Order id: %s", id))
}
