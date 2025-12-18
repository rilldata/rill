package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/formatter"
)

// ResolveTemplatedStringOptions contains options for resolving templated strings with metrics_sql support.
type ResolveTemplatedStringOptions struct {
	InstanceID      string
	Claims          *SecurityClaims
	Body            string
	UseFormatTokens bool
}

// ResolveTemplatedStringWithMetricsSQL resolves a templated string that may contain metrics_sql functions.
// This is the shared implementation used by both the server RPC and reconcilers.
func (r *Runtime) ResolveTemplatedStringWithMetricsSQL(ctx context.Context, opts ResolveTemplatedStringOptions) (string, error) {
	inst, err := r.Instance(ctx, opts.InstanceID)
	if err != nil {
		return "", err
	}

	// Helper function to get measure formatters for a metrics view
	getMeasureFormatters := func(metricsView string) (map[string]formatter.Formatter, error) {
		if metricsView == "" {
			return nil, nil
		}

		ctrl, err := r.Controller(ctx, opts.InstanceID)
		if err != nil {
			return nil, err
		}

		mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: metricsView}, false)
		if err != nil {
			// Metrics view not found - return empty formatters
			return nil, nil
		}

		mvSpec := mv.GetMetricsView()
		if mvSpec == nil {
			return nil, nil
		}

		formatters := make(map[string]formatter.Formatter)
		for _, measure := range mvSpec.State.ValidSpec.Measures {
			if measure.FormatPreset != "" {
				f, err := formatter.NewPresetFormatter(measure.FormatPreset, false)
				if err != nil {
					// Log but continue - formatter error shouldn't break the whole report
					continue
				}
				formatters[measure.Name] = f
			}
		}

		return formatters, nil
	}

	// Helper function to resolve metrics SQL queries
	resolveMetricsSQL := func(sql string, unary bool) ([]map[string]any, error) {
		resolveRes, err := r.Resolve(ctx, &ResolveOptions{
			InstanceID: opts.InstanceID,
			Resolver:   "metrics_sql",
			ResolverProperties: map[string]any{
				"sql": sql,
			},
			Args:   nil,
			Claims: opts.Claims,
		})
		if err != nil {
			return nil, err
		}
		defer resolveRes.Close()

		// Get metrics view name from resolver metadata
		var metricsView string
		if meta := resolveRes.Meta(); meta != nil {
			metricsView, _ = meta["metrics_view"].(string)
		}

		// Get formatters for measures in this metrics view
		measureFormatters, err := getMeasureFormatters(metricsView)
		if err != nil {
			// Log but continue - formatter errors shouldn't break the report
			measureFormatters = nil
		}

		var rows []map[string]any
		for {
			row, err := resolveRes.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, fmt.Errorf("failed to get result: %w", err)
			}

			// Apply formatters to measure values
			if measureFormatters != nil {
				for field, val := range row {
					if formatter, ok := measureFormatters[field]; ok && val != nil {
						formatted, err := formatter.StringFormat(val)
						if err == nil {
							row[field] = formatted
						}
						// If formatting fails, keep original value
					}
				}
			}

			if len(rows) > 0 && unary {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, but the query returned multiple")
			}
			rows = append(rows, row)
		}

		if unary {
			if len(rows) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, got none")
			}
			if len(rows[0]) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating only allows one result field, got %d", len(rows[0]))
			}
		}

		return rows, nil
	}

	// Prepare template data with metrics_sql functions
	var userAttrs map[string]any
	if opts.Claims != nil {
		userAttrs = opts.Claims.UserAttributes
	}
	if userAttrs == nil {
		userAttrs = make(map[string]any)
	}

	templateData := parser.TemplateData{
		User:      userAttrs,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
		ExtraFuncs: map[string]any{
			"metrics_sql": func(sql string) (string, error) {
				rows, err := resolveMetricsSQL(sql, true)
				if err != nil {
					return "", err
				}
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
				return resolveMetricsSQL(sql, false)
			},
		},
	}

	// Resolve the template
	body, err := parser.ResolveTemplate(opts.Body, templateData, false)
	if err != nil {
		return "", fmt.Errorf("failed to resolve template: %w", err)
	}

	return body, nil
}
