package reconcilers_test

import (
	"testing"

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
	e1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, e1.GetCanvas().State.ValidSpec)

	// Change the model so it breaks the metrics view and canvas, check valid spec is preserved
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'bar' as bar, 2 as y`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 4, 0)
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	e1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, e1.GetCanvas().State.ValidSpec)

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
	e1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, e1.GetCanvas().State.ValidSpec)
	require.NotEmpty(t, e1.Meta.ReconcileError)
}
