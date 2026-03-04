package ai_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestAnalystBasic(t *testing.T) {
	// Setup a basic metrics view with an "event_time" time dimension, "country" dimension, and "count" and "revenue" measures.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 100 AS revenue
  UNION ALL
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 10 AS revenue
  UNION ALL
  SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 100 AS revenue
  UNION ALL
  SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 10 AS revenue
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
- name: revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Analyst agent question
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "What country has the highest revenue? Answer with a single country name and nothing else.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.AnalystAgentName, res.Agent)
	require.Equal(t, "United States", res.Response)

	// Analyst agent question that references the previous response
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Repeat the answer you gave to my last question.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.AnalystAgentName, res.Agent)
	require.Equal(t, "United States", res.Response)
}

func TestAnalystOpenRTB(t *testing.T) {
	// Setup runtime instance with the OpenRTB dataset
	n, files := testruntime.ProjectOpenRTB(t)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files:       files,
		FrontendURL: "https://ui.rilldata.com/-/dashboards/bids_metrics",
	})
	testruntime.RequireReconcileState(t, rt, instanceID, n, 0, 0)

	// Test it remembers previous tool calls over
	t.Run("MultipleTurns", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		// Check the only sub-call is the seeded list_metrics_views call
		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "What metrics views are available?",
		})
		require.NoError(t, err)
		calls := s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)

		// It should make one sub-call, get_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Tell me about the auction metrics (but don't query the data)",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)
		require.Equal(t, ai.GetMetricsViewName, calls[0].Tool)

		// It should make one sub-call, get_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Now tell me about the other dataset (but don't query the data)",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)
		require.Equal(t, ai.GetMetricsViewName, calls[0].Tool)

		// It should remember the previous turns and only make one sub-call: query_metrics_view_summary and query_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Tell me which non-US country has the most auctions. Make the minimal number of tool calls necessary to answer.",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 2)
		require.Equal(t, ai.QueryMetricsViewSummaryName, calls[0].Tool)
		require.Equal(t, ai.QueryMetricsViewName, calls[1].Tool)
	})

	t.Run("DashboardContext", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		// It should make three sub-calls: query_metrics_view_summary, get_metrics_view, query_metrics_view
		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:    "Tell me which app_site_name has the most impressions. When calling tools, you must only make one call total to the `query_metrics_view` tool.",
			Explore:   "bids_metrics",
			TimeStart: parseTestTime(t, "2023-09-11T00:00:00Z"),
			TimeEnd:   parseTestTime(t, "2023-09-14T00:00:00Z"),
			Where: &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{
						{Name: "device_os"},
						{Value: "Android"},
					},
				},
			},
		})
		require.NoError(t, err)
		calls := s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 3)
		require.Equal(t, ai.QueryMetricsViewSummaryName, calls[0].Tool)
		require.Equal(t, ai.GetMetricsViewName, calls[1].Tool)
		require.Equal(t, ai.QueryMetricsViewName, calls[2].Tool)

		// Map the request sent and assert that context was honored.
		rawQry, err := s.UnmarshalMessageContent(calls[2])
		require.NoError(t, err)
		var qry metricsview.Query
		err = mapstructureutil.WeakDecode(rawQry, &qry)
		require.NoError(t, err)
		// Assert that the time range is sent using the context
		require.Equal(t, parseTestTime(t, "2023-09-11T00:00:00Z"), qry.TimeRange.Start)
		require.Equal(t, parseTestTime(t, "2023-09-14T00:00:00Z"), qry.TimeRange.End)
		// Assert that the filter is sent using the context.
		// Checking using the SQL representation since the LLM's use of nesting in the expression is unstable.
		exprSQL, err := metricsview.ExpressionToSQL(qry.Where)
		require.NoError(t, err)
		require.Equal(t, "device_os = 'Android'", strings.Trim(exprSQL, "()"))

		rawRes, err := s.UnmarshalMessageContent(res.Result)
		require.NoError(t, err)
		var agentRes ai.AnalystAgentResult
		err = mapstructureutil.WeakDecode(rawRes, &agentRes)
		require.NoError(t, err)
		expectedCitationUrl := fmt.Sprintf(`https://ui.rilldata.com/-/dashboards/bids_metrics/-/ai/%s/message/%s/-/open`, s.ID(), calls[2].ID)
		require.Contains(t, agentRes.Response, expectedCitationUrl)
	})

	t.Run("CanvasContext", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		// It should make three sub-calls: query_metrics_view_summary, get_metrics_view, query_metrics_view
		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:          "Tell me which advertiser_name has the highest overall_spend. Make the minimal number of tool calls necessary to answer.",
			Canvas:          "bids_canvas",
			CanvasComponent: "bids_canvas--component-2-0",
			TimeStart:       parseTestTime(t, "2023-09-11T00:00:00Z"),
			TimeEnd:         parseTestTime(t, "2023-09-14T00:00:00Z"),
			WherePerMetricsView: map[string]*metricsview.Expression{
				"bids_metrics": {
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorEq,
						Expressions: []*metricsview.Expression{
							{Name: "auction_type"},
							{Value: "First Price"},
						},
					},
				},
			},
		})
		require.NoError(t, err)
		calls := s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 4)
		require.Equal(t, ai.GetCanvasName, calls[0].Tool)
		require.Equal(t, ai.QueryMetricsViewSummaryName, calls[1].Tool)
		require.Equal(t, ai.GetMetricsViewName, calls[2].Tool)
		require.Equal(t, ai.QueryMetricsViewName, calls[3].Tool)

		// Map the request sent and assert that context was honored.
		rawQry, err := s.UnmarshalMessageContent(calls[3])
		require.NoError(t, err)
		var qry metricsview.Query
		err = mapstructureutil.WeakDecode(rawQry, &qry)
		require.NoError(t, err)
		// Assert that the time range is sent using the context
		require.Equal(t, parseTestTime(t, "2023-09-11T00:00:00Z"), qry.TimeRange.Start)
		require.Equal(t, parseTestTime(t, "2023-09-14T00:00:00Z"), qry.TimeRange.End)

		// Assert that the filter is sent using the context
		// Checking using the SQL representation since the LLM's use of nesting in the expression is unstable.
		require.NotNil(t, qry.Where)
		exprSQL, err := metricsview.ExpressionToSQL(qry.Where)
		require.NoError(t, err)
		require.Contains(t, exprSQL, "auction_type = 'First Price'")
		require.Contains(t, exprSQL, "app_or_site = 'App'")
	})
}

