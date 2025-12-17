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

func TestTargetsQuery(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'DK' as country, 1 as val union all
select '2024-01-02T00:00:00Z'::TIMESTAMP as time, 'US' as country, 2 as val union all
select '2024-01-03T00:00:00Z'::TIMESTAMP as time, 'US' as country, 3 as val`,
			"targets.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'budget_2024' as target, 1000.0 as value
union all
select '2024-02-01T00:00:00Z'::TIMESTAMP as time, 'budget_2024' as target, 1200.0 as value
union all
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'budget_2025' as target, 1500.0 as value
union all
select '2024-02-01T00:00:00Z'::TIMESTAMP as time, 'budget_2025' as target, 1800.0 as value`,
			"metrics_view.yaml": `
version: 1
type: metrics_view
model: model
timeseries: time
measures:
  - name: val_sum
    expression: sum(val)
targets:
  - model: targets
    measures: [val_sum]`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mvRes := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics_view")
	require.NotNil(t, mvRes)
	mv := mvRes.GetMetricsView()
	require.NotNil(t, mv)
	validSpec := mv.State.ValidSpec
	require.NotNil(t, validSpec)
	require.Len(t, validSpec.Targets, 1)

	security, err := rt.ResolveSecurity(context.Background(), instanceID, nil, mvRes)
	require.NoError(t, err)

	e, err := executor.New(context.Background(), rt, instanceID, validSpec, false, security, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	// Test querying targets
	qry := &metricsview.TargetsQuery{
		MetricsView: "metrics_view",
		Measures:    []string{"val_sum"},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		TimeZone: "UTC",
		Priority: 0,
	}

	rows, err := e.Targets(context.Background(), qry)
	require.NoError(t, err)
	require.Greater(t, len(rows), 0)

	// Verify we got target rows
	hasTarget := false
	for _, row := range rows {
		if target, ok := row["target"].(string); ok {
			require.Contains(t, []string{"budget_2024", "budget_2025"}, target)
			hasTarget = true
		}
		if value, ok := row["value"].(float64); ok {
			require.Greater(t, value, 0.0)
		}
	}
	require.True(t, hasTarget, "expected to find target in target rows")
}

func TestTargetsQueryWithGrain(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'DK' as country, 1 as val union all
select '2024-01-02T00:00:00Z'::TIMESTAMP as time, 'US' as country, 2 as val`,
			"targets.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'budget_2024' as series, 'day' as grain, 1000.0 as value
union all
select '2024-02-01T00:00:00Z'::TIMESTAMP as time, 'budget_2024' as series, 'month' as grain, 1200.0 as value`,
			"metrics_view.yaml": `
version: 1
type: metrics_view
model: model
timeseries: time
measures:
  - name: val_sum
    expression: sum(val)
targets:
  - model: targets
    measures: [val_sum]`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mvRes := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics_view")
	require.NotNil(t, mvRes)
	mv := mvRes.GetMetricsView()
	require.NotNil(t, mv)
	validSpec := mv.State.ValidSpec
	require.NotNil(t, validSpec)
	require.Len(t, validSpec.Targets, 1)
	require.True(t, validSpec.Targets[0].HasGrain)

	security, err := rt.ResolveSecurity(context.Background(), instanceID, nil, mvRes)
	require.NoError(t, err)

	e, err := executor.New(context.Background(), rt, instanceID, validSpec, false, security, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	// Test querying targets with grain filter
	qry := &metricsview.TargetsQuery{
		MetricsView: "metrics_view",
		Measures:    []string{"val_sum"},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		TimeZone:  "UTC",
		TimeGrain: metricsview.TimeGrainDay,
		Priority:  0,
	}

	rows, err := e.Targets(context.Background(), qry)
	require.NoError(t, err)
	// Should filter to only day grain targets
	for _, row := range rows {
		if grain, ok := row["grain"].(string); ok {
			require.Equal(t, "day", grain)
		}
	}
}

func TestTargetsQueryWithSeriesFilter(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'DK' as country, 1 as val`,
			"targets.sql": `
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'budget_2024' as target, 1000.0 as value
union all
select '2024-01-01T00:00:00Z'::TIMESTAMP as time, 'budget_2025' as target, 1500.0 as value`,
			"metrics_view.yaml": `
version: 1
type: metrics_view
model: model
timeseries: time
measures:
  - name: val_sum
    expression: sum(val)
targets:
  - model: targets
    measures: [val_sum]`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	mvRes := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "metrics_view")
	require.NotNil(t, mvRes)
	mv := mvRes.GetMetricsView()
	require.NotNil(t, mv)
	validSpec := mv.State.ValidSpec
	require.NotNil(t, validSpec)

	security, err := rt.ResolveSecurity(context.Background(), instanceID, nil, mvRes)
	require.NoError(t, err)

	e, err := executor.New(context.Background(), rt, instanceID, validSpec, false, security, 0, nil)
	require.NoError(t, err)
	defer e.Close()

	// Test querying targets with target filter
	qry := &metricsview.TargetsQuery{
		MetricsView: "metrics_view",
		Measures:    []string{"val_sum"},
		Target:      "budget_2024",
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		TimeZone: "UTC",
		Priority: 0,
	}

	rows, err := e.Targets(context.Background(), qry)
	require.NoError(t, err)
	require.Greater(t, len(rows), 0)

	// All rows should have the filtered target
	for _, row := range rows {
		target, ok := row["target"].(string)
		require.True(t, ok, "row should have target field")
		require.Equal(t, "budget_2024", target)
	}
}

