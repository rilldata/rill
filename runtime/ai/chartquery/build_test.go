package chartquery

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update the golden files")

// TestBuildGolden checks the generated queries for every chart type against golden files in testdata.
// The main query is built with a fixed domain so the top-N filters are visible. Run with -update to regenerate.
func TestBuildGolden(t *testing.T) {
	timeRange := map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}
	comparisonTimeRange := map[string]any{"start": "2023-12-01T00:00:00Z", "end": "2024-01-01T00:00:00Z"}
	where := map[string]any{"cond": map[string]any{"op": "eq", "exprs": []any{map[string]any{"name": "country"}, map[string]any{"val": "US"}}}}

	cases := []struct {
		name      string
		chartType string
		spec      map[string]any
		domain    map[string][]any
	}{
		{
			name:      "bar_chart",
			chartType: BarChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "where": where, "color": "primary",
				"x": map[string]any{"field": "advertiser", "type": "nominal", "limit": 10, "sort": "-y"},
				"y": map[string]any{"field": "bids", "type": "quantitative", "zeroBasedOrigin": true},
			},
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name:      "bar_chart_horizontal",
			chartType: BarChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "color": "primary",
				"x": map[string]any{"field": "bids", "type": "quantitative"},
				"y": map[string]any{"field": "advertiser", "type": "nominal", "limit": 100, "sort": "-x"},
			},
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name:      "bar_chart_delta_sort",
			chartType: BarChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "comparison_time_range": comparisonTimeRange, "color": "primary",
				"x": map[string]any{"field": "advertiser", "type": "nominal", "limit": 5, "sort": "-y_delta"},
				"y": map[string]any{"field": "bids", "type": "quantitative"},
			},
			domain: map[string][]any{"advertiser": {"A", "B"}},
		},
		{
			name:      "bar_chart_show_null",
			chartType: BarChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "color": "primary",
				"x": map[string]any{"field": "advertiser", "type": "nominal", "showNull": true, "sort": "x"},
				"y": map[string]any{"field": "bids", "type": "quantitative"},
			},
			domain: map[string][]any{"advertiser": {"A", nil}},
		},
		{
			name:      "line_chart",
			chartType: LineChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_MONTH",
				"color": map[string]any{"field": "device", "type": "nominal", "limit": 3},
				"x":     map[string]any{"field": "__time", "type": "temporal", "limit": 20, "sort": "-y"},
				"y":     map[string]any{"field": "bids", "type": "quantitative"},
			},
			domain: map[string][]any{"device": {"iOS", "Android"}},
		},
		{
			name:      "line_chart_comparison",
			chartType: LineChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "comparison_time_range": comparisonTimeRange, "color": "primary",
				"x": map[string]any{"field": "__time", "type": "temporal"},
				"y": map[string]any{"field": "bids", "type": "quantitative"},
			},
		},
		{
			name:      "area_chart",
			chartType: AreaChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z", "time_zone": "America/New_York"},
				"color": map[string]any{"field": "app_or_site", "type": "nominal"},
				"x":     map[string]any{"field": "__time", "type": "temporal", "showNull": true},
				"y":     map[string]any{"field": "requests", "type": "quantitative"},
			},
			domain: map[string][]any{"app_or_site": {"app"}},
		},
		{
			name:      "stacked_bar_multi_measure",
			chartType: StackedBar,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_WEEK",
				"color": map[string]any{"field": "rill_measures", "type": "value"},
				"x":     map[string]any{"field": "__time", "type": "temporal"},
				"y":     map[string]any{"field": "clicks", "type": "quantitative", "fields": []any{"clicks", "impressions"}},
			},
		},
		{
			name:      "stacked_bar_normalized",
			chartType: StackedBarNormalized,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"color": map[string]any{"field": "username", "type": "nominal", "limit": 3},
				"x":     map[string]any{"field": "date", "type": "temporal"},
				"y":     map[string]any{"field": "commits", "type": "quantitative"},
			},
			domain: map[string][]any{"username": {"alice", "bob"}},
		},
		{
			name:      "donut_chart",
			chartType: DonutChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"color":   map[string]any{"field": "username", "type": "nominal", "limit": 20},
				"measure": map[string]any{"field": "commits", "type": "quantitative", "showTotal": true},
			},
			domain: map[string][]any{"username": {"alice", "bob"}},
		},
		{
			name:      "pie_chart_custom_order",
			chartType: PieChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"color":   map[string]any{"field": "weekday", "type": "nominal", "sort": []any{"Mon", "Tue", "Wed"}},
				"measure": map[string]any{"field": "commits", "type": "quantitative"},
			},
			domain: map[string][]any{"weekday": {"Mon", "Tue", "Wed"}},
		},
		{
			name:      "funnel_chart_dimension",
			chartType: FunnelChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "breakdownMode": "dimension", "color": "stage", "mode": "width",
				"measure": map[string]any{"field": "users", "type": "quantitative"},
				"stage":   map[string]any{"field": "stage", "type": "nominal", "limit": 15},
			},
			domain: map[string][]any{"stage": {"visit", "signup"}},
		},
		{
			name:      "funnel_chart_measures",
			chartType: FunnelChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "breakdownMode": "measures", "color": "name",
				"measure": map[string]any{"field": "impressions", "type": "quantitative", "fields": []any{"impressions", "video_starts", "video_completes"}},
			},
		},
		{
			name:      "heatmap",
			chartType: Heatmap,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"color": map[string]any{"field": "bids", "type": "quantitative"},
				"x":     map[string]any{"field": "day", "type": "nominal", "limit": 10, "sort": []any{"Mon", "Tue"}},
				"y":     map[string]any{"field": "hour", "type": "nominal", "limit": 24, "sort": "-color"},
			},
			domain: map[string][]any{"day": {"Mon", "Tue"}, "hour": {"0", "1"}},
		},
		{
			name:      "heatmap_temporal",
			chartType: Heatmap,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_DAY",
				"color": map[string]any{"field": "bids", "type": "quantitative"},
				"x":     map[string]any{"field": "__time", "type": "temporal"},
				"y":     map[string]any{"field": "device", "type": "nominal", "limit": 100},
			},
			domain: map[string][]any{"device": {"iOS"}},
		},
		{
			name:      "combo_chart_temporal",
			chartType: ComboChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_DAY",
				"color": map[string]any{"field": "rill_measures", "type": "value"},
				"x":     map[string]any{"field": "__time", "type": "temporal"},
				"y1":    map[string]any{"field": "qps", "type": "quantitative", "mark": "bar"},
				"y2":    map[string]any{"field": "requests", "type": "quantitative", "mark": "line"},
			},
		},
		{
			name:      "combo_chart_nominal",
			chartType: ComboChart,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"x":  map[string]any{"field": "country", "type": "nominal", "limit": 5},
				"y1": map[string]any{"field": "qps", "type": "quantitative", "mark": "bar"},
				"y2": map[string]any{"field": "requests", "type": "quantitative", "mark": "line"},
			},
			domain: map[string][]any{"country": {"US", "CA"}},
		},
		{
			name:      "scatter_plot",
			chartType: ScatterPlot,
			spec: map[string]any{
				"metrics_view": "mv", "time_range": timeRange,
				"x":         map[string]any{"field": "bids", "type": "quantitative"},
				"y":         map[string]any{"field": "clicks", "type": "quantitative"},
				"dimension": map[string]any{"field": "advertiser", "type": "nominal", "limit": 10},
				"size":      map[string]any{"field": "impressions", "type": "quantitative"},
				"color":     map[string]any{"field": "device", "type": "nominal", "limit": 5},
			},
			domain: map[string][]any{"device": {"iOS", "Android"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := ParseSpec(tc.chartType, tc.spec)
			require.NoError(t, err)
			plan, err := Build(spec)
			require.NoError(t, err)

			got := struct {
				TimeGrain  string                        `json:"time_grain,omitempty"`
				Horizontal bool                          `json:"horizontal,omitempty"`
				RowLimit   int                           `json:"row_limit"`
				TopN       []TopNQuery                   `json:"top_n,omitempty"`
				Main       *metricsview.Query            `json:"main"`
				Extra      map[string]*metricsview.Query `json:"extra,omitempty"`
				Measures   []string                      `json:"measure_domain,omitempty"`
			}{
				TimeGrain:  spec.TimeGrainName(),
				Horizontal: spec.Horizontal,
				RowLimit:   plan.RowLimit,
				TopN:       plan.TopN,
				Main:       plan.Main(tc.domain),
				Extra:      plan.Extra,
				Measures:   plan.MeasureDomain,
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

func TestParseSpecErrors(t *testing.T) {
	timeRange := map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}

	_, err := ParseSpec("sparkline", map[string]any{"metrics_view": "mv", "time_range": timeRange})
	require.ErrorContains(t, err, "unsupported chart type")

	_, err = ParseSpec(BarChart, map[string]any{"time_range": timeRange})
	require.ErrorContains(t, err, "metrics_view")

	_, err = ParseSpec(BarChart, map[string]any{"metrics_view": "mv"})
	require.ErrorContains(t, err, "time_range")

	_, err = ParseSpec(LineChart, map[string]any{
		"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_FORTNIGHT",
		"x": map[string]any{"field": "__time", "type": "temporal"},
		"y": map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.ErrorContains(t, err, "invalid time_grain")

	// A grain without a temporal field is ignored rather than reported.
	spec, err := ParseSpec(BarChart, map[string]any{
		"metrics_view": "mv", "time_range": timeRange, "time_grain": "TIME_GRAIN_MONTH",
		"x": map[string]any{"field": "advertiser", "type": "nominal"},
		"y": map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.NoError(t, err)
	require.Empty(t, spec.TimeGrainName())

	// Build rejects specs that lack the fields the chart type needs.
	spec, err = ParseSpec(DonutChart, map[string]any{"metrics_view": "mv", "time_range": timeRange})
	require.NoError(t, err)
	_, err = Build(spec)
	require.ErrorContains(t, err, "measure")
}

func TestSpecFields(t *testing.T) {
	spec, err := ParseSpec(StackedBar, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
		"color": map[string]any{"field": "rill_measures", "type": "value"},
		"x":     map[string]any{"field": "__time", "type": "temporal"},
		"y":     map[string]any{"field": "clicks", "type": "quantitative", "fields": []any{"clicks", "impressions"}},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"__time", "clicks", "impressions"}, spec.Fields())
	require.Equal(t, "TIME_GRAIN_DAY", spec.TimeGrainName())
}
