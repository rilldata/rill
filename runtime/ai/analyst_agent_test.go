package ai_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// newSessionWithHeaders creates a test AI session with optional headers.
func newSessionWithHeaders(t *testing.T, rt *runtime.Runtime, instanceID string, userAgent string, headers http.Header) *ai.Session {
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     claims,
		UserAgent:  userAgent,
		Headers:    headers,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Flush(t.Context())
		require.NoError(t, err)
	})

	return s
}

func TestAnalystBasic(t *testing.T) {
	// Setup a basic metrics view with an "event_time" time dimension, "country" dimension, and "count" and "revenue" measures.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
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
	_, err := s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt: "What country has the highest revenue? Answer with a single country name and nothing else.",
	})
	require.NoError(t, err)
	require.Equal(t, "analyst_agent", res.Agent)
	require.Equal(t, "United States", res.Response)

	// Analyst agent question that references the previous response
	_, err = s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt: "Repeat the answer you gave to my last question.",
	})
	require.NoError(t, err)
	require.Equal(t, "analyst_agent", res.Agent)
	require.Equal(t, "United States", res.Response)
}

func TestAnalystOpenRTB(t *testing.T) {
	// Setup runtime instance with the OpenRTB dataset
	n, files := testruntime.ProjectOpenRTB(t)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
		Files:     files,
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
		require.Equal(t, "get_metrics_view", calls[0].Tool)

		// It should make two sub-calls: get_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Now tell me about the other dataset (but don't query the data)",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 1)
		require.Equal(t, "get_metrics_view", calls[0].Tool)

		// It should remember the previous turns and only make one sub-call: query_metrics_view_summary and query_metrics_view
		res, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.RouterAgentArgs{
			Prompt: "Tell me which non-US country has the most auctions",
		})
		require.NoError(t, err)
		calls = s.Messages(ai.FilterByParent(res.Call.ID), ai.FilterByType(ai.MessageTypeCall))
		require.Len(t, calls, 2)
		require.Equal(t, "query_metrics_view_summary", calls[0].Tool)
		require.Equal(t, "query_metrics_view", calls[1].Tool)
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
		require.Equal(t, "query_metrics_view_summary", calls[0].Tool)
		require.Equal(t, "get_metrics_view", calls[1].Tool)
		require.Equal(t, "query_metrics_view", calls[2].Tool)

		// Map the request sent and assert that context was honored.
		rawQry, err := s.UnmarshalMessageContent(calls[2])
		require.NoError(t, err)
		var qry metricsview.Query
		err = mapstructureutil.WeakDecode(rawQry, &qry)
		require.NoError(t, err)
		// Assert that the time range is sent using the context
		require.Equal(t, parseTestTime(t, "2023-09-11T00:00:00Z"), qry.TimeRange.Start)
		require.Equal(t, parseTestTime(t, "2023-09-14T00:00:00Z"), qry.TimeRange.End)
		// Assert that the filter is sent using the context
		require.NotNil(t, qry.Where)
		require.NotNil(t, qry.Where.Condition)
		require.Len(t, qry.Where.Condition.Expressions, 2)
		require.Equal(t, "device_os", qry.Where.Condition.Expressions[0].Name)
		require.Equal(t, "Android", qry.Where.Condition.Expressions[1].Value)
	})

}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}

func TestAnalystCheckAccess(t *testing.T) {
	// Setup runtime instance with the OpenRTB dataset
	n, files := testruntime.ProjectOpenRTB(t)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
		Files:     files,
	})
	testruntime.RequireReconcileState(t, rt, instanceID, n, 0, 0)

	t.Run("CheckAccess Denied - non-rill user agent", func(t *testing.T) {
		s := newSession(t, rt, instanceID)
		s.CatalogSession().UserAgent = "some-other-agent/1.0"

		agent := &ai.AnalystAgent{Runtime: rt}

		require.False(t, agent.CheckAccess(ai.WithSession(t.Context(), s)))
	})

	t.Run("CheckAccess Denied - no X-Rill-Agent header and non-rill user agent", func(t *testing.T) {
		s := newSession(t, rt, instanceID)
		s.CatalogSession().UserAgent = "some-other-agent/1.0"

		agent := &ai.AnalystAgent{Runtime: rt}

		ctx := ai.WithSession(t.Context(), s)
		require.False(t, agent.CheckAccess(ctx))
	})

	t.Run("CheckAccess Allowed - rill user agent", func(t *testing.T) {
		s := newSession(t, rt, instanceID)

		s.CatalogSession().UserAgent = "rill-evals/1.0"

		agent := &ai.AnalystAgent{Runtime: rt}

		ctx := ai.WithSession(t.Context(), s)
		require.True(t, agent.CheckAccess(ctx))
	})

	t.Run("CheckAccess Allowed - X-Rill-Agent header", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("X-Rill-Agent", "gemini")

		s := newSessionWithHeaders(t, rt, instanceID, "node", headers)

		agent := &ai.AnalystAgent{Runtime: rt}

		ctx := ai.WithSession(t.Context(), s)
		require.True(t, agent.CheckAccess(ctx))
	})
}
