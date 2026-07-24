package canvas

import (
	"encoding/json"
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
// allowTemplated should be true when the component declares params. In that case, properties may contain
// unresolved template placeholders (e.g. {{ .params.measure }}), which makes field-membership validation
// impossible for the templated values; those membership checks are skipped and the bound values are
// validated by the canvas reconciler (ValidateParamBindings) instead. All other checks still run:
// required keys must be present, values must have the right types, and enum-valued and non-templated
// field properties are validated as usual.
//
// Note: metrics views referenced through markdown content cannot be validated here.
// This is because the upstream parser can't extract refs from templates, so the metrics views cannot be passed through to here.
// Warning: if you try to fix this, note that the refs must be added in the parser, not looked up dynamically here;
// a dynamic lookup will have a race condition where the metrics view may not have been reconciled yet.
func ValidateRendererProperties(renderer string, props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec, allowTemplated bool) error {
	v := rendererValidator{
		metricsViews: metricsViews,
		lenient:      allowTemplated && hasTemplatedString(props),
	}
	switch renderer {
	case "line_chart", "bar_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		return v.validateCartesianChart(props)
	case "donut_chart", "pie_chart":
		return v.validateCircularChart(props)
	case "scatter_plot":
		return v.validateScatterPlot(props)
	case "funnel_chart":
		return v.validateFunnelChart(props)
	case "heatmap":
		return v.validateHeatmap(props)
	case "combo_chart":
		return v.validateComboChart(props)
	case "markdown":
		return validateMarkdown(props)
	case "image":
		return validateImage(props)
	case "kpi":
		return v.validateKPI(props)
	case "kpi_grid":
		return v.validateKPIGrid(props)
	case "table":
		return v.validateTable(props)
	case "pivot":
		return v.validatePivot(props)
	case "leaderboard":
		return v.validateLeaderboard(props)
	case "custom_chart":
		return validateCustomChart(props)
	default:
		return fmt.Errorf("unsupported renderer %q", renderer)
	}
}

// rendererValidator holds the context for validating renderer properties.
// When lenient is true, the properties may contain unresolved template placeholders:
// field-membership checks are skipped for templated values (and for all fields when the
// metrics view name itself is templated), while structural checks still run.
type rendererValidator struct {
	metricsViews map[string]*runtimev1.MetricsViewSpec
	lenient      bool
}

// validateCustomChart validates properties for custom_chart.
// It only rejects malformed values, not incomplete ones: the visual editor persists draft custom
// charts with empty properties, so completeness is enforced at render time instead.
// metrics_sql queries are validated at query time, and vega_spec is validated as JSON only when
// it contains no template placeholders.
func validateCustomChart(props map[string]any) error {
	if raw, ok := pathutil.GetPath(props, "metrics_sql"); ok {
		switch v := raw.(type) {
		case string:
			// Nothing to check.
		case []any:
			for i, e := range v {
				if _, ok := e.(string); !ok {
					return fmt.Errorf("renderer property 'metrics_sql' entry at index %d must be a string", i)
				}
			}
		default:
			return errors.New("renderer property 'metrics_sql' must be a string or an array of strings")
		}
	}

	vegaSpec, ok, err := getOptionalPathString(props, "vega_spec")
	if err != nil {
		return err
	}
	if ok && strings.TrimSpace(vegaSpec) != "" && !strings.Contains(vegaSpec, "{{") {
		var m map[string]any
		if err := json.Unmarshal([]byte(vegaSpec), &m); err != nil {
			return fmt.Errorf("renderer property 'vega_spec' is not a valid JSON object: %w", err)
		}
	}

	return nil
}

// hasTemplatedString reports whether any string nested in the value contains template placeholders.
func hasTemplatedString(val any) bool {
	switch val := val.(type) {
	case string:
		return strings.Contains(val, "{{")
	case map[string]any:
		for _, v := range val {
			if hasTemplatedString(v) {
				return true
			}
		}
	case []any:
		for _, v := range val {
			if hasTemplatedString(v) {
				return true
			}
		}
	}
	return false
}

// validateCartesianChart validates properties for line_chart, bar_chart, area_chart, stacked_bar, and stacked_bar_normalized.
func (v rendererValidator) validateCartesianChart(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	xField, ok := pathutil.GetPathString(props, "x.field")
	if !ok {
		return errors.New("renderer properties must include a string 'x.field' property")
	}
	if !v.hasDimension(mv, xField) {
		return fmt.Errorf("referenced x.field %q is not a dimension in metrics view %q", xField, mvn)
	}

	yField, ok := pathutil.GetPathString(props, "y.field")
	if !ok {
		return errors.New("renderer properties must include a string 'y.field' property")
	}
	if !v.hasMeasure(mv, yField) {
		return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
	}

	// Validate optional multi-field measures (y.fields)
	yFields, err := getPathStringSlice(props, "y.fields")
	if err != nil {
		return err
	}
	for _, f := range yFields {
		if !v.hasMeasure(mv, f) {
			return fmt.Errorf("referenced y.fields value %q is not a measure in metrics view %q", f, mvn)
		}
	}

	// Validate optional color field: can be a plain string (skip) or a map with a "field" key (validate as dimension)
	return v.validateOptionalColorDimensionField(mv, mvn, props)
}

// validateCircularChart validates properties for donut_chart and pie_chart.
func (v rendererValidator) validateCircularChart(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	measureField, ok := pathutil.GetPathString(props, "measure.field")
	if !ok {
		return errors.New("renderer properties must include a string 'measure.field' property")
	}
	if !v.hasMeasure(mv, measureField) {
		return fmt.Errorf("referenced measure.field %q is not a measure in metrics view %q", measureField, mvn)
	}

	return v.validateOptionalDimensionField(mv, mvn, props, "color.field")
}

// validateScatterPlot validates properties for scatter_plot.
func (v rendererValidator) validateScatterPlot(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	xField, ok := pathutil.GetPathString(props, "x.field")
	if !ok {
		return errors.New("renderer properties must include a string 'x.field' property")
	}
	if !v.hasMeasure(mv, xField) {
		return fmt.Errorf("referenced x.field %q is not a measure in metrics view %q", xField, mvn)
	}

	yField, ok := pathutil.GetPathString(props, "y.field")
	if !ok {
		return errors.New("renderer properties must include a string 'y.field' property")
	}
	if !v.hasMeasure(mv, yField) {
		return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "dimension.field"); err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "size.field"); err != nil {
		return err
	}

	// Color can be a plain string or a map with a "field" key
	return v.validateOptionalColorDimensionField(mv, mvn, props)
}

