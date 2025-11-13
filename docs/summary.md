# 事务重构与 Mock 测试总结

## 一、你的理解确认 ✅

### 1. 事务实现的本质

**完全正确！** 现在的实现：

```go
// 之前（sqlc 层事务）：需要为每个事务定义参数和结果类型
type VerifyEmailTxParams struct {
    CodeID int64
    UserID int64
}
type VerifyEmailTxResult struct {
    User User
}

// 现在（Service 层事务）：直接在闭包中组装，复用现有方法
s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
    q.MarkCodeAsUsed(ctx, codeID)      // 复用
    q.VerifyUserEmail(ctx, userID)     // 复用
    return nil
})
```

**优势对比：**

| 特性 | 之前（sqlc 层） | 现在（Service 层） |
|------|----------------|-------------------|
| 参数类型 | 需要新定义 TxParams | ✅ 复用现有参数 |
| 代码位置 | db/sqlc/tx_*.go | ✅ Service 层（业务逻辑集中） |
| 跨领域事务 | ❌ 不知道放哪 | ✅ 很自然 |
| 灵活性 | ❌ 固定的事务模板 | ✅ 灵活组装 |
| 可维护性 | ❌ 事务散落各处 | ✅ 业务逻辑清晰 |

---

## 二、微服务拆分策略

### 演进路径（推荐）

```
阶段 1: 单体应用（现在 → 6 个月）
├── 单一数据库（PostgreSQL）
├── 单一 Store 接口
├── 本地事务（ACID 保证）
└── 专注业务功能开发

         ↓

阶段 2: 模块化单体（6-12 个月）
├── 仍是单体应用
├── 领域边界清晰
├── 准备数据库拆分
└── 引入事件总线

         ↓

阶段 3: 数据库拆分（12-18 个月）
├── 每个领域独立数据库
├── 多个 Store（user_db.Store, order_db.Store）
├── 跨领域调用通过 Service
└── Saga 模式处理跨领域事务

         ↓

阶段 4: 完全微服务（18+ 个月）
├── 每个服务独立部署
├── gRPC/REST 通信
├── 事件驱动架构
└── Kubernetes 编排
```

### Store 拆分策略

**当前（单体应用）：**
```go
// 一个 Store，所有领域共享
type Store interface {
    Querier  // 所有表的查询方法
    ExecTx(ctx, fn func(Querier) error) error
}

// 所有 Repository 都使用同一个 Store
type UserRepository interface { Store }
type OrderRepository interface { Store }
```

**数据库拆分后：**
```go
// 每个领域独立的 Store
// db/user_db/sqlc/store.go
type UserStore interface {
    UserQuerier  // 只有 User 表
    ExecTx(...)
}

// db/order_db/sqlc/store.go
type OrderStore interface {
    OrderQuerier  // 只有 Order 表
    ExecTx(...)
}

// internal/user/repository.go
type Repository interface {
    user_sqlc.UserStore  // 使用 User 数据库
}
```

### 跨服务事务解决方案

#### 1. Saga 模式（推荐 🌟）

```go
func (s *orderService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    // 步骤 1: 扣减库存
    err := s.productService.DeductStock(ctx, req.Items)
    if err != nil {
        return err
    }

    // 步骤 2: 创建订单
    order, err := s.orderRepo.CreateOrder(ctx, req)
    if err != nil {
        // 补偿：恢复库存
        s.productService.RestoreStock(ctx, req.Items)
        return err
    }

    // 步骤 3: 清空购物车
    err = s.cartService.ClearCart(ctx, req.UserID)
    if err != nil {
        // 补偿：取消订单 + 恢复库存
        s.orderRepo.CancelOrder(ctx, order.ID)
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

#### 2. 事件驱动（最终一致性）

```go
// Order 服务发布事件
func (s *orderService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    order, err := s.orderRepo.CreateOrder(ctx, req)

    // 发布事件
    s.eventBus.Publish("order.created", &OrderCreatedEvent{
        OrderID: order.ID,
        Items:   req.Items,
    })

    return nil
}

// Product 服务监听事件
func (s *productService) OnOrderCreated(event *OrderCreatedEvent) {
    s.productRepo.DeductStock(ctx, event.Items)
}
```

**优点：**
- ✅ 服务完全解耦
- ✅ 高可用
- ✅ 易于扩展

**缺点：**
- ❌ 最终一致性（有延迟）
- ❌ 需要处理重复消息

### 我的建议

**当前阶段（0-6 个月）：**
- ✅ 保持单体应用
- ✅ 单一 Store
- ✅ 专注业务开发

**何时考虑微服务：**
- 团队规模 > 10 人
- DAU > 10 万
- 领域 > 10 个
- 需要独立扩展某些服务

**核心原则：不要过早微服务化！**

---

## 三、Mock 测试实现 ✅

### 1. Makefile 配置

```makefile
mock:
	@echo "Generating mocks..."
	@mockgen -package mockdb -destination db/mock/store.go gomall/db/sqlc Store
	@echo "✅ Store mock generated: db/mock/store.go"
