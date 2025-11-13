package category

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gomall/utils/response"
)

type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all category routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	categories := router.Group("/categories")
	{
		// Public endpoints (no auth required)
		categories.GET("", h.ListCategories)
		categories.GET("/tree", h.GetCategoryTree)
		categories.GET("/roots", h.GetRoots)
		categories.GET("/:id", h.GetCategory)
		categories.GET("/:id/children", h.GetChildren)

		// Protected endpoints (require auth) - uncomment when auth is needed
		// categories.POST("", middleware.AuthMiddleware(), h.CreateCategory)
		// categories.PUT("/:id", middleware.AuthMiddleware(), h.UpdateCategory)
		// categories.DELETE("/:id", middleware.AuthMiddleware(), h.DeleteCategory)

		// For now, allow without auth for testing
		categories.POST("", h.CreateCategory)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
	}
}

// CreateCategory godoc
// @Summary      Create Category
// @Description  Create a new category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        request  body      CreateCategoryRequest  true  "Category information"
// @Success      201      {object}  response.Response{data=CategoryResponse}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    category,
	})
}

// GetCategory godoc
// @Summary      Get Category
// @Description  Get category by ID
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  response.Response{data=CategoryResponse}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /categories/{id} [get]
func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid category id")
		return
	}

	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, category)
}

// UpdateCategory godoc
// @Summary      Update Category
// @Description  Update category information
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Category ID"
// @Param        request  body      UpdateCategoryRequest  true  "Category information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid category id")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	err = h.service.UpdateCategory(c.Request.Context(), id, req)
	if err != nil {
		if err.Error() == "category not found" || err.Error() == "parent category not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "category cannot be its own parent" {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "category updated successfully"})
}

// DeleteCategory godoc
// @Summary      Delete Category
// @Description  Soft delete a category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid category id")
		return
	}

	err = h.service.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "cannot delete category with children" || err.Error() == "cannot delete category with products" {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "category deleted successfully"})
}

// ListCategories godoc
// @Summary      List Categories
// @Description  Get all categories with optional filter
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        is_active  query     bool  false  "Filter by active status"
// @Success      200        {object}  response.Response{data=[]CategoryResponse}
// @Failure      500        {object}  response.Response
// @Router       /categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	var isActive *bool
	if c.Query("is_active") != "" {
		val := c.Query("is_active") == "true"
		isActive = &val
	}

	categories, err := h.service.ListCategories(c.Request.Context(), isActive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, categories)
}

// GetCategoryTree godoc
// @Summary      Get Category Tree
// @Description  Get hierarchical tree of all active categories
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=[]CategoryTreeNode}
// @Failure      500  {object}  response.Response
// @Router       /categories/tree [get]
func (h *Handler) GetCategoryTree(c *gin.Context) {
	tree, err := h.service.GetCategoryTree(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, tree)
}

// GetChildren godoc
// @Summary      Get Category Children
// @Description  Get all children of a category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  response.Response{data=[]CategoryResponse}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /categories/{id}/children [get]
func (h *Handler) GetChildren(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid category id")
		return
	}

	children, err := h.service.GetChildren(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, children)
}

// GetRoots godoc
// @Summary      Get Root Categories
// @Description  Get all root level categories
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=[]CategoryResponse}
// @Failure      500  {object}  response.Response
// @Router       /categories/roots [get]
func (h *Handler) GetRoots(c *gin.Context) {
	roots, err := h.service.GetRoots(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, roots)
}
