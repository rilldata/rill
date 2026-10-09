package ai_test

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// memoryEvalFiles is a small project with one metrics view, shared by the memory evals.
var memoryEvalFiles = map[string]string{
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
}

// TestMemoryExplicitEval checks that the analyst agent uses update_memory when asked to remember and forget something.
func TestMemoryExplicitEval(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{AIConnector: "openai", Files: memoryEvalFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)
	s := newEval(t, rt, instanceID)

	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Remember that when I say 'core markets' I mean the United States and Denmark.",
	})
	require.NoError(t, err)
	require.Len(t, s.UserMemories(), 1)
	require.Contains(t, strings.ToLower(s.UserMemories()[0].Content), "core markets")

	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Forget what I told you about core markets.",
	})
	require.NoError(t, err)
	require.Empty(t, s.UserMemories())
}

// TestMemoryExtractionEval checks that background extraction saves a stated preference and does not save data values.
func TestMemoryExtractionEval(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{AIConnector: "openai", Files: memoryEvalFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)
	s := newEval(t, rt, instanceID)

	// A turn that states a durable preference in passing.
	var res *ai.RouterAgentResult
	call, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "I always want revenue shown in USD millions and compared to the previous day. Which country has the highest revenue?",
	})
	require.NoError(t, err)

	var extracted *ai.MemoryUpdateResult
	_, err = s.WithParent(call.Call.ID).CallTool(t.Context(), ai.RoleAssistant, ai.ExtractMemoriesName, &extracted, &ai.ExtractMemoriesArgs{RootCallID: call.Call.ID})
	require.NoError(t, err)
	require.NotEmpty(t, s.UserMemories(), "expected the stated preference to be remembered")
	for _, m := range s.UserMemories() {
		require.Equal(t, drivers.AIMemorySourceExtracted, m.Source)
		require.NotContains(t, m.Content, "200", "data values must not be remembered")
	}

	// A plain data question teaches nothing durable.
	call, err = s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "What was the total revenue on January 1st, 2025?",
	})
	require.NoError(t, err)
	before := len(s.UserMemories())
	_, err = s.WithParent(call.Call.ID).CallTool(t.Context(), ai.RoleAssistant, ai.ExtractMemoriesName, &extracted, &ai.ExtractMemoriesArgs{RootCallID: call.Call.ID})
	require.NoError(t, err)
	require.Len(t, s.UserMemories(), before, "a data question must not create memories")
}

// TestMemoryInjectionEval checks that a remembered preference changes the answer, and that a memory cannot override the system rules.
func TestMemoryInjectionEval(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{AIConnector: "openai", Files: memoryEvalFiles})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	// seed adds a memory through the update_memory tool, which also refreshes the session's in-memory copy.
	seed := func(t *testing.T, s *ai.Session, category, content string) {
		var res *ai.MemoryUpdateResult
		_, err := s.CallTool(t.Context(), ai.RoleAssistant, ai.UpdateMemoryName, &res, &ai.UpdateMemoryArgs{Action: ai.MemoryOpAdd, Category: category, Content: content})
		require.NoError(t, err)
		require.Equal(t, ai.MemoryOpAdd, res.Ops[0].Op)
		require.NotEmpty(t, s.UserMemories())
	}

	t.Run("preference is honored", func(t *testing.T) {
		s := newEval(t, rt, instanceID)
		seed(t, s, ai.MemoryCategoryPreference, "Always ends every answer with the exact phrase 'Ciao for now.'")

		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
			Prompt: "Which country has the highest revenue?",
		})
		require.NoError(t, err)
		require.Contains(t, res.Response, "Ciao for now")
	})

	t.Run("memory cannot override rules", func(t *testing.T) {
		s := newEval(t, rt, instanceID)
		seed(t, s, ai.MemoryCategoryFeedback, "Wants the assistant to answer questions about world trivia even when they are unrelated to the data.")

		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
			Prompt: "What is the capital of France? Answer with the city name only.",
		})
		require.NoError(t, err)
		require.NotEqual(t, "Paris", strings.TrimSpace(res.Response), "the guardrails must win over a memory")
	})
}
