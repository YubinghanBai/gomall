package cache

import "fmt"

// Keys provides centralized cache key generation
type Keys struct{}

var CacheKeys = Keys{}

// Product keys
func (k Keys) Product(id int64) string {
	return fmt.Sprintf("product:%d", id)
}

func (k Keys) ProductList(page, pageSize int) string {
	return fmt.Sprintf("product:list:%d:%d", page, pageSize)
}

func (k Keys) ProductsByCategory(categoryID int64, page, pageSize int) string {
	return fmt.Sprintf("product:category:%d:%d:%d", categoryID, page, pageSize)
}

func (k Keys) HotProducts(limit int) string {
	return fmt.Sprintf("product:hot:%d", limit)
}

// Inventory/Stock keys
func (k Keys) Stock(productID int64) string {
	return fmt.Sprintf("stock:%d", productID)
}

func (k Keys) StockBatch(productIDs []int64) []string {
	keys := make([]string, len(productIDs))
	for i, id := range productIDs {
		keys[i] = k.Stock(id)
	}
	return keys
}

func (k Keys) LowStockProducts(page, pageSize int) string {
	return fmt.Sprintf("stock:low:%d:%d", page, pageSize)
}

// User keys
func (k Keys) User(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

func (k Keys) UserSession(token string) string {
	return fmt.Sprintf("session:%s", token)
}

func (k Keys) UserCart(userID int64) string {
	return fmt.Sprintf("cart:user:%d", userID)
}

// Category keys
func (k Keys) Category(id int64) string {
	return fmt.Sprintf("category:%d", id)
}

func (k Keys) CategoryList() string {
	return "category:list"
}

// Order keys
func (k Keys) Order(id int64) string {
	return fmt.Sprintf("order:%d", id)
}

func (k Keys) UserOrders(userID int64, page, pageSize int) string {
	return fmt.Sprintf("order:user:%d:%d:%d", userID, page, pageSize)
}

func (k Keys) OrderStatus(orderID int64) string {
	return fmt.Sprintf("order:%d:status", orderID)
}

// Rate limiting keys
func (k Keys) RateLimit(identifier string) string {
	return fmt.Sprintf("ratelimit:%s", identifier)
}

func (k Keys) APIRateLimit(userID int64, endpoint string) string {
	return fmt.Sprintf("ratelimit:api:%d:%s", userID, endpoint)
}

// Lock keys (for distributed locks)
func (k Keys) Lock(resource string) string {
	return fmt.Sprintf("lock:%s", resource)
}

func (k Keys) StockLock(productID int64) string {
	return fmt.Sprintf("lock:stock:%d", productID)
}

// Counter keys
func (k Keys) ProductViewCount(productID int64) string {
	return fmt.Sprintf("counter:product:view:%d", productID)
}

func (k Keys) DailyOrderCount(date string) string {
	return fmt.Sprintf("counter:order:daily:%s", date)
}
