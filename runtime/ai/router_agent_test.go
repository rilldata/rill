package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestRouterAgent(t *testing.T) {
	// Setup empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
	})

	// NOTE: We use a single eval session, so each subsequent prompt will have the context of the previous ones.
	s := newEval(t, rt, instanceID)
	cases := []struct {
		prompt string
		agent  string
	}{
		{
			prompt: "What country has the highest revenue?",
			agent:  ai.AnalystAgentName,
		},
		{
			prompt: "Repeat the answer you gave to my last question",
			agent:  ai.AnalystAgentName,
		},
		{
			prompt: "Create a model called 'sales_data' that selects all columns from the 'orders' table.",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "Do another one for the 'customers' table.",
			agent:  ai.DeveloperAgentName,
		},
		{
			prompt: "What is 2 + 2?",
			agent:  ai.AnalystAgentName,
		},
	}
	for _, c := range cases {
		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
			Prompt:      c.prompt,
			SkipHandoff: true,
		})
		require.NoError(t, err)
		require.Equal(t, c.agent, res.Agent, "prompt: %q", c.prompt)
	}
}
