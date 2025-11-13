package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Cache      CacheConfig      `mapstructure:"cache"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Email      EmailConfig      `mapstructure:"email"`
	Pagination PaginationConfig `mapstructure:"pagination"`
	Inventory  InventoryConfig  `mapstructure:"inventory"`
	Order      OrderConfig      `mapstructure:"order"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           string        `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN returns PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// CacheConfig holds cache TTL configuration
type CacheConfig struct {
	ProductTTL     time.Duration `mapstructure:"product_ttl"`
	ProductListTTL time.Duration `mapstructure:"product_list_ttl"`
	StockTTL       time.Duration `mapstructure:"stock_ttl"`
	UserSessionTTL time.Duration `mapstructure:"user_session_ttl"`
	HotProductTTL  time.Duration `mapstructure:"hot_product_ttl"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SenderName     string `mapstructure:"sender_name"`
	SenderEmail    string `mapstructure:"sender_email"`
	SenderPassword string `mapstructure:"sender_password"`
}

// PaginationConfig holds pagination configuration
type PaginationConfig struct {
	DefaultPageSize int `mapstructure:"default_page_size"`
	MaxPageSize     int `mapstructure:"max_page_size"`
}

// InventoryConfig holds inventory configuration
type InventoryConfig struct {
	ReservationTTL  time.Duration `mapstructure:"reservation_ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// OrderConfig holds order configuration
type OrderConfig struct {
	PaymentTimeout      time.Duration `mapstructure:"payment_timeout"`
	AutoCancelInterval time.Duration `mapstructure:"auto_cancel_interval"`
}
