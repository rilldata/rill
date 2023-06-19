package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"math"
)

// RequestRateLimiter offers rate limiting functionality using a Redis-based rate limiter from the `go-redis/redis_rate`.
// The RequestRateLimiter supports the concept of 'No-operation' (Noop) that performs no rate limiting.
// This can be useful in local/testing environments or when rate limiting is not required.
type RequestRateLimiter struct {
	*redis_rate.Limiter
}

func NewLimiter(client *redis.Client) *RequestRateLimiter {
	return &RequestRateLimiter{Limiter: redis_rate.NewLimiter(client)}
}

func NewNoop() *RequestRateLimiter {
	return &RequestRateLimiter{Limiter: nil}
}

func (l *RequestRateLimiter) Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	if l.Limiter == nil {
		return nil
	}

	if limit == Unlimited {
		return nil
	}

	if limit.IsZero() {
		return NewQuotaExceededError("Resource quota not provided")
	}

	rateResult, err := l.Allow(ctx, limitKey, limit)
	if err != nil {
		return err
	}

	if rateResult.Allowed == 0 {
		return NewQuotaExceededError(fmt.Sprintf("Rate limit exceeded. Try again in %v seconds", rateResult.RetryAfter))
	}

	return nil
}

var Default = redis_rate.PerMinute(60)

var Sensitive = redis_rate.PerMinute(10)

var Public = redis_rate.PerMinute(250)

var Unlimited = redis_rate.PerSecond(math.MaxInt)

var Zero = redis_rate.Limit{}

type QuotaExceededError struct {
	message string
}

func (e QuotaExceededError) Error() string {
	return e.message
}

func NewQuotaExceededError(message string) QuotaExceededError {
	return QuotaExceededError{message}
}
