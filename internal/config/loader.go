package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var globalConfig *Config

// Load loads configuration from config file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file path
	if configPath == "" {
		// Default to config/config.yaml
		configPath = "config/config.yaml"
	}

	// Set config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Allow environment variables to override config file
	v.AutomaticEnv()

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override sensitive values from environment variables if they exist
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}
	if emailPassword := os.Getenv("EMAIL_PASSWORD"); emailPassword != "" {
		cfg.Email.SenderPassword = emailPassword
	}

	globalConfig = &cfg
	return &cfg, nil
}

// LoadFromEnv loads configuration primarily from environment variables
// This is useful for production environments
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("SERVER_PORT", ":8080"),
			Mode: getEnvOrDefault("SERVER_MODE", "release"),
		},
		Database: DatabaseConfig{
			Driver:   getEnvOrDefault("DB_DRIVER", "postgres"),
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvIntOrDefault("DB_PORT", 5432),
			User:     getEnvOrDefault("DB_USER", "root"),
			Password: getEnvOrDefault("DB_PASSWORD", ""),
			DBName:   getEnvOrDefault("DB_NAME", "gomall"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     getEnvIntOrDefault("REDIS_PORT", 6379),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       getEnvIntOrDefault("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: getEnvOrDefault("JWT_SECRET", ""),
		},
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global config instance
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, call Load() or LoadFromEnv() first")
	}
	return globalConfig
}

// MustLoad loads config and panics on error
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// FindConfigFile searches for config file in common locations
func FindConfigFile() (string, error) {
	searchPaths := []string{
		"config/config.yaml",
		"config/config.yml",
		"../config/config.yaml",
		"../config/config.yml",
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("config file not found in any of: %v", searchPaths)
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}
