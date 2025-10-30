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
			agent:  "analyst_agent",
		},
		{
			prompt: "Repeat the answer you gave to my last question",
			agent:  "analyst_agent",
		},
		{
			prompt: "Create a model called 'sales_data' that selects all columns from the 'orders' table.",
			agent:  "developer_agent",
		},
		{
			prompt: "Do another one for the 'customers' table.",
			agent:  "developer_agent",
		},
		{
			prompt: "What is 2 + 2?",
			agent:  "analyst_agent",
		},
	}
	for _, c := range cases {
		var res *ai.RouterAgentResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
			Prompt:      c.prompt,
			SkipHandoff: true,
		})
		require.NoError(t, err)
		require.Equal(t, c.agent, res.Agent, "prompt: %q", c.prompt)
	}
}
