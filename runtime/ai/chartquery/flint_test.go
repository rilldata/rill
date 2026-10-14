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
				"encodings": map[string]any{
					"x":     "__time",
					"y":     "bids",
					"color": map[string]any{"field": "device", "limit": 3},
				},
			},
			data:   flintTestData(map[string]any{"comparison_time_range": comparisonTimeRange, "time_grain": "TIME_GRAIN_MONTH"}),
			domain: map[string][]any{"device": {"iOS", "Android"}},
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
				ChartType string             `json:"chart_type"`
				TimeGrain string             `json:"time_grain,omitempty"`
				TimeField string             `json:"time_field,omitempty"`
				Fields    []string           `json:"fields"`
				Measures  []string           `json:"measures"`
				RowLimit  int                `json:"row_limit"`
				TopN      []TopNQuery        `json:"top_n,omitempty"`
				Main      *metricsview.Query `json:"main"`
			}{
				ChartType: spec.ChartType,
				TimeGrain: spec.TimeGrainName(),
				TimeField: plan.TimeField,
				Fields:    spec.Fields(),
				Measures:  spec.Measures(),
				RowLimit:  plan.RowLimit,
				TopN:      plan.TopN,
				Main:      plan.Main(tc.domain),
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
