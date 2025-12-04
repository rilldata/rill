package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestMetricsViewTimeCaseInsensitive(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"m1.sql": `SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS TiMe, 1 AS num`,
			"mv1.yaml": `
type: metrics_view
model: m1
timeseries: TiMe
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv2.yaml": `
type: metrics_view
model: m1
timeseries: TiMe
dimensions:
- column: TiMe
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv3.yaml": `
type: metrics_view
model: m1
timeseries: time
dimensions:
- name: time
  column: TiMe
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv4.yaml": `
type: metrics_view
model: m1
timeseries: time
dimensions:
- column: TiMe
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv5.yaml": `
type: metrics_view
model: m1
timeseries: TiMe
dimensions:
- name: time
  column: TiMe
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 5, 1, 2)

	r := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.Empty(t, r.Meta.ReconcileError)
	d := r.GetMetricsView().State.ValidSpec.Dimensions[0]
	require.Equal(t, runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, d.Type)
	require.Equal(t, runtimev1.Type_CODE_TIMESTAMP, d.DataType.Code)

	r = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.Empty(t, r.Meta.ReconcileError)
	d = r.GetMetricsView().State.ValidSpec.Dimensions[0]
	require.Equal(t, runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, d.Type)
	require.Equal(t, runtimev1.Type_CODE_TIMESTAMP, d.DataType.Code)

	r = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv3")
	require.Empty(t, r.Meta.ReconcileError)
	d = r.GetMetricsView().State.ValidSpec.Dimensions[0]
	require.Equal(t, runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, d.Type)
	require.Equal(t, runtimev1.Type_CODE_TIMESTAMP, d.DataType.Code)

	testruntime.RequireParseErrors(t, rt, id, map[string]string{
		"/mv4.yaml": "does not match the case of time dimension",
		"/mv5.yaml": "does not match the case of time dimension",
	})
}
