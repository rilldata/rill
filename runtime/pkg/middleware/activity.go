package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ActivityStreamServerInterceptor(activityClient *activity.Client) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		claims := auth.GetClaims(ctx, "")
		ctx = setRequestActivityAttributes(ctx, claims, info.FullMethod)

		// Tag gRPC traffic as the "ui" source (dashboard / direct query API); chat handlers override it to "chat".
		ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceUI)

		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = ctx

		var code codes.Code
		start := time.Now()
		defer func() {
			emitRequestMetric(ctx, activityClient, claims, start, code.String())
		}()

		err := handler(srv, wss)
		code = extractCode(err)
		return err
	}
}

func ActivityUnaryServerInterceptor(activityClient *activity.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		claims := auth.GetClaims(ctx, "")
		ctx = setRequestActivityAttributes(ctx, claims, info.FullMethod)

		// Tag gRPC traffic as the "ui" source (dashboard / direct query API); chat handlers override it to "chat".
		ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceUI)

		var code codes.Code
		start := time.Now()
		defer func() {
			// Emitted after the handler, so org/project attributes added by the handler (via addInstanceRequestAttributes)
			// are present on the context and attached to the request event.
			emitRequestMetric(ctx, activityClient, claims, start, code.String())
		}()

		res, err := handler(ctx, req)
		code = extractCode(err)
		return res, err
	}
}

// ActivityHTTPMiddleware emits the request_time_ms usage metric for HTTP API requests that don't pass through the gRPC
// interceptor (e.g. the custom REST API and MCP). Together with the gRPC interceptors this gives one request event per
// API call across all surfaces, tagged with the given source. Billing logic (which sources count, distinct embedded
// users, etc.) is applied downstream in the metrics project.
func ActivityHTTPMiddleware(activityClient *activity.Client, source runtime.RequestSource) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := auth.GetClaims(r.Context(), "")
			ctx := setRequestActivityAttributes(r.Context(), claims, r.URL.Path)
			ctx = runtime.WithRequestSource(ctx, source)

			start := time.Now()
			defer func() {
				emitRequestMetric(ctx, activityClient, claims, start, "")
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setRequestActivityAttributes sets the user_id and request_method activity attributes on the context. For anonymous
// embedded users (no subject), it synthesizes a stable non-PII user id so distinct embedded users can be counted
// downstream. The user_id is only set when non-empty, to avoid overwriting an upstream-set user id (e.g. the CLI on
// localhost, where the user id is set at start time and individual requests are unauthenticated).
func setRequestActivityAttributes(ctx context.Context, claims *runtime.SecurityClaims, method string) context.Context {
	var userID string
	if claims != nil {
		userID = claims.UserID
		if userID == "" && isEmbed(claims.UserAttributes) {
			userID = anonymousUserID(claims.UserAttributes)
		}
	}

	if userID != "" {
		return activity.SetAttributes(ctx,
			attribute.String(activity.AttrKeyUserID, userID),
			attribute.String("request_method", method),
		)
	}
	return activity.SetAttributes(ctx, attribute.String("request_method", method))
}

// emitRequestMetric emits the request_time_ms usage metric with generic attributes (source, embed) so the metrics
// project can derive billable API calls and distinct embedded users downstream. code is the gRPC status (empty for HTTP).
func emitRequestMetric(ctx context.Context, activityClient *activity.Client, claims *runtime.SecurityClaims, start time.Time, code string) {
	embed := claims != nil && isEmbed(claims.UserAttributes)
	activityClient.RecordMetric(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()),
		attribute.String("grpc_code", code),
		attribute.String("source", string(runtime.RequestSourceFromContext(ctx))),
		attribute.Bool("embed", embed),
	)
}

func isEmbed(attrs map[string]any) bool {
	embed, _ := attrs["embed"].(bool)
	return embed
}

// anonymousUserID returns a deterministic, non-PII identifier for an anonymous embedded user,
// derived from their user attributes. Downstream billing counts distinct values of this identifier.
func anonymousUserID(attrs map[string]any) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%v\x00", k, attrs[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func extractCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	if s, ok := status.FromError(err); ok {
		return s.Code()
	}

	return codes.Unknown
}
