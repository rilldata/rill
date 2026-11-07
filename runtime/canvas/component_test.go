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

func TestValidateHorizontalBarChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: a quantitative x holds the measure and y holds the dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: y
    type: quantitative
    fields: [y, z]
  y:
    field: foo
    type: nominal
    sort: -x
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: a quantitative x must reference a measure.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
stacked_bar:
  metrics_view: mv1
  x:
    field: foo
    type: quantitative
  y:
    field: bar
    type: nominal
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", `x.field "foo" is not a measure`)

	// Invalid: with the measure on x, y must reference a dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: y
    type: quantitative
  y:
    field: z
    type: nominal
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", `y.field "z" is not a dimension`)

	// Invalid: line charts cannot be drawn horizontally.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
line_chart:
  metrics_view: mv1
  x:
    field: y
    type: quantitative
  y:
    field: foo
    type: nominal
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "only supported by bar charts")
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

	// Valid: encoded time-grain column (the flat table frontend encodes a time dimension at a
	// chosen grain as "{timeDimension}_rill_{GRAIN}").
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  columns:
  - ts_rill_TIME_GRAIN_MONTH
  - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: object entries with per-column overrides and the presentation properties.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  fit_to_width: true
  wrap: true
  wrap_headers: true
  wrap_lines: 3
  sort_by: y
  sort_dir: asc
  sort_comparison: delta
  columns:
  - name: foo
    width: 240
    wrap: true
    label: Foo label
    align: center
  - name: ts_rill_TIME_GRAIN_MONTH
    width: 120
  - name: y
    width: 90
    format_d3: ".3s"
    align: left
    delta:
      width: 80
    percent_change:
      width: 70
  - bar
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid overrides and presentation properties.
	invalid := []struct {
		name        string
		yaml        string
		errContains string
	}{
		{"fit_to_width not a boolean", `
type: component
table:
  metrics_view: mv1
  fit_to_width: "yes"
  columns: [foo, y]
`, "fit_to_width"},
		{"width as a string", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: foo
    width: "140"
`, "'width' must be an integer"},
		{"fractional width", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: foo
    width: 140.5
`, "'width' must be an integer"},
		{"measure width below the minimum", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: y
    width: 59
`, "between 60 and 300"},
		{"dimension width above the maximum", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: foo
    width: 601
`, "between 100 and 600"},
		{"wrap on a measure", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: y
    wrap: true
`, "'wrap' only applies to dimension columns"},
		{"number format on a dimension", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: foo
    format_d3: ".2f"
`, "number formats only apply to measures"},
		{"both number formats", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: y
    format_preset: humanize
    format_d3: ".2f"
`, "cannot set both"},
		{"unknown format preset", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: y
    format_preset: money
`, "'format_preset' must be one of"},
		{"object without a name", `
type: component
table:
  metrics_view: mv1
  columns:
  - width: 100
`, "non-empty 'name'"},
		{"duplicate field", `
type: component
table:
  metrics_view: mv1
  columns:
  - foo
  - name: foo
    width: 120
`, "more than once"},
		{"wrap_lines out of range", `
type: component
table:
  metrics_view: mv1
  wrap_lines: 9
  columns: [foo, y]
`, "wrap_lines"},
		{"sort_by not in the component", `
type: component
table:
  metrics_view: mv1
  sort_by: bar
  columns: [foo, y]
`, "sort_by"},
		{"unknown sort_dir", `
type: component
table:
  metrics_view: mv1
  sort_by: y
  sort_dir: down
  columns: [foo, y]
`, "sort_dir"},
		{"sort_dir without sort_by", `
type: component
table:
  metrics_view: mv1
  sort_dir: asc
  columns: [foo, y]
`, "requires"},
		{"unknown sort_comparison", `
type: component
table:
  metrics_view: mv1
  sort_by: y
  sort_comparison: previous
  columns: [foo, y]
`, "sort_comparison"},
		{"sort_comparison on a dimension", `
type: component
table:
  metrics_view: mv1
  sort_by: foo
  sort_comparison: delta
  columns: [foo, y]
`, "name a measure"},
		{"sort_comparison without sort_by", `
type: component
table:
  metrics_view: mv1
  sort_comparison: percent_change
  columns: [foo, y]
`, "requires"},
		{"delta on a dimension", `
type: component
table:
  metrics_view: mv1
  columns:
  - name: foo
    delta:
      width: 80
  - y
`, "only applies to measures"},
		{"delta width out of bounds", `
type: component
table:
  metrics_view: mv1
  columns:
  - foo
  - name: y
    delta:
      width: 20
`, "'delta.width' must be between"},
		{"percent_change not an object", `
type: component
table:
  metrics_view: mv1
  columns:
  - foo
  - name: y
    percent_change: 80
`, "must be an object"},
		{"delta on an adhoc measure", `
type: component
table:
  metrics_view: mv1
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
  columns:
  - foo
  - name: profit
    delta:
      width: 80
`, "adhoc measures have no comparison columns"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": tc.yaml})
			testruntime.ReconcileParserAndWait(t, rt, id)
			testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
			testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", tc.errContains)
		})
	}
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

	// Valid: encoded time-grain dimensions (the canvas pivot frontend encodes a time
	// dimension at a chosen grain as "{timeDimension}_rill_{GRAIN}").
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  measures:
  - y
  row_dimensions:
  - ts_rill_TIME_GRAIN_DAY
  col_dimensions:
  - ts_rill_TIME_GRAIN_MONTH
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: encoded time dimension with an unknown grain.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  col_dimensions:
  - ts_rill_TIME_GRAIN_BOGUS
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")

	// Valid: object entries with per-column overrides; the column dimension only carries a label.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  fit_to_width: true
  sort_by: y
  sort_comparison: percent_change
  measures:
  - name: y
    width: 120
    format_preset: percentage
    delta:
      width: 80
    align: center
  - z
  row_dimensions:
  - name: foo
    width: 260
    wrap: true
    label: Foo
  col_dimensions:
  - name: bar
    label: Bar group
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: sorting by a row dimension.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  sort_by: foo
  sort_dir: desc
  measures: [y]
  row_dimensions: [foo]
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid overrides on pivoted dimensions and sorts by them.
	invalid := []struct {
		name        string
		yaml        string
		errContains string
	}{
		{"width on a column dimension", `
type: component
pivot:
  metrics_view: mv1
  measures: [y]
  row_dimensions: [foo]
  col_dimensions:
  - name: bar
    width: 120
`, "take no 'width'"},
		{"align on a column dimension", `
type: component
pivot:
  metrics_view: mv1
  measures: [y]
  col_dimensions:
  - name: bar
    align: center
`, "take no 'align'"},
		{"sort by a column dimension", `
type: component
pivot:
  metrics_view: mv1
  sort_by: bar
  measures: [y]
  row_dimensions: [foo]
  col_dimensions: [bar]
`, "sort_by"},
		{"measure width above the maximum", `
type: component
pivot:
  metrics_view: mv1
  measures:
  - name: y
    width: 301
`, "between 60 and 300"},
		{"align on the first row dimension", `
type: component
pivot:
  metrics_view: mv1
  measures: [y]
  row_dimensions:
  - name: foo
    align: center
`, "'align' is not supported on row dimensions"},
		{"width on a later row dimension", `
type: component
pivot:
  metrics_view: mv1
  measures: [y]
  row_dimensions:
  - foo
  - name: bar
    width: 200
`, "only the first row dimension renders a column"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": tc.yaml})
			testruntime.ReconcileParserAndWait(t, rt, id)
			testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
			testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", tc.errContains)
		})
	}
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

