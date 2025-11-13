package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gomall/db/sqlc"
	"gomall/internal/domain/inventory"
	"gomall/internal/domain/product"
	"gomall/utils"
)

// Service defines the business logic interface for order domain
type Service interface {
	// Order CRUD operations
	CreateOrder(ctx context.Context, userID int64, req CreateOrderRequest) (*OrderResponse, error)
	GetOrder(ctx context.Context, userID int64, orderID int64) (*OrderResponse, error)
	GetOrderByOrderNo(ctx context.Context, userID int64, orderNo string) (*OrderResponse, error)
	ListUserOrders(ctx context.Context, userID int64, req ListOrdersRequest) (*PaginatedOrdersResponse, error)

	// Order status management
	UpdateOrderStatus(ctx context.Context, userID int64, orderID int64, req UpdateOrderStatusRequest) error
	CancelOrder(ctx context.Context, userID int64, orderID int64) error

	// Payment and shipping
	UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error
	UpdateShipStatus(ctx context.Context, orderID int64, status string) error
	PayOrder(ctx context.Context, userID int64, orderID int64)error
	ShipOrder(ctx context.Context,orderID int64) error
	CompleteOrder(ctx context.Context, orderID int64) error


}

type service struct {
	repo Repository
	inventoryService inventory.Service
	productService product.Service
}

// NewService creates a new Service instance
func NewService(repo Repository, inventoryService inventory.Service, productService product.Service) Service {
	return &service{
		repo: repo,
		inventoryService: inventoryService,
		productService: productService,
	}
}

// CreateOrder creates a new order with items (atomic transaction)
func (s *service) CreateOrder(ctx context.Context, userID int64, req CreateOrderRequest) (*OrderResponse, error) {
	

	maxRetries:=3
	var lastErr error

	for attempt:=0;attempt<maxRetries;attempt++{
		result,err:=s.createOrderWithRetry(ctx,userID,req)
		if err==nil{
			return result, nil
		}

		if strings.Contains(err.Error(), "concurrent update") || 
		   strings.Contains(err.Error(), "version") {
			lastErr = err
			// 
			time.Sleep(time.Millisecond * time.Duration(10*(attempt+1)))
			continue
		}
		
		
		return nil, err
	}

	return nil, fmt.Errorf("failed after %d retries: %w",maxRetries,lastErr)

}

func (s *service) createOrderWithRetry(ctx context.Context, userID int64, req CreateOrderRequest) (*OrderResponse,error){
	
	var result OrderResponse

	//batch get product infomation outside of transaction
	productIDs:=make([]int64,len(req.Items))
	for i,item:=range req.Items{
		productIDs[i]=item.ProductID
	}

	products,err:=s.productService.GetProductsByIDs(ctx,productIDs)
	if err!=nil{
		return nil, fmt.Errorf("failed to get products: %w",err)
	}

	for _, item := range req.Items {
		if _, exists := products[item.ProductID]; !exists {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}
	}

	//validate all products exist
	stockChecks := make([]inventory.StockCheckItem, len(req.Items))
	for i, item := range req.Items {
		stockChecks[i] = inventory.StockCheckItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	
	checkResults, err := s.inventoryService.BatchCheckStockAvailability(ctx, stockChecks)
	if err != nil {
		return nil, fmt.Errorf("failed to check stock availability: %w", err)
	}
	

	//validate all products are sufficient
	for productID, check := range checkResults {
		if !check.IsAvailable {
			return nil, fmt.Errorf("product %d insufficient stock, available: %d, requested: %d", 
				productID, check.AvailableStock, check.RequestedQty)
		}
	}

	err = s.repo.ExecTx(ctx, func(q sqlc.Querier) error {

		// 1. Generate order number
		orderNo := generateOrderNo()

		// 2. Calculate total amount
		
		var totalAmount int64 = 0
		for _, item := range req.Items {
			product:=products[item.ProductID]
			totalAmount+=int64(item.Quantity)*product.Price
		}

		// 3. Calculate pay amount
		payAmount := totalAmount - req.DiscountAmount + req.ShippingFee

		// 4. Create order
		order, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
			OrderNo:         orderNo,
			UserID:          userID,
			TotalAmount:     totalAmount,
			DiscountAmount:  req.DiscountAmount,
			ShippingFee:     req.ShippingFee,
			PayAmount:       payAmount,
			Status:          "pending",
			PaymentStatus:   "unpaid",
			ShipStatus:      "unshipped",
			ReceiverName:    req.ReceiverName,
			ReceiverPhone:   req.ReceiverPhone,
			ReceiverAddress: req.ReceiverAddress,
			ReceiverZipCode: utils.Ptr(req.ReceiverZipCode),
			Remark:          utils.Ptr(req.Remark),
		})
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// 5. Create order items
		items := make([]sqlc.OrderItem, 0, len(req.Items))
		for _, itemReq := range req.Items {
			product:=products[itemReq.ProductID]
			unitPrice:=product.Price
			totalPrice := int64(itemReq.Quantity) * unitPrice

			var productImage *string
			if product.MainImage != "" {
				productImage = &product.MainImage
			}

			item, err := q.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
				OrderID:      order.ID,
				ProductID:    itemReq.ProductID,
				ProductName:  fmt.Sprintf("Product %d", itemReq.ProductID), // placeholder
				ProductImage: productImage,                                          // placeholder
				Quantity:     itemReq.Quantity,
				UnitPrice:    unitPrice,
				TotalPrice:   totalPrice,
			})
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
			items = append(items, item)
		}

		//7. Reserve stock 

		for _,item:=range req.Items{
			err=s.inventoryService.ReserveStock(ctx, inventory.ReserveStockRequest{
				ProductID: item.ProductID,
				Quantity: item.Quantity,
				OrderID: order.ID,
			},30)

			if err!=nil{
				return fmt.Errorf("failed ti reserve stock for product %d: %w",item.ProductID,err)
			}
		}

		// 8. Convert to response
		result = toOrderResponse(order, items)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOrder retrieves an order by ID with all its items
