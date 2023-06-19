package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"reflect"
)

// LimiterInterceptor embeds the RequestRateLimiter and contains limits for authenticated and anonymous requests.
// It uses AuthContextInspector to classify a request as authenticated/anonymous.
// The interceptor provides mechanisms to handle rate limiting for three types of requests:
// unary RPC, streaming RPC and HTTP.
type LimiterInterceptor struct {
	RequestRateLimiter
	ctxInspector AuthContextInspector
	anonLimit    redis_rate.Limit
	authLimit    redis_rate.Limit
}

type AuthContextInspector interface {
	IsAuthenticated(ctx context.Context) bool
	GetAuthID(ctx context.Context) string
}

func NewInterceptor(l RequestRateLimiter, i AuthContextInspector, anonLimit, authLimit redis_rate.Limit) *LimiterInterceptor {
	return &LimiterInterceptor{
		RequestRateLimiter: l,
		ctxInspector:       i,
		authLimit:          authLimit,
		anonLimit:          anonLimit,
	}
}

func (i *LimiterInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := i.Limit(ctx, i.gRPCRequestLimitKeyByCtx(ctx, req), i.limitByCtx(ctx)); err != nil {
			if errors.As(err, &QuotaExceededError{}) {
				return nil, status.Errorf(codes.ResourceExhausted, err.Error())
			}
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (i *LimiterInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &wrappedServerStream{ss, i})
	}
}

func (i *LimiterInterceptor) HTTPHandler(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := i.Limit(ctx, i.httpRequestLimitKeyByCtx(ctx, r, route), i.limitByCtx(ctx)); err != nil {
			if errors.As(err, &QuotaExceededError{}) {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (i *LimiterInterceptor) limitByCtx(ctx context.Context) redis_rate.Limit {
	var limit redis_rate.Limit
	if i.ctxInspector.IsAuthenticated(ctx) {
		limit = i.authLimit
	} else {
		limit = i.anonLimit
	}
	return limit
}

func (i *LimiterInterceptor) httpRequestLimitKeyByCtx(ctx context.Context, req *http.Request, route string) string {
	if i.ctxInspector.IsAuthenticated(ctx) {
		return AuthReqLimitKey(route, i.ctxInspector.GetAuthID(ctx))
	}
	return AnonReqLimitKey(route, observability.HTTPPeer(req))
}

func (i *LimiterInterceptor) gRPCRequestLimitKeyByCtx(ctx context.Context, req interface{}) string {
	if i.ctxInspector.IsAuthenticated(ctx) {
		return AuthReqLimitKey(GRPCRequestName(req), i.ctxInspector.GetAuthID(ctx))
	}
	return AnonReqLimitKey(GRPCRequestName(req), observability.GrpcPeer(ctx))
}

func AuthReqLimitKey(reqName string, authID string) string {
	return fmt.Sprintf("auth:%s:%s", reqName, authID)
}

func AnonReqLimitKey(reqName string, ipAddr string) string {
	return fmt.Sprintf("anon:%s:%s", reqName, ipAddr)
}

func GRPCRequestName(req interface{}) string {
	if req == nil {
		return "unnamed"
	}
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

type wrappedServerStream struct {
	grpc.ServerStream
	i *LimiterInterceptor
}

func (wss *wrappedServerStream) RecvMsg(req interface{}) error {
	ctx := wss.Context()
	if err := wss.i.Limit(ctx, wss.i.gRPCRequestLimitKeyByCtx(ctx, req), wss.i.limitByCtx(ctx)); err != nil {
		if errors.As(err, &QuotaExceededError{}) {
			return status.Errorf(codes.ResourceExhausted, err.Error())
		}
		return err
	}

	return wss.ServerStream.RecvMsg(req)
}
