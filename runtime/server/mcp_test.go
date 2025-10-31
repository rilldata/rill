package server

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMCP(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m.sql": `
SELECT 'US' AS country
`,
			// Metrics view
			"mv1.yaml": `
type: metrics_view
model: m
dimensions:
- column: country
measures:
- expression: COUNT(*)
explore:
  skip: true
`,
			// Metrics view
			"mv2.yaml": `
type: metrics_view
model: m
dimensions:
- column: country
measures:
- expression: COUNT(*)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Create a test server for the MCP handler with auth middleware
	httpSrv := httptest.NewServer(auth.HTTPMiddleware(srv.aud, srv.mcpHandler()))
	defer httpSrv.Close()

	// Connect an MCP client
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "mcp/test", Version: "1.0.0"}, nil)
	conn, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, nil)
	require.NoError(t, err)
	defer conn.Close()

	// TODO: Use JWT with limited permissions when mcp-go supports client-side auth.
	// jwt, err := auth.NewDevToken(nil, []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI})
	// require.NoError(t, err)

	// Test tool listings
	tools, err := conn.ListTools(t.Context(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	expectedTools := []string{
		"list_metrics_views",
		"get_metrics_view",
		"query_metrics_view",
		"query_metrics_view_summary",
	}
	require.Len(t, tools.Tools, len(expectedTools))
	for _, tool := range tools.Tools {
		require.Contains(t, expectedTools, tool.Name)
		require.NotEmpty(t, tool.Name)
		require.NotEmpty(t, tool.Description)
		require.NotEmpty(t, tool.InputSchema)
	}

	// Test metrics view listing
	mvs, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: "list_metrics_views"})
	require.NoError(t, err)
	require.False(t, mvs.IsError)
	mvsText := mvs.Content[0].(*mcp.TextContent).Text
	require.Contains(t, mvsText, "mv1")
	require.Contains(t, mvsText, "mv2")

	// Test that it handles missing parameters
	_, err = conn.CallTool(t.Context(), &mcp.CallToolParams{Name: "get_metrics_view"})
	require.ErrorContains(t, err, "missing properties")
}
