package ai_test

import (
	"strings"
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
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt: "Develop a dashboard of Shopify orders using mock data. Please proceed without asking clarifying questions.",
	})
	require.NoError(t, err)
	require.Equal(t, ai.DeveloperAgentName, res.Agent)

	// Verify it created a Shopify orders model
	msg, ok := s.LatestMessage(
		ai.FilterByTool(ai.DevelopFileName),
		ai.FilterByType(ai.MessageTypeCall),
	)
	require.True(t, ok)
	args := s.MustUnmarshalMessageContent(msg).(*ai.DevelopFileArgs)
	require.Contains(t, []string{"explore", "canvas", "metrics_view"}, args.Type)

	// Check that it doesn't have any parse or reconcile errors.
	testruntime.RequireReconcileState(t, rt, instanceID, -1, 0, 0)

	// Check there's a model and metrics view created related to shopify
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	models, err := ctrl.List(t.Context(), runtime.ResourceKindModel, "", false)
	require.NoError(t, err)
	metricsViews, err := ctrl.List(t.Context(), runtime.ResourceKindMetricsView, "", false)
	require.NoError(t, err)

	foundModel := false
	for _, m := range models {
		if strings.Contains(m.Meta.Name.Name, "shopify") {
			foundModel = true
			break
		}
	}
	require.True(t, foundModel, "expected a model related to shopify")

	foundMV := false
	for _, mv := range metricsViews {
		if strings.Contains(mv.Meta.Name.Name, "shopify") {
			foundMV = true
			break
		}
	}
	require.True(t, foundMV, "expected a metrics view related to shopify")
}
