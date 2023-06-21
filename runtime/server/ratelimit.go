package server

import (
	"context"
	"github.com/go-redis/redis_rate/v10"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc"
)

type RuntimeUserCtxInspector struct{}

func (i RuntimeUserCtxInspector) IsAuthenticated(ctx context.Context) bool {
	return !auth.IsAnonymous(ctx)
}

func (i RuntimeUserCtxInspector) GetAuthID(ctx context.Context) string {
	return auth.GetClaims(ctx).Subject()
}

var ctxInspector ratelimit.CtxInspector = RuntimeUserCtxInspector{}

func limiterUnaryServerInterceptor(l ratelimit.Limiter, anonLimit, authLimit redis_rate.Limit) grpc.UnaryServerInterceptor {
	return ratelimit.NewInterceptor(l, ctxInspector, anonLimit, authLimit).UnaryServerInterceptor()
}

func limiterStreamServerInterceptor(l ratelimit.Limiter, anonLimit, authLimit redis_rate.Limit) grpc.StreamServerInterceptor {
	return ratelimit.NewInterceptor(l, ctxInspector, anonLimit, authLimit).StreamServerInterceptor()
}
