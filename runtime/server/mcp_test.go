package server

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/ai"
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
SELECT 'US' AS country, TIMESTAMP '2024-01-15 00:00:00' AS event_time
`,
			// Metrics view
			"mv1.yaml": `
type: metrics_view
model: m
timeseries: event_time
dimensions:
- column: country
measures:
- name: row_count
  expression: COUNT(*)
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
		ai.ListMetricsViewsName,
		ai.GetMetricsViewName,
		ai.QueryMetricsViewName,
		ai.QueryMetricsViewSummaryName,
		ai.ProjectStatusName,
		ai.QuerySQLName,
		ai.ListTablesName,
		ai.ShowTableName,
		ai.ListBucketsName,
		ai.ListBucketObjectsName,
		ai.ListFilesName,
		ai.ReadFileName,
		ai.SearchFilesName,
		ai.WriteFileName,
	}
	require.Len(t, tools.Tools, len(expectedTools))
	for _, tool := range tools.Tools {
		require.Contains(t, expectedTools, tool.Name)
		require.NotEmpty(t, tool.Name)
		require.NotEmpty(t, tool.Description)
		require.NotEmpty(t, tool.InputSchema)
	}

	// Test metrics view listing
	mvs, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.ListMetricsViewsName})
	require.NoError(t, err)
	require.False(t, mvs.IsError)
	mvsText := mvs.Content[0].(*mcp.TextContent).Text
	require.Contains(t, mvsText, "mv1")
	require.Contains(t, mvsText, "mv2")

	// Test that it handles missing parameters
	_, err = conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.GetMetricsViewName})
	require.ErrorContains(t, err, "missing properties")

	// Test a query with object/array-typed arguments sent as JSON-encoded strings.
	// Some MCP clients stringify nested arguments (see https://github.com/anthropics/claude-code/issues/25865);
	// the server should tolerantly decode them.
	res, err := conn.CallTool(t.Context(), &mcp.CallToolParams{
		Name: ai.QueryMetricsViewName,
		Arguments: map[string]any{
			"metrics_view": "mv1",
			"dimensions":   `[{"name": "country"}]`,
			"measures":     []any{map[string]any{"name": "row_count"}},
			"time_range":   `{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}`,
			"where":        `{"cond": {"op": "eq", "exprs": [{"name": "country"}, {"val": "US"}]}}`,
			"limit":        10,
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	resText := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, resText, "US")
	require.Contains(t, resText, "row_count")

	// Test a query where comparison_time_range, having, sort and a nested measure compute are also stringified
	res, err = conn.CallTool(t.Context(), &mcp.CallToolParams{
		Name: ai.QueryMetricsViewName,
		Arguments: map[string]any{
			"metrics_view": "mv1",
			"dimensions":   `[{"name": "country"}]`,
			"measures": []any{
				map[string]any{"name": "row_count"},
				map[string]any{"name": "row_count_prev", "compute": `{"comparison_value": {"measure": "row_count"}}`},
			},
			"time_range":            `{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}`,
			"comparison_time_range": `{"start": "2023-12-01T00:00:00Z", "end": "2024-01-01T00:00:00Z"}`,
			"having":                `{"cond": {"op": "gt", "exprs": [{"name": "row_count"}, {"val": 0}]}}`,
			"sort":                  `[{"name": "row_count", "desc": true}]`,
			"limit":                 10,
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	resText = res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, resText, "US")
	require.Contains(t, resText, "row_count_prev")

	// Test that a string that is not valid JSON still fails schema validation
	_, err = conn.CallTool(t.Context(), &mcp.CallToolParams{
		Name: ai.QueryMetricsViewName,
		Arguments: map[string]any{
			"metrics_view": "mv1",
			"measures":     []any{map[string]any{"name": "row_count"}},
			"time_range":   "last 7 days",
		},
	})
	require.ErrorContains(t, err, `want "object"`)
}
