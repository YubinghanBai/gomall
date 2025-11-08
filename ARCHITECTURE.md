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

