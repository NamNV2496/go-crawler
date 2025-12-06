package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	redisratev9 "github.com/go-redis/redis_rate/v9"
	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/pkg/logging"
)

const (
	END_LIMIT_TIME = "end_limit_time"
	BLOCK_COUNT    = "count"
)

var DEFAUL_RATELIMI_OPTION = &RatelimitOpt{
	BlockRetention: time.Hour * 24,
	CalculateBlockDuration: func(count int) time.Duration {
		return time.Hour * 24 * time.Duration(count)
	},
}

type IRateLimit interface {
	Allow(ctx context.Context, path, key string, limits ...redisratev9.Limit) (pass bool, err error)
}

type RatelimitOpt struct {
	// BlockRetention is the time to block request
	BlockRetention time.Duration
	// CalculateBlockDuration is the function to calculate block duration
	CalculateBlockDuration func(count int) time.Duration
}

type RateLimit struct {
	// redis to save blocked status
	redisClient *redis.Client
	// limiter for rate limit
	limiter *redisratev9.Limiter
	// option for rate limiter
	opts *RatelimitOpt
}

func NewRateLimitWithOption(
	conf *configs.Config,
	opts *RatelimitOpt,
) *RateLimit {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	rateLimiter := redisratev9.NewLimiter(client)
	if opts == nil {
		opts = DEFAUL_RATELIMI_OPTION
	}
	return &RateLimit{
		redisClient: client,
		limiter:     rateLimiter,
		opts:        opts,
	}
}

var _ IRateLimit = &RateLimit{}

func LimitPerSecond(rate, burst int) redisratev9.Limit {
	return redisratev9.Limit{
		Rate:   rate,
		Period: time.Second,
		Burst:  burst,
	}
}

func LimitPerMinute(rate, burst int) redisratev9.Limit {
	return redisratev9.Limit{
		Rate:   rate,
		Period: time.Minute,
		Burst:  burst,
	}
}

func LimitCustom(rate, burst int, period time.Duration) redisratev9.Limit {
	return redisratev9.Limit{
		Rate:   rate,
		Period: period,
		Burst:  burst,
	}
}

func (_self *RateLimit) Allow(ctx context.Context, path, key string, limits ...redisratev9.Limit) (pass bool, err error) {
	ctx = logging.AppendPrefix(ctx, "Allow")
	logging.Infof(ctx, "allow")
	if len(limits) == 0 {
		return true, nil
	}
	// check is the request blocked
	redisKey := buildBlockedRedisKey(path, key)
	isBlocked, err := _self.isBlocked(ctx, redisKey)
	if err != nil {
		return false, fmt.Errorf("request is blocked: %w", err)
	}
	if isBlocked {
		return false, nil
	}

	// rate limit execute
	isLimited, err := _self.rateLimit(ctx, path, key, limits...)
	if err != nil {
		return false, err
	}
	if isLimited {
		// block if reach limit
		if err := _self.block(ctx, path, key); err != nil {
			return false, err
		}
		return false, nil // <- Not allowed, rate limit hit!
	}
	return true, nil
}

func (_self *RateLimit) rateLimit(ctx context.Context, path, key string, limits ...redisratev9.Limit) (bool, error) {
	for _, limit := range limits {
		keyLimit := buildLimitRedisKey(path, key, limit)
		result, err := _self.limiter.Allow(ctx, keyLimit, limit)
		if err != nil {
			return false, err
		}
		if result.Allowed == 0 {
			return true, nil
		}
	}
	return false, nil
}

func (_self *RateLimit) isBlocked(ctx context.Context, key string) (bool, error) {
	endTime, err := _self.getBlockEndTime(ctx, key)
	if err != nil {
		return false, err
	}
	if endTime.IsZero() || time.Now().After(endTime) {
		return false, nil
	}
	return true, nil
}

func buildBlockedRedisKey(path, key string) string {
	return fmt.Sprintf("blocked.%s.%s", path, key)
}

func buildLimitRedisKey(path, key string, limit redisratev9.Limit) string {
	return fmt.Sprintf("ratelimit.%s.%s.%d", path, key, limit.Rate)
}

func (_self *RateLimit) getBlockEndTime(ctx context.Context, key string) (time.Time, error) {
	endTimeStr, err := _self.redisClient.HGet(ctx, key, BLOCK_COUNT).Result()
	if err != nil && err != redis.Nil {
		return time.Time{}, err
	}
	if endTimeStr == "" {
		return time.Time{}, nil
	}
	endTime, err := time.Parse(time.RFC3339Nano, endTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	return endTime, nil
}

func (_self *RateLimit) setBlockEndTime(ctx context.Context, key string, endTime time.Time) error {
	return _self.redisClient.HSet(ctx, key, BLOCK_COUNT, endTime.Format(time.RFC3339Nano)).Err()
}

func (_self *RateLimit) block(ctx context.Context, path, key string) error {
	blockedKey := buildBlockedRedisKey(path, key)
	// increase block counter
	count, err := _self.redisClient.HIncrBy(ctx, blockedKey, BLOCK_COUNT, 1).Result()
	if err != nil {
		return err
	}
	// calculate new end time
	newEndTime := _self.opts.CalculateBlockDuration(int(count))
	// set block end time
	err = _self.setBlockEndTime(ctx, blockedKey, time.Now().Add(newEndTime))
	if err != nil {
		return err
	}
	// reset retention
	_self.ResetRetention(ctx, blockedKey)
	return nil
}

func (_self *RateLimit) ResetRetention(ctx context.Context, blockedKey string) error {
	return _self.redisClient.Expire(ctx, blockedKey, _self.opts.BlockRetention).Err()
}
