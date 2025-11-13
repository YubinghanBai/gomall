package category

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gomall/db/sqlc"
)

// Repository defines the interface for category data access
type Repository interface {
	// CreateCategory CRUD operations
	CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error)
	GetCategoryByID(ctx context.Context, id int64) (sqlc.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (sqlc.Category, error)
	UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) error
	DeleteCategory(ctx context.Context, id int64) error

	// ListCategories Query operations
	ListCategories(ctx context.Context, isActive *bool) ([]sqlc.Category, error)
	GetRootCategories(ctx context.Context) ([]sqlc.Category, error)
	GetCategoryChildren(ctx context.Context, parentID int64) ([]sqlc.Category, error)

	// CountCategoryChildren Business checks
	CountCategoryChildren(ctx context.Context, parentID int64) (int64, error)
	CountProductsByCategory(ctx context.Context, categoryID int64) (int64, error)

	// ExecTx Transaction support
	ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error
}

type repository struct {
	store sqlc.Store
}

// NewRepository creates a new Repository instance
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		store: sqlc.NewStore(pool),
	}
}

func (r *repository) CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	return r.store.CreateCategory(ctx, arg)
}

func (r *repository) GetCategoryByID(ctx context.Context, id int64) (sqlc.Category, error) {
	return r.store.GetCategoryByID(ctx, id)
}

func (r *repository) GetCategoryBySlug(ctx context.Context, slug string) (sqlc.Category, error) {
	return r.store.GetCategoryBySlug(ctx, &slug)
}

func (r *repository) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) error {
	return r.store.UpdateCategory(ctx, arg)
}

func (r *repository) DeleteCategory(ctx context.Context, id int64) error {
	return r.store.DeleteCategory(ctx, id)
}

func (r *repository) ListCategories(ctx context.Context, isActive *bool) ([]sqlc.Category, error) {
	if isActive == nil {
		return r.store.ListCategories(ctx, false)
	}
	return r.store.ListCategories(ctx, *isActive)
}

func (r *repository) GetRootCategories(ctx context.Context) ([]sqlc.Category, error) {
	return r.store.GetRootCategories(ctx)
}

func (r *repository) GetCategoryChildren(ctx context.Context, parentID int64) ([]sqlc.Category, error) {
	return r.store.GetCategoryChildren(ctx, &parentID)
}

func (r *repository) CountCategoryChildren(ctx context.Context, parentID int64) (int64, error) {
	return r.store.CountCategoryChildren(ctx, &parentID)
}

func (r *repository) CountProductsByCategory(ctx context.Context, categoryID int64) (int64, error) {
	return r.store.CountProductsByCategory(ctx, categoryID)
}

func (r *repository) ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	return r.store.ExecTx(ctx, fn)
}