func TestValidateCustomChart(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: static metrics_sql string and a Flint spec.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  spec:
    chartType: Bar Chart
    encodings:
      x: {field: foo}
      y: {field: y}
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: channels may bind a bare field name instead of an encoding mapping.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  spec:
    chartType: Bar Chart
    encodings:
      x: foo
      y: [y]
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Valid: incomplete drafts are allowed; the visual editor persists custom charts with
	// missing or empty properties while the user is still building them.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: metrics_sql of the wrong type.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: 42
  spec:
    chartType: Bar Chart
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "must be a string or an array of strings")

	// Valid: an ejected component draws a Vega-Lite spec instead of a chart spec.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  vega_spec: '{"mark": "bar"}'
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: a component draws one or the other, not both.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  vega_spec: '{"mark": "bar"}'
  spec:
    chartType: Bar Chart
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 1, 1)
	testruntime.RequireParseErrors(t, rt, id, map[string]string{
		"/c1.yaml": `"spec" and "vega_spec" are mutually exclusive`,
	})

	// Invalid: vega_spec is a JSON string, not a mapping.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  vega_spec:
    mark: bar
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 1, 1)
	testruntime.RequireParseErrors(t, rt, id, map[string]string{
		"/c1.yaml": `"vega_spec" must be a string`,
	})

	// Invalid: spec must be a mapping, not a JSON string.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  spec: '{"chartType": "Bar Chart"}'
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 1, 1)
	testruntime.RequireParseErrors(t, rt, id, map[string]string{
		"/c1.yaml": `"spec" must be a mapping`,
	})

	// Invalid: empty chartType.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  spec:
    chartType: ""
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "'spec.chartType' must not be empty")

	// Invalid: an encoding channel that is neither a field name nor an encoding mapping.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
