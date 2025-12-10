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

func TestMetricsViewTimeTypes(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"m1.sql": `SELECT '2024-01-01'::DATE AS date, '2024-01-01T00:00:00Z'::TIMESTAMP AS time, 'foo' AS str, 1 AS num`,
			"mv_none.yaml": `
type: metrics_view
model: m1
dimensions:
- column: time
- column: date
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv_time.yaml": `
type: metrics_view
model: m1
timeseries: time
dimensions:
- column: time
- column: date
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv_date.yaml": `
type: metrics_view
model: m1
timeseries: date
dimensions:
- column: time
- column: date
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv_time_legacy.yaml": `
type: metrics_view
model: m1
timeseries: time
dimensions:
- column: date
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
			"mv_date_legacy.yaml": `
type: metrics_view
model: m1
timeseries: date
dimensions:
- column: time
measures:
- name: num
  expression: sum(num)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 7, 0, 0)

	// Expectations
	cases := []struct {
		metricsView string
		dimension   string
		typ         runtimev1.MetricsViewSpec_DimensionType
		dataTyp     runtimev1.Type_Code
	}{
		{"mv_none", "time", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_TIMESTAMP},
		{"mv_none", "date", runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL, runtimev1.Type_CODE_DATE},
		{"mv_time", "time", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_TIMESTAMP},
		{"mv_time", "date", runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL, runtimev1.Type_CODE_DATE},
		{"mv_date", "time", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_TIMESTAMP},
		{"mv_date", "date", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_DATE},
		{"mv_time_legacy", "time", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_TIMESTAMP},
		{"mv_time_legacy", "date", runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL, runtimev1.Type_CODE_DATE},
		{"mv_date_legacy", "time", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_TIMESTAMP},
		{"mv_date_legacy", "date", runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, runtimev1.Type_CODE_DATE},
	}
	for _, c := range cases {
		t.Run(c.metricsView+"_"+c.dimension, func(t *testing.T) {
			ctrl, err := rt.Controller(t.Context(), id)
			require.NoError(t, err)
			mv, err := ctrl.Get(t.Context(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: c.metricsView}, false)
			require.NoError(t, err)
			validSpec := mv.GetMetricsView().State.ValidSpec
			require.NotNil(t, validSpec)

			var found bool
			for _, d := range validSpec.Dimensions {
				if d.Name == c.dimension {
					found = true
					require.Equal(t, c.typ, d.Type)
					require.Equal(t, c.dataTyp, d.DataType.Code)
				}
			}
			require.True(t, found, "dimension %s not found in metrics view %s", c.dimension, c.metricsView)
		})
	}
}
