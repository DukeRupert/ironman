package orders

import (
	"log/slog"
	"net/http"
	"strconv"

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
