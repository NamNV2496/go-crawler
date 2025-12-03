package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/redis/go-redis/v9"
)

type CacheObj interface {
	HashKey(key any) string
	Seriablize(key any) string
	Deserialize(data any, output any) error
}

type ICache[E CacheObj] interface {
	Get(ctx context.Context, key any) (*E, error)
	Set(ctx context.Context, key, value any, expiredTime time.Duration) error
	Incr(ctx context.Context, key any) *redis.IntCmd
	Decr(ctx context.Context, key any) *redis.IntCmd
	Expire(ctx context.Context, key any, expiredTime time.Duration) error
}

type cache[E CacheObj] struct {
	client *redis.Client
}

func NewCache[E CacheObj](
	conf *configs.Config,
) *cache[E] {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})
	return &cache[E]{
		client: rdb,
	}
}

func (_self cache[E]) Get(ctx context.Context, key any) (*E, error) {
	val, err := _self.client.Get(ctx, key.(string)).Result()
	if err != nil {
		slog.Debug("failed to get cache", slog.String("key", key.(string)), slog.Any("error", err))
		return nil, err
	}
	var resp E
	err = resp.Deserialize(val, &resp)
	if err != nil {
		slog.Debug("failed to parse cache", slog.String("key", key.(string)), slog.Any("error", err))
		return nil, err
	}
	return &resp, nil
}

func (_self cache[E]) Set(ctx context.Context, key, value any, expiredTime time.Duration) error {
	err := _self.client.Set(ctx, key.(string), value.(string), expiredTime).Err()
	if err != nil {
		slog.Debug("failed to set cache", slog.String("key", key.(string)), slog.Any("error", err))
		return err
	}
	return nil
}

func (_self cache[E]) Incr(ctx context.Context, key any) *redis.IntCmd {
	return _self.client.Incr(ctx, key.(string))
}

func (_self cache[E]) Decr(ctx context.Context, key any) *redis.IntCmd {
	return _self.client.Decr(ctx, key.(string))
}

func (_self cache[E]) Expire(ctx context.Context, key any, expiredTime time.Duration) error {
	return _self.client.Expire(ctx, key.(string), expiredTime).Err()
}
