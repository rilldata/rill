package reconcilers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractMetricsViewsFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single metrics_sql call",
			content:  `{{ metrics_sql "SELECT total_revenue FROM sales_metrics" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name: "multiple metrics_sql calls",
			content: `
Total revenue: {{ metrics_sql "SELECT total_revenue FROM sales_metrics" }}
Total orders: {{ metrics_sql "SELECT total_orders FROM order_metrics" }}
`,
			expected: []string{"sales_metrics", "order_metrics"},
		},
		{
			name:     "metrics_sql with WHERE clause",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics WHERE country = 'US'" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "metrics_sql with GROUP BY",
			content:  `{{ metrics_sql "SELECT country, revenue FROM sales_metrics WHERE active = true GROUP BY country" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "metrics_sql with ORDER BY",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics ORDER BY revenue DESC" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "metrics_sql with LIMIT",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics LIMIT 10" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "metrics_sql with HAVING",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics HAVING revenue > 100" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "no metrics_sql calls",
			content:  `# Just some markdown\nNo template functions here.`,
			expected: []string{},
		},
		{
			name:     "other template functions ignored",
			content:  `{{ upper "hello" }} {{ metrics_sql "SELECT revenue FROM sales_metrics" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name: "duplicate metrics views",
			content: `
{{ metrics_sql "SELECT revenue FROM sales_metrics" }}
{{ metrics_sql "SELECT orders FROM sales_metrics" }}
`,
			expected: []string{"sales_metrics"},
		},
		{
			name:     "malformed template (no closing quote)",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics }}`,
			expected: []string{},
		},
		{
			name:     "malformed template (no FROM clause)",
			content:  `{{ metrics_sql "SELECT revenue" }}`,
			expected: []string{},
		},
		{
			name:     "complex SQL query",
			content:  `{{ metrics_sql "SELECT SUM(revenue) FROM sales_metrics WHERE country IN ('US', 'UK') AND active = true GROUP BY country HAVING SUM(revenue) > 1000 ORDER BY SUM(revenue) DESC LIMIT 10" }}`,
			expected: []string{"sales_metrics"},
		},
		{
			name: "markdown with mixed content",
			content: `# Sales Report

Total revenue: {{ metrics_sql "SELECT total_revenue FROM sales_metrics" }}

## By Country
{{ metrics_sql "SELECT country, revenue FROM country_metrics WHERE active = true" }}

_Last updated: {{ now }}_`,
			expected: []string{"sales_metrics", "country_metrics"},
		},
		{
			name:     "metrics view name with underscores",
			content:  `{{ metrics_sql "SELECT revenue FROM sales_metrics_v2" }}`,
			expected: []string{"sales_metrics_v2"},
		},
		{
			name:     "metrics view name with numbers",
			content:  `{{ metrics_sql "SELECT revenue FROM metrics2024" }}`,
			expected: []string{"metrics2024"},
		},
		{
			name:     "metrics_sql with whitespace variations",
			content:  `{{metrics_sql "SELECT revenue FROM sales_metrics"}}`,
			expected: []string{"sales_metrics"},
		},
		{
			name: "multiple references same metrics view different queries",
			content: `
Revenue: {{ metrics_sql "SELECT SUM(revenue) FROM sales_metrics" }}
Count: {{ metrics_sql "SELECT COUNT(*) FROM sales_metrics" }}
`,
			expected: []string{"sales_metrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMetricsViewsFromTemplate(tt.content)
			require.Equal(t, tt.expected, result)
		})
	}
}
