# 从单体到微服务的演进策略

## 一、当前架构（单体应用）

```
gomall (单体应用)
├── internal/
│   ├── user/
│   │   ├── service.go      # 业务逻辑
│   │   └── repository.go   # = sqlc.Store
│   ├── order/
│   │   ├── service.go
│   │   └── repository.go   # = sqlc.Store
│   └── product/
│       ├── service.go
│       └── repository.go   # = sqlc.Store
└── db/sqlc/
    ├── store.go            # Store 接口（单一数据库）
    └── *.sql.go            # 所有表的查询方法
```

**特点：**
- ✅ 单一数据库（PostgreSQL）
- ✅ 本地事务（ACID 保证）
- ✅ 直接函数调用（无网络开销）

---

## 二、微服务拆分策略

### 阶段 1: 模块化单体（现在 → 6 个月）

**目标：** 为微服务做准备，但仍是单体应用

```
gomall (模块化单体)
├── internal/
│   ├── user/          # User 领域（未来的 User 服务）
│   │   ├── service.go
│   │   └── repository.go
│   ├── order/         # Order 领域（未来的 Order 服务）
│   │   ├── service.go
│   │   └── repository.go
│   └── product/       # Product 领域（未来的 Product 服务）
│       ├── service.go
│       └── repository.go
└── db/
    └── sqlc/
        └── store.go   # 仍然是单一 Store
```

**关键原则：**
1. **领域边界清晰** - 每个领域只访问自己的表
2. **避免跨领域直接查询** - 通过 Service 层调用
3. **事件化思维** - 准备引入事件总线

**示例：**

```go
// ❌ 错误：Order Service 直接访问 User 表
func (s *orderService) CreateOrder(ctx context.Context) error {
    user, err := s.store.GetUserByID(ctx, userID)  // 跨领域查询
}

// ✅ 正确：Order Service 调用 User Service
func (s *orderService) CreateOrder(ctx context.Context) error {
    user, err := s.userService.GetProfile(ctx, userID)  // 通过 Service
}
```

---

### 阶段 2: 数据库拆分（6-12 个月）

**目标：** 每个领域有独立数据库（仍在单体应用中）

```
gomall (单体应用 + 多数据库)
├── internal/
│   ├── user/
│   │   ├── service.go
│   │   └── repository.go   # → user_db.Store
│   ├── order/
│   │   ├── service.go
│   │   └── repository.go   # → order_db.Store
│   └── product/
│       ├── service.go
│       └── repository.go   # → product_db.Store
└── db/
    ├── user_db/sqlc/       # User 数据库
    ├── order_db/sqlc/      # Order 数据库
    └── product_db/sqlc/    # Product 数据库
```

**重构步骤：**

1. **按领域拆分 SQL 文件**

```bash
# 之前：所有表在一起
db/queries/
  ├── user.sql
  ├── order.sql
  └── product.sql

# 之后：按领域分离
db/user_db/queries/user.sql
db/order_db/queries/order.sql
db/product_db/queries/product.sql
```

2. **生成独立的 Store**

```yaml
# db/user_db/sqlc.yaml
version: "2"
sql:
  - schema: "db/user_db/schema.sql"
    queries: "db/user_db/queries"
    engine: "postgresql"
    gen:
      go:
        package: "user_sqlc"
        out: "db/user_db/sqlc"
```

```go
// internal/user/repository.go
type Repository interface {
    user_sqlc.Store  // 使用 User 数据库的 Store
}

// internal/order/repository.go
type Repository interface {
    order_sqlc.Store  // 使用 Order 数据库的 Store
}
```

3. **跨领域调用通过 Service**

```go
// internal/order/service.go
type service struct {
    orderRepo   Repository      // Order 数据库
    userService user.Service    // User Service（跨领域调用）
    productService product.Service
}

func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    // 1. 验证用户（跨领域调用 User Service）
    user, err := s.userService.GetProfile(ctx, req.UserID)

    // 2. 验证商品（跨领域调用 Product Service）
    for _, item := range req.Items {
        product, err := s.productService.GetProduct(ctx, item.ProductID)
    }

    // 3. 创建订单（本领域事务）
    return s.orderRepo.ExecTx(ctx, func(q order_sqlc.Querier) error {
        order, err := q.CreateOrder(ctx, ...)
        return err
    })
}
```

**问题：跨领域事务怎么办？**

---

## 三、跨服务事务解决方案

### 方案 1: Saga 模式（推荐 🌟）

**原理：** 将跨服务事务拆分为多个本地事务 + 补偿操作

```go
// internal/order/service.go
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    var orderID int64

    // 步骤 1: 扣减库存（Product 服务）
    err := s.productService.DeductStock(ctx, req.Items)
    if err != nil {
        return err
    }

    // 步骤 2: 创建订单（Order 服务）
    order, err := s.orderRepo.CreateOrder(ctx, req)
    if err != nil {
        // 补偿：恢复库存
        s.productService.RestoreStock(ctx, req.Items)
        return err
    }
    orderID = order.ID

    // 步骤 3: 清空购物车（Cart 服务）
    err = s.cartService.ClearCart(ctx, req.UserID)
    if err != nil {
        // 补偿：取消订单 + 恢复库存
        s.orderRepo.CancelOrder(ctx, orderID)
        s.productService.RestoreStock(ctx, req.Items)
        return err
    }

    return nil
}
```

