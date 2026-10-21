package ai_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
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

// TestCreateChartExternalClient checks the result an external MCP client receives: the chart spec plus the chart-ready data,
// field metadata, metrics view summary, theme, title and the link to open the chart in Rill.
func TestCreateChartExternalClient(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:       createChartTestFiles("theme: brand\n"),
		FrontendURL: "https://ui.rilldata.com/demo/bids",
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	var res *ai.CreateChartResult
	toolRes, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, barChartArgs())
	require.NoError(t, err)

	// The contract fields that are always present; the spec echoes the limit that was applied (see TestCreateChartSpecLimits)
	require.Equal(t, "chart", res.Widget)
	require.Equal(t, "bar_chart", res.ChartType)
	require.Equal(t, "bids_metrics", res.Spec["metrics_view"])
	require.EqualValues(t, 2, res.Spec["x"].(map[string]any)["limit"])
	require.Equal(t, "Chart created successfully: bar_chart", res.Message)

	// The external-only fields
	require.Equal(t, "Total Bids per Advertiser Name", res.Title)
	require.Equal(t, &ai.ChartMetricsView{Name: "bids_metrics", DisplayName: "Bids", TimeDimension: "event_time", SmallestTimeGrain: "TIME_GRAIN_HOUR"}, res.MetricsView)
	require.NotNil(t, res.ChartDataResult)
	require.Empty(t, res.TimeGrain, "no temporal field in the chart")

	// Top 2 advertisers by total bids, nulls excluded, in the top-N order
	require.Equal(t, []any{"Acme", "Beta"}, res.DomainValues["advertiser_name"])
	require.Len(t, res.Data, 2)
	require.Equal(t, "Acme", res.Data[0]["advertiser_name"])
	require.EqualValues(t, 15, res.Data[0]["total_bids"])
	require.Equal(t, "Beta", res.Data[1]["advertiser_name"])
	require.EqualValues(t, 7, res.Data[1]["total_bids"])
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
	for _, k := range []string{"widget", "chart_type", "spec", "message", "title", "metrics_view", "data", "fields", "domain_values", "resolved_time_range", "theme", "open_url", "truncated", "truncation_warning"} {
		require.Contains(t, keys, k)
	}
	for _, k := range []string{"version", "time_grain", "totals", "resolved_comparison_time_range", "rows_truncated"} {
		require.NotContains(t, keys, k)
	}
}

// TestCreateChartSpecLimits checks that an external result's spec carries the effective limit of every channel that got a top-N
// (the default where the spec set none, and on the input's y channel for a horizontal bar chart),
// while the in-app result returns the spec as given.
func TestCreateChartSpecLimits(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	timeRange := map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-08T00:00:00Z"}
	external := newChartSession(t, rt, instanceID, "claude-ai/1.0")

	// The defaults of a nominal x axis and a color split
	var stacked *ai.CreateChartResult
	_, err := external.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &stacked, ai.CreateChartArgs{
		"chart_type": "stacked_bar",
		"spec": map[string]any{
			"metrics_view": "bids_metrics",
			"time_range":   timeRange,
			"color":        map[string]any{"field": "device_os", "type": "nominal"},
			"x":            map[string]any{"field": "advertiser_name", "type": "nominal", "sort": "-y"},
			"y":            map[string]any{"field": "total_bids", "type": "quantitative"},
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, 20, stacked.Spec["x"].(map[string]any)["limit"])
	require.EqualValues(t, 10, stacked.Spec["color"].(map[string]any)["limit"])
	require.Equal(t, "-y", stacked.Spec["x"].(map[string]any)["sort"], "the rest of the channel is kept")
	require.NotContains(t, stacked.Spec["y"], "limit")

	// A horizontal bar chart carries the dimension, and so the limit, on y
	horizontal := ai.CreateChartArgs{
		"chart_type": "bar_chart",
		"spec": map[string]any{
			"metrics_view": "bids_metrics",
			"time_range":   timeRange,
			"color":        "primary",
			"x":            map[string]any{"field": "total_bids", "type": "quantitative"},
			"y":            map[string]any{"field": "advertiser_name", "type": "nominal", "sort": "-x"},
		},
	}
	var bar *ai.CreateChartResult
	_, err = external.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &bar, horizontal)
	require.NoError(t, err)
	require.EqualValues(t, 20, bar.Spec["y"].(map[string]any)["limit"])
	require.NotContains(t, bar.Spec["x"], "limit")
	require.Equal(t, []any{"Acme", "Beta", "Gamma"}, bar.DomainValues["advertiser_name"])

	// The in-app result returns the spec as given
	rill := newChartSession(t, rt, instanceID, "rill/1.0")
	var inApp *ai.CreateChartResult
	_, err = rill.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &inApp, horizontal)
	require.NoError(t, err)
	require.NotContains(t, inApp.Spec["y"], "limit")
	require.Nil(t, inApp.ChartDataResult)
}

