package executor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/starrocks"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestTableFilter_QueriesTimestampsAndSummary(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/orders.sql": `
SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'old' AS snapshot, 'active' AS status, 'US' AS country, 100 AS amount
UNION ALL
SELECT '2025-01-10T00:00:00Z'::TIMESTAMP AS ts, 'old' AS snapshot, 'active' AS status, 'US' AS country, 200 AS amount
UNION ALL
SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'active' AS status, 'US' AS country, 10 AS amount
UNION ALL
SELECT '2025-01-03T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'pending' AS status, 'US' AS country, 20 AS amount
UNION ALL
SELECT '2025-01-04T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'zzz_cancelled' AS status, 'US' AS country, 30 AS amount
UNION ALL
SELECT '2025-01-04T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'active' AS status, 'EU' AS country, 40 AS amount
`,
			"metrics/orders_or.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: ts
table_filter: "snapshot = 'latest' AND (status = 'active' OR status = 'pending')"
dimensions:
  - column: country
  - column: status
measures:
  - name: total_amount
    expression: SUM(amount)
  - name: order_count
    expression: COUNT(*)
`,
			"metrics/orders_top_or.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: ts
table_filter: "status = 'pending' OR country = 'EU'"
dimensions:
  - column: country
  - column: status
measures:
  - name: total_amount
    expression: SUM(amount)
`,
			"metrics/orders_with_security.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: ts
table_filter: "snapshot = 'latest'"
security:
  access: true
  row_filter: "country = 'US'"
dimensions:
  - column: country
  - column: status
measures:
  - name: total_amount
    expression: SUM(amount)
`,
			"metrics/orders_empty.yaml": `
version: 1
type: metrics_view
model: orders
timeseries: ts
table_filter: "snapshot = 'none'"
dimensions:
  - column: country
measures:
  - name: total_amount
    expression: SUM(amount)
`,
			"metrics/orders_child.yaml": `
