package server_test

import (
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveTemplatedString_MetricsSQL(t *testing.T) {
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

	server, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	tt := []struct {
		name                string
		body                string
		useFormatTokens     bool
		additionalWhere     map[string]*runtimev1.Expression
		additionalTimeRange *runtimev1.TimeRange
		expected            []string
		expectEqual         bool
	}{
		{
			name:            "WithFormatTokens",
			body:            `Total: {{ metrics_sql "SELECT total_revenue FROM mv1" }}`,
			useFormatTokens: true,
			additionalWhere: nil,
			expected:        []string{`__RILL__FORMAT__({"metrics_view":"mv1","field":"total_revenue","value":700})`},
			expectEqual:     false,
		},
		{
			name: "MultipleQueries",
			body: `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}
		Orders: {{ metrics_sql "SELECT total_orders FROM mv1" }}`,
			useFormatTokens: false,
			additionalWhere: nil,
			expected:        []string{"Revenue: 700", "Orders: 50"},
			expectEqual:     false,
		},
		{
			name:            "MultipleMetricsViewsWithDifferentFilters",
			body:            `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}, Sales: {{ metrics_sql "SELECT total_sales FROM mv2" }}`,
			useFormatTokens: false,
			additionalWhere: map[string]*runtimev1.Expression{
				"mv1": {
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_IN,
							Exprs: []*runtimev1.Expression{
								{Expression: &runtimev1.Expression_Ident{Ident: "country"}},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewListValue(&structpb.ListValue{
											Values: []*structpb.Value{
												structpb.NewStringValue("US"),
												structpb.NewStringValue("UK"),
											},
										}),
									},
								},
							},
						},
					},
				},
				"mv2": {
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{Expression: &runtimev1.Expression_Ident{Ident: "category"}},
								{Expression: &runtimev1.Expression_Val{Val: structpb.NewStringValue("Electronics")}},
							},
						},
					},
				},
			},
			expected:    []string{"Revenue: 300, Sales: 500"},
			expectEqual: true,
		},
		{
			name:            "WithAdditionalTimeRange",
			body:            `Revenue: {{ metrics_sql "SELECT total_revenue FROM mv1" }}`,
			useFormatTokens: false,
			additionalWhere: nil,
			additionalTimeRange: &runtimev1.TimeRange{
				Start: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				End:   timestamppb.New(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			},
			expected:    []string{"Revenue: 300"},
			expectEqual: true,
		},
		{
			name:            "WithRefs",
			body:            `Revenue in US and UK from 2024-01-01 to 2024-02-01: {{ metrics_sql "SELECT total_revenue FROM {{ ref \"mv1\" }}" }}`,
			useFormatTokens: false,
			additionalWhere: map[string]*runtimev1.Expression{
				"mv1": {
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_IN,
							Exprs: []*runtimev1.Expression{
								{Expression: &runtimev1.Expression_Ident{Ident: "country"}},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewListValue(&structpb.ListValue{
											Values: []*structpb.Value{
												structpb.NewStringValue("US"),
												structpb.NewStringValue("UK"),
											},
										}),
									},
								},
							},
						},
					},
				},
			},
			additionalTimeRange: nil,
			expected:            []string{"Revenue in US and UK from 2024-01-01 to 2024-02-01: 300"},
			expectEqual:         true,
		},
		{
			name:            "WithMultipleRefs",
			body:            `{{ metrics_sql "SELECT total_revenue FROM {{ ref \"mv1\" }}" }} and {{ metrics_sql "SELECT total_sales FROM {{ ref \"mv2\" }}" }}`,
			useFormatTokens: false,
			additionalWhere: nil,
			expected:        []string{"700 and 900"},
			expectEqual:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
				InstanceId:                   instanceID,
				Body:                         tc.body,
				UseFormatTokens:              tc.useFormatTokens,
				AdditionalWhereByMetricsView: tc.additionalWhere,
				AdditionalTimeRange:          tc.additionalTimeRange,
			})
			require.NoError(t, err)

			if tc.expectEqual {
				require.Equal(t, strings.Join(tc.expected, ""), res.Body)
			} else {
				for _, exp := range tc.expected {
					require.Contains(t, res.Body, exp)
				}
			}
		})
	}
}

func TestResolveTemplatedString_MetricsSQL_MultiField_MultiRow(t *testing.T) {
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

	server, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	tt := []struct {
		name            string
		body            string
		useFormatTokens bool
		expected        []string
	}{
		{
			name: "SingleRowSingleField_BackwardCompatible",
			body: `Total: {{ metrics_sql "select total_bids from bids_metrics" }}`,
			expected: []string{
				"Total: 5",
			},
		},
		{
			name: "MultipleRowsMultipleFields_WithRange",
			body: `{{ $data := metrics_sql_rows "select overall_spend, advertiser_name from bids_metrics order by advertiser_name limit 3" }}
{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}
{{ end }}`,
			expected: []string{
				"- Amazon: 3000",
				"- Apple: 2500",
				"- Google: 1000",
			},
		},
		{
			name: "MultipleRowsSingleField_WithRange",
			body: `{{ $data := metrics_sql_rows "select advertiser_name from bids_metrics order by advertiser_name limit 3" }}
{{ range $data }}- {{ .advertiser_name }}
{{ end }}`,
			expected: []string{
				"- Amazon",
				"- Apple",
				"- Google",
			},
		},
		{
			name: "ComplexTemplate_WithConditional",
			body: `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from bids_metrics order by overall_spend desc limit 3" }}
Top Spenders:
{{ range $data }}{{ if gt .overall_spend 2000.0 }}- **{{ .advertiser_name }}**: ${{ .overall_spend }}
{{ end }}{{ end }}`,
			expected: []string{
				"Top Spenders:",
				"- **Amazon**: $3000",
				"- **Apple**: $2500",
			},
		},
		{
			name: "NestedRef_MultipleRows",
			body: `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from {{ ref \"bids_metrics\" }} order by advertiser_name limit 2" }}
{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}
{{ end }}`,
			expected: []string{
				"- Amazon: 3000",
				"- Apple: 2500",
			},
		},
		{
			name: "MultipleRows_WithFormatTokens",
			body: `{{ $data := metrics_sql_rows "select advertiser_name, overall_spend from bids_metrics order by advertiser_name limit 2" }}
{{ range $data }}- {{ .advertiser_name }}: {{ .overall_spend }}
{{ end }}`,
			useFormatTokens: true,
			expected: []string{
				`__RILL__FORMAT__({"metrics_view":"bids_metrics","field":"overall_spend","value":3000})`,
				`__RILL__FORMAT__({"metrics_view":"bids_metrics","field":"overall_spend","value":2500})`,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := server.ResolveTemplatedString(testCtx(), &runtimev1.ResolveTemplatedStringRequest{
				InstanceId:      instanceID,
				Body:            tc.body,
				UseFormatTokens: tc.useFormatTokens,
			})
			require.NoError(t, err)

			for _, exp := range tc.expected {
				require.Contains(t, res.Body, exp, "Expected output to contain: %s\nFull output:\n%s", exp, res.Body)
			}
		})
	}
}