```

### 2. 生成 Mock

```bash
# 生成 Store 接口的 mock
make mock

# 生成的文件
db/mock/store.go
```

### 3. 测试示例

**基础查询测试：**

```go
func TestGetProfile(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStore := mockdb.NewMockStore(ctrl)

    // 设置期望
    mockStore.EXPECT().
        GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
        Return(user, nil)

    // 执行测试
    service := NewService(nil, mockStore, nil, nil)
    user, err := service.GetProfile(ctx, 1)

    // 断言
    require.NoError(t, err)
}
```

**事务测试：**

```go
func TestVerifyEmail(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStore := mockdb.NewMockStore(ctrl)

    // 期望：获取验证码
    mockStore.EXPECT().
        GetVerificationCode(gomock.Any(), gomock.Any()).
        Return(code, nil)

    // 期望：执行事务
    mockStore.EXPECT().
        ExecTx(gomock.Any(), gomock.Any()).
        DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
            // 模拟事务成功
            return nil
        })

    service := NewService(nil, mockStore, nil, nil)
    err := service.VerifyEmail(ctx, req)

    require.NoError(t, err)
}
```

### 4. 运行测试

```bash
# 运行所有测试
make test

# 运行特定测试
go test -v ./internal/user/... -run TestVerifyEmail

# 生成覆盖率
make test-coverage
```

### 5. 测试结果 ✅

```
=== RUN   TestGetProfile
--- PASS: TestGetProfile (0.00s)
=== RUN   TestGetProfile_NotFound
--- PASS: TestGetProfile_NotFound (0.00s)
=== RUN   TestVerifyEmail_Success
--- PASS: TestVerifyEmail_Success (0.00s)
=== RUN   TestVerifyEmail_InvalidCode
--- PASS: TestVerifyEmail_InvalidCode (0.00s)
=== RUN   TestVerifyEmail_Expired
--- PASS: TestVerifyEmail_Expired (0.00s)
=== RUN   TestResetPassword_Success
--- PASS: TestResetPassword_Success (0.06s)
=== RUN   TestResetPassword_TransactionFailed
--- PASS: TestResetPassword_TransactionFailed (0.05s)
PASS
ok  	gomall/internal/user	0.287s
```

---

## 四、关键收获

### 1. 事务架构演进

```
❌ 之前：业务事务在 sqlc 层
├── db/sqlc/tx_verify_email.go (100+ 行)
├── db/sqlc/tx_reset_password.go (150+ 行)
└── 问题：跨领域事务不知道放哪

✅ 现在：业务逻辑在 Service 层
├── db/sqlc/store.go (只有 Querier + ExecTx)
└── internal/user/service.go (业务逻辑清晰)
```

### 2. 核心优势

1. **更灵活** - 不需要为每个事务定义参数类型
2. **更清晰** - 业务逻辑在 Service 层，不藏在 db 层
3. **可扩展** - 跨领域事务很自然
4. **可测试** - Mock Store 接口即可

### 3. 微服务演进

```
单体应用（现在）
  ↓ 6 个月后
模块化单体
  ↓ 12 个月后
数据库拆分
  ↓ 18 个月后
完全微服务
```

**不要过早微服务化！单体应用可以支撑到：**
- 10 万+ DAU
- 100+ 张表
- 10+ 个开发人员

### 4. 测试覆盖

- ✅ 基础查询测试
- ✅ 事务成功场景
- ✅ 事务失败场景
- ✅ 验证码过期场景
- ✅ 数据不存在场景

---

## 五、下一步计划

### 短期（本周）
- ✅ 事务重构完成
- ✅ Mock 测试完成
- 🔨 继续开发 Product、Order 领域

### 中期（1-3 个月）
- 🔨 实现订单功能
- 🔨 实现支付功能
- 📊 监控事务复杂度

### 长期（3-6 个月）
- 🔄 考虑数据库拆分（如果需要）
- 🔄 引入事件总线（NATS/Kafka）
- 🚀 考虑微服务拆分（如果业务需要）

---

## 六、相关文档

- 📖 `docs/transaction-refactor-example.md` - 事务重构详细示例
- 📖 `docs/gorm-vs-sqlc-transactions.md` - GORM vs sqlc 对比
- 📖 `docs/transaction-best-practices.md` - 优秀项目实践
- 📖 `docs/microservices-migration.md` - 微服务演进策略
- 📖 `docs/mock-testing-guide.md` - Mock 测试指南

---

**总结：你的理解完全正确！现在的架构清晰、灵活、可扩展，专注业务开发即可！** 🎉