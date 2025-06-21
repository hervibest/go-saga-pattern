package adapter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheAdapter interface {
	TTL(ctx context.Context, key string) (time.Duration, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type cacheAdapter struct {
	redisClient *redis.Client
}

func NewCacheAdapter(redisClient *redis.Client) CacheAdapter {
	return &cacheAdapter{
		redisClient: redisClient,
	}
}

func (a *cacheAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return a.redisClient.TTL(ctx, key).Result()
}

func (a *cacheAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.redisClient.Set(ctx, key, value, expiration).Err()
}

func (a *cacheAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.redisClient.Get(ctx, key).Result()
}

func (a *cacheAdapter) Del(ctx context.Context, keys ...string) error {
	return a.redisClient.Del(ctx, keys...).Err()
}