// validateFunnelChart validates properties for funnel_chart.
func (v rendererValidator) validateFunnelChart(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "measure.field"); err != nil {
		return err
	}

	// Validate optional multi-field measures (measure.fields)
	fields, err := getPathStringSlice(props, "measure.fields")
	if err != nil {
		return err
	}
	for _, f := range fields {
		if !v.hasMeasure(mv, f) {
			return fmt.Errorf("referenced measure.fields value %q is not a measure in metrics view %q", f, mvn)
		}
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "stage.field"); err != nil {
		return err
	}

	// Optional enum-valued funnel fields. The renderer falls back to defaults for
	// unknown values, but we reject typos here so authors get a clear error.
	if err := v.validateOptionalStringEnum(props, "breakdownMode", []string{"dimension", "measures"}); err != nil {
		return err
	}
	if err := v.validateOptionalStringEnum(props, "mode", []string{"width", "order"}); err != nil {
		return err
	}
	if err := v.validateOptionalStringEnum(props, "color", []string{"stage", "measure", "name", "value"}); err != nil {
		return err
	}
	return v.validateOptionalStringEnum(props, "percentMode", []string{"top", "previous"})
}

// validateHeatmap validates properties for heatmap.
func (v rendererValidator) validateHeatmap(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "x.field"); err != nil {
		return err
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "y.field"); err != nil {
		return err
	}

	// Note: for heatmap, color is a measure (not a dimension like other charts)
	return v.validateOptionalMeasureField(mv, mvn, props, "color.field")
}

