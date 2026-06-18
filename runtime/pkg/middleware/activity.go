package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
		subject := auth.GetClaims(ctx, "").UserID

		// Only set the user ID attribute if it is not empty. This prevents overwriting a user ID attribute set upstream.
		// (For example, in the CLI on local, the user ID is set at start time, and individual requests on localhost are unauthenticated.)
		if subject != "" {
			ctx = activity.SetAttributes(ctx,
				attribute.String(activity.AttrKeyUserID, subject),
				attribute.String("request_method", info.FullMethod),
			)
		} else {
			ctx = activity.SetAttributes(ctx,
				attribute.String("request_method", info.FullMethod),
			)
		}

		// Tag interactive/gRPC traffic as the "ui" source so billable queries it triggers are not billed as programmatic.
		ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceUI)

		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = ctx

		var code codes.Code
		start := time.Now()
		defer func() {
			// Emit usage metric
			activityClient.RecordMetric(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()), attribute.String("grpc_code", code.String()))
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
		subject := claims.UserID

		// Only set the user ID attribute if it is not empty. This prevents overwriting a user ID attribute set upstream.
		// (For example, in the CLI on local, the user ID is set at start time, and individual requests on localhost are unauthenticated.)
		if subject != "" {
			ctx = activity.SetAttributes(ctx,
				attribute.String(activity.AttrKeyUserID, subject),
				attribute.String("request_method", info.FullMethod),
			)
		} else {
			ctx = activity.SetAttributes(ctx,
				attribute.String("request_method", info.FullMethod),
			)
		}

		// Tag interactive/gRPC traffic as the "ui" source so billable queries it triggers are not billed as programmatic.
		ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceUI)

		var code codes.Code
		start := time.Now()
		defer func() {
			// Emit usage metrics. This runs after the handler, so org/project attributes added by the handler (via
			// addInstanceRequestAttributes) are present on the context and attached to the emitted events.
			activityClient.RecordMetric(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()), attribute.String("grpc_code", code.String()))
			recordEmbeddedUsage(ctx, activityClient, claims)
		}()

		res, err := handler(ctx, req)
		code = extractCode(err)
		return res, err
	}
}

// recordEmbeddedUsage emits a generic embedded_user_request metric for requests made by embedded users (those with
// the "embed" attribute set on their token), carrying a user_id so the metrics project can derive distinct active
// embedded users. Classification (external vs anonymous) is left to the metrics project via the user_id (external
// users have an ext_-prefixed id). Org/project attributes come from the context (set by the instance-specific handlers).
func recordEmbeddedUsage(ctx context.Context, activityClient *activity.Client, claims *runtime.SecurityClaims) {
	if claims == nil || !isEmbed(claims.UserAttributes) {
		return
	}
	activityClient.RecordMetric(ctx, "embedded_user_request", 1, attribute.String(activity.AttrKeyUserID, getEmbedUserID(claims)))
}

// getEmbedUserID returns the identifier for an embedded user: the external user subject (ext_-prefixed) if an
// external_user_id was provided, otherwise a non-PII hash of the user attributes (for anonymous embeds).
func getEmbedUserID(claims *runtime.SecurityClaims) string {
	if claims.UserID != "" {
		return claims.UserID
	}
	return anonymousUserID(claims.UserAttributes)
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