func (s *service) GetOrder(ctx context.Context, userID int64, orderID int64) (*OrderResponse, error) {
	// Get order
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Verify ownership
	if order.UserID != userID {
		return nil, errors.New("unauthorized access to order")
	}

	// Get order items
	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	response := toOrderResponse(order, items)
	return &response, nil
}

// GetOrderByOrderNo retrieves an order by order number
func (s *service) GetOrderByOrderNo(ctx context.Context, userID int64, orderNo string) (*OrderResponse, error) {
	// Get order
	order, err := s.repo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Verify ownership
	if order.UserID != userID {
		return nil, errors.New("unauthorized access to order")
	}

	// Get order items
	items, err := s.repo.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	response := toOrderResponse(order, items)
	return &response, nil
}

// ListUserOrders lists orders for a user with pagination
func (s *service) ListUserOrders(ctx context.Context, userID int64, req ListOrdersRequest) (*PaginatedOrdersResponse, error) {
	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Get orders
	orders, err := s.repo.ListUserOrders(ctx, sqlc.ListUserOrdersParams{
		UserID: userID,
		Limit:  req.PageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	// Get total count
	total, err := s.repo.CountUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get order IDs for batch loading items
	orderIDs := make([]int64, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.ID
	}

	// Batch load order items
	itemsMap := make(map[int64][]sqlc.OrderItem)
	if len(orderIDs) > 0 {
		allItems, err := s.repo.GetOrderItemsByIDs(ctx, orderIDs)
		if err == nil {
			for _, item := range allItems {
				itemsMap[item.OrderID] = append(itemsMap[item.OrderID], item)
			}
		}
	}

	// Convert to response
	orderResponses := make([]OrderResponse, len(orders))
	for i, o := range orders {
		orderResponses[i] = toOrderResponse(o, itemsMap[o.ID])
	}

	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &PaginatedOrdersResponse{
		Orders:     orderResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateOrderStatus updates the order status
func (s *service) UpdateOrderStatus(ctx context.Context, userID int64, orderID int64, req UpdateOrderStatusRequest) error {
	// Verify order ownership
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return errors.New("unauthorized access to order")
	}

	// Update status
	err = s.repo.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		Status: req.Status,
		ID:     orderID,
	})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// CancelOrder cancels an order
func (s *service) CancelOrder(ctx context.Context, userID int64, orderID int64) error {
	return s.repo.ExecTx(ctx,func(q sqlc.Querier) error {

		// Verify order ownership
		order, err := s.repo.GetOrderByID(ctx, orderID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("order not found")
			}
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.UserID != userID {
			return errors.New("unauthorized access to order")
		}

		// Check if order can be cancelled
		if order.Status == "completed" || order.Status == "cancelled" {
			return errors.New("order cannot be cancelled")
		}


		// Get order items
		items,err:=q.GetOrderItems(ctx,orderID)
		if err!=nil{
			return fmt.Errorf("failed to get order items: %w",err)
		}

		//Release reserved stock(only if order is pending/unpaid)
		if order.Status=="pending" &&order.PaymentStatus=="unpaid"{
			for _,item:=range items{
				err=s.inventoryService.ReleaseStock(ctx, inventory.ReleaseStockRequest{
					ProductID: item.ProductID,
					Quantity: item.Quantity,
					OrderID: orderID,
				})
				if err!=nil{
					fmt.Printf("failed to release stock for product %d: %v\n",item.ProductID,err)
				}
			}
		}

		// Cancel order
		err = s.repo.CancelOrder(ctx, orderID)
		if err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}

		return nil
	})
}

// UpdatePaymentStatus updates the payment status (usually called by payment service)
func (s *service) UpdatePaymentStatus(ctx context.Context, orderID int64, status string) error {
	err := s.repo.UpdateOrderPaymentStatus(ctx, sqlc.UpdateOrderPaymentStatusParams{
		PaymentStatus: status,
		ID:            orderID,
	})
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// If payment is successful, update order status to paid
	if status == "paid" {
		err = s.repo.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
			Status: "paid",
			ID:     orderID,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	return nil
}

// UpdateShipStatus updates the shipping status (usually called by shipping service)
func (s *service) UpdateShipStatus(ctx context.Context, orderID int64, status string) error {
	err := s.repo.UpdateOrderShipStatus(ctx, sqlc.UpdateOrderShipStatusParams{
		ShipStatus: status,
		ID:         orderID,
	})
	if err != nil {
		return fmt.Errorf("failed to update ship status: %w", err)
	}

	// If shipped, update order status
	if status == "shipped" {
		err = s.repo.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
			Status: "shipped",
			ID:     orderID,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	return nil
}

//PayOrder handles order payment (deduct reserved stock)
func (s *service) PayOrder(ctx context.Context, userID int64, orderID int64) error{
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1.Verify order ownership
		order,err:=q.GetOrderByID(ctx,orderID)
		if err!=nil{
			if errors.Is(err,sql.ErrNoRows){
				return errors.New("order not found")
			}
			return fmt.Errorf("failed to get order: %w",err)
		}

		if order.UserID!=userID{
			return errors.New("unauthorized access to order")
		}

		// 2. check ordder Status
		if order.Status!="pending"{
			return fmt.Errorf("order status is %s, cannot pay",order.Status)
		}

		// 3.Get order items
		items,err:=q.GetOrderItems(ctx,orderID)
		if err!=nil{
			return fmt.Errorf("failed to get order items: %w",err)
		}
		//TODO: ADD payment module
		//4. Deduct reserved stock for each item
		for _,item:=range items{
			err=s.inventoryService.DeductStock(ctx,inventory.DeductStockRequest{
				ProductID: item.ProductID,
				Quantity: item.Quantity,
				OrderID: item.OrderID,
			})
			if err!=nil{
				return fmt.Errorf("failed to deduct stock for product %d: %w",item.ProductID,err)
			}
		}

		// 5. Update order status
		err=q.UpdateOrderStatus(ctx,sqlc.UpdateOrderStatusParams{
			Status: "paid",
			ID: orderID,
		})
		if err!=nil{
			return fmt.Errorf("failed to update payment status: %w",err)
		}

		return nil
	})
}

// ShipOrder marks order as shipped

func (s *service) ShipOrder(ctx context.Context,orderID int64) error{
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get order
		order,err:=q.GetOrderByID(ctx,orderID)
		if err!=nil{
			return fmt.Errorf("failed to get order: %w",err)
		}

		// 2. Check if order is paid 
		if order.Status !="paid"{
			return fmt.Errorf("order status is %s, cannot ship",order.Status)
		}
		// TODO: 但是在实际中往往是会过几个小时才会发货那这边这个逻辑怎么办 使用异步队列吗还是
		// 3. Update order status
		err=q.UpdateOrderStatus(ctx,sqlc.UpdateOrderStatusParams{
			Status:"shipped",
			ID:orderID,
		})
		if err!=nil{
			return fmt.Errorf("failed to update ship status: %w",err)
		}
		return nil
	})
}

func (s *service) CompleteOrder(ctx context.Context, orderID int64) error{
	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 1. Get order
		order, err := q.GetOrderByID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		// 2. Check if order is shipped
		if order.Status != "shipped" {
			return fmt.Errorf("order status is %s, cannot complete", order.Status)
		}

		// 3. Update order status
		err = q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
			Status: "completed",
			ID:     orderID,
		})
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		return nil
	})
}




func generateOrderNo() string {
	// Generate order number with timestamp
	// Format: ORD + YYYYMMDDHHMMSS + random suffix
	now := time.Now()
	return fmt.Sprintf("ORD%s%03d", now.Format("20060102150405"), now.Nanosecond()%1000)
}
