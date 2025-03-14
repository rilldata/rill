package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"

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
		// If the error is a server connection error, we should not return an error.
		// This is because the server may be temporarily unavailable, and we should not block the request.
		// The client should retry the request.
		if isServerConnError(err) {
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

// isServerError checks if the error is a server connection error.
func isServerConnError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	if httpErr, ok := err.(interface{ StatusCode() int }); ok {
		return httpErr.StatusCode() >= 500 && httpErr.StatusCode() < 600
	}

	return false
}
