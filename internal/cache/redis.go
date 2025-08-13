package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(redisURL string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	cli := redis.NewClient(opt)
	return &Redis{Client: cli}, nil
}

// AllowSlidingWindow increments a counter per key with TTL window and returns whether allowed under max.
func (r *Redis) AllowSlidingWindow(ctx context.Context, key string, window time.Duration, max int64) (bool, int64, error) {
	pipe := r.Client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	val := incr.Val()
	return val <= max, val, nil
}