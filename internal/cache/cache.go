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
