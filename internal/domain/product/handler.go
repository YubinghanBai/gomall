package product

import (
	"github.com/gin-gonic/gin"
	"gomall/utils/response"
	"net/http"
	"strconv"
)

// Handler handles product-related HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all product routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		// Public endpoints (no auth required)
		products.GET("", h.ListProducts)            // GET /products
		products.GET("/:id", h.GetProduct)           // GET /products/:id
		products.GET("/search", h.SearchProducts)    // GET /products/search
		products.GET("/featured", h.GetFeatured)     // GET /products/featured
		products.GET("/category/:category_id", h.GetByCategory) // GET /products/category/:category_id
		products.GET("/price-range", h.GetByPriceRange)         // GET /products/price-range

		// Protected endpoints (require auth)
		products.POST("", h.CreateProduct)                    // POST /products
		products.PUT("/:id", h.UpdateProduct)                 // PUT /products/:id
		products.DELETE("/:id", h.DeleteProduct)              // DELETE /products/:id
		products.PUT("/:id/stock", h.UpdateStock)             // PUT /products/:id/stock
		products.GET("/low-stock", h.GetLowStock)             // GET /products/low-stock
		products.POST("/:id/images", h.AddImages)             // POST /products/:id/images
		products.PUT("/:id/images/:image_id/main", h.SetMainImage) // PUT /products/:id/images/:image_id/main
		products.DELETE("/images/:image_id", h.DeleteImage)   // DELETE /products/images/:image_id
	}
}

// CreateProduct godoc
// @Summary      Create Product
// @Description  Create a new product with images
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        request  body      CreateProductRequest  true  "Product information"
// @Success      201      {object}  response.Response{data=ProductDetailResponse}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    product,
	})
}

// GetProduct godoc
// @Summary      Get Product
// @Description  Get product details by ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.Response{data=ProductDetailResponse}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, product)
}

// UpdateProduct godoc
// @Summary      Update Product
// @Description  Update product information
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Product ID"
// @Param        request  body      UpdateProductRequest  true  "Product information"
// @Success      200      {object}  response.Response{data=ProductResponse}
// @Failure      400      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, req)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, product)
}

// DeleteProduct godoc
// @Summary      Delete Product
// @Description  Delete a product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	err = h.service.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListProducts godoc
// @Summary      List Products
// @Description  List products with filtering and pagination
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        category_id  query     int     false  "Category ID"
// @Param        status       query     string  false  "Status (draft/published/off_shelf)"
// @Param        page         query     int     false  "Page number (default: 1)"
// @Param        page_size    query     int     false  "Page size (default: 20)"
// @Success      200          {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      400          {object}  response.Response
// @Failure      500          {object}  response.Response
// @Router       /products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	var req ListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	products, err := h.service.ListProducts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// SearchProducts godoc
// @Summary      Search Products
// @Description  Search products by keyword
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        keyword   query     string  true   "Search keyword"
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        page_size query     int     false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /products/search [get]
func (h *Handler) SearchProducts(c *gin.Context) {
	var req SearchProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	products, err := h.service.SearchProducts(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// GetFeatured godoc
// @Summary      Get Featured Products
// @Description  Get featured/hot products
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      500       {object}  response.Response
// @Router       /products/featured [get]
func (h *Handler) GetFeatured(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	products, err := h.service.GetFeaturedProducts(c.Request.Context(), int32(page), int32(pageSize))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// GetByCategory godoc
// @Summary      Get Products by Category
// @Description  Get products filtered by category
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        category_id path      int  true   "Category ID"
// @Param        page        query     int  false  "Page number (default: 1)"
// @Param        page_size   query     int  false  "Page size (default: 20)"
// @Success      200         {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      400         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /products/category/{category_id} [get]
func (h *Handler) GetByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseInt(c.Param("category_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid category id")
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	products, err := h.service.GetProductsByCategory(c.Request.Context(), categoryID, int32(page), int32(pageSize))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// GetByPriceRange godoc
// @Summary      Get Products by Price Range
// @Description  Get products filtered by price range
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        min_price query     int  true   "Minimum price"
// @Param        max_price query     int  true   "Maximum price"
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /products/price-range [get]
func (h *Handler) GetByPriceRange(c *gin.Context) {
	var req PriceRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	products, err := h.service.GetProductsByPriceRange(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// UpdateStock godoc
// @Summary      Update Product Stock
// @Description  Update product stock (increment or decrement)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id       path      int                  true  "Product ID"
// @Param        request  body      UpdateStockRequest  true  "Stock delta"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /products/{id}/stock [put]
func (h *Handler) UpdateStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err = h.service.UpdateStock(c.Request.Context(), id, req.Delta)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetLowStock godoc
// @Summary      Get Low Stock Products
// @Description  Get products with stock below threshold
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        page_size query     int  false  "Page size (default: 20)"
// @Success      200       {object}  response.Response{data=PaginatedProductsResponse}
// @Failure      500       {object}  response.Response
// @Router       /products/low-stock [get]
func (h *Handler) GetLowStock(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	products, err := h.service.GetLowStockProducts(c.Request.Context(), int32(page), int32(pageSize))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, products)
}

// AddImages godoc
// @Summary      Add Product Images
// @Description  Add images to a product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Product ID"
// @Param        request  body      []ImageRequest  true  "Images"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /products/{id}/images [post]
func (h *Handler) AddImages(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var images []ImageRequest
	if err := c.ShouldBindJSON(&images); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err = h.service.AddProductImages(c.Request.Context(), id, images)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// SetMainImage godoc
// @Summary      Set Main Product Image
// @Description  Set the main image for a product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id        path      int  true  "Product ID"
// @Param        image_id  path      int  true  "Image ID"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /products/{id}/images/{image_id}/main [put]
func (h *Handler) SetMainImage(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product id")
		return
	}

	imageID, err := strconv.ParseInt(c.Param("image_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid image id")
		return
	}

	err = h.service.SetMainImage(c.Request.Context(), productID, imageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteImage godoc
// @Summary      Delete Product Image
// @Description  Delete a product image
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        image_id path      int  true  "Image ID"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /products/images/{image_id} [delete]
func (h *Handler) DeleteImage(c *gin.Context) {
	imageID, err := strconv.ParseInt(c.Param("image_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid image id")
		return
	}

	err = h.service.DeleteProductImage(c.Request.Context(), imageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}
