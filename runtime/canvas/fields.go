package canvas

import (
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// timeDimensionDisplayName labels the metrics view's primary time dimension.
// A metrics view carries no meaningful display name for its timestamp column,
// so charts would otherwise label axes and tooltips with the raw column name.
// Keep in sync with TIME_DISPLAY_NAME in web-common/src/features/custom-viz/flint/semantic-types.ts.
const timeDimensionDisplayName = "Time"

// FieldTemplateData builds the ".fields" template namespace: the metrics view metadata of every
// field-typed param, keyed by param name.
//
// Params substitute a field's *name* through ".params.<name>", which is all a query needs.
// A hand-authored Vega-Lite spec also needs the field's presentation metadata (what to title an
// axis, how to format a value), and that lives in the metrics view rather than in the binding.
// Exposing it per param rather than per field name keeps the lookup unambiguous when a component
// declares several metrics_view params: each field-typed param names the one it resolves against.
//
// The exposed keys are:
//
//	name:          the bound field name, same as .params.<name>
//	display_name:  the field's display name ("Time" for the primary time dimension)
//	type:          the param's declared type ("measure", "dimension" or "time_dimension")
//	format_d3:     the measure's d3 format string, if it declares one
//	format_preset: the measure's format preset, if it declares one
//	format_type:   the name of the Vega expression formatter Rill registers for the measure,
//	               for use as "formatType" in a vega_spec
//
// Every field-typed param gets an entry, even one whose metrics view or field cannot be resolved:
// its keys are then empty strings, so a reference degrades to nothing rather than to Go's
// "<no value>" placeholder, which would reach the renderer as a field or formatter name.
func FieldTemplateData(params []*runtimev1.ComponentParam, args map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) map[string]any {
	res := make(map[string]any)
	for _, p := range params {
		switch p.Type {
		case "measure", "dimension", "time_dimension":
		default:
			continue
		}

		field, ok := args[p.Name].(string)
		if !ok || isTemplated(field) {
			field = ""
		}

		var mv *runtimev1.MetricsViewSpec
		for _, mvp := range params {
			if mvp.Name != p.MetricsViewParam {
				continue
			}
			if name, ok := args[mvp.Name].(string); ok {
				mv = metricsViews[name]
			}
			break
		}

		res[p.Name] = fieldMetadata(mv, p.Type, field)
	}
	return res
}

// FieldMetadataKeys are the keys ".fields.<param>" exposes.
// The parser validates references against them, since a mistyped key resolves to Go's
// "<no value>" placeholder rather than to an error.
var FieldMetadataKeys = []string{"name", "display_name", "type", "format_d3", "format_preset", "format_type"}

// fieldMetadata resolves one field's metadata from the metrics view it belongs to.
func fieldMetadata(mv *runtimev1.MetricsViewSpec, paramType, field string) map[string]any {
	res := map[string]any{
		"name":          field,
		"display_name":  field,
		"type":          paramType,
		"format_d3":     "",
		"format_preset": "",
		"format_type":   "",
	}
	if mv == nil || field == "" {
		return res
	}

	for _, m := range mv.Measures {
		if m.Name != field {
			continue
		}
		if m.DisplayName != "" {
			res["display_name"] = m.DisplayName
		}
		res["format_d3"] = m.FormatD3
		res["format_preset"] = m.FormatPreset
		res["format_type"] = VegaFormatType(field)
		return res
	}

	// The primary time dimension is not always present in the dimensions list, and the display name
	// it does carry defaults to the column name, so it is replaced rather than used.
	if mv.TimeDimension == field {
		res["display_name"] = timeDimensionDisplayName
		return res
	}

	for _, d := range mv.Dimensions {
		if d.Name == field {
			if d.DisplayName != "" {
				res["display_name"] = d.DisplayName
			}
			return res
		}
	}

	return res
}

// VegaFormatType returns the name of the Vega expression function that Rill registers to format
// values of the given measure. Vega-Lite compiles a custom "formatType" into an expression function
// call, so the name is reduced to a JavaScript identifier-safe subset and measure names with spaces
// or operators stay usable.
//
// Keep in sync with sanitizeFieldName in web-common/src/components/vega/util.ts.
func VegaFormatType(measure string) string {
	var b strings.Builder
	for _, r := range measure {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '$':
			b.WriteRune(r)
		default:
			b.WriteString("_u")
			b.WriteString(strings.ToLower(strconv.FormatInt(int64(r), 16)))
			b.WriteString("_")
		}
	}
	if b.Len() == 0 {
		return "rill_field"
	}
	return "rill_" + b.String()
}
