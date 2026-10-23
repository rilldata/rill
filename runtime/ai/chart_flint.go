package ai

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/chartquery"
)

// The Flint shape of create_chart and chart_data, served to external MCP clients (see CreateChart.MCPSpec):
// a Flint chart spec (chartType, encodings, chartProperties) bound to a metrics view through a data block,
// as opposed to the in-app shape (chart_type and spec) that Rill's own chat renders itself.
// The chart types come from the embedded Flint catalog (see chartquery.Catalog).

// ChartInput is the effective Flint input echoed in a create_chart result:
// the chart spec with the applied top-N limits and the data binding with the resolved grain.
type ChartInput struct {
	ChartSpec map[string]any `json:"chart_spec"`
	Data      map[string]any `json:"data"`
}

// flintChartNotes are the rules of the chart types that need more than their channel list, shown in the tool description.
var flintChartNotes = map[string]string{
	"Bar Chart":              "one dimension on x and one measure on y; a second dimension on color stacks segments; put the measure on x and the dimension on y for horizontal bars",
	"Grouped Bar Chart":      "side-by-side bars: the clustering dimension goes on group, not on color",
	"Stacked Bar Chart":      "parts of a whole: the stacked dimension on color; chartProperties.stackMode is stacked, normalize or layered",
	"Lollipop Chart":         "a bar chart drawn as stems with dots",
	"Waterfall Chart":        "x is the step (the time field or a dimension), y the signed change; leave color unbound (it is reserved for a start/delta/end type field); chartProperties.totals is first, last or both",
	"Bullet Chart":           "y the dimension, x the measure, goal the target measure (required)",
	"Pie Chart":              "the measure on size, the dimension on color; chartProperties.innerRadius above 0 makes it a donut",
	"Donut Chart":            "the measure on size, the dimension on color",
	"Rose Chart":             "polar bars: the dimension on x, the measure on y",
	"Radar Chart":            "the dimension on x (one axis per value), the measure on y, an optional series dimension on color",
	"Heatmap":                "two dimensions on x and y (one may be the time field), the measure on color",
	"Bar Table":              "the dimension on y, the measure on x: compact bars with value labels",
	"Line Chart":             "the time field on x, the measure on y, an optional series dimension on color; several measures on y draw one line each",
	"Area Chart":             "like a line chart with filled areas; chartProperties.stackMode stacks a series dimension",
	"Sparkline":              "one small trend line per value of the series dimension (color)",
	"Bump Chart":             "rank over time: the time field on x, the measure on y, the ranked dimension on color",
	"Slope Chart":            "two periods: the time field on x at a grain that yields two buckets, the measure on y, the dimension on color",
	"Streamgraph":            "stacked areas around a center line: the time field on x, the measure on y, the dimension on color",
	"Range Area Chart":       "a band between two measures: the low measure on y and the high one on y2",
	"Pyramid Chart":          "back-to-back bars: the measure on x, the band dimension on y, a two-valued dimension on color",
	"Ranged Dot Plot":        "a dumbbell per category: the measure on x, the dimension on y, a two-valued dimension on color",
	"Candlestick Chart":      "the time field on x and four measures on open, high, low and close",
	"Gantt Chart":            "needs a start on x and an end on x2; rarely expressible with a metrics view",
	"Scatter Plot":           "two measures on x and y, one point per value of the dimension on detail; optional color and size",
	"Regression":             "a scatter plot with a fitted line: two measures on x and y, the dimension on detail",
	"Connected Scatter Plot": "two measures on x and y connected in the order of the time field on order",
	"Histogram":              "the measure on x, binned over one row per value of the dimension on detail",
	"Density Plot":           "the measure on x, estimated over one row per value of the dimension on detail",
	"ECDF Plot":              "the measure on x, cumulated over one row per value of the dimension on detail",
	"Boxplot":                "the dimension on x, the measure on y, one observation per value of the dimension on detail",
	"Violin Plot":            "the dimension on x, the measure on y, one observation per value of the dimension on detail",
	"Strip Plot":             "the dimension on x, the measure on y, one point per value of the dimension on detail",
}