custom_chart:
  metrics_sql: SELECT foo, y FROM mv1
  spec:
    chartType: Bar Chart
    encodings:
      x: 42
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "must be a field name or an encoding mapping")

	// Valid: a parameterized component with templated properties reconciles.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
  - name: measure
    type: measure
    required: true
custom_chart:
  metrics_sql: SELECT foo, {{ .params.measure }} AS value FROM {{ .params.metrics_view }}
  spec:
    chartType: Bar Chart
    encodings:
      x: {field: foo}
      y: {field: "{{ .params.measure }}"}
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: a parameterized component still gets structural checks despite templated properties.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
custom_chart:
  metrics_sql: 42
  spec:
    chartType: Bar Chart
    encodings:
      x: {field: "{{ .params.metrics_view }}"}
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "must be a string or an array of strings")
}

func TestValidateParameterizedRenderer(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// Valid: all fields templated; membership checks are deferred to param binding.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
  - name: measure
    type: measure
    required: true
kpi:
  metrics_view: "{{ .params.metrics_view }}"
  measure: "{{ .params.measure }}"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: templated properties do not exempt the component from structural checks;
	// the required 'measure' property is missing.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
kpi:
  metrics_view: "{{ .params.metrics_view }}"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "'measure' property")

	// Valid: static metrics view with a templated field; the static x.field is still validated.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: measure
    type: measure
    required: true
    metrics_view: metrics_view
  - name: metrics_view
    type: metrics_view
    default: mv1
line_chart:
  metrics_view: mv1
  x:
    field: foo
  y:
    field: "{{ .params.measure }}"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Invalid: static x.field is a measure, caught even though other fields are templated.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: measure
    type: measure
    required: true
    metrics_view: metrics_view
  - name: metrics_view
    type: metrics_view
    default: mv1
line_chart:
  metrics_view: mv1
  x:
    field: y
  y:
    field: "{{ .params.measure }}"
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a dimension")
}

func TestValidateUnknownRendererWithParams(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// A parameterized component skips field validation, but an unknown renderer
	// (e.g. an erroneous `renderer:` wrapper around the real block) must still fail.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
renderer:
  custom_chart:
    metrics_sql: SELECT foo, y FROM {{ .params.metrics_view }}
    vega_spec: '{"mark": "bar"}'
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", `unsupported renderer "renderer"`)
}

func TestValidateEphemeralMeasures(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: metricsViewFiles(),
	})

	// A kpi_grid referencing a ephemeral measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures: [y, profit]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A leaderboard referencing a ephemeral measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
leaderboard:
  metrics_view: mv1
  measures: [profit]
  dimensions: [foo]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A table with a ephemeral measure column should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
table:
  metrics_view: mv1
  columns: [foo, y, profit]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A pivot with a ephemeral measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pivot:
  metrics_view: mv1
  measures: [profit]
  row_dimensions: [foo]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A bar chart using a ephemeral measure on the y axis (single and multi) should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
    type: nominal
  y:
    field: profit
    type: quantitative
    fields: [y, profit]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A pie chart using a ephemeral measure should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
pie_chart:
  metrics_view: mv1
  measure:
    field: profit
    type: quantitative
  color:
    field: foo
    type: nominal
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A heatmap using a ephemeral measure as the color field should be valid.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
heatmap:
  metrics_view: mv1
  x:
    field: foo
    type: nominal
  color:
    field: profit
    type: quantitative
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// A chart referencing an undefined ephemeral measure should fail.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
bar_chart:
  metrics_view: mv1
  x:
    field: foo
    type: nominal
  y:
    field: missing
    type: quantitative
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "not a measure in metrics view")

	// An invalid expression should fail validation.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures: [profit]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: sum(y)
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "aggregate function")

	// An expression referencing an unknown measure should fail validation.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures: [profit]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - unknown
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "not a measure in metrics view")

	// A measure not defined anywhere should still fail.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: component
kpi_grid:
  metrics_view: mv1
  measures: [missing]
  adhoc_measures:
  - name: profit
    display_name: Profit
    expression: y - z
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindComponent, "c1", "is not a measure")
}