func TestAnalystCharts(t *testing.T) {
	// Setup runtime instance with the OpenRTB dataset
	n, files := testruntime.ProjectOpenRTB(t)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files:       files,
		FrontendURL: "https://ui.rilldata.com/-/dashboards/bids_metrics",
	})
	testruntime.RequireReconcileState(t, rt, instanceID, n, 0, 0)

	t.Run("BasicBarChart", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:  "Show me the top 10 advertisers by total bids as a bar chart. Use the bids_metrics dataset and the full available time range. Create a single chart and nothing else.",
			Explore: "bids_metrics",
		})
		require.NoError(t, err)

		chartCalls := findChartCalls(s, res.Call.ID)
		require.GreaterOrEqual(t, len(chartCalls), 1, "expected at least one create_chart call")

		spec := requireValidChartSpec(t, s, chartCalls[0], []string{"bar_chart"})
		if x, ok := spec["x"].(map[string]any); ok {
			if field, ok := x["field"].(string); ok {
				require.Equal(t, "advertiser_name", field)
			}
		}
		if y, ok := spec["y"].(map[string]any); ok {
			if field, ok := y["field"].(string); ok {
				require.Equal(t, "total_bids", field)
			}
		}
	})

	t.Run("TimeSeriesChart", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:  "Show me the trend of total bids over time as a line chart. Use the bids_metrics dataset and the full available time range. Create a single chart and nothing else.",
			Explore: "bids_metrics",
		})
		require.NoError(t, err)

		chartCalls := findChartCalls(s, res.Call.ID)
		require.GreaterOrEqual(t, len(chartCalls), 1, "expected at least one create_chart call")

		spec := requireValidChartSpec(t, s, chartCalls[0], []string{"line_chart", "area_chart"})
		if x, ok := spec["x"].(map[string]any); ok {
			if field, ok := x["field"].(string); ok {
				require.Equal(t, "__time", field)
			}
		}
		_, hasTimeGrain := spec["time_grain"]
		require.True(t, hasTimeGrain, "time series chart should have time_grain")
	})

	t.Run("ChartWithDashboardContext", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:    "Show me total bids by device OS as a bar chart. Create a single chart and nothing else.",
			Explore:   "bids_metrics",
			TimeStart: parseTestTime(t, "2023-09-11T00:00:00Z"),
			TimeEnd:   parseTestTime(t, "2023-09-14T00:00:00Z"),
			Where: &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorEq,
					Expressions: []*metricsview.Expression{
						{Name: "auction_type"},
						{Value: "First Price"},
					},
				},
			},
		})
		require.NoError(t, err)

		chartCalls := findChartCalls(s, res.Call.ID)
		require.GreaterOrEqual(t, len(chartCalls), 1, "expected at least one create_chart call")

		spec := requireValidChartSpec(t, s, chartCalls[0], nil)

		// Verify time range honors the context
		timeRange, ok := spec["time_range"].(map[string]any)
		require.True(t, ok, "spec must contain time_range object")
		start, ok := timeRange["start"].(string)
		require.True(t, ok, "time_range.start must be a string")
		require.Contains(t, start, "2023-09-11", "time_range.start should honor the provided context")
		end, ok := timeRange["end"].(string)
		require.True(t, ok, "time_range.end must be a string")
		require.Contains(t, end, "2023-09-14", "time_range.end should honor the provided context")
	})

	t.Run("ChartsDisabled", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:        "Show me the top advertisers by total bids. Create a chart if possible.",
			Explore:       "bids_metrics",
			DisableCharts: true,
		})
		require.NoError(t, err)

		// Verify NO create_chart calls were made
		chartCalls := findChartCalls(s, res.Call.ID)
		require.Empty(t, chartCalls, "expected no create_chart calls when charts are disabled")

		// Verify we still got a response
		rawRes, err := s.UnmarshalMessageContent(res.Result)
		require.NoError(t, err)
		require.NotNil(t, rawRes)
	})
}

