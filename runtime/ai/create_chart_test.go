package ai_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// createChartTestFiles is a project with a time series metrics view and a default theme.
func createChartTestFiles(rillYAML string) map[string]string {
	return map[string]string{
		"rill.yaml": rillYAML,
		"bids.sql": `
SELECT * FROM (VALUES
	('Acme', 'iOS', TIMESTAMP '2024-01-01 10:00:00', 10),
	('Acme', 'Android', TIMESTAMP '2024-01-02 10:00:00', 5),
	('Beta', 'iOS', TIMESTAMP '2024-01-02 11:00:00', 7),
	('Gamma', 'iOS', TIMESTAMP '2024-01-03 09:00:00', 1),
	(NULL, 'iOS', TIMESTAMP '2024-01-03 12:00:00', 100)
) AS t(advertiser_name, device_os, event_time, bids)
`,
		"bids_metrics.yaml": `
type: metrics_view
display_name: Bids
model: bids
timeseries: event_time
smallest_time_grain: hour
dimensions:
- column: advertiser_name
- column: device_os
measures:
- name: total_bids
  expression: SUM(bids)
  format_preset: humanize
- name: bid_count
  expression: COUNT(*)
- name: avg_bids
  expression: AVG(bids)
explore:
  skip: true
`,
		"themes/brand.yaml": `
type: theme
colors:
  primary: "#4f46e5"
  secondary: "#f59e0b"
`,
	}
}

func newChartSession(t *testing.T, rt *runtime.Runtime, instanceID, userAgent string) *ai.Session {
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     claims,
		UserAgent:  userAgent,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, s.Flush(t.Context()))
	})
	return s
}

// barChartArgs are create_chart arguments in the in-app shape Rill's own chat uses (see CreateChart.Spec).
func barChartArgs() ai.CreateChartArgs {
	return ai.CreateChartArgs{
		"chart_type": "bar_chart",
		"spec": map[string]any{
			"metrics_view": "bids_metrics",
			"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-08T00:00:00Z"},
			"color":        "primary",
			"x":            map[string]any{"field": "advertiser_name", "type": "nominal", "limit": 2, "sort": "-y"},
			"y":            map[string]any{"field": "total_bids", "type": "quantitative"},
		},
	}
}

// TestCreateChartInputLimits checks that an external result's input carries the effective limit of every channel that got a top-N:
// the default where the chart spec set none, on an axis and on a series channel, and on the y channel of a horizontal bar chart, which holds the dimension.
func TestCreateChartInputLimits(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	var stacked *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &stacked, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Stacked Bar Chart", "encodings": map[string]any{
			"x":     map[string]any{"field": "advertiser_name", "sortBy": "y"},
			"y":     "total_bids",
			"color": "device_os",
		}},
		"data": flintBarChartArgs()["data"],
	})
	require.NoError(t, err)
	encodings := stacked.Input.ChartSpec["encodings"].(map[string]any)
	require.Equal(t, map[string]any{"field": "advertiser_name", "sortBy": "y", "limit": float64(20)}, encodings["x"], "the rest of the binding is kept")
	require.Equal(t, map[string]any{"field": "device_os", "limit": float64(10)}, encodings["color"])
	require.Equal(t, "total_bids", encodings["y"])

	var bar *ai.CreateChartResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &bar, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "total_bids", "y": "advertiser_name"}},
		"data":       flintBarChartArgs()["data"],
	})
	require.NoError(t, err)
	encodings = bar.Input.ChartSpec["encodings"].(map[string]any)
	require.Equal(t, map[string]any{"field": "advertiser_name", "limit": float64(20)}, encodings["y"])
	require.Equal(t, "total_bids", encodings["x"])
	require.Equal(t, []any{"Acme", "Beta", "Gamma"}, bar.DomainValues["advertiser_name"])
}

