package executor_test

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

const dedupeDimensionsTestMetricsView = `
type: metrics_view
version: 1
model: events
timeseries: ts
dimensions:
  - name: ts
    column: ts
measures:
  - name: events
    expression: count(*)
`

// The time dimension detail view compared by the metrics view's own time dimension requests that dimension both plain and at the selected grain.
// The query must return one row per grain bucket with a single "ts" column.
// Before the dedupe, DuckDB returned two "ts" columns with a row per raw timestamp, and ClickHouse rejected the duplicate alias.
func TestQueryTimeDimensionRequestedPlainAndFlooredDuckDB(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"events.sql": `
SELECT ts FROM (VALUES (TIMESTAMP '2024-01-05'), (TIMESTAMP '2024-01-20'), (TIMESTAMP '2024-02-03'), (TIMESTAMP '2024-03-01'), (TIMESTAMP '2024-03-15'), (TIMESTAMP '2024-03-30')) AS t(ts)
`,
			"mv.yaml": dedupeDimensionsTestMetricsView,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)
	requireTimeDimensionRequestedPlainAndFlooredRows(t, rt, instanceID)
}

func TestQueryTimeDimensionRequestedPlainAndFlooredClickHouse(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"events.sql": `
-- @connector: clickhouse
select toDateTime('2024-01-05 00:00:00', 'UTC') as ts
union all select toDateTime('2024-01-20 00:00:00', 'UTC')
union all select toDateTime('2024-02-03 00:00:00', 'UTC')
union all select toDateTime('2024-03-01 00:00:00', 'UTC')
union all select toDateTime('2024-03-15 00:00:00', 'UTC')
union all select toDateTime('2024-03-30 00:00:00', 'UTC')
`,
			"mv.yaml": dedupeDimensionsTestMetricsView,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)
	requireTimeDimensionRequestedPlainAndFlooredRows(t, rt, instanceID)
}

// requireTimeDimensionRequestedPlainAndFlooredRows runs the query the time dimension detail view sends when compared by the time dimension itself,
// and checks that it returns the monthly buckets sorted by the measure.
func requireTimeDimensionRequestedPlainAndFlooredRows(t *testing.T, rt *runtime.Runtime, instanceID string) {
	t.Helper()
	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	ctx := context.Background()
	e, err := executor.New(ctx, rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	limit := int64(10000)
	res, err := e.Query(ctx, &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "ts"},
			{Name: "ts", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainMonth}}},
		},
		Measures: []metricsview.Measure{{Name: "events"}},
		Sort:     []metricsview.Sort{{Name: "events", Desc: true}, {Name: "ts"}},
		Limit:    &limit,
	}, nil)
	require.NoError(t, err)
	defer res.Close()

	var columns []string
	for _, f := range res.Schema.Fields {
		columns = append(columns, f.Name)
	}
	require.Equal(t, []string{"ts", "events"}, columns)

	want := []struct {
		ts     time.Time
		events int64
	}{
		{ts: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), events: 3},
		{ts: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), events: 2},
		{ts: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), events: 1},
	}
	for _, w := range want {
		require.True(t, res.Next())
		row := make(map[string]any)
		require.NoError(t, res.MapScan(row))
		require.Len(t, row, 2)
		ts, ok := row["ts"].(time.Time)
		require.True(t, ok, "ts is %T", row["ts"])
		require.Equal(t, w.ts, ts.UTC())
		require.EqualValues(t, w.events, row["events"])
	}
	require.False(t, res.Next())
	require.NoError(t, res.Err())
}
