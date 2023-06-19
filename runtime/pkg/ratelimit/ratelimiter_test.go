package ratelimit

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestRequestRateLimiter_Limit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	opts, err := redis.ParseURL("redis://" + mr.Addr())
	if err != nil {
		panic(err)
	}
	limiter := NewLimiter(redis.NewClient(opts))

	t.Run("allowed & not allowed keyed authorized request", func(t *testing.T) {
		mr.FlushDB()
		ctx := context.Background()

		err := limiter.Limit(ctx, "testKey", redis_rate.PerMinute(1))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = limiter.Limit(ctx, "testKey", redis_rate.PerMinute(1))
		if !errors.As(err, &QuotaExceededError{}) {
			t.Errorf("QuotaExceededError expected: %v", err)
		}
	})
}
