package ai

import (
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/ai/chartquery"
	"github.com/stretchr/testify/require"
)

// TestFlintChartNotesMatchCatalog keeps the hand-written chart notes of the external description tied to the generated catalog:
// every noted chart type exists, and every chart type is described.
func TestFlintChartNotesMatchCatalog(t *testing.T) {
	catalog := chartquery.Catalog()
	for name := range flintChartNotes {
		require.NotNil(t, catalog.ChartType(name), "note for a chart type that is not in the catalog: %q", name)
	}
	description := flintCreateChartDescription()
	for _, ct := range catalog.ChartTypes {
		require.Contains(t, description, "- "+ct.Name+": ", "chart type missing from the description")
	}
	require.NotContains(t, description, "KPI Card")
	require.Less(t, len(description), 12_000, "the description is sent with every tool list")

	schema := flintChartsInputSchema()
	require.Equal(t, []string{"chart_spec", "data"}, schema.Required)
	require.NotContains(t, schema.Properties, "chart_type", "external clients use the Flint shape only")
	require.NotContains(t, schema.Properties, "spec")
	require.Len(t, schema.Properties["chart_spec"].Properties["chartType"].Enum, len(catalog.ChartTypes))
	require.Contains(t, schema.Defs, "FlintEncoding")
	require.Contains(t, schema.Defs, "Expression")
	require.Contains(t, schema.Defs, "TimeGrain")
	require.NotContains(t, schema.Defs, "ChartSpec", "the in-app chart spec is not part of the schema")
	require.NotContains(t, schema.Defs, "FieldConfig")
	resolved, err := schema.Resolve(nil)
	require.NoError(t, err, "every reference resolves without the in-app definitions")
	require.Error(t, resolved.Validate(map[string]any{"chart_type": "bar_chart", "spec": map[string]any{}}), "the in-app shape does not validate")
}

func TestFlintInputWithLimits(t *testing.T) {
	input := &ChartInput{
		ChartSpec: map[string]any{
			"chartType": "Bar Chart",
			"encodings": map[string]any{
				"x":     "advertiser",
				"y":     []any{"bids", "clicks"},
				"color": []any{map[string]any{"field": "device", "sortBy": "y"}, "country"},
			},
		},
		Data:          map[string]any{"metrics_view": "mv"},
		SemanticTypes: map[string]any{"bids": "Amount"},
	}
	res := flintInputWithLimits(input, map[string]map[string]int{"x": {"advertiser": 20}, "color": {"device": 5, "country": 3}}, "TIME_GRAIN_DAY")

	encodings := res.ChartSpec["encodings"].(map[string]any)
	require.Equal(t, map[string]any{"field": "advertiser", "limit": 20}, encodings["x"], "a field name becomes an object")
	require.Equal(t, []any{map[string]any{"field": "device", "sortBy": "y", "limit": 5}, map[string]any{"field": "country", "limit": 3}}, encodings["color"], "each field of a channel gets its own limit")
	require.Equal(t, []any{"bids", "clicks"}, encodings["y"], "channels without a limit are echoed as given")
	require.Equal(t, "TIME_GRAIN_DAY", res.Data["time_grain"])
	require.Equal(t, map[string]any{"bids": "Amount"}, res.SemanticTypes, "the caller's semantic types are echoed")

	// The input maps are not modified.
	require.Equal(t, []any{map[string]any{"field": "device", "sortBy": "y"}, "country"}, input.ChartSpec["encodings"].(map[string]any)["color"])
	require.NotContains(t, input.Data, "time_grain")
}

// TestFieldKinds checks that dimensions holding dates or timestamps are time fields, whatever the metrics view types them as.
func TestFieldKinds(t *testing.T) {
	kinds := fieldKinds(&runtimev1.MetricsViewSpec{
		TimeDimension: "event_time",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "event_time", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME},
			{Name: "order_time", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
			// A date column other than the time dimension is typed categorical by the metrics view.
			{Name: "order_date", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL, DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}},
			{Name: "country", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL, DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{{Name: "bids"}},
	})
	require.Equal(t, map[string]chartquery.FieldKind{
		"event_time": chartquery.FieldKindTime,
		"order_time": chartquery.FieldKindTime,
		"order_date": chartquery.FieldKindTime,
		"country":    chartquery.FieldKindDimension,
		"bids":       chartquery.FieldKindMeasure,
	}, kinds)
}

func TestFlintTitle(t *testing.T) {
	kinds := map[string]chartquery.FieldKind{"__time": chartquery.FieldKindTime, "advertiser": chartquery.FieldKindDimension, "device": chartquery.FieldKindDimension, "bids": chartquery.FieldKindMeasure, "clicks": chartquery.FieldKindMeasure}
	fields := map[string]*ChartField{
		"__time":     {Name: "__time", Kind: "time", DisplayName: "__time"},
		"advertiser": {Name: "advertiser", Kind: "dimension", DisplayName: "Advertiser"},
		"device":     {Name: "device", Kind: "dimension", DisplayName: "Device"},
		"bids":       {Name: "bids", Kind: "measure", DisplayName: "Bids"},
		"clicks":     {Name: "clicks", Kind: "measure", DisplayName: "Clicks"},
	}
	data := map[string]any{"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}}
	title := func(chartSpec map[string]any) string {
		parsed, err := chartquery.ParseFlintSpec(chartSpec, data, kinds)
		require.NoError(t, err)
		return flintTitle(parsed, chartSpec, fields)
	}

	require.Equal(t, "Bids by Advertiser", title(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "y": "bids"}}))
	require.Equal(t, "Bids, Clicks over Time split by Device", title(map[string]any{"chartType": "Line Chart", "encodings": map[string]any{"x": "__time", "y": []any{"bids", "clicks"}, "color": "device"}}))
	require.Equal(t, "Bids by Device", title(map[string]any{"chartType": "Pie Chart", "encodings": map[string]any{"size": "bids", "color": "device"}}), "a series channel is the category when nothing else is")
	require.Equal(t, "Bids by Advertiser", title(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"y": "advertiser", "x": "bids"}}), "horizontal layout")
	require.Equal(t, "Bids by Device", title(map[string]any{"chartType": "Heatmap", "encodings": map[string]any{"x": "device", "y": "advertiser", "color": "bids"}}))
	require.Equal(t, "Bids vs Clicks for Advertiser", title(map[string]any{"chartType": "Scatter Plot", "encodings": map[string]any{"x": "bids", "y": "clicks", "detail": "advertiser"}}), "the detail dimension is the grain of a scatter plot")
	require.Equal(t, "My headline", title(map[string]any{"chartType": "Bar Chart", "title": " My headline ", "encodings": map[string]any{"x": "advertiser", "y": "bids"}}))
	require.True(t, strings.HasPrefix(title(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "y": "bids", "column": "device"}}), "Bids by Advertiser"), "facets do not name the chart")
}
