package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gomall/db/sqlc"
	"gomall/utils"
)

// Service defines the business logic interface for product domain
type Service interface {
	// CreateProduct Product CRUD operations
	CreateProduct(ctx context.Context, req CreateProductRequest) (*ProductDetailResponse, error)
	GetProduct(ctx context.Context, productID int64) (*ProductDetailResponse, error)
	UpdateProduct(ctx context.Context, productID int64, req UpdateProductRequest) (*ProductResponse, error)
	DeleteProduct(ctx context.Context, productID int64) error
	ListProducts(ctx context.Context, req ListProductsRequest) (*PaginatedProductsResponse, error)

	// SearchProducts Search and filtering
	SearchProducts(ctx context.Context, req SearchProductsRequest) (*PaginatedProductsResponse, error)
	GetFeaturedProducts(ctx context.Context, page, pageSize int32) (*PaginatedProductsResponse, error)
	GetProductsByIDs(ctx context.Context,productIDs []int64) (map[int64]*ProductResponse,error)
	GetProductsByCategory(ctx context.Context, categoryID int64, page, pageSize int32) (*PaginatedProductsResponse, error)
	GetProductsByPriceRange(ctx context.Context, req PriceRangeRequest) (*PaginatedProductsResponse, error)

	// UpdateStock Stock management
	UpdateStock(ctx context.Context, productID int64, delta int32) error
	CheckStock(ctx context.Context, productID int64, quantity int32) (bool, error)
	GetLowStockProducts(ctx context.Context, page, pageSize int32) (*PaginatedProductsResponse, error)

	// AddProductImages Image management
	AddProductImages(ctx context.Context, productID int64, images []ImageRequest) error
	SetMainImage(ctx context.Context, productID, imageID int64) error
	DeleteProductImage(ctx context.Context, imageID int64) error

	// IncrementViews Statistics
	IncrementViews(ctx context.Context, productID int64) error
}

type service struct {
	repo Repository
}

