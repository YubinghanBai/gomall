package product

import (
	"gomall/db/sqlc"
	"gomall/utils"
	"time"
)

// ImageRequest Image represents a product image
type ImageRequest struct {
	ImageURL string `json:"image_url" binding:"required,url"`
	Sort     int32  `json:"sort"`
	IsMain   bool   `json:"is_main"`
}

type ImageResponse struct {
	ID       int64     `json:"id"`
	ImageURL string    `json:"image_url"`
	Sort     int32     `json:"sort"`
	IsMain   bool      `json:"is_main"`
	CreateAt time.Time `json:"created_at"`
}

// Product Request DTOs

type CreateProductRequest struct {
	Name              string         `json:"name" binding:"required,min=1,max=200"`
	Description       string         `json:"description"`
	Brand             string         `json:"brand" binding:"max=100"`
	Price             int64          `json:"price" binding:"required,min=0"`
	OriginPrice       int64          `json:"origin_price" binding:"min=0"`
	CostPrice         int64          `json:"cost_price" binding:"min=0"`
	Stock             int32          `json:"stock" binding:"required,min=0"`
	LowStockThreshold int32          `json:"low_stock_threshold" binding:"min=0"`
	CategoryID        int64          `json:"category_id" binding:"required"`
	Status            string         `json:"status" binding:"required,oneof=draft published off_shelf"`
	IsFeatured        bool           `json:"is_featured"`
	Specifications    string         `json:"specifications"`
	Images            []ImageRequest `json:"images"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty"`
	Brand       *string `json:"brand,omitempty" binding:"omitempty,max=100"`
	Price       *int64  `json:"price,omitempty" binding:"omitempty,min=0"`
	OriginPrice *int64  `json:"origin_price,omitempty" binding:"omitempty,min=0"`
	Stock       *int32  `json:"stock,omitempty" binding:"omitempty,min=0"`
	CategoryID  *int64  `json:"category_id,omitempty"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=draft published off_shelf"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
}

type UpdateStockRequest struct {
	Delta int32 `json:"delta" binding:"required"`
}

type ListProductsRequest struct {
	CategoryID *int64  `form:"category_id"`
	Status     *string `form:"status" binding:"omitempty,oneof=draft published off_shelf"`
	Page       int32   `form:"page" binding:"min=1"`
	PageSize   int32   `form:"page_size" binding:"min=1,max=100"`
}

type SearchProductsRequest struct {
	Keyword  string `form:"keyword" binding:"required,min=1"`
	Page     int32  `form:"page" binding:"min=1"`
	PageSize int32  `form:"page_size" binding:"min=1,max=100"`
}

type PriceRangeRequest struct {
	MinPrice int64 `form:"min_price" binding:"min=0"`
	MaxPrice int64 `form:"max_price" binding:"min=0"`
	Page     int32 `form:"page" binding:"min=1"`
	PageSize int32 `form:"page_size" binding:"min=1,max=100"`
}

// Product Response DTOs

type ProductResponse struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Brand             string    `json:"brand"`
	Price             int64     `json:"price"`
	OriginPrice       int64     `json:"origin_price"`
	Stock             int32     `json:"stock"`
	LowStockThreshold int32     `json:"low_stock_threshold"`
	SalesCount        int32     `json:"sales_count"`
	ViewCount         int32     `json:"view_count"`
	CategoryID        int64     `json:"category_id"`
	Status            string    `json:"status"`
	IsFeatured        bool      `json:"is_featured"`
	Specifications    string    `json:"specifications,omitempty"`
	MainImage         string    `json:"main_image,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ProductDetailResponse struct {
	ID                int64           `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Brand             string          `json:"brand"`
	Price             int64           `json:"price"`
	OriginPrice       int64           `json:"origin_price"`
	CostPrice         int64           `json:"cost_price"`
	Stock             int32           `json:"stock"`
	LowStockThreshold int32           `json:"low_stock_threshold"`
	SalesCount        int32           `json:"sales_count"`
	ViewCount         int32           `json:"view_count"`
	CategoryID        int64           `json:"category_id"`
	Status            string          `json:"status"`
	IsFeatured        bool            `json:"is_featured"`
	Specifications    string          `json:"specifications,omitempty"`
	Images            []ImageResponse `json:"images"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type PaginatedProductsResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int64             `json:"total"`
	Page       int32             `json:"page"`
	PageSize   int32             `json:"page_size"`
	TotalPages int32             `json:"total_pages"`
}

// Conversion functions

func toProductResponse(product sqlc.Product, mainImageURL string) ProductResponse {
	specs := ""
	if len(product.Specifications) > 0 {
		specs = string(product.Specifications)
	}

	return ProductResponse{
		ID:                product.ID,
		Name:              product.Name,
		Description:       utils.PtrValue(product.Description),
		Brand:             utils.PtrValue(product.Brand),
		Price:             product.Price,
		OriginPrice:       product.OriginPrice,
		Stock:             product.Stock,
		LowStockThreshold: product.LowStockThreshold,
		SalesCount:        product.SalesCount,
		ViewCount:         product.ViewCount,
		CategoryID:        product.CategoryID,
		Status:            product.Status,
		IsFeatured:        product.IsFeatured,
		Specifications:    specs,
		MainImage:         mainImageURL,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}

func toProductDetailResponse(product sqlc.Product, images []sqlc.ProductImage) ProductDetailResponse {
	specs := ""
	if len(product.Specifications) > 0 {
		specs = string(product.Specifications)
	}

	imageResponses := make([]ImageResponse, 0, len(images))
	for _, img := range images {
		imageResponses = append(imageResponses, ImageResponse{
			ID:       img.ID,
			ImageURL: img.ImageUrl,
			Sort:     utils.PtrValue(img.Sort),
			IsMain:   utils.PtrValue(img.IsMain),
			CreateAt: img.CreatedAt,
		})
	}

	return ProductDetailResponse{
		ID:                product.ID,
		Name:              product.Name,
		Description:       utils.PtrValue(product.Description),
		Brand:             utils.PtrValue(product.Brand),
		Price:             product.Price,
		OriginPrice:       product.OriginPrice,
		CostPrice:         utils.PtrValue(product.CostPrice),
		Stock:             product.Stock,
		LowStockThreshold: product.LowStockThreshold,
		SalesCount:        product.SalesCount,
		ViewCount:         product.ViewCount,
		CategoryID:        product.CategoryID,
		Status:            product.Status,
		IsFeatured:        product.IsFeatured,
		Specifications:    specs,
		Images:            imageResponses,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}