// findChartCalls finds all create_chart tool calls among children of the given parent call.
func findChartCalls(s *ai.Session, parentID string) []*ai.Message {
	return s.Messages(
		ai.FilterByParent(parentID),
		ai.FilterByType(ai.MessageTypeCall),
		ai.FilterByTool(ai.CreateChartName),
	)
}

// requireValidChartSpec unmarshals a create_chart call message, validates the chart spec structure,
// and returns the spec map for further assertions. If validChartTypes is non-empty, it asserts the
// chart_type is one of the allowed values.
func requireValidChartSpec(t *testing.T, s *ai.Session, chartCall *ai.Message, validChartTypes []string) map[string]any {
	t.Helper()

	rawArgs, err := s.UnmarshalMessageContent(chartCall)
	require.NoError(t, err)

	// CreateChartArgs is a named type (map[string]any), so try both type assertions
	var args map[string]any
	switch v := rawArgs.(type) {
	case ai.CreateChartArgs:
		args = map[string]any(v)
	case map[string]any:
		args = v
	default:
		require.Fail(t, "expected CreateChartArgs or map[string]any, got %T", rawArgs)
	}

	chartType, ok := args["chart_type"].(string)
	require.True(t, ok, "chart_type must be a string")
	require.NotEmpty(t, chartType)
	if len(validChartTypes) > 0 {
		require.Contains(t, validChartTypes, chartType, "chart_type %q not in expected types %v", chartType, validChartTypes)
	}

	spec, ok := args["spec"].(map[string]any)
	require.True(t, ok, "spec must be an object")

	metricsView, ok := spec["metrics_view"].(string)
	require.True(t, ok, "spec.metrics_view must be a string")
	require.NotEmpty(t, metricsView)

	_, hasTimeRange := spec["time_range"]
	require.True(t, hasTimeRange, "spec must contain time_range")

	return spec
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
