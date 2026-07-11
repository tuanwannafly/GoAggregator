package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string, db int) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   db,
		}),
	}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	raw, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return ErrMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, raw, ttl).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

type NoopCache struct{}

func (NoopCache) Get(ctx context.Context, key string, dest any) error {
	return ErrMiss
}

func (NoopCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

// MemoryCache is an in-memory cache implementation for testing
type MemoryCache struct {
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheEntry),
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest any) error {
	entry, ok := c.data[key]
	if !ok {
		return ErrMiss
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return ErrMiss
	}
	return json.Unmarshal(entry.value, dest)
}

func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.data[key] = cacheEntry{
		value:     raw,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}
