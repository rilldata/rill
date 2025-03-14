package ratelimit

import (
	"context"
	"fmt"
	"math"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// Limiter returns an error if quota per key is exceeded.
type Limiter interface {
	Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error
	Ping(ctx context.Context) error
}

// Redis offers rate limiting functionality using a Redis-based rate limiter.
type Redis struct {
	*redis_rate.Limiter
	ping func(ctx context.Context) error
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{
		Limiter: redis_rate.NewLimiter(client),
		ping: func(ctx context.Context) error {
			status := client.Ping(ctx)
			return status.Err()
		},
	}
}

func (l *Redis) Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	if limit == Unlimited {
		return nil
	}

	if limit.IsZero() {
		return NewQuotaExceededError("Resource quota not provided")
	}

	rateResult, err := l.Allow(ctx, limitKey, limit)
	if err != nil {
		// Check if the error is a 5xx like error
		if isServerError(err) {
			// Allow the request to proceed without rate limiting
			return nil
		}
		return err
	}

	if rateResult.Allowed == 0 {
		return NewQuotaExceededError(fmt.Sprintf("Rate limit exceeded. Try again in %v seconds", rateResult.RetryAfter))
	}

	return nil
}

func (l *Redis) Ping(ctx context.Context) error {
	return l.ping(ctx)
}

// Noop performs no rate limiting.
// This can be useful in local/testing environments or when rate limiting is not required.
type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n Noop) Limit(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	return nil
}

func (n Noop) Ping(ctx context.Context) error {
	return nil
}

var Default = redis_rate.PerMinute(180)

var Sensitive = redis_rate.PerMinute(30)

var Public = redis_rate.PerMinute(750)

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

func AuthLimitKey(methodName, authID string) string {
	return fmt.Sprintf("auth:%s:%s", methodName, authID)
}

func AnonLimitKey(methodName, peer string) string {
	return fmt.Sprintf("anon:%s:%s", methodName, peer)
}

// isServerError checks if the error is a server-side error (5xx)
func isServerError(err error) bool {
	// go-redis server errors are typically returned as redis.Error
	if _, ok := err.(redis.Error); ok {
		// implement additional logic here to check for specific error messages
		return true
	}
	return false
}