// flintChartsInputSchema returns the JSON schema of the Flint shape of the chart arguments.
// It reuses the $defs of the in-app schema (Expression, TimeGrain, ...) for the data block.
func flintChartsInputSchema() *jsonschema.Schema {
	encoding := &jsonschema.Schema{
		Type:        "object",
		Required:    []string{"field"},
		Description: "A field bound to a channel with optional sorting and top-N.",
		Properties: map[string]*jsonschema.Schema{
			"field":     {Type: "string", Description: "A measure, dimension or the time dimension of the metrics view."},
			"type":      {Type: "string", Enum: []any{"quantitative", "nominal", "ordinal", "temporal"}, Description: "Overrides the type derived from the field; rarely needed."},
			"aggregate": {Type: "string", Enum: []any{"count", "sum", "average", "mean"}, Description: "Rarely needed: measures are already aggregated."},
			"sortOrder": {Type: "string", Enum: []any{"ascending", "descending"}, Description: "Direction of the sort; descending by default for a sort by a measure."},
			"sortBy":    {Type: "string", Description: "Channel (e.g. \"y\") or field to sort this channel's values by; the first measure by default."},
			"scheme":    {Type: "string", Description: "Color scheme for the color channel, e.g. viridis or tableau10."},
			"limit":     {Type: "integer", Minimum: ptr(1.0), Description: "Top-N limit for a dimension on this channel: defaults to 20 on x and y, 10 on color, group, shape and strokeDash, 6 on column and row."},
		},
	}
	encodingValue := &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Type: "string", Description: "A field name."},
			{Ref: "#/$defs/FlintEncoding"},
			{Type: "array", Items: &jsonschema.Schema{AnyOf: []*jsonschema.Schema{{Type: "string"}, {Ref: "#/$defs/FlintEncoding"}}}, Description: "Several fields on one channel, e.g. several measures on y as one series each."},
		},
	}
	chartSpec := &jsonschema.Schema{
		Type:        "object",
		Required:    []string{"chartType", "encodings"},
		Description: "What to draw: the chart type and which field goes on which channel. Rill derives scales, axes, formats and colors.",
		Properties: map[string]*jsonschema.Schema{
			"chartType":       {Type: "string", Enum: anyStrings(chartquery.Catalog().Names()), Description: "The chart type, exactly as listed in the tool description."},
			"title":           {Type: "string", Description: "Headline shown above the chart; optional."},
			"subtitle":        {Type: "string", Description: "What is measured, of whom, when; optional."},
			"encodings":       {Type: "object", AdditionalProperties: encodingValue, Description: "Channel -> field, e.g. {\"x\": \"advertiser\", \"y\": {\"field\": \"total_bids\"}}. The channels of each chart type are listed in the tool description; detail binds a dimension that sets the row grain without being drawn."},
			"chartProperties": {Type: "object", Description: "Per-chart-type tuning, e.g. {\"stackMode\": \"normalize\"} or {\"innerRadius\": 50}; the keys of each chart type are listed in the tool description."},
		},
	}
	legacy := chartsInputSchema()
	data := &jsonschema.Schema{
		Type:        "object",
		Required:    []string{"metrics_view", "time_range"},
		Description: "The data binding: which metrics view to query and over which time range. Rill runs the query; never pass rows.",
		Properties: map[string]*jsonschema.Schema{
			"metrics_view":          {Type: "string", Description: "The metrics view to query."},
			"time_range":            {Type: "object", Required: []string{"start", "end"}, Properties: map[string]*jsonschema.Schema{"start": {Type: "string", Description: "ISO 8601 start, inclusive."}, "end": {Type: "string", Description: "ISO 8601 end, exclusive."}, "time_zone": {Type: "string", Description: "IANA time zone; defaults to UTC."}}, Description: "The period to chart."},
			"comparison_time_range": {Type: "object", Required: []string{"start", "end"}, Properties: map[string]*jsonschema.Schema{"start": {Type: "string"}, "end": {Type: "string"}}, Description: "Optional period to compare with: every measure also returns its value in this period, drawn as a second series."},
			"where":                 {Ref: "#/$defs/Expression", Description: "Optional filter, with the same structure as in query_metrics_view."},
			"time_grain":            {Ref: "#/$defs/TimeGrain", Description: "Bucket size of the time field when one is bound; defaults to TIME_GRAIN_DAY."},
		},
	}
	defs := maps.Clone(legacy.Defs)
	defs["FlintEncoding"] = encoding
	// The root requires neither shape: hosts cache a connector's tool list across conversations, so a client that still holds
	// the previous shape (chart_type and spec) keeps passing validation and gets the handler's guidance instead of a protocol error.
	// Both of its properties are described as deprecated so a client that reads this schema uses the Flint shape.
	deprecated := "Deprecated: the previous shape of this tool, kept for clients with a cached tool list. Use chart_spec and data instead."
	return &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"chart_spec":     chartSpec,
			"data":           data,
			"chart_type":     {Type: "string", Description: deprecated},
			"spec":           {Ref: "#/$defs/ChartSpec", Description: deprecated},
			"semantic_types": {Type: "object", AdditionalProperties: &jsonschema.Schema{Type: "string"}, Description: "Optional overrides of the semantic type Rill derives per field (e.g. {\"revenue\": \"Amount\"}); only needed when the derived type is wrong."},
		},
		Defs: defs,
	}
}

