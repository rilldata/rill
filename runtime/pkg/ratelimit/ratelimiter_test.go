package ratelimit

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis_rate/v10"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/peer"
	"net"
	"testing"
)

type mockClaims struct {
	subject string
}

func (m *mockClaims) Subject() string {
	return m.subject
}

func (m *mockClaims) Can(p auth.Permission) bool {
	return false
}

func (m *mockClaims) CanInstance(instanceID string, p auth.Permission) bool {
	return false
}

func TestRequestRateLimiter_LimitKeyedAuthRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("allowed & not allowed keyed authorized request", func(t *testing.T) {
		mr.FlushDB()
		ctx := auth.WithClaims(context.Background(), &mockClaims{subject: "subject"})

		err := limiter.LimitKeyedAuthRequest(ctx, "testKey", redis_rate.PerMinute(1))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = limiter.LimitKeyedAuthRequest(ctx, "testKey", redis_rate.PerMinute(1))
		if !errors.As(err, &QuotaExceededError{}) {
			t.Errorf("QuotaExceededError expected: %v", err)
		}
	})
}

func TestRequestRateLimiter_LimitKeyedAnonRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("allowed & not allowed keyed anonymous request", func(t *testing.T) {
		mr.FlushDB()
		ctx := context.Background()

		err := limiter.LimitKeyedAnonRequest(ctx, "testKey", redis_rate.PerMinute(1))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = limiter.LimitKeyedAnonRequest(ctx, "testKey", redis_rate.PerMinute(1))
		if !errors.As(err, &QuotaExceededError{}) {
			t.Errorf("QuotaExceededError expected: %v", err)
		}
	})
}

func TestRequestRateLimiter_LimitAuthRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("allowed & not allowed authorized request", func(t *testing.T) {
		mr.FlushDB()
		ctx := auth.WithClaims(context.Background(), &mockClaims{subject: "subject"})

		request := &runtimev1.QueryRequest{}
		err := limiter.LimitAuthRequest(ctx, request, redis_rate.PerMinute(1))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = limiter.LimitAuthRequest(ctx, request, redis_rate.PerMinute(1))
		if !errors.As(err, &QuotaExceededError{}) {
			t.Errorf("QuotaExceededError expected: %v", err)
		}
	})
}

func TestRequestRateLimiter_LimitAnonRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("allowed & not allowed anonymous request", func(t *testing.T) {
		mr.FlushDB()
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}})

		req := &runtimev1.QueryRequest{}
		err := limiter.LimitAnonRequest(ctx, req, redis_rate.PerMinute(1))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = limiter.LimitAnonRequest(ctx, req, redis_rate.PerMinute(1))
		if !errors.As(err, &QuotaExceededError{}) {
			t.Errorf("QuotaExceededError expected: %v", err)
		}
	})
}