// TestCreateChartTemporalComparison checks time flooring, comparison values and the time grain in the result,
// for a project without a theme or a frontend URL.
func TestCreateChartTemporalComparison(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "chatgpt/1.0")

	var res *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Line Chart", "encodings": map[string]any{"x": "event_time", "y": "total_bids"}},
		"data": map[string]any{
			"metrics_view":          "bids_metrics",
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-04T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2023-12-31T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
			"time_grain":            "TIME_GRAIN_DAY",
		},
	})
	require.NoError(t, err)

	require.Equal(t, "Total Bids over Time", res.Title)
	require.Equal(t, "TIME_GRAIN_DAY", res.TimeGrain)
	require.Len(t, res.Data, 2)
	require.Equal(t, "2024-01-02T00:00:00Z", res.Data[0]["event_time"], "temporal values are RFC3339 UTC bucket starts")
	require.EqualValues(t, 12, res.Data[0]["total_bids"])
	require.Contains(t, res.Data[0], "total_bids_prev")
	require.Equal(t, "time", res.Fields["event_time"].Kind)
	require.Equal(t, "Event Time", res.Fields["event_time"].DisplayName, "the metrics view's display name is kept for the time dimension")
	require.NotNil(t, res.ResolvedComparisonTimeRange)
	require.Empty(t, res.OpenURL, "no frontend URL configured")
	require.Nil(t, res.Theme, "no theme configured")
	require.Empty(t, res.DomainValues)
	require.False(t, res.Truncated)
	require.False(t, res.RowsTruncated)
	require.Equal(t, "event_time", res.Input.ChartSpec["encodings"].(map[string]any)["x"], "a time axis gets no top-N")
}

// TestCreateChartDonutTotals checks the total query of a donut chart whose slices were not cut.
func TestCreateChartDonutTotals(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "cursor/1.0")

	var res *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Donut Chart", "encodings": map[string]any{"size": "total_bids", "color": map[string]any{"field": "device_os", "limit": 10}}},
		"data":       flintBarChartArgs()["data"],
	})
	require.NoError(t, err)
	require.Equal(t, "Total Bids by Device Os", res.Title)
	require.Equal(t, []any{"iOS", "Android"}, res.DomainValues["device_os"])
	require.EqualValues(t, 10, res.Input.ChartSpec["encodings"].(map[string]any)["color"].(map[string]any)["limit"])
	require.Len(t, res.Data, 2)
	require.EqualValues(t, 123, res.Totals["total_bids"])
	require.False(t, res.Truncated)
}

// TestCreateChartEmptyResult checks the contract for a time range with no rows: data is present as an empty array,
// and so are the field metadata and the resolved time range, so that the widget can show its empty state. chart_data behaves the same.
func TestCreateChartEmptyResult(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	// The data is from January 2024; this range is a year earlier.
	args := flintBarChartArgs()
	args["data"].(map[string]any)["time_range"] = map[string]any{"start": "2023-01-01T00:00:00Z", "end": "2023-01-08T00:00:00Z"}

	var res *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, args)
	require.NoError(t, err)
	require.Empty(t, res.Data)
	require.Len(t, res.Fields, 2)
	require.Equal(t, "dimension", res.Fields["advertiser_name"].Kind)
	require.Equal(t, "measure", res.Fields["total_bids"].Kind)
	require.NotNil(t, res.ResolvedTimeRange)
	require.Equal(t, "2023-01-01T00:00:00Z", res.ResolvedTimeRange.Start.UTC().Format("2006-01-02T15:04:05Z07:00"))
	require.False(t, res.Truncated)
	requireValidChartResult(t, res)

	data, err := json.Marshal(res)
	require.NoError(t, err)
	var keys map[string]any
	require.NoError(t, json.Unmarshal(data, &keys))
	require.Equal(t, []any{}, keys["data"])
	require.Contains(t, keys, "fields")
	require.Contains(t, keys, "resolved_time_range")

	var dataRes *ai.ChartDataResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, ai.ChartDataArgs(args))
	require.NoError(t, err)
	require.Empty(t, dataRes.Data)
	require.Len(t, dataRes.Fields, 2)
	require.NotNil(t, dataRes.ResolvedTimeRange)
	data, err = json.Marshal(dataRes)
	require.NoError(t, err)
	keys = nil
	require.NoError(t, json.Unmarshal(data, &keys))
	require.Equal(t, []any{}, keys["data"])
	require.Contains(t, keys, "fields")
	require.Contains(t, keys, "resolved_time_range")
}

