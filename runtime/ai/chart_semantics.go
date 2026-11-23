package ai

import (
	"strings"

	"github.com/rilldata/rill/runtime/ai/chartquery"
)

// Flint semantic types of chart fields, derived from the metrics view the same way the canvas custom viz derives them.
// This is a port of deriveFlintFields in web-common/src/features/custom-viz/flint/semantic-types.ts: keep the two in sync.
// A semantic type tells the chart compiler how to format, scale and color a field; a wrong one degrades to plain rendering,
// so only names of Flint's own registry are emitted.

// chartTimeDisplayName labels the time dimension, which carries no meaningful display name of its own.
// Keep in sync with TIME_DISPLAY_NAME in semantic-types.ts and timeDimensionDisplayName in runtime/canvas/fields.go.
const chartTimeDisplayName = "Time"

// chartTimeSemanticTypes maps the grain a time field is bucketed at to Flint's temporal type; finer grains fall back to DateTime.
var chartTimeSemanticTypes = map[string]string{
	"TIME_GRAIN_YEAR":    "Year",
	"TIME_GRAIN_QUARTER": "YearQuarter",
	"TIME_GRAIN_MONTH":   "YearMonth",
	"TIME_GRAIN_WEEK":    "YearWeek",
	"TIME_GRAIN_DAY":     "Date",
}

// chartIntegerTypes are the type codes of integer measures, which are counts often enough that Flint's integer tick strategy is the better default.
var chartIntegerTypes = map[string]bool{
	"INT8": true, "INT16": true, "INT32": true, "INT64": true, "INT128": true, "INT256": true,
	"UINT8": true, "UINT16": true, "UINT32": true, "UINT64": true, "UINT128": true, "UINT256": true,
}

// chartSemantics derives the Flint semantic type and display name of every field of a chart result.
// The comparison value of a measure (<measure>_prev) takes the measure's type and a "(previous period)" display name.
// overrides are the caller's semantic_types: a type name per field, or an annotation object with a semanticType.
func chartSemantics(fields map[string]*ChartField, timeGrain string, comparison bool, overrides map[string]any) (semanticTypes, displayNames map[string]string) {
	semanticTypes = make(map[string]string, len(fields))
	displayNames = make(map[string]string, len(fields))
	for name, f := range fields {
		semanticTypes[name] = fieldSemanticType(f, timeGrain)
		displayNames[name] = f.DisplayName
		if f.Kind == "time" {
			displayNames[name] = chartTimeDisplayName
		}
		if displayNames[name] == "" {
			displayNames[name] = name
		}
	}
	for name, override := range overrides {
		if _, ok := fields[name]; !ok {
			continue
		}
		switch v := override.(type) {
		case string:
			if v != "" {
				semanticTypes[name] = v
			}
		case map[string]any:
			if t, ok := v["semanticType"].(string); ok && t != "" {
				semanticTypes[name] = t
			}
		}
	}
	if comparison {
		for name, f := range fields {
			if f.Kind != "measure" {
				continue
			}
			prev := name + chartquery.PrevSuffix
			semanticTypes[prev] = semanticTypes[name]
			displayNames[prev] = displayNames[name] + " (previous period)"
		}
	}
	return semanticTypes, displayNames
}

// fieldSemanticType picks the type of a field from its role: measures by how they are formatted (the preset encodes the intent:
// money, ratio, elapsed time), time fields by the grain they are bucketed at, and dimensions are categories.
func fieldSemanticType(f *ChartField, timeGrain string) string {
	switch f.Kind {
	case "time":
		if t, ok := chartTimeSemanticTypes[timeGrain]; ok {
			return t
		}
		return "DateTime"
	case "dimension":
		return "Category"
	}
	switch f.FormatPreset {
	case "currency_usd", "currency_eur":
		return "Amount"
	case "percentage":
		return "Percentage"
	case "interval_ms":
		return "Duration"
	}
	if strings.Contains(f.FormatD3, "$") {
		return "Amount"
	}
	if strings.Contains(f.FormatD3, "%") {
		return "Percentage"
	}
	if chartIntegerTypes[f.Type] {
		return "Count"
	}
	return "Number"
}
