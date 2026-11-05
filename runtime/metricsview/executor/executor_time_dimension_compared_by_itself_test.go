package executor_test

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

const timeDimensionComparedByItselfMetricsView = `
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

// The time dimension detail view compared by the metrics view's own time dimension needs both the raw timestamps, for its rows,
// and their time floor, for its columns, so it requests the floor under an alias.
// A query that requests both under one name is rejected:
// ClickHouse fails on its duplicate alias, and DuckDB returns two "ts" columns where one overwrites the other.
// Repeating the plain dimension is allowed, since both columns hold the same values.
func TestQueryTimeDimensionComparedByItselfDuckDB(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"events.sql": `
SELECT ts FROM (VALUES (TIMESTAMP '2024-01-05'), (TIMESTAMP '2024-01-20'), (TIMESTAMP '2024-02-03'), (TIMESTAMP '2024-03-01'), (TIMESTAMP '2024-03-15'), (TIMESTAMP '2024-03-30')) AS t(ts)
`,
			"mv.yaml": timeDimensionComparedByItselfMetricsView,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)
	requireTimeDimensionComparedByItselfRows(t, rt, instanceID)
}

func TestQueryTimeDimensionComparedByItselfClickHouse(t *testing.T) {
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
			"mv.yaml": timeDimensionComparedByItselfMetricsView,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)
	requireTimeDimensionComparedByItselfRows(t, rt, instanceID)
}

func requireTimeDimensionComparedByItselfRows(t *testing.T, rt *runtime.Runtime, instanceID string) {
	t.Helper()
	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	ctx := context.Background()
	e, err := executor.New(ctx, rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	monthFloor := &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainMonth}}
	day := func(month, day int) time.Time { return time.Date(2024, time.Month(month), day, 0, 0, 0, 0, time.UTC) }

	t.Run("the floor under an alias keeps a row per raw timestamp", func(t *testing.T) {
		limit := int64(10000)
		res, err := e.Query(ctx, &metricsview.Query{
			Dimensions: []metricsview.Dimension{{Name: "ts"}, {Name: "ts_rill_TIME_GRAIN_MONTH", Compute: monthFloor}},
			Measures:   []metricsview.Measure{{Name: "events"}},
			Sort:       []metricsview.Sort{{Name: "events", Desc: true}, {Name: "ts_rill_TIME_GRAIN_MONTH"}, {Name: "ts"}},
			Limit:      &limit,
		}, nil)
		require.NoError(t, err)
		requireTimeDimensionRows(t, res, []string{"ts", "ts_rill_TIME_GRAIN_MONTH", "events"}, []map[string]any{
			{"ts": day(1, 5), "ts_rill_TIME_GRAIN_MONTH": day(1, 1), "events": 1},
			{"ts": day(1, 20), "ts_rill_TIME_GRAIN_MONTH": day(1, 1), "events": 1},
			{"ts": day(2, 3), "ts_rill_TIME_GRAIN_MONTH": day(2, 1), "events": 1},
			{"ts": day(3, 1), "ts_rill_TIME_GRAIN_MONTH": day(3, 1), "events": 1},
			{"ts": day(3, 15), "ts_rill_TIME_GRAIN_MONTH": day(3, 1), "events": 1},
			{"ts": day(3, 30), "ts_rill_TIME_GRAIN_MONTH": day(3, 1), "events": 1},
		})
	})

	t.Run("the floor under the dimension's own name is rejected", func(t *testing.T) {
		_, err := e.Query(ctx, &metricsview.Query{
			Dimensions: []metricsview.Dimension{{Name: "ts"}, {Name: "ts", Compute: monthFloor}},
			Measures:   []metricsview.Measure{{Name: "events"}},
		}, nil)
		require.ErrorContains(t, err, `dimension "ts" is requested more than once with different computations`)
	})

	t.Run("the plain dimension repeated is allowed", func(t *testing.T) {
		limit := int64(10000)
		res, err := e.Query(ctx, &metricsview.Query{
			Dimensions: []metricsview.Dimension{{Name: "ts"}, {Name: "ts"}},
			Measures:   []metricsview.Measure{{Name: "events"}},
			Sort:       []metricsview.Sort{{Name: "ts"}},
			Limit:      &limit,
		}, nil)
		require.NoError(t, err)
		requireTimeDimensionRows(t, res, []string{"ts", "ts", "events"}, []map[string]any{
			{"ts": day(1, 5), "events": 1},
			{"ts": day(1, 20), "events": 1},
			{"ts": day(2, 3), "events": 1},
			{"ts": day(3, 1), "events": 1},
			{"ts": day(3, 15), "events": 1},
			{"ts": day(3, 30), "events": 1},
		})
	})
}

// requireTimeDimensionRows checks the result's columns and rows in order, comparing timestamps in UTC and counts by value.
func requireTimeDimensionRows(t *testing.T, res *drivers.Result, wantColumns []string, wantRows []map[string]any) {
	t.Helper()
	defer res.Close()

	var columns []string
	for _, f := range res.Schema.Fields {
		columns = append(columns, f.Name)
	}
	require.Equal(t, wantColumns, columns)

	for _, want := range wantRows {
		require.True(t, res.Next())
		row := make(map[string]any)
		require.NoError(t, res.MapScan(row))
		require.Len(t, row, len(want))
		for k, v := range want {
			if ts, ok := v.(time.Time); ok {
				got, ok := row[k].(time.Time)
				require.True(t, ok, "%s is %T", k, row[k])
				require.Equal(t, ts, got.UTC())
				continue
			}
			require.EqualValues(t, v, row[k])
		}
	}
	require.False(t, res.Next())
	require.NoError(t, res.Err())
}
