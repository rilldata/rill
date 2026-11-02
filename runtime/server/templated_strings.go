package server

import (
	"context"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ResolveTemplatedString(ctx context.Context, req *runtimev1.ResolveTemplatedStringRequest) (*runtimev1.ResolveTemplatedStringResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.component", req.Component),
		attribute.Bool("args.use_format_tokens", req.UseFormatTokens),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Error(codes.PermissionDenied, "does not have access to query data")
	}

	// The template is either the saved content of a markdown component or the request body.
	// A component's content is trusted if it was declared in the project's source files, so its templates may access project variables.
	// A body comes from the caller and is not trusted, so its templates cannot.
	text := req.Body
	trusted := false
	if req.Component != "" {
		var err error
		text, trusted, err = s.markdownContent(ctx, req.InstanceId, req.Component, claims)
		if err != nil {
			return nil, err
		}
	}

	additionalWhereByMetricsView := map[string]map[string]any{}
	for mv, expr := range req.AdditionalWhereByMetricsView {
		var err error
		additionalWhereByMetricsView[mv], err = metricsview.NewExpressionFromProto(expr).AsMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert additional where expression for metrics view %q: %w", mv, err)
		}
	}

	var additionalTimeRange map[string]any
	var timeZone string
	if req.AdditionalTimeRange != nil {
		var err error
		additionalTimeRange, err = metricsview.NewTimeRangeFromProto(req.AdditionalTimeRange).AsMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert additional time range: %w", err)
		}
		timeZone = req.AdditionalTimeRange.TimeZone
	}

	resolveRes, _, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: req.InstanceId,
		Resolver:   "text",
		ResolverProperties: map[string]any{
			"text":                             text,
			"use_format_tokens":                req.UseFormatTokens,
			"additional_where_by_metrics_view": additionalWhereByMetricsView,
			"additional_time_range":            additionalTimeRange,
			"time_zone":                        timeZone,
		},
		TrustedProperties: trusted,
		Claims:            claims,
	})
	if err != nil {
		return nil, mapGRPCErrorWithFallback(err, codes.InvalidArgument)
	}
	defer resolveRes.Close()

	row, err := resolveRes.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, status.Errorf(codes.Internal, "text resolver returned no rows")
		}
		return nil, err
	}

	body, _ := row["text"].(string)
	return &runtimev1.ResolveTemplatedStringResponse{
		Body: body,
	}, nil
}

// markdownContent returns the content of a markdown component, after checking that the caller may access the component.
// It also reports whether the content is trusted: a component of an admin-managed canvas (e.g. a personal canvas) was written by the canvas's creator,
// not declared in the project's source files, so its templates must not access project variables.
func (s *Server) markdownContent(ctx context.Context, instanceID, component string, claims *runtime.SecurityClaims) (string, bool, error) {
	res, canvasRes, err := s.accessibleComponent(ctx, instanceID, component, claims)
	if err != nil {
		return "", false, err
	}
	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		return "", false, status.Errorf(codes.FailedPrecondition, "component %q is invalid", component)
	}
	if spec.Renderer != "markdown" {
		return "", false, status.Errorf(codes.InvalidArgument, "component %q is not a markdown component", component)
	}
	content, _ := spec.RendererProperties.AsMap()["content"].(string)
	trusted := canvasRes == nil || !runtime.IsAdminManaged(canvasRes.GetCanvas().State.ValidSpec.Annotations)
	return content, trusted, nil
}
