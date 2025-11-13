package product

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"

	"gomall/db/sqlc"
)

// Repository defines the data access interface for product domain
// This interface only includes product-related methods to maintain clear domain boundaries
type Repository interface {
	// CreateProduct Product CRUD operations
	CreateProduct(ctx context.Context, arg sqlc.CreateProductParams) (sqlc.Product, error)
	GetProductByID(ctx context.Context, id int64) (sqlc.Product, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]sqlc.Product, error)
	UpdateProduct(ctx context.Context, arg sqlc.UpdateProductParams) error
	DeleteProduct(ctx context.Context, id int64) error

	// ListProducts Product listing and search
	ListProducts(ctx context.Context, arg sqlc.ListProductsParams) ([]sqlc.Product, error)
	ListProductsByCategory(ctx context.Context, arg sqlc.ListProductsByCategoryParams) ([]sqlc.Product, error)
	ListFeaturedProducts(ctx context.Context, arg sqlc.ListFeaturedProductsParams) ([]sqlc.Product, error)
	ListProductsByPriceRange(ctx context.Context, arg sqlc.ListProductsByPriceRangeParams) ([]sqlc.Product, error)
	SearchProducts(ctx context.Context, arg sqlc.SearchProductsParams) ([]sqlc.Product, error)
	CountProducts(ctx context.Context) (int64, error)

	// UpdateProductStock Stock management operations
	UpdateProductStock(ctx context.Context, arg sqlc.UpdateProductStockParams) error
	DecrementProductStock(ctx context.Context, arg sqlc.DecrementProductStockParams) error
	UpdateProductStockWithVersion(ctx context.Context, arg sqlc.UpdateProductStockWithVersionParams) (sqlc.Product, error)
	GetLowStockProducts(ctx context.Context, arg sqlc.GetLowStockProductsParams) ([]sqlc.Product, error)

	// IncrementProductViews Statistics operations
	IncrementProductViews(ctx context.Context, id int64) error
	IncrementProductSales(ctx context.Context, arg sqlc.IncrementProductSalesParams) error

	// UpdateProductsStatus Batch operations
	UpdateProductsStatus(ctx context.Context, arg sqlc.UpdateProductsStatusParams) error

	// CreateProductImage Product image operations
	CreateProductImage(ctx context.Context, arg sqlc.CreateProductImageParams) (sqlc.ProductImage, error)
	GetProductImages(ctx context.Context, productID int64) ([]sqlc.ProductImage, error)
	GetProductMainImage(ctx context.Context, productID int64) (sqlc.ProductImage, error)
	GetImagesByProductIDs(ctx context.Context, productIDs []int64) ([]sqlc.ProductImage, error)
	UpdateProductImage(ctx context.Context, arg sqlc.UpdateProductImageParams) error
	DeleteProductImage(ctx context.Context, id int64) error
	DeleteProductImages(ctx context.Context, productID int64) error

	// ExecTx Transaction support
	ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error
}

// repository implements Repository interface
type repository struct {
	store sqlc.Store
}

// NewRepository creates a new Repository instance
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		store: sqlc.NewStore(pool),
	}
}

// Product CRUD operations

func (r *repository) CreateProduct(ctx context.Context, arg sqlc.CreateProductParams) (sqlc.Product, error) {
	return r.store.CreateProduct(ctx, arg)
}

func (r *repository) GetProductByID(ctx context.Context, id int64) (sqlc.Product, error) {
	return r.store.GetProductByID(ctx, id)
}

func (r *repository) GetProductsByIDs(ctx context.Context, ids []int64) ([]sqlc.Product, error) {
	return r.store.GetProductsByIDs(ctx, ids)
}

func (r *repository) UpdateProduct(ctx context.Context, arg sqlc.UpdateProductParams) error {
	return r.store.UpdateProduct(ctx, arg)
}

func (r *repository) DeleteProduct(ctx context.Context, id int64) error {
	return r.store.DeleteProduct(ctx, id)
}

