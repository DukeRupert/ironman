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

// HandleOrders for the /orders route
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
	case "orderspace":
		slog.Info("Fetching order details", "order_type", orderType, "order_id", id)
		order, err := s.orderspaceClient.GetOrder(id)
		if err != nil {
			slog.Error("Failed to fetch orderspace order details: %s", err)
			return c.String(http.StatusServiceUnavailable, "Failed to retrieve order details from OrderSpace")
		}
		// Comprehensive debug logging of the order response
		slog.Debug("Order response received",
			// Basic order info
			"order_id", order.ID,
			"order_number", order.Number,
			"status", order.Status,
			"created", order.Created,
			"currency", order.Currency,

			// Customer info
			"customer_id", order.CustomerID,
			"company_name", order.CompanyName,
			"phone", order.Phone,
			"email_orders", order.EmailAddresses.Orders,
			"email_dispatches", order.EmailAddresses.Dispatches,
			"email_invoices", order.EmailAddresses.Invoices,

			// Order details
			"delivery_date", order.DeliveryDate,
			"reference", order.Reference,
			"customer_po_number", order.CustomerPONumber,
			"shipping_type", order.ShippingType,
			"created_by", order.CreatedBy,

			// Financial info
			"net_total", order.NetTotal,
			"gross_total", order.GrossTotal,

			// Addresses
			"shipping_company", order.ShippingAddress.CompanyName,
			"shipping_contact", order.ShippingAddress.ContactName,
			"shipping_line1", order.ShippingAddress.Line1,
			"shipping_city", order.ShippingAddress.City,
			"shipping_country", order.ShippingAddress.Country,
			"billing_company", order.BillingAddress.CompanyName,
			"billing_contact", order.BillingAddress.ContactName,
			"billing_line1", order.BillingAddress.Line1,
			"billing_city", order.BillingAddress.City,
			"billing_country", order.BillingAddress.Country,

			// Order lines summary
			"order_lines_count", len(order.OrderLines),

			// Notes
			"customer_note", order.CustomerNote,
			"internal_note", order.InternalNote,
			"standing_order_id", order.StandingOrderID,
		)
		page := views.OrderDetailPage(*order)
		return page.Render(c.Request().Context(), c.Response())
	case "woocommerce":
		slog.Info("Fetching order details", "order_type", orderType, "order_id", id)
		orderID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Malformed order id cannot be converted to an integer: %s", err)
			return c.NoContent(http.StatusBadRequest)
		}
		order, err := s.wooClient.GetOrder(orderID)
		if err != nil {
			slog.Error("Failed to fetch woocommerce order details: %s", err)
			return c.NoContent(http.StatusBadRequest)
		}
		page := views.WooOrderDetailPage(*order)
		return page.Render(c.Request().Context(), c.Response())
	}

	// Convert for template
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
