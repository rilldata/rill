package executor_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
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
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "publisher", Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "records", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "invalid_nested_aggregation", Expression: "MAX(COUNT(DISTINCT publisher))", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "invalid_partition", Expression: "AVG(bid_price) OVER (PARTITION BY publisher)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0)
	require.NoError(t, err)

	res, err := e.ValidateAndNormalizeMetricsView(context.Background())
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
			"rill.yaml": "olap_connector: clickhouse",
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
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

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

func TestValidateAndNormalizeMetricsViewSmallestTimeGrain(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"model.sql": `
SELECT '2025-05-01T00:00:00Z'::TIMESTAMP AS t_timestamp, '2025-05-01'::DATE AS t_date
		`,
		`ok_none.yaml`: `
version: 1
type: metrics_view
model: model
measures:
- expression: count(*)
`,
		`ok_timestamp_default.yaml`: `
version: 1
type: metrics_view
model: model
timeseries: t_timestamp
measures:
- expression: count(*)
`,
		`ok_date_default.yaml`: `
version: 1
type: metrics_view
model: model
timeseries: t_date
measures:
- expression: count(*)
`,
		`ok_timestamp_month.yaml`: `
version: 1
type: metrics_view
model: model
timeseries: t_timestamp
smallest_time_grain: month
measures:
- expression: count(*)
`,
		`ok_date_month.yaml`: `
version: 1
type: metrics_view
model: model
timeseries: t_date
smallest_time_grain: month
measures:
- expression: count(*)
`,
		`ok_dimension_date_default.yaml`: `
version: 1
type: metrics_view
model: model
dimensions:
- column: t_date
measures:
- expression: count(*)
`,
		`ok_dimension_date.yaml`: `
version: 1
type: metrics_view
model: model
dimensions:
- column: t_date
  smallest_time_grain: month
measures:
- expression: count(*)
`,
		`fail_date_hour.yaml`: `
version: 1
type: metrics_view
model: model
timeseries: t_date
smallest_time_grain: hour
measures:
- expression: count(*)
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 10, 1, 0)

	mv := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_none")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_timestamp_default")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_SECOND, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_date_default")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_DAY, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_timestamp_month")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MONTH, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_date_month")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MONTH, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_dimension_date_default")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_DAY, mv.GetMetricsView().State.ValidSpec.Dimensions[0].SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "ok_dimension_date")
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, mv.GetMetricsView().State.ValidSpec.SmallestTimeGrain)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MONTH, mv.GetMetricsView().State.ValidSpec.Dimensions[0].SmallestTimeGrain)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "fail_date_hour")
	require.NotEmpty(t, mv.Meta.ReconcileError)
	require.Contains(t, mv.Meta.ReconcileError, "smaller than the smallest possible grain")
}

func TestValidateAnnotations(t *testing.T) {
	// Start a test runtime with a simple ClickHouse model.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"simple_model.sql": `
select 'DK' as country, 1 as val union all
select 'US' as country, 2 as val union all
select 'US' as country, 3 as val union all
select 'US' as country, 4 as val union all
select 'DK' as country, 5 as val
`,
			"model.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'DK' as country, 1 as val union all
select '2024-01-02T00:00:00Z'::TIMESTAMP as time, 'US' as country, 2 as val union all
select '2024-01-03T00:00:00Z'::TIMESTAMP as time, 'US' as country, 3 as val union all
select '2024-01-04T00:00:00Z'::TIMESTAMP as time, 'US' as country, 4 as val union all
select '2024-01-05T00:00:00Z'::TIMESTAMP as time, 'DK' as country, 5 as val`,
			"simple_annotation.sql": `
select '2022-01-05T00:00:00Z' as time, '1st event' as description
union all
select '2022-02-16T00:00:00Z' as time, '2nd event' as description
union all
select '2022-03-27T00:00:00Z' as time, '3rd event' as description`,
			"time_end_annotation.sql": `
select '2022-01-05T00:00:00Z' as time, '2022-01-09T00:00:00Z' as time_end, '1st event' as description
union all
select '2022-02-16T00:00:00Z' as time, '2022-02-20T00:00:00Z' as time_end, '2nd event' as description
union all
select '2022-03-27T00:00:00Z' as time, '2022-04-11T00:00:00Z' as time_end, '3rd event' as description`,
			"grain_annotation.sql": `
select '2022-01-05T00:00:00Z' as time, 'day' as duration, '1st event' as description
union all
select '2022-02-16T00:00:00Z' as time, 'month' as duration, '2nd event' as description
union all
select '2022-03-27T00:00:00Z' as time, 'hour' as duration, '3rd event' as description`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)

	cases := []struct {
		name          string
		partial       string
		annotation    []*runtimev1.MetricsViewSpec_Annotation
		errorContains string
	}{
		{
			name: "simple model annotation",
			partial: `
annotations:
  - model: simple_annotation`,
			annotation: []*runtimev1.MetricsViewSpec_Annotation{
				{
					Name:      "simple_annotation",
					Connector: "duckdb",
					Table:     "simple_annotation",
					Model:     "simple_annotation",
					Measures:  []string{"val_sum"},
				},
			},
			errorContains: "",
		},
		{
			name: "simple talbe annotation",
			partial: `
annotations:
  - table: simple_annotation`,
			annotation: []*runtimev1.MetricsViewSpec_Annotation{
				{
					Name:     "simple_annotation",
					Table:    "simple_annotation",
					Measures: []string{"val_sum"},
				},
			},
			errorContains: "",
		},
		{
			name: "with time_end",
			partial: `
annotations:
  - model: time_end_annotation`,
			annotation: []*runtimev1.MetricsViewSpec_Annotation{
				{
					Name:       "time_end_annotation",
					Connector:  "duckdb",
					Table:      "time_end_annotation",
					Model:      "time_end_annotation",
					Measures:   []string{"val_sum"},
					HasTimeEnd: true,
				},
			},
			errorContains: "",
		},
		{
			name: "with grain",
			partial: `
annotations:
  - model: grain_annotation`,
			annotation: []*runtimev1.MetricsViewSpec_Annotation{
				{
					Name:        "grain_annotation",
					Connector:   "duckdb",
					Table:       "grain_annotation",
					Model:       "grain_annotation",
					Measures:    []string{"val_sum"},
					HasDuration: true,
				},
			},
			errorContains: "",
		},
		{
			name: "missing model",
			partial: `
annotations:
  - model: missing_model`,
			errorContains: `failed to get table details "missing_model" for annotation "missing_model"`,
		},
		{
			name: "model without time",
			partial: `
annotations:
  - model: simple_model`,
			errorContains: `table "simple_model" for annotation "simple_model" does not have the required "time" column`,
		},
		{
			name: "model without description",
			partial: `
annotations:
  - model: model`,
			errorContains: `table "model" for annotation "model" does not have the required "description" column`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			metricsView := `
version: 1
type: metrics_view
model: model
timeseries: time
dimensions:
  - name: country
    expression: country
measures:
  - name: val_sum
    expression: sum(val)
` + c.partial

			testruntime.PutFiles(t, rt, instanceID, map[string]string{"metrics_view.yaml": metricsView})
			testruntime.ReconcileParserAndWait(t, rt, instanceID)
			testruntime.RequireReconcileState(t, rt, instanceID, 7, -1, 0)

			r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics_view")
			if c.errorContains != "" {
				require.Contains(t, r.Meta.ReconcileError, c.errorContains)
			} else {
				require.Empty(t, r.Meta.ReconcileError)
				require.Equal(t, c.annotation, r.GetMetricsView().State.ValidSpec.Annotations)
			}
		})
	}
}
