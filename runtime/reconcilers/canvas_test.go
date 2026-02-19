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
	require.Len(t, c1.Meta.Refs, 3) // 2 components + 1 MetricsView (mv1)
	var foundMvRef bool
	for _, ref := range c1.Meta.Refs {
		if ref.Kind == runtime.ResourceKindMetricsView {
			require.Equal(t, "mv1", ref.Name)
			foundMvRef = true
			continue
		}
		require.Equal(t, runtime.ResourceKindComponent, ref.Kind)
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, ref.Name)
		require.NotNil(t, r.GetComponent().State.ValidSpec)
	}
	require.True(t, foundMvRef, "expected a MetricsView ref for mv1")

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
	require.Len(t, c1.Meta.Refs, 4) // 2 components + 2 MetricsView (doesnt_exist + mv1)
	var valid, invalid int
	mvRefNames := make(map[string]bool)
	for _, ref := range c1.Meta.Refs {
		if ref.Kind == runtime.ResourceKindMetricsView {
			mvRefNames[ref.Name] = true
			continue
		}
		require.Equal(t, runtime.ResourceKindComponent, ref.Kind)
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, ref.Name)
		if r.GetComponent().State.ValidSpec == nil {
			invalid++
		} else {
			valid++
		}
	}
	require.Equal(t, 1, valid)
	require.Equal(t, 1, invalid)
	require.True(t, mvRefNames["mv1"], "expected MetricsView ref for mv1")
	require.True(t, mvRefNames["doesnt_exist"], "expected MetricsView ref for doesnt_exist")
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

		// Find the component ref (canvas refs now include MetricsView refs too)
		var compRefName string
		for _, ref := range c1.Meta.Refs {
			if ref.Kind == runtime.ResourceKindComponent {
				compRefName = ref.Name
				break
			}
		}
		require.NotEmpty(t, compRefName)

		comp1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, compRefName)
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
