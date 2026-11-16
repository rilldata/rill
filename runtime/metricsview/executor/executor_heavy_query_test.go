package executor_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// Queries that read past the heavy query row threshold (offset plus limit) must hold a concurrency slot until their result is closed,
// so that a caller paginating deep into a large result in parallel (as an agent extracting data page by page does) cannot exhaust the project.
// Smaller queries, like those issued by dashboards, are never limited.
func TestHeavyQueryConcurrency(t *testing.T) {
	rt, instanceID, mv := newHeavyQueryTestInstance(t, 20)

	// acquireSlot holds a heavy query slot for the user, failing if one isn't free right away.
	acquireSlot := func(userID string) (func(), error) {
		return rt.AcquireHeavyQuerySlot(t.Context(), instanceID, userID, 2, 1, 0)
	}

	// u1 has a heavy query running.
	releaseU1, err := acquireSlot("u1")
	require.NoError(t, err)

	// Another heavy query from u1 waits for its slot (offset plus limit is 30, above the threshold of 20).
	u1 := newHeavyQueryTestExecutor(t, rt, instanceID, mv, map[string]any{"id": "u1"})
	_, err = heavyQueryTestRows(t, u1, heavyQueryTestQuery(10, ptr(int64(20))), 200*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// So does one without a limit, which counts as reading its offset plus the interactive row cap of 20.
	_, err = heavyQueryTestRows(t, u1, heavyQueryTestQuery(10, nil), 200*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// Queries at or below the threshold run right away, including ones without a limit or offset.
	n, err := heavyQueryTestRows(t, u1, heavyQueryTestQuery(0, ptr(int64(20))), 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 20, n)
	n, err = heavyQueryTestRows(t, u1, &metricsview.Query{MetricsView: "mv1", Measures: []metricsview.Measure{{Name: "count"}}}, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	// Another user still gets the second project slot, and holds it until the result is closed.
	u2 := newHeavyQueryTestExecutor(t, rt, instanceID, mv, map[string]any{"id": "u2"})
	res, err := u2.Query(t.Context(), heavyQueryTestQuery(10, ptr(int64(20))), nil)
	require.NoError(t, err)
	_, err = acquireSlot("u3")
	require.ErrorIs(t, err, runtime.ErrTooManyConcurrentQueries)
	require.NoError(t, res.Close())
	releaseU3, err := acquireSlot("u3")
	require.NoError(t, err)
	releaseU3()

	// Embeds only count toward the project limit, since every viewer of an embed can share the identity of the user who issued it:
	// an embed with u1's identity runs while u1's slot is taken, and holds the second project slot.
	embed := newHeavyQueryTestExecutor(t, rt, instanceID, mv, map[string]any{"id": "u1", "embed": true})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	res, err = embed.Query(ctx, heavyQueryTestQuery(10, ptr(int64(20))), nil)
	require.NoError(t, err)
	_, err = acquireSlot("u3")
	require.ErrorIs(t, err, runtime.ErrTooManyConcurrentQueries)
	require.NoError(t, res.Close())

	// Once u1's running query finishes, its next heavy query runs and returns the requested page.
	releaseU1()
	n, err = heavyQueryTestRows(t, u1, heavyQueryTestQuery(40, ptr(int64(20))), 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 10, n)

	// A failed heavy query releases its slot: with u1's slots free again, two heavy queries from different users can run.
	_, err = heavyQueryTestRows(t, u1, &metricsview.Query{MetricsView: "mv1", Dimensions: []metricsview.Dimension{{Name: "missing"}}, Offset: ptr(int64(30))}, 10*time.Second)
	require.Error(t, err)
	releaseU1, err = acquireSlot("u1")
	require.NoError(t, err)
	releaseU1()
}

// Without a limit or an interactive row cap, a query reads every row past its offset, so it is heavy as soon as it has an offset.
func TestHeavyQueryConcurrencyWithoutRowCap(t *testing.T) {
	rt, instanceID, mv := newHeavyQueryTestInstance(t, 0)

	// u1 has a heavy query running.
	release, err := rt.AcquireHeavyQuerySlot(t.Context(), instanceID, "u1", 2, 1, 0)
	require.NoError(t, err)
	defer release()

	// A query from u1 that skips a single row waits for u1's slot.
	u1 := newHeavyQueryTestExecutor(t, rt, instanceID, mv, map[string]any{"id": "u1"})
	_, err = heavyQueryTestRows(t, u1, heavyQueryTestQuery(1, nil), 200*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// Without an offset, it runs right away and returns every row.
	n, err := heavyQueryTestRows(t, u1, heavyQueryTestQuery(0, nil), 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 50, n)
}

// newHeavyQueryTestInstance creates an instance with a metrics view over 50 dimension values,
// a heavy query threshold of 20 rows, a limit of 2 heavy queries per project and 1 per user, and the given interactive row cap.
func newHeavyQueryTestInstance(t *testing.T, rowCap int) (*runtime.Runtime, string, *runtimev1.MetricsViewSpec) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    "SELECT range AS id, 'v' || (range % 50)::VARCHAR AS dim FROM range(100)",
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
  - name: dim
    column: dim
measures:
  - name: count
    expression: count(*)
`,
		},
		Variables: map[string]string{
			"rill.interactive_sql_row_limit":                strconv.Itoa(rowCap),
			"rill.metrics.heavy_query_row_threshold":        "20",
			"rill.metrics.heavy_query_concurrency":          "2",
			"rill.metrics.heavy_query_concurrency_per_user": "1",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv1")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)
	return rt, instanceID, mv
}

func newHeavyQueryTestExecutor(t *testing.T, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, userAttrs map[string]any) *executor.Executor {
	e, err := executor.New(t.Context(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, userAttrs)
	require.NoError(t, err)
	t.Cleanup(e.Close)
	return e
}

func heavyQueryTestQuery(offset int64, limit *int64) *metricsview.Query {
	return &metricsview.Query{
		MetricsView: "mv1",
		Dimensions:  []metricsview.Dimension{{Name: "dim"}},
		Measures:    []metricsview.Measure{{Name: "count"}},
		Sort:        []metricsview.Sort{{Name: "dim"}},
		Offset:      &offset,
		Limit:       limit,
	}
}

// heavyQueryTestRows runs the query and returns the number of rows, or an error if it can't start within the timeout.
func heavyQueryTestRows(t *testing.T, e *executor.Executor, qry *metricsview.Query, timeout time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()
	res, err := e.Query(ctx, qry, nil)
	if err != nil {
		return 0, err
	}
	defer res.Close()
	n := 0
	for res.Next() {
		n++
	}
	return n, res.Err()
}

func ptr[T any](v T) *T {
	return &v
}