// validateComboChart validates properties for combo_chart.
func (v rendererValidator) validateComboChart(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "x.field"); err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "y1.field"); err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "y2.field"); err != nil {
		return err
	}

	// Combo chart color is typically {field: "measures", type: "value"} for dual-axis mode
	return v.validateOptionalColorDimensionField(mv, mvn, props)
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
func (v rendererValidator) validateKPI(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	// KPI uses a top-level "measure" string, not "measure.field"
	measure, ok := pathutil.GetPathString(props, "measure")
	if !ok {
		return errors.New("renderer properties for kpi must include a string 'measure' property")
	}
	if !v.hasMeasure(mv, measure) {
		return fmt.Errorf("referenced measure %q is not a measure in metrics view %q", measure, mvn)
	}

	return nil
}

// validateKPIGrid validates properties for kpi_grid.
func (v rendererValidator) validateKPIGrid(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	measures, err := getPathStringSlice(props, "measures")
	if err != nil {
		return err
	}
	if len(measures) == 0 {
		return errors.New("renderer properties for kpi_grid must include a non-empty 'measures' array of strings")
	}
	for _, m := range measures {
		if !v.hasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}

	return nil
}

// validateTable validates properties for table.
func (v rendererValidator) validateTable(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	columns, err := getPathStringSlice(props, "columns")
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return errors.New("renderer properties for table must include a non-empty 'columns' array of strings")
	}
	for _, col := range columns {
		if !v.hasDimension(mv, col) && !v.hasMeasure(mv, col) && !isEncodedTimeDimension(mv, col) {
			return fmt.Errorf("referenced columns value %q is not a dimension or measure in metrics view %q", col, mvn)
		}
	}

	return nil
}

