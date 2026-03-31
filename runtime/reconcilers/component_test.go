package reconcilers_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
)

func TestComponentValidationLineChart(t *testing.T) {
	// Setup model and metrics
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"m1.sql": `SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'foo' as foo, 1 as y`,
			"mv1.yaml": `
version: 1
type: metrics_view
model: m1
timeseries: ts
dimensions:
- column: foo
measures:
- name: y
  expression: sum(y)
`,
		},
	})

	// X time and Y measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: ts
  y:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// X categorical and Y measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// X measure should be invalid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: y
  y:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}
