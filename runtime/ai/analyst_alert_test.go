package ai_test

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestAnalystAlert(t *testing.T) {
	alertArgs := &ai.RouterAgentArgs{
		Prompt:           "There are orders from Denmark",
		Agent:            ai.AnalystAgentName,
		AnalystAgentArgs: &ai.AnalystAgentArgs{IsAlert: true},
	}
	analysis := "<summary>Denmark has orders.</summary> Orders come from Denmark."

	cases := []struct {
		verdict string
		want    ai.AlertOutcome
		wantErr string
	}{
		{verdict: `{"outcome": "met"}`, want: ai.AlertOutcomeMet},
		{verdict: `{"outcome": "not_met"}`, want: ai.AlertOutcomeNotMet},
		{verdict: `{"outcome": "cannot_determine"}`, want: ai.AlertOutcomeCannotDetermine},
		{verdict: `{"outcome": "maybe"}`, wantErr: "invalid verdict"},
		{verdict: `The condition holds.`, wantErr: "verdict"},
	}
	for _, tc := range cases {
		t.Run(tc.verdict, func(t *testing.T) {
			var schemas []*jsonschema.Schema
			recordSchema := func(turn turnFunc) turnFunc {
				return func(opts *drivers.CompleteOptions) *aiv1.CompletionMessage {
					schemas = append(schemas, opts.OutputSchema)
					return turn(opts)
				}
			}
			script := &scriptedAIService{turns: []turnFunc{recordSchema(textTurn(analysis)), recordSchema(textTurn(tc.verdict))}}
			s := newSkillReferencesSession(t, script)

			var res *ai.RouterAgentResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, alertArgs)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, res.AlertOutcome)
			require.Equal(t, analysis, res.Response)

			// The analysis runs with tools and the verdict with an output schema, in separate requests,
			// since some providers reject a request that combines them.
			require.Len(t, schemas, 2)
			require.Contains(t, script.tools[0], ai.QueryMetricsViewName)
			require.Nil(t, schemas[0])
			require.Empty(t, script.tools[1])
			require.Contains(t, schemas[1].Properties, "outcome")
			// The verdict is given on the analysis
			require.Contains(t, script.inputs[1][len(script.inputs[1])-1].Content[0].GetText(), analysis)
		})
	}
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
		want      ai.AlertOutcome
	}{
		{"Revenue in Spain dropped sharply in the last week compared with the previous weeks", ai.AlertOutcomeMet},
		{"Spain is the country with the lowest revenue in the last week", ai.AlertOutcomeMet},
		{"Revenue in France dropped sharply in the last week compared with the previous weeks", ai.AlertOutcomeNotMet},
		{"There are orders from Italy", ai.AlertOutcomeNotMet},
		// The data has no profit
		{"The profit margin in Spain fell below 10% in the last week", ai.AlertOutcomeCannotDetermine},
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
			require.Equal(t, tc.want, res.AlertOutcome, res.Response)
		})
	}
}