// validatePivot validates properties for pivot.
func (v rendererValidator) validatePivot(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	measures, err := getPathStringSlice(props, "measures")
	if err != nil {
		return err
	}
	rowDims, err := getPathStringSlice(props, "row_dimensions")
	if err != nil {
		return err
	}
	colDims, err := getPathStringSlice(props, "col_dimensions")
	if err != nil {
		return err
	}

	if len(measures) == 0 && len(rowDims) == 0 && len(colDims) == 0 {
		return errors.New("renderer properties for pivot must include at least one of 'measures', 'row_dimensions', or 'col_dimensions'")
	}

	for _, m := range measures {
		if !v.hasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}
	for _, d := range rowDims {
		if !v.hasDimension(mv, d) && !isEncodedTimeDimension(mv, d) {
			return fmt.Errorf("referenced row_dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}
	for _, d := range colDims {
		if !v.hasDimension(mv, d) && !isEncodedTimeDimension(mv, d) {
			return fmt.Errorf("referenced col_dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}

	return nil
}

// validateLeaderboard validates properties for leaderboard.
func (v rendererValidator) validateLeaderboard(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	measures, err := getPathStringSlice(props, "measures")
	if err != nil {
		return err
	}
	dimensions, err := getPathStringSlice(props, "dimensions")
	if err != nil {
		return err
	}

	if len(measures) == 0 && len(dimensions) == 0 {
		return errors.New("renderer properties for leaderboard must include at least one 'measures' or 'dimensions' entry")
	}

	for _, m := range measures {
		if !v.hasMeasure(mv, m) {
			return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m, mvn)
		}
	}
	for _, d := range dimensions {
		if !v.hasDimension(mv, d) {
			return fmt.Errorf("referenced dimensions value %q is not a dimension in metrics view %q", d, mvn)
		}
	}

	return nil
}

// requireMetricsView extracts and validates the "metrics_view" property from renderer props.
// It returns the metrics view name, spec, and nil error on success.
// In lenient mode, a templated metrics view name returns a nil spec without error;
// the membership helpers skip their checks when the spec is nil.
func (v rendererValidator) requireMetricsView(props map[string]any) (string, *runtimev1.MetricsViewSpec, error) {
	mvn, ok := pathutil.GetPathString(props, "metrics_view")
	if !ok {
		return "", nil, errors.New("renderer properties must include a string 'metrics_view' property")
	}
	if v.lenient && strings.Contains(mvn, "{{") {
		return mvn, nil, nil
	}
	mv := v.metricsViews[mvn]
	if mv == nil {
		return "", nil, fmt.Errorf("referenced metrics view %q is invalid", mvn)
	}
	return mvn, mv, nil
}

// hasDimension reports whether the metrics view has a dimension with the given name.
// In lenient mode the check passes when it cannot run: the metrics view is unresolved
// (templated name, mv is nil) or the field name itself is templated.
func (v rendererValidator) hasDimension(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	if v.lenient && (mv == nil || strings.Contains(fieldName, "{{")) {
		return true
	}
	return metricsViewHasDimension(mv, fieldName)
}

// hasMeasure reports whether the metrics view has a measure with the given name.
// In lenient mode the check passes when it cannot run; see hasDimension.
func (v rendererValidator) hasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	if v.lenient && (mv == nil || strings.Contains(fieldName, "{{")) {
		return true
	}
	return metricsViewHasMeasure(mv, fieldName)
}

// validateOptionalDimensionField validates that a field at the given path, if present, is a dimension in the metrics view.
func (v rendererValidator) validateOptionalDimensionField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any, path string) error {
	field, ok, err := getOptionalPathString(props, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !v.hasDimension(mv, field) {
		return fmt.Errorf("referenced %s %q is not a dimension in metrics view %q", path, field, mvName)
	}
	return nil
}

// validateOptionalMeasureField validates that a field at the given path, if present, is a measure in the metrics view.
func (v rendererValidator) validateOptionalMeasureField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any, path string) error {
	field, ok, err := getOptionalPathString(props, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !v.hasMeasure(mv, field) {
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
func (v rendererValidator) validateOptionalColorDimensionField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any) error {
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
	return v.validateOptionalDimensionField(mv, mvName, props, "color.field")
}

// getPathStringSlice extracts a []string from a nested path in the props map.
// Returns (nil, nil) if the path is absent. Returns an error if the value is
// present but malformed (not an array, or an array containing non-strings).
func getPathStringSlice(props map[string]any, path string) ([]string, error) {
	raw, ok := pathutil.GetPath(props, path)
	if !ok {
		return nil, nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("renderer property %q is malformed: must be an array of strings", path)
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("renderer property %q is malformed: must be an array of strings", path)
		}
		result = append(result, s)
	}
	return result, nil
}

// getOptionalPathString extracts a string from a nested path in the props map.
// Returns ("", false, nil) if the path is absent, (value, true, nil) if present
// and a string, and an error if present but not a string.
func getOptionalPathString(props map[string]any, path string) (string, bool, error) {
	raw, ok := pathutil.GetPath(props, path)
	if !ok {
		return "", false, nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", false, fmt.Errorf("renderer property %q is malformed: must be a string", path)
	}
	return s, true, nil
}

// validateOptionalStringEnum validates that an optional string field at the given
// path, if present, equals one of the allowed values.
// In lenient mode, templated values are skipped since they only resolve at render time.
func (v rendererValidator) validateOptionalStringEnum(props map[string]any, path string, allowed []string) error {
	value, ok, err := getOptionalPathString(props, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if v.lenient && strings.Contains(value, "{{") {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("renderer property %q must be one of %v, got %q", path, allowed, value)
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

// isEncodedTimeDimension reports whether fieldName is an encoded time-grain dimension of the
// form "{timeDimension}_rill_{GRAIN}" (e.g. "ts_rill_TIME_GRAIN_MONTH"). The canvas pivot and
// table frontends encode a time dimension at a chosen grain this way; it is decoded back into
// a proper aggregation dimension at query time on the client.
//
// Note: this only recognizes the metrics view's primary time dimension, matching the frontend,
// which currently only encodes the primary one. If pivots/tables start allowing secondary time
// dimensions (time-type dimensions declared in the dimension list), this check must be extended
// to recognize those prefixes as well.
func isEncodedTimeDimension(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	if mv.TimeDimension == "" {
		return false
	}
	grain, ok := strings.CutPrefix(fieldName, mv.TimeDimension+"_rill_")
	if !ok {
		return false
	}
	v, ok := runtimev1.TimeGrain_value[grain]
	return ok && v != int32(runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED)
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
