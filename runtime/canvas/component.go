package canvas

import (
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
)

// ValidateRendererProperties validates the renderer properties for a component.
// The provided metricsViews should contain every valid metrics view referenced by the component (as determined in the parser).
// If the renderer properties reference a metrics view not in metricsViews, assume the metrics view is invalid or does not exist (don't look it up separately in the catalog).
//
// Note: that metrics views referenced through markdown or metrics_sql cannot be validated here.
// This is because the upstream parser can't extract refs from them, so the metrics views cannot be passed through to here.
// Warning: if you try to fix this, note that the refs must be added in the parser, not looked up dynamically here;
// a dynamic lookup will have a race condition where the metrics view may not have been reconciled yet.
func ValidateRendererProperties(renderer string, props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	switch renderer {
	case "line_chart", "bar_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		return validateCartesianChart(props, metricsViews)
	case "donut_chart", "pie_chart":
		return validateCircularChart(props, metricsViews)
	case "scatter_plot":
		return validateScatterPlot(props, metricsViews)
	case "funnel_chart":
		return validateFunnelChart(props, metricsViews)
	case "heatmap":
		return validateHeatmap(props, metricsViews)
	case "combo_chart":
		return validateComboChart(props, metricsViews)
	case "markdown":
		return validateMarkdown(props)
	case "image":
		return validateImage(props)
	case "kpi":
		return validateKPI(props, metricsViews)
	case "kpi_grid":
		return validateKPIGrid(props, metricsViews)
	case "table":
		return validateTable(props, metricsViews)
	case "pivot":
		return validatePivot(props, metricsViews)
	case "leaderboard":
		return validateLeaderboard(props, metricsViews)
	default:
		return fmt.Errorf("unsupported renderer %q", renderer)
	}
}

// validateCartesianChart validates properties for line_chart, bar_chart, area_chart, stacked_bar, and stacked_bar_normalized.
func validateCartesianChart(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	xField, ok := pathutil.GetPathString(props, "x.field")
	if !ok {
		return errors.New("renderer properties must include a string 'x.field' property")
	}
	if !metricsViewHasDimension(mv, xField) {
		return fmt.Errorf("referenced x.field %q is not a dimension in metrics view %q", xField, mvn)
	}

	yField, ok := pathutil.GetPathString(props, "y.field")
	if !ok {
		return errors.New("renderer properties must include a string 'y.field' property")
	}
	if !metricsViewHasMeasure(mv, yField) {
		return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
	}

	// Validate optional multi-field measures (y.fields)
	if yFields, ok := getPathStringSlice(props, "y.fields"); ok {
		for _, f := range yFields {
			if !metricsViewHasMeasure(mv, f) {
				return fmt.Errorf("referenced y.fields value %q is not a measure in metrics view %q", f, mvn)
			}
		}
	}

	// Validate optional color field: can be a plain string (skip) or a map with a "field" key (validate as dimension)
	if err := validateOptionalColorDimensionField(mv, mvn, props); err != nil {
		return err
	}

	return nil
}

// validateCircularChart validates properties for donut_chart and pie_chart.
func validateCircularChart(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	measureField, ok := pathutil.GetPathString(props, "measure.field")
	if !ok {
		return errors.New("renderer properties must include a string 'measure.field' property")
	}
	if !metricsViewHasMeasure(mv, measureField) {
		return fmt.Errorf("referenced measure.field %q is not a measure in metrics view %q", measureField, mvn)
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "color.field"); err != nil {
		return err
	}

	return nil
}

// validateScatterPlot validates properties for scatter_plot.
func validateScatterPlot(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	xField, ok := pathutil.GetPathString(props, "x.field")
	if !ok {
		return errors.New("renderer properties must include a string 'x.field' property")
	}
	if !metricsViewHasMeasure(mv, xField) {
		return fmt.Errorf("referenced x.field %q is not a measure in metrics view %q", xField, mvn)
	}

	yField, ok := pathutil.GetPathString(props, "y.field")
	if !ok {
		return errors.New("renderer properties must include a string 'y.field' property")
	}
	if !metricsViewHasMeasure(mv, yField) {
		return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "dimension.field"); err != nil {
		return err
	}

	if err := validateOptionalMeasureField(mv, mvn, props, "size.field"); err != nil {
		return err
	}

	// Color can be a plain string or a map with a "field" key
	if err := validateOptionalColorDimensionField(mv, mvn, props); err != nil {
		return err
	}

	return nil
}

