package inventory

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gomall/db/sqlc"
)

// Repository defines the interface for inventory data access
type Repository interface {
	// Inventory operations
	CreateInventory(ctx context.Context, arg sqlc.CreateInventoryParams) (sqlc.Inventory, error)
	GetInventoryByProductID(ctx context.Context, productID int64) (sqlc.Inventory, error)
	GetInventoryByID(ctx context.Context, id int64) (sqlc.Inventory, error)
	ListInventories(ctx context.Context, arg sqlc.ListInventoriesParams) ([]sqlc.Inventory, error)
	ListLowStockInventories(ctx context.Context, arg sqlc.ListLowStockInventoriesParams) ([]sqlc.Inventory, error)
	CountInventories(ctx context.Context) (int64, error)
	CountLowStockInventories(ctx context.Context) (int64, error)
	UpdateInventoryStock(ctx context.Context, arg sqlc.UpdateInventoryStockParams) error
	ReserveStock(ctx context.Context, arg sqlc.ReserveStockParams) error
	ReleaseReservedStock(ctx context.Context, arg sqlc.ReleaseReservedStockParams) error
	DeductReservedStock(ctx context.Context, arg sqlc.DeductReservedStockParams) error
	AddAvailableStock(ctx context.Context, arg sqlc.AddAvailableStockParams) error
	UpdateLowStockThreshold(ctx context.Context, arg sqlc.UpdateLowStockThresholdParams) error
	DeleteInventory(ctx context.Context, productID int64) error

	// Inventory log operations
	CreateInventoryLog(ctx context.Context, arg sqlc.CreateInventoryLogParams) (sqlc.InventoryLog, error)
	GetInventoryLogsByProductID(ctx context.Context, arg sqlc.GetInventoryLogsByProductIDParams) ([]sqlc.InventoryLog, error)
	GetInventoryLogsByOrderID(ctx context.Context, orderID int64) ([]sqlc.InventoryLog, error)
	CountInventoryLogsByProductID(ctx context.Context, productID int64) (int64, error)

	// Inventory reservation operations
	CreateInventoryReservation(ctx context.Context, arg sqlc.CreateInventoryReservationParams) (sqlc.InventoryReservation, error)
	GetInventoryReservationByID(ctx context.Context, id int64) (sqlc.InventoryReservation, error)
	GetInventoryReservationByOrderID(ctx context.Context, orderID int64) ([]sqlc.InventoryReservation, error)
	GetActiveReservationsByProductID(ctx context.Context, productID int64) ([]sqlc.InventoryReservation, error)
	UpdateReservationStatus(ctx context.Context, arg sqlc.UpdateReservationStatusParams) error
	ConfirmReservation(ctx context.Context, orderID int64) error
	CancelReservation(ctx context.Context, orderID int64) error
	GetExpiredReservations(ctx context.Context, limit int32) ([]sqlc.InventoryReservation, error)
	DeleteReservation(ctx context.Context, id int64) error

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

// Inventory operations

func (r *repository) CreateInventory(ctx context.Context, arg sqlc.CreateInventoryParams) (sqlc.Inventory, error) {
	return r.store.CreateInventory(ctx, arg)
}

func (r *repository) GetInventoryByProductID(ctx context.Context, productID int64) (sqlc.Inventory, error) {
	return r.store.GetInventoryByProductID(ctx, productID)
}

func (r *repository) GetInventoryByID(ctx context.Context, id int64) (sqlc.Inventory, error) {
	return r.store.GetInventoryByID(ctx, id)
}

func (r *repository) ListInventories(ctx context.Context, arg sqlc.ListInventoriesParams) ([]sqlc.Inventory, error) {
	return r.store.ListInventories(ctx, arg)
}

func (r *repository) ListLowStockInventories(ctx context.Context, arg sqlc.ListLowStockInventoriesParams) ([]sqlc.Inventory, error) {
	return r.store.ListLowStockInventories(ctx, arg)
}

func (r *repository) CountInventories(ctx context.Context) (int64, error) {
	return r.store.CountInventories(ctx)
}

func (r *repository) CountLowStockInventories(ctx context.Context) (int64, error) {
	return r.store.CountLowStockInventories(ctx)
}

func (r *repository) UpdateInventoryStock(ctx context.Context, arg sqlc.UpdateInventoryStockParams) error {
	return r.store.UpdateInventoryStock(ctx, arg)
}

func (r *repository) ReserveStock(ctx context.Context, arg sqlc.ReserveStockParams) error {
	return r.store.ReserveStock(ctx, arg)
}

func (r *repository) ReleaseReservedStock(ctx context.Context, arg sqlc.ReleaseReservedStockParams) error {
	return r.store.ReleaseReservedStock(ctx, arg)
}

func (r *repository) DeductReservedStock(ctx context.Context, arg sqlc.DeductReservedStockParams) error {
	return r.store.DeductReservedStock(ctx, arg)
}

func (r *repository) AddAvailableStock(ctx context.Context, arg sqlc.AddAvailableStockParams) error {
	return r.store.AddAvailableStock(ctx, arg)
}

func (r *repository) UpdateLowStockThreshold(ctx context.Context, arg sqlc.UpdateLowStockThresholdParams) error {
	return r.store.UpdateLowStockThreshold(ctx, arg)
}

func (r *repository) DeleteInventory(ctx context.Context, productID int64) error {
	return r.store.DeleteInventory(ctx, productID)
}

// Inventory log operations

func (r *repository) CreateInventoryLog(ctx context.Context, arg sqlc.CreateInventoryLogParams) (sqlc.InventoryLog, error) {
	return r.store.CreateInventoryLog(ctx, arg)
}

func (r *repository) GetInventoryLogsByProductID(ctx context.Context, arg sqlc.GetInventoryLogsByProductIDParams) ([]sqlc.InventoryLog, error) {
	return r.store.GetInventoryLogsByProductID(ctx, arg)
}

func (r *repository) GetInventoryLogsByOrderID(ctx context.Context, orderID int64) ([]sqlc.InventoryLog, error) {
	return r.store.GetInventoryLogsByOrderID(ctx, orderID)
}

func (r *repository) CountInventoryLogsByProductID(ctx context.Context, productID int64) (int64, error) {
	return r.store.CountInventoryLogsByProductID(ctx, productID)
}

// Inventory reservation operations

func (r *repository) CreateInventoryReservation(ctx context.Context, arg sqlc.CreateInventoryReservationParams) (sqlc.InventoryReservation, error) {
	return r.store.CreateInventoryReservation(ctx, arg)
}

func (r *repository) GetInventoryReservationByID(ctx context.Context, id int64) (sqlc.InventoryReservation, error) {
	return r.store.GetInventoryReservationByID(ctx, id)
}

func (r *repository) GetInventoryReservationByOrderID(ctx context.Context, orderID int64) ([]sqlc.InventoryReservation, error) {
	return r.store.GetInventoryReservationByOrderID(ctx, orderID)
}

func (r *repository) GetActiveReservationsByProductID(ctx context.Context, productID int64) ([]sqlc.InventoryReservation, error) {
	return r.store.GetActiveReservationsByProductID(ctx, productID)
}

func (r *repository) UpdateReservationStatus(ctx context.Context, arg sqlc.UpdateReservationStatusParams) error {
	return r.store.UpdateReservationStatus(ctx, arg)
}

func (r *repository) ConfirmReservation(ctx context.Context, orderID int64) error {
	return r.store.ConfirmReservation(ctx, orderID)
}

func (r *repository) CancelReservation(ctx context.Context, orderID int64) error {
	return r.store.CancelReservation(ctx, orderID)
}

func (r *repository) GetExpiredReservations(ctx context.Context, limit int32) ([]sqlc.InventoryReservation, error) {
	return r.store.GetExpiredReservations(ctx, limit)
}

func (r *repository) DeleteReservation(ctx context.Context, id int64) error {
	return r.store.DeleteReservation(ctx, id)
}

// Transaction support

func (r *repository) ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	return r.store.ExecTx(ctx, fn)
}
