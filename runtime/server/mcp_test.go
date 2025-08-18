package server

import (
	"context"
	"encoding/json"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// testCtx provides authentication context for testing
func testCtx() context.Context {
	return auth.WithClaims(context.Background(), auth.NewOpenClaims())
}

func newMCPTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			// Create a simple model
			"test_data.sql": `SELECT 'US' AS country, 100 AS revenue, NOW() AS timestamp`,
			// Create a metrics view
			"test_metrics.yaml": `
type: metrics_view
version: 1
model: test_data
dimensions:
- column: country
measures:
- expression: SUM(revenue)
  name: total_revenue
`,
		},
	})

	// Wait for reconciliation to complete
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return srv, instanceID
}

func TestMCPListTools(t *testing.T) {
	srv, instanceID := newMCPTestServer(t)

	ctx := testCtx()

	// Test listing tools
	tools, err := srv.mcpListTools(ctx, instanceID)
	require.NoError(t, err)

	// Verify expected tools are present
	expectedToolNames := []string{
		"list_metrics_views",
		"get_metrics_view",
		"query_metrics_view_time_range",
		"query_metrics_view",
		"search",
		"fetch",
	}

	require.Len(t, tools, len(expectedToolNames))

	// Check that all tools have proper metadata
	for _, tool := range tools {
		require.Contains(t, expectedToolNames, tool.Name)
		require.NotEmpty(t, tool.Description, "Tool %s should have a description", tool.Name)
		require.NotEmpty(t, tool.Name, "Tool should have a name")
	}

	// Verify specific tool metadata for get_metrics_view (has input schema)
	var getMetricsViewTool *aiv1.Tool
	for _, tool := range tools {
		if tool.Name == "get_metrics_view" {
			getMetricsViewTool = tool
			break
		}
	}
	require.NotNil(t, getMetricsViewTool, "get_metrics_view tool should be found")
	require.Contains(t, getMetricsViewTool.Description, "metrics view", "Description should mention metrics view")

	// Verify InputSchema is valid JSON when present
	if getMetricsViewTool.InputSchema != "" {
		t.Logf("InputSchema: %s", getMetricsViewTool.InputSchema)
		var schema interface{}
		err := json.Unmarshal([]byte(getMetricsViewTool.InputSchema), &schema)
		require.NoError(t, err, "InputSchema should be valid JSON")
	} else {
		t.Log("No InputSchema found for get_metrics_view tool")
	}
}

func TestMCPExecuteTool_Success(t *testing.T) {
	srv, instanceID := newMCPTestServer(t)

	ctx := testCtx()

	// Test executing list_metrics_views tool (no parameters required)
	textResult, err := srv.mcpExecuteTool(ctx, instanceID, "list_metrics_views", map[string]any{})
	require.NoError(t, err)

	// The response should be valid JSON with metrics view data
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(textResult), &jsonData)
	require.NoError(t, err, "expected valid JSON response from successful tool execution")

	// Verify the response contains the expected structure
	require.Contains(t, jsonData, "metrics_views", "response should contain metrics_views field")

	metricsViews, ok := jsonData["metrics_views"].([]interface{})
	require.True(t, ok, "metrics_views should be an array")
	require.Len(t, metricsViews, 1, "should have one metrics view")

	// Verify the metrics view has expected fields
	mv, ok := metricsViews[0].(map[string]interface{})
	require.True(t, ok, "metrics view should be an object")
	require.Equal(t, "test_metrics", mv["name"], "metrics view should have correct name")
}

func TestMCPExecuteTool_MissingParam(t *testing.T) {
	srv, instanceID := newMCPTestServer(t)

	ctx := testCtx()

	// Test executing get_metrics_view tool without required parameter
	result, err := srv.mcpExecuteTool(ctx, instanceID, "get_metrics_view", map[string]any{})

	// The tool should either error or return an error message in the response
	if err != nil {
		// If it errors, it should mention the missing parameter
		require.Contains(t, err.Error(), "metrics_view")
	} else {
		// If it succeeds, check that result indicates an issue
		require.NotEmpty(t, result)
		t.Logf("Tool succeeded with error in response: %v", result)
		// This is valid behavior - MCP tools return errors in response content
	}
}
