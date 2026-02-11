package ai_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
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
		FrontendURL: "http://localhost:8080",
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

		// It should make two sub-calls: get_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Tell me about the auction metrics (but don't query the data)",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)
		require.Equal(t, ai.GetMetricsViewName, calls[0].Tool)

		// It should make two sub-calls: get_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Now tell me about the other dataset (but don't query the data)",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)
		require.Equal(t, ai.GetMetricsViewName, calls[0].Tool)

		// It should remember the previous turns and only make one sub-call: query_metrics_view_summary and query_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Tell me which non-US country has the most auctions",
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
			Prompt:    "Tell me which app_site_name has the most impressions",
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

		assertQuery := func(qry metricsview.Query) {
			// Assert that the time range is sent using the context
			require.Equal(t, parseTestTime(t, "2023-09-11T00:00:00Z"), qry.TimeRange.Start)
			require.Equal(t, parseTestTime(t, "2023-09-14T00:00:00Z"), qry.TimeRange.End)
			// Assert that the filter is sent using the context
			require.NotNil(t, qry.Where)
			require.NotNil(t, qry.Where.Condition)
			expr := maybeUnwrapJoinerExpr(qry.Where)
			require.Len(t, expr.Condition.Expressions, 2)
			require.Equal(t, "device_os", expr.Condition.Expressions[0].Name)
			require.Equal(t, "Android", expr.Condition.Expressions[1].Value)
		}

		// Map the request sent and assert that context was honored.
		rawQry, err := s.UnmarshalMessageContent(calls[2])
		require.NoError(t, err)
		var qry metricsview.Query
		err = mapstructureutil.WeakDecode(rawQry, &qry)
		require.NoError(t, err)
		assertQuery(qry)

		// Get query in the citation url in the final summary message and assert that context was honored.
		citationUrlQuery := agentMessageToMetricsResolverQuery(t, s, res.Result)
		assertQuery(citationUrlQuery)
	})

	t.Run("CanvasContext", func(t *testing.T) {
		s := newEval(t, rt, instanceID)

		// It should make three sub-calls: query_metrics_view_summary, get_metrics_view, query_metrics_view
		res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{
			Prompt:          "Tell me which advertiser_name has the highest overall_spend",
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

		assertQuery := func(qry metricsview.Query) {
			// Assert that the time range is sent using the context
			require.Equal(t, parseTestTime(t, "2023-09-11T00:00:00Z"), qry.TimeRange.Start)
			require.Equal(t, parseTestTime(t, "2023-09-14T00:00:00Z"), qry.TimeRange.End)
			// Assert that the filter is sent using the context
			require.NotNil(t, qry.Where)
			require.NotNil(t, qry.Where.Condition)
			expr := maybeUnwrapJoinerExpr(qry.Where)
			require.Len(t, expr.Condition.Expressions, 2)

			auctionTypeCond := expr.Condition.Expressions[0]
			appOrSiteCond := expr.Condition.Expressions[1]
			if auctionTypeCond.Condition.Expressions[0].Name != "auction_type" {
				auctionTypeCond = expr.Condition.Expressions[1]
				appOrSiteCond = expr.Condition.Expressions[0]
			}

			require.Equal(t, "auction_type", auctionTypeCond.Condition.Expressions[0].Name)
			require.Equal(t, "First Price", auctionTypeCond.Condition.Expressions[1].Value)
			require.Equal(t, "app_or_site", appOrSiteCond.Condition.Expressions[0].Name)
			require.Equal(t, "App", appOrSiteCond.Condition.Expressions[1].Value)
		}

		// Map the request sent and assert that context was honored.
		rawQry, err := s.UnmarshalMessageContent(calls[3])
		require.NoError(t, err)
		var qry metricsview.Query
		err = mapstructureutil.WeakDecode(rawQry, &qry)
		require.NoError(t, err)
		assertQuery(qry)

		// Get query in the citation url in the final summary message and assert that context was honored.
		citationUrlQuery := agentMessageToMetricsResolverQuery(t, s, res.Result)
		assertQuery(citationUrlQuery)
	})
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}

var citationUrlRegex = regexp.MustCompile(`open-query\?query=(.*?)\)`)

func agentMessageToMetricsResolverQuery(t *testing.T, s *ai.Session, msg *ai.Message) metricsview.Query {
	rawRes, err := s.UnmarshalMessageContent(msg)
	require.NoError(t, err)
	var analystRes ai.AnalystAgentResult
	err = mapstructureutil.WeakDecode(rawRes, &analystRes)
	require.NoError(t, err)

	matches := citationUrlRegex.FindStringSubmatch(analystRes.Response)
	require.Len(t, matches, 2)
	queryParam, err := url.QueryUnescape(matches[1])
	require.NoError(t, err)
	var qry metricsview.Query
	err = json.Unmarshal([]byte(queryParam), &qry)
	if err != nil {
		fmt.Println(queryParam)
	}
	require.NoError(t, err)
	return qry
}

func maybeUnwrapJoinerExpr(expr *metricsview.Expression) *metricsview.Expression {
	if expr.Condition == nil ||
		(expr.Condition.Operator != metricsview.OperatorOr && expr.Condition.Operator != metricsview.OperatorAnd) ||
		len(expr.Condition.Expressions) != 1 {
		return expr
	}
	return expr.Condition.Expressions[0]
}
