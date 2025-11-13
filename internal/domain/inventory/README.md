# Inventory Domain (库存领域)

## 概述

Inventory 领域负责管理电商系统中的库存，这是防止超卖、确保订单履约的核心模块。

## 已实现的核心特性

### 1. 库存管理 (Inventory Management)
- ✅ 创建和查询库存
- ✅ 可用库存和预留库存分离管理
- ✅ 低库存阈值告警
- ✅ 库存列表查询（支持分页）
- ✅ 低库存商品查询

### 2. 防止超卖 (Overselling Prevention)
- ✅ **乐观锁机制** (Optimistic Locking)
  - 使用 `version` 字段实现并发控制
  - 每次更新库存时检查 version，防止并发更新冲突

- ✅ **库存预留机制** (Stock Reservation)
  - 下单时先预留库存（从 available_stock 转到 reserved_stock）
  - 预留有过期时间（默认30分钟），超时自动释放
  - 支付成功后从预留库存扣除
  - 订单取消时释放预留库存

### 3. 库存操作 (Stock Operations)
- ✅ **预留库存** (Reserve) - 下单时调用
- ✅ **释放库存** (Release) - 订单取消时调用
- ✅ **扣减库存** (Deduct) - 订单支付成功时调用
- ✅ **入库补货** (Restock) - 采购入库时调用
- ✅ **库存调整** (Adjust) - 盘点或修正时调用

### 4. 审计和追溯 (Audit Trail)
- ✅ **库存日志** (Inventory Logs)
  - 记录所有库存变动
  - 包含操作类型、数量、操作前后状态
  - 记录操作人和原因
  - 关联订单ID（如有）

### 5. 库存预留管理 (Reservation Management)
- ✅ 预留记录追踪
- ✅ 预留状态管理（active/confirmed/cancelled/expired）
- ✅ 自动清理过期预留

### 6. 库存查询 (Stock Check)
- ✅ 单个商品库存检查
- ✅ 批量商品库存检查
- ✅ 实时库存状态查询

## 数据库设计亮点

### 1. 库存表 (inventory)
```sql
- available_stock: 可用库存（可以卖的）
- reserved_stock: 预留库存（已下单未支付的）
- total_stock: 计算列 = available_stock + reserved_stock
- version: 乐观锁版本号
- low_stock_threshold: 低库存阈值
```

### 2. 库存日志表 (inventory_logs)
- 完整记录所有库存变动
- 支持审计和问题追溯
- 记录变动前后的状态

### 3. 库存预留表 (inventory_reservations)
- 跟踪每个订单的库存预留
- 支持过期自动释放
- 防止重复预留（product_id + order_id 唯一约束）

## 防止超卖的完整流程

### 场景1：正常下单流程
```
1. 用户下单 → ReserveStock()
   - 检查 available_stock >= quantity
   - 使用乐观锁更新：available_stock -= quantity, reserved_stock += quantity
   - 创建预留记录（30分钟后过期）

2. 用户支付成功 → DeductStock()
   - reserved_stock -= quantity
   - 确认预留状态为 confirmed

3. 发货完成
   - 库存已扣减，无需操作
```

### 场景2：订单取消
```
1. 用户取消订单 → ReleaseStock()
   - available_stock += quantity
   - reserved_stock -= quantity
   - 预留状态设置为 cancelled
```

### 场景3：超时未支付
```
1. 定时任务 → CleanupExpiredReservations()
   - 查找所有过期的 active 预留
   - 调用 ReleaseStock() 释放库存
   - 预留状态设置为 expired
```

## 未来扩展特性建议

### 高级特性 (Advanced Features)

#### 1. 分布式库存管理
- [ ] **多仓库库存** (Multi-warehouse)
  - 支持多个仓库的库存分配
  - 智能路由（就近发货）
  - 仓库间调拨

- [ ] **库存池化** (Inventory Pooling)
  - 虚拟库存池
  - 跨区域库存共享

#### 2. 高级防超卖策略
- [ ] **分布式锁** (Distributed Lock)
  - Redis 分布式锁
  - 适用于高并发场景

- [ ] **库存分片** (Inventory Sharding)
  - 按商品ID分片存储
  - 提高并发能力

- [ ] **预留配额管理** (Quota Management)
  - 为大客户/活动预留配额
  - 支持配额优先级

#### 3. 智能库存管理
- [ ] **安全库存计算** (Safety Stock)
  - 基于历史销售数据
  - 自动调整低库存阈值

- [ ] **库存预测** (Inventory Forecasting)
  - 机器学习预测需求
  - 提前补货提醒

- [ ] **ABC分类管理**
  - 按销量/价值分类商品
  - 差异化库存策略

#### 4. 性能优化
- [ ] **缓存机制** (Caching)
  - Redis 缓存热门商品库存
  - 减少数据库压力
  - 缓存一致性保证

