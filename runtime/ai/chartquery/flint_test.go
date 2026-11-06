package chartquery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

// flintTestKinds classifies the fields of the metrics view the Flint tests bind.
var flintTestKinds = map[string]FieldKind{
	"__time":      FieldKindTime,
	"advertiser":  FieldKindDimension,
	"device":      FieldKindDimension,
	"country":     FieldKindDimension,
	"campaign_id": FieldKindDimension,
	"bids":        FieldKindMeasure,
	"clicks":      FieldKindMeasure,
	"spend":       FieldKindMeasure,
	"target":      FieldKindMeasure,
}

func flintTestData(extra map[string]any) map[string]any {
	data := map[string]any{
		"metrics_view": "mv",
		"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

// TestBuildFlintGolden checks the generated queries of Flint chart specs against golden files in testdata (flint_*.json).
// The main query is built with a fixed domain so the top-N filters are visible. Run with -update to regenerate.
func TestBuildFlintGolden(t *testing.T) {
	comparisonTimeRange := map[string]any{"start": "2023-12-01T00:00:00Z", "end": "2024-01-01T00:00:00Z"}
	where := map[string]any{"cond": map[string]any{"op": "eq", "exprs": []any{map[string]any{"name": "country"}, map[string]any{"val": "US"}}}}

	cases := []struct {
		name      string
		chartSpec map[string]any
		data      map[string]any
		domain    map[string][]any
	}{
		{
			name: "flint_bar_chart",
			chartSpec: map[string]any{
				"chartType": "Bar Chart",
				"encodings": map[string]any{
					"x": map[string]any{"field": "advertiser", "limit": 10, "sortBy": "y", "sortOrder": "descending"},
					"y": "bids",
				},
				"chartProperties": map[string]any{"cornerRadius": 2},
			},
			data:   flintTestData(map[string]any{"where": where}),
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name: "flint_bar_chart_sorted_by_name",
			chartSpec: map[string]any{
				"chartType": "Bar Chart",
				"encodings": map[string]any{
					"x": map[string]any{"field": "advertiser", "sortBy": "x", "sortOrder": "descending"},
					"y": "bids",
				},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"advertiser": {"B", "A"}},
		},
		{
			name: "flint_line_chart_comparison",
			chartSpec: map[string]any{
				"chartType": "Line Chart",
				"encodings": map[string]any{"x": "__time", "y": "bids"},
			},
			data: flintTestData(map[string]any{"comparison_time_range": comparisonTimeRange, "time_grain": "TIME_GRAIN_MONTH", "where": where}),
		},
		{
			name: "flint_rose_chart",
			chartSpec: map[string]any{
				"chartType": "Rose Chart",
				"encodings": map[string]any{"x": "advertiser", "y": "bids"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name: "flint_line_chart_temporal_dimension",
			chartSpec: map[string]any{
				"chartType": "Line Chart",
				"encodings": map[string]any{"x": map[string]any{"field": "country", "type": "temporal"}, "y": "bids", "color": "device"},
			},
			data:   flintTestData(map[string]any{"time_grain": "TIME_GRAIN_WEEK"}),
			domain: map[string][]any{"device": {"iOS"}},
		},
		{
			name: "flint_grouped_bar_chart",
			chartSpec: map[string]any{
				"chartType": "Grouped Bar Chart",
				"encodings": map[string]any{"x": "advertiser", "y": "bids", "group": "device"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"advertiser": {"A", "B"}, "device": {"iOS", "Android"}},
		},
		{
			name: "flint_stacked_bar_chart",
			chartSpec: map[string]any{
				"chartType":       "Stacked Bar Chart",
				"encodings":       map[string]any{"x": "__time", "y": "bids", "color": "device"},
				"chartProperties": map[string]any{"stackMode": "normalize"},
			},
			data:   flintTestData(map[string]any{"time_grain": "TIME_GRAIN_WEEK"}),
			domain: map[string][]any{"device": {"iOS"}},
		},
		{
			name: "flint_waterfall_chart",
			chartSpec: map[string]any{
				"chartType":       "Waterfall Chart",
				"encodings":       map[string]any{"x": "__time", "y": "spend"},
				"chartProperties": map[string]any{"totals": "both"},
			},
			data: flintTestData(map[string]any{"time_grain": "TIME_GRAIN_MONTH"}),
		},
		{
			name: "flint_pie_chart",
			chartSpec: map[string]any{
				"chartType":       "Pie Chart",
				"encodings":       map[string]any{"size": "bids", "color": map[string]any{"field": "device", "limit": 5}},
				"chartProperties": map[string]any{"innerRadius": 40},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"device": {"iOS", "Android"}},
		},
		{
			name: "flint_heatmap",
			chartSpec: map[string]any{
				"chartType": "Heatmap",
				"encodings": map[string]any{"x": "device", "y": map[string]any{"field": "advertiser", "limit": 100}, "color": "bids"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"device": {"iOS"}, "advertiser": {"A"}},
		},
		{
			name: "flint_slope_chart",
			chartSpec: map[string]any{
				"chartType": "Slope Chart",
				"encodings": map[string]any{"x": "__time", "y": "bids", "color": "advertiser"},
			},
			data:   flintTestData(map[string]any{"time_grain": "TIME_GRAIN_YEAR"}),
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name: "flint_scatter_plot_detail",
			chartSpec: map[string]any{
				"chartType": "Scatter Plot",
				"encodings": map[string]any{"x": "bids", "y": "clicks", "size": "spend", "detail": "campaign_id", "color": "device"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"device": {"iOS", "Android"}},
		},
		{
			name: "flint_bar_chart_facet",
			chartSpec: map[string]any{
				"chartType": "Bar Chart",
				"encodings": map[string]any{"x": "advertiser", "y": "bids", "column": "country"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"advertiser": {"A", "B"}, "country": {"US", "CA"}},
		},
		{
			name: "flint_multi_measure_fold",
			chartSpec: map[string]any{
				"chartType": "Line Chart",
				"encodings": map[string]any{"x": "__time", "y": []any{"bids", "clicks"}},
			},
			data: flintTestData(map[string]any{"time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z", "time_zone": "America/New_York"}}),
		},
		{
			name: "flint_bullet_chart",
			chartSpec: map[string]any{
				"chartType": "Bullet Chart",
				"encodings": map[string]any{"y": "advertiser", "x": "bids", "goal": "target"},
			},
			data:   flintTestData(nil),
			domain: map[string][]any{"advertiser": {"A"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := ParseFlintSpec(tc.chartSpec, tc.data, flintTestKinds)
			require.NoError(t, err)
			plan, err := BuildFlint(spec)
			require.NoError(t, err)

			got := struct {
				ChartType string                        `json:"chart_type"`
				TimeGrain string                        `json:"time_grain,omitempty"`
				TimeField string                        `json:"time_field,omitempty"`
				Fields    []string                      `json:"fields"`
				Measures  []string                      `json:"measures"`
				RowLimit  int                           `json:"row_limit"`
				TopN      []TopNQuery                   `json:"top_n,omitempty"`
				Main      *metricsview.Query            `json:"main"`
				Extra     map[string]*metricsview.Query `json:"extra,omitempty"`
				Other     *metricsview.Query            `json:"other,omitempty"`
			}{
				ChartType: spec.ChartType,
				TimeGrain: spec.TimeGrainName(),
				TimeField: plan.TimeField,
				Fields:    spec.Fields(),
				Measures:  spec.Measures(),
				RowLimit:  plan.RowLimit,
				TopN:      plan.TopN,
				Main:      plan.Main(tc.domain),
				Extra:     plan.Extra,
			}
			if plan.Other != nil {
				got.Other = plan.Other.Query(tc.domain)
			}
			gotJSON, err := json.MarshalIndent(got, "", "  ")
			require.NoError(t, err)
			gotJSON = append(gotJSON, '\n')

			path := filepath.Join("testdata", tc.name+".json")
			if *update {
				require.NoError(t, os.WriteFile(path, gotJSON, 0o644))
			}
			want, err := os.ReadFile(path)
			require.NoError(t, err, "missing golden file; run with -update")
			require.JSONEq(t, string(want), string(gotJSON))
		})
	}
}

func TestParseFlintSpecErrors(t *testing.T) {
	data := flintTestData(nil)

	_, err := ParseFlintSpec(map[string]any{"chartType": "Sankey", "encodings": map[string]any{"x": "advertiser"}}, data, flintTestKinds)
	require.ErrorContains(t, err, `unknown chart type "Sankey"`)
	require.ErrorContains(t, err, "Bar Chart")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "y": "bids", "angle": "clicks"}}, data, flintTestKinds)
	require.ErrorContains(t, err, `does not accept channel "angle"`)
	require.ErrorContains(t, err, "x, y, color")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "publisher", "y": "bids"}}, data, flintTestKinds)
	require.ErrorContains(t, err, `field "publisher" on channel "x" not found in metrics view "mv"`)

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": map[string]any{"field": "advertiser", "sort": "-y"}, "y": "bids"}}, data, flintTestKinds)
	require.ErrorContains(t, err, `unknown encoding key "sort"`)

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": map[string]any{"field": "advertiser", "sortOrder": "desc"}, "y": "bids"}}, data, flintTestKinds)
	require.ErrorContains(t, err, "invalid sortOrder")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart"}, data, flintTestKinds)
	require.ErrorContains(t, err, "encodings")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser"}}, map[string]any{"time_range": data["time_range"]}, flintTestKinds)
	require.ErrorContains(t, err, "metrics_view")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser"}}, map[string]any{"metrics_view": "mv"}, flintTestKinds)
	require.ErrorContains(t, err, "time_range")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Line Chart", "encodings": map[string]any{"x": "__time", "y": "bids"}}, flintTestData(map[string]any{"time_grain": "TIME_GRAIN_FORTNIGHT"}), flintTestKinds)
	require.ErrorContains(t, err, "invalid time_grain")

	// A grain without a time field is ignored rather than reported.
	spec, err := ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "y": "bids"}}, flintTestData(map[string]any{"time_grain": "TIME_GRAIN_MONTH"}), flintTestKinds)
	require.NoError(t, err)
	require.Empty(t, spec.TimeGrainName())

	// Build rejects a chart that binds no measure.
	spec, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "color": "device"}}, data, flintTestKinds)
	require.NoError(t, err)
	_, err = BuildFlint(spec)
	require.ErrorContains(t, err, "at least one measure")

	_, err = ParseFlintSpec(map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": map[string]any{"field": "advertiser", "type": "categorical"}, "y": "bids"}}, data, flintTestKinds)
	require.ErrorContains(t, err, `invalid type "categorical"`)

	// The data block: a filter needs a condition, and a comparison both ends.
	bar := map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser", "y": "bids"}}
	for _, where := range []any{map[string]any{}, map[string]any{"name": "country"}, "country = 'US'"} {
		_, err = ParseFlintSpec(bar, flintTestData(map[string]any{"where": where}), flintTestKinds)
		require.ErrorContains(t, err, "data.where must be an object with a 'cond' property", "where %v", where)
	}
	for _, comparison := range []any{map[string]any{"start": "2023-12-01T00:00:00Z"}, map[string]any{"end": "2024-01-01T00:00:00Z"}, map[string]any{}} {
		_, err = ParseFlintSpec(bar, flintTestData(map[string]any{"comparison_time_range": comparison}), flintTestKinds)
		require.ErrorContains(t, err, "data.comparison_time_range must contain both 'start' and 'end'", "comparison %v", comparison)
	}
}

