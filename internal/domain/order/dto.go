package order

import (
	sqlc "gomall/db/sqlc"
	"gomall/utils"
	"time"
)

// Request DTOs

type OrderItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	ReceiverName    string             `json:"receiver_name" binding:"required,min=1,max=50"`
	ReceiverPhone   string             `json:"receiver_phone" binding:"required,min=1,max=20"`
	ReceiverAddress string             `json:"receiver_address" binding:"required,min=1,max=500"`
	ReceiverZipCode string             `json:"receiver_zip_code,omitempty" binding:"omitempty,max=20"`
	Remark          string             `json:"remark,omitempty"`
	DiscountAmount  int64              `json:"discount_amount,omitempty" binding:"min=0"`
	ShippingFee     int64              `json:"shipping_fee,omitempty" binding:"min=0"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending paid shipped completed cancelled refunded"`
}

type ListOrdersRequest struct {
	Page     int32 `form:"page" binding:"min=1"`
	PageSize int32 `form:"page_size" binding:"min=1,max=100"`
}

// Response DTOs

type OrderItemResponse struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image,omitempty"`
	Quantity     int32  `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	TotalPrice   int64  `json:"total_price"`
}

type OrderResponse struct {
	ID              int64               `json:"id"`
	OrderNo         string              `json:"order_no"`
	UserID          int64               `json:"user_id"`
	TotalAmount     int64               `json:"total_amount"`
	DiscountAmount  int64               `json:"discount_amount"`
	ShippingFee     int64               `json:"shipping_fee"`
	PayAmount       int64               `json:"pay_amount"`
	Status          string              `json:"status"`
	PaymentStatus   string              `json:"payment_status"`
	ShipStatus      string              `json:"ship_status"`
	ReceiverName    string              `json:"receiver_name"`
	ReceiverPhone   string              `json:"receiver_phone"`
	ReceiverAddress string              `json:"receiver_address"`
	ReceiverZipCode string              `json:"receiver_zip_code,omitempty"`
	Remark          string              `json:"remark,omitempty"`
	PaidAt          *time.Time          `json:"paid_at,omitempty"`
	ShippedAt       *time.Time          `json:"shipped_at,omitempty"`
	CompletedAt     *time.Time          `json:"completed_at,omitempty"`
	CancelledAt     *time.Time          `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Items           []OrderItemResponse `json:"items,omitempty"`
}

type PaginatedOrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	Total      int64           `json:"total"`
	Page       int32           `json:"page"`
	PageSize   int32           `json:"page_size"`
	TotalPages int32           `json:"total_pages"`
}

// Conversion functions

func toOrderResponse(order sqlc.Order, items []sqlc.OrderItem) OrderResponse {
	itemResponses := make([]OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: utils.PtrValue(item.ProductImage),
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			TotalPrice:   item.TotalPrice,
		}
	}

	return OrderResponse{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		DiscountAmount:  order.DiscountAmount,
		ShippingFee:     order.ShippingFee,
		PayAmount:       order.PayAmount,
		Status:          order.Status,
		PaymentStatus:   order.PaymentStatus,
		ShipStatus:      order.ShipStatus,
		ReceiverName:    order.ReceiverName,
		ReceiverPhone:   order.ReceiverPhone,
		ReceiverAddress: order.ReceiverAddress,
		ReceiverZipCode: utils.PtrValue(order.ReceiverZipCode),
		Remark:          utils.PtrValue(order.Remark),
		PaidAt:          order.PaidAt.Ptr(),
		ShippedAt:       order.ShippedAt.Ptr(),
		CompletedAt:     order.CompletedAt.Ptr(),
		CancelledAt:     order.CancelledAt.Ptr(),
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           itemResponses,
	}
}

func toOrderResponseWithoutItems(order sqlc.Order) OrderResponse {
	return OrderResponse{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		DiscountAmount:  order.DiscountAmount,
		ShippingFee:     order.ShippingFee,
		PayAmount:       order.PayAmount,
		Status:          order.Status,
		PaymentStatus:   order.PaymentStatus,
		ShipStatus:      order.ShipStatus,
		ReceiverName:    order.ReceiverName,
		ReceiverPhone:   order.ReceiverPhone,
		ReceiverAddress: order.ReceiverAddress,
		ReceiverZipCode: utils.PtrValue(order.ReceiverZipCode),
		Remark:          utils.PtrValue(order.Remark),
		PaidAt:          order.PaidAt.Ptr(),
		ShippedAt:       order.ShippedAt.Ptr(),
		CompletedAt:     order.CompletedAt.Ptr(),
		CancelledAt:     order.CancelledAt.Ptr(),
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