version: 1
type: metrics_view
parent: orders_or
table_filter: "country = 'US'"
parent_dimensions: '*'
parent_measures: '*'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)

	getValidSpec := func(name string) *runtimev1.MetricsViewSpec {
		res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
		require.NoError(t, err)
		require.NotNil(t, res.GetMetricsView().State.ValidSpec)
		return res.GetMetricsView().State.ValidSpec
	}

	t.Run("OR table_filter applies under SkipChecks with query where and time range", func(t *testing.T) {
		mv := getValidSpec("orders_or")
		e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
		require.NoError(t, err)
		defer e.Close()

		// Aggregation with where country = 'US' must parenthesize table_filter so EU/old rows never leak.
		res, err := e.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}, {Name: "order_count"}},
			Where: &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{{Name: "country"}, {Value: "US"}},
				},
			},
			TimeRange: &metricsview.TimeRange{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, res.Next())
		row := map[string]any{}
		require.NoError(t, res.MapScan(row))
		require.Equal(t, "30", fmt.Sprint(row["total_amount"]))
		require.EqualValues(t, 2, row["order_count"])
		require.NoError(t, res.Close())

		// Rows query also enforces table_filter.
		rowsRes, err := e.Query(context.Background(), &metricsview.Query{
			Rows:  true,
			Limit: ptrInt64(100),
		}, nil)
		require.NoError(t, err)
		var rowCount int
		for rowsRes.Next() {
			rowCount++
		}
		require.NoError(t, rowsRes.Err())
		require.NoError(t, rowsRes.Close())
		require.Equal(t, 3, rowCount)

		// Timestamps exclude old snapshot rows (2025-01-01 and 2025-01-10).
		ts, err := e.Timestamps(context.Background(), "")
		require.NoError(t, err)
		require.Equal(t, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), ts.Min)
		require.Equal(t, time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), ts.Max)
		require.Equal(t, time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), ts.Watermark)

		// Summary reflects only filtered rows (zzz_cancelled on 2025-01-04 must be excluded by table_filter in categorical summary).
		summary, err := e.Summary(context.Background())
		require.NoError(t, err)
		require.Equal(t, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), summary.DefaultTimeDimension.MinValue)
		require.Equal(t, time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), summary.DefaultTimeDimension.MaxValue)
		var foundStatus bool
		for _, d := range summary.Dimensions {
			if d.Name == "status" {
				foundStatus = true
				require.Equal(t, "pending", d.MaxValue)
			}
		}
		require.True(t, foundStatus, "expected status dimension in summary")

		// Search excludes values outside table_filter while returning matching values inside it.
		searchExcluded, err := e.Search(context.Background(), &metricsview.SearchQuery{
			Dimensions: []string{"status"},
			Search:     "zzz",
		}, nil)
		require.NoError(t, err)
		require.Empty(t, searchExcluded)

		searchIncluded, err := e.Search(context.Background(), &metricsview.SearchQuery{
			Dimensions: []string{"status"},
			Search:     "pend",
		}, nil)
		require.NoError(t, err)
		require.Len(t, searchIncluded, 1)
		require.Equal(t, "pending", searchIncluded[0].Value)
	})

	t.Run("unparenthesized top-level OR table_filter preserves precedence with query where", func(t *testing.T) {
		mv := getValidSpec("orders_top_or")
		e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
		require.NoError(t, err)
		defer e.Close()

		// Without parens around `status = 'pending' OR country = 'EU'`, `AND country = 'US'` would still match status='pending' AND leak if reversed;
		// filtering `country = 'US'` must yield only the US/pending row (20), excluding EU/active (40).
		res, err := e.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
			Where: &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{{Name: "country"}, {Value: "US"}},
				},
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, res.Next())
		row := map[string]any{}
		require.NoError(t, res.MapScan(row))
		require.NoError(t, res.Close())
		require.Equal(t, "20", fmt.Sprint(row["total_amount"]))

		// Filtering `status = 'active'` must yield only the EU/active row (40), excluding US/active (10, 100, 200).
		res2, err := e.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
			Where: &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{{Name: "status"}, {Value: "active"}},
				},
			},
		}, nil)
		require.NoError(t, err)
		require.True(t, res2.Next())
		row2 := map[string]any{}
		require.NoError(t, res2.MapScan(row2))
		require.NoError(t, res2.Close())
		require.Equal(t, "40", fmt.Sprint(row2["total_amount"]))
	})

	t.Run("combines with security.row_filter and persists under SkipChecks", func(t *testing.T) {
		mvRes, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "orders_with_security"}, false)
		require.NoError(t, err)
		mv := mvRes.GetMetricsView().State.ValidSpec

		// With SkipChecks (ResolvedSecurityOpen): row_filter is skipped, table_filter still applies (10+20+30+40=100).
		eOpen, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
		require.NoError(t, err)
		defer eOpen.Close()
		resOpen, err := eOpen.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
		}, nil)
		require.NoError(t, err)
		require.True(t, resOpen.Next())
		rowOpen := map[string]any{}
		require.NoError(t, resOpen.MapScan(rowOpen))
		require.NoError(t, resOpen.Close())
		require.Equal(t, "100", fmt.Sprint(rowOpen["total_amount"]))

		// With normal user claims: both table_filter (snapshot='latest') AND row_filter (country='US') apply (10+20+30=60).
		sec, err := rt.ResolveSecurity(context.Background(), instanceID, &runtime.SecurityClaims{
			UserAttributes: map[string]any{},
			Permissions:    []runtime.Permission{runtime.ReadMetrics},
		}, mvRes)
		require.NoError(t, err)
		require.True(t, sec.CanAccess())

		eSec, err := executor.New(context.Background(), rt, instanceID, mv, false, sec, 0, nil)
		require.NoError(t, err)
		defer eSec.Close()
		resSec, err := eSec.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
		}, nil)
		require.NoError(t, err)
		defer resSec.Close()
		require.True(t, resSec.Next())
		rowSec := map[string]any{}
		require.NoError(t, resSec.MapScan(rowSec))
		require.Equal(t, "60", fmt.Sprint(rowSec["total_amount"]))
	})

	t.Run("zero-row table_filter returns empty aggregation and zero timestamps without error", func(t *testing.T) {
		mv := getValidSpec("orders_empty")
		e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
		require.NoError(t, err)
		defer e.Close()

		ts, err := e.Timestamps(context.Background(), "")
		require.NoError(t, err)
		require.True(t, ts.Min.IsZero())
		require.True(t, ts.Max.IsZero())
		require.True(t, ts.Watermark.IsZero())

		res, err := e.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
		}, nil)
		require.NoError(t, err)
		require.True(t, res.Next())
		row := map[string]any{}
		require.NoError(t, res.MapScan(row))
		require.NoError(t, res.Close())
		require.Nil(t, row["total_amount"])

		summary, err := e.Summary(context.Background())
		require.NoError(t, err)
		require.NotNil(t, summary)
	})

	t.Run("derived view enforces parent AND child table_filter on queries and timestamps", func(t *testing.T) {
		mv := getValidSpec("orders_child")
		e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
		require.NoError(t, err)
		defer e.Close()

		res, err := e.Query(context.Background(), &metricsview.Query{
			Measures: []metricsview.Measure{{Name: "total_amount"}},
		}, nil)
		require.NoError(t, err)
		require.True(t, res.Next())
		row := map[string]any{}
		require.NoError(t, res.MapScan(row))
		require.NoError(t, res.Close())
		require.Equal(t, "30", fmt.Sprint(row["total_amount"]))

		ts, err := e.Timestamps(context.Background(), "")
		require.NoError(t, err)
		require.Equal(t, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), ts.Min)
		require.Equal(t, time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), ts.Max)
	})
}

