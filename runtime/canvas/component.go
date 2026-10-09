package canvas

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
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
		return v.validateCartesianChart(renderer, props)
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
	case "map":
		return v.validateMap(props)
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
// metrics_sql queries are validated at query time; vega_spec (inline canvas charts) is validated as
// JSON only when it contains no template placeholders; spec (Flint, component files) is checked
// structurally. Which of the two is permitted is enforced at parse time, not here.
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

	if raw, ok := props["spec"]; ok {
		if err := validateFlintSpec(raw); err != nil {
			return err
		}
	}

	return nil
}

// validateFlintSpec structurally checks a Flint chart spec. It deliberately does not check chartType
// against Flint's registry: the registry lives in the JS package, so the frontend validates that.
func validateFlintSpec(raw any) error {
	spec, ok := raw.(map[string]any)
	if !ok {
		return errors.New("renderer property 'spec' must be a mapping")
	}

	if v, ok := spec["chartType"]; ok {
		s, ok := v.(string)
		if !ok {
			return errors.New("renderer property 'spec.chartType' must be a string")
		}
		if strings.TrimSpace(s) == "" {
			return errors.New("renderer property 'spec.chartType' must not be empty")
		}
	}

	raw, ok = spec["encodings"]
	if !ok {
		return nil
	}
	encodings, ok := raw.(map[string]any)
	if !ok {
		return errors.New("renderer property 'spec.encodings' must be a mapping")
	}
	for channel, raw := range encodings {
		// A channel is either a field-name shorthand, an encoding object, or an array of either.
		entries, ok := raw.([]any)
		if !ok {
			entries = []any{raw}
		}
		for _, entry := range entries {
			switch e := entry.(type) {
			case string:
				// Field-name shorthand; nothing to check.
			case map[string]any:
				if v, ok := e["field"]; ok {
					if _, ok := v.(string); !ok {
						return fmt.Errorf("renderer property 'spec.encodings.%s.field' must be a string", channel)
					}
				}
			default:
				return fmt.Errorf("renderer property 'spec.encodings.%s' must be a field name or an encoding mapping", channel)
			}
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

// The dimension is normally drawn on x and the measure on y.
// The bar renderers also accept the measure on x (a quantitative x.type) and the dimension on y, which draws horizontal bars.
func (v rendererValidator) validateCartesianChart(renderer string, props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	dimensionAxis, measureAxis := "x", "y"
	xType, _, err := getOptionalPathString(props, "x.type")
	if err != nil {
		return err
	}
	if xType == "quantitative" {
		switch renderer {
		case "bar_chart", "stacked_bar", "stacked_bar_normalized":
			dimensionAxis, measureAxis = "y", "x"
		default:
			return fmt.Errorf("renderer %q requires a dimension on x; a quantitative x (horizontal layout) is only supported by bar charts", renderer)
		}
	}

	dimensionField, ok := pathutil.GetPathString(props, dimensionAxis+".field")
	if !ok {
		return fmt.Errorf("renderer properties must include a string '%s.field' property", dimensionAxis)
	}
	if !v.hasDimension(mv, dimensionField) {
		return fmt.Errorf("referenced %s.field %q is not a dimension in metrics view %q", dimensionAxis, dimensionField, mvn)
	}

	measureField, ok := pathutil.GetPathString(props, measureAxis+".field")
	if !ok {
		return fmt.Errorf("renderer properties must include a string '%s.field' property", measureAxis)
	}
	if !v.hasMeasure(mv, measureField, ephemeralNames) {
		return fmt.Errorf("referenced %s.field %q is not a measure in metrics view %q", measureAxis, measureField, mvn)
	}

	// Validate optional multi-field measures (e.g. y.fields)
	measureFields, err := getPathStringSlice(props, measureAxis+".fields")
	if err != nil {
		return err
	}
	for _, f := range measureFields {
		if !v.hasMeasure(mv, f, ephemeralNames) {
			return fmt.Errorf("referenced %s.fields value %q is not a measure in metrics view %q", measureAxis, f, mvn)
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	measureField, ok := pathutil.GetPathString(props, "measure.field")
	if !ok {
		return errors.New("renderer properties must include a string 'measure.field' property")
	}
	if !v.hasMeasure(mv, measureField, ephemeralNames) {
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	xField, ok := pathutil.GetPathString(props, "x.field")
	if !ok {
		return errors.New("renderer properties must include a string 'x.field' property")
	}
	if !v.hasMeasure(mv, xField, ephemeralNames) {
		return fmt.Errorf("referenced x.field %q is not a measure in metrics view %q", xField, mvn)
	}

	yField, ok := pathutil.GetPathString(props, "y.field")
	if !ok {
		return errors.New("renderer properties must include a string 'y.field' property")
	}
	if !v.hasMeasure(mv, yField, ephemeralNames) {
		return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
	}

	if err := v.validateOptionalDimensionField(mv, mvn, props, "dimension.field"); err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "size.field", ephemeralNames); err != nil {
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "measure.field", ephemeralNames); err != nil {
		return err
	}

	// Validate optional multi-field measures (measure.fields)
	fields, err := getPathStringSlice(props, "measure.fields")
	if err != nil {
		return err
	}
	for _, f := range fields {
		if !v.hasMeasure(mv, f, ephemeralNames) {
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	// Note: for heatmap, color is a measure (not a dimension like other charts)
	return v.validateOptionalMeasureField(mv, mvn, props, "color.field", ephemeralNames)
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "y1.field", ephemeralNames); err != nil {
		return err
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "y2.field", ephemeralNames); err != nil {
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
	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}
	if !v.hasMeasure(mv, measure, ephemeralNames) {
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
	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}
	for _, m := range measures {
		if !v.hasMeasure(mv, m, ephemeralNames) {
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

	columns, err := getPathFieldList(props, "columns")
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return errors.New("renderer properties for table must include a non-empty 'columns' array")
	}
	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}
	sortable := make([]string, 0, len(columns))
	for _, col := range columns {
		isMeasure := v.hasMeasure(mv, col.Name, ephemeralNames)
		if !v.hasDimension(mv, col.Name) && !isMeasure && !isEncodedTimeDimension(mv, col.Name) {
			return fmt.Errorf("referenced columns value %q is not a dimension or measure in metrics view %q", col.Name, mvn)
		}
		sortable = append(sortable, col.Name)
		// In lenient mode an unresolved column's kind is unknown, so its overrides cannot be checked against a role.
		if v.lenient && (mv == nil || strings.Contains(col.Name, "{{")) {
			continue
		}
		role := fieldRoleTableDimension
		if metricsViewHasMeasure(mv, col.Name) {
			role = fieldRoleMeasure
		} else if isMeasure {
			role = fieldRoleAdhocMeasure
		}
		if err := validateFieldEntryProps("columns", col, role); err != nil {
			return err
		}
	}

	return v.validateTablePresentation(props, sortable, func(name string) bool { return v.hasMeasure(mv, name) })
}

// validatePivot validates properties for pivot.
func (v rendererValidator) validatePivot(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	measures, err := getPathFieldList(props, "measures")
	if err != nil {
		return err
	}
	rowDims, err := getPathFieldList(props, "row_dimensions")
	if err != nil {
		return err
	}
	colDims, err := getPathFieldList(props, "col_dimensions")
	if err != nil {
		return err
	}

	if len(measures) == 0 && len(rowDims) == 0 && len(colDims) == 0 {
		return errors.New("renderer properties for pivot must include at least one of 'measures', 'row_dimensions', or 'col_dimensions'")
	}

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}
	// Measures and row dimensions can be sorted by; column dimensions cannot.
	sortable := make([]string, 0, len(measures)+len(rowDims))
	for _, m := range measures {
		role := fieldRoleMeasure
		if !v.hasMeasure(mv, m.Name) {
			if !ephemeralNames[m.Name] {
				return fmt.Errorf("referenced measures value %q is not a measure in metrics view %q", m.Name, mvn)
			}
			role = fieldRoleAdhocMeasure
		}
		if err := validateFieldEntryProps("measures", m, role); err != nil {
			return err
		}
		sortable = append(sortable, m.Name)
	}
	for i, d := range rowDims {
		if !v.hasDimension(mv, d.Name) && !isEncodedTimeDimension(mv, d.Name) {
			return fmt.Errorf("referenced row_dimensions value %q is not a dimension in metrics view %q", d.Name, mvn)
		}
		// All row dimensions share one merged row-header column keyed by the first one.
		role := fieldRoleRowDimension
		if i == 0 {
			role = fieldRoleRowHeader
		}
		if err := validateFieldEntryProps("row_dimensions", d, role); err != nil {
			return err
		}
		sortable = append(sortable, d.Name)
	}
	for _, d := range colDims {
		if !v.hasDimension(mv, d.Name) && !isEncodedTimeDimension(mv, d.Name) {
			return fmt.Errorf("referenced col_dimensions value %q is not a dimension in metrics view %q", d.Name, mvn)
		}
		if err := validateFieldEntryProps("col_dimensions", d, fieldRoleColumnDimension); err != nil {
			return err
		}
	}

	return v.validateTablePresentation(props, sortable, func(name string) bool { return v.hasMeasure(mv, name) })
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

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}
	for _, m := range measures {
		if !v.hasMeasure(mv, m, ephemeralNames) {
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

// validateMap validates properties for map.
func (v rendererValidator) validateMap(props map[string]any) error {
	mvn, mv, err := v.requireMetricsView(props)
	if err != nil {
		return err
	}

	ephemeralNames, err := ephemeralMeasureNames(props, mvn, mv)
	if err != nil {
		return err
	}

	geoDim, ok := pathutil.GetPathString(props, "geo_dimension.field")
	if !ok {
		return errors.New("renderer properties for map must include a string 'geo_dimension.field' property")
	}
	if !v.hasDimension(mv, geoDim) {
		return fmt.Errorf("referenced geo_dimension.field %q is not a dimension in metrics view %q", geoDim, mvn)
	}

	// Color on a map is always measure-driven.
	colorMeasure, ok := pathutil.GetPathString(props, "color.measure")
	if !ok {
		return errors.New("renderer properties for map must include a string 'color.measure' property")
	}
	if !v.hasMeasure(mv, colorMeasure, ephemeralNames) {
		return fmt.Errorf("referenced color.measure %q is not a measure in metrics view %q", colorMeasure, mvn)
	}

	if err := v.validateOptionalMeasureField(mv, mvn, props, "size_measure.field", ephemeralNames); err != nil {
		return err
	}

	return v.validateOptionalDimensionField(mv, mvn, props, "tooltip_dimension.field")
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
func (v rendererValidator) hasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string, ephemeralNames ...map[string]bool) bool {
	if v.lenient && (mv == nil || strings.Contains(fieldName, "{{")) {
		return true
	}
	for _, names := range ephemeralNames {
		if names[fieldName] {
			return true
		}
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

// validateOptionalMeasureField validates that a field at the given path, if present,
// is a measure in the metrics view or one of the component's ephemeral measures.
func (v rendererValidator) validateOptionalMeasureField(mv *runtimev1.MetricsViewSpec, mvName string, props map[string]any, path string, ephemeralNames ...map[string]bool) error {
	field, ok, err := getOptionalPathString(props, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !v.hasMeasure(mv, field, ephemeralNames...) {
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

// fieldEntry is an entry of a table's or pivot's field list (columns, measures, row_dimensions, col_dimensions):
// a field name, optionally with per-column presentation overrides when written as an object.
type fieldEntry struct {
	Name  string
	Props map[string]any
}

// getPathFieldList extracts a field list from the props map.
// Each entry is either a string (the field name) or an object with a non-empty "name" and optional overrides.
// Returns (nil, nil) if the path is absent and an error if an entry is malformed or a name repeats.
func getPathFieldList(props map[string]any, path string) ([]fieldEntry, error) {
	raw, ok := pathutil.GetPath(props, path)
	if !ok {
		return nil, nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("renderer property %q is malformed: must be an array of field names or field objects", path)
	}
	result := make([]fieldEntry, 0, len(arr))
	seen := make(map[string]bool, len(arr))
	for _, v := range arr {
		var entry fieldEntry
		switch t := v.(type) {
		case string:
			entry.Name = t
		case map[string]any:
			name, _ := t["name"].(string)
			if name == "" {
				return nil, fmt.Errorf("renderer property %q is malformed: field objects must have a non-empty 'name'", path)
			}
			entry.Name = name
			entry.Props = t
		default:
			return nil, fmt.Errorf("renderer property %q is malformed: must be an array of field names or field objects", path)
		}
		if seen[entry.Name] {
			return nil, fmt.Errorf("renderer property %q lists field %q more than once", path, entry.Name)
		}
		seen[entry.Name] = true
		result = append(result, entry)
	}
	return result, nil
}

// Column width bounds in pixels, mirroring the resizer bounds of the frontend (pivot-column-width-utils.ts).
// A configured width outside these would snap back on the first drag.
const (
	minMeasureColumnWidth   = 60
	maxMeasureColumnWidth   = 300
	minDimensionColumnWidth = 100
	maxDimensionColumnWidth = 600
)

var columnAligns = []string{"left", "center", "right"}

var measureFormatPresets = []string{"humanize", "none", "currency_usd", "currency_eur", "percentage", "interval_ms"}

// fieldRole describes which per-column overrides a field entry may carry.
type fieldRole int

const (
	// fieldRoleTableDimension is a dimension or time column of a table: width, wrap, align and label.
	fieldRoleTableDimension fieldRole = iota
	// fieldRoleMeasure is a measure of the metrics view in a table or pivot: width, align, label, the number formats,
	// and the 'delta' and 'percent_change' objects for its comparison columns.
	fieldRoleMeasure
	// fieldRoleAdhocMeasure is an adhoc measure of the component: like a measure, but it has no comparison columns.
	fieldRoleAdhocMeasure
	// fieldRoleRowHeader is the first row dimension of a pivot, which keys the merged row-header column: width, wrap and label.
	fieldRoleRowHeader
	// fieldRoleRowDimension is a row dimension after the first, which renders no column of its own: label only.
	fieldRoleRowDimension
	// fieldRoleColumnDimension is a column dimension, a header group spanning its measure columns: label only.
	fieldRoleColumnDimension
)

// isMeasure reports whether the role renders a measure column: a measure of the metrics view or an adhoc measure.
func (r fieldRole) isMeasure() bool {
	return r == fieldRoleMeasure || r == fieldRoleAdhocMeasure
}

// validateFieldEntryProps validates the per-column overrides of a field entry against its role:
// width (an integer within the bounds of the role), wrap, align, label, and the number format keys.
// Unknown keys are ignored so that future keys need no lockstep validation.
func validateFieldEntryProps(path string, entry fieldEntry, role fieldRole) error {
	if entry.Props == nil {
		return nil
	}
	prefix := fmt.Sprintf("field %q in %q", entry.Name, path)

	if raw, ok := entry.Props["width"]; ok {
		switch role {
		case fieldRoleColumnDimension:
			return fmt.Errorf("%s: column dimensions span their measure columns and take no 'width'; set it on the measures instead", prefix)
		case fieldRoleRowDimension:
			return fmt.Errorf("%s: only the first row dimension renders a column; set 'width' on it", prefix)
		}
		// Numbers arrive as float64 after the structpb round trip.
		f, ok := raw.(float64)
		if !ok || f != math.Trunc(f) {
			return fmt.Errorf("%s: 'width' must be an integer number of pixels", prefix)
		}
		minWidth, maxWidth := minDimensionColumnWidth, maxDimensionColumnWidth
		if role.isMeasure() {
			minWidth, maxWidth = minMeasureColumnWidth, maxMeasureColumnWidth
		}
		if f < float64(minWidth) || f > float64(maxWidth) {
			return fmt.Errorf("%s: 'width' must be between %d and %d pixels", prefix, minWidth, maxWidth)
		}
	}

	if raw, ok := entry.Props["wrap"]; ok {
		switch role {
		case fieldRoleColumnDimension:
			return fmt.Errorf("%s: column dimensions take no 'wrap'", prefix)
		case fieldRoleRowDimension:
			return fmt.Errorf("%s: only the first row dimension renders a column; set 'wrap' on it", prefix)
		case fieldRoleMeasure, fieldRoleAdhocMeasure:
			return fmt.Errorf("%s: 'wrap' only applies to dimension columns", prefix)
		}
		if _, ok := raw.(bool); !ok {
			return fmt.Errorf("%s: 'wrap' must be a boolean", prefix)
		}
	}

	if raw, ok := entry.Props["align"]; ok {
		switch role {
		case fieldRoleColumnDimension:
			return fmt.Errorf("%s: column dimensions take no 'align'", prefix)
		case fieldRoleRowHeader, fieldRoleRowDimension:
			return fmt.Errorf("%s: 'align' is not supported on row dimensions", prefix)
		}
		s, ok := raw.(string)
		if !ok || !slices.Contains(columnAligns, s) {
			return fmt.Errorf("%s: 'align' must be one of %v", prefix, columnAligns)
		}
	}

	if raw, ok := entry.Props["label"]; ok {
		if _, ok := raw.(string); !ok {
			return fmt.Errorf("%s: 'label' must be a string", prefix)
		}
	}

	rawPreset, hasPreset := entry.Props["format_preset"]
	rawD3, hasD3 := entry.Props["format_d3"]
	if (hasPreset || hasD3) && !role.isMeasure() {
		return fmt.Errorf("%s: number formats only apply to measures", prefix)
	}
	if hasPreset && hasD3 {
		return fmt.Errorf("%s: cannot set both 'format_preset' and 'format_d3'", prefix)
	}
	if hasPreset {
		s, ok := rawPreset.(string)
		if !ok || !slices.Contains(measureFormatPresets, s) {
			return fmt.Errorf("%s: 'format_preset' must be one of %v", prefix, measureFormatPresets)
		}
	}
	if hasD3 {
		s, ok := rawD3.(string)
		if !ok || s == "" {
			return fmt.Errorf("%s: 'format_d3' must be a non-empty string", prefix)
		}
	}

	// The comparison columns a measure gets while the component compares.
	for _, key := range []string{"delta", "percent_change"} {
		raw, ok := entry.Props[key]
		if !ok {
			continue
		}
		switch role {
		case fieldRoleAdhocMeasure:
			return fmt.Errorf("%s: adhoc measures have no comparison columns, so '%s' is not allowed", prefix, key)
		case fieldRoleMeasure:
		default:
			return fmt.Errorf("%s: '%s' configures a measure's comparison column and only applies to measures", prefix, key)
		}
		nested, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: '%s' must be an object with the comparison column's overrides, such as 'width'", prefix, key)
		}
		if rawWidth, ok := nested["width"]; ok {
			f, ok := rawWidth.(float64)
			if !ok || f != math.Trunc(f) {
				return fmt.Errorf("%s: '%s.width' must be an integer number of pixels", prefix, key)
			}
			if f < minMeasureColumnWidth || f > maxMeasureColumnWidth {
				return fmt.Errorf("%s: '%s.width' must be between %d and %d pixels", prefix, key, minMeasureColumnWidth, maxMeasureColumnWidth)
			}
		}
	}

	return nil
}

// validateTablePresentation validates the presentation properties shared by tables and pivots:
// fit_to_width, wrap, wrap_headers, wrap_lines, sort_by, sort_dir and sort_comparison.
// sortable lists the field names that sort_by may reference;
// hasComparison reports whether the pivot builds comparison columns for a field (a measure of the metrics view, not an adhoc one).
func (v rendererValidator) validateTablePresentation(props map[string]any, sortable []string, hasComparison func(name string) bool) error {
	for _, key := range []string{"fit_to_width", "wrap", "wrap_headers"} {
		if _, _, err := getOptionalPathBool(props, key); err != nil {
			return err
		}
	}

	if raw, ok := props["wrap_lines"]; ok {
		f, ok := raw.(float64)
		if !ok || f != math.Trunc(f) || f < 1 || f > 5 {
			return errors.New("renderer property \"wrap_lines\" must be an integer between 1 and 5")
		}
	}

	sortBy, hasSortBy, err := getOptionalPathString(props, "sort_by")
	if err != nil {
		return err
	}
	if hasSortBy && sortBy != "" && !slices.Contains(sortable, sortBy) && !(v.lenient && strings.Contains(sortBy, "{{")) {
		return fmt.Errorf("renderer property \"sort_by\" references %q, which is not a sortable field of the component (one of %v)", sortBy, sortable)
	}
	if err := v.validateOptionalStringEnum(props, "sort_dir", []string{"asc", "desc"}); err != nil {
		return err
	}
	if _, hasSortDir := props["sort_dir"]; hasSortDir && (!hasSortBy || sortBy == "") {
		return errors.New("renderer property \"sort_dir\" requires \"sort_by\"")
	}

	if err := v.validateOptionalStringEnum(props, "sort_comparison", []string{"delta", "percent_change"}); err != nil {
		return err
	}
	if _, hasSortComparison := props["sort_comparison"]; hasSortComparison {
		if !hasSortBy || sortBy == "" {
			return errors.New("renderer property \"sort_comparison\" requires \"sort_by\"")
		}
		if !hasComparison(sortBy) {
			return fmt.Errorf("renderer property \"sort_comparison\" requires \"sort_by\" to name a measure of the metrics view, got %q", sortBy)
		}
	}

	return nil
}

// getOptionalPathBool extracts a bool from a nested path in the props map.
// Returns (false, false, nil) if the path is absent, (value, true, nil) if present
// and a bool, and an error if present but not a bool.
func getOptionalPathBool(props map[string]any, path string) (bool, bool, error) {
	raw, ok := pathutil.GetPath(props, path)
	if !ok {
		return false, false, nil
	}
	b, ok := raw.(bool)
	if !ok {
		return false, false, fmt.Errorf("renderer property %q is malformed: must be a boolean", path)
	}
	return b, true, nil
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
	if mv == nil {
		return false
	}
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
	if mv == nil || mv.TimeDimension == "" {
		return false
	}
	grain, ok := strings.CutPrefix(fieldName, mv.TimeDimension+"_rill_")
	if !ok {
		return false
	}
	v, ok := runtimev1.TimeGrain_value[grain]
	return ok && v != int32(runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED)
}

// ephemeralMeasureNames extracts and validates the optional "adhoc_measures" renderer property.
// Each entry defines an ephemeral measure derived from existing measures via an arithmetic expression;
// the returned set contains the names that may be referenced alongside the metrics view's own measures.
func ephemeralMeasureNames(props map[string]any, mvn string, mv *runtimev1.MetricsViewSpec) (map[string]bool, error) {
	raw, ok := props["adhoc_measures"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, errors.New("renderer property 'adhoc_measures' must be an array")
	}
	names := make(map[string]bool, len(list))
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, errors.New("entries in 'adhoc_measures' must be objects with 'name' and 'expression'")
		}
		name, _ := entry["name"].(string)
		expression, _ := entry["expression"].(string)
		if name == "" || expression == "" {
			return nil, errors.New("entries in 'adhoc_measures' must have a non-empty 'name' and 'expression'")
		}
		// Mirror metricsview.AST.checkNameForComputedField, which also rejects the time dimension.
		// It is often absent from mv.Dimensions, so checking it here surfaces the collision at parse time rather than at query time.
		if mv != nil && (metricsViewHasMeasure(mv, name) || metricsViewHasDimension(mv, name) || name == mv.TimeDimension) {
			return nil, fmt.Errorf("ephemeral measure %q collides with a field in metrics view %q", name, mvn)
		}
		if names[name] {
			return nil, fmt.Errorf("duplicate ephemeral measure %q", name)
		}
		parsed, err := metricsview.ParseMeasureExpression(expression)
		if err != nil {
			return nil, fmt.Errorf("ephemeral measure %q: %w", name, err)
		}
		for _, ref := range parsed.Refs() {
			if mv != nil && !metricsViewHasMeasure(mv, ref) {
				return nil, fmt.Errorf("ephemeral measure %q references %q, which is not a measure in metrics view %q", name, ref, mvn)
			}
		}
		names[name] = true
	}
	return names, nil
}

// metricsViewHasMeasure returns true if the metrics view has a measure with the given name.
func metricsViewHasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	if mv == nil {
		return false
	}
	for _, m := range mv.Measures {
		if m.Name == fieldName {
			return true
		}
	}
	return false
}