// TestCreateChartAccess covers the user agent and feature flag matrix of create_chart and chart_data.
// Whether the MCP server serves the widget plays no part here: servers without the widget leave chart_data out at registration time
// (see TestMCPAppsWithoutWidget in runtime/server).
func TestCreateChartAccess(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	access := func(t *testing.T, s *ai.Session, name string) bool {
		tool, ok := s.Tool(name)
		require.True(t, ok)
		allowed, err := tool.CheckAccess(ai.WithSession(t.Context(), s))
		require.NoError(t, err)
		return allowed
	}

	// Default flags: every client gets create_chart; only external clients get chart_data, since in-app clients fetch chart data themselves.
	rill := newChartSession(t, rt, instanceID, "rill/1.0")
	external := newChartSession(t, rt, instanceID, "claude-ai/1.0")
	require.True(t, access(t, rill, ai.CreateChartName))
	require.False(t, access(t, rill, ai.ChartDataName))
	require.True(t, access(t, external, ai.CreateChartName))
	require.True(t, access(t, external, ai.ChartDataName))

	// The in-app result carries exactly the spec fields, not the data, and still passes the output schema (see requireValidChartResult).
	var res *ai.CreateChartResult
	_, err := rill.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, barChartArgs())
	require.NoError(t, err)
	require.Nil(t, res.ChartDataResult)
	requireValidChartResult(t, res)
	data, err := json.Marshal(res)
	require.NoError(t, err)
	var keys map[string]any
	require.NoError(t, json.Unmarshal(data, &keys))
	require.ElementsMatch(t, []string{"widget", "chart_type", "spec", "message"}, mapKeys(keys))

	// chart_data returns the data fields only.
	var dataRes *ai.ChartDataResult
	_, err = external.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, ai.ChartDataArgs(flintBarChartArgs()))
	require.NoError(t, err)
	require.Len(t, dataRes.Data, 2)
	require.Equal(t, []any{"Acme", "Beta"}, dataRes.DomainValues["advertiser_name"])
	require.Contains(t, dataRes.Fields, "total_bids")

	// chat_charts is the master switch for both tools. A session caches its feature flags, so the change is checked with new sessions.
	inst, err := rt.Instance(t.Context(), instanceID)
	require.NoError(t, err)
	inst.FeatureFlags = map[string]string{"chat_charts": "false"}
	require.NoError(t, rt.EditInstance(t.Context(), inst, false))
	rill = newChartSession(t, rt, instanceID, "rill/1.0")
	external = newChartSession(t, rt, instanceID, "claude-ai/1.0")
	require.False(t, access(t, rill, ai.CreateChartName))
	require.False(t, access(t, external, ai.CreateChartName))
	require.False(t, access(t, external, ai.ChartDataName))

	// mcp_charts turns charts off for external clients alone: Rill's own chat keeps create_chart.
	inst.FeatureFlags = map[string]string{"mcp_charts": "false"}
	require.NoError(t, rt.EditInstance(t.Context(), inst, false))
	rill = newChartSession(t, rt, instanceID, "rill/1.0")
	external = newChartSession(t, rt, instanceID, "claude-ai/1.0")
	require.True(t, access(t, rill, ai.CreateChartName))
	require.False(t, access(t, external, ai.CreateChartName))
	require.False(t, access(t, external, ai.ChartDataName))

	// Without ReadMetrics there is no access at all.
	inst.FeatureFlags = nil
	require.NoError(t, rt.EditInstance(t.Context(), inst, false))
	limited := newChartSessionWithClaims(t, rt, instanceID, "claude-ai/1.0", &runtime.SecurityClaims{UserID: uuid.NewString(), Permissions: []runtime.Permission{runtime.UseAI}})
	require.False(t, access(t, limited, ai.CreateChartName))
	require.False(t, access(t, limited, ai.ChartDataName))
}

func newChartSessionWithClaims(t *testing.T, rt *runtime.Runtime, instanceID, userAgent string, claims *runtime.SecurityClaims) *ai.Session {
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     claims,
		UserAgent:  userAgent,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, s.Flush(t.Context()))
	})
	return s
}

