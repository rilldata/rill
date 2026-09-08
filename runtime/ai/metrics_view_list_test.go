package ai_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestListMetricsViewsAIInstructions verifies that the project's ai_instructions are returned
// to external MCP clients, but not to Rill's own agents (which receive them directly in their prompts).
func TestListMetricsViewsAIInstructions(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": `
ai_instructions: |
  Revenue always refers to net revenue.
`,
			"models/orders.yaml": `
type: model
materialize: true
sql: SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 100 AS revenue
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
measures:
- name: revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	newSessionWithUserAgent := func(t *testing.T, userAgent string) *ai.Session {
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

	// External MCP client: ai_instructions is included
	s := newSessionWithUserAgent(t, "mcp-client")
	var res *ai.ListMetricsViewsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.Contains(t, res.AIInstructions, "Revenue always refers to net revenue.")
	require.Len(t, res.MetricsViews, 1)

	// Rill's own agents: no ai_instructions (they are injected into agent prompts instead)
	s = newSessionWithUserAgent(t, "rill-web")
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.Empty(t, res.AIInstructions)
	require.Len(t, res.MetricsViews, 1)
}
