package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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

// externalUserSubjectPrefix is the prefix of the JWT subject assigned to embedded external users.
// Keep in sync with admin/server/deployment.go::subjectForExternalUser.
const externalUserSubjectPrefix = "ext_"

// InstanceAttributesFunc resolves the activity attributes (org_id, project_id, etc.) for an instance.
// It is used to attribute billable usage metrics to the right organization and project.
type InstanceAttributesFunc func(ctx context.Context, instanceID string) []attribute.KeyValue

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

func ActivityUnaryServerInterceptor(activityClient *activity.Client, instanceAttrs InstanceAttributesFunc) grpc.UnaryServerInterceptor {
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

		// Emit billable embedded-user API-call metrics (best-effort).
		recordEmbeddedUserAPICall(ctx, activityClient, instanceAttrs, claims, req)

		var code codes.Code
		start := time.Now()
		defer func() {
			// Emit usage metric
			activityClient.RecordMetric(ctx, "request_time_ms", float64(time.Since(start).Milliseconds()), attribute.String("grpc_code", code.String()))
		}()

		res, err := handler(ctx, req)
		code = extractCode(err)
		return res, err
	}
}

// recordEmbeddedUserAPICall emits a billable API-call metric for requests made by embedded users. Both metrics
// carry the user identity in the user_id attribute so the metrics project can derive distinct active users:
//   - external_user_api_call: the request is authenticated for an embedded external user (an external_user_id was
//     passed). The user_id attribute (set above) carries the "ext_"-prefixed hashed external user ID.
//   - external_anonymous_user_api_call: an embedded request with no external_user_id (and no Rill user). The
//     user_id attribute carries an "anon_"-prefixed hash of the user attributes.
//
// Both are per-request counters. It is a no-op for regular dashboard, API, and owner-preview traffic. The instance
// attributes (org_id, project_id) are attached so the usage can be attributed to the right organization and project.
func recordEmbeddedUserAPICall(ctx context.Context, activityClient *activity.Client, instanceAttrs InstanceAttributesFunc, claims *runtime.SecurityClaims, req interface{}) {
	if claims == nil || instanceAttrs == nil {
		return
	}

	var metricName string
	var idAttr attribute.KeyValue
	switch {
	case strings.HasPrefix(claims.UserID, externalUserSubjectPrefix):
		// The user_id attribute (set above) already carries the hashed external user ID for distinct counting.
		metricName = "external_user_api_call"
	case claims.UserID == "" && isEmbed(claims.UserAttributes):
		metricName = "external_anonymous_user_api_call"
		idAttr = attribute.String(activity.AttrKeyUserID, "anon_"+anonymousUserID(claims.UserAttributes))
	default:
		return
	}

	// The instance ID is needed to attribute usage to the org and project.
	r, ok := req.(interface{ GetInstanceId() string })
	if !ok {
		return
	}
	instanceID := r.GetInstanceId()
	if instanceID == "" {
		return
	}

	attrs := instanceAttrs(ctx, instanceID)
	if idAttr.Valid() {
		attrs = append(attrs, idAttr)
	}
	activityClient.RecordMetric(ctx, metricName, 1, attrs...)
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
