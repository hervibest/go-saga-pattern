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
	HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HSet(ctx context.Context, key string, values ...interface{}) error
	SAdd(ctx context.Context, key string, members ...interface{}) error
	XRead(ctx context.Context, args *redis.XReadArgs) ([]redis.XStream, error)
	XAdd(ctx context.Context, args *redis.XAddArgs) error
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error
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

func (a *cacheAdapter) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return a.redisClient.HMGet(ctx, key, fields...).Result()
}

func (a *cacheAdapter) HSet(ctx context.Context, key string, values ...interface{}) error {
	return a.redisClient.HSet(ctx, key, values...).Err()
}

func (a *cacheAdapter) HDel(ctx context.Context, key string, fields ...string) error {
	return a.redisClient.HDel(ctx, key, fields...).Err()
}

func (a *cacheAdapter) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return a.redisClient.SAdd(ctx, key, members...).Err()
}

func (a *cacheAdapter) XRead(ctx context.Context, args *redis.XReadArgs) ([]redis.XStream, error) {
	return a.redisClient.XRead(ctx, args).Result()
}

func (a *cacheAdapter) XAdd(ctx context.Context, args *redis.XAddArgs) error {
	return a.redisClient.XAdd(ctx, args).Err()
}

func (a *cacheAdapter) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return a.redisClient.PSubscribe(ctx, channels...)
}

func (a *cacheAdapter) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.redisClient.SetEx(ctx, key, value, expiration).Err()
}