func TestTableFilter_RollupDoesNotInheritBaseTableFilter(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/base_events.sql": `
SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'old' AS snapshot, 'Google' AS publisher, 999 AS impressions
UNION ALL
SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'Google' AS publisher, 10 AS impressions
UNION ALL
SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS ts, 'latest' AS snapshot, 'Facebook' AS publisher, 20 AS impressions
`,
			// Rollup intentionally omits the raw `snapshot` column.
			"models/rollup_day.sql": `
SELECT date_trunc('day', ts) AS ts, publisher, SUM(impressions) AS impressions
FROM base_events
WHERE snapshot = 'latest'
GROUP BY 1, 2
`,
			"metrics/mv.yaml": `
version: 1
type: metrics_view
model: base_events
timeseries: ts
table_filter: "snapshot = 'latest'"
dimensions:
  - column: publisher
measures:
  - name: impressions
    expression: SUM(impressions)
rollups:
  - model: rollup_day
    time_grain: day
    dimensions: [publisher]
    measures: [impressions]
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "mv"}, false)
	require.NoError(t, err)
	mv := res.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	e, err := executor.New(context.Background(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	// Timestamps must resolve both base and rollup without failing on missing `snapshot` column in rollup_day.
	ts, err := e.Timestamps(context.Background(), "")
	require.NoError(t, err)
	require.Contains(t, ts.Rollups, "rollup_day")
	require.Equal(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), ts.Rollups["rollup_day"].Min)

	// Eligible day-grain query routes to rollup_day without injecting base table_filter.
	qRes, err := e.Query(context.Background(), &metricsview.Query{
		Dimensions: []metricsview.Dimension{{
			Name:    "ts",
			Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainDay}},
		}},
		Measures: []metricsview.Measure{{Name: "impressions"}},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}, nil)
	require.NoError(t, err)
	require.NoError(t, qRes.Close())
	require.Equal(t, "rollup_day", e.LatestQueryTable())

	// Raw rows query bypasses rollup and still applies base table_filter (excludes old snapshot row).
	rowsRes, err := e.Query(context.Background(), &metricsview.Query{
		Rows:  true,
		Limit: ptrInt64(100),
	}, nil)
	require.NoError(t, err)
	require.Equal(t, "base_events", e.LatestQueryTable())
	var count int
	for rowsRes.Next() {
		count++
	}
	require.NoError(t, rowsRes.Err())
	require.NoError(t, rowsRes.Close())
	require.Equal(t, 2, count)
}

func TestTableFilter_StarRocksASTSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Database:       "hive_ctl",
		DatabaseSchema: "dwd",
		Table:          "orders_1h_f",
		TableFilter:    "datetime = 'latest' OR datetime = 'current'",
		TimeDimension:  "ts",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "ts", Column: "ts", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME},
			{Name: "region", Column: "region"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "cnt", Expression: "CAST(COUNT(*) AS DOUBLE)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, &metricsview.Query{
		Measures: []metricsview.Measure{{Name: "cnt"}},
		Where: &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator:    metricsview.OperatorEq,
				Expressions: []*metricsview.Expression{{Name: "region"}, {Value: "APAC"}},
			},
		},
	}, starrocks.DialectStarRocks)
	require.NoError(t, err)

	sql, args, err := ast.SQL()
	require.NoError(t, err)
	require.Contains(t, sql, "WHERE (datetime = 'latest' OR datetime = 'current') AND (((`region`) = ?))")
	require.Equal(t, []any{"APAC"}, args)
}

func ptrInt64(v int64) *int64 {
	return &v
}
