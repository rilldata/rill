package middleware

import (
	"context"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"google.golang.org/grpc"
	"net/http"
)

// This is a collection of gRPC and HTTP interceptors that call fn per request.

// RequestStreamServerInterceptor wraps a ServerStream and calls fn on each RecvMsg.
func RequestStreamServerInterceptor(fn func(info Metadata) error) grpc.StreamServerInterceptor {
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

		wss := wrappedServerStream{ss, *newMetadata(ss.Context(), method, observability.GrpcPeer(ss.Context())), fn}
		return handler(srv, &wss)
	}
}

// RequestUnaryServerInterceptor calls fn on each request
func RequestUnaryServerInterceptor(fn func(info Metadata) error) grpc.UnaryServerInterceptor {
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

		if err = fn(*newMetadata(ctx, method, observability.GrpcPeer(ctx))); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	info Metadata
	fn   func(info Metadata) error
}

func (wss *wrappedServerStream) RecvMsg(m interface{}) error {
	if err := wss.fn(wss.info); err != nil {
		return err
	}
	return wss.ServerStream.RecvMsg(m)
}

// RequestHTTPHandler calls fn on each request.
func RequestHTTPHandler(route string, fn func(info Metadata) error, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := fn(*newMetadata(ctx, route, observability.HTTPPeer(r))); err != nil {
			switch err := err.(type) {
			case *HTTPError:
				http.Error(w, err.Error(), err.Code)
			default:
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

func (h *HTTPError) Error() string {
	return h.Message
}

type Metadata struct {
	Ctx    context.Context
	Method string
	Peer   string
}

func newMetadata(ctx context.Context, method string, peer string) *Metadata {
	return &Metadata{Ctx: ctx, Method: method, Peer: peer}
}
