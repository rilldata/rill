package executor

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

func TestRewriteQueryForRollup_NoRollups(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{{Name: "count"}},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_RawRows(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{Table: "rollup", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY},
			},
		},
	}

	qry := &metricsview.Query{Rows: true}
	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_MissingDimension(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
				{Name: "domain", Column: "domain"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher"}, // missing "domain"
					Measures:   []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "domain"}, // not in rollup
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
	}

	// Rejected at eligibility (missing dimension); no watermark fetch needed
	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_MissingMeasure(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
				{Name: "total_clicks", Expression: `SUM("clicks")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
					// missing total_clicks
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_clicks"}, // not in rollup
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_GrainNotDerivable(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "weekly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Query for month grain; month is not derivable from week
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}},
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
	}

	// Rejected at eligibility (grain not derivable); no watermark fetch needed
	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_StartNotAligned(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), // not aligned to day
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Rejected at eligibility (start not aligned to rollup grain); no watermark fetch needed
	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_WhereDimensionMissing(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
				{Name: "domain", Column: "domain"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher"}, // no "domain"
					Measures:   []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		Where: &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorEq,
				Expressions: []*metricsview.Expression{
					{Name: "domain"},
					{Value: "example.com"},
				},
			},
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_ComparisonTimeRange(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		ComparisonTimeRange: &metricsview.TimeRange{
			Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_DerivedMeasure(t *testing.T) {
	// Derived measures are not in the rollup's measure list, so querying one by name should not match
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`, Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				{Name: "derived_measure", Expression: `total_impressions * 2`, Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "derived_measure"},
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_ComputedMeasureRejected(t *testing.T) {
	// Computed measures (count, count_distinct, etc.) produce incorrect results on
	// pre-aggregated rollup tables; queries with them should be rejected.
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher"},
					Measures:   []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "publisher"},
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
			{Name: "__count__", Compute: &metricsview.MeasureCompute{Count: true}},
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_TimezoneRejection(t *testing.T) {
	// Day+ rollup with mismatched timezone should be rejected at eligibility.
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_Rollup{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					TimeZone:  "", // UTC
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeZone: "America/New_York",
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result, err := e.rewriteQueryForRollup(context.Background(), qry)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestCollectWhereDimensions(t *testing.T) {
	expr := &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorAnd,
			Expressions: []*metricsview.Expression{
				{
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorEq,
						Expressions: []*metricsview.Expression{
							{Name: "publisher"},
							{Value: "google"},
						},
					},
				},
				{
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorIn,
						Expressions: []*metricsview.Expression{
							{Name: "domain"},
							{Value: "a.com"},
							{Value: "b.com"},
						},
					},
				},
			},
		},
	}

	dims := collectWhereDimensions(expr)
	require.True(t, dims["publisher"])
	require.True(t, dims["domain"])
	require.Len(t, dims, 2)
}

func TestCollectWhereDimensions_Nil(t *testing.T) {
	dims := collectWhereDimensions(nil)
	require.Empty(t, dims)
}

func TestNormalizeTimezone(t *testing.T) {
	assertTZ := func(expected, input string) {
		t.Helper()
		got, err := normalizeTimezone(input)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	}
	assertTZ("UTC", "")
	assertTZ("UTC", "UTC")
	assertTZ("UTC", "Etc/UTC")
	assertTZ("UTC", "utc")
	assertTZ("America/New_York", "America/New_York")
	_, err := normalizeTimezone("Not/A_Timezone")
	require.Error(t, err)
}
