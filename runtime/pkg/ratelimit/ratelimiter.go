package ratelimit

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

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

// Common rate limit configurations
var (
	Default   = redis_rate.PerMinute(180)
	Sensitive = redis_rate.PerMinute(30)
	Public    = redis_rate.PerMinute(750)
	Unlimited = redis_rate.PerSecond(math.MaxInt)
	Zero      = redis_rate.Limit{}
)

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

// Common Redis error messages
const (
	errMaxClients  = "ERR max number of clients reached"
	errLoading     = "LOADING "
	errReadOnly    = "READONLY "
	errMasterDown  = "MASTERDOWN "
	errClusterDown = "CLUSTERDOWN "
	errTryAgain    = "TRYAGAIN "
)

// If the error is a server connection error, we should not return an error. This is because the server may be temporarily unavailable, and we should not block the request. The client should retry the request.
func isServerConnError(err error) bool {
	switch err {
	case io.EOF:
		return true
	case io.ErrUnexpectedEOF:
		return true
	default:
		return parseRedisError(err)
	}
}

func parseRedisError(err error) bool {
	s := err.Error()
	if s == errMaxClients {
		return true
	}
	if strings.HasPrefix(s, errLoading) {
		return true
	}
	if strings.HasPrefix(s, errReadOnly) {
		return true
	}
	if strings.HasPrefix(s, errMasterDown) {
		return true
	}
	if strings.HasPrefix(s, errClusterDown) {
		return true
	}
	if strings.HasPrefix(s, errTryAgain) {
		return true
	}
	return false
}
