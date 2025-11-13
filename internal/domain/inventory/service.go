package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gomall/db/sqlc"
	dberrors "gomall/db"
	"gomall/utils"
)

// Service defines the business logic interface for inventory domain
type Service interface {
	// Inventory CRUD operations
	CreateInventory(ctx context.Context, req CreateInventoryRequest) (*InventoryResponse, error)
	GetInventoryByProductID(ctx context.Context, productID int64) (*InventoryResponse, error)
	ListInventories(ctx context.Context, req ListInventoriesRequest) (*PaginatedInventoriesResponse, error)
	ListLowStockInventories(ctx context.Context, page, pageSize int32) (*PaginatedInventoriesResponse, error)
	UpdateLowStockThreshold(ctx context.Context, productID int64, req UpdateLowStockThresholdRequest) error

	// Stock operations with optimistic locking
	ReserveStock(ctx context.Context, req ReserveStockRequest, expiresInMinutes int) error
	ReleaseStock(ctx context.Context, req ReleaseStockRequest) error
	DeductStock(ctx context.Context, req DeductStockRequest) error
	RestockInventory(ctx context.Context, req RestockRequest, operatorID *int64) error
	AdjustStock(ctx context.Context, req AdjustStockRequest, operatorID *int64) error

	// Stock check operations
	CheckStockAvailability(ctx context.Context, productID int64, quantity int32) (*StockCheckResponse, error)
	BatchCheckStockAvailability(ctx context.Context, items []StockCheckItem) (map[int64]*StockCheckResponse, error)

	// Reservation operations
	ConfirmReservation(ctx context.Context, orderID int64) error
	CancelReservation(ctx context.Context, orderID int64) error
	CleanupExpiredReservations(ctx context.Context) error

	// Inventory log operations
	GetInventoryLogs(ctx context.Context, req ListInventoryLogsRequest) (*PaginatedInventoryLogsResponse, error)
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

// CreateInventory creates a new inventory record for a product
func (s *service) CreateInventory(ctx context.Context, req CreateInventoryRequest) (*InventoryResponse, error) {
	// Check if inventory already exists
	_, err := s.repo.GetInventoryByProductID(ctx, req.ProductID)
	if err == nil {
		return nil, errors.New("inventory already exists for this product")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing inventory: %w", err)
	}

	// Create inventory
	inventory, err := s.repo.CreateInventory(ctx, sqlc.CreateInventoryParams{
		ProductID:         req.ProductID,
		AvailableStock:    req.AvailableStock,
		ReservedStock:     req.ReservedStock,
		LowStockThreshold: utils.Ptr(req.LowStockThreshold),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	response := toInventoryResponse(inventory)
	return &response, nil
}

// GetInventoryByProductID retrieves inventory by product ID
func (s *service) GetInventoryByProductID(ctx context.Context, productID int64) (*InventoryResponse, error) {
	inventory, err := s.repo.GetInventoryByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("inventory not found")
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	response := toInventoryResponse(inventory)
	return &response, nil
}

// ListInventories lists all inventories with pagination
func (s *service) ListInventories(ctx context.Context, req ListInventoriesRequest) (*PaginatedInventoriesResponse, error) {
	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	inventories, err := s.repo.ListInventories(ctx, sqlc.ListInventoriesParams{
		Limit:  req.PageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	total, err := s.repo.CountInventories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count inventories: %w", err)
	}

	responses := make([]InventoryResponse, len(inventories))
	for i, inv := range inventories {
		responses[i] = toInventoryResponse(inv)
	}

	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &PaginatedInventoriesResponse{
		Inventories: responses,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
	}, nil
}

// ListLowStockInventories lists inventories with low stock
func (s *service) ListLowStockInventories(ctx context.Context, page, pageSize int32) (*PaginatedInventoriesResponse, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	inventories, err := s.repo.ListLowStockInventories(ctx, sqlc.ListLowStockInventoriesParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list low stock inventories: %w", err)
	}

	total, err := s.repo.CountLowStockInventories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count low stock inventories: %w", err)
	}

	responses := make([]InventoryResponse, len(inventories))
	for i, inv := range inventories {
		responses[i] = toInventoryResponse(inv)
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &PaginatedInventoriesResponse{
		Inventories: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// ReserveStock reserves stock for an order with optimistic locking (防止超卖)
func (s *service) ReserveStock(ctx context.Context, req ReserveStockRequest, expiresInMinutes int) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get current inventory
		inventory, err := q.GetInventoryByProductID(ctx, req.ProductID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("inventory not found")
			}
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		// 2. Check if enough stock available
		if inventory.AvailableStock < req.Quantity {
			return dberrors.ErrInsufficientStock
		}

		// 3. Reserve stock with optimistic locking
		err = q.ReserveStock(ctx, sqlc.ReserveStockParams{
			AvailableStock: req.Quantity,
			ProductID:      req.ProductID,
			Version:        inventory.Version,
		})
		if err != nil {
			return fmt.Errorf("failed to reserve stock (possible concurrent update): %w", err)
		}

		// 4. Create reservation record
		expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)
		_, err = q.CreateInventoryReservation(ctx, sqlc.CreateInventoryReservationParams{
			ProductID: req.ProductID,
			OrderID:   req.OrderID,
			Quantity:  req.Quantity,
			Status:    utils.Ptr("active"),
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create reservation: %w", err)
		}

		// 5. Log the operation
		_, err = q.CreateInventoryLog(ctx, sqlc.CreateInventoryLogParams{
			ProductID:       req.ProductID,
			OrderID:         &req.OrderID,
			ChangeType:      "reserve",
			QuantityChange:  req.Quantity,
			BeforeAvailable: inventory.AvailableStock,
			AfterAvailable:  inventory.AvailableStock - req.Quantity,
			BeforeReserved:  inventory.ReservedStock,
			AfterReserved:   inventory.ReservedStock + req.Quantity,
			Reason:          utils.Ptr("Stock reserved for order"),
			OperatorID:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create inventory log: %w", err)
		}

		return nil
	})
}

// ReleaseStock releases reserved stock (e.g., when order is cancelled)
func (s *service) ReleaseStock(ctx context.Context, req ReleaseStockRequest) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get current inventory
		inventory, err := q.GetInventoryByProductID(ctx, req.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		// 2. Release reserved stock with optimistic locking
		err = q.ReleaseReservedStock(ctx, sqlc.ReleaseReservedStockParams{
			AvailableStock: req.Quantity,
			ProductID:      req.ProductID,
			Version:        inventory.Version,
		})
		if err != nil {
			return fmt.Errorf("failed to release stock: %w", err)
		}

		// 3. Update reservation status
		err = q.CancelReservation(ctx, req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to cancel reservation: %w", err)
		}

		// 4. Log the operation
		_, err = q.CreateInventoryLog(ctx, sqlc.CreateInventoryLogParams{
			ProductID:       req.ProductID,
			OrderID:         &req.OrderID,
			ChangeType:      "release",
			QuantityChange:  -req.Quantity,
			BeforeAvailable: inventory.AvailableStock,
			AfterAvailable:  inventory.AvailableStock + req.Quantity,
			BeforeReserved:  inventory.ReservedStock,
			AfterReserved:   inventory.ReservedStock - req.Quantity,
			Reason:          utils.Ptr("Stock released from cancelled order"),
			OperatorID:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create inventory log: %w", err)
		}

		return nil
	})
}

// DeductStock deducts reserved stock (e.g., when order is confirmed/paid)
func (s *service) DeductStock(ctx context.Context, req DeductStockRequest) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get current inventory
		inventory, err := q.GetInventoryByProductID(ctx, req.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		// 2. Deduct reserved stock with optimistic locking
		err = q.DeductReservedStock(ctx, sqlc.DeductReservedStockParams{
			ReservedStock: req.Quantity,
			ProductID:     req.ProductID,
			Version:       inventory.Version,
		})
		if err != nil {
			return fmt.Errorf("failed to deduct stock: %w", err)
		}

		// 3. Confirm reservation
		err = q.ConfirmReservation(ctx, req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to confirm reservation: %w", err)
		}

		// 4. Log the operation
		_, err = q.CreateInventoryLog(ctx, sqlc.CreateInventoryLogParams{
			ProductID:       req.ProductID,
			OrderID:         &req.OrderID,
			ChangeType:      "deduct",
			QuantityChange:  -req.Quantity,
			BeforeAvailable: inventory.AvailableStock,
			AfterAvailable:  inventory.AvailableStock,
			BeforeReserved:  inventory.ReservedStock,
			AfterReserved:   inventory.ReservedStock - req.Quantity,
			Reason:          utils.Ptr("Stock deducted for confirmed order"),
			OperatorID:      nil,
		})
		if err != nil {
			return fmt.Errorf("failed to create inventory log: %w", err)
		}

		return nil
	})
}

// RestockInventory adds stock to inventory
func (s *service) RestockInventory(ctx context.Context, req RestockRequest, operatorID *int64) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get current inventory
		inventory, err := q.GetInventoryByProductID(ctx, req.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		// 2. Add stock
		err = q.AddAvailableStock(ctx, sqlc.AddAvailableStockParams{
			AvailableStock: req.Quantity,
			ProductID:      req.ProductID,
		})
		if err != nil {
			return fmt.Errorf("failed to add stock: %w", err)
		}

		// 3. Log the operation
		_, err = q.CreateInventoryLog(ctx, sqlc.CreateInventoryLogParams{
			ProductID:       req.ProductID,
			OrderID:         nil,
			ChangeType:      "restock",
			QuantityChange:  req.Quantity,
			BeforeAvailable: inventory.AvailableStock,
			AfterAvailable:  inventory.AvailableStock + req.Quantity,
			BeforeReserved:  inventory.ReservedStock,
			AfterReserved:   inventory.ReservedStock,
			Reason:          utils.Ptr(req.Reason),
			OperatorID:      operatorID,
		})
		if err != nil {
			return fmt.Errorf("failed to create inventory log: %w", err)
		}

		return nil
	})
}

