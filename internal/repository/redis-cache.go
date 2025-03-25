package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
	}, nil
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(key string) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return result, true
}

// Set adds a value to the cache
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(ctx, key, data, expiration)
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c.client.Del(ctx, key)
}

// Close closes the Redis client connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}