- [ ] **读写分离** (Read-Write Splitting)
  - 库存查询走从库
  - 库存更新走主库

- [ ] **异步处理** (Async Processing)
  - MQ 异步扣减库存
  - 提高响应速度

#### 5. 库存监控和告警
- [ ] **实时监控大屏**
  - 库存变动趋势
  - 低库存商品数量
  - 预留超时率

- [ ] **智能告警** (Smart Alerts)
  - 低库存告警
  - 异常库存变动告警
  - 预留超时告警

#### 6. 库存盘点
- [ ] **定期盘点** (Inventory Count)
  - 支持盘点单创建
  - 盘盈盘亏处理
  - 差异分析报告

- [ ] **循环盘点** (Cycle Counting)
  - 按ABC分类循环盘点
  - 减少停业盘点时间

#### 7. 库存履约
- [ ] **可承诺量计算** (ATP - Available to Promise)
  - 考虑在途库存
  - 考虑预计到货
  - 更准确的库存可用性

- [ ] **预售管理** (Pre-sale)
  - 支持预售商品
  - 分批发货

#### 8. 历史数据管理
- [ ] **库存快照** (Inventory Snapshot)
  - 定期保存库存快照
  - 支持历史查询

- [ ] **日志归档** (Log Archiving)
  - 历史日志归档
  - 冷热数据分离

## API 端点说明

### 公开端点
- `GET /inventory/check/:product_id` - 检查单个商品库存
- `POST /inventory/check/batch` - 批量检查商品库存

### 管理端点（需要认证）
- `POST /inventory` - 创建库存记录
- `GET /inventory` - 查询库存列表
- `GET /inventory/product/:product_id` - 查询商品库存
- `GET /inventory/low-stock` - 查询低库存商品
- `POST /inventory/restock` - 补货
- `POST /inventory/adjust` - 调整库存
- `PUT /inventory/:product_id/threshold` - 更新低库存阈值
- `GET /inventory/logs/:product_id` - 查询库存日志

### 内部端点（系统调用）
- `POST /inventory/reserve` - 预留库存
- `POST /inventory/release` - 释放库存
- `POST /inventory/deduct` - 扣减库存
- `POST /inventory/cleanup-expired` - 清理过期预留

## 与其他领域的集成

### Order 领域
```go
// 创建订单时
1. 先调用 inventory.ReserveStock() 预留库存
2. 如果预留成功，创建订单
3. 如果预留失败，提示库存不足

// 支付成功后
1. 调用 inventory.DeductStock() 扣减库存
2. 调用 inventory.ConfirmReservation() 确认预留

// 取消订单时
1. 调用 inventory.ReleaseStock() 释放库存
2. 调用 inventory.CancelReservation() 取消预留
```

### Product 领域
```go
// 创建商品时
1. 创建商品后自动创建库存记录
2. 初始库存为0或指定值

// 更新库存阈值
1. 通过 inventory.UpdateLowStockThreshold() 更新
```

## 监控指标建议

### 关键指标
1. **库存周转率** - 衡量库存效率
2. **库存准确率** - 实物与系统一致性
3. **缺货率** - 影响用户体验
4. **超卖事件数** - 系统可靠性
5. **预留超时率** - 订单转化率

### 性能指标
1. **库存查询 QPS**
2. **库存更新 TPS**
3. **预留操作平均耗时**
4. **乐观锁冲突率**

## 最佳实践

### 1. 并发控制
- 始终使用乐观锁更新库存
- 捕获版本冲突，重试或提示用户

### 2. 预留时间设置
- 根据支付流程合理设置预留时间
- 考虑支付渠道的响应时间

### 3. 批量操作
- 大批量操作使用事务
- 考虑分批处理避免长事务

### 4. 日志管理
- 定期归档历史日志
- 保留关键操作日志用于审计

### 5. 监控告警
- 设置低库存告警
- 监控异常库存变动
- 关注乐观锁冲突率

## 技术栈

- **数据库**: PostgreSQL
- **ORM/Query Builder**: SQLC
- **并发控制**: 乐观锁 (Optimistic Locking)
- **事务管理**: Database Transaction
- **日志**: 结构化日志记录所有库存变动

## 测试建议

### 单元测试
- 测试各种库存操作
- 测试并发场景（乐观锁）
- 测试边界条件（库存为0、负数等）

### 集成测试
- 测试完整的订单-库存流程
- 测试预留超时场景
- 测试批量操作

### 压力测试
- 高并发下单场景
- 秒杀场景模拟
- 验证无超卖发生

## 总结

Inventory 领域是电商系统的核心，当前实现已经包含了：
1. ✅ 完善的防超卖机制（乐观锁 + 预留机制）
2. ✅ 完整的库存操作和审计追溯
3. ✅ 灵活的库存查询和管理
4. ✅ 可扩展的架构设计

未来可以根据业务需求逐步添加多仓库管理、智能补货、缓存优化等高级特性。
