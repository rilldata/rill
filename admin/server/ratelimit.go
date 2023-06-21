package server

import (
	"github.com/go-redis/redis_rate/v10"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"google.golang.org/grpc"
	"net/http"
)

func limiterUnaryServerInterceptor(l ratelimit.Limiter, anonLimit, authLimit redis_rate.Limit) grpc.UnaryServerInterceptor {
	return ratelimit.NewInterceptor(l, auth.CtxInspector{}, anonLimit, authLimit).UnaryServerInterceptor()
}

func limiterStreamServerInterceptor(l ratelimit.Limiter, anonLimit, authLimit redis_rate.Limit) grpc.StreamServerInterceptor {
	return ratelimit.NewInterceptor(l, auth.CtxInspector{}, anonLimit, authLimit).StreamServerInterceptor()
}

func limiterHTTPHandler(route string, l ratelimit.Limiter, anonLimit redis_rate.Limit, next http.Handler) http.Handler {
	return ratelimit.NewInterceptor(l, auth.CtxInspector{}, anonLimit, ratelimit.Unlimited).HTTPHandler(route, next)
}