**优点：**
- ✅ 无需分布式事务协调器
- ✅ 性能好（本地事务）
- ✅ 易于理解和调试

**缺点：**
- ❌ 需要实现补偿逻辑
- ❌ 非原子性（中间状态可见）

---

### 方案 2: 事件驱动（最终一致性）

**原理：** 通过事件总线（Kafka/NATS）实现最终一致性

```go
// internal/order/service.go
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    // 1. 创建订单（本地事务）
    order, err := s.orderRepo.ExecTx(ctx, func(q order_sqlc.Querier) error {
        order, err := q.CreateOrder(ctx, ...)

        // 2. 发布事件（在同一事务中）
        event := &OrderCreatedEvent{
            OrderID: order.ID,
            UserID:  req.UserID,
            Items:   req.Items,
        }
        err = q.SaveOutboxEvent(ctx, event)  // Outbox 模式

        return err
    })

    // 3. 异步发送事件到消息队列
    s.eventBus.Publish("order.created", event)

    return nil
}

// Product 服务监听事件
func (s *productService) OnOrderCreated(event *OrderCreatedEvent) {
    // 扣减库存
    s.productRepo.ExecTx(ctx, func(q product_sqlc.Querier) error {
        for _, item := range event.Items {
            q.DeductStock(ctx, item.ProductID, item.Quantity)
        }
        return nil
    })
}
```

**优点：**
- ✅ 服务完全解耦
- ✅ 高可用（消息队列）
- ✅ 易于扩展

**缺点：**
- ❌ 最终一致性（有延迟）
- ❌ 需要处理重复消息
- ❌ 调试复杂

---

### 方案 3: 两阶段提交（不推荐）

**原理：** 使用分布式事务协调器（如 Seata）

```go
// 需要分布式事务管理器
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    return s.dtm.TransactionDo(ctx, func(ctx context.Context) error {
        // 所有操作在分布式事务中
        s.productService.DeductStock(ctx, req.Items)
        s.orderRepo.CreateOrder(ctx, req)
        s.cartService.ClearCart(ctx, req.UserID)
        return nil
    })
}
```

**缺点：**
- ❌ 性能差（阻塞锁）
- ❌ 复杂度高
- ❌ 可用性低（协调器单点）

---

## 四、阶段 3: 完全微服务（12+ 个月）

```
用户服务 (user-service)
├── cmd/server/
├── internal/user/
└── db/user_db/sqlc/

订单服务 (order-service)
├── cmd/server/
├── internal/order/
└── db/order_db/sqlc/

商品服务 (product-service)
├── cmd/server/
├── internal/product/
└── db/product_db/sqlc/

API 网关 (gateway)
└── 路由到各个服务
```

**通信方式：**

1. **同步调用：gRPC**

```protobuf
// api/user/v1/user.proto
service UserService {
    rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
}
```

```go
// internal/order/service.go
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    // 通过 gRPC 调用 User 服务
    user, err := s.userClient.GetProfile(ctx, &userpb.GetProfileRequest{
        UserId: req.UserID,
    })
}
```

2. **异步通信：事件总线**

```go
// 发布事件
s.eventBus.Publish("order.created", event)

// 订阅事件
s.eventBus.Subscribe("order.created", s.OnOrderCreated)
```

---

## 五、Store 拆分总结

### 单体应用（现在）

```go
// 一个 Store，所有领域共享
type Store interface {
    Querier  // 所有表的查询方法
    ExecTx(...)
}
```

### 模块化单体（6 个月）

```go
// 仍然一个 Store，但领域边界清晰
// User Service 只调用 User 相关方法
// Order Service 只调用 Order 相关方法
```

### 数据库拆分（12 个月）

```go
// 每个领域独立的 Store
type UserStore interface {
    UserQuerier  // 只有 User 表的方法
    ExecTx(...)
}

type OrderStore interface {
    OrderQuerier  // 只有 Order 表的方法
    ExecTx(...)
}
```

### 完全微服务（18+ 个月）

```go
// 每个服务独立部署，独立数据库
// 通过 gRPC/REST/事件总线通信
// 跨服务事务使用 Saga/事件驱动
```

---

## 六、我的建议

### 当前阶段（0-6 个月）
- ✅ 保持单体应用 + 单一 Store
- ✅ 专注业务功能开发
- ✅ 建立清晰的领域边界

### 中期（6-12 个月）
- 🔄 开始数据库拆分
- 🔄 引入事件总线（NATS/Kafka）
- 🔄 实现 Saga 模式

### 长期（12+ 个月）
- 🚀 微服务拆分
- 🚀 gRPC 通信
- 🚀 Kubernetes 部署

**核心原则：不要过早微服务化！**

单体应用可以支撑到：
- 10 万+ DAU
- 100+ 张表
- 10+ 个开发人员

在此之前，保持单体更高效！