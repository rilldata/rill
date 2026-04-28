package reconcilers_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestLegacyAlert(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: MetricsView
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

	_, metricsRes := newMetricsView("mv1", "bar", "__time", []any{"count(*)", runtimev1.Type_CODE_INT64}, []any{"country", runtimev1.Type_CODE_STRING})
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
					DisplayName:          "Test Alert",
					RefreshSchedule:      &runtimev1.Schedule{RefUpdate: true},
					WatermarkInherit:     true,
					IntervalsIsoDuration: "P1D",
					Resolver:             "legacy_metrics",
					ResolverProperties: must(structpb.NewStruct(map[string]any{
						"query_name":      "MetricsViewAggregation",
						"query_args_json": "{\"dimensions\":[{\"name\":\"country\"}],\"having\":{\"cond\":{\"exprs\":[{\"ident\":\"measure_0\"},{\"val\":4}],\"op\":\"OPERATION_GTE\"}},\"measures\":[{\"name\":\"measure_0\"}],\"metrics_view\":\"mv1\",\"time_range\":{\"iso_duration\":\"P1W\"}}",
					})),
					NotifyOnRecover:      false,
					NotifyOnFail:         true,
					NotifyOnError:        false,
					Renotify:             false,
					RenotifyAfterSeconds: 0,
					Notifiers: []*runtimev1.Notifier{{
						Connector:  "email",
						Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"somebody@example.com"}})),
					}},
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
					"Measure 0": "4",
				})),
			},
			SentNotifications: true,
			ExecutionTime:     timestamppb.New(time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)),
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Check that the alert was sent
	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.Equal(t, emails[0].ToEmail, "somebody@example.com")
	require.Contains(t, emails[0].Body, "Denmark")
}

func TestAlert(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: MetricsView
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
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	_, metricsRes := newMetricsView("mv1", "bar", "__time", []any{"count(*)", runtimev1.Type_CODE_INT64}, []any{"country", runtimev1.Type_CODE_STRING})
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
					DisplayName:          "Test Alert",
					RefreshSchedule:      &runtimev1.Schedule{RefUpdate: true},
					WatermarkInherit:     true,
					IntervalsIsoDuration: "P1D",
					Resolver:             "legacy_metrics",
					ResolverProperties: must(structpb.NewStruct(map[string]any{
						"query_name":      "MetricsViewAggregation",
						"query_args_json": "{\"dimensions\":[{\"name\":\"country\"}],\"having\":{\"cond\":{\"exprs\":[{\"ident\":\"measure_0\"},{\"val\":4}],\"op\":\"OPERATION_GTE\"}},\"measures\":[{\"name\":\"measure_0\"}],\"metrics_view\":\"mv1\",\"time_range\":{\"iso_duration\":\"P1W\"}}",
					})),
					NotifyOnRecover:      false,
					NotifyOnFail:         true,
					NotifyOnError:        false,
					Renotify:             false,
					RenotifyAfterSeconds: 0,
					Notifiers:            []*runtimev1.Notifier{{Connector: "email", Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"somebody@example.com"}}))}},
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
					"Measure 0": "4",
				})),
			},
			SentNotifications: true,
			ExecutionTime:     timestamppb.New(time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)),
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Check that the alert was sent
	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.Equal(t, emails[0].ToEmail, "somebody@example.com")
	require.Contains(t, emails[0].Body, "Denmark")
}