// requireValidChartResult validates a create_chart result the way the MCP server does: as JSON against the tool's inferred output schema.
// Both the external shape (with data, as [] when empty) and the in-app shape (without the data fields) must pass,
// which is why ChartDataResult.Data is omitzero rather than required.
func requireValidChartResult(t *testing.T, res *ai.CreateChartResult) {
	t.Helper()
	schema, ok := ai.MCPToolSpecs()[ai.CreateChartName].OutputSchema.(*jsonschema.Schema)
	require.True(t, ok)
	resolved, err := schema.Resolve(nil)
	require.NoError(t, err)
	data, err := json.Marshal(res)
	require.NoError(t, err)
	var v any
	require.NoError(t, json.Unmarshal(data, &v))
	require.NoError(t, resolved.Validate(v))
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// flintBarChartArgs are create_chart arguments in the Flint shape external clients use (see CreateChart.MCPSpec).
func flintBarChartArgs() ai.CreateChartArgs {
	return ai.CreateChartArgs{
		"chart_spec": map[string]any{
			"chartType": "Bar Chart",
			"encodings": map[string]any{
				"x": map[string]any{"field": "advertiser_name", "limit": 2, "sortBy": "y", "sortOrder": "descending"},
				"y": "total_bids",
			},
		},
		"data": map[string]any{
			"metrics_view": "bids_metrics",
			"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-08T00:00:00Z"},
		},
	}
}

