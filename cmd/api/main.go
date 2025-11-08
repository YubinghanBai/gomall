package main

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "gomall/docs"
	"gomall/utils/mail"
	"log"

	"github.com/gin-gonic/gin"
	"gomall/config"
	"gomall/db"
	"gomall/internal/user"
	"gomall/utils/token"
)

// @title           GoMall API
// @version         1.0
// @description     Online Shopping API Service
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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

	// 5. Init Router
	gin.SetMode(cfg.ServerConfig.Mode)
	r := gin.Default()

	//Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API Route
	api := r.Group("/api/v1")
	{
		// Register User Route
		userHandler.RegisterRoutes(api)
	}

	// 6. Start Service
	log.Printf("🚀 Server starting on %s", cfg.ServerConfig.Port)
	if err := r.Run(cfg.ServerConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
