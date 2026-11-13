package ai_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// Queries and answers for a day-over-day question on the orders metrics view of newOrdersSession.
var (
	revenueTimeSeriesQuery = map[string]any{
		"metrics_view": "orders",
		"dimensions": []any{
			map[string]any{"name": "order_date", "compute": map[string]any{"time_floor": map[string]any{"dimension": "order_date", "grain": "day"}}},
		},
		"measures":   []any{map[string]any{"name": "revenue"}},
		"time_range": map[string]any{"start": "2025-03-10T00:00:00Z", "end": "2025-03-12T00:00:00Z"},
		"sort":       []any{map[string]any{"name": "order_date", "desc": true}},
		"limit":      10,
	}
	revenueComparisonQuery = map[string]any{
		"metrics_view": "orders",
		"measures": []any{
			map[string]any{"name": "revenue"},
			map[string]any{"name": "revenue__delta_abs", "compute": map[string]any{"comparison_delta": map[string]any{"measure": "revenue"}}},
			map[string]any{"name": "revenue__delta_rel", "compute": map[string]any{"comparison_ratio": map[string]any{"measure": "revenue"}}},
		},
		"time_range":            map[string]any{"start": "2025-03-11T00:00:00Z", "end": "2025-03-12T00:00:00Z"},
		"comparison_time_range": map[string]any{"start": "2025-03-10T00:00:00Z", "end": "2025-03-11T00:00:00Z"},
		"limit":                 10,
	}
	// An answer with a delta the model computed from the time series rows itself, and got wrong; the right delta is 31,559,455.
	wrongRevenueAnswer = "Revenue rose by +35,559,455 day over day, from 3,152,968,451 on Mar 10 to 3,184,527,906 on Mar 11."
	rightRevenueAnswer = "Revenue rose by +31,559,455 (1.00%) day over day, to 3,184,527,906 on Mar 11."
)

func TestAnalystRevisesUngroundedNumbers(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{
		toolCallTurn(ai.QueryMetricsViewName, revenueTimeSeriesQuery),
		textTurn(wrongRevenueAnswer),
		toolCallTurn(ai.QueryMetricsViewName, revenueComparisonQuery),
		textTurn(rightRevenueAnswer),
	}}
	s := newOrdersSession(t, script)

	var res *ai.AnalystAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, &res, &ai.AnalystAgentArgs{
		Prompt: "What's the day over day change in revenue?",
	})
	require.NoError(t, err)
	require.Equal(t, rightRevenueAnswer, res.Response)
	require.Equal(t, 4, script.calls)

	// The model was shown its rejected answer and told which numbers no tool result contains, and how to get them.
	input := script.inputs[2]
	require.Equal(t, wrongRevenueAnswer, input[len(input)-2].Content[0].GetText())
	feedback := input[len(input)-1]
	require.Equal(t, string(ai.RoleUser), feedback.Role)
	require.Contains(t, feedback.Content[0].GetText(), "+35,559,455")
	require.Contains(t, feedback.Content[0].GetText(), "comparison_delta")
	require.NotContains(t, feedback.Content[0].GetText(), "3,152,968,451")

	// The comparison query returns the right delta.
	results := s.Messages(ai.FilterByType(ai.MessageTypeResult), ai.FilterByTool(ai.QueryMetricsViewName))
	require.Len(t, results, 2)
	var comparison ai.QueryMetricsViewResult
	require.NoError(t, json.Unmarshal([]byte(results[1].Content), &comparison))
	require.Len(t, comparison.Data, 1)
	require.EqualValues(t, 3184527906, comparison.Data[0][0])
	require.EqualValues(t, 31559455, comparison.Data[0][1])
	require.InDelta(t, 0.010009, comparison.Data[0][2], 1e-6)

	// The rejected answer was never added to the session, so the user never saw it.
	for _, m := range s.Messages() {
		require.NotContains(t, m.Content, "35,559,455")
	}
}

func TestAnalystRevisesUngroundedNumbersOnce(t *testing.T) {
	// The model ignores the feedback, and its revision is returned as is.
	revised := strings.Replace(wrongRevenueAnswer, "rose", "increased", 1)
	script := &scriptedAIService{turns: []turnFunc{
		toolCallTurn(ai.QueryMetricsViewName, revenueTimeSeriesQuery),
		textTurn(wrongRevenueAnswer),
		textTurn(revised),
	}}
	s := newOrdersSession(t, script)

	var res *ai.AnalystAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, &res, &ai.AnalystAgentArgs{
		Prompt: "What's the day over day change in revenue?",
	})
	require.NoError(t, err)
	require.Equal(t, revised, res.Response)
	require.Equal(t, 3, script.calls)
}

func TestAnalystKeepsGroundedNumbers(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{
		toolCallTurn(ai.QueryMetricsViewName, revenueComparisonQuery),
		textTurn(rightRevenueAnswer),
	}}
	s := newOrdersSession(t, script)

	var res *ai.AnalystAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, &res, &ai.AnalystAgentArgs{
		Prompt: "What's the day over day change in revenue?",
	})
	require.NoError(t, err)
	require.Equal(t, rightRevenueAnswer, res.Response)
	require.Equal(t, 2, script.calls)
}

// newOrdersSession creates a session on a project with a daily orders metrics view, backed by the given simulated model.
func newOrdersSession(t *testing.T, script *scriptedAIService) *ai.Session {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT '2025-03-10T00:00:00Z'::TIMESTAMP AS order_date, 'web' AS channel, 3152968451::DOUBLE AS amount
  UNION ALL
  SELECT '2025-03-11T00:00:00Z'::TIMESTAMP AS order_date, 'web' AS channel, 3184527906::DOUBLE AS amount
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: order_date
dimensions:
- column: channel
measures:
- name: revenue
  expression: SUM(amount)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newSession(t, rt, instanceID)
	s.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
		return script, func() {}, nil
	})
	return s
}