// TestCreateChartFlint checks the result an external MCP client receives for the Flint shape of the arguments:
// the effective input (chart spec with the applied limits, data block with the resolved grain), the chart-ready rows with their field metadata,
// the chart compiler's semantic types and display names, the metrics view summary, the theme, a title from the bound fields and the link to open the chart in Rill.
// It also checks that errors name the valid chart types and channels, and that each client is held to the shape it was taught.
func TestCreateChartFlint(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:       createChartTestFiles("theme: brand\n"),
		FrontendURL: "https://ui.rilldata.com/demo/bids",
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	var res *ai.CreateChartResult
	toolRes, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, flintBarChartArgs())
	require.NoError(t, err)

	require.Equal(t, "chart", res.Widget)
	require.Equal(t, "Bar Chart", res.ChartType)
	require.Nil(t, res.Spec, "the in-app spec is not part of an external result")
	require.Equal(t, "Chart created successfully: Bar Chart", res.Message)
	require.Equal(t, "Total Bids by Advertiser Name", res.Title)
	require.Equal(t, &ai.ChartMetricsView{Name: "bids_metrics", DisplayName: "Bids", TimeDimension: "event_time", SmallestTimeGrain: "TIME_GRAIN_HOUR"}, res.MetricsView)

	// The effective input: the chart spec with the applied limit, the rest of the encodings as given
	require.NotNil(t, res.Input)
	require.Equal(t, "Bar Chart", res.Input.ChartSpec["chartType"])
	encodings := res.Input.ChartSpec["encodings"].(map[string]any)
	require.EqualValues(t, 2, encodings["x"].(map[string]any)["limit"])
	require.Equal(t, "y", encodings["x"].(map[string]any)["sortBy"])
	require.Equal(t, "total_bids", encodings["y"], "a plain field without a top-N is echoed as is")
	require.Equal(t, "bids_metrics", res.Input.Data["metrics_view"])
	require.NotContains(t, res.Input.Data, "time_grain", "no time field is bound")

	// Top 2 advertisers by total bids, nulls excluded, in the top-N order
	require.Equal(t, []any{"Acme", "Beta"}, res.DomainValues["advertiser_name"])
	require.Len(t, res.Data, 2)
	require.Equal(t, "Acme", res.Data[0]["advertiser_name"])
	require.EqualValues(t, 15, res.Data[0]["total_bids"])
	require.Equal(t, "Beta", res.Data[1]["advertiser_name"])
	require.EqualValues(t, 7, res.Data[1]["total_bids"])
	require.Empty(t, res.TimeGrain, "no time field is bound")
	require.True(t, res.Truncated, "a third advertiser exists beyond the limit")
	require.False(t, res.RowsTruncated, "only a top-N was cut")
	require.Contains(t, res.TruncationWarning, `only the top 2 values of "advertiser_name" are shown`)

	// Field metadata for the renderer
	require.Len(t, res.Fields, 2)
	require.Equal(t, &ai.ChartField{Name: "advertiser_name", Kind: "dimension", DisplayName: "Advertiser Name", Type: "STRING"}, res.Fields["advertiser_name"])
	require.Equal(t, "measure", res.Fields["total_bids"].Kind)
	require.Equal(t, "Total Bids", res.Fields["total_bids"].DisplayName)
	require.Equal(t, "humanize", res.Fields["total_bids"].FormatPreset)
	require.NotEmpty(t, res.Fields["total_bids"].Type)
	require.Equal(t, "2024-01-01T00:00:00Z", res.ResolvedTimeRange.Start.UTC().Format("2006-01-02T15:04:05Z07:00"))
	require.Nil(t, res.ResolvedComparisonTimeRange)
	require.Nil(t, res.Totals)

	// The chart compiler's inputs: an integer measure is a Count, a dimension a Category
	require.Equal(t, map[string]string{"advertiser_name": "Category", "total_bids": "Count"}, res.SemanticTypes)
	require.Equal(t, map[string]string{"advertiser_name": "Advertiser Name", "total_bids": "Total Bids"}, res.FieldDisplayNames)

	// The project theme as protojson
	require.Equal(t, "#4f46e5", res.Theme["primaryColorRaw"])
	require.Equal(t, "#f59e0b", res.Theme["secondaryColorRaw"])
	require.Equal(t, fmt.Sprintf("https://ui.rilldata.com/demo/bids/-/ai/%s/message/%s/-/open", s.ID(), toolRes.Call.ID), res.OpenURL)
	requireValidChartResult(t, res)

	// The exact JSON an external client receives
	data, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)
	t.Logf("create_chart result for an external client:\n%s", data)
	var keys map[string]any
	require.NoError(t, json.Unmarshal(data, &keys))
	for _, k := range []string{"widget", "chart_type", "input", "message", "title", "metrics_view", "data", "fields", "semantic_types", "field_display_names", "domain_values", "resolved_time_range", "theme", "open_url", "truncated", "truncation_warning"} {
		require.Contains(t, keys, k)
	}
	for _, k := range []string{"spec", "version", "time_grain", "totals", "resolved_comparison_time_range", "rows_truncated"} {
		require.NotContains(t, keys, k)
	}

	// A time series with a comparison and a headline: the time field is floored at the grain, its semantic type follows the grain,
	// the comparison values take the measure's type (here the caller's override) and a "(previous period)" display name.
	lineArgs := ai.CreateChartArgs{
		"chart_spec": map[string]any{
			"chartType": "Line Chart",
			"title":     "  Bids per day  ",
			"encodings": map[string]any{"x": "event_time", "y": "total_bids"},
		},
		"data": map[string]any{
			"metrics_view":          "bids_metrics",
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-04T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2023-12-31T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
		},
		"semantic_types": map[string]any{"total_bids": "Quantity"},
	}
	var line *ai.CreateChartResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &line, lineArgs)
	require.NoError(t, err)
	require.Equal(t, "Bids per day", line.Title, "the headline wins")
	require.Equal(t, "TIME_GRAIN_DAY", line.TimeGrain, "the grain defaults to day")
	require.Equal(t, "TIME_GRAIN_DAY", line.Input.Data["time_grain"])
	require.Equal(t, map[string]any{"total_bids": "Quantity"}, line.Input.SemanticTypes, "the input echoes the caller's overrides for the widget's refetches")
	require.Equal(t, "Date", line.SemanticTypes["event_time"])
	require.Equal(t, "Quantity", line.SemanticTypes["total_bids"], "the caller's override wins")
	require.Equal(t, "Quantity", line.SemanticTypes["total_bids_prev"])
	require.Equal(t, "Time", line.FieldDisplayNames["event_time"])
	require.Equal(t, "Total Bids (previous period)", line.FieldDisplayNames["total_bids_prev"])
	require.Equal(t, "time", line.Fields["event_time"].Kind)
	require.Contains(t, line.Data[0], "total_bids_prev")
	require.Equal(t, "2024-01-02T00:00:00Z", line.Data[0]["event_time"])
	require.NotNil(t, line.ResolvedComparisonTimeRange)

	// The widget draws the previous period as a second series, so a comparison with a series dimension is rejected rather than charted without it.
	lineArgs["chart_spec"].(map[string]any)["encodings"].(map[string]any)["color"] = "device_os"
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &line, lineArgs)
	require.ErrorContains(t, err, `cannot also bind "device_os" on color`)

	// Without a headline, a time field on x reads "over Time" and a series dimension "split by"
	var area *ai.CreateChartResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &area, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Area Chart", "encodings": map[string]any{"x": "event_time", "y": "total_bids", "color": "device_os"}},
		"data":       map[string]any{"metrics_view": "bids_metrics", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-08T00:00:00Z"}, "time_grain": "TIME_GRAIN_MONTH"},
	})
	require.NoError(t, err)
	require.Equal(t, "Total Bids over Time split by Device Os", area.Title)
	require.Equal(t, "YearMonth", area.SemanticTypes["event_time"])

	// Errors name the valid chart types and channels so the model can repair the call
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Sankey", "encodings": map[string]any{"x": "advertiser_name", "y": "total_bids"}},
		"data":       flintBarChartArgs()["data"],
	})
	require.ErrorContains(t, err, `unknown chart type "Sankey"`)
	require.ErrorContains(t, err, "Waterfall Chart")
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "advertiser_name", "y": "total_bids", "goal": "bid_count"}},
		"data":       flintBarChartArgs()["data"],
	})
	require.ErrorContains(t, err, `does not accept channel "goal"`)
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_spec": map[string]any{"chartType": "Bar Chart", "encodings": map[string]any{"x": "publisher", "y": "total_bids"}},
		"data":       flintBarChartArgs()["data"],
	})
	require.ErrorContains(t, err, `field "publisher" on channel "x" not found`)

	// chart_data takes the Flint shape too and returns the data section with the compiler inputs,
	// the caller's semantic types applied, and the effective input as create_chart returns it.
	dataArgs := ai.ChartDataArgs(flintBarChartArgs())
	dataArgs["chart_spec"].(map[string]any)["encodings"].(map[string]any)["x"] = "advertiser_name"
	dataArgs["semantic_types"] = map[string]any{"total_bids": "Amount"}
	var dataRes *ai.ChartDataOutput
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, dataArgs)
	require.NoError(t, err)
	require.Len(t, dataRes.Data, 3)
	require.Equal(t, "Category", dataRes.SemanticTypes["advertiser_name"])
	require.Equal(t, "Amount", dataRes.SemanticTypes["total_bids"])
	require.Equal(t, "Total Bids", dataRes.FieldDisplayNames["total_bids"])
	require.NotNil(t, dataRes.Input)
	require.Equal(t, map[string]any{"field": "advertiser_name", "limit": float64(20)}, dataRes.Input.ChartSpec["encodings"].(map[string]any)["x"], "the default limit that was applied")
	require.Equal(t, map[string]any{"total_bids": "Amount"}, dataRes.Input.SemanticTypes)

	// Each client is held to the shape it was taught: Rill's own chat to the in-app shape, external clients to the Flint shape.
	rill := newChartSession(t, rt, instanceID, "rill/1.0")
	var inApp *ai.CreateChartResult
	_, err = rill.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &inApp, flintBarChartArgs())
	require.ErrorContains(t, err, "chart_spec and data are not supported here: pass chart_type")
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, barChartArgs())
	require.ErrorContains(t, err, "chart_spec and data are required")
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, ai.ChartDataArgs(barChartArgs()))
	require.ErrorContains(t, err, "chart_spec and data are required")

	// A call in neither shape is told about the shape the client was taught.
	_, err = rill.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &inApp, ai.CreateChartArgs{})
	require.ErrorContains(t, err, "chart_type is required")
	require.NotContains(t, err.Error(), "chart_spec")
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{})
	require.ErrorContains(t, err, "chart_spec and data are required")
}

