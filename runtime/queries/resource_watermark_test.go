package queries_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestResourceWatermark_MetricsView(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2022-01-02T00:00:00Z")
	require.NoError(t, err)

	rt, instanceID := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"/models/foo.sql": fmt.Sprintf(`SELECT '%s'::TIMESTAMP as time`, ts.Add(-time.Second).Format(time.RFC3339)),
		"/metrics/bare.yaml": `
version: 1
type: metrics_view
model: foo
measures:
- name: a
  expression: count(*)
`,
		"/metrics/with_time_dimension.yaml": `
version: 1
type: metrics_view
model: foo
timeseries: time
measures:
- name: a
  expression: count(*)
`,
		"/metrics/with_watermark_expression.yaml": `
version: 1
type: metrics_view
model: foo
timeseries: time
watermark: max(time) - interval '1 day'
measures:
- name: a
  expression: count(*)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	// "bare" should have no watermark
	q := &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "bare",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Nil(t, q.Result)

	// "with_time_dimension" should have a watermark of the max time
	q = &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "with_time_dimension",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, ts, *q.Result)

	// "with_watermark_expression" should have a watermark that's one day less than the max time
	q = &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "with_watermark_expression",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, ts.Add(-24*time.Hour), *q.Result)
}

func TestResourceWatermark_MetricsViewTableFilter(t *testing.T) {
	oldTS, err := time.Parse(time.RFC3339, "2022-01-10T00:00:00Z")
	require.NoError(t, err)
	latestTS, err := time.Parse(time.RFC3339, "2022-01-05T00:00:00Z")
	require.NoError(t, err)

	rt, instanceID := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"/models/snapshots.sql": fmt.Sprintf(`
SELECT '%s'::TIMESTAMP AS time, 'old' AS snapshot
UNION ALL
SELECT '%s'::TIMESTAMP AS time, 'latest' AS snapshot
`, oldTS.Add(-time.Second).Format(time.RFC3339), latestTS.Add(-time.Second).Format(time.RFC3339)),
		"/metrics/filtered.yaml": `
version: 1
type: metrics_view
model: snapshots
timeseries: time
table_filter: "snapshot = 'latest'"
measures:
- name: a
  expression: count(*)
`,
		"/metrics/filtered_expr.yaml": `
version: 1
type: metrics_view
model: snapshots
timeseries: time
table_filter: "snapshot = 'latest'"
watermark: max(time) - interval '1 day'
measures:
- name: a
  expression: count(*)
`,
		"/metrics/filtered_empty.yaml": `
version: 1
type: metrics_view
model: snapshots
timeseries: time
table_filter: "snapshot = 'nonexistent'"
measures:
- name: a
  expression: count(*)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	q := &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "filtered",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotNil(t, q.Result)
	require.Equal(t, latestTS, *q.Result)

	q = &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "filtered_expr",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotNil(t, q.Result)
	require.Equal(t, latestTS.Add(-24*time.Hour), *q.Result)

	// Zero-row table_filter returns SQL NULL from MAX(...) and must resolve to nil without error.
	q = &queries.ResourceWatermark{
		ResourceKind: runtime.ResourceKindMetricsView,
		ResourceName: "filtered_empty",
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Nil(t, q.Result)
}
