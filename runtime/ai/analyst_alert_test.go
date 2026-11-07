package ai_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestAlertVerdictUnmarshal(t *testing.T) {
	var v ai.AlertVerdict
	err := json.Unmarshal([]byte(`{"met": false, "reason": "Revenue is within its usual range."}`), &v)
	require.NoError(t, err)
	require.Equal(t, ai.AlertVerdict{Met: false, Reason: "Revenue is within its usual range."}, v)

	// A malformed verdict is an error, not a verdict that the condition is not met
	for _, in := range []string{
		`The condition holds.`,
		`{"reason": "Revenue fell."}`,
		`{"met": null, "reason": "Revenue fell."}`,
		`{"met": "yes", "reason": "Revenue fell."}`,
		`{"met": true}`,
		`{"met": true, "reason": "  "}`,
	} {
		err := json.Unmarshal([]byte(in), &v)
		require.Error(t, err, in)
	}
}

func TestAnalystAlert(t *testing.T) {
	alertArgs := &ai.RouterAgentArgs{
		Prompt:           "There are orders from Denmark",
		Agent:            ai.AnalystAgentName,
		AnalystAgentArgs: &ai.AnalystAgentArgs{IsAlert: true},
	}

	t.Run("verdict", func(t *testing.T) {
		var outputSchema *jsonschema.Schema
		script := &scriptedAIService{turns: []turnFunc{func(opts *drivers.CompleteOptions) *aiv1.CompletionMessage {
			outputSchema = opts.OutputSchema
			return textTurn(`{"met": true, "reason": "Denmark has orders."}`)(opts)
		}}}
		s := newSkillReferencesSession(t, script)

		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, alertArgs)
		require.NoError(t, err)
		require.Equal(t, &ai.AlertVerdict{Met: true, Reason: "Denmark has orders."}, res.Verdict)

		// The verdict is requested through the output schema
		require.NotNil(t, outputSchema)
		require.Contains(t, outputSchema.Properties, "met")
		require.Contains(t, outputSchema.Properties, "reason")

		// A verdict can't show charts
		require.Contains(t, script.tools[0], ai.QueryMetricsViewName)
		require.NotContains(t, script.tools[0], ai.CreateChartName)
	})

	t.Run("malformed verdict", func(t *testing.T) {
		s := newSkillReferencesSession(t, &scriptedAIService{turns: []turnFunc{textTurn(`The condition holds.`)}})
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, nil, alertArgs)
		require.Error(t, err)
	})

	t.Run("iteration limit", func(t *testing.T) {
		// Outside of alerts, the loop forces a final answer without tools after the last iteration.
		// For an alert, that would be a verdict that doesn't rest on the data, so it fails instead.
		turns := slices.Repeat([]turnFunc{toolCallTurn(ai.ListMetricsViewsName, map[string]any{})}, 20)
		s := newSkillReferencesSession(t, &scriptedAIService{turns: turns})
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, nil, alertArgs)
		require.ErrorContains(t, err, "iterations")
	})
}

func TestAnalystAlertEval(t *testing.T) {
	// Eight weeks of daily revenue for three countries.
	// Revenue is around 100 a day for every country, except Spain, which drops to around 30 a day in the last week.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT
    d::TIMESTAMP AS event_time,
    c AS country,
    CASE WHEN c = 'Spain' AND d >= DATE '2025-10-19' THEN 30 ELSE 100 END + (hash(d::DATE, c) % 7)::INTEGER AS revenue
  FROM range(DATE '2025-08-31', DATE '2025-10-26', INTERVAL 1 DAY) AS t(d), (VALUES ('Spain'), ('France'), ('Germany')) AS countries(c)
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
dimensions:
- column: country
measures:
- name: revenue
  expression: SUM(revenue)
`,
			"explores/orders.yaml": `
type: explore
metrics_view: orders
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	cases := []struct {
		condition string
		met       bool
	}{
		{"Revenue in Spain dropped sharply in the last week compared with the previous weeks", true},
		{"Spain is the country with the lowest revenue in the last week", true},
		{"Revenue in France dropped sharply in the last week compared with the previous weeks", false},
		{"There are orders from Italy", false},
	}
	for _, tc := range cases {
		t.Run(tc.condition, func(t *testing.T) {
			s := newEval(t, rt, instanceID)
			var res *ai.RouterAgentResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, &ai.RouterAgentArgs{
				Prompt:           tc.condition,
				Agent:            ai.AnalystAgentName,
				AnalystAgentArgs: &ai.AnalystAgentArgs{Explore: "orders", IsAlert: true},
			})
			require.NoError(t, err)
			require.NotNil(t, res.Verdict)
			require.Equal(t, tc.met, res.Verdict.Met, res.Verdict.Reason)
		})
	}
}