// validateFunnelChart validates properties for funnel_chart.
func validateFunnelChart(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	if err := validateOptionalMeasureField(mv, mvn, props, "measure.field"); err != nil {
		return err
	}

	// Validate optional multi-field measures (measure.fields)
	if fields, ok := getPathStringSlice(props, "measure.fields"); ok {
		for _, f := range fields {
			if !metricsViewHasMeasure(mv, f) {
				return fmt.Errorf("referenced measure.fields value %q is not a measure in metrics view %q", f, mvn)
			}
		}
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "stage.field"); err != nil {
		return err
	}

	return nil
}

// validateHeatmap validates properties for heatmap.
func validateHeatmap(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "x.field"); err != nil {
		return err
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "y.field"); err != nil {
		return err
	}

	// Note: for heatmap, color is a measure (not a dimension like other charts)
	if err := validateOptionalMeasureField(mv, mvn, props, "color.field"); err != nil {
		return err
	}

	return nil
}

// validateComboChart validates properties for combo_chart.
func validateComboChart(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	if err := validateOptionalDimensionField(mv, mvn, props, "x.field"); err != nil {
		return err
	}

	if err := validateOptionalMeasureField(mv, mvn, props, "y1.field"); err != nil {
		return err
	}

	if err := validateOptionalMeasureField(mv, mvn, props, "y2.field"); err != nil {
		return err
	}

	// Combo chart color is typically {field: "measures", type: "value"} for dual-axis mode
	if err := validateOptionalColorDimensionField(mv, mvn, props); err != nil {
		return err
	}

	return nil
}

// validateMarkdown validates properties for markdown.
func validateMarkdown(props map[string]any) error {
	content, ok := pathutil.GetPathString(props, "content")
	if !ok || strings.TrimSpace(content) == "" {
		return errors.New("renderer properties for markdown must include a non-empty string 'content' property")
	}
	return nil
}

// validateImage validates properties for image.
func validateImage(props map[string]any) error {
	url, ok := pathutil.GetPathString(props, "url")
	if !ok || strings.TrimSpace(url) == "" {
		return errors.New("renderer properties for image must include a non-empty string 'url' property")
	}
	return nil
}

// validateKPI validates properties for kpi.
func validateKPI(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	// KPI uses a top-level "measure" string, not "measure.field"
	measure, ok := pathutil.GetPathString(props, "measure")
	if !ok {
		return errors.New("renderer properties for kpi must include a string 'measure' property")
	}
	if !metricsViewHasMeasure(mv, measure) {
		return fmt.Errorf("referenced measure %q is not a measure in metrics view %q", measure, mvn)
	}

	return nil
}

// validateKPIGrid validates properties for kpi_grid.
// kpi_grid supports metrics_sql as an alternative to metrics_view; field validation is skipped in that case.
func validateKPIGrid(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		// kpi_grid supports metrics_sql as an alternative data source
		if _, hasSQL := pathutil.GetPath(props, "metrics_sql"); hasSQL {
			return nil
		}
		return err
	}

	measures, ok := getPathStringSlice(props, "measures")
	if !ok || len(measures) == 0 {
		return errors.New("renderer properties for kpi_grid must include a non-empty 'measures' array")
	}
	for _, m := range measures {
		if !metricsViewHasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}

	return nil
}

// validateTable validates properties for table.
func validateTable(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	columns, ok := getPathStringSlice(props, "columns")
	if !ok || len(columns) == 0 {
		return errors.New("renderer properties for table must include a non-empty 'columns' array")
	}
	for _, col := range columns {
		if !metricsViewHasDimension(mv, col) && !metricsViewHasMeasure(mv, col) {
			return fmt.Errorf("referenced columns value %q is not a dimension or measure in metrics view %q", col, mvn)
		}
	}

	return nil
}

