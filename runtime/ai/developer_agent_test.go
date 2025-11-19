package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestDeveloperShopify(t *testing.T) {
	// Setup a basic empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
		Files: map[string]string{
			"rill.yaml": `
olap_connector: duckdb
`,
			"connectors/duckdb.yaml": `
type: connector
driver: duckdb
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Initialize eval
	s := newEval(t, rt, instanceID)

	// Ask it to add a Shopify dashboard
	var res *ai.RouterAgentResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt: "Develop a dashboard of Shopify orders",
	})
	require.NoError(t, err)
	require.Equal(t, "developer_agent", res.Agent)

	// Verify it created a Shopify orders model
	msg, ok := s.LatestMessage(
		ai.FilterByTool("develop_model"),
		ai.FilterByType(ai.MessageTypeResult),
	)
	require.True(t, ok)
	require.Equal(t, `{"model_name":"shopify_orders"}`, msg.Content)

	// Check that it added three new resources without errors (model, metrics view, explore)
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "shopify_orders")
	require.NotEmpty(t, model.GetModel().State.ResultTable)

	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "shopify_orders")
	require.NotEmpty(t, mv.GetMetricsView().State.ValidSpec)
}