// flintCreateChartDescription is the description of create_chart for external MCP clients, built from the catalog.
var flintCreateChartDescription = sync.OnceValue(func() string {
	var b strings.Builder
	b.WriteString(`# Chart Visualization Tool

Create a chart from a metrics view. The chart is rendered interactively in clients that support it; other clients receive the chart's data as a table together with a link. Whenever the result includes an ` + "`open_url`" + `, surface it to the user so they can open the chart in Rill.

Query the data with query_metrics_view first, then describe the chart: the chart type and which field goes on which channel. Rill runs the query, derives scales, axes, number formats, sorting, stacking, colors and layout, and returns the rows. Do not pass colors, formats or axis settings, and reference only measures, dimensions and the time dimension of the metrics view. The rill-charts skill (in the ai_instructions of list_metrics_views, or via load_skill) says which chart fits which question and what to check first.

## Arguments

` + "```json" + `
{
  "chart_spec": {
    "chartType": "Bar Chart",
    "title": "Bids by advertiser",
    "encodings": {
      "x": {"field": "advertiser_name", "limit": 10, "sortBy": "y", "sortOrder": "descending"},
      "y": "total_bids",
      "color": "device_os"
    },
    "chartProperties": {"cornerRadius": 2}
  },
  "data": {
    "metrics_view": "bids_metrics",
    "time_range": {"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
    "time_grain": "TIME_GRAIN_DAY"
  }
}
` + "```" + `

- ` + "`chart_spec.encodings`" + ` maps a channel to a field name or to an object with ` + "`field`" + ` and optionally ` + "`sortOrder`" + ` (ascending or descending), ` + "`sortBy`" + ` (a channel such as "y", or a field) and ` + "`limit`" + ` (top-N for a dimension; defaults to 20 on an axis, 10 on a series channel, 6 on a facet). An array of fields on one channel draws one series per field.
- ` + "`detail`" + ` may be bound on any chart type to a dimension that sets the row grain without being drawn: one row per value. Charts of individual observations (scatter, histogram, box plots) need it.
- ` + "`data.time_range`" + ` is required (ISO 8601 ` + "`start`" + ` inclusive, ` + "`end`" + ` exclusive, optional ` + "`time_zone`" + `). ` + "`data.time_grain`" + ` buckets the time field (default TIME_GRAIN_DAY). ` + "`data.where`" + ` filters with the same expression structure as query_metrics_view. ` + "`data.comparison_time_range`" + ` adds the previous period of every measure as a second series; use it with a time field on x or a dimension axis.
- Results are capped (at most 1500 rows); the result says when data was cut and how to narrow the chart.

## Chart types and their channels

`)
	for _, ct := range chartquery.Catalog().ChartTypes {
		fmt.Fprintf(&b, "- %s: %s", ct.Name, strings.Join(ct.Channels, ", "))
		if note, ok := flintChartNotes[ct.Name]; ok {
			b.WriteString(". ")
			b.WriteString(note)
		}
		if keys := propertyKeys(ct); len(keys) > 0 {
			fmt.Fprintf(&b, ". chartProperties: %s", strings.Join(keys, ", "))
		}
		b.WriteString("\n")
	}
	b.WriteString(`
Rule of thumb for bars: a single series is a Bar Chart; parts of a whole are a Stacked Bar Chart with the parts on color; side-by-side values are a Grouped Bar Chart with the clustering dimension on group.
`)
	return b.String()
})

