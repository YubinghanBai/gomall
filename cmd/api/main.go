package main

import (
	"gomall/utils/mail"
	"log"

	"github.com/gin-gonic/gin"
	"gomall/config"
	"gomall/db"
	"gomall/internal/user"
	"gomall/utils/token"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 连接数据库
	pool, err := db.NewPostgresPool(&cfg.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("✅ Connected to database successfully")

	// 3. 创建 Token Maker
	tokenMaker, err := token.NewJWTMaker(cfg.JWTConfig.Secret)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}

	emailSender := mail.NewGmailSender(
		cfg.EmailConfig.SenderName,
		cfg.EmailConfig.SenderEmail,
		cfg.EmailConfig.SenderPassword)

	// 4. 初始化依赖（User 领域）
	userRepo := user.NewRepository(pool)
	userService := user.NewService(cfg, userRepo, tokenMaker, emailSender)
	userHandler := user.NewHandler(userService, tokenMaker)

	// 5. 初始化路由
	gin.SetMode(cfg.ServerConfig.Mode)
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由
	api := r.Group("/api/v1")
	{
		// 注册 User 路由
		userHandler.RegisterRoutes(api)
	}

	// 6. 启动服务
	log.Printf("🚀 Server starting on %s", cfg.ServerConfig.Port)
	if err := r.Run(cfg.ServerConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
