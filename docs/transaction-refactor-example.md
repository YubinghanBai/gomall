# 事务重构方案示例

## 方案 B：Service 层控制事务（推荐）

### 1. 修改 Store 接口，暴露通用事务方法

```go
// db/sqlc/store.go
type Store interface {
    Querier
    // 通用事务执行方法（给 Service 层使用）
    ExecTx(ctx context.Context, fn func(Querier) error) error

    // 可选：保留一些复杂的事务封装
    // VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

func (store *SQLStore) ExecTx(ctx context.Context, fn func(Querier) error) error {
    tx, err := store.connPool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    q := New(tx)
    err = fn(q)  // 注意这里传的是 Querier 接口
    if err != nil {
        if rbErr := tx.Rollback(ctx); rbErr != nil {
            return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
        }
        return err
    }

    return tx.Commit(ctx)
}
```

### 2. Service 层使用事务

```go
// internal/user/service.go

func (s *service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
    // 1. 查找验证码（不在事务中）
    verificationCode, err := s.repo.GetVerificationCode(ctx, sqlc.GetVerificationCodeParams{
        Email: req.Email,
        Code:  req.Code,
        Type:  "email_verification",
    })
    if err != nil {
        return fmt.Errorf("failed to get verification code: %w", err)
    }

    // 2. 检查是否过期（不在事务中）
    if time.Now().After(verificationCode.ExpiresAt) {
        return errors.New("verification code has expired")
    }

    // 3. 在事务中执行（Service 层控制）
    return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
        // 标记验证码已使用
        if err := q.MarkCodeAsUsed(ctx, verificationCode.ID); err != nil {
            return fmt.Errorf("failed to mark code as used: %w", err)
        }

        // 验证用户邮箱
        if err := q.VerifyUserEmail(ctx, verificationCode.UserID); err != nil {
            return fmt.Errorf("failed to verify user email: %w", err)
        }

        return nil
    })
}
```

### 3. 跨领域事务示例

```go
// internal/order/service.go

type OrderService struct {
    store        sqlc.Store  // 注意：依赖 Store 而不是特定的 Repository
    productRepo  product.Repository
    cartRepo     cart.Repository
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []OrderItem) error {
    // Service 层控制整个事务
    return s.store.ExecTx(ctx, func(q sqlc.Querier) error {
        var totalAmount int64

        // 1. 验证并扣减库存
        for _, item := range items {
            product, err := q.GetProduct(ctx, item.ProductID)
            if err != nil {
                return fmt.Errorf("product not found: %w", err)
            }

            if product.Stock < item.Quantity {
                return fmt.Errorf("insufficient stock for product %s", product.Name)
            }

            // 扣减库存
            err = q.UpdateProductStock(ctx, sqlc.UpdateProductStockParams{
                ID:       item.ProductID,
                Quantity: -item.Quantity,
            })
            if err != nil {
                return fmt.Errorf("failed to deduct stock: %w", err)
            }

            totalAmount += product.Price * int64(item.Quantity)
        }

        // 2. 创建订单
        order, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
            UserID:      userID,
            TotalAmount: totalAmount,
            Status:      "pending",
        })
        if err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }

        // 3. 创建订单项
        for _, item := range items {
            _, err := q.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
                OrderID:   order.ID,
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
            })
            if err != nil {
                return fmt.Errorf("failed to create order item: %w", err)
            }
        }

        // 4. 清空购物车
        err = q.ClearCart(ctx, userID)
        if err != nil {
            return fmt.Errorf("failed to clear cart: %w", err)
        }

        return nil
    })
}
```

## 方案 C：独立的 Transaction Manager（企业级）

### 1. 创建事务管理器

```go
// pkg/transaction/manager.go
package transaction

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Manager interface {
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type manager struct {
    pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) Manager {
    return &manager{pool: pool}
}

// txKey 用于在 context 中存储事务
type txKey struct{}

func (m *manager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
    // 检查是否已在事务中
    if tx := GetTx(ctx); tx != nil {
        // 已在事务中，直接执行（支持嵌套）
        return fn(ctx)
    }

    // 开始新事务
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return err
    }

    // 将事务存入 context
    ctx = context.WithValue(ctx, txKey{}, tx)

    // 执行业务逻辑
    err = fn(ctx)
    if err != nil {
        tx.Rollback(ctx)
        return err
    }

    return tx.Commit(ctx)
}

// GetTx 从 context 中获取事务
func GetTx(ctx context.Context) pgx.Tx {
    if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
        return tx
    }
    return nil
}
```

### 2. Repository 使用事务

```go
// internal/user/repository.go
type repository struct {
    pool *pgxpool.Pool
}

func (r *repository) getDB(ctx context.Context) sqlc.DBTX {
    // 优先使用 context 中的事务
    if tx := transaction.GetTx(ctx); tx != nil {
        return tx
    }
    // 否则使用连接池
    return r.pool
}

func (r *repository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
    q := sqlc.New(r.getDB(ctx))  // 自动选择事务或连接池
    return q.CreateUser(ctx, arg)
}
```

### 3. Service 层使用

```go
// internal/user/service.go
type service struct {
    repo    Repository
    txMgr   transaction.Manager
}

func (s *service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
    return s.txMgr.WithTransaction(ctx, func(ctx context.Context) error {
        // 所有 repo 操作都会自动使用同一个事务
        err := s.repo.MarkCodeAsUsed(ctx, codeID)
        err = s.repo.VerifyUserEmail(ctx, userID)
        return err
    })
}

// 跨领域事务
func (s *orderService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    return s.txMgr.WithTransaction(ctx, func(ctx context.Context) error {
        // 所有 repo 调用自动在同一事务中
        err := s.productRepo.DeductStock(ctx, ...)
        err = s.orderRepo.CreateOrder(ctx, ...)
        err = s.cartRepo.ClearCart(ctx, ...)
        return err
    })
}
```

**优点：**
- ✅ 最灵活，支持嵌套事务
- ✅ 业务代码干净，不关心事务细节
- ✅ 跨领域事务非常容易
- ❌ 实现复杂度高
- ❌ 对团队要求高

## 对比总结

| 方案 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| **A: sqlc 层** | 小项目、单领域 | 简单直接 | 业务逻辑泄露、不可扩展 |
| **B: Service 层** | 中型项目 ⭐ | 清晰、灵活 | 需要暴露 ExecTx |
| **C: Transaction Manager** | 大型项目 | 最灵活 | 实现复杂 |

## 推荐路径

### 当前阶段（1-3 个月）
保持 **方案 A（当前实现）**，原因：
- 项目还在早期
- 只有 User 领域
- 简单快速迭代

### 中期（3-6 个月）
迁移到 **方案 B（Service 层事务）**，当：
- 有 3+ 个领域
- 出现跨领域事务需求
- 团队规模 2-5 人

### 长期（6+ 个月）
考虑 **方案 C（Transaction Manager）**，当：
- 有 10+ 个领域
- 复杂的嵌套事务
- 团队规模 5+ 人

## 重构步骤（从 A → B）

1. **修改 Store 接口**：添加 `ExecTx(ctx, func(Querier) error) error`
2. **保留现有事务方法**：`VerifyEmailTx` 等作为兼容层
3. **逐步迁移 Service**：新功能用 ExecTx，旧功能保持不变
4. **最终清理**：删除 `tx_*.go` 文件
