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
		"/dashboards/bare.yaml": `
model: foo
measures:
- name: a
  expression: count(*)
`,
		"/dashboards/with_time_dimension.yaml": `
model: foo
timeseries: time
measures:
- name: a
  expression: count(*)
`,
		"/dashboards/with_watermark_expression.yaml": `
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