// TestParseFlintSpecDrawable covers the bindings a chart cannot draw, which are rejected with how to restructure the chart.
func TestParseFlintSpecDrawable(t *testing.T) {
	parse := func(chartType string, encodings map[string]any, extra map[string]any) error {
		_, err := ParseFlintSpec(map[string]any{"chartType": chartType, "encodings": encodings}, flintTestData(extra), flintTestKinds)
		return err
	}

	// sortBy names a bound channel, a bound field, or a list of values for Flint; anything else is an error rather than ignored.
	require.NoError(t, parse("Bar Chart", map[string]any{"x": map[string]any{"field": "advertiser", "sortBy": "y"}, "y": "bids"}, nil))
	require.NoError(t, parse("Bar Chart", map[string]any{"x": map[string]any{"field": "advertiser", "sortBy": "bids"}, "y": "bids"}, nil))
	require.NoError(t, parse("Bar Chart", map[string]any{"x": map[string]any{"field": "advertiser", "sortBy": `["B", "A"]`}, "y": "bids"}, nil))
	require.ErrorContains(t, parse("Bar Chart", map[string]any{"x": map[string]any{"field": "advertiser", "sortBy": "clicks"}, "y": "bids"}, nil), `channel "x" sorts by "clicks", which is neither a bound channel nor a bound field`)
	require.ErrorContains(t, parse("Bar Chart", map[string]any{"x": map[string]any{"field": "advertiser", "sortBy": "color"}, "y": "bids"}, nil), `sorts by "color"`)

	// Charts that draw one mark per value of x need the time field or a dimension there; horizontal layouts keep the measure on x.
	require.ErrorContains(t, parse("Line Chart", map[string]any{"x": "spend", "y": "bids"}, nil), `Line Chart needs the time field or a dimension on x, not the measure "spend"`)
	require.ErrorContains(t, parse("Waterfall Chart", map[string]any{"y": "bids"}, nil), "Waterfall Chart needs the time field or a dimension on x")
	for _, chartType := range []string{"Area Chart", "Streamgraph", "Sparkline", "Bump Chart", "Slope Chart"} {
		require.ErrorContains(t, parse(chartType, map[string]any{"x": "spend", "y": "bids"}, nil), "needs the time field or a dimension on x", chartType)
	}
	require.ErrorContains(t, parse("Range Area Chart", map[string]any{"x": "spend", "y": "bids", "y2": "clicks"}, nil), "needs the time field or a dimension on x")
	require.ErrorContains(t, parse("Candlestick Chart", map[string]any{"x": "spend", "open": "bids", "high": "bids", "low": "bids", "close": "bids"}, nil), "needs the time field or a dimension on x")
	require.NoError(t, parse("Line Chart", map[string]any{"x": "advertiser", "y": "bids"}, nil))
	require.NoError(t, parse("Bar Chart", map[string]any{"x": "bids", "y": "advertiser"}, nil))
	require.NoError(t, parse("Lollipop Chart", map[string]any{"x": "bids", "y": "advertiser"}, nil))
	require.NoError(t, parse("Bullet Chart", map[string]any{"x": "bids", "y": "advertiser", "goal": "target"}, nil))
	require.NoError(t, parse("Bar Table", map[string]any{"x": "bids", "y": "advertiser"}, nil))

	// A comparison is drawn as a second series: one measure on y (or x), the series channel free, on a chart type that folds it.
	comparison := map[string]any{"comparison_time_range": map[string]any{"start": "2023-12-01T00:00:00Z", "end": "2024-01-01T00:00:00Z"}}
	require.NoError(t, parse("Line Chart", map[string]any{"x": "__time", "y": "bids"}, comparison))
	require.NoError(t, parse("Bar Chart", map[string]any{"x": "advertiser", "y": "bids"}, comparison))
	require.NoError(t, parse("Bar Chart", map[string]any{"y": "advertiser", "x": "bids"}, comparison), "horizontal bars fold the measure on x")
	require.NoError(t, parse("Grouped Bar Chart", map[string]any{"x": "advertiser", "y": "bids"}, comparison))
	require.ErrorContains(t, parse("Line Chart", map[string]any{"x": "__time", "y": "bids", "color": "device"}, comparison), `cannot also bind "device" on color`)
	require.ErrorContains(t, parse("Bar Chart", map[string]any{"x": "advertiser", "y": "bids", "color": "device"}, comparison), `second series on group, so the chart cannot also bind "device" on color`)
	require.ErrorContains(t, parse("Grouped Bar Chart", map[string]any{"x": "advertiser", "y": "bids", "group": "device"}, comparison), `cannot also bind "device" on group`)
	require.ErrorContains(t, parse("Line Chart", map[string]any{"x": "__time", "y": []any{"bids", "clicks"}}, comparison), "needs exactly one measure on y")
	require.ErrorContains(t, parse("Line Chart", map[string]any{"x": "__time", "y": []any{"bids"}}, comparison), "needs exactly one measure on y", "the widget does not fold an array")
	err := parse("Pie Chart", map[string]any{"size": "bids", "color": "device"}, comparison)
	require.ErrorContains(t, err, "a Pie Chart cannot draw data.comparison_time_range")
	require.ErrorContains(t, err, "Area Chart, Bar Chart, Grouped Bar Chart, Line Chart, Lollipop Chart, Slope Chart")
}

