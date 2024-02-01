package reconcilers_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAlert(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/dashboards/mv1.yaml": `
title: mv1
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/alerts/a1.yaml": `
kind: alert
title: Test Alert
refs:
- kind: MetricsView
  name: mv1
watermark: inherit
intervals:
  duration: P1D
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: measure_0
    time_range:
      iso_duration: P1W
    having:
      cond:
        op: OPERATION_GTE
        exprs:
        - ident: measure_0
        - val: 4
email:
  recipients:
    - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	_, metricsRes := newMetricsView("mv1", "bar", "__time", []string{"count(*)"}, []string{"country"})
	testruntime.RequireResource(t, rt, id, metricsRes)

	a1 := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: "a1"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "mv1"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/alerts/a1.yaml"},
		},
		Resource: &runtimev1.Resource_Alert{
			Alert: &runtimev1.Alert{
				Spec: &runtimev1.AlertSpec{
					Title:                "Test Alert",
					RefreshSchedule:      &runtimev1.Schedule{RefUpdate: true},
					WatermarkInherit:     true,
					IntervalsIsoDuration: "P1D",
					QueryName:            "MetricsViewAggregation",
					QueryArgsJson:        "{\"dimensions\":[{\"name\":\"country\"}],\"having\":{\"cond\":{\"exprs\":[{\"ident\":\"measure_0\"},{\"val\":4}],\"op\":\"OPERATION_GTE\"}},\"measures\":[{\"name\":\"measure_0\"}],\"metrics_view\":\"mv1\",\"time_range\":{\"iso_duration\":\"P1W\"}}",
					EmailRecipients:      []string{"somebody@example.com"},
					EmailOnFail:          true,
				},
				State: &runtimev1.AlertState{
					ExecutionCount: 1,
					ExecutionHistory: []*runtimev1.AlertExecution{
						{
							Result:        &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS},
							ExecutionTime: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		},
	}
	testruntime.RequireResource(t, rt, id, a1)

	// Extract alert state
	as1 := a1.GetAlert().State

	// Add data for two more days, check it executes for each day (all passing)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-02T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-03T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
	`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	as1.ExecutionCount = 3
	as1.ExecutionHistory = slices.Insert(as1.ExecutionHistory, 0,
		&runtimev1.AlertExecution{
			Result:        &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS},
			ExecutionTime: timestamppb.New(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		&runtimev1.AlertExecution{
			Result:        &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS},
			ExecutionTime: timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Add data for another day such that the assertion fails
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-02T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-03T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-03T12:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-04T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
	`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	as1.ExecutionCount = 4
	as1.ExecutionHistory = slices.Insert(as1.ExecutionHistory, 0,
		&runtimev1.AlertExecution{
			Result: &runtimev1.AssertionResult{
				Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
				FailRow: must(structpb.NewStruct(map[string]any{
					"country":   "Denmark",
					"measure_0": 4,
				})),
			},
			SentEmails:    true,
			ExecutionTime: timestamppb.New(time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)),
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Check that the alert was sent
	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.Equal(t, emails[0].ToEmail, "somebody@example.com")
	require.Contains(t, emails[0].Body, "Denmark")
}

func newMetricsView(name, table, timeDim string, measures, dimensions []string) (*runtimev1.MetricsViewV2, *runtimev1.Resource) {
	metrics := &runtimev1.MetricsViewV2{
		Spec: &runtimev1.MetricsViewSpec{
			Connector:     "duckdb",
			Table:         table,
			Title:         name,
			TimeDimension: timeDim,
			Measures:      make([]*runtimev1.MetricsViewSpec_MeasureV2, len(measures)),
			Dimensions:    make([]*runtimev1.MetricsViewSpec_DimensionV2, len(dimensions)),
		},
		State: &runtimev1.MetricsViewState{
			ValidSpec: &runtimev1.MetricsViewSpec{
				Connector:     "duckdb",
				Table:         table,
				Title:         name,
				TimeDimension: timeDim,
				Measures:      make([]*runtimev1.MetricsViewSpec_MeasureV2, len(measures)),
				Dimensions:    make([]*runtimev1.MetricsViewSpec_DimensionV2, len(dimensions)),
			},
		},
	}
	for i, measure := range measures {
		metrics.Spec.Measures[i] = &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       fmt.Sprintf("measure_%d", i),
			Expression: measure,
		}
		metrics.State.ValidSpec.Measures[i] = &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       fmt.Sprintf("measure_%d", i),
			Expression: measure,
		}
	}
	for i, dimension := range dimensions {
		metrics.Spec.Dimensions[i] = &runtimev1.MetricsViewSpec_DimensionV2{
			Name:   dimension,
			Column: dimension,
		}
		metrics.State.ValidSpec.Dimensions[i] = &runtimev1.MetricsViewSpec_DimensionV2{
			Name:   dimension,
			Column: dimension,
		}
	}
	metricsRes := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: table}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{fmt.Sprintf("/dashboards/%s.yaml", name)},
		},
		Resource: &runtimev1.Resource_MetricsView{
			MetricsView: metrics,
		},
	}
	return metrics, metricsRes
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
