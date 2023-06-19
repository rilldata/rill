package server

import (
	"context"
	"github.com/go-redis/redis_rate/v10"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"google.golang.org/grpc"
	"net/http"
)

type AdminUserCtxInspector struct{}

func (i AdminUserCtxInspector) IsAuthenticated(ctx context.Context) bool {
	return !auth.IsAnonymous(ctx)
}

func (i AdminUserCtxInspector) GetAuthID(ctx context.Context) string {
	return auth.GetClaims(ctx).OwnerID()
}

var ctxInspector ratelimit.AuthContextInspector = AdminUserCtxInspector{}

func limiterUnaryServerInterceptor(l ratelimit.RequestRateLimiter, anonLimit, authLimit redis_rate.Limit) grpc.UnaryServerInterceptor {
	return ratelimit.NewInterceptor(l, ctxInspector, anonLimit, authLimit).UnaryServerInterceptor()
}

func limiterStreamServerInterceptor(l ratelimit.RequestRateLimiter, anonLimit, authLimit redis_rate.Limit) grpc.StreamServerInterceptor {
	return ratelimit.NewInterceptor(l, ctxInspector, anonLimit, authLimit).StreamServerInterceptor()
}

func LimiterHTTPHandler(route string, l ratelimit.RequestRateLimiter, anonLimit redis_rate.Limit, next http.Handler) http.Handler {
	return ratelimit.NewInterceptor(l, ctxInspector, anonLimit, ratelimit.Unlimited).HTTPHandler(route, next)
}
