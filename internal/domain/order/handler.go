package order

import (
	"github.com/gin-gonic/gin"
	"gomall/internal/common/middleware"
	"gomall/utils/response"
	"net/http"
	"strconv"
)

// Handler handles order-related HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all order routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	orders.Use(middleware.AuthMiddleware(nil)) // Apply auth middleware to all order routes
	{
		orders.POST("", h.CreateOrder)                // POST /orders
		orders.GET("", h.ListOrders)                  // GET /orders
		orders.GET("/:id", h.GetOrder)                // GET /orders/:id
		orders.GET("/order-no/:order_no", h.GetOrderByOrderNo) // GET /orders/order-no/:order_no
		orders.PUT("/:id/status", h.UpdateOrderStatus)         // PUT /orders/:id/status
		orders.POST("/:id/cancel", h.CancelOrder)              // POST /orders/:id/cancel
		orders.POST("/:id/pay", h.PayOrder)                    // POST /orders/:id/pay
		orders.POST("/:id/ship", h.ShipOrder)                  // POST /orders/:id/ship
		orders.POST("/:id/complete", h.CompleteOrder)          // POST /orders/:id/complete
	}
}

// CreateOrder godoc
// @Summary      Create Order
// @Description  Create a new order with items
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      CreateOrderRequest  true  "Order information"
// @Success      201      {object}  response.Response{data=OrderResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), payload.UserID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
}

// GetOrder godoc
// @Summary      Get Order
// @Description  Get order details by ID
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  response.Response{data=OrderResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), payload.UserID, id)
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "unauthorized access to order" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, order)
}

// GetOrderByOrderNo godoc
// @Summary      Get Order by Order Number
// @Description  Get order details by order number
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        order_no path      string  true  "Order Number"
// @Success      200      {object}  response.Response{data=OrderResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /orders/order-no/{order_no} [get]
func (h *Handler) GetOrderByOrderNo(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, "order number is required")
		return
	}

	order, err := h.service.GetOrderByOrderNo(c.Request.Context(), payload.UserID, orderNo)
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "unauthorized access to order" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, order)
}

// ListOrders godoc
// @Summary      List Orders
// @Description  List orders for the current user with pagination
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedOrdersResponse}
// @Failure      400       {object}  response.Response
// @Failure      401       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	orders, err := h.service.ListUserOrders(c.Request.Context(), payload.UserID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, orders)
}

// UpdateOrderStatus godoc
// @Summary      Update Order Status
// @Description  Update the status of an order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                       true  "Order ID"
// @Param        request  body      UpdateOrderStatusRequest  true  "Status information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /orders/{id}/status [put]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err = h.service.UpdateOrderStatus(c.Request.Context(), payload.UserID, id, req)
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "unauthorized access to order" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// CancelOrder godoc
// @Summary      Cancel Order
// @Description  Cancel an order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	err = h.service.CancelOrder(c.Request.Context(), payload.UserID, id)
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "unauthorized access to order" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "order cannot be cancelled" {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "order cancelled successfully"})
}


// PayOrder godoc
// @Summary      Pay Order
// @Description  Mark order as paid and deduct stock
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /orders/{id}/pay [post]
func (h *Handler) PayOrder(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	err = h.service.PayOrder(c.Request.Context(), payload.UserID, id)
	if err != nil {
		if err.Error() == "order not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
				if err.Error() == "unauthorized access to order" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "order paid successfully"})
}

// ShipOrder godoc
// @Summary      Ship Order
// @Description  Mark order as shipped (admin only)
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /orders/{id}/ship [post]
func (h *Handler) ShipOrder(c *gin.Context) {
	// TODO: Add admin permission check
	
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}
		err = h.service.ShipOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "order shipped successfully"})
}

// CompleteOrder godoc
// @Summary      Complete Order
// @Description  Mark order as completed
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /orders/{id}/complete [post]
func (h *Handler) CompleteOrder(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
				response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	err = h.service.CompleteOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "order completed successfully"})
}