package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

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
			"metrics_sql": func(sql string) (string, error) {
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

				// Get only column in the only row
				row, err := resolveRes.Next()
				if err != nil {
					return "", fmt.Errorf("failed to get result: %w", err)
				}
				if len(row) != 1 {
					return "", fmt.Errorf("metrics_sql in templating only allows one result field, got %d", len(row))
				}
				_, err = resolveRes.Next()
				if err == nil {
					return "", fmt.Errorf("metrics_sql in templating must return one row, but the query returned multiple")
				}
				var field string
				var val any
				for k, v := range row {
					field = k
					val = v
				}

				// Return value wrapped in a format token if requested
				if req.UseFormatTokens {
					if meta := resolveRes.Meta(); meta != nil {
						mv, _ := meta["metrics_view"].(string)
						if mv != "" {
							if dims, err := s.metricsViewDimensions(ctx, req.InstanceId, mv); err == nil {
								if !dims[field] {
									payload := tokenPayload{MetricsView: mv, Field: field, Value: val}
									if b, err := json.Marshal(payload); err == nil {
										return fmt.Sprintf("__RILL__FORMAT__(%s)", string(b)), nil
									}
								}
							}
						}
					}
				}

				// Return stringified raw value
				return fmt.Sprintf("%v", val), nil
			},
			"metrics_sql_rows": func(sql string) (any, error) {
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
					return nil, err
				}
				defer resolveRes.Close()

				var rows []map[string]any
				for {
					row, err := resolveRes.Next()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return nil, fmt.Errorf("failed to get result: %w", err)
					}
					rows = append(rows, row)
				}

				// Get metrics view from metadata for format tokens
				var mv string
				var fieldTypes map[string]string
				if meta := resolveRes.Meta(); meta != nil {
					mv, _ = meta["metrics_view"].(string)
					if mv != "" {
						if dims, err := s.metricsViewDimensions(ctx, req.InstanceId, mv); err == nil {
							fieldTypes = make(map[string]string, len(dims))
							for k, v := range dims {
								if v {
									fieldTypes[k] = "dimension"
								}
							}
						}
					}
				}
				if !req.UseFormatTokens || mv == "" {
					return rows, nil
				}

				dims, _ := s.metricsViewDimensions(ctx, req.InstanceId, mv)
				formattedRows := make([]map[string]any, len(rows))
				for i, row := range rows {
					formattedRow := make(map[string]any, len(row))
					for field, val := range row {
						if dims != nil && dims[field] {
							formattedRow[field] = val
						} else {
							payload := tokenPayload{MetricsView: mv, Field: field, Value: val}
							if b, err := json.Marshal(payload); err == nil {
								formattedRow[field] = fmt.Sprintf("__RILL__FORMAT__(%s)", string(b))
							}
						}
					}
					formattedRows[i] = formattedRow
				}
				return formattedRows, nil
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

// metricsViewDimensions retrieves the metrics view resource for the given instance and metrics view name
func (s *Server) metricsViewDimensions(ctx context.Context, instanceID, metricsView string) (map[string]bool, error) {
	if metricsView == "" {
		return nil, nil
	}

	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := r.GetMetricsView()
	if mv == nil || mv.State == nil || mv.State.ValidSpec == nil {
		return nil, fmt.Errorf("metrics view %q has no valid spec", metricsView)
	}

	res := make(map[string]bool)
	for _, d := range mv.State.ValidSpec.Dimensions {
		if d != nil && d.Name != "" {
			res[d.Name] = true
		}
	}

	if mv.State.ValidSpec.TimeDimension != "" {
		res[mv.State.ValidSpec.TimeDimension] = true
	}

	return res, nil
}

type tokenPayload struct {
	MetricsView string `json:"metrics_view"`
	Field       string `json:"field"`
	Value       any    `json:"value"`
}
