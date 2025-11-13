package inventory

import (
	"github.com/gin-gonic/gin"
	"gomall/internal/common/middleware"
	"gomall/utils/response"
	"net/http"
	"strconv"
)

// Handler handles inventory-related HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all inventory routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	inventory := router.Group("/inventory")
	{
		// Public endpoints (check stock)
		inventory.GET("/check/:product_id", h.CheckStock) // GET /inventory/check/:product_id
		inventory.POST("/check/batch", h.BatchCheckStock) // POST /inventory/check/batch

		// Protected endpoints (require auth - admin only in production)
		inventory.POST("", h.CreateInventory)                           // POST /inventory
		inventory.GET("", h.ListInventories)                            // GET /inventory
		inventory.GET("/product/:product_id", h.GetInventoryByProduct)  // GET /inventory/product/:product_id
		inventory.GET("/low-stock", h.ListLowStock)                     // GET /inventory/low-stock
		inventory.POST("/restock", h.Restock)                           // POST /inventory/restock
		inventory.POST("/adjust", h.AdjustStock)                        // POST /inventory/adjust
		inventory.PUT("/:product_id/threshold", h.UpdateThreshold)      // PUT /inventory/:product_id/threshold
		inventory.GET("/logs/:product_id", h.GetInventoryLogs)          // GET /inventory/logs/:product_id

		// Reservation management (internal use)
		inventory.POST("/reserve", h.ReserveStock)                      // POST /inventory/reserve
		inventory.POST("/release", h.ReleaseStock)                      // POST /inventory/release
		inventory.POST("/deduct", h.DeductStock)                        // POST /inventory/deduct
		inventory.POST("/cleanup-expired", h.CleanupExpiredReservations) // POST /inventory/cleanup-expired
	}
}

// CreateInventory godoc
// @Summary      Create Inventory
// @Description  Create inventory record for a product
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      CreateInventoryRequest  true  "Inventory information"
// @Success      201      {object}  response.Response{data=InventoryResponse}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory [post]
func (h *Handler) CreateInventory(c *gin.Context) {
	var req CreateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	inventory, err := h.service.CreateInventory(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    inventory,
	})
}

// GetInventoryByProduct godoc
// @Summary      Get Inventory by Product
// @Description  Get inventory details for a specific product
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Param        product_id path      int  true  "Product ID"
// @Success      200        {object}  response.Response{data=InventoryResponse}
// @Failure      404        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /inventory/product/{product_id} [get]
func (h *Handler) GetInventoryByProduct(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	inventory, err := h.service.GetInventoryByProductID(c.Request.Context(), productID)
	if err != nil {
		if err.Error() == "inventory not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, inventory)
}

// ListInventories godoc
// @Summary      List Inventories
// @Description  List all inventories with pagination
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedInventoriesResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /inventory [get]
func (h *Handler) ListInventories(c *gin.Context) {
	var req ListInventoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	inventories, err := h.service.ListInventories(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, inventories)
}

// ListLowStock godoc
// @Summary      List Low Stock Items
// @Description  List inventories with low stock levels
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedInventoriesResponse}
// @Failure      500       {object}  response.Response
// @Router       /inventory/low-stock [get]
func (h *Handler) ListLowStock(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	inventories, err := h.service.ListLowStockInventories(c.Request.Context(), int32(page), int32(pageSize))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, inventories)
}

// CheckStock godoc
// @Summary      Check Stock Availability
// @Description  Check if stock is available for a product
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Param        product_id path      int  true  "Product ID"
// @Param        quantity   query     int  true  "Requested quantity"
// @Success      200        {object}  response.Response{data=StockCheckResponse}
// @Failure      400        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /inventory/check/{product_id} [get]
func (h *Handler) CheckStock(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	quantity, err := strconv.ParseInt(c.Query("quantity"), 10, 32)
	if err != nil || quantity <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid quantity")
		return
	}

	check, err := h.service.CheckStockAvailability(c.Request.Context(), productID, int32(quantity))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, check)
}

