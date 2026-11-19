package queries_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// The exact expression is deliberately different from the default one (rather than a real approximate/exact pair),
// so that the tests can tell which expression was used.
var exactMeasuresTestFiles = map[string]string{
	"rill.yaml": "",
	"m1.sql": `SELECT * FROM (VALUES
  (TIMESTAMP '2024-01-01 00:00:00', 'a'),
  (TIMESTAMP '2024-01-02 00:00:00', 'b'),
  (TIMESTAMP '2024-01-02 00:00:00', 'a')
) t(ts, publisher)`,
	"mv1.yaml": `
type: metrics_view
model: m1
timeseries: ts
dimensions:
  - name: publisher
    column: publisher
measures:
  - name: records
    expression: count(*)
    exact_expression: count(*) + 100
`,
}

func TestMetricsViewAggregation_exact(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: exactMeasuresTestFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	tests := []struct {
		name     string
		measures []*runtimev1.MetricsViewAggregationMeasure
		want     float64
	}{
		{name: "approximate by default", measures: []*runtimev1.MetricsViewAggregationMeasure{{Name: "records"}}, want: 3},
		{name: "measure-level flag", measures: []*runtimev1.MetricsViewAggregationMeasure{{Name: "records", Exact: true}}, want: 103},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queries.MetricsViewAggregation{
				MetricsViewName: "mv1",
				Measures:        tt.measures,
				SecurityClaims:  testClaims(),
			}
			require.NoError(t, q.Resolve(t.Context(), rt, instanceID, 0))
			require.Len(t, q.Result.Data, 1)
			require.Equal(t, tt.want, q.Result.Data[0].Fields["records"].GetNumberValue())
		})
	}
}

func TestMetricsViewTimeSeries_exact(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: exactMeasuresTestFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	tests := []struct {
		name              string
		exactMeasureNames []string
		want              []float64 // per day, ordered by time
	}{
		{name: "approximate by default", want: []float64{1, 2}},
		{name: "exact measure names", exactMeasureNames: []string{"records"}, want: []float64{101, 102}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queries.MetricsViewTimeSeries{
				MetricsViewName:   "mv1",
				MeasureNames:      []string{"records"},
				ExactMeasureNames: tt.exactMeasureNames,
				TimeGranularity:   runtimev1.TimeGrain_TIME_GRAIN_DAY,
				SecurityClaims:    testClaims(),
			}
			require.NoError(t, q.Resolve(t.Context(), rt, instanceID, 0))
			require.Len(t, q.Result.Data, len(tt.want))
			for i, want := range tt.want {
				require.Equal(t, want, q.Result.Data[i].Records.Fields["records"].GetNumberValue())
			}
		})
	}
}
