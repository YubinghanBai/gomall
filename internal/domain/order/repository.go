package order

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gomall/db/sqlc"
)

// Repository defines the interface for order data access
type Repository interface {
	// CreateOrder Order operations
	CreateOrder(ctx context.Context, arg sqlc.CreateOrderParams) (sqlc.Order, error)
	GetOrderByID(ctx context.Context, id int64) (sqlc.Order, error)
	GetOrderByOrderNo(ctx context.Context, orderNo string) (sqlc.Order, error)
	ListUserOrders(ctx context.Context, arg sqlc.ListUserOrdersParams) ([]sqlc.Order, error)
	CountUserOrders(ctx context.Context, userID int64) (int64, error)
	UpdateOrderStatus(ctx context.Context, arg sqlc.UpdateOrderStatusParams) error
	UpdateOrderPaymentStatus(ctx context.Context, arg sqlc.UpdateOrderPaymentStatusParams) error
	UpdateOrderShipStatus(ctx context.Context, arg sqlc.UpdateOrderShipStatusParams) error
	CancelOrder(ctx context.Context, id int64) error

	// Order item operations
	CreateOrderItem(ctx context.Context, arg sqlc.CreateOrderItemParams) (sqlc.OrderItem, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]sqlc.OrderItem, error)
	GetOrderItemsByIDs(ctx context.Context, orderIDs []int64) ([]sqlc.OrderItem, error)

	// Transaction support
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

func (r *repository) CreateOrder(ctx context.Context, arg sqlc.CreateOrderParams) (sqlc.Order, error) {
	return r.store.CreateOrder(ctx, arg)
}

func (r *repository) GetOrderByID(ctx context.Context, id int64) (sqlc.Order, error) {
	return r.store.GetOrderByID(ctx, id)
}

func (r *repository) GetOrderByOrderNo(ctx context.Context, orderNo string) (sqlc.Order, error) {
	return r.store.GetOrderByOrderNo(ctx, orderNo)
}

func (r *repository) ListUserOrders(ctx context.Context, arg sqlc.ListUserOrdersParams) ([]sqlc.Order, error) {
	return r.store.ListUserOrders(ctx, arg)
}

func (r *repository) CountUserOrders(ctx context.Context, userID int64) (int64, error) {
	return r.store.CountUserOrders(ctx, userID)
}

func (r *repository) UpdateOrderStatus(ctx context.Context, arg sqlc.UpdateOrderStatusParams) error {
	return r.store.UpdateOrderStatus(ctx, arg)
}

func (r *repository) UpdateOrderPaymentStatus(ctx context.Context, arg sqlc.UpdateOrderPaymentStatusParams) error {
	return r.store.UpdateOrderPaymentStatus(ctx, arg)
}

func (r *repository) UpdateOrderShipStatus(ctx context.Context, arg sqlc.UpdateOrderShipStatusParams) error {
	return r.store.UpdateOrderShipStatus(ctx, arg)
}

func (r *repository) CancelOrder(ctx context.Context, id int64) error {
	return r.store.CancelOrder(ctx, id)
}

func (r *repository) CreateOrderItem(ctx context.Context, arg sqlc.CreateOrderItemParams) (sqlc.OrderItem, error) {
	return r.store.CreateOrderItem(ctx, arg)
}

func (r *repository) GetOrderItems(ctx context.Context, orderID int64) ([]sqlc.OrderItem, error) {
	return r.store.GetOrderItems(ctx, orderID)
}

func (r *repository) GetOrderItemsByIDs(ctx context.Context, orderIDs []int64) ([]sqlc.OrderItem, error) {
	return r.store.GetOrderItemsByIDs(ctx, orderIDs)
}

func (r *repository) ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	return r.store.ExecTx(ctx, fn)
}