// TestCreateChartFlintOtherSlice checks the "Other" slice of a pie chart whose slices were cut to the top N:
// a sum gets the rest of the total as one more slice, an average gets none, since its slices share no whole.
func TestCreateChartFlintOtherSlice(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)
	s := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	pie := func(t *testing.T, measure string) *ai.CreateChartResult {
		var res *ai.CreateChartResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
			"chart_spec": map[string]any{"chartType": "Pie Chart", "encodings": map[string]any{"size": measure, "color": map[string]any{"field": "advertiser_name", "limit": 2}}},
			"data":       flintBarChartArgs()["data"],
		})
		require.NoError(t, err)
		return res
	}

	// Acme (15) and Beta (7) are shown; Gamma (1) is the rest of the 23 bids of known advertisers.
	sum := pie(t, "total_bids")
	require.Len(t, sum.Data, 3)
	require.Equal(t, "Other", sum.Data[2]["advertiser_name"])
	require.EqualValues(t, 1, sum.Data[2]["total_bids"])
	require.Equal(t, []any{"Acme", "Beta", "Other"}, sum.DomainValues["advertiser_name"])
	require.EqualValues(t, 23, sum.Totals["total_bids"])
	require.True(t, sum.Truncated)
	require.Contains(t, sum.TruncationWarning, `the rest are combined into "Other"`)

	avg := pie(t, "avg_bids")
	require.Len(t, avg.Data, 2)
	require.Equal(t, []any{"Acme", "Beta"}, avg.DomainValues["advertiser_name"])
	require.NotContains(t, avg.TruncationWarning, "Other")
}

