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

// Queries that read past the heavy query row threshold (offset plus limit) must hold a concurrency slot until their result is closed,
// so that a caller paginating deep into a large result in parallel (as an agent extracting data page by page does) cannot exhaust the project.
// Smaller queries, like those issued by dashboards, are never limited.
func TestHeavyQueryConcurrency(t *testing.T) {
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
			"rill.metrics.heavy_query_row_threshold":        "20",
			"rill.metrics.heavy_query_concurrency":          "2",
			"rill.metrics.heavy_query_concurrency_per_user": "1",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	r := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv1")
	mv := r.GetMetricsView().State.ValidSpec
	require.NotNil(t, mv)

	newExecutor := func(userID string) *executor.Executor {
		e, err := executor.New(t.Context(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, map[string]any{"id": userID})
		require.NoError(t, err)
		t.Cleanup(e.Close)
		return e
	}
	newQuery := func(offset, limit int64) *metricsview.Query {
		return &metricsview.Query{
			MetricsView: "mv1",
			Dimensions:  []metricsview.Dimension{{Name: "dim"}},
			Measures:    []metricsview.Measure{{Name: "count"}},
			Sort:        []metricsview.Sort{{Name: "dim"}},
			Offset:      &offset,
			Limit:       &limit,
		}
	}
	// queryRows runs the query and returns the number of rows, or an error if it can't start within the timeout.
	queryRows := func(e *executor.Executor, qry *metricsview.Query, timeout time.Duration) (int, error) {
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
	// acquireSlot holds a heavy query slot for the user, failing if one isn't free right away.
	acquireSlot := func(userID string) (func(), error) {
		return rt.AcquireHeavyQuerySlot(t.Context(), instanceID, userID, 2, 1, 0)
	}

	// u1 has a heavy query running.
	releaseU1, err := acquireSlot("u1")
	require.NoError(t, err)

	// Another heavy query from u1 waits for its slot (offset plus limit is 30, above the threshold of 20).
	u1 := newExecutor("u1")
	_, err = queryRows(u1, newQuery(10, 20), 200*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// Queries at or below the threshold run right away.
	n, err := queryRows(u1, newQuery(0, 20), 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 20, n)

	// Another user still gets the second project slot, and holds it until the result is closed.
	u2 := newExecutor("u2")
	res, err := u2.Query(t.Context(), newQuery(10, 20), nil)
	require.NoError(t, err)
	_, err = acquireSlot("u3")
	require.ErrorIs(t, err, runtime.ErrTooManyConcurrentQueries)
	require.NoError(t, res.Close())
	releaseU3, err := acquireSlot("u3")
	require.NoError(t, err)
	releaseU3()

	// Once u1's running query finishes, its next heavy query runs and returns the requested page.
	releaseU1()
	n, err = queryRows(u1, newQuery(40, 20), 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, 10, n)

	// A failed heavy query releases its slot: with u1's slots free again, two heavy queries from different users can run.
	_, err = queryRows(u1, &metricsview.Query{MetricsView: "mv1", Dimensions: []metricsview.Dimension{{Name: "missing"}}, Offset: ptr(int64(30))}, 10*time.Second)
	require.Error(t, err)
	releaseU1, err = acquireSlot("u1")
	require.NoError(t, err)
	releaseU1()
}

func ptr[T any](v T) *T {
	return &v
}
