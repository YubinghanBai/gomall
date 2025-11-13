package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg Config) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Ping to test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &redisCache{client: client}, nil
}

// Get retrieves a value from cache
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, err
}

// Set stores a value in cache with expiration
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		// For complex types, use JSON encoding
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(jsonData)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Delete removes a key from cache
func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// MGet retrieves multiple values
func (r *redisCache) MGet(ctx context.Context, keys ...string) ([]string, error) {
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]string, len(vals))
	for i, val := range vals {
		if val != nil {
			result[i] = val.(string)
		}
	}
	return result, nil
}

// MSet sets multiple key-value pairs
func (r *redisCache) MSet(ctx context.Context, kvs map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()
	for key, value := range kvs {
		var data string
		switch v := value.(type) {
		case string:
			data = v
		case []byte:
			data = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
			}
			data = string(jsonData)
		}
		pipe.Set(ctx, key, data, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// HGet retrieves a field from a hash
func (r *redisCache) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("field not found: %s", field)
	}
	return val, err
}

// HSet sets a field in a hash
func (r *redisCache) HSet(ctx context.Context, key, field string, value interface{}) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(jsonData)
	}
	return r.client.HSet(ctx, key, field, data).Err()
}

// HGetAll retrieves all fields from a hash
func (r *redisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// Incr increments a counter
func (r *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr decrements a counter
func (r *redisCache) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// IncrBy increments a counter by value
func (r *redisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// DecrBy decrements a counter by value
func (r *redisCache) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}

// Expire sets expiration on a key
func (r *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL gets time to live for a key
func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}