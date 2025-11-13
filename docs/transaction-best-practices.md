# 优秀项目事务处理实践

## 1. techschool/simplebank (你的参考项目)

**GitHub Stars:** 5k+
**作者:** Tech School (教学项目)

### 架构
```
db/sqlc/
  ├── store.go          # Store 接口
  ├── exec_tx.go        # 事务封装
  ├── tx_transfer.go    # 转账事务
  └── tx_create_user.go # 创建用户事务
```

### 评价
- ✅ **优点:** 简单直接，适合学习
- ❌ **缺点:** 所有事务都在 db 层，不适合大型项目
- 📊 **项目规模:** 小型（教学用途）
- 🎯 **适用场景:** 学习 sqlc、银行转账简单业务

---

## 2. golang-standards/project-layout

**GitHub Stars:** 47k+
**模式:** Clean Architecture

### 架构
```
internal/
  ├── usecase/
  │   └── order.go      # 业务逻辑 + 事务控制
  ├── repository/
  │   └── postgres/
  │       └── order.go  # 数据访问
  └── entity/
      └── order.go      # 领域模型
```

### 事务处理方式
```go
// internal/usecase/order.go
type OrderUseCase struct {
    repo     repository.OrderRepo
    txManager transaction.Manager
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, order *entity.Order) error {
    return uc.txManager.WithTransaction(ctx, func(ctx context.Context) error {
        // 业务逻辑在 UseCase 层
        err := uc.repo.Create(ctx, order)
        err = uc.repo.UpdateStock(ctx, order.Items)
        return err
    })
}
```

### 评价
- ✅ **优点:** 标准的 Clean Architecture，职责清晰
- ✅ **优点:** 事务在 UseCase 层，业务逻辑集中
- ⚠️ **注意:** 需要实现 Transaction Manager
- 🎯 **适用场景:** 中大型项目

---

## 3. go-kratos/kratos (B站开源)

**GitHub Stars:** 23k+
**公司:** Bilibili

### 架构
```
internal/
  ├── biz/
  │   └── order.go      # 业务逻辑
  ├── data/
  │   ├── data.go       # Data 层封装（含事务）
  │   └── order.go      # Repository 实现
  └── service/
      └── order.go      # gRPC/HTTP Service
```

### 事务处理方式
```go
// internal/data/data.go
type Data struct {
    db *gorm.DB
}

type Transaction interface {
    InTx(context.Context, func(ctx context.Context) error) error
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
    return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        ctx = context.WithValue(ctx, txKey{}, tx)
        return fn(ctx)
    })
}

// internal/biz/order.go
type OrderUseCase struct {
    data Transaction
    repo OrderRepo
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context) error {
    return uc.data.InTx(ctx, func(ctx context.Context) error {
        // 业务逻辑
        err := uc.repo.Create(ctx, order)
        err = uc.repo.UpdateStock(ctx, items)
        return err
    })
}
```

### 评价
- ✅ **优点:** 工业级框架，经过大规模验证
- ✅ **优点:** Transaction 抽象为接口，可测试性强
- ⚠️ **注意:** 使用 GORM（性能非最优）
- 🎯 **适用场景:** 微服务架构、大型项目

---

## 4. shopspring/decimal (电商常用库)

**GitHub Stars:** 6k+
**领域:** 电商金额计算

虽然不是完整项目，但提供了一个教训：

### 电商项目的特殊需求
```go
// ❌ 错误：使用 float64
price := 19.99
quantity := 3
total := price * quantity  // 59.97000000000001

// ✅ 正确：使用 decimal
price := decimal.NewFromFloat(19.99)
quantity := decimal.NewFromInt(3)
total := price.Mul(quantity)  // 精确的 59.97
```

### 对你的启示
电商项目的事务中，金额计算必须精确：

```go
// db/sqlc/tx_create_order.go
func (store *SQLStore) CreateOrderTx(ctx context.Context, arg CreateOrderTxParams) error {
    return store.execTx(ctx, func(q *Queries) error {
        // ✅ 使用 int64 存储金额（单位：分）
        totalAmount := int64(0)
        for _, item := range arg.Items {
            totalAmount += item.Price * int64(item.Quantity)
        }

        order, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
            TotalAmount: totalAmount,  // 59.97 元 = 5997 分
        })
        return err
    })
}
```

---

## 5. uber-go/guide (Uber Go 编程指南)

**GitHub Stars:** 15k+
**公司:** Uber

### 事务处理建议

#### ❌ 不推荐：数据层暴露事务细节
```go
// Bad
type UserRepo interface {
    BeginTx() (*sql.Tx, error)
    CreateUser(tx *sql.Tx, user User) error
    CreateProfile(tx *sql.Tx, profile Profile) error
}

func (s *Service) CreateUserWithProfile() error {
    tx, _ := s.repo.BeginTx()
    defer tx.Rollback()

    s.repo.CreateUser(tx, user)
    s.repo.CreateProfile(tx, profile)

    return tx.Commit()
}
```

