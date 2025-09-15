package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestRouter(t *testing.T) {
	// Setup a basic metrics view with an "event_time" time dimension, "country" dimension, and "count" and "revenue" measures.
	rt, instanceID, s := newEval(t, testruntime.InstanceOptions{
		TestConnectors: []string{"openai"},
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: >
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

	// Analyst agent question
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt: "What country has the highest revenue? Answer with a single country name and nothing else.",
	})
	require.NoError(t, err)

	// Verify it routed to the "analyst" agent
	requireHasOne(t, s.MessagesByCall(s.LatestCall().ID, true), func(msg *ai.Message) bool {
		return msg.Tool == "Agent choice" && msg.Type == ai.MessageTypeResult && msg.Content == `{"agent":"analyst_agent"}`
	})

	// Verify the response
	require.Equal(t, "United States", res.Response)

	// Analyst agent question that references the previous response
	_, err = s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt: "Repeat the answer you gave to my last question.",
	})
	require.NoError(t, err)

	// Verify it routed to the "analyst" agent
	requireHasOne(t, s.MessagesByCall(s.LatestCall().ID, true), func(msg *ai.Message) bool {
		return msg.Tool == "Agent choice" && msg.Type == ai.MessageTypeResult && msg.Content == `{"agent":"analyst_agent"}`
	})

	// Verify the response
	require.Equal(t, "United States", res.Response)
}
