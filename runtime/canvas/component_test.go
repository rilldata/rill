package canvas_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
)

// metricsViewFiles returns the standard model and metrics view fixture for chart tests.
func metricsViewFiles() map[string]string {
	return map[string]string{
		"m1.sql": `SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'foo' as foo, 'bar' as bar, 1 as y, 2 as z`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
timeseries: ts
dimensions:
- column: foo
- column: bar
measures:
- name: y
  expression: sum(y)
- name: z
  expression: sum(z)
`,
	}
}

func TestValidateLineChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
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

	// Y dimension should be invalid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}

func TestValidateBarChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid bar chart.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: x is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
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

func TestValidateCartesianMultiField(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid multi-field y.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
    fields:
    - y
    - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: one of the y.fields is a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
    fields:
    - y
    - foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}

func TestValidateCartesianColorField(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: color as a field config with dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
  color:
    field: bar
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: color as a plain string (should be skipped).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
  color: "primary"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: color.field is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
  color:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateCartesianRillMeasures(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: color with rill_measures (virtual field for multi-measure mode).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
stacked_bar:
  metrics_view: mv1
  color:
    field: rill_measures
    type: value
    legendOrientation: top
  x:
    field: ts
    type: temporal
  y:
    field: y
    fields:
    - y
    - z
    type: quantitative
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
}

func TestValidateComboChartColorMeasures(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: combo chart with color field "measures" type "value" (virtual field for dual-axis mode).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
combo_chart:
  metrics_view: mv1
  color:
    field: measures
    type: value
    legendOrientation: top
  x:
    field: foo
  y1:
    field: y
    mark: bar
  y2:
    field: z
    mark: line
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
}

func TestValidateDonutChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid donut chart.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
donut_chart:
  metrics_view: mv1
  measure:
    field: y
  color:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: measure.field is a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
donut_chart:
  metrics_view: mv1
  measure:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: missing measure.field.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
donut_chart:
  metrics_view: mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "measure.field")

	// Invalid: color.field is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pie_chart:
  metrics_view: mv1
  measure:
    field: y
  color:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateScatterPlot(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid scatter plot.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
scatter_plot:
  metrics_view: mv1
  x:
    field: y
  y:
    field: z
  dimension:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: x.field is a dimension (scatter expects measure).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
scatter_plot:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: dimension.field is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
scatter_plot:
  metrics_view: mv1
  x:
    field: y
  y:
    field: z
  dimension:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateFunnelChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid funnel chart with measure and stage.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
funnel_chart:
  metrics_view: mv1
  measure:
    field: y
  stage:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: funnel with only metrics_view (all fields optional).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
funnel_chart:
  metrics_view: mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: funnel with multi-field measures.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
funnel_chart:
  metrics_view: mv1
  measure:
    field: y
    fields:
    - y
    - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: stage.field is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
funnel_chart:
  metrics_view: mv1
  stage:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")

	// Invalid: measure.fields contains a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
funnel_chart:
  metrics_view: mv1
  measure:
    field: y
    fields:
    - y
    - foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}

func TestValidateHeatmap(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid heatmap.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
heatmap:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: bar
  color:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: x.field is a measure (heatmap expects dimensions).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
heatmap:
  metrics_view: mv1
  x:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")

	// Invalid: color.field is a dimension (heatmap expects measure for color).
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
heatmap:
  metrics_view: mv1
  color:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}

func TestValidateComboChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid combo chart.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
combo_chart:
  metrics_view: mv1
  x:
    field: foo
  y1:
    field: y
  y2:
    field: z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: y1.field is a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
combo_chart:
  metrics_view: mv1
  x:
    field: foo
  y1:
    field: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: x.field is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
combo_chart:
  metrics_view: mv1
  x:
    field: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateMarkdown(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{},
	})

	// Valid markdown.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
markdown:
  content: "# Hello World"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	// Invalid: empty content.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
markdown:
  content: ""
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "content")

	// Invalid: missing content.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
markdown:
  apply_formatting: true
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "content")
}

func TestValidateImage(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{},
	})

	// Valid image.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
image:
  url: "https://example.com/image.png"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	// Invalid: empty url.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
image:
  url: ""
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "url")

	// Invalid: missing url.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
image:
  alignment:
    horizontal: center
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "url")
}

func TestValidateKPI(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid KPI.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi:
  metrics_view: mv1
  measure: y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: measure is a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi:
  metrics_view: mv1
  measure: foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: missing measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi:
  metrics_view: mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "measure")
}

func TestValidateKPIGrid(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid KPI grid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures:
  - y
  - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: empty measures array.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures: []
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "measures")

	// Invalid: one measure doesn't exist.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures:
  - y
  - nonexistent
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}

func TestValidateTable(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid table with dimensions and measures as columns.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  columns:
  - foo
  - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: empty columns.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  columns: []
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "columns")

	// Invalid: column doesn't exist as dimension or measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  columns:
  - foo
  - nonexistent
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension or measure")
}

func TestValidatePivot(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid pivot with all fields.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  measures:
  - y
  row_dimensions:
  - foo
  col_dimensions:
  - bar
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: only measures.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  measures:
  - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: only row_dimensions.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  row_dimensions:
  - foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: all arrays empty.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "at least one")

	// Invalid: measure doesn't exist.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  measures:
  - nonexistent
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: row_dimensions value is a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  row_dimensions:
  - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateLeaderboard(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid leaderboard.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
  measures:
  - y
  dimensions:
  - foo
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: only measures.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
  measures:
  - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: no measures or dimensions.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "at least one")

	// Invalid: measure doesn't exist.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
  measures:
  - nonexistent
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")

	// Invalid: dimension doesn't exist.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
  dimensions:
  - nonexistent
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}
