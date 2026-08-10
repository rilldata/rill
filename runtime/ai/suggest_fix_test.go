package ai

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestResolveSuggestFixProposal(t *testing.T) {
	metricsViews := []*runtimev1.Resource{
		{
			Meta: &runtimev1.ResourceMeta{
				Name:      &runtimev1.ResourceName{Kind: "rill.runtime.v1.MetricsView", Name: "sales"},
				FilePaths: []string{"/metrics/sales.yaml"},
			},
		},
	}

	t.Run("project instruction targets rill.yaml", func(t *testing.T) {
		res, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:          SuggestFixActionAddProjectInstruction,
			Summary:         "s",
			InstructionText: "Revenue means net_revenue.",
			EvalQuestion:    "What was revenue in 2024?",
			EvalExpected:    "Reports net_revenue.",
		}, metricsViews)
		require.NoError(t, err)
		require.Equal(t, "/rill.yaml", res.FilePath)
		require.Equal(t, "Revenue means net_revenue.", res.Instruction)
		require.Equal(t, "What was revenue in 2024?", res.EvalQuestion)
		require.Equal(t, "Reports net_revenue.", res.EvalExpectedAnswer)
	})

	t.Run("metrics view instruction targets its file", func(t *testing.T) {
		res, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:          SuggestFixActionAddMetricsViewInstruction,
			MetricsView:     "Sales",
			InstructionText: "abc refers to net_revenue.",
		}, metricsViews)
		require.NoError(t, err)
		require.Equal(t, "/metrics/sales.yaml", res.FilePath)
		require.Equal(t, "abc refers to net_revenue.", res.Instruction)
	})

	t.Run("none still carries the eval draft", func(t *testing.T) {
		res, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:       SuggestFixActionNone,
			Summary:      "Nothing to change.",
			EvalQuestion: "q",
			EvalExpected: "a",
		}, metricsViews)
		require.NoError(t, err)
		require.Empty(t, res.FilePath)
		require.Empty(t, res.Instruction)
		require.Equal(t, "q", res.EvalQuestion)
	})

	t.Run("measure proposal targets the metrics view file", func(t *testing.T) {
		res, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:      SuggestFixActionAddMeasure,
			MetricsView: "sales",
			MeasureYAML: "name: avg_order_value\nexpression: SUM(revenue) / COUNT(*)\ndescription: Average order value.",
		}, metricsViews)
		require.NoError(t, err)
		require.Equal(t, "/metrics/sales.yaml", res.FilePath)
		require.Contains(t, res.MeasureYAML, "avg_order_value")
		require.Empty(t, res.Instruction)
	})

	t.Run("errors on a measure without a name or expression", func(t *testing.T) {
		_, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:      SuggestFixActionAddMeasure,
			MetricsView: "sales",
			MeasureYAML: "name: avg_order_value",
		}, metricsViews)
		require.ErrorContains(t, err, "must include a name and an expression")
	})

	t.Run("errors on a malformed measure definition", func(t *testing.T) {
		_, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:      SuggestFixActionAddMeasure,
			MetricsView: "sales",
			MeasureYAML: "- not\n- a\n- mapping",
		}, metricsViews)
		require.ErrorContains(t, err, "not valid YAML")
	})

	t.Run("errors on missing instruction text", func(t *testing.T) {
		_, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action: SuggestFixActionAddProjectInstruction,
		}, metricsViews)
		require.ErrorContains(t, err, "missing instruction text")
	})

	t.Run("errors on unknown metrics view", func(t *testing.T) {
		_, err := resolveSuggestFixProposal(&feedbackFixProposal{
			Action:          SuggestFixActionAddMetricsViewInstruction,
			MetricsView:     "nope",
			InstructionText: "rule",
		}, metricsViews)
		require.ErrorContains(t, err, `unknown metrics view "nope"`)
	})

	t.Run("errors on unknown action", func(t *testing.T) {
		_, err := resolveSuggestFixProposal(&feedbackFixProposal{Action: "edit_measure_description"}, metricsViews)
		require.ErrorContains(t, err, "unknown action")
	})
}

func TestParseSuggestFixProposal(t *testing.T) {
	msg := NewTextCompletionMessage(RoleAssistant, `{"action": "add_metrics_view_instruction", "summary": "Add an alias.", "metrics_view": "sales", "instruction_text": "abc refers to net_revenue."}`)
	proposal, err := parseSuggestFixProposal(msg)
	require.NoError(t, err)
	require.Equal(t, SuggestFixActionAddMetricsViewInstruction, proposal.Action)
	require.Equal(t, "sales", proposal.MetricsView)

	// Tolerates markdown fences
	msg = NewTextCompletionMessage(RoleAssistant, "```json\n{\"action\": \"none\", \"summary\": \"s\"}\n```")
	proposal, err = parseSuggestFixProposal(msg)
	require.NoError(t, err)
	require.Equal(t, SuggestFixActionNone, proposal.Action)
}