#### ✅ 推荐：Service 层控制事务
```go
// Good
type UserRepo interface {
    CreateUser(ctx context.Context, user User) error
    CreateProfile(ctx context.Context, profile Profile) error
}

type TxManager interface {
    WithTx(ctx context.Context, fn func(context.Context) error) error
}

func (s *Service) CreateUserWithProfile() error {
    return s.txMgr.WithTx(ctx, func(ctx context.Context) error {
        s.repo.CreateUser(ctx, user)
        s.repo.CreateProfile(ctx, profile)
        return nil
    })
}
```

---

## 6. gitea/gitea (Git 托管平台)

**GitHub Stars:** 44k+
**语言:** Go

### 架构
```
models/
  ├── db/
  │   ├── engine.go     # 数据库引擎
  │   └── context.go    # DB Context（含事务）
  └── user/
      └── user.go       # 领域模型 + 数据访问
```

### 事务处理方式
```go
// models/db/context.go
func WithTx(ctx context.Context, f func(ctx context.Context) error) error {
    e := GetEngine(ctx)
    if e == nil {
        return fmt.Errorf("no db engine in context")
    }

    return e.Transaction(func(sess *xorm.Session) error {
        return f(NewContext(ctx, sess))
    })
}

// modules/user/user.go
func CreateUser(ctx context.Context, u *User) error {
    return db.WithTx(ctx, func(ctx context.Context) error {
        // 插入用户
        _, err := db.GetEngine(ctx).Insert(u)

        // 创建默认仓库
        _, err = repo.Create(ctx, &repo.Repository{
            OwnerID: u.ID,
        })

        return err
    })
}
```

### 评价
- ✅ **优点:** Context 传递事务，简洁优雅
- ✅ **优点:** 支持嵌套事务
- ⚠️ **注意:** 使用 xorm（不是最现代的选择）
- 🎯 **适用场景:** 传统 Web 应用

---

## 7. gorm.io/gorm (最流行的 Go ORM)

**GitHub Stars:** 36k+

### 事务最佳实践
```go
// 1. 自动事务（推荐）
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    if err := tx.Create(&profile).Error; err != nil {
        return err
    }

    return nil
})

// 2. 手动事务（灵活）
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error

// 3. SavePoint（嵌套事务）
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&user1)

    tx.SavePoint("sp1")
    tx.Create(&user2)
    tx.RollbackTo("sp1")  // 只回滚到 sp1

    return nil
})
```

---

## 总结对比

| 项目 | 事务位置 | ORM/SQL | 适用规模 | 推荐度 |
|------|---------|---------|----------|--------|
| **simplebank** | db 层 | sqlc | 小型 | ⭐⭐⭐ (学习) |
| **project-layout** | UseCase 层 | 抽象 | 中大型 | ⭐⭐⭐⭐⭐ |
| **kratos** | Biz 层 | GORM | 大型 | ⭐⭐⭐⭐ |
| **uber-go** | Service 层 | 抽象 | 企业级 | ⭐⭐⭐⭐⭐ |
| **gitea** | Context 层 | xorm | 中型 | ⭐⭐⭐ |

## 对 gomall 的建议 🎯

### 当前阶段（第 1-2 个月）
```
✅ 保持 simplebank 风格
   - 事务在 db/sqlc/ 层
   - 简单直接
   - 快速迭代
```

### 成长阶段（第 3-4 个月）
```
🔄 迁移到 Service 层事务
   - 参考 uber-go guide
   - 添加 Store.ExecTx() 方法
   - 逐步迁移现有事务
```

### 成熟阶段（第 5+ 个月）
```
🚀 引入 Transaction Manager
   - 参考 kratos/gitea
   - Context 传递事务
   - 支持嵌套事务
```

### 架构演进路径

```
阶段 1: simplebank 模式
db/sqlc/
  ├── tx_verify_email.go
  └── tx_reset_password.go

         ↓ (3 个月后)

阶段 2: Service 层控制
internal/user/
  └── service.go
      └── repo.ExecTx(ctx, func(q Querier) error { ... })

         ↓ (6 个月后)

阶段 3: Transaction Manager
pkg/transaction/
  └── manager.go
internal/user/
  └── service.go
      └── txMgr.WithTx(ctx, func(ctx) error { ... })
```

### 核心原则

1. **不要过早优化** - 当前保持简单
2. **监控复杂度** - 当事务文件 > 10 个时重构
3. **渐进式迁移** - 不要一次性重写
4. **保持类型安全** - 坚持使用 sqlc
5. **优先性能** - 电商项目性能至关重要

**我的建议：坚持 sqlc，3 个月后迁移到 Service 层事务！** ✨
