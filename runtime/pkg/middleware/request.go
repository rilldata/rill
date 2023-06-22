package middleware

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"net/http"

	"github.com/rilldata/rill/runtime/pkg/observability"
)

// This is a collection of gRPC and HTTP interceptors that call fn per request.

// RequestStreamServerInterceptor wraps a ServerStream and calls fn on each RecvMsg.
func RequestStreamServerInterceptor(fn func(Metadata) error) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		_, method, err := parseFullMethod(info.FullMethod)
		if err != nil {
			return err
		}

		wss := wrappedServerStream{ss, method, observability.GrpcPeer(ss.Context()), fn}
		return handler(srv, &wss)
	}
}

// RequestUnaryServerInterceptor calls fn on each request
func RequestUnaryServerInterceptor(fn func(Metadata) error) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		_, method, err := parseFullMethod(info.FullMethod)
		if err != nil {
			return nil, err
		}

		md := Metadata{ctx, req, method, observability.GrpcPeer(ctx)}
		if err = fn(md); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	method string
	peer   string
	fn     func(Metadata) error
}

func (wss *wrappedServerStream) RecvMsg(m interface{}) error {
	md := Metadata{wss.Context(), m, wss.method, wss.peer}
	if err := wss.fn(md); err != nil {
		return err
	}
	return wss.ServerStream.RecvMsg(m)
}

// RequestHTTPHandler calls fn on each request.
func RequestHTTPHandler(route string, fn func(Metadata) error, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md := Metadata{r.Context(), r, route, observability.HTTPPeer(r)}
		if err := fn(md); err != nil {
			var httpError *HTTPError
			if errors.As(err, &httpError) {
				http.Error(w, err.Error(), httpError.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

type HTTPError struct {
	Code    int
	Message string
}

func NewHTTPError(code int, msg string) *HTTPError {
	return &HTTPError{code, msg}
}

func (e *HTTPError) Error() string {
	return e.Message
}

type Metadata struct {
	Ctx    context.Context
	Req    interface{}
	Method string
	Peer   string
}
