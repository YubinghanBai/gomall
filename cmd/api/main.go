package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "gomall/docs"

	"gomall/db"
	"gomall/internal/cache"
	"gomall/internal/config"
	"gomall/internal/domain/category"
	"gomall/internal/domain/inventory"
	"gomall/internal/domain/order"
	"gomall/internal/domain/product"
	"gomall/internal/domain/user"
	"gomall/utils/mail"
	"gomall/utils/token"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	// 1. Load Config
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to database
	pool, err := db.NewPostgresPool(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("✅ Connected to database successfully")

	// 3. Initialize Redis Cache
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	cacheClient, err := cache.NewRedisCache(cache.Config{
		Addr:         redisAddr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis successfully")

	// 4. Create Token Maker
	tokenMaker, err := token.NewJWTMaker(cfg.JWT.Secret)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}

	emailSender := mail.NewGmailSender(
		cfg.Email.SenderName,
		cfg.Email.SenderEmail,
		cfg.Email.SenderPassword)

	// 5. Initialize User domain
	userRepo := user.NewRepository(pool)
	userService := user.NewService(cfg, userRepo, tokenMaker, emailSender)
	userHandler := user.NewHandler(userService, tokenMaker)

	// Initialize Product domain
	productRepo := product.NewRepository(pool)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// Initialize Category domain
	categoryRepo := category.NewRepository(pool)
	categoryService := category.NewService(categoryRepo)
	categoryHandler := category.NewHandler(categoryService)

	// Inventory
	inventoryRepo := inventory.NewRepository(pool)
	inventoryService := inventory.NewService(inventoryRepo)
	inventoryHandler := inventory.NewHandler(inventoryService)

	// Order
	orderRepo := order.NewRepository(pool)
	orderService := order.NewService(orderRepo, inventoryService, productService)
	orderHandler := order.NewHandler(orderService)

	// 6. Init Router
	gin.SetMode(cfg.Server.Mode)
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

		// Register Product Route
		productHandler.RegisterRoutes(api)

		// Register Category Route
		categoryHandler.RegisterRoutes(api)

		// Register Inventory Route
		inventoryHandler.RegisterRoutes(api)  

		//Register Order Route
		orderHandler.RegisterRoutes(api)

	}

	go startInventoryCleanupJob(inventoryService)

	// 7. Start Service
	log.Printf("🚀 Server starting on %s", cfg.Server.Port)
	log.Printf("📦 Redis cache client initialized: %v", cacheClient != nil)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


func startInventoryCleanupJob(inventoryService inventory.Service){
	ticker:=time.NewTicker(5*time.Minute)
	defer ticker.Stop()

	log.Println("Inventory cleanup job started,running every 5 minutes")

	for{
		select{
		case <-ticker.C:
			ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)
			err:=inventoryService.CleanupExpiredReservations(ctx)
			if err!=nil{
				log.Printf("Failed to cleanup expired reservations: %v",err)
			}else{
				log.Println("Successfully cleanned up expired reservations")
			}
			cancel()
		}
	}
}