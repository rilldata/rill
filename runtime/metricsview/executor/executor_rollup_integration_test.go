package executor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

const (
	rollupTestDailyTable  = "rollup_day"
	rollupTestWeeklyTable = "rollup_week"
	rollupTestMonthTable  = "rollup_month"
	rollupTestMVName      = "mv"
)

// rollupTestFiles returns the project files for the rollup integration tests.
func rollupTestFiles() map[string]string {
	return map[string]string{
		"rill.yaml": "",
		"models/base_events.sql": `
SELECT
	ts AS timestamp,
	CASE (row_number() OVER ()) % 3
		WHEN 0 THEN 'Google'
		WHEN 1 THEN 'Facebook'
		ELSE 'Microsoft'
	END AS publisher,
	CASE (row_number() OVER ()) % 2
		WHEN 0 THEN 'news.com'
		ELSE 'sports.com'
	END AS domain,
	'US' AS country,
	10 AS impressions,
	2 AS clicks
FROM generate_series(TIMESTAMP '2024-01-01 00:00:00', TIMESTAMP '2024-03-31 23:00:00', INTERVAL '1 HOUR') t(ts)
`,
		"models/rollup_day.sql": `
SELECT date_trunc('day', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events GROUP BY 1, 2, 3
`,
		// Weekly rollup: only Jan+Feb (not March), for coverage gap tests
		"models/rollup_week.sql": `
SELECT date_trunc('week', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events WHERE timestamp < TIMESTAMP '2024-03-01' GROUP BY 1, 2, 3
`,
		"models/rollup_month.sql": `
SELECT date_trunc('month', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events GROUP BY 1, 2, 3
`,
		"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
  - name: country
    column: country
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
  - name: total_clicks
    expression: 'SUM("clicks")'
rollups:
  - model: rollup_day
    time_grain: day
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
      - total_clicks
  - model: rollup_week
    time_grain: week
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
      - total_clicks
  - model: rollup_month
    time_grain: month
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
      - total_clicks
explore:
  skip: true
`,
	}
}

// newRollupTestRuntime creates a runtime with the rollup test project files.
func newRollupTestRuntime(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: rollupTestFiles(),
	})
	// 1 project_parser + 4 models + 1 metrics_view = 6 resources; 0 errors
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)
	return rt, instanceID
}

// newRollupTestExecutor creates an Executor backed by a real runtime and OLAP store.
func newRollupTestExecutor(t *testing.T, rt *runtime.Runtime, instanceID string) *executor.Executor {
	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, rollupTestMVName)
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	e, err := executor.New(context.Background(), rt, instanceID, rollupTestMVName, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	return e
}

func TestRollupIntegration(t *testing.T) {
	rt, instanceID := newRollupTestRuntime(t)

	t.Run("routing", func(t *testing.T) {
		t.Run("day_grain_selects_daily", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
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

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, rollupTestDailyTable, table)
		})

		t.Run("month_grain_selects_monthly", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

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

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Both daily and monthly are eligible; monthly is coarser
			require.Equal(t, rollupTestMonthTable, table)
		})

		t.Run("year_grain_selects_monthly", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainYear}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Daily and monthly are eligible (year derivable from both); monthly is coarser
			require.Equal(t, rollupTestMonthTable, table)
		})

		t.Run("no_grain_selects_coarsest", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// All 3 eligible (no grain check); monthly is coarsest
			require.Equal(t, rollupTestMonthTable, table)
		})

		t.Run("week_in_coverage_selects_weekly", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			// Both start and end must be Mondays for week alignment.
			// Jan 1, 2024 = Monday; Feb 5, 2024 = Monday (5 weeks later)
			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainWeek}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
				},
			}

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Daily and weekly eligible (week derivable from day); weekly is coarser
			require.Equal(t, rollupTestWeeklyTable, table)
		})

		t.Run("week_beyond_coverage_falls_to_daily", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainWeek}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Weekly lacks March data; monthly ineligible (week not derivable from month); daily covers all
			require.Equal(t, rollupTestDailyTable, table)
		})

		t.Run("misaligned_start_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00 not aligned to any day+ grain
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("missing_dimension_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "country"}, // country is not in any rollup
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("computed_measure_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
					{Name: "__count__", Compute: &metricsview.MeasureCompute{Count: true}},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("spine_query_uses_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

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

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Monthly is coarsest eligible
			require.Equal(t, rollupTestMonthTable, table)
		})

		t.Run("no_time_range_skips_partial_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				// No TimeRange: means "all data"
			}

			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Weekly rollup is partial (Jan+Feb only); daily and monthly cover full range.
			// Monthly is coarsest eligible.
			require.Equal(t, rollupTestMonthTable, table)
		})

		t.Run("hour_grain_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainHour}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			}

			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("where_filter_dimension_not_in_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				Where: &metricsview.Expression{
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorEq,
						Expressions: []*metricsview.Expression{
							{Name: "country"},
							{Value: "US"},
						},
					},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("time_range_coverage_not_covered", func(t *testing.T) {
			// Only rollup covers Feb+; query from Jan to Mar cannot be covered
			files := map[string]string{
				"rill.yaml": "",
				"models/base_events.sql": rollupTestFiles()["models/base_events.sql"],
				"models/rollup_day.sql": `
SELECT date_trunc('day', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events WHERE timestamp >= TIMESTAMP '2024-02-01' GROUP BY 1, 2, 3`,
				"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
rollups:
  - model: rollup_day
    time_grain: day
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
explore:
  skip: true`,
			}
			customRT, customID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: files})
			testruntime.RequireReconcileState(t, customRT, customID, 4, 0, 0)
			e := newRollupTestExecutor(t, customRT, customID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainDay}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})

		t.Run("query_wider_than_data", func(t *testing.T) {
			// Query range much wider than base data; clamping makes rollup cover it
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainDay}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Query clamped to base data range; daily is the only eligible rollup for day grain
			require.Equal(t, rollupTestDailyTable, table)
		})

		t.Run("prefer_smallest_data_range", func(t *testing.T) {
			// Two monthly rollups: wide (Jan-Mar) and narrow (Feb-Mar); query Feb-Apr picks narrow
			files := map[string]string{
				"rill.yaml": "",
				"models/base_events.sql": rollupTestFiles()["models/base_events.sql"],
				"models/rollup_month_wide.sql": `
SELECT date_trunc('month', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events GROUP BY 1, 2, 3`,
				"models/rollup_month_narrow.sql": `
SELECT date_trunc('month', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events WHERE timestamp >= TIMESTAMP '2024-02-01' GROUP BY 1, 2, 3`,
				"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
rollups:
  - model: rollup_month_wide
    time_grain: month
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
  - model: rollup_month_narrow
    time_grain: month
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
explore:
  skip: true`,
			}
			customRT, customID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: files})
			testruntime.RequireReconcileState(t, customRT, customID, 5, 0, 0)
			e := newRollupTestExecutor(t, customRT, customID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			// Both monthly rollups cover the range; narrow has smaller data range
			require.Equal(t, "rollup_month_narrow", table)
		})
	})

	t.Run("no_time_range_checks_coverage", func(t *testing.T) {
		t.Run("only_partial_rollup_returns_nil", func(t *testing.T) {
			// Only rollup is partial (Jan+Feb); no time range requires full coverage
			files := map[string]string{
				"rill.yaml": "",
				"models/base_events.sql": rollupTestFiles()["models/base_events.sql"],
				"models/rollup_week.sql": `
SELECT date_trunc('week', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events WHERE timestamp < TIMESTAMP '2024-03-01' GROUP BY 1, 2, 3`,
				"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
rollups:
  - model: rollup_week
    time_grain: week
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
explore:
  skip: true`,
			}
			customRT, customID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: files})
			testruntime.RequireReconcileState(t, customRT, customID, 4, 0, 0)
			e := newRollupTestExecutor(t, customRT, customID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
			}
			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})
	})

	t.Run("first_day_of_week", func(t *testing.T) {
		weeklyOnlyFiles := func(fdow int) map[string]string {
			return map[string]string{
				"rill.yaml": "",
				"models/base_events.sql": rollupTestFiles()["models/base_events.sql"],
				"models/rollup_week.sql": `
SELECT date_trunc('week', timestamp) AS timestamp, publisher, domain,
	SUM(impressions) AS impressions, SUM(clicks) AS clicks
FROM base_events WHERE timestamp < TIMESTAMP '2024-03-01' GROUP BY 1, 2, 3`,
				"metrics_views/mv.yaml": fmt.Sprintf(`
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
first_day_of_week: %d
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
rollups:
  - model: rollup_week
    time_grain: week
    dimensions:
      - publisher
      - domain
    measures:
      - total_impressions
explore:
  skip: true`, fdow),
			}
		}

		t.Run("sunday_aligned_routes", func(t *testing.T) {
			// first_day_of_week=7 (Sunday); Sunday-aligned boundaries should route
			customRT, customID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: weeklyOnlyFiles(7)})
			testruntime.RequireReconcileState(t, customRT, customID, 4, 0, 0)
			e := newRollupTestExecutor(t, customRT, customID)
			defer e.Close()

			// 2024-01-07 = Sunday, 2024-02-04 = Sunday
			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainWeek}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 4, 0, 0, 0, 0, time.UTC),
				},
			}
			table, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, "rollup_week", table)
		})

		t.Run("monday_rejects_sunday_boundaries", func(t *testing.T) {
			// first_day_of_week=1 (Monday); same Sunday boundaries are not aligned
			customRT, customID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: weeklyOnlyFiles(1)})
			testruntime.RequireReconcileState(t, customRT, customID, 4, 0, 0)
			e := newRollupTestExecutor(t, customRT, customID)
			defer e.Close()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainWeek}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 4, 0, 0, 0, 0, time.UTC),
				},
			}
			_, ok, err := e.RewriteQueryForRollupTest(context.Background(), qry)
			require.NoError(t, err)
			require.False(t, ok)
		})
	})

	t.Run("correctness", func(t *testing.T) {
		t.Run("daily_agg_correctness", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()
			ctx := context.Background()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "timestamp", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainDay}}},
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				Sort: []metricsview.Sort{
					{Name: "timestamp"},
					{Name: "publisher"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				},
			}

			res, err := e.Query(ctx, qry, nil)
			require.NoError(t, err)
			defer res.Close()

			type row struct {
				ts          time.Time
				publisher   string
				impressions float64
			}
			var rows []row
			for res.Next() {
				var r row
				require.NoError(t, res.Scan(&r.ts, &r.publisher, &r.impressions))
				rows = append(rows, r)
			}
			require.NoError(t, res.Err())

			// 3 publishers x 2 days = 6 rows
			require.Len(t, rows, 6)

			// Each publisher gets 8 rows per day (24 hours / 3 publishers) => 80 impressions
			for _, r := range rows {
				require.Equal(t, float64(80), r.impressions, "publisher=%s, day=%s", r.publisher, r.ts)
			}
		})

		t.Run("monthly_agg_correctness", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()
			ctx := context.Background()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "timestamp", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				Sort: []metricsview.Sort{
					{Name: "timestamp"},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			res, err := e.Query(ctx, qry, nil)
			require.NoError(t, err)
			defer res.Close()

			type row struct {
				ts          time.Time
				impressions float64
			}
			var rows []row
			for res.Next() {
				var r row
				require.NoError(t, res.Scan(&r.ts, &r.impressions))
				rows = append(rows, r)
			}
			require.NoError(t, res.Err())

			// 3 months: Jan, Feb, Mar
			require.Len(t, rows, 3)

			// Jan: 744 hours * 10 = 7440
			require.Equal(t, float64(7440), rows[0].impressions, "January")
			// Feb: 696 hours * 10 = 6960 (2024 is leap year: 29 days * 24 = 696)
			require.Equal(t, float64(6960), rows[1].impressions, "February")
			// Mar: 744 hours * 10 = 7440
			require.Equal(t, float64(7440), rows[2].impressions, "March")
		})

		t.Run("no_grain_with_filter_correctness", func(t *testing.T) {
			e := newRollupTestExecutor(t, rt, instanceID)
			defer e.Close()
			ctx := context.Background()

			qry := &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "publisher"},
				},
				Measures: []metricsview.Measure{
					{Name: "total_impressions"},
				},
				Where: &metricsview.Expression{
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorEq,
						Expressions: []*metricsview.Expression{
							{Name: "publisher"},
							{Value: "Google"},
						},
					},
				},
				TimeRange: &metricsview.TimeRange{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			res, err := e.Query(ctx, qry, nil)
			require.NoError(t, err)
			defer res.Close()

			type row struct {
				publisher   string
				impressions float64
			}
			var rows []row
			for res.Next() {
				var r row
				require.NoError(t, res.Scan(&r.publisher, &r.impressions))
				rows = append(rows, r)
			}
			require.NoError(t, res.Err())

			require.Len(t, rows, 1)
			require.Equal(t, "Google", rows[0].publisher)
			// Google gets 248 of 744 hours in January (every 3rd row starting at row 3)
			// 248 * 10 = 2480
			require.Equal(t, float64(2480), rows[0].impressions)
		})
	})
}