// TestFlintTimeFields checks that a dimension declared temporal is a time field on all of its channels:
// it is floored at the grain and gets no top-N.
func TestFlintTimeFields(t *testing.T) {
	spec, err := ParseFlintSpec(map[string]any{
		"chartType": "Bar Chart",
		"encodings": map[string]any{"x": map[string]any{"field": "country", "type": "temporal"}, "y": "bids", "column": "country"},
	}, flintTestData(nil), flintTestKinds)
	require.NoError(t, err)
	require.Equal(t, "country", spec.TimeField())
	require.Equal(t, "TIME_GRAIN_DAY", spec.TimeGrainName())
	require.Empty(t, spec.Dimensions())
	plan, err := BuildFlint(spec)
	require.NoError(t, err)
	require.Empty(t, plan.TopN)
	require.Equal(t, "country", plan.TimeField)
}

// TestBuildFlintLimits checks the default top-N limits per chart type and channel: the angular axis of a polar chart defaults to 10.
func TestBuildFlintLimits(t *testing.T) {
	limit := func(chartType string, x any) int {
		spec, err := ParseFlintSpec(map[string]any{"chartType": chartType, "encodings": map[string]any{"x": x, "y": "bids"}}, flintTestData(nil), flintTestKinds)
		require.NoError(t, err)
		plan, err := BuildFlint(spec)
		require.NoError(t, err)
		require.Len(t, plan.TopN, 1)
		return plan.TopN[0].Limit
	}
	require.Equal(t, defaultNominalLimit, limit("Bar Chart", "advertiser"))
	require.Equal(t, defaultPolarLimit, limit("Rose Chart", "advertiser"))
	require.Equal(t, defaultPolarLimit, limit("Radar Chart", "advertiser"))
	require.Equal(t, 30, limit("Rose Chart", map[string]any{"field": "advertiser", "limit": 30}), "an explicit limit is kept")
}

func TestFlintSpecFields(t *testing.T) {
	spec, err := ParseFlintSpec(map[string]any{
		"chartType": "Line Chart",
		"encodings": map[string]any{"color": "device", "y": []any{"bids", "clicks", "bids"}, "x": "__time", "detail": "campaign_id"},
	}, flintTestData(nil), flintTestKinds)
	require.NoError(t, err)
	// Channel order of the chart type (x, y, color, ..., detail), not the input's key order
	require.Equal(t, []string{"__time", "bids", "clicks", "device", "campaign_id"}, spec.Fields())
	require.Equal(t, []string{"bids", "clicks"}, spec.Measures())
	require.Equal(t, "__time", spec.TimeField())
	require.Equal(t, "TIME_GRAIN_DAY", spec.TimeGrainName())

	require.Equal(t, "0.4.1", Catalog().FlintVersion)
	require.NotNil(t, Catalog().ChartType("Waterfall Chart"))
	require.Nil(t, Catalog().ChartType("Map"), "geographic charts are left out of the catalog")
	require.True(t, Catalog().ChartType("Bar Chart").AcceptsChannel(DetailChannel))
}