func TestAlertDataYAMLSQL(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
-- @materialize: true
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: Model
  name: bar
watermark: inherit
intervals:
  duration: P1D
  check_unclosed: true
data:
  sql: |-
    select * from bar where country <> 'Denmark'
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	a1 := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: "a1"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: "bar"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/alerts/a1.yaml"},
		},
		Resource: &runtimev1.Resource_Alert{
			Alert: &runtimev1.Alert{
				Spec: &runtimev1.AlertSpec{
					DisplayName:            "Test Alert",
					RefreshSchedule:        &runtimev1.Schedule{RefUpdate: true},
					WatermarkInherit:       true,
					IntervalsIsoDuration:   "P1D",
					IntervalsCheckUnclosed: true,
					Resolver:               "sql",
					ResolverProperties:     must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select * from bar where country <> 'Denmark'"})),
					NotifyOnRecover:        false,
					NotifyOnFail:           true,
					NotifyOnError:          false,
					Renotify:               false,
					RenotifyAfterSeconds:   0,
					Notifiers:              []*runtimev1.Notifier{{Connector: "email", Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"somebody@example.com"}}))}},
				},
				State: &runtimev1.AlertState{
					ExecutionCount: 1,
					ExecutionHistory: []*runtimev1.AlertExecution{
						{
							Result:        &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS},
							ExecutionTime: nil,
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
	as1.ExecutionCount = 2
	as1.ExecutionHistory = slices.Insert(as1.ExecutionHistory, 0,
		&runtimev1.AlertExecution{
			Result:        &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS},
			ExecutionTime: nil,
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
SELECT '2024-01-04T00:00:00Z'::TIMESTAMP as __time, 'Sweden' as country
	`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	as1.ExecutionCount = 3
	as1.ExecutionHistory = slices.Insert(as1.ExecutionHistory, 0,
		&runtimev1.AlertExecution{
			Result: &runtimev1.AssertionResult{
				Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
				FailRow: must(structpb.NewStruct(map[string]any{
					"country": "Sweden",
					"__time":  "2024-01-04T00:00:00Z",
				})),
			},
			SentNotifications: true,
			ExecutionTime:     nil,
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Check that the alert was sent
	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.Equal(t, emails[0].ToEmail, "somebody@example.com")
	require.Contains(t, emails[0].Body, "Sweden")
}

func TestAlertDataYAMLMetricsSQL(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/alerts/a1.yaml": `
type: alert
display_name: Test Alert
refs:
- type: MetricsView
  name: mv1
watermark: inherit
intervals:
  duration: P1D
data:
  metrics_sql: |-
    select measure_0 from mv1 where country <> 'Denmark' having measure_0 > 0
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	_, metricsRes := newMetricsView("mv1", "bar", "__time", []any{"count(*)", runtimev1.Type_CODE_INT64}, []any{"country", runtimev1.Type_CODE_STRING})
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
					DisplayName:          "Test Alert",
					RefreshSchedule:      &runtimev1.Schedule{RefUpdate: true},
					WatermarkInherit:     true,
					IntervalsIsoDuration: "P1D",
					Resolver:             "metrics_sql",
					ResolverProperties:   must(structpb.NewStruct(map[string]any{"sql": "select measure_0 from mv1 where country <> 'Denmark' having measure_0 > 0"})),
					NotifyOnRecover:      false,
					NotifyOnFail:         true,
					NotifyOnError:        false,
					Renotify:             false,
					RenotifyAfterSeconds: 0,
					Notifiers:            []*runtimev1.Notifier{{Connector: "email", Properties: must(structpb.NewStruct(map[string]any{"recipients": []any{"somebody@example.com"}}))}},
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

	// Add data for another day such that the assertion fails
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
UNION ALL
SELECT '2024-01-02T00:00:00Z'::TIMESTAMP as __time, 'Sweden' as country
	`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	as1.ExecutionCount = 2
	as1.ExecutionHistory = slices.Insert(as1.ExecutionHistory, 0,
		&runtimev1.AlertExecution{
			Result: &runtimev1.AssertionResult{
				Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
				FailRow: must(structpb.NewStruct(map[string]any{
					"measure_0": 1,
				})),
			},
			SentNotifications: true,
			ExecutionTime:     timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
	)
	testruntime.RequireResource(t, rt, id, a1)

	// Check that the alert was sent
	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.Equal(t, emails[0].ToEmail, "somebody@example.com")
	require.Contains(t, emails[0].Body, "measure_0")
}

func newMetricsView(name, model, timeDim string, measures, dimensions []any) (*runtimev1.MetricsView, *runtimev1.Resource) {
	metrics := &runtimev1.MetricsView{
		Spec: &runtimev1.MetricsViewSpec{
			Connector:     "duckdb",
			Model:         model,
			DisplayName:   parser.ToDisplayName(name),
			TimeDimension: timeDim,
			Measures:      make([]*runtimev1.MetricsViewSpec_Measure, len(measures)/2),
			Dimensions:    make([]*runtimev1.MetricsViewSpec_Dimension, len(dimensions)/2),
		},
		State: &runtimev1.MetricsViewState{
			ValidSpec: &runtimev1.MetricsViewSpec{
				Connector:         "duckdb",
				Table:             model,
				Model:             model,
				DisplayName:       parser.ToDisplayName(name),
				TimeDimension:     timeDim,
				SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
				Measures:          make([]*runtimev1.MetricsViewSpec_Measure, len(measures)/2),
				Dimensions:        make([]*runtimev1.MetricsViewSpec_Dimension, len(dimensions)/2),
			},
		},
	}
	for i := range len(measures) / 2 {
		name := fmt.Sprintf("measure_%d", i)
		idx := i * 2
		expr := measures[idx].(string)
		metrics.Spec.Measures[i] = &runtimev1.MetricsViewSpec_Measure{
			Name:        name,
			DisplayName: parser.ToDisplayName(name),
			Expression:  expr,
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
		}
		metrics.State.ValidSpec.Measures[i] = &runtimev1.MetricsViewSpec_Measure{
			Name:        name,
			DisplayName: parser.ToDisplayName(name),
			Expression:  expr,
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DataType:    &runtimev1.Type{Code: measures[idx+1].(runtimev1.Type_Code), Nullable: true},
		}
	}
	for i := range len(dimensions) / 2 {
		idx := i * 2
		name := dimensions[idx].(string)
		metrics.Spec.Dimensions[i] = &runtimev1.MetricsViewSpec_Dimension{
			Name:        name,
			DisplayName: parser.ToDisplayName(name),
			Column:      name,
		}
		metrics.State.ValidSpec.Dimensions[i] = &runtimev1.MetricsViewSpec_Dimension{
			Name:        name,
			DisplayName: parser.ToDisplayName(name),
			Column:      name,
			DataType:    &runtimev1.Type{Code: dimensions[idx+1].(runtimev1.Type_Code), Nullable: true},
			Type:        runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL,
		}
	}
	// prepend the time dimension to metrics.Spec.Dimensions and metrics.State.ValidSpec.Dimensions
	metrics.Spec.Dimensions = slices.Insert(metrics.Spec.Dimensions, 0, &runtimev1.MetricsViewSpec_Dimension{
		Name:        timeDim,
		Column:      timeDim,
		DisplayName: parser.ToDisplayName(timeDim),
	})
	metrics.State.ValidSpec.Dimensions = slices.Insert(metrics.State.ValidSpec.Dimensions, 0, &runtimev1.MetricsViewSpec_Dimension{
		Name:              timeDim,
		Column:            timeDim,
		DisplayName:       parser.ToDisplayName(timeDim),
		DataType:          &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true},
		Type:              runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME,
		SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
	})

	metricsRes := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: model}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{fmt.Sprintf("/metrics/%s.yaml", name)},
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