// BatchCheckStock godoc
// @Summary      Batch Check Stock Availability
// @Description  Check stock availability for multiple products
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Param        request  body      []StockCheckItem  true  "Products to check"
// @Success      200      {object}  response.Response{data=map[int64]StockCheckResponse}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/check/batch [post]
func (h *Handler) BatchCheckStock(c *gin.Context) {
	var items []StockCheckItem
	if err := c.ShouldBindJSON(&items); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := h.service.BatchCheckStockAvailability(c.Request.Context(), items)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// ReserveStock godoc
// @Summary      Reserve Stock
// @Description  Reserve stock for an order (prevents overselling)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      ReserveStockRequest  true  "Reserve information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/reserve [post]
func (h *Handler) ReserveStock(c *gin.Context) {
	var req ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Default reservation expires in 30 minutes
	err := h.service.ReserveStock(c.Request.Context(), req, 30)
	if err != nil {
		if err.Error() == "insufficient stock" {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock reserved successfully"})
}

// ReleaseStock godoc
// @Summary      Release Reserved Stock
// @Description  Release reserved stock (e.g., when order is cancelled)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      ReleaseStockRequest  true  "Release information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/release [post]
func (h *Handler) ReleaseStock(c *gin.Context) {
	var req ReleaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err := h.service.ReleaseStock(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock released successfully"})
}

// DeductStock godoc
// @Summary      Deduct Reserved Stock
// @Description  Deduct reserved stock (e.g., when order is paid/confirmed)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      DeductStockRequest  true  "Deduct information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/deduct [post]
func (h *Handler) DeductStock(c *gin.Context) {
	var req DeductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err := h.service.DeductStock(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock deducted successfully"})
}

// Restock godoc
// @Summary      Restock Inventory
// @Description  Add stock to inventory
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      RestockRequest  true  "Restock information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/restock [post]
func (h *Handler) Restock(c *gin.Context) {
	var req RestockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get operator ID from middleware if available
	payload := middleware.GetPayload(c)
	var operatorID *int64
	if payload != nil {
		operatorID = &payload.UserID
	}

	err := h.service.RestockInventory(c.Request.Context(), req, operatorID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "inventory restocked successfully"})
}

// AdjustStock godoc
// @Summary      Adjust Stock
// @Description  Adjust inventory stock (can be positive or negative)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      AdjustStockRequest  true  "Adjustment information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /inventory/adjust [post]
func (h *Handler) AdjustStock(c *gin.Context) {
	var req AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get operator ID from middleware if available
	payload := middleware.GetPayload(c)
	var operatorID *int64
	if payload != nil {
		operatorID = &payload.UserID
	}

	err := h.service.AdjustStock(c.Request.Context(), req, operatorID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock adjusted successfully"})
}

// UpdateThreshold godoc
// @Summary      Update Low Stock Threshold
// @Description  Update the low stock threshold for a product
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        product_id path      int                            true  "Product ID"
// @Param        request    body      UpdateLowStockThresholdRequest true  "Threshold information"
// @Success      200        {object}  response.Response
// @Failure      400        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /inventory/{product_id}/threshold [put]
func (h *Handler) UpdateThreshold(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var req UpdateLowStockThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err = h.service.UpdateLowStockThreshold(c.Request.Context(), productID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "threshold updated successfully"})
}

// GetInventoryLogs godoc
// @Summary      Get Inventory Logs
// @Description  Get inventory change logs for a product
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        product_id path      int  true   "Product ID"
// @Param        page       query     int  false  "Page number (default: 1)"
// @Param        page_size  query     int  false  "Page size (default: 20)"
// @Success      200        {object}  response.Response{data=PaginatedInventoryLogsResponse}
// @Failure      400        {object}  response.Response
// @Failure      500        {object}  response.Response
// @Router       /inventory/logs/{product_id} [get]
func (h *Handler) GetInventoryLogs(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var req ListInventoryLogsRequest
	req.ProductID = productID
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	logs, err := h.service.GetInventoryLogs(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, logs)
}

// CleanupExpiredReservations godoc
// @Summary      Cleanup Expired Reservations
// @Description  Clean up expired stock reservations (admin/cron job)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /inventory/cleanup-expired [post]
func (h *Handler) CleanupExpiredReservations(c *gin.Context) {
	err := h.service.CleanupExpiredReservations(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "expired reservations cleaned up successfully"})
}
