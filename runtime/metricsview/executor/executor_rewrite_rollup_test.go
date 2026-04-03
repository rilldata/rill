package executor

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

func TestRewriteQueryForRollup_BasicMatch(t *testing.T) {
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
				{Name: "total_clicks", Expression: `SUM("clicks")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher", "domain"},
					Measures:   []string{"total_impressions", "total_clicks"},
				},
			},
		},
	}

	// Pre-populate watermark cache so no OLAP store is needed
	key := watermarkCacheKey("", "", "", "daily_rollup")
	setWatermark(key, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "publisher"},
			{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainDay}}},
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.NotNil(t, result.spec)
	require.Equal(t, "daily_rollup", result.spec.Table)
	require.Empty(t, result.spec.Model)
	require.Nil(t, result.spec.Rollups)

	// Base expressions should be preserved (no rewriting needed)
	for _, m := range result.spec.Measures {
		if m.Name == "total_impressions" {
			require.Equal(t, `SUM("impressions")`, m.Expression)
		}
	}
}

func TestRewriteQueryForRollup_NoRollups(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
		},
	}

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{{Name: "count"}},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_RawRows(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{Table: "rollup", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY},
			},
		},
	}

	qry := &metricsview.Query{Rows: true}
	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_SpineAllowed(t *testing.T) {
	// Spine queries (used by timeseries for null-filling) should still use rollups
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	setWatermark(watermarkCacheKey("", "", "", "daily_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	qry := &metricsview.Query{
		Spine: &metricsview.Spine{},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.Equal(t, "daily_rollup", result.spec.Table)
}

func TestRewriteQueryForRollup_MissingDimension(t *testing.T) {
	e := &Executor{
		instanceID: "test-missing-dim",
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
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher"}, // missing "domain"
					Measures:   []string{"total_impressions"},
				},
			},
		},
	}

	// Provide watermarks so the coverage check doesn't short-circuit; the test should fail on dimension eligibility
	setWatermark(watermarkCacheKey("test-missing-dim", "", "", "base_table"),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
	setWatermark(watermarkCacheKey("test-missing-dim", "", "", "daily_rollup"),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "domain"}, // not in rollup
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_MissingMeasure(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
				{Name: "total_clicks", Expression: `SUM("clicks")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_GrainNotDerivable(t *testing.T) {
	e := &Executor{
		instanceID: "test-grain-not-derivable",
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "weekly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Provide watermarks so the coverage check doesn't short-circuit; the test should fail on grain derivability
	setWatermark(watermarkCacheKey("test-grain-not-derivable", "", "", "base_table"),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
	setWatermark(watermarkCacheKey("test-grain-not-derivable", "", "", "weekly_rollup"),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	// Query for month grain; month is not derivable from week
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}},
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_TimeRangeNotAligned(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_WhereDimensionMissing(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
				{Name: "domain", Column: "domain"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_PreferCoarsestGrain(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "hourly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
					Measures:  []string{"total_impressions"},
				},
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Pre-populate watermark cache for both rollups
	setWatermark(watermarkCacheKey("", "", "", "hourly_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
	setWatermark(watermarkCacheKey("", "", "", "daily_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	// Query for month grain; both hourly and daily are eligible, but daily is coarser
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}},
		},
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.NotNil(t, result.spec)
	require.Equal(t, "daily_rollup", result.spec.Table)
}

func TestRewriteQueryForRollup_ComparisonTimeRange(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_DerivedMeasure(t *testing.T) {
	// Derived measures are not in the rollup's measure list, so querying one by name should not match
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`, Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
				{Name: "derived_measure", Expression: `total_impressions * 2`, Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_ComputedMeasureRejected(t *testing.T) {
	// Computed measures (count, count_distinct, etc.) produce incorrect results on
	// pre-aggregated rollup tables; queries with them should be rejected.
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_NoTimeGrainQuery(t *testing.T) {
	// Query without time grain (pure aggregation) should still match rollup
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table: "base_table",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
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
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.NotNil(t, result.spec)
	require.Equal(t, "daily_rollup", result.spec.Table)
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

func TestRewriteQueryForRollup_TimezoneMatching(t *testing.T) {
	baseMV := func(rollupTZ string) *runtimev1.MetricsViewSpec {
		return &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "daily_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Timezone:  rollupTZ,
					Measures:  []string{"total_impressions"},
				},
			},
		}
	}

	baseQuery := func(tz string) *metricsview.Query {
		return &metricsview.Query{
			Measures: []metricsview.Measure{
				{Name: "total_impressions"},
			},
			TimeZone: tz,
			TimeRange: &metricsview.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		}
	}

	// Pre-populate watermark cache for all timezone test variants
	setWatermark(watermarkCacheKey("", "", "", "daily_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	t.Run("day rollup UTC, query tz New York: falls back", func(t *testing.T) {
		e := &Executor{metricsView: baseMV("")}
		result := e.rewriteQueryForRollup(context.Background(), baseQuery("America/New_York"))
		require.Nil(t, result)
	})

	t.Run("day rollup New York, query tz New York: routes", func(t *testing.T) {
		e := &Executor{metricsView: baseMV("America/New_York")}
		// Use time range aligned to New York day boundaries
		ny, _ := time.LoadLocation("America/New_York")
		qry := baseQuery("America/New_York")
		qry.TimeRange = &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, ny),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, ny),
		}
		result := e.rewriteQueryForRollup(context.Background(), qry)
		require.NotNil(t, result)
		require.NotNil(t, result.spec)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})

	t.Run("day rollup unset, query tz UTC: routes", func(t *testing.T) {
		e := &Executor{metricsView: baseMV("")}
		result := e.rewriteQueryForRollup(context.Background(), baseQuery("UTC"))
		require.NotNil(t, result)
		require.NotNil(t, result.spec)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})

	t.Run("day rollup unset, query tz empty: routes", func(t *testing.T) {
		e := &Executor{metricsView: baseMV("")}
		result := e.rewriteQueryForRollup(context.Background(), baseQuery(""))
		require.NotNil(t, result)
		require.NotNil(t, result.spec)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})

	t.Run("hour rollup, query tz New York: routes (sub-day safe)", func(t *testing.T) {
		mv := baseMV("")
		mv.Rollups[0].TimeGrain = runtimev1.TimeGrain_TIME_GRAIN_HOUR
		e := &Executor{metricsView: mv}
		qry := baseQuery("America/New_York")
		// Align to hour boundaries
		qry.TimeRange = &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		}
		result := e.rewriteQueryForRollup(context.Background(), qry)
		require.NotNil(t, result)
		require.NotNil(t, result.spec)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})
}

func TestNormalizeTimezone(t *testing.T) {
	require.Equal(t, "UTC", normalizeTimezone(""))
	require.Equal(t, "UTC", normalizeTimezone("UTC"))
	require.Equal(t, "UTC", normalizeTimezone("Etc/UTC"))
	require.Equal(t, "UTC", normalizeTimezone("utc"))
	require.Equal(t, "America/New_York", normalizeTimezone("America/New_York"))
}

func TestRewriteQueryForRollup_TimeRangeCoverage_Covered(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "monthly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Rollup has data from 2020-01-01 to 2024-12-01
	setWatermark(watermarkCacheKey("", "", "", "monthly_rollup"),
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.Equal(t, "monthly_rollup", result.spec.Table)
}

func TestRewriteQueryForRollup_TimeRangeCoverage_NotCovered(t *testing.T) {
	e := &Executor{
		instanceID: "test-not-covered",
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "monthly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Base table has data from 2020-01-01 to 2024-12-01
	setWatermark(watermarkCacheKey("test-not-covered", "", "", "base_table"),
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))

	// Rollup only has data from 2023-01-01 to 2024-06-01 (less than base table)
	setWatermark(watermarkCacheKey("test-not-covered", "", "", "monthly_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	// Query starts before rollup data; base table has data there too → rollup should be rejected
	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.Nil(t, result)
}

func TestRewriteQueryForRollup_TimeRangeCoverage_EffectiveEnd(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "monthly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Monthly rollup max is 2024-05-01; effective end = 2024-06-01 (max + 1 month)
	setWatermark(watermarkCacheKey("", "", "", "monthly_rollup"),
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC))

	// Query end is exactly 2024-06-01; should match via effective end
	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.Equal(t, "monthly_rollup", result.spec.Table)
}

func TestRewriteQueryForRollup_PreferSmallestDataRange(t *testing.T) {
	e := &Executor{
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "monthly_rollup_wide",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
				{
					Table:     "monthly_rollup_narrow",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Wide rollup: 10 years of data
	setWatermark(watermarkCacheKey("", "", "", "monthly_rollup_wide"),
		time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))
	// Narrow rollup: 2 years of data
	setWatermark(watermarkCacheKey("", "", "", "monthly_rollup_narrow"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))

	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.Equal(t, "monthly_rollup_narrow", result.spec.Table)
}

func TestRewriteQueryForRollup_QueryWiderThanData(t *testing.T) {
	// When the query range extends beyond both the base table and rollup,
	// the rollup should still be eligible because it has the same data as the base table.
	e := &Executor{
		instanceID: "test-wider",
		metricsView: &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:     "monthly_rollup",
					TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
					Measures:  []string{"total_impressions"},
				},
			},
		},
	}

	// Base table has data from 2023-01-01 to 2024-06-01
	setWatermark(watermarkCacheKey("test-wider", "", "", "base_table"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	// Rollup also has data from 2023-01-01 to 2024-06-01 (same as base)
	setWatermark(watermarkCacheKey("test-wider", "", "", "monthly_rollup"),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

	// Query range is much wider than both base and rollup
	qry := &metricsview.Query{
		Measures: []metricsview.Measure{
			{Name: "total_impressions"},
		},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// query range is clamped to base table range, and rollup covers it.
	result := e.rewriteQueryForRollup(context.Background(), qry)
	require.NotNil(t, result)
	require.Equal(t, "monthly_rollup", result.spec.Table)
}

func TestRewriteQueryForRollup_NoTimeRange_ChecksCoverage(t *testing.T) {
	baseMV := func() *runtimev1.MetricsViewSpec {
		return &runtimev1.MetricsViewSpec{
			Table:         "base_table",
			TimeDimension: "timestamp",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "publisher", Column: "publisher"},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total_impressions", Expression: `SUM("impressions")`},
			},
			Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
				{
					Table:      "daily_rollup",
					TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
					Dimensions: []string{"publisher"},
					Measures:   []string{"total_impressions"},
				},
			},
		}
	}

	baseQuery := func() *metricsview.Query {
		return &metricsview.Query{
			Dimensions: []metricsview.Dimension{{Name: "publisher"}},
			Measures:   []metricsview.Measure{{Name: "total_impressions"}},
		}
	}

	t.Run("full_coverage_uses_rollup", func(t *testing.T) {
		e := &Executor{instanceID: "no-tr-full", metricsView: baseMV()}

		// Base and rollup cover the same range
		setWatermark(watermarkCacheKey("no-tr-full", "", "", "base_table"),
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
		setWatermark(watermarkCacheKey("no-tr-full", "", "", "daily_rollup"),
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

		result := e.rewriteQueryForRollup(context.Background(), baseQuery())
		require.NotNil(t, result)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})

	t.Run("partial_coverage_returns_nil", func(t *testing.T) {
		e := &Executor{instanceID: "no-tr-partial", metricsView: baseMV()}

		// Base covers Jan-Jun, rollup only covers Mar-Jun
		setWatermark(watermarkCacheKey("no-tr-partial", "", "", "base_table"),
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
		setWatermark(watermarkCacheKey("no-tr-partial", "", "", "daily_rollup"),
			time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

		result := e.rewriteQueryForRollup(context.Background(), baseQuery())
		require.Nil(t, result)
	})

	t.Run("no_base_watermark_returns_nil", func(t *testing.T) {
		e := &Executor{instanceID: "no-tr-no-base", metricsView: baseMV()}

		// No base watermark cached, no OLAP store => fetchBaseWatermark returns false
		result := e.rewriteQueryForRollup(context.Background(), baseQuery())
		require.Nil(t, result)
	})

	t.Run("no_time_dimension_skips_coverage", func(t *testing.T) {
		mv := baseMV()
		mv.TimeDimension = "" // no time dimension
		e := &Executor{instanceID: "no-tr-no-td", metricsView: mv}

		// No watermarks needed; coverage check is skipped entirely
		result := e.rewriteQueryForRollup(context.Background(), baseQuery())
		require.NotNil(t, result)
		require.Equal(t, "daily_rollup", result.spec.Table)
	})
}
