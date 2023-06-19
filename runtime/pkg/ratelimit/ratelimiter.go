package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"math"
	"net/http"
	"reflect"
)

// RequestRateLimiter is a struct that provides rate limiting functionality
// leveraging the Redis-based rate limiter from the `go-redis/redis_rate` package.
// It offers methods to control rate limiting for different types of requests
// (anonymous vs authenticated) and for different request keys.
//
// A Redis client is initialized in the constructor with the provided Redis address.
// If no address is provided, the rate limiter is not set up.
//
// The rate limits can be adjusted on a per-key basis, and different limits
// can be set up for authenticated and anonymous requests. These limits are applied
// each time a request is made, and an error is returned if the rate limit is exceeded.
//
// The package also defines several rate limiting constants, which can be used
// to adjust the rate limiting limits: Default, Sensitive, Public, Unlimited, and Zero.
//
// Additionally, a custom error type, QuotaExceededError, is used to indicate
// when the rate limit for a particular key has been exceeded.
type RequestRateLimiter struct {
	*redis_rate.Limiter
}

// NewRequestRateLimiter is a constructor for creating a new RequestRateLimiter instance.
// It takes a string representing the Redis address as input. If the address string
// is not empty, a new Redis client is initialized with this address and a
// rate limiter is set up using this client.
//
// If the Redis address is empty, the constructor still returns a valid
// RequestRateLimiter instance, but the Limiter within it is not initialized,
// effectively making this rate limiter a no-op. E.g. in local environment.
func NewRequestRateLimiter(redisAddr string) *RequestRateLimiter {
	// if RedisAddr is not passed then rateLimiter doesn't limit user requests
	var limiter *redis_rate.Limiter
	if redisAddr != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		limiter = redis_rate.NewLimiter(redisClient)
	}
	return &RequestRateLimiter{Limiter: limiter}
}

func (l *RequestRateLimiter) LimitKeyedRequest(ctx context.Context, limitKey string, authLimit, anonLimit redis_rate.Limit) error {
	if l.Limiter == nil {
		return nil
	}

	var limit redis_rate.Limit
	if isAnonymous(ctx) {
		// Anonymous request
		limit = anonLimit
	} else {
		// Authorized request
		limit = authLimit
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

func (l *RequestRateLimiter) LimitKeyedAuthRequest(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	return l.LimitKeyedRequest(ctx, limitKey, limit, Unlimited)
}

func (l *RequestRateLimiter) LimitKeyedAnonRequest(ctx context.Context, limitKey string, limit redis_rate.Limit) error {
	return l.LimitKeyedRequest(ctx, limitKey, Unlimited, limit)
}

func (l *RequestRateLimiter) LimitRequest(ctx context.Context, req interface{}, authLimit, anonLimit redis_rate.Limit) error {
	return l.LimitKeyedRequest(ctx, getRequestLimitKey(ctx, req), authLimit, anonLimit)
}

func (l *RequestRateLimiter) LimitAuthRequest(ctx context.Context, req interface{}, limit redis_rate.Limit) error {
	return l.LimitRequest(ctx, req, limit, Unlimited)
}

func (l *RequestRateLimiter) LimitAnonRequest(ctx context.Context, req interface{}, limit redis_rate.Limit) error {
	return l.LimitRequest(ctx, req, Unlimited, limit)
}

var Default = redis_rate.PerMinute(60)

var Sensitive = redis_rate.PerMinute(10)

var Public = redis_rate.PerMinute(250)

var Unlimited = redis_rate.PerSecond(math.MaxInt)

var Zero = redis_rate.Limit{}

func isAnonymous(ctx context.Context) bool {
	claims := auth.GetClaims(ctx)
	isAnonymous := claims == nil || claims.Subject() == ""
	return isAnonymous
}

func getRequestLimitKey(ctx context.Context, req interface{}) string {
	if isAnonymous(ctx) {
		// Anonymous request
		var ipAddr string
		switch r := req.(type) {
		case *http.Request:
			ipAddr = observability.HTTPPeer(r)
		default:
			ipAddr = observability.GrpcPeer(ctx)
		}
		return fmt.Sprintf("anonymous-req-rate-limit-%s-%s", getRequestName(req), ipAddr)
	}
	// Authorized request
	return fmt.Sprintf("authorized-req-rate-limit-%s-%s", getRequestName(req), auth.GetClaims(ctx).Subject())
}

func getRequestName(req interface{}) string {
	if req == nil {
		return "unnamed"
	}
	switch r := req.(type) {
	case *http.Request:
		return r.URL.Path
	default:
		t := reflect.TypeOf(req)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t.Name()
	}
}

type QuotaExceededError struct {
	message string
}

func (e QuotaExceededError) Error() string {
	return e.message
}

func NewQuotaExceededError(message string) QuotaExceededError {
	return QuotaExceededError{message}
}