// AdjustStock adjusts inventory (can be positive or negative)
func (s *service) AdjustStock(ctx context.Context, req AdjustStockRequest, operatorID *int64) error {
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get current inventory
		inventory, err := q.GetInventoryByProductID(ctx, req.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		// 2. Calculate new stock
		newAvailableStock := inventory.AvailableStock + req.Quantity
		if newAvailableStock < 0 {
			return errors.New("adjustment would result in negative stock")
		}

		// 3. Update stock
		err = q.UpdateInventoryStock(ctx, sqlc.UpdateInventoryStockParams{
			AvailableStock: newAvailableStock,
			ReservedStock:  inventory.ReservedStock,
			ProductID:      req.ProductID,
			Version:        inventory.Version,
		})
		if err != nil {
			return fmt.Errorf("failed to adjust stock: %w", err)
		}

		// 4. Log the operation
		_, err = q.CreateInventoryLog(ctx, sqlc.CreateInventoryLogParams{
			ProductID:       req.ProductID,
			OrderID:         nil,
			ChangeType:      "adjust",
			QuantityChange:  req.Quantity,
			BeforeAvailable: inventory.AvailableStock,
			AfterAvailable:  newAvailableStock,
			BeforeReserved:  inventory.ReservedStock,
			AfterReserved:   inventory.ReservedStock,
			Reason:          utils.Ptr(req.Reason),
			OperatorID:      operatorID,
		})
		if err != nil {
			return fmt.Errorf("failed to create inventory log: %w", err)
		}

		return nil
	})
}

