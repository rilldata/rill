package reconcilers_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestCanvasValidState(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + canvas
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
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
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	// Change the model so it breaks the metrics view and canvas, check valid spec is preserved
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'bar' as bar, 2 as y`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 4, 0)
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Len(t, c1.Meta.Refs, 2)
	for _, componentName := range c1.Meta.Refs {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, componentName.Name)
		require.NotNil(t, r.GetComponent().State.ValidSpec)
	}

	// Fix everything
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'foo' as foo, 1 as x`})
	testruntime.ReconcileParserAndWait(t, rt, id)

	// Break one canvas component. Check valid spec is preserved.
	testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: doesnt_exist
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 2, 0)
	mv1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.NotEmpty(t, c1.Meta.ReconcileError)
	require.Len(t, c1.Meta.Refs, 2)
	var valid, invalid int
	for _, componentName := range c1.Meta.Refs {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, componentName.Name)
		if r.GetComponent().State.ValidSpec == nil {
			invalid++
		} else {
			valid++
		}
	}
	require.Equal(t, 1, valid)
	require.Equal(t, 1, invalid)
}

func TestCanvasValidateMetricsViewTimeConsistency(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create two models, two metrics views with consistent time settings, and a canvas referencing both
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
first_day_of_week: 2
first_month_of_year: 3
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
first_day_of_week: 2
first_month_of_year: 3
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
      - kpi_grid:
          metrics_view: mv2
          measures:
            - y
`,
	})

	// Reconcile and verify success
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

	// Verify that the canvas and metrics views have valid specs
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv1.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv1.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Change one metrics view to have inconsistent time settings
	testruntime.PutFiles(t, rt, id, map[string]string{
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
first_day_of_week: 1
first_month_of_year: 1
`,
	})

	// Reconcile and verify that the metrics view gets updated
	testruntime.ReconcileParserAndWait(t, rt, id)

	// Verify that mv2's valid spec got updated
	mv2 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(1), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(1), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Verify that the canvas reconciliation fails with the expected error
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotEmpty(t, c1.Meta.ReconcileError)
	require.Contains(t, c1.Meta.ReconcileError, "inconsistent first_day_of_week")

	// With StageChanges==true, the valid spec should be preserved
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	// Fix the metrics view to have consistent time settings again
	testruntime.PutFiles(t, rt, id, map[string]string{
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
first_day_of_week: 2
first_month_of_year: 3
`,
	})

	// Reconcile and verify success again
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

	// Verify that mv2's valid spec got updated again
	mv2 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Verify that the canvas reconciles successfully again
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Empty(t, c1.Meta.ReconcileError)
}

func TestCanvasDataRefreshedOn(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + canvas
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `
-- @materialize: true
SELECT 'foo' as foo, 1 as x
`,
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
		"c1.yaml": `
type: canvas
rows:
  - items:
    - kpi_grid:
        metrics_view: mv1
        measures:
          - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	getAndCheckRefreshedOn := func() time.Time {
		c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
		require.NotNil(t, c1.GetCanvas().State.DataRefreshedOn)

		comp1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, c1.Meta.Refs[0].Name)
		require.NotNil(t, comp1.GetComponent().State.DataRefreshedOn)

		mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
		require.NotNil(t, mv1.GetMetricsView().State.DataRefreshedOn)

		require.Equal(t, c1.GetCanvas().State.DataRefreshedOn, comp1.GetComponent().State.DataRefreshedOn)
		require.Equal(t, c1.GetCanvas().State.DataRefreshedOn, mv1.GetMetricsView().State.DataRefreshedOn)

		return c1.GetCanvas().State.DataRefreshedOn.AsTime()
	}

	refreshedOn1 := getAndCheckRefreshedOn()
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m1"})
	refreshedOn2 := getAndCheckRefreshedOn()
	require.Greater(t, refreshedOn2, refreshedOn1)
}

func TestCanvasResolveTransitiveAccess(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create three models, three metrics views, and a canvas with components using metrics_view, metrics_sql, and markdown.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"m3.sql": `SELECT 'baz' as baz, 3 as z`,
		"m4.sql": `SELECT 'qux' as qux, 4 as w`,
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
		"mv4.yaml": `
version: 1
type: metrics_view
model: m4
dimensions:
- column: qux
measures:
- name: w
  expression: sum(w)
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
  - items:
      - custom_chart:
          metrics_sql:
            - "SELECT bar, y FROM mv2"
            - "SELECT qux, w FROM mv4"
  - items:
      - markdown:
          content: 'Total z: {{ metrics_sql "SELECT z FROM mv3" }}'
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 14, 0, 0)

	// Build claims with a transitive access rule on the canvas
	ctx := t.Context()
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: runtime.ResourceKindCanvas,
							Name: "c1",
						},
					},
				},
			},
		},
	}

	// Resolve security for mv1 (referenced via metrics_view); should be accessible
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err := rt.ResolveSecurity(ctx, id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv2 (referenced via metrics_sql); should be accessible
	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv2)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv3 (referenced via metrics_sql in markdown content); should be accessible
	mv3 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv3")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv3)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv4 (referenced via a metrics_sql list in a multi-query custom chart); should be accessible
	mv4 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv4")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv4)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
}

func TestCanvasItemParamBindings(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create a model, a metrics view, a parameterized component, and a canvas binding values to its params.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'foo' as foo, 1 as x`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
timeseries: ts
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"trend.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
  - name: measure
    type: measure
    required: true
  - name: time_dim
    type: time_dimension
    required: true
  - name: limit
    type: number
    default: 500
custom_chart:
  metrics_sql: SELECT {{ .params.time_dim }} AS ts, {{ .params.measure }} AS value FROM {{ .params.metrics_view }} LIMIT {{ .params.limit }}
  spec:
    chartType: Line Chart
    encodings:
      x: {field: ts}
      y: {field: value}
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: x
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	// The canvas should be valid and have a parser-added ref to the param-bound metrics view.
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	var hasMVRef bool
	for _, ref := range c1.Meta.Refs {
		if ref.Kind == runtime.ResourceKindMetricsView && ref.Name == "mv1" {
			hasMVRef = true
		}
	}
	require.True(t, hasMVRef, "expected canvas to have a ref to the param-bound metrics view")

	// The param-bound metrics view should be accessible through transitive access on the canvas.
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "c1"},
					},
				},
			},
		},
	}
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err := rt.ResolveSecurity(t.Context(), id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Invalid: a bound measure that doesn't exist in the metrics view.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: nonexistent
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", `not a measure in metrics view "mv1"`)

	// Invalid: a missing required param.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", `missing value for required param "measure"`)

	// Invalid: passing params to a component that doesn't declare any.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"plain.yaml": `
type: component
kpi:
  metrics_view: mv1
  measure: x
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: plain
        params:
          foo: 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", "does not declare any params")

	// Valid again: fix the canvas and check it recovers.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: x
          time_dim: ts
          limit: 100
      - component: plain
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
}
