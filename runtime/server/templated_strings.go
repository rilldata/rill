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
					// The "metrics" resolver returns the metrics view in the metadata.
					// (This is a bit of a hacky way to pass this info along, but it avoids turning format tokens into a deeper concept.)
					var mv string
					var isDimension bool
					if meta := resolveRes.Meta(); meta != nil {
						mv, _ = meta["metrics_view"].(string)
						isDimension = isFieldDimension(meta, field)
					}
					if mv != "" && !isDimension {
						return fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, mv, field, val), nil
					}
					// Fallthrough to raw value if we can't find the metrics view or if it's a dimension
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
						if err.Error() == "EOF" || err.Error() == "sql: no rows in result set" {
							break
						}
						return nil, fmt.Errorf("failed to get result: %w", err)
					}
					rows = append(rows, row)
				}

				// Handle empty result
				if len(rows) == 0 {
					return nil, fmt.Errorf("metrics_sql_rows query returned no results")
				}

				// Get metrics view from metadata for format tokens
				var mv string
				var fieldTypes map[string]string
				if meta := resolveRes.Meta(); meta != nil {
					mv, _ = meta["metrics_view"].(string)
					fieldTypes = buildFieldTypeMap(meta)
				}

				// Apply format tokens if requested
				if req.UseFormatTokens && mv != "" {
					formattedRows := make([]map[string]any, len(rows))
					for i, row := range rows {
						formattedRow := make(map[string]any)
						for field, val := range row {
							// Only apply formatting if the field is not a dimension
							if fieldTypes != nil && fieldTypes[field] == "dimension" {
								formattedRow[field] = val
							} else {
								formattedRow[field] = fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, mv, field, val)
							}
						}
						formattedRows[i] = formattedRow
					}
					return formattedRows, nil
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

// isFieldDimension checks if a field is a dimension by looking it up in the metadata fields.
func isFieldDimension(meta map[string]any, fieldName string) bool {
	fields := getFieldsFromMeta(meta)
	if fields == nil {
		return false
	}
	for _, f := range fields {
		name, _ := f["name"].(string)
		fieldType, _ := f["type"].(string)
		if name == fieldName && fieldType == "dimension" {
			return true
		}
	}
	return false
}

// buildFieldTypeMap builds a map of field names to their types from metadata.
func buildFieldTypeMap(meta map[string]any) map[string]string {
	fieldTypes := make(map[string]string)
	fields := getFieldsFromMeta(meta)
	if fields == nil {
		return fieldTypes
	}
	for _, f := range fields {
		name, _ := f["name"].(string)
		fieldType, _ := f["type"].(string)
		if name != "" && fieldType != "" {
			fieldTypes[name] = fieldType
		}
	}
	return fieldTypes
}

// getFieldsFromMeta extracts the fields array from metadata, handling different type assertions.
func getFieldsFromMeta(meta map[string]any) []map[string]any {
	fieldsVal, ok := meta["fields"]
	if !ok {
		return nil
	}

	// Try asserting as []map[string]any first (direct type)
	if fields, ok := fieldsVal.([]map[string]any); ok {
		return fields
	}

	// Try asserting as []any and then convert each element
	if fieldsAny, ok := fieldsVal.([]any); ok {
		fields := make([]map[string]any, 0, len(fieldsAny))
		for _, f := range fieldsAny {
			if fieldMap, ok := f.(map[string]any); ok {
				fields = append(fields, fieldMap)
			}
		}
		return fields
	}

	return nil
}
