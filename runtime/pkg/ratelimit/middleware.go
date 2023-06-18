package ratelimit

import (
	"context"
	"errors"
	"github.com/go-redis/redis_rate/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func (l *RequestRateLimiter) Middleware() *RequestRateLimiterInterceptor {
	return &RequestRateLimiterInterceptor{
		RequestRateLimiter: *l,
		authLimit:          Unlimited,
		anonLimit:          Unlimited,
	}
}

type RequestRateLimiterInterceptor struct {
	RequestRateLimiter
	authLimit redis_rate.Limit
	anonLimit redis_rate.Limit
}

func (l *RequestRateLimiterInterceptor) WithAuthLimit(limit redis_rate.Limit) *RequestRateLimiterInterceptor {
	return &RequestRateLimiterInterceptor{
		RequestRateLimiter: l.RequestRateLimiter,
		authLimit:          limit,
		anonLimit:          l.anonLimit,
	}
}

func (l *RequestRateLimiterInterceptor) WithAnonLimit(limit redis_rate.Limit) *RequestRateLimiterInterceptor {
	return &RequestRateLimiterInterceptor{
		RequestRateLimiter: l.RequestRateLimiter,
		authLimit:          l.authLimit,
		anonLimit:          limit,
	}
}

func (l *RequestRateLimiterInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		err := l.LimitRequest(ctx, req, l.authLimit, l.anonLimit)

		if errors.As(err, &QuotaExceededError{}) {
			return nil, status.Errorf(codes.ResourceExhausted, err.Error())
		}

		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (l *RequestRateLimiterInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		w := &wrappedRequestLimitingStream{ss, nil}
		recvMsg := func(m interface{}) error {
			err := l.LimitRequest(w.Context(), m, l.authLimit, l.anonLimit)

			if errors.As(err, &QuotaExceededError{}) {
				return status.Errorf(codes.ResourceExhausted, err.Error())
			}

			if err != nil {
				return err
			}

			// Call the original RecvMsg.
			return w.ServerStream.RecvMsg(m)
		}
		w.recvMsg = recvMsg
		return handler(srv, w)
	}
}

func (l *RequestRateLimiterInterceptor) HTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := l.LimitRequest(r.Context(), r, l.authLimit, l.anonLimit)

		if errors.As(err, &QuotaExceededError{}) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type wrappedRequestLimitingStream struct {
	grpc.ServerStream
	recvMsg func(m interface{}) error
}

func (w *wrappedRequestLimitingStream) RecvMsg(m interface{}) error {
	return w.recvMsg(m)
}
