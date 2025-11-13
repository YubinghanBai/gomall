User Request
-> Handler (receive request, validate parameter, return response)
-> Service(Core business logic, transaction,  rely on Repository and other services)
-> Repository(CRUD for database , wrap up sqlc)
-> sqlc Code (execute SQL, type transform , rely on pgxpool.Pool)
-> Database(PostgreSQL)

Why we need wrap up sqlc Code
- Provide interface to mock 
- Combine multiple SQL to perform complex queries
- Could add cache and log

Rely on interface instead of Instance

Satisfy SRP and OCP




## Authorization
Authorization: <type> <credentials>
```
[1] 用户输入用户名 + 密码
↓
[2] 后端验证数据库密码是否正确
↓
[3] 生成 JWT token（access_token + refresh_token）
↓
[4] 返回 token 给前端
↓
[5] 前端保存 token（localStorage / cookie）
↓
[6] 每次请求都带上 Authorization: Bearer <token>
↓
[7] Gin AuthMiddleware 拦截请求
↓
[8] VerifyToken 校验签名、过期、payload
↓
[9] 注入用户信息 → Handler 继续执行
```



Product 模块实现计划

📊 Product vs User 核心区别

| 特性   | User 模块                      | Product 模块                 |
  |------|------------------------------|----------------------------|
| 业务类型 | 认证授权                         | 商品管理                       |
| 关联表  | sessions, verification_codes | product_images, categories |
| 并发关注 | Session 阻塞                   | 库存扣减（高并发）                  |
| 事务场景 | 注册+发邮件、重置密码                  | 创建商品+添加图片、库存更新             |
| 流量模式 | 写为主                          | 读:写 = 100:1（需要缓存）          |
| 锁需求  | 低                            | 高（库存更新需要分布式锁）              |
| 缓存需求 | Session cache                | 商品详情、搜索结果、分类列表             |

🎯 Product 模块特有功能

1. 库存管理（最关键）
   - 扣减库存（需要原子性 + 并发控制）
   - 库存预警（低于阈值提醒）
   - 库存流水（可选，用于审计）
2. 图片管理（一对多）
   - 批量添加图片
   - 设置主图（需要原子更新）
   - 图片排序
3. 搜索与筛选
   - 全文搜索
   - 分类筛选
   - 价格区间
   - 销量排序
4. 统计功能
   - 浏览量统计
   - 销量统计
   - 分类统计


阶段 1：数据库层补充（可选优化）

     检查并补充以下 SQL 查询到 db/query/product.sql：
     - 批量查询商品（按 IDs）
     - 版本化库存更新（乐观锁）
     - 批量状态更新
     - 低库存商品查询
     - 价格区间筛选

     运行 make sqlc 生成新方法

     阶段 2：Repository 层

     创建 internal/product/repository.go：
     - 定义领域特定接口（仅包含 Product 相关的20个方法）
     - 包含事务支持 ExecTx
     - 实现所有方法（委托给 sqlc.Store）
     - 关键设计：接口为未来 Redis 缓存和分布式锁预留扩展点

     阶段 3：DTO 层

     创建 internal/product/dto.go：
     - Request DTOs（CreateProductRequest, UpdateProductRequest, ListProductsRequest 等）
     - Response DTOs（ProductResponse, ProductDetailResponse, PaginatedProductsResponse）
     - 验证标签（binding, validate）
     - 图片相关 DTOs

     阶段 4：Service 层

     创建 internal/product/service.go：

     基础 CRUD
     - CreateProduct, GetProduct, UpdateProduct, DeleteProduct, ListProducts

     库存管理（预留锁机制接口）
     - UpdateStock（增减库存）
     - CheckStock（检查库存）
     - GetLowStockProducts（库存预警）

     图片管理（使用事务）
     - CreateProductWithImages（事务：创建商品 + 批量添加图片）
     - SetMainImage（事务：取消旧主图 + 设置新主图）
     - AddProductImages, DeleteProductImage

     搜索与筛选
     - SearchProducts（全文搜索）
     - GetFeaturedProducts（精选商品）
     - GetProductsByCategory（分类筛选）

     统计
     - IncrementViews, IncrementSales

     阶段 5：Handler 层

     创建 internal/product/handler.go：
     - 路由定义在文件开头（方便查看所有 endpoints）
     - REST API endpoints（CRUD、搜索、库存等）
     - 输入验证、错误处理
     - Swagger 注解（仿照 User 模块）

     阶段 6：路由注册

     在 cmd/api/main.go 中：
     - 初始化 Product handler
     - 注册路由组 /api/v1/products

     关键事务场景

     事务 1：创建商品 + 添加图片
     ExecTx(ctx, func(q) {
         product := q.CreateProduct(...)
         for img := range images {
             q.CreateProductImage(product.ID, img)
         }
     })

     事务 2：设置主图
     ExecTx(ctx, func(q) {
         // 取消所有主图标记
         images := q.GetProductImages(productID)
         for img in images {
             if img.IsMain { q.UpdateProductImage(img.ID, IsMain=false) }
         }
         // 设置新主图
         q.UpdateProductImage(newImageID, IsMain=true)
     })

     事务 3：库存扣减（为分布式锁预留）
     // 当前：使用数据库行锁
     DecrementProductStock(WHERE stock >= quantity)

     // 未来：添加 Redis 分布式锁
     lock := redlock.Acquire("lock:product:stock:{id}")
     defer lock.Release()

     未来扩展点设计

     1. Repository 接口：返回 interface 而非具体类型，方便切换实现（加缓存层）
     2. Service 层：库存方法预留 lock 参数，未来可注入 Redis 分布式锁
     3. 缓存接口：在 Service 构造函数预留 cache 参数（可选）
     4. 搜索接口：抽象搜索方法，未来可切换为 Elasticsearch

     测试

     - 运行 make test 验证所有测试通过
     - 运行 make mock 生成 mock
     - 手动测试 API endpoints
     - 验证事务正确性（创建商品 + 图片应原子完成）



