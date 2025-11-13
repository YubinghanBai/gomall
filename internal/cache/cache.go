package cache

import (
	"context"
	"time"
)

// Cache 定义通用缓存接口
type Cache interface {
	
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	
	MGet(ctx context.Context, keys ...string) ([]string, error)
	MSet(ctx context.Context, kvs map[string]interface{}, expiration time.Duration) error
	
	
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key, field string, value interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	
	
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	
	
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// Config Redis Config
type Config struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}