package resolvers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestText_MetricsSQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model1.sql": `
SELECT 'US' AS country, DATE '2024-01-01' AS order_date, 100 AS revenue, 5 AS orders
UNION ALL
SELECT 'UK' AS country, DATE '2024-01-15' AS order_date, 200 AS revenue, 10 AS orders
UNION ALL
SELECT 'CA' AS country, DATE '2024-02-01' AS order_date, 150 AS revenue, 15 AS orders
UNION ALL
SELECT 'CA' AS country, DATE '2024-02-15' AS order_date, 250 AS revenue, 20 AS orders
`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: model1
timeseries: order_date
dimensions:
- column: country
- column: order_date
  name: order_date
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: total_orders
  expression: SUM(orders)
`,
			"model2.sql": `
SELECT 'Electronics' AS category, 500 AS sales
UNION ALL
SELECT 'Clothing' AS category, 250 AS sales
UNION ALL
SELECT 'Food' AS category, 150 AS sales
`,
			"mv2.yaml": `
type: metrics_view
version: 1
model: model2
dimensions:
- column: category
measures:
- name: total_sales
  expression: SUM(sales)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	tt := []struct {
		name        string
		props       map[string]any
		expected    []string
		expectEqual bool
	}{
		{
			name: "WithFormatTokens",
			props: map[string]any{
				"text":              `Total: {{ metrics_sql "SELECT total_revenue FROM mv1" }}`,
				"use_format_tokens": true,
			},
			expected: []string{`__RILL__FORMAT__({"metrics_view":"mv1","field":"total_revenue","value":700})`},
		},
		{
			name: "MultipleQueries",
			props: map[string]any{
				"text": "Revenue: {{ metrics_sql \"SELECT total_revenue FROM mv1\" }}\nOrders: {{ metrics_sql \"SELECT total_orders FROM mv1\" }}",
			},
			expected: []string{"Revenue: 700", "Orders: 50"},
		},
		{
			name: "MultipleMetricsViewsWithDifferentFilters",
			props: map[string]any{
				"text": `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}, Sales: {{ metrics_sql "SELECT total_sales FROM mv2" }}`,
				"additional_where_by_metrics_view": map[string]any{
					"mv1": map[string]any{
						"cond": map[string]any{
							"op": "in",
							"exprs": []any{
								map[string]any{"name": "country"},
								map[string]any{"val": []any{"US", "UK"}},
							},
						},
					},
					"mv2": map[string]any{
						"cond": map[string]any{
							"op": "eq",
							"exprs": []any{
								map[string]any{"name": "category"},
								map[string]any{"val": "Electronics"},
							},
						},
					},
				},
			},
			expected:    []string{"Revenue: 300, Sales: 500"},
			expectEqual: true,
		},
		{
			name: "WithAdditionalTimeRange",
			props: map[string]any{
				"text": `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}`,
				"additional_time_range": map[string]any{
					"start": "2024-01-01T00:00:00Z",
					"end":   "2024-02-01T00:00:00Z",
				},
			},
			expected:    []string{"Revenue: 300"},
			expectEqual: true,
		},
		{
			name: "WithRefs",
			props: map[string]any{
				"text": `Revenue: {{ metrics_sql "SELECT total_revenue FROM {{ ref \"mv1\" }}" }}`,
			},
			expected:    []string{"Revenue: 700"},
			expectEqual: true,
		},
		{
			name: "WithMultipleRefs",
			props: map[string]any{
				"text": `{{ metrics_sql "SELECT total_revenue FROM {{ ref \"mv1\" }}" }} and {{ metrics_sql "SELECT total_sales FROM {{ ref \"mv2\" }}" }}`,
			},
			expected:    []string{"700 and 900"},
			expectEqual: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := resolveText(t, rt, instanceID, tc.props)
			if tc.expectEqual {
				require.Equal(t, strings.Join(tc.expected, ""), body)
			} else {
				for _, exp := range tc.expected {
					require.Contains(t, body, exp)
				}
			}
		})
	}
}

func TestText_MetricsSQLRows(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"bids.sql": `
SELECT 'Google' AS advertiser_name, 1000 AS overall_spend, DATE '2024-01-01' AS bid_date
UNION ALL
SELECT 'Microsoft' AS advertiser_name, 2000 AS overall_spend, DATE '2024-01-02' AS bid_date
UNION ALL
SELECT 'Yahoo' AS advertiser_name, 1500 AS overall_spend, DATE '2024-01-03' AS bid_date
UNION ALL
SELECT 'Amazon' AS advertiser_name, 3000 AS overall_spend, DATE '2024-01-04' AS bid_date
UNION ALL
SELECT 'Apple' AS advertiser_name, 2500 AS overall_spend, DATE '2024-01-05' AS bid_date
`,
			"bids_metrics.yaml": `
type: metrics_view
version: 1
model: bids
timeseries: bid_date
dimensions:
- column: advertiser_name
measures:
- name: total_bids
  expression: COUNT(*)
- name: overall_spend
  expression: SUM(overall_spend)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	tt := []struct {
		name     string
		props    map[string]any
		expected []string
	}{
		{
			name: "SingleRowSingleField",
			props: map[string]any{
				"text": `Total: {{ metrics_sql "select total_bids from bids_metrics" }}`,
			},
			expected: []string{"Total: 5"},
		},
		{
			name: "MultipleRowsMultipleFields",
			props: map[string]any{
				"text": `{{ $data := metrics_sql_rows "select overall_spend, advertiser_name from bids_metrics order by advertiser_name limit 3" }}` +
					`{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}` + "\n" + `{{ end }}`,
			},
			expected: []string{
				"- Amazon: 3000",
				"- Apple: 2500",
				"- Google: 1000",
			},
		},
		{
			name: "MultipleRowsSingleField",
			props: map[string]any{
				"text": `{{ $data := metrics_sql_rows "select advertiser_name from bids_metrics order by advertiser_name limit 3" }}` +
					`{{ range $data }}- {{ .advertiser_name }}` + "\n" + `{{ end }}`,
			},
			expected: []string{
				"- Amazon",
				"- Apple",
				"- Google",
			},
		},
		{
			name: "ComplexTemplateWithConditional",
			props: map[string]any{
				"text": `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from bids_metrics order by overall_spend desc limit 3" }}` +
					"Top Spenders:\n" +
					`{{ range $data }}{{ if gt .overall_spend 2000.0 }}- **{{ .advertiser_name }}**: ${{ .overall_spend }}` + "\n" +
					`{{ end }}{{ end }}`,
			},
			expected: []string{
				"Top Spenders:",
				"- **Amazon**: $3000",
				"- **Apple**: $2500",
			},
		},
		{
			name: "NestedRefMultipleRows",
			props: map[string]any{
				"text": `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from {{ ref \"bids_metrics\" }} order by advertiser_name limit 2" }}` +
					`{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}` + "\n" + `{{ end }}`,
			},
			expected: []string{
				"- Amazon: 3000",
				"- Apple: 2500",
			},
		},
		{
			name: "MultipleRowsWithFormatTokens",
			props: map[string]any{
				"text": `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from bids_metrics order by advertiser_name limit 2" }}` +
					`{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}` + "\n" + `{{ end }}`,
				"use_format_tokens": true,
			},
			expected: []string{
				`__RILL__FORMAT__({"metrics_view":"bids_metrics","field":"overall_spend","value":3000})`,
				`__RILL__FORMAT__({"metrics_view":"bids_metrics","field":"overall_spend","value":2500})`,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			body := resolveText(t, rt, instanceID, tc.props)
			for _, exp := range tc.expected {
				require.Contains(t, body, exp, "Expected output to contain: %s\nFull output:\n%s", exp, body)
			}
		})
	}
}

// resolveText calls the text resolver and returns the resolved text string.
func resolveText(t *testing.T, rt *runtime.Runtime, instanceID string, props map[string]any) string {
	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "text",
		ResolverProperties: props,
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Len(t, rows, 1)

	text, ok := rows[0]["text"].(string)
	require.True(t, ok)
	return text
}
