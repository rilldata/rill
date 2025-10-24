package markdown

import (
	"context"
	"fmt"
	"regexp"
	"text/template"
)

// customFunctions returns custom template functions for markdown rendering
func (r *Renderer) customFunctions(ctx context.Context, renderCtx RenderContext) template.FuncMap {
	return template.FuncMap{
		// metrics_sql executes a Metrics SQL query and returns results
		// For single value: {{ metrics_sql "select requests from auction_metrics" }} -> returns formatted token
		// For multiple rows: {{ $data := metrics_sql "select name, revenue from products limit 5" }} -> returns array with tokens
		"metrics_sql": func(sql string) (any, error) {
			rows, err := r.executeQuery(ctx, sql, renderCtx)
			if err != nil {
				return nil, err
			}

			// Extract metrics view from SQL for formatting tokens
			metricsView := extractMetricsView(sql)

			// If single row with single column, return formatted token
			if len(rows) == 1 && len(rows[0]) == 1 {
				for col, v := range rows[0] {
					if metricsView != "" {
						return formatToken(metricsView, col, v), nil
					}
					return v, nil
				}
			}

			// For multi-row or multi-column results, wrap numeric values in format tokens
			if metricsView != "" {
				for i := range rows {
					for col, val := range rows[i] {
						// Only wrap numeric values
						if isNumeric(val) {
							rows[i][col] = formatToken(metricsView, col, val)
						}
					}
				}
			}

			return rows, nil
		},

		// first gets the first value from a result set
		// Usage: {{ $data := metrics_sql "select revenue, orders from sales limit 1" }}{{ $data | first "revenue" }}
		"first": func(key string, rows []map[string]any) any {
			if len(rows) > 0 {
				return rows[0][key]
			}
			return nil
		},
	}
}

// extractMetricsView extracts the metrics view name from a SQL query
func extractMetricsView(sql string) string {
	re := regexp.MustCompile(`(?i)from\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// formatToken creates a format token for the frontend to parse
// Format: __RILL_FORMAT__metricsview_measure_value__END__
func formatToken(metricsView, measureOrDim string, value any) string {
	return fmt.Sprintf("__RILL_FORMAT__%s_%s_%v__END__", metricsView, measureOrDim, value)
}

// isNumeric checks if a value is numeric
func isNumeric(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}
