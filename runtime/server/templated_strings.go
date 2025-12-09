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

	// Cache and getter function for metrics view measures.
	//
	// NOTE: It doesn't have to enforce field access security policies because those are already enforced in the metrics SQL resolver.
	// So any field returned by the resolver is already allowed to be accessed by the user.
	measuresCache := make(map[string]map[string]bool)
	isMeasure := func(metricsView, measure string) (bool, error) {
		if metricsView == "" {
			return false, nil
		}
		measures, ok := measuresCache[metricsView]
		if !ok {
			var err error
			measures, err = s.metricsViewMeasures(ctx, req.InstanceId, metricsView)
			if err != nil {
				return false, err
			}
			measuresCache[metricsView] = measures
		}
		_, ok = measures[measure]
		return ok, nil
	}

	// Utility function for resolving metrics SQL and handling req.UseFormatTokens.
	// If unary is true, it expects the query to return exactly one row with one column.
	resolveMetricsSQL := func(sql string, unary bool) ([]map[string]any, error) {
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
			return nil, err
		}
		defer resolveRes.Close()

		// Read all rows (or just one if unary)
		var rows []map[string]any
		for {
			row, err := resolveRes.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, fmt.Errorf("failed to get result: %w", err)
			}
			if len(rows) > 0 && unary {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, but the query returned multiple")
			}
			rows = append(rows, row)
		}

		// If unary, validate we got exactly one row with one column
		if unary {
			if len(rows) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, got none")
			}
			if len(rows[0]) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating only allows one result field, got %d", len(rows[0]))
			}
		}

		// When using format tokens, wrap each measure value with a format token (only measures, not dimensions).
		if req.UseFormatTokens {
			// The "metrics" resolver returns the metrics view in the metadata.
			// (This is a bit of a hacky way to pass this info along, but it avoids turning format tokens into a deeper concept.)
			var mv string
			if meta := resolveRes.Meta(); meta != nil {
				mv, _ = meta["metrics_view"].(string)
			}

			for _, row := range rows {
				for field, val := range row {
					// Skip if it's not a measure
					ok, err := isMeasure(mv, field)
					if err != nil {
						return nil, err
					}
					if !ok {
						continue
					}

					// Wrap in format token
					data, err := json.Marshal(resolveTemplatedStringFormatToken{MetricsView: mv, Field: field, Value: val})
					if err != nil {
						return nil, fmt.Errorf("failed to marshal measure value %v as JSON: %w", val, err)
					}
					row[field] = fmt.Sprintf("__RILL__FORMAT__(%s)", string(data))
				}
			}
		}

		// Return the rows
		return rows, nil
	}

	// Prepare template data.
	// We add two extra functions:
	// - metrics_sql: executes a metrics SQL query and returns the first field of the first row as a string.
	// - metrics_sql_rows: executes a metrics SQL query and returns all rows as a []map[string]any.
	templateData := parser.TemplateData{
		User:      claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
		ExtraFuncs: map[string]any{
			"metrics_sql": func(sql string) (string, error) {
				// Resolve with unary=true
				rows, err := resolveMetricsSQL(sql, true)
				if err != nil {
					return "", err
				}

				// Guaranteed to have one row with one column.
				if len(rows) > 0 {
					for _, val := range rows[0] {
						if val, ok := val.(string); ok {
							return val, nil
						}
						return fmt.Sprintf("%v", val), nil
					}
				}
				return "", fmt.Errorf("unreachable: no value in single-column single-row result")
			},
			"metrics_sql_rows": func(sql string) (any, error) {
				// Resolve with unary=false.
				// We can return the rows as-is.
				return resolveMetricsSQL(sql, false)
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

// metricsViewMeasures retrieves the measure names for a metrics view.
func (s *Server) metricsViewMeasures(ctx context.Context, instanceID, metricsView string) (map[string]bool, error) {
	if metricsView == "" {
		return nil, nil
	}

	_, mv, err := lookupMetricsView(ctx, s.runtime, instanceID, metricsView)
	if err != nil {
		return nil, err
	}

	spec := mv.ValidSpec
	measures := make(map[string]bool)
	for _, d := range spec.Measures {
		measures[d.Name] = true
	}

	return measures, nil
}

// resolveTemplatedStringFormatToken is the payload inside a __RILL__FORMAT__(...) token generated by ResolveTemplatedString.
type resolveTemplatedStringFormatToken struct {
	MetricsView string `json:"metrics_view"`
	Field       string `json:"field"`
	Value       any    `json:"value"`
}