// NewService creates a new Service instance
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateProduct creates a new product with images (atomic transaction)
func (s *service) CreateProduct(ctx context.Context, req CreateProductRequest) (*ProductDetailResponse, error) {
	var result ProductDetailResponse

	err := s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Create product
		var specs []byte
		if req.Specifications != "" {
			specs = []byte(req.Specifications)
		}

		product, err := q.CreateProduct(ctx, sqlc.CreateProductParams{
			Name:              req.Name,
			Description:       stringToNullString(req.Description),
			Brand:             stringToNullString(req.Brand),
			Price:             req.Price,
			OriginPrice:       req.OriginPrice,
			CostPrice:         &req.CostPrice,
			Stock:             req.Stock,
			LowStockThreshold: req.LowStockThreshold,
			CategoryID:        req.CategoryID,
			Status:            req.Status,
			IsFeatured:        req.IsFeatured,
			Specifications:    specs,
		})
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		// 2. Create product images if provided
		images := make([]sqlc.ProductImage, 0, len(req.Images))
		for i, img := range req.Images {
			// First image is main by default if not specified
			isMain := img.IsMain
			if i == 0 && !anyImageIsMain(req.Images) {
				isMain = true
			}

			image, err := q.CreateProductImage(ctx, sqlc.CreateProductImageParams{
				ProductID: product.ID,
				ImageUrl:  img.ImageURL,
				Sort:      &img.Sort,
				IsMain:    &isMain,
			})
			if err != nil {
				return fmt.Errorf("failed to create product image: %w", err)
			}
			images = append(images, image)
		}

		result = toProductDetailResponse(product, images)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProduct retrieves a product by ID with all its images
func (s *service) GetProduct(ctx context.Context, productID int64) (*ProductDetailResponse, error) {
	// Get product
	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get product images
	images, err := s.repo.GetProductImages(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}

	// Increment view count asynchronously (don't block on error)
	go func() {
		_ = s.repo.IncrementProductViews(context.Background(), productID)
	}()

	response := toProductDetailResponse(product, images)
	return &response, nil
}

// UpdateProduct updates product information
func (s *service) UpdateProduct(ctx context.Context, productID int64, req UpdateProductRequest) (*ProductResponse, error) {
	// Check if product exists
	_, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update product
	err = s.repo.UpdateProduct(ctx, sqlc.UpdateProductParams{
		Name:        req.Name,
		Description: req.Description,
		Brand:       req.Brand,
		Price:       req.Price,
		OriginPrice: req.OriginPrice,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Status:      req.Status,
		IsFeatured:  req.IsFeatured,
		ID:          productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Get updated product
	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated product: %w", err)
	}

	// Get main image URL
	mainImageURL := ""
	mainImage, err := s.repo.GetProductMainImage(ctx, productID)
	if err == nil {
		mainImageURL = mainImage.ImageUrl
	}

	response := toProductResponse(product, mainImageURL)
	return &response, nil
}

// DeleteProduct soft deletes a product
func (s *service) DeleteProduct(ctx context.Context, productID int64) error {
	return s.repo.DeleteProduct(ctx, productID)
}

// ListProducts lists products with filtering and pagination
func (s *service) ListProducts(ctx context.Context, req ListProductsRequest) (*PaginatedProductsResponse, error) {
	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Get products
	products, err := s.repo.ListProducts(ctx, sqlc.ListProductsParams{
		CategoryID: req.CategoryID,
		Status:     req.Status,
		Limit:      req.PageSize,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Get total count
	total, err := s.repo.CountProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Get product IDs for batch loading images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	// Batch load main images
	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	// Convert to response
	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchProducts searches products by keyword
func (s *service) SearchProducts(ctx context.Context, req SearchProductsRequest) (*PaginatedProductsResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	keyword := "%" + req.Keyword + "%"
	products, err := s.repo.SearchProducts(ctx, sqlc.SearchProductsParams{
		Column1: &keyword,
		Limit:   req.PageSize,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Get main images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	// Note: For simplicity, we're not counting total search results
	// In production, you might want to add a CountSearchProducts query
	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      int64(len(products)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: 1,
	}, nil
}

// GetFeaturedProducts gets featured products
func (s *service) GetFeaturedProducts(ctx context.Context, page, pageSize int32) (*PaginatedProductsResponse, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	products, err := s.repo.ListFeaturedProducts(ctx, sqlc.ListFeaturedProductsParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get featured products: %w", err)
	}

	// Get main images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      int64(len(products)),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

// GetProductsByCategory gets products by category
func (s *service) GetProductsByCategory(ctx context.Context, categoryID int64, page, pageSize int32) (*PaginatedProductsResponse, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	products, err := s.repo.ListProductsByCategory(ctx, sqlc.ListProductsByCategoryParams{
		CategoryID: categoryID,
		Limit:      pageSize,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	// Get main images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      int64(len(products)),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}
// GetProductsByIDs gets products by IDs
func (s *service) GetProductsByIDs(ctx context.Context, productIDs []int64) (map[int64]*ProductResponse,error){

	if len(productIDs) == 0 {
		return make(map[int64]*ProductResponse), nil
	}

	result := make(map[int64]*ProductResponse)
	
	for _, id := range productIDs {
		product, err := s.repo.GetProductByID(ctx, id)
		if err != nil {
			if errors.Is(err,sql.ErrNoRows){
				continue
			}
			return nil,fmt.Errorf("failed to get product %d: %w",id,err)
		}

		if product.Status != "published" {
			continue
		}

		mainImageURL := ""
		mainImage, err := s.repo.GetProductMainImage(ctx, id)
		if err == nil {
			mainImageURL = mainImage.ImageUrl
		}

		response := toProductResponse(product, mainImageURL)
		result[id] = &response
	}

	
	return result, nil
}

// GetProductsByPriceRange gets products within price range
func (s *service) GetProductsByPriceRange(ctx context.Context, req PriceRangeRequest) (*PaginatedProductsResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	products, err := s.repo.ListProductsByPriceRange(ctx, sqlc.ListProductsByPriceRangeParams{
		Price:   req.MinPrice,
		Price_2: req.MaxPrice,
		Limit:   req.PageSize,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get products by price range: %w", err)
	}

	// Get main images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      int64(len(products)),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: 1,
	}, nil
}

// UpdateStock updates product stock
func (s *service) UpdateStock(ctx context.Context, productID int64, delta int32) error {
	return s.repo.UpdateProductStock(ctx, sqlc.UpdateProductStockParams{
		Stock: delta,
		ID:    productID,
	})
}

// CheckStock checks if product has sufficient stock
func (s *service) CheckStock(ctx context.Context, productID int64, quantity int32) (bool, error) {
	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return false, err
	}

	return product.Stock >= quantity, nil
}

// GetLowStockProducts gets products with low stock
func (s *service) GetLowStockProducts(ctx context.Context, page, pageSize int32) (*PaginatedProductsResponse, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	products, err := s.repo.GetLowStockProducts(ctx, sqlc.GetLowStockProductsParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Get main images
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	mainImages := make(map[int64]string)
	if len(productIDs) > 0 {
		images, err := s.repo.GetImagesByProductIDs(ctx, productIDs)
		if err == nil {
			for _, img := range images {
				if utils.PtrValue(img.IsMain) {
					mainImages[img.ProductID] = img.ImageUrl
				}
			}
		}
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = toProductResponse(p, mainImages[p.ID])
	}

	return &PaginatedProductsResponse{
		Products:   productResponses,
		Total:      int64(len(products)),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

// AddProductImages adds images to a product
func (s *service) AddProductImages(ctx context.Context, productID int64, images []ImageRequest) error {
	for _, img := range images {
		_, err := s.repo.CreateProductImage(ctx, sqlc.CreateProductImageParams{
			ProductID: productID,
			ImageUrl:  img.ImageURL,
			Sort:      &img.Sort,
			IsMain:    &img.IsMain,
		})
		if err != nil {
			return fmt.Errorf("failed to create product image: %w", err)
		}
	}
	return nil
}

// SetMainImage sets the main image for a product (atomic transaction)
func (s *service) SetMainImage(ctx context.Context, productID, imageID int64) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get all images for this product
		images, err := q.GetProductImages(ctx, productID)
		if err != nil {
			return fmt.Errorf("failed to get product images: %w", err)
		}

		// 2. Unset all main flags
		for _, img := range images {
			if utils.PtrValue(img.IsMain) {
				falseVal := false
				err := q.UpdateProductImage(ctx, sqlc.UpdateProductImageParams{
					ID:     img.ID,
					IsMain: &falseVal,
				})
				if err != nil {
					return fmt.Errorf("failed to unset main image: %w", err)
				}
			}
		}

		// 3. Set new main image
		trueVal := true
		err = q.UpdateProductImage(ctx, sqlc.UpdateProductImageParams{
			ID:     imageID,
			IsMain: &trueVal,
		})
		if err != nil {
			return fmt.Errorf("failed to set main image: %w", err)
		}

		return nil
	})
}

// DeleteProductImage deletes a product image
func (s *service) DeleteProductImage(ctx context.Context, imageID int64) error {
	return s.repo.DeleteProductImage(ctx, imageID)
}

// IncrementViews increments product view count
func (s *service) IncrementViews(ctx context.Context, productID int64) error {
	return s.repo.IncrementProductViews(ctx, productID)
}

// Helper functions

func stringToNullString(s string) *string {
	if s == "" {
		return nil
	}
	return utils.Ptr(s)
}

func anyImageIsMain(images []ImageRequest) bool {
	for _, img := range images {
		if img.IsMain {
			return true
		}
	}
	return false
}
