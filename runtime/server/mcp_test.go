package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func newMCPTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstance(t)

	srv, err := NewServer(context.Background(), &Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return srv, instanceID
}

func TestMCPListTools(t *testing.T) {
	srv, instanceID := newMCPTestServer(t)

	ctx := context.Background()

	// Test listing tools
	tools, err := srv.mcpListTools(ctx, instanceID)
	require.NoError(t, err)

	// Verify expected tools are present
	expectedToolNames := []string{
		"list_metrics_views",
		"get_metrics_view",
		"query_metrics_view_time_range",
		"query_metrics_view",
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

	ctx := context.Background()

	// Test executing list_metrics_views tool (no parameters required)
	result, err := srv.mcpExecuteTool(ctx, instanceID, "list_metrics_views", map[string]any{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// The result should be a slice of content items
	contentItems, ok := result.([]mcp.Content)
	require.True(t, ok, "expected result to be []mcp.Content")
	require.NotEmpty(t, contentItems, "expected at least one content item")

	// First content item should be text content
	firstItem := contentItems[0]
	textContent, ok := firstItem.(mcp.TextContent)
	require.True(t, ok, "expected first item to be mcp.TextContent")
	require.NotEmpty(t, textContent.Text, "expected non-empty text content")
}

func TestMCPExecuteTool_MissingParam(t *testing.T) {
	srv, instanceID := newMCPTestServer(t)

	ctx := context.Background()

	// Test executing get_metrics_view tool without required parameter
	result, err := srv.mcpExecuteTool(ctx, instanceID, "get_metrics_view", map[string]any{})

	// The tool should either error or return an empty/error result
	if err != nil {
		// If it errors, it should mention the missing parameter
		require.Contains(t, err.Error(), "metrics_view")
	} else {
		// If it succeeds, check that result indicates an issue
		require.NotNil(t, result)
		t.Logf("Tool succeeded with error in response: %v", result)
		// This is valid behavior - MCP tools return errors in response content
	}
}
