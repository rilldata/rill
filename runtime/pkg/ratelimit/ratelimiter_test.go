package ratelimit

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
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
	limiter := NewRedis(redis.NewClient(opts))

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

func TestAuthReqLimitKey(t *testing.T) {
	assert.Equal(t, "auth:TestMethod:authID", AuthLimitKey("TestMethod", "authID"))
}

func TestAnonReqLimitKey(t *testing.T) {
	assert.Equal(t, "anon:TestMethod:192.168.0.1", AnonLimitKey("TestMethod", "192.168.0.1"))
}
