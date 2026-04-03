package executor

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

const (
	rollupTestDailyTable  = "rollup_day"
	rollupTestWeeklyTable = "rollup_week"
	rollupTestMonthTable  = "rollup_month"
	rollupTestBaseTable   = "base_events"
)

// setupRollupOLAP creates a DuckDB OLAP store with base_events, rollup_day,
// rollup_week (Jan+Feb only), and rollup_month tables.
func setupRollupOLAP(t *testing.T) drivers.OLAPStore {
	conn, err := drivers.Open("duckdb", "duckdb", "default",
		map[string]any{"dsn": ":memory:?access_mode=read_write", "pool_size": "4"},
		storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	ctx := context.Background()

	// Base events: ~2184 rows (91 days * 24 hours/day), 3 publishers cycling, 2 domains alternating
	execSQL(t, olap, ctx, `
		CREATE TABLE base_events AS
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
	`)

	// Daily rollup: full date range
	execSQL(t, olap, ctx, `
		CREATE TABLE rollup_day AS
		SELECT date_trunc('day', timestamp) AS timestamp, publisher, domain,
			SUM(impressions) AS impressions, SUM(clicks) AS clicks
		FROM base_events GROUP BY 1, 2, 3
	`)

	// Weekly rollup: only Jan+Feb (not March), for coverage gap tests
	execSQL(t, olap, ctx, `
		CREATE TABLE rollup_week AS
		SELECT date_trunc('week', timestamp) AS timestamp, publisher, domain,
			SUM(impressions) AS impressions, SUM(clicks) AS clicks
		FROM base_events WHERE timestamp < TIMESTAMP '2024-03-01' GROUP BY 1, 2, 3
	`)

	// Monthly rollup: full date range
	execSQL(t, olap, ctx, `
		CREATE TABLE rollup_month AS
		SELECT date_trunc('month', timestamp) AS timestamp, publisher, domain,
			SUM(impressions) AS impressions, SUM(clicks) AS clicks
		FROM base_events GROUP BY 1, 2, 3
	`)

	return olap
}

func execSQL(t *testing.T, olap drivers.OLAPStore, ctx context.Context, sql string) {
	t.Helper()
	res, err := olap.Query(ctx, &drivers.Statement{Query: sql})
	require.NoError(t, err)
	res.Close()
}

// rollupTestSpec returns a MetricsViewSpec matching the test tables.
func rollupTestSpec() *runtimev1.MetricsViewSpec {
	return &runtimev1.MetricsViewSpec{
		Table:         rollupTestBaseTable,
		TimeDimension: "timestamp",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "publisher", Column: "publisher"},
			{Name: "domain", Column: "domain"},
			{Name: "country", Column: "country"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "total_impressions", Expression: `SUM("impressions")`},
			{Name: "total_clicks", Expression: `SUM("clicks")`},
		},
		Rollups: []*runtimev1.MetricsViewSpec_RollupTable{
			{
				Table:      rollupTestDailyTable,
				TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_DAY,
				Dimensions: []string{"publisher", "domain"},
				Measures:   []string{"total_impressions", "total_clicks"},
			},
			{
				Table:      rollupTestWeeklyTable,
				TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_WEEK,
				Dimensions: []string{"publisher", "domain"},
				Measures:   []string{"total_impressions", "total_clicks"},
			},
			{
				Table:      rollupTestMonthTable,
				TimeGrain:  runtimev1.TimeGrain_TIME_GRAIN_MONTH,
				Dimensions: []string{"publisher", "domain"},
				Measures:   []string{"total_impressions", "total_clicks"},
			},
		},
	}
}

// newRollupTestExecutor creates an Executor backed by a real DuckDB OLAP store.
// The caller should defer e.Close().
func newRollupTestExecutor(t *testing.T, olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec) *Executor {
	return &Executor{
		instanceID:  t.Name(),
		metricsView: mv,
		security:    runtime.ResolvedSecurityOpen,
		olap:        olap,
		olapRelease: func() {},
		timestamps:  make(map[string]metricsview.TimestampsResult),
	}
}

func TestRollupIntegration(t *testing.T) {
	olap := setupRollupOLAP(t)
	mv := rollupTestSpec()

	t.Run("routing", func(t *testing.T) {
		t.Run("day_grain_selects_daily", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			require.Equal(t, rollupTestDailyTable, result.spec.Table)
		})

		t.Run("month_grain_selects_monthly", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Both daily and monthly are eligible; monthly is coarser
			require.Equal(t, rollupTestMonthTable, result.spec.Table)
		})

		t.Run("year_grain_selects_monthly", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Daily and monthly are eligible (year derivable from both); monthly is coarser
			require.Equal(t, rollupTestMonthTable, result.spec.Table)
		})

		t.Run("no_grain_selects_coarsest", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// All 3 eligible (no grain check); monthly is coarsest
			require.Equal(t, rollupTestMonthTable, result.spec.Table)
		})

		t.Run("week_in_coverage_selects_weekly", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Daily and weekly eligible (week derivable from day); weekly is coarser
			require.Equal(t, rollupTestWeeklyTable, result.spec.Table)
		})

		t.Run("week_beyond_coverage_falls_to_daily", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Weekly lacks March data; monthly ineligible (week not derivable from month); daily covers all
			require.Equal(t, rollupTestDailyTable, result.spec.Table)
		})

		t.Run("misaligned_start_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.Nil(t, result)
		})

		t.Run("missing_dimension_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.Nil(t, result)
		})

		t.Run("computed_measure_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.Nil(t, result)
		})

		t.Run("spine_query_uses_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Monthly is coarsest eligible
			require.Equal(t, rollupTestMonthTable, result.spec.Table)
		})

		t.Run("no_time_range_skips_partial_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.NotNil(t, result)
			// Weekly rollup is partial (Jan+Feb only); daily and monthly cover full range.
			// Monthly is coarsest eligible.
			require.Equal(t, rollupTestMonthTable, result.spec.Table)
		})

		t.Run("hour_grain_returns_nil", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.Nil(t, result)
		})

		t.Run("where_filter_dimension_not_in_rollup", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			result := e.rewriteQueryForRollup(context.Background(), qry)
			require.Nil(t, result)
		})
	})

	t.Run("correctness", func(t *testing.T) {
		t.Run("daily_agg_correctness", func(t *testing.T) {
			e := newRollupTestExecutor(t, olap, mv)
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

			// Build AST and execute directly (bypasses Query() which needs rt)
			ast, err := metricsview.NewAST(mv, e.security, qry, olap.Dialect())
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)

			res, err := olap.Query(ctx, &drivers.Statement{Query: sql, Args: args})
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
			e := newRollupTestExecutor(t, olap, mv)
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

			// Use the monthly rollup spec for correctness verification
			monthlySpec := buildSyntheticSpec(mv, mv.Rollups[2])
			ast, err := metricsview.NewAST(monthlySpec, e.security, qry, olap.Dialect())
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)

			res, err := olap.Query(ctx, &drivers.Statement{Query: sql, Args: args})
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
			e := newRollupTestExecutor(t, olap, mv)
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

			// Query base table directly for the filtered correctness test
			// (WHERE on publisher is in the rollup, so rollup could be used too)
			ast, err := metricsview.NewAST(mv, e.security, qry, olap.Dialect())
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)

			res, err := olap.Query(ctx, &drivers.Statement{Query: sql, Args: args})
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
