package inventory

import (
	"gomall/db/sqlc"
	"gomall/utils"
	"time"
)

// Request DTOs

type CreateInventoryRequest struct {
	ProductID         int64 `json:"product_id" binding:"required"`
	AvailableStock    int32 `json:"available_stock" binding:"required,min=0"`
	ReservedStock     int32 `json:"reserved_stock,omitempty" binding:"min=0"`
	LowStockThreshold int32 `json:"low_stock_threshold,omitempty" binding:"min=0"`
}

type UpdateInventoryStockRequest struct {
	AvailableStock int32 `json:"available_stock" binding:"required,min=0"`
	ReservedStock  int32 `json:"reserved_stock,omitempty" binding:"min=0"`
}

type StockCheckItem struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
}


type ReserveStockRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
	OrderID   int64 `json:"order_id" binding:"required"`
}

type ReleaseStockRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
	OrderID   int64 `json:"order_id" binding:"required"`
}

type DeductStockRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
	OrderID   int64 `json:"order_id" binding:"required"`
}

type RestockRequest struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required,min=1"`
	Reason    string `json:"reason,omitempty" binding:"max=500"`
}

type AdjustStockRequest struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required"`
	Reason    string `json:"reason" binding:"required,max=500"`
}

type UpdateLowStockThresholdRequest struct {
	Threshold int32 `json:"threshold" binding:"required,min=0"`
}

type ListInventoriesRequest struct {
	Page     int32 `form:"page" binding:"min=1"`
	PageSize int32 `form:"page_size" binding:"min=1,max=100"`
}

type ListInventoryLogsRequest struct {
	ProductID int64 `form:"product_id" binding:"required"`
	Page      int32 `form:"page" binding:"min=1"`
	PageSize  int32 `form:"page_size" binding:"min=1,max=100"`
}

// Response DTOs

type InventoryResponse struct {
	ID                int64     `json:"id"`
	ProductID         int64     `json:"product_id"`
	AvailableStock    int32     `json:"available_stock"`
	ReservedStock     int32     `json:"reserved_stock"`
	TotalStock        int32     `json:"total_stock"`
	LowStockThreshold int32     `json:"low_stock_threshold"`
	IsLowStock        bool      `json:"is_low_stock"`
	Version           int64     `json:"version"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type InventoryLogResponse struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	OrderID         *int64    `json:"order_id,omitempty"`
	ChangeType      string    `json:"change_type"`
	QuantityChange  int32     `json:"quantity_change"`
	BeforeAvailable int32     `json:"before_available"`
	AfterAvailable  int32     `json:"after_available"`
	BeforeReserved  int32     `json:"before_reserved"`
	AfterReserved   int32     `json:"after_reserved"`
	Reason          string    `json:"reason,omitempty"`
	OperatorID      *int64    `json:"operator_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type InventoryReservationResponse struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	OrderID   int64     `json:"order_id"`
	Quantity  int32     `json:"quantity"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaginatedInventoriesResponse struct {
	Inventories []InventoryResponse `json:"inventories"`
	Total       int64               `json:"total"`
	Page        int32               `json:"page"`
	PageSize    int32               `json:"page_size"`
	TotalPages  int32               `json:"total_pages"`
}

type PaginatedInventoryLogsResponse struct {
	Logs       []InventoryLogResponse `json:"logs"`
	Total      int64                  `json:"total"`
	Page       int32                  `json:"page"`
	PageSize   int32                  `json:"page_size"`
	TotalPages int32                  `json:"total_pages"`
}

type StockCheckResponse struct {
	ProductID      int64 `json:"product_id"`
	AvailableStock int32 `json:"available_stock"`
	ReservedStock  int32 `json:"reserved_stock"`
	IsAvailable    bool  `json:"is_available"`
	RequestedQty   int32 `json:"requested_qty"`
}

// Conversion functions

func toInventoryResponse(inv sqlc.Inventory) InventoryResponse {
	totalStock := inv.AvailableStock + inv.ReservedStock
	threshold := utils.PtrValue(inv.LowStockThreshold)
	isLowStock := inv.AvailableStock <= threshold

	return InventoryResponse{
		ID:                inv.ID,
		ProductID:         inv.ProductID,
		AvailableStock:    inv.AvailableStock,
		ReservedStock:     inv.ReservedStock,
		TotalStock:        totalStock,
		LowStockThreshold: threshold,
		IsLowStock:        isLowStock,
		Version:           inv.Version,
		CreatedAt:         inv.CreatedAt,
		UpdatedAt:         inv.UpdatedAt,
	}
}

func toInventoryLogResponse(log sqlc.InventoryLog) InventoryLogResponse {
	return InventoryLogResponse{
		ID:              log.ID,
		ProductID:       log.ProductID,
		OrderID:         log.OrderID,
		ChangeType:      log.ChangeType,
		QuantityChange:  log.QuantityChange,
		BeforeAvailable: log.BeforeAvailable,
		AfterAvailable:  log.AfterAvailable,
		BeforeReserved:  log.BeforeReserved,
		AfterReserved:   log.AfterReserved,
		Reason:          utils.PtrValue(log.Reason),
		OperatorID:      log.OperatorID,
		CreatedAt:       log.CreatedAt,
	}
}

func toInventoryReservationResponse(res sqlc.InventoryReservation) InventoryReservationResponse {
	return InventoryReservationResponse{
		ID:        res.ID,
		ProductID: res.ProductID,
		OrderID:   res.OrderID,
		Quantity:  res.Quantity,
		Status:    utils.PtrValue(res.Status),
		ExpiresAt: res.ExpiresAt,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}
}
