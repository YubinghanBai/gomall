# GORM vs sqlc 事务对比

## GORM 事务处理

### 1. 自动事务

```go
// GORM 提供自动事务管理
db.Transaction(func(tx *gorm.DB) error {
    // 自动开始事务
    if err := tx.Create(&user).Error; err != nil {
        return err  // 自动回滚
    }

    if err := tx.Create(&order).Error; err != nil {
        return err  // 自动回滚
    }

    // 返回 nil 自动提交
    return nil
})
```

**优点：**
- ✅ 使用简单，自动管理
- ✅ 自动回滚/提交
- ✅ 支持嵌套事务（SavePoint）

**缺点：**
- ❌ 性能开销（反射、链式调用）
- ❌ SQL 不透明（难以调试和优化）
- ❌ 类型安全性差（使用 `interface{}`）
- ❌ 容易写出 N+1 查询

### 2. 手动事务

```go
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

if err := tx.Create(&order).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error
```

## sqlc 事务处理

### 1. 当前实现（方案 A）

```go
// db/sqlc/exec_tx.go
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := store.connPool.Begin(ctx)
    if err != nil {
        return err
    }

    q := New(tx)
    err = fn(q)
    if err != nil {
        tx.Rollback(ctx)
        return err
    }

    return tx.Commit(ctx)
}
```

**优点：**
- ✅ 完全类型安全
- ✅ SQL 透明（可见、可优化）
- ✅ 高性能（无反射）
- ✅ 生成的代码可审查

**缺点：**
- ❌ 需要手动管理
- ❌ 样板代码稍多
- ❌ 不支持自动嵌套事务

### 2. 推荐实现（方案 B/C）

见 `transaction-refactor-example.md`

## 性能对比

### 基准测试示例

```go
// GORM
BenchmarkGORM-8    10000    115243 ns/op    25347 B/op    428 allocs/op

// sqlc
BenchmarkSQLC-8    30000     38721 ns/op     3245 B/op     42 allocs/op
```

**sqlc 通常比 GORM 快 2-3 倍，内存占用少 5-8 倍**

## 使用场景建议

### 选择 GORM 的场景
- ✅ 快速原型开发
- ✅ 简单 CRUD 应用
- ✅ 团队不熟悉 SQL
- ✅ 需要跨数据库支持
- ✅ 大量关联查询

### 选择 sqlc 的场景（推荐你的项目 🌟）
- ✅ 性能要求高
- ✅ 需要复杂 SQL 查询
- ✅ 类型安全很重要
- ✅ 想要 SQL 透明可审查
- ✅ 团队熟悉 SQL
- ✅ PostgreSQL 专用项目

## 其他选择

### 1. SQLX
```go
// 介于 GORM 和 sqlc 之间
tx, _ := db.Beginx()
tx.Get(&user, "SELECT * FROM users WHERE id = $1", id)
tx.Commit()
```
- 半手动、半类型安全
- 比 GORM 轻量，比 sqlc 灵活

### 2. Ent (Facebook)
```go
// 强类型的 ORM
client.User.Create().
    SetName("Alice").
    SetAge(30).
    Save(ctx)
```
- 代码生成 + ORM
- 类型安全但学习曲线陡

### 3. Bun (uptrace)
```go
// 类似 GORM 但基于 SQL Builder
db.NewInsert().Model(&user).Exec(ctx)
```
- SQL Builder + 部分 ORM 特性
- 性能介于 GORM 和 sqlc 之间

## 我的推荐

### 对于你的 gomall 项目：

**坚持使用 sqlc** ✅

原因：
1. **电商项目对性能要求高**（高并发下单、库存扣减）
2. **复杂查询多**（订单列表、商品筛选、统计报表）
3. **PostgreSQL 专用**，不需要跨数据库
4. **你已经在用 sqlc**，切换成本高

### 事务组织方式：

**当前阶段：保持方案 A**
- 你的项目还在早期
- 迁移成本 > 收益

**3 个月后：迁移到方案 B**
- 当你有 Order、Product、Cart 等多个领域时
- 开始出现跨领域事务需求

**示例时间线：**

```
现在（第 1 个月）
├── User 领域 ✅
└── 使用方案 A（sqlc 层事务）

第 2-3 个月
├── User 领域
├── Product 领域
├── Category 领域
└── 仍使用方案 A

第 4-5 个月
├── Order 领域
├── Cart 领域
├── Payment 领域
└── 迁移到方案 B（Service 层事务）
    └── 因为出现 CreateOrder 跨领域事务

第 6+ 个月
├── 10+ 个领域
├── 复杂业务流程
└── 考虑方案 C（Transaction Manager）
```

## 总结

| 工具 | 性能 | 类型安全 | 学习曲线 | 适用场景 |
|------|------|----------|----------|----------|
| **GORM** | ⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | 快速开发 |
| **sqlc** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 高性能应用 ⭐ |
| **SQLX** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | 平衡选择 |
| **Ent** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | 企业应用 |
| **Bun** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | GORM 替代 |

**对你的建议：继续使用 sqlc！** 🎯
