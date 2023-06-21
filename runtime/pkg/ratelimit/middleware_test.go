package ratelimit

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/grpc/peer"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestContextInspector struct{}

func (i TestContextInspector) IsAuthenticated(ctx context.Context) bool {
	_, ok := ctx.Value("authorization").(string)
	return ok
}

func (i TestContextInspector) GetAuthID(ctx context.Context) string {
	value, ok := ctx.Value("authorization").(string)
	if !ok {
		return ""
	}
	return value
}

var localhost, _ = net.ResolveIPAddr("", "127.0.0.1")

func TestUnaryServerInterceptor(t *testing.T) {
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

	t.Run("No Error (anonymous request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, Unlimited)
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: localhost})
		nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
		_, err := interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)

		if !errors.Is(err, nil) {
			t.Errorf("Expected error %v, got %v", nil, err)
		}
	})

	t.Run("No Error (authenticated request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, Unlimited)
		ctx := context.WithValue(context.Background(), "authorization", "userid")
		nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
		_, err := interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)

		if !errors.Is(err, nil) {
			t.Errorf("Expected error %v, got %v", nil, err)
		}
	})

	t.Run("Quota Exceeded Error (anonymous request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, redis_rate.PerMinute(1), Unlimited)
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: localhost})
		nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}
		_, err := interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)
		_, err = interceptor.UnaryServerInterceptor()(ctx, runtimev1.QueryRequest{}, nil, nextHandler)

		if !strings.HasPrefix(err.Error(), "rpc error: code = ResourceExhausted") {
			t.Errorf("Expected ResourceExhausted rpc error, got %v", err)
		}
	})

	t.Run("Quota Exceeded Error (authenticated request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, redis_rate.PerMinute(1))
		ctx := context.WithValue(context.Background(), "authorization", "userid")
		ctx = peer.NewContext(ctx, &peer.Peer{Addr: localhost})
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

	opts, err := redis.ParseURL("redis://" + mr.Addr())
	if err != nil {
		panic(err)
	}
	limiter := NewRedis(redis.NewClient(opts))

	t.Run("No Error (anonymous request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, Unlimited)
		handler := interceptor.HTTPHandler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "http://rilldata.com/foo", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("No Error (authenticated request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, Unlimited)
		handler := interceptor.HTTPHandler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "http://rilldata.com/foo", nil)
		req = req.WithContext(context.WithValue(context.Background(), "authorization", "userid"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %v, got %v", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Quota Exceeded Error (anonymous request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, redis_rate.PerMinute(1), Unlimited)
		handler := interceptor.HTTPHandler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

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

	t.Run("Quota Exceeded Error (authenticated request)", func(t *testing.T) {
		mr.FlushDB()
		interceptor := NewInterceptor(limiter, TestContextInspector{}, Unlimited, redis_rate.PerMinute(1))
		handler := interceptor.HTTPHandler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "http://rilldata.com/foo", nil)
		req = req.WithContext(context.WithValue(context.Background(), "authorization", "userid"))
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