// validatePivot validates properties for pivot.
func validatePivot(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	measures, _ := getPathStringSlice(props, "measures")
	rowDims, _ := getPathStringSlice(props, "row_dimensions")
	colDims, _ := getPathStringSlice(props, "col_dimensions")

	if len(measures) == 0 && len(rowDims) == 0 && len(colDims) == 0 {
		return errors.New("renderer properties for pivot must include at least one of 'measures', 'row_dimensions', or 'col_dimensions'")
	}

	for _, m := range measures {
		if !metricsViewHasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}
	for _, d := range rowDims {
		if !metricsViewHasDimension(mv, d) {
			return fmt.Errorf("referenced row_dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}
	for _, d := range colDims {
		if !metricsViewHasDimension(mv, d) {
			return fmt.Errorf("referenced col_dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}

	return nil
}

// validateLeaderboard validates properties for leaderboard.
func validateLeaderboard(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	mvn, mv, err := requireMetricsView(props, metricsViews)
	if err != nil {
		return err
	}

	measures, _ := getPathStringSlice(props, "measures")
	dimensions, _ := getPathStringSlice(props, "dimensions")

	if len(measures) == 0 && len(dimensions) == 0 {
		return errors.New("renderer properties for leaderboard must include at least one 'measures' or 'dimensions' entry")
	}

	for _, m := range measures {
		if !metricsViewHasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}
	for _, d := range dimensions {
		if !metricsViewHasDimension(mv, d) {
			return fmt.Errorf("referenced dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}

	return nil
}

// requireMetricsView extracts and validates the "metrics_view" property from renderer props.
// It returns the metrics view name, spec, and nil error on success.
func requireMetricsView(props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) (string, *runtimev1.MetricsViewSpec, error) {
	mvn, ok := pathutil.GetPathString(props, "metrics_view")
	if !ok {
		return "", nil, errors.New("renderer properties must include a string 'metrics_view' property")
	}
	mv := metricsViews[mvn]
	if mv == nil {
		return "", nil, fmt.Errorf("referenced metrics view %q is invalid", mvn)
	}
	return mvn, mv, nil
}

// validateOptionalDimensionField validates that a field at the given path, if present, is a dimension in the metrics view.
func validateOptionalDimensionField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any, path string) error {
	field, ok := pathutil.GetPathString(props, path)
	if !ok {
		return nil
	}
	if !metricsViewHasDimension(mv, field) {
		return fmt.Errorf("referenced %s %q is not a dimension in metrics view %q", path, field, mvName)
	}
	return nil
}

// validateOptionalMeasureField validates that a field at the given path, if present, is a measure in the metrics view.
func validateOptionalMeasureField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any, path string) error {
	field, ok := pathutil.GetPathString(props, path)
	if !ok {
		return nil
	}
	if !metricsViewHasMeasure(mv, field) {
		return fmt.Errorf("referenced %s %q is not a measure in metrics view %q", path, field, mvName)
	}
	return nil
}

// validateOptionalColorDimensionField handles the special case where "color" can be:
//   - a plain string (e.g. "primary", "stage"): skip validation
//   - a map with type "value" (e.g. {field: "rill_measures", type: "value"}): skip validation; this is a virtual field for multi-measure mode
//   - a map with a "field" key: validate color.field as a dimension
//
// This pattern is used by cartesian charts, scatter plots, and combo charts.
func validateOptionalColorDimensionField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any) error {
	raw, ok := pathutil.GetPath(props, "color")
	if !ok {
		return nil
	}
	// If color is a plain string, skip validation (it's a color literal, not a field reference)
	if _, isString := raw.(string); isString {
		return nil
	}
	// If color has type "value", it's a virtual field (e.g. rill_measures or measures for multi-measure mode); skip validation
	if colorType, ok := pathutil.GetPathString(props, "color.type"); ok && colorType == "value" {
		return nil
	}
	// Otherwise validate color.field as a dimension
	return validateOptionalDimensionField(mv, mvName, props, "color.field")
}

// getPathStringSlice extracts a []string from a nested path in the props map.
// Returns false if the path doesn't exist or the value is not a []any of strings.
func getPathStringSlice(props map[string]any, path string) ([]string, bool) {
	raw, ok := pathutil.GetPath(props, path)
	if !ok {
		return nil, false
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		result = append(result, s)
	}
	return result, true
}

// metricsViewHasDimension returns true if the metrics view has a dimension with the given name.
func metricsViewHasDimension(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, d := range mv.Dimensions {
		if d.Name == fieldName {
			return true
		}
	}
	return false
}

// metricsViewHasMeasure returns true if the metrics view has a measure with the given name.
func metricsViewHasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, m := range mv.Measures {
		if m.Name == fieldName {
			return true
		}
	}
	return false
}