// propertyKeys lists the tunable properties of a chart type, leaving out the axis settings Rill derives.
func propertyKeys(ct *chartquery.FlintChartType) []string {
	var keys []string
	for _, p := range ct.Properties {
		switch p.Key {
		case "independentYAxis", "logScale_x", "logScale_y", "includeZero_x", "includeZero_y", "xAxisType", "yAxisType":
			continue
		}
		keys = append(keys, p.Key)
	}
	return keys
}

// parseFlintChartRequest validates the Flint shape of the chart arguments against the metrics view the caller has access to.
func parseFlintChartRequest(ctx context.Context, rt *runtime.Runtime, s *Session, args map[string]any) (*chartRequest, error) {
	chartSpec, ok := args["chart_spec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("chart_spec is required and must be an object")
	}
	data, ok := args["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is required and must be an object with 'metrics_view' and 'time_range'")
	}
	metricsView, ok := data["metrics_view"].(string)
	if !ok || metricsView == "" {
		return nil, fmt.Errorf("data must contain a 'metrics_view' field")
	}
	mv, err := loadMetricsView(ctx, rt, s, metricsView)
	if err != nil {
		return nil, err
	}

	parsed, err := chartquery.ParseFlintSpec(chartSpec, data, fieldKinds(mv))
	if err != nil {
		return nil, err
	}
	semanticTypes, _ := args["semantic_types"].(map[string]any)

	return &chartRequest{
		chartType:     parsed.ChartType,
		flint:         parsed,
		input:         &ChartInput{ChartSpec: chartSpec, Data: data},
		semanticTypes: semanticTypes,
		mv:            mv,
	}, nil
}

// fieldKinds classifies the fields of a metrics view: its time dimension and dimensions of the time type are time fields.
func fieldKinds(mv *runtimev1.MetricsViewSpec) map[string]chartquery.FieldKind {
	kinds := make(map[string]chartquery.FieldKind, len(mv.Measures)+len(mv.Dimensions)+1)
	for _, m := range mv.Measures {
		kinds[m.Name] = chartquery.FieldKindMeasure
	}
	for _, d := range mv.Dimensions {
		if d.Type == runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME {
			kinds[d.Name] = chartquery.FieldKindTime
		} else {
			kinds[d.Name] = chartquery.FieldKindDimension
		}
	}
	if mv.TimeDimension != "" {
		kinds[mv.TimeDimension] = chartquery.FieldKindTime
	}
	return kinds
}

