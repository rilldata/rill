package metricsview_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestValidateMetricsView(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	mv := &runtimev1.MetricsViewSpec{
		Connector:     "duckdb",
		Table:         "ad_bids",
		DisplayName:   "Ad Bids",
		TimeDimension: "timestamp",
		Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
			{Name: "publisher", Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
			{Name: "records", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "invalid_nested_aggregation", Expression: "MAX(COUNT(DISTINCT publisher))", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "invalid_partition", Expression: "AVG(bid_price) OVER (PARTITION BY publisher)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	e, err := metricsview.NewExecutor(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)

	res, err := e.ValidateMetricsView(context.Background())
	require.NoError(t, err)
	require.Empty(t, res.TimeDimensionErr)
	require.Empty(t, res.DimensionErrs)
	require.Empty(t, res.OtherErrs)

	require.Len(t, res.MeasureErrs, 2)
	require.Equal(t, 1, res.MeasureErrs[0].Idx)
	require.Equal(t, 2, res.MeasureErrs[1].Idx)
}

// ClickHouse does not support expression aliases that collide with column names.
// Check that such metrics views are rejected.
func TestValidateMetricsViewClickHouseNames(t *testing.T) {
	// Start a test runtime with a simple ClickHouse model.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
-- @connector: clickhouse
select parseDateTimeBestEffort('2024-01-01T00:00:00Z') as time, 'DK' as country, 1 as val union all
select parseDateTimeBestEffort('2024-01-02T00:00:00Z') as time, 'US' as country, 2 as val union all
select parseDateTimeBestEffort('2024-01-03T00:00:00Z') as time, 'US' as country, 3 as val union all
select parseDateTimeBestEffort('2024-01-04T00:00:00Z') as time, 'US' as country, 4 as val union all
select parseDateTimeBestEffort('2024-01-05T00:00:00Z') as time, 'DK' as country, 5 as val
`,
		},
	})

	// Test cases of metrics view YAML partials defining dimensions and measures.
	cases := []struct {
		name          string
		partial       string
		errorContains string
	}{
		{
			name: "simple",
			partial: `
dimensions:
  - name: country
    expression: country
measures:
  - name: val_sum
    expression: sum(val)
`,
		},
		{
			name: "measure collides with column",
			partial: `
dimensions:
  - name: country
    expression: country
measures:
  - name: val
    expression: sum(val)
`,
			errorContains: `invalid measure "val": measures cannot have the same name as a column`,
		},
		{
			name: "dimension expression collides with column",
			partial: `
dimensions:
  - name: country
    expression: UPPER(country)
measures:
  - name: val_sum
    expression: sum(val)
		`,
			errorContains: `invalid dimension "country": dimensions that use`,
		},
		{
			name: "dimension expression collides with case insensitive column",
			partial: `
dimensions:
  - name: Country
    expression: UPPER(country)
measures:
  - name: val_sum
    expression: sum(val)
		`,
			errorContains: `invalid dimension "Country": dimensions that use`,
		},
		{
			name: "dimension column is allowed",
			partial: `
dimensions:
  - name: country
    column: country
measures:
  - name: val_sum
    expression: sum(val)
		`,
		},
		{
			name: "dimension expression matching the name is allowed",
			partial: `
dimensions:
  - name: country
    expression: country
measures:
  - name: val_sum
    expression: sum(val)
		`,
		},
	}

	// Execute the test cases
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			metricsView := `
version: 1
type: metrics_view
model: model
timeseries: time
` + c.partial

			testruntime.PutFiles(t, rt, instanceID, map[string]string{"metrics_view.yaml": metricsView})
			testruntime.ReconcileParserAndWait(t, rt, instanceID)
			testruntime.RequireReconcileState(t, rt, instanceID, 3, -1, 0)

			r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics_view")
			if c.errorContains != "" {
				require.Contains(t, r.Meta.ReconcileError, c.errorContains)
			} else {
				require.Empty(t, r.Meta.ReconcileError)
			}
		})
	}
}