// CheckStockAvailability checks if stock is available
func (s *service) CheckStockAvailability(ctx context.Context, productID int64, quantity int32) (*StockCheckResponse, error) {
	inventory, err := s.repo.GetInventoryByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("inventory not found")
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return &StockCheckResponse{
		ProductID:      productID,
		AvailableStock: inventory.AvailableStock,
		ReservedStock:  inventory.ReservedStock,
		IsAvailable:    inventory.AvailableStock >= quantity,
		RequestedQty:   quantity,
	}, nil
}

// BatchCheckStockAvailability checks stock availability for multiple items
func (s *service) BatchCheckStockAvailability(ctx context.Context, items []StockCheckItem) (map[int64]*StockCheckResponse, error) {
	result := make(map[int64]*StockCheckResponse)

	for _, item := range items {
		check, err := s.CheckStockAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		result[item.ProductID] = check
	}

	return result, nil
}

// ConfirmReservation confirms a reservation (usually after payment)
func (s *service) ConfirmReservation(ctx context.Context, orderID int64) error {
	return s.repo.ConfirmReservation(ctx, orderID)
}

// CancelReservation cancels a reservation
func (s *service) CancelReservation(ctx context.Context, orderID int64) error {
	return s.repo.CancelReservation(ctx, orderID)
}

// CleanupExpiredReservations cleans up expired reservations
func (s *service) CleanupExpiredReservations(ctx context.Context) error {
	// Get expired reservations in batches
	expiredReservations, err := s.repo.GetExpiredReservations(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get expired reservations: %w", err)
	}

	for _, reservation := range expiredReservations {
		// Release stock for each expired reservation
		err = s.ReleaseStock(ctx, ReleaseStockRequest{
			ProductID: reservation.ProductID,
			Quantity:  reservation.Quantity,
			OrderID:   reservation.OrderID,
		})
		if err != nil {
			// Log error but continue processing
			fmt.Printf("failed to release stock for expired reservation %d: %v\n", reservation.ID, err)
		}
	}

	return nil
}

// GetInventoryLogs retrieves inventory logs
func (s *service) GetInventoryLogs(ctx context.Context, req ListInventoryLogsRequest) (*PaginatedInventoryLogsResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	logs, err := s.repo.GetInventoryLogsByProductID(ctx, sqlc.GetInventoryLogsByProductIDParams{
		ProductID: req.ProductID,
		Limit:     req.PageSize,
		Offset:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory logs: %w", err)
	}

	total, err := s.repo.CountInventoryLogsByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to count inventory logs: %w", err)
	}

	responses := make([]InventoryLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = toInventoryLogResponse(log)
	}

	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &PaginatedInventoryLogsResponse{
		Logs:       responses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateLowStockThreshold updates the low stock threshold
func (s *service) UpdateLowStockThreshold(ctx context.Context, productID int64, req UpdateLowStockThresholdRequest) error {
	return s.repo.UpdateLowStockThreshold(ctx, sqlc.UpdateLowStockThresholdParams{
		LowStockThreshold: utils.Ptr(req.Threshold),
		ProductID:         productID,
	})
}