// TestCreateChartTemporalComparison checks time flooring, comparison values and the time grain in the result.
func TestCreateChartTemporalComparison(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "chatgpt/1.0")

	var res *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_type": "line_chart",
		"spec": map[string]any{
			"metrics_view":          "bids_metrics",
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-04T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2023-12-31T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
			"time_grain":            "TIME_GRAIN_DAY",
			"color":                 "primary",
			"x":                     map[string]any{"field": "event_time", "type": "temporal"},
			"y":                     map[string]any{"field": "total_bids", "type": "quantitative"},
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
	require.NotContains(t, res.Spec["x"], "limit", "a temporal axis gets no top-N")
}

// TestCreateChartDonutTotals checks the total query of circular charts.
func TestCreateChartDonutTotals(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: createChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newChartSession(t, rt, instanceID, "cursor/1.0")

	var res *ai.CreateChartResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &res, ai.CreateChartArgs{
		"chart_type": "donut_chart",
		"spec": map[string]any{
			"metrics_view": "bids_metrics",
			"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-01-08T00:00:00Z"},
			"color":        map[string]any{"field": "device_os", "type": "nominal", "limit": 10},
			"measure":      map[string]any{"field": "total_bids", "type": "quantitative"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Total Bids split by Device Os", res.Title)
	require.Equal(t, []any{"iOS", "Android"}, res.DomainValues["device_os"])
	require.EqualValues(t, 10, res.Spec["color"].(map[string]any)["limit"])
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
	args := barChartArgs()
	args["spec"].(map[string]any)["time_range"] = map[string]any{"start": "2023-01-01T00:00:00Z", "end": "2023-01-08T00:00:00Z"}

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
	_, err = external.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, ai.ChartDataArgs(barChartArgs()))
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

// TestCreateChartFlint checks the Flint shape of the arguments: the result echoes the effective input (chart spec with the applied
// limits, data block with the resolved grain), carries the chart compiler's semantic types and display names next to the rows,
// titles the chart from the bound fields, and reports bad chart types with the valid names.
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
	require.Nil(t, res.Spec, "the in-app spec is not part of a Flint result")
	require.Equal(t, "Chart created successfully: Bar Chart", res.Message)
	require.Equal(t, "Total Bids by Advertiser Name", res.Title)

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
	require.True(t, res.Truncated)
	require.Equal(t, "dimension", res.Fields["advertiser_name"].Kind)
	require.Equal(t, "measure", res.Fields["total_bids"].Kind)

	// The chart compiler's inputs: an integer measure is a Count, a dimension a Category
	require.Equal(t, map[string]string{"advertiser_name": "Category", "total_bids": "Count"}, res.SemanticTypes)
	require.Equal(t, map[string]string{"advertiser_name": "Advertiser Name", "total_bids": "Total Bids"}, res.FieldDisplayNames)

	require.Equal(t, "#4f46e5", res.Theme["primaryColorRaw"])
	require.Equal(t, fmt.Sprintf("https://ui.rilldata.com/demo/bids/-/ai/%s/message/%s/-/open", s.ID(), toolRes.Call.ID), res.OpenURL)
	requireValidChartResult(t, res)

	data, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)
	t.Logf("create_chart result for a Flint call:\n%s", data)
	var keys map[string]any
	require.NoError(t, json.Unmarshal(data, &keys))
	for _, k := range []string{"widget", "chart_type", "input", "message", "title", "metrics_view", "data", "fields", "semantic_types", "field_display_names", "domain_values", "resolved_time_range", "theme", "open_url", "truncated"} {
		require.Contains(t, keys, k)
	}
	require.NotContains(t, keys, "spec")

	// A time series with a comparison and a headline: the time field is floored at the grain, its semantic type follows the grain,
	// the comparison values take the measure's type (here the caller's override) and a "(previous period)" display name.
	var line *ai.CreateChartResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &line, ai.CreateChartArgs{
		"chart_spec": map[string]any{
			"chartType": "Line Chart",
			"title":     "  Bids per day  ",
			"encodings": map[string]any{"x": "event_time", "y": "total_bids", "color": "device_os"},
		},
		"data": map[string]any{
			"metrics_view":          "bids_metrics",
			"time_range":            map[string]any{"start": "2024-01-02T00:00:00Z", "end": "2024-01-04T00:00:00Z"},
			"comparison_time_range": map[string]any{"start": "2023-12-31T00:00:00Z", "end": "2024-01-02T00:00:00Z"},
		},
		"semantic_types": map[string]any{"total_bids": "Quantity"},
	})
	require.NoError(t, err)
	require.Equal(t, "Bids per day", line.Title, "the headline wins")
	require.Equal(t, "TIME_GRAIN_DAY", line.TimeGrain, "the grain defaults to day")
	require.Equal(t, "TIME_GRAIN_DAY", line.Input.Data["time_grain"])
	require.EqualValues(t, 10, line.Input.ChartSpec["encodings"].(map[string]any)["color"].(map[string]any)["limit"], "a series channel defaults to 10")
	require.Equal(t, "Date", line.SemanticTypes["event_time"])
	require.Equal(t, "Quantity", line.SemanticTypes["total_bids"], "the caller's override wins")
	require.Equal(t, "Quantity", line.SemanticTypes["total_bids_prev"])
	require.Equal(t, "Category", line.SemanticTypes["device_os"])
	require.Equal(t, "Time", line.FieldDisplayNames["event_time"])
	require.Equal(t, "Total Bids (previous period)", line.FieldDisplayNames["total_bids_prev"])
	require.Equal(t, "time", line.Fields["event_time"].Kind)
	require.Contains(t, line.Data[0], "total_bids_prev")
	require.Equal(t, "2024-01-02T00:00:00Z", line.Data[0]["event_time"])
	require.NotNil(t, line.ResolvedComparisonTimeRange)

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

	// chart_data takes the Flint shape too and returns the data section with the compiler inputs
	var dataRes *ai.ChartDataResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ChartDataName, &dataRes, ai.ChartDataArgs(flintBarChartArgs()))
	require.NoError(t, err)
	require.Len(t, dataRes.Data, 2)
	require.Equal(t, "Category", dataRes.SemanticTypes["advertiser_name"])
	require.Equal(t, "Total Bids", dataRes.FieldDisplayNames["total_bids"])

	// The in-app client is not offered the Flint shape, but a call in that shape still echoes the input without data
	rill := newChartSession(t, rt, instanceID, "rill/1.0")
	var inApp *ai.CreateChartResult
	_, err = rill.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, &inApp, flintBarChartArgs())
	require.NoError(t, err)
	require.Nil(t, inApp.ChartDataResult)
	require.Nil(t, inApp.Spec)
	require.Equal(t, "Bar Chart", inApp.Input.ChartSpec["chartType"])
}