// Product listing and search

func (r *repository) ListProducts(ctx context.Context, arg sqlc.ListProductsParams) ([]sqlc.Product, error) {
	return r.store.ListProducts(ctx, arg)
}

func (r *repository) ListProductsByCategory(ctx context.Context, arg sqlc.ListProductsByCategoryParams) ([]sqlc.Product, error) {
	return r.store.ListProductsByCategory(ctx, arg)
}

func (r *repository) ListFeaturedProducts(ctx context.Context, arg sqlc.ListFeaturedProductsParams) ([]sqlc.Product, error) {
	return r.store.ListFeaturedProducts(ctx, arg)
}

func (r *repository) ListProductsByPriceRange(ctx context.Context, arg sqlc.ListProductsByPriceRangeParams) ([]sqlc.Product, error) {
	return r.store.ListProductsByPriceRange(ctx, arg)
}

func (r *repository) SearchProducts(ctx context.Context, arg sqlc.SearchProductsParams) ([]sqlc.Product, error) {
	return r.store.SearchProducts(ctx, arg)
}

func (r *repository) CountProducts(ctx context.Context) (int64, error) {
	return r.store.CountProducts(ctx)
}

// Stock management operations

func (r *repository) UpdateProductStock(ctx context.Context, arg sqlc.UpdateProductStockParams) error {
	return r.store.UpdateProductStock(ctx, arg)
}

func (r *repository) DecrementProductStock(ctx context.Context, arg sqlc.DecrementProductStockParams) error {
	return r.store.DecrementProductStock(ctx, arg)
}

func (r *repository) UpdateProductStockWithVersion(ctx context.Context, arg sqlc.UpdateProductStockWithVersionParams) (sqlc.Product, error) {
	return r.store.UpdateProductStockWithVersion(ctx, arg)
}

func (r *repository) GetLowStockProducts(ctx context.Context, arg sqlc.GetLowStockProductsParams) ([]sqlc.Product, error) {
	return r.store.GetLowStockProducts(ctx, arg)
}

// Statistics operations

func (r *repository) IncrementProductViews(ctx context.Context, id int64) error {
	return r.store.IncrementProductViews(ctx, id)
}

func (r *repository) IncrementProductSales(ctx context.Context, arg sqlc.IncrementProductSalesParams) error {
	return r.store.IncrementProductSales(ctx, arg)
}

// Batch operations

func (r *repository) UpdateProductsStatus(ctx context.Context, arg sqlc.UpdateProductsStatusParams) error {
	return r.store.UpdateProductsStatus(ctx, arg)
}

// Product image operations

func (r *repository) CreateProductImage(ctx context.Context, arg sqlc.CreateProductImageParams) (sqlc.ProductImage, error) {
	return r.store.CreateProductImage(ctx, arg)
}

func (r *repository) GetProductImages(ctx context.Context, productID int64) ([]sqlc.ProductImage, error) {
	return r.store.GetProductImages(ctx, productID)
}

func (r *repository) GetProductMainImage(ctx context.Context, productID int64) (sqlc.ProductImage, error) {
	return r.store.GetProductMainImage(ctx, productID)
}

func (r *repository) GetImagesByProductIDs(ctx context.Context, productIDs []int64) ([]sqlc.ProductImage, error) {
	return r.store.GetImagesByProductIDs(ctx, productIDs)
}

func (r *repository) UpdateProductImage(ctx context.Context, arg sqlc.UpdateProductImageParams) error {
	return r.store.UpdateProductImage(ctx, arg)
}

func (r *repository) DeleteProductImage(ctx context.Context, id int64) error {
	return r.store.DeleteProductImage(ctx, id)
}

func (r *repository) DeleteProductImages(ctx context.Context, productID int64) error {
	return r.store.DeleteProductImages(ctx, productID)
}

// Transaction support

func (r *repository) ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	return r.store.ExecTx(ctx, fn)
}
