package ratelimit

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis_rate/v10"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/server/auth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnaryServerInterceptor(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("No Error", func(t *testing.T) {
		mr.FlushDB()
		interceptor := limiter.Middleware()
		ctx := auth.WithClaims(context.Background(), &mockClaims{subject: "subject"})
		nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
		_, err := interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)

		if !errors.Is(err, nil) {
			t.Errorf("Expected error %v, got %v", nil, err)
		}
	})

	t.Run("Quota Exceeded Error", func(t *testing.T) {
		mr.FlushDB()
		interceptor := limiter.Middleware().WithAuthLimit(redis_rate.PerMinute(1))
		ctx := auth.WithClaims(context.Background(), &mockClaims{subject: "subject"})
		nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
		_, err := interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)
		_, err = interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)

		if !strings.HasPrefix(err.Error(), "rpc error: code = ResourceExhausted") {
			t.Errorf("Expected ResourceExhausted rpc error, got %v", err)

		}
	})
}

func TestHTTPHandler(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mr.FlushDB()
	defer mr.Close()

	limiter := NewRequestRateLimiter(mr.Addr())

	t.Run("No Error", func(t *testing.T) {
		mr.FlushDB()
		handler := limiter.Middleware().HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "http://rilldata.com/foo", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Quota Exceeded Error", func(t *testing.T) {
		mr.FlushDB()
		interceptor := limiter.Middleware().WithAnonLimit(redis_rate.PerMinute(1))
		handler := interceptor.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "http://rilldata.com/foo", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected status %v, got %v", http.StatusTooManyRequests, resp.StatusCode)
		}
	})
}
