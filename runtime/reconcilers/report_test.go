package reconcilers_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestReportAIExploreWatermarkInherit(t *testing.T) {
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
		"/explores/e1.yaml": `
type: explore
metrics_view: mv1
`,
		// AI report that inherits its watermark through the explore
		"/reports/r1.yaml": `
type: report
display_name: AI Report
refresh:
  cron: 0 8 * * *
watermark: inherit
data:
  ai:
    prompt: Analyze key metrics
    explore: e1
notify:
  email:
    recipients:
      - somebody@example.com
`,
		// Same report without watermark: inherit, which should use the trigger time
		"/reports/r2.yaml": `
type: report
display_name: AI Report
refresh:
  cron: 0 8 * * *
data:
  ai:
    prompt: Analyze key metrics
    explore: e1
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)

	r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.Meta.Refs, 1)
	require.Equal(t, runtime.ResourceKindExplore, r1.Meta.Refs[0].Kind)
	require.Equal(t, "e1", r1.Meta.Refs[0].Name)

	// Trigger both reports.
	// The execution's report time is recorded before the report is sent, so it is available regardless of whether the AI session could be created.
	triggerTime := time.Now()
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"})

	// r1 should have resolved the watermark of the explore's metrics view.
	// The watermark is max(__time) plus one second, since it's used as an exclusive upper bound (see ResourceWatermark).
	r1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.GetReport().State.ExecutionHistory, 1)
	require.Equal(t, time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC), r1.GetReport().State.ExecutionHistory[0].ReportTime.AsTime())

	// r2 should have used the trigger time
	r2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r2")
	require.Len(t, r2.GetReport().State.ExecutionHistory, 1)
	require.WithinDuration(t, triggerTime, r2.GetReport().State.ExecutionHistory[0].ReportTime.AsTime(), time.Minute)
}

func TestReportCanvasResolveTransitiveAccess(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create a canvas with components referencing mv1 and mv2, an unrelated metrics view mv3,
	// and a canvas PDF report referencing the canvas.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"m3.sql": `SELECT 'baz' as baz, 3 as z`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
`,
		"mv3.yaml": `
version: 1
type: metrics_view
model: m3
dimensions:
- column: baz
measures:
- name: z
  expression: sum(z)
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - custom_chart:
          metrics_sql: "SELECT bar, y FROM mv2"
`,
		"r1.yaml": `
type: report
display_name: My Canvas PDF Report
export:
  format: pdf
notify:
  email:
    recipients:
      - user_1@example.com
annotations:
  canvas: c1
  metrics_view_filters: '{"mv1":{"cond":{"op":"OPERATION_EQ","exprs":[{"ident":"foo"},{"val":"foo"}]}}}'
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 11, 0, 0)

	// Build claims with a transitive access rule on the report
	ctx := t.Context()
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: runtime.ResourceKindReport,
							Name: "r1",
						},
					},
				},
			},
		},
	}

	// The canvas referenced by the report should be accessible
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	sec, err := rt.ResolveSecurity(ctx, id, claims, c1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// The canvas's components should be accessible
	for _, ref := range c1.Meta.Refs {
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		comp := testruntime.GetResource(t, rt, id, ref.Kind, ref.Name)
		sec, err = rt.ResolveSecurity(ctx, id, claims, comp)
		require.NoError(t, err)
		require.True(t, sec.CanAccess())
	}

	// The metrics views referenced by the canvas's components should be accessible.
	// The filters captured in the report's "metrics_view_filters" annotation should be applied as a query filter.
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
	require.NotNil(t, sec.QueryFilter())

	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv2)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
	require.Nil(t, sec.QueryFilter())

	// A metrics view not referenced by the canvas should not be accessible
	mv3 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv3")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv3)
	require.NoError(t, err)
	require.False(t, sec.CanAccess())
}