// TestCreateChartDashboardFilters tests that create_chart checks the chart's 'where' against the dashboard's filters.
// The matching itself is tested in metricsview.ValidateExpressionNarrows; this covers the wiring around it.
func TestCreateChartDashboardFilters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 'desktop' AS device, 100 AS revenue
  UNION ALL
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 'mobile' AS device, 10 AS revenue
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
dimensions:
- column: country
- column: device
measures:
- name: revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newSession(t, rt, instanceID)

	// chartArgs builds a bar chart of revenue by country on the given metrics view, with the given 'where' (omitted when nil).
	chartArgs := func(metricsView string, where map[string]any) ai.CreateChartArgs {
		spec := map[string]any{
			"metrics_view": metricsView,
			"time_range":   map[string]any{"start": "2025-01-01T00:00:00Z", "end": "2025-01-02T00:00:00Z"},
			"x":            map[string]any{"field": "country", "type": "nominal"},
			"y":            map[string]any{"field": "revenue", "type": "quantitative"},
		}
		if where != nil {
			spec["where"] = where
		}
		return ai.CreateChartArgs{"chart_type": "bar_chart", "spec": spec}
	}
	// cond builds a chart 'where' such as country = 'Denmark', in the JSON shape the model sends.
	cond := func(op, dimension string, value any) map[string]any {
		return map[string]any{"cond": map[string]any{
			"op":    op,
			"exprs": []any{map[string]any{"name": dimension}, map[string]any{"val": value}},
		}}
	}

	// The user is viewing the "orders" dashboard filtered to Denmark, in the shape the dashboards send.
	denmarkFilter := &metricsview.Expression{Condition: &metricsview.Condition{
		Operator:    metricsview.OperatorIn,
		Expressions: []*metricsview.Expression{{Name: "country"}, {Value: "Denmark"}},
	}}
	denmarkCtx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": denmarkFilter})

	t.Run("rejects a chart without a filter", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs("orders", nil))
		require.ErrorContains(t, err, "the chart has no 'where' filter")
		require.ErrorContains(t, err, "country IN ('Denmark')")
	})

	t.Run("rejects a chart whose filter does not narrow the dashboard's filter", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs("orders", cond("eq", "device", "mobile")))
		require.ErrorContains(t, err, `no filter on "country"`)
	})

	t.Run("matches the metrics view name case-insensitively", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs("Orders", nil))
		require.ErrorContains(t, err, "the chart has no 'where' filter")
	})

	t.Run("accepts a chart that restates the dashboard's filter", func(t *testing.T) {
		var res *ai.CreateChartResult
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, &res, chartArgs("orders", cond("eq", "country", "Denmark")))
		require.NoError(t, err)
		require.NotNil(t, res.Spec["where"])
	})

	t.Run("accepts an explicit empty filter", func(t *testing.T) {
		var res *ai.CreateChartResult
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, &res, chartArgs("orders", map[string]any{"cond": map[string]any{"op": "and", "exprs": []any{}}}))
		require.NoError(t, err)
		require.Equal(t, "bar_chart", res.ChartType)
	})

	t.Run("ignores filters on other metrics views", func(t *testing.T) {
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"other": denmarkFilter})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs("orders", nil))
		require.NoError(t, err)
	})

	t.Run("ignores empty dashboard filters", func(t *testing.T) {
		emptyFilter := &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd}}
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": emptyFilter})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs("orders", nil))
		require.NoError(t, err)
	})

	t.Run("accepts a chart when the dashboard has no filters", func(t *testing.T) {
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, nil, chartArgs("orders", nil))
		require.NoError(t, err)
	})
}
