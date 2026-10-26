package executor_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// A comparison leaderboard on a UUID dimension with a Decimal measure must take the two-phase path:
// the base results are inlined with a cast back to their ClickHouse types, which the second phase joins on and filters the comparison range with.
// Before the cast, UUID and Decimal results were rejected as not inlinable and the query silently aggregated the whole comparison range instead.
func TestRewriteTwoPhaseComparisonsUUIDDimension(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			// 20 users with five events on each of ten days; users 9 and 19 have a NULL id, and the price makes every user's revenue distinct.
			"events.sql": `
-- @connector: clickhouse
select
  toDateTime('2024-01-01 00:00:00') + interval (number % 10) day as event_time,
  if(intDiv(number, 10) % 10 = 9, NULL, toUUID(UUIDNumToString(MD5(toString(intDiv(number, 10) % 20))))) as user_id,
  toDecimal64(intDiv(number, 10) % 20 + 0.25 + (number % 10) * 0.1, 2) as price
from numbers(1000)
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: events
timeseries: event_time
dimensions:
  - name: user_id
    column: user_id
measures:
  - name: events
    expression: count(*)
  - name: revenue
    expression: sum(price)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	ctx := context.Background()
	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	require.NoError(t, err)
	defer release()

	e, err := executor.New(ctx, rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	limit := int64(5)
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{{Name: "user_id"}},
		Measures: []metricsview.Measure{
			{Name: "events"},
			{Name: "revenue"},
			{Name: "events__prev", Compute: &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{Measure: "events"}}},
			{Name: "revenue__prev", Compute: &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{Measure: "revenue"}}},
			{Name: "revenue__delta", Compute: &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{Measure: "revenue"}}},
		},
		Sort:                []metricsview.Sort{{Name: "revenue", Desc: true}},
		TimeRange:           &metricsview.TimeRange{Start: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)},
		ComparisonTimeRange: &metricsview.TimeRange{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
		Limit:               &limit,
	}

	res, err := e.Query(ctx, qry, nil)
	require.NoError(t, err)
	got := twoPhaseTestResultRows(t, res)

	// The same leaderboard as a single query, with the limit applied to the base range only.
	table := olap.Dialect().EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table)
	want := twoPhaseTestRows(t, olap, fmt.Sprintf(`
		SELECT b.user_id AS user_id, b.events AS events, b.revenue AS revenue, c.events AS events__prev, c.revenue AS revenue__prev, b.revenue - c.revenue AS revenue__delta
		FROM (SELECT user_id, count(*) AS events, sum(price) AS revenue FROM %[1]s WHERE event_time >= toDateTime('2024-01-06 00:00:00') AND event_time < toDateTime('2024-01-11 00:00:00') GROUP BY user_id ORDER BY revenue DESC LIMIT 5) b
		LEFT JOIN (SELECT user_id, count(*) AS events, sum(price) AS revenue FROM %[1]s WHERE event_time >= toDateTime('2024-01-01 00:00:00') AND event_time < toDateTime('2024-01-06 00:00:00') GROUP BY user_id) c
		ON isNotDistinctFrom(b.user_id, c.user_id)
		ORDER BY revenue DESC`, table), nil)
	require.Len(t, want, 5)
	require.Equal(t, "<nil>", want[0]["user_id"], "the NULL user must be in the leaderboard so that the rewrite's IS NULL branch is exercised")
	require.Equal(t, want, got)

	// The second phase must have run with the base results inlined, cast back to the dimension's and measure's types.
	err = olap.Exec(ctx, &drivers.Statement{Query: "SYSTEM FLUSH LOGS"})
	require.NoError(t, err)
	logged := twoPhaseTestRows(t, olap, "SELECT count() AS n FROM system.query_log WHERE type = 'QueryFinish' AND query LIKE '%FROM VALUES%' AND query LIKE '%CAST(%AS Nullable(UUID))%' AND query LIKE '%CAST(%AS Decimal(38, 2))%'", nil)
	require.Equal(t, "1", logged[0]["n"])
}

func twoPhaseTestRows(t *testing.T, olap drivers.OLAPStore, sql string, args []any) []map[string]string {
	t.Helper()
	res, err := olap.Query(context.Background(), &drivers.Statement{Query: sql, Args: args})
	require.NoError(t, err)
	return twoPhaseTestResultRows(t, res)
}

// twoPhaseTestResultRows returns the rows with every value formatted as a string, because VALUES infers the narrowest integer type for inlined counts.
func twoPhaseTestResultRows(t *testing.T, res *drivers.Result) []map[string]string {
	t.Helper()
	defer res.Close()
	var rows []map[string]string
	for res.Next() {
		row := make(map[string]any)
		require.NoError(t, res.MapScan(row))
		formatted := make(map[string]string, len(row))
		for k, v := range row {
			formatted[k] = fmt.Sprint(v)
		}
		rows = append(rows, formatted)
	}
	require.NoError(t, res.Err())
	return rows
}
