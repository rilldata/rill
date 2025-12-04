package server

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/parser"
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
		attribute.Bool("args.use_format_tokens", req.UseFormatTokens),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Errorf(codes.FailedPrecondition, "does not have access to query data")
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	additionalWhereByMetricsView := map[string]map[string]any{}
	for mv, expr := range req.AdditionalWhereByMetricsView {
		additionalWhereByMetricsView[mv], err = metricsview.NewExpressionFromProto(expr).AsMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert additional where expression for metrics view %q: %w", mv, err)
		}
	}

	var additionalTimeRange map[string]any
	var timeZone string
	if req.AdditionalTimeRange != nil {
		additionalTimeRange, err = metricsview.NewTimeRangeFromProto(req.AdditionalTimeRange).AsMap()
		if err != nil {
			return nil, fmt.Errorf("failed to convert additional time range: %w", err)
		}
		timeZone = req.AdditionalTimeRange.TimeZone
	}

	templateData := parser.TemplateData{
		User:      claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
		ExtraFuncs: map[string]any{
			"metrics_sql": func(sql string) (any, error) {
				// Run metrics SQL resolver
				resolveRes, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
					InstanceID: req.InstanceId,
					Resolver:   "metrics_sql",
					ResolverProperties: map[string]any{
						"sql":                              sql,
						"time_zone":                        timeZone,
						"additional_where_by_metrics_view": additionalWhereByMetricsView,
						"additional_time_range":            additionalTimeRange,
					},
					Args:   nil,
					Claims: claims,
				})
				if err != nil {
					return "", err
				}
				defer resolveRes.Close()

				var rows []map[string]any
				for {
					row, err := resolveRes.Next()
					if err != nil {
						if err.Error() == "EOF" || err.Error() == "sql: no rows in result set" {
							break
						}
						return "", fmt.Errorf("failed to get result: %w", err)
					}
					rows = append(rows, row)
				}

				if len(rows) == 0 {
					return "", fmt.Errorf("metrics_sql query returned no results")
				}

				// Get metrics view from metadata for format tokens
				var mv string
				if meta := resolveRes.Meta(); meta != nil {
					mv, _ = meta["metrics_view"].(string)
				}

				if len(rows) == 1 && len(rows[0]) == 1 {
					var field string
					var val any
					for k, v := range rows[0] {
						field = k
						val = v
					}

					// Return value wrapped in a format token if requested
					if req.UseFormatTokens && mv != "" {
						return fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, mv, field, val), nil
					}
					return fmt.Sprintf("%v", val), nil
				}

				// Single row, multiple fields: return as map
				if len(rows) == 1 {
					return rows[0], nil
				}

				return rows, nil
			},
		},
	}

	// Resolve the template
	body, err := parser.ResolveTemplate(req.Body, templateData, false)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			return nil, err
		}
		return nil, status.Errorf(codes.InvalidArgument, "failed to resolve template: %s", err.Error())
	}

	return &runtimev1.ResolveTemplatedStringResponse{
		Body: body,
	}, nil
}
