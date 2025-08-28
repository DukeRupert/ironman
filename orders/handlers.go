package orders

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dukerupert/ironman/views"
	"github.com/labstack/echo"
)

// Handler for the orders page
func (s *Service) HandleOrders(c echo.Context) error {
	// Get page parameter
	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Get orders (this is your existing logic)
	paginatedOrders, err := s.GetUnifiedOrdersPaginated(page, 20) // 20 orders per page
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

// HandleOrderDetailWithHTMLTemplate handles individual order detail pages
func (s *Service) HandleOrderDetailWithHTMLTemplate(c echo.Context) error {
	// Get order ID from URL parameter
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing order ID",
		})
	}

	// Determine order type based on ID format
	orderType := "woocommerce"
	if strings.Contains(id, "or_") {
		orderType = "orderspace"
	}

	var data views.TemplateData
	var templateName string

	switch orderType {
	case "orderspace":
		// Fetch Orderspace order
		order, err := s.orderspaceClient.GetOrder(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch order details",
			})
		}

		data = views.TemplateData{
			Order:     *order,
			OrderType: "orderspace",
		}
		templateName = "orderspace_order_detail_page"

	case "woocommerce":
		// Convert string ID to int for WooCommerce
		orderID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid order ID format",
			})
		}

		// Fetch WooCommerce order
		order, err := s.wooClient.GetOrder(orderID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch order details",
			})
		}

		data = views.TemplateData{
			Order:     *order,
			OrderType: "woocommerce",
		}
		templateName = "woo_order_detail_page"

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Unknown order type",
		})
	}

	// Check if this is an HTMX request (for partial updates)
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return just the detail view component for HTMX swapping
		if orderType == "orderspace" {
			return c.Render(http.StatusOK, "orderspace_order_detail_view", data.Order)
		} else {
			return c.Render(http.StatusOK, "woo_order_detail_view", data.Order)
		}
	}

	// Return full page for regular requests
	return c.Render(http.StatusOK, templateName, data)
}
