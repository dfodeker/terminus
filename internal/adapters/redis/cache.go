// internal/adapters/redis/cache.go
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements the ports.Cache interface using Redis
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Redis-based cache
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// GetTyped retrieves a cached value and unmarshals it into the provided type
func GetTyped[T any](ctx context.Context, c *Cache, key string) (*T, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