// flintInputWithLimits returns the effective input: the chart spec with the applied top-N limit of each channel filled in
// (as a copy; the input maps are not modified) and the data block with the resolved time grain.
func flintInputWithLimits(input *ChartInput, limits map[string]int, timeGrain string) *ChartInput {
	chartSpec := maps.Clone(input.ChartSpec)
	if encodings, ok := input.ChartSpec["encodings"].(map[string]any); ok && len(limits) > 0 {
		encodings = maps.Clone(encodings)
		for channel, limit := range limits {
			value, ok := encodings[channel]
			if !ok {
				continue
			}
			encodings[channel] = encodingWithLimit(value, limit)
		}
		chartSpec["encodings"] = encodings
	}
	data := maps.Clone(input.Data)
	if timeGrain != "" {
		data["time_grain"] = timeGrain
	}
	return &ChartInput{ChartSpec: chartSpec, Data: data}
}

// encodingWithLimit returns the encoding value with the limit set: a field name becomes an object,
// an object is copied, and an array gets the limit on its first field, which is the one the top-N selected.
func encodingWithLimit(value any, limit int) any {
	switch v := value.(type) {
	case string:
		return map[string]any{"field": v, "limit": limit}
	case map[string]any:
		res := maps.Clone(v)
		res["limit"] = limit
		return res
	case []any:
		if len(v) == 0 {
			return v
		}
		res := slices.Clone(v)
		res[0] = encodingWithLimit(res[0], limit)
		return res
	default:
		return value
	}
}

// flintTitle builds the title of a Flint chart from its headline, or from the bound fields' display names:
// "<measures> over Time" or "<measures> by <dimension>", plus "split by <series>" when a series dimension is bound.
func flintTitle(parsed *chartquery.FlintSpec, chartSpec map[string]any, fields map[string]*ChartField) string {
	if title, ok := chartSpec["title"].(string); ok && strings.TrimSpace(title) != "" {
		return strings.TrimSpace(title)
	}
	label := func(name string) string {
		if f, ok := fields[name]; ok {
			if f.Kind == "time" {
				return chartTimeDisplayName
			}
			if f.DisplayName != "" {
				return f.DisplayName
			}
		}
		return name
	}
	labels := func(names []string) string {
		res := make([]string, len(names))
		for i, n := range names {
			res[i] = label(n)
		}
		return strings.Join(res, ", ")
	}

	measures := parsed.Measures()
	// The category of the chart: the first dimension or time field on an axis, else on any channel other than a series or facet one;
	// without one, the detail dimension (the row grain) or the series dimension names it.
	var category, series, detail *chartquery.FlintEncoding
	for _, e := range parsed.Encodings {
		if e.Kind == chartquery.FieldKindMeasure {
			continue
		}
		switch e.Channel {
		case "color", "group", "shape", "strokeDash":
			if series == nil {
				series = e
			}
		case chartquery.DetailChannel:
			if detail == nil {
				detail = e
			}
		case "column", "row", "opacity":
		default:
			if category == nil {
				category = e
			}
		}
	}
	preposition := "by"
	switch {
	case category != nil && category.Kind == chartquery.FieldKindTime:
		preposition = "over"
	case category == nil && detail != nil:
		category, preposition = detail, "for"
	case category == nil && series != nil:
		category, series = series, nil
	}

	// Two measures on the axes (scatter plots) read as "x vs y"; otherwise the measures are listed.
	title := labels(measures)
	if x, y := parsed.FieldOn("x"), parsed.FieldOn("y"); x != nil && y != nil && x.Kind == chartquery.FieldKindMeasure && y.Kind == chartquery.FieldKindMeasure {
		title = label(x.Field) + " vs " + label(y.Field)
	}
	if category != nil {
		title = fmt.Sprintf("%s %s %s", title, preposition, label(category.Field))
	}
	if series != nil {
		title += " split by " + label(series.Field)
	}
	return title
}

func anyStrings(values []string) []any {
	res := make([]any, len(values))
	for i, v := range values {
		res[i] = v
	}
	return res
}

func ptr[T any](v T) *T {
	return &v
}
