package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
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

	require.Contains(t, conn.InitializeResult().Instructions, "## Skills")

	// Test tool listings
	tools, err := conn.ListTools(t.Context(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	expectedTools := []string{
		ai.ListMetricsViewsName,
		ai.GetMetricsViewName,
		ai.ListCanvasesName,
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
		ai.ListSkillsName,
		ai.LoadSkillName,
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

	// Test that it handles missing parameters.
	// Since go-sdk v1.7, input validation failures are reported as tool errors rather than protocol errors.
	missing, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.GetMetricsViewName})
	require.NoError(t, err)
	require.True(t, missing.IsError)
	require.Contains(t, missing.Content[0].(*mcp.TextContent).Text, "missing properties")

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
	invalid, err := conn.CallTool(t.Context(), &mcp.CallToolParams{
		Name: ai.QueryMetricsViewName,
		Arguments: map[string]any{
			"metrics_view": "mv1",
			"measures":     []any{map[string]any{"name": "row_count"}},
			"time_range":   "last 7 days",
		},
	})
	require.NoError(t, err)
	require.True(t, invalid.IsError)
	require.Contains(t, invalid.Content[0].(*mcp.TextContent).Text, `want "object"`)
}

// TestMCPSkillTools asserts that the skill tools are advertised to callers that can use AI, and only to them.
// TestMCP covers that they are advertised even when the project defines no skills.
func TestMCPSkillTools(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"skills/glossary/SKILL.md": `---
description: Business glossary.
---

ARPU excludes trial users.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	httpSrv := httptest.NewServer(auth.HTTPMiddleware(srv.aud, srv.mcpHandler()))
	defer httpSrv.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "mcp/test", Version: "1.0.0"}, nil)
	conn, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, nil)
	require.NoError(t, err)
	defer conn.Close()

	toolNames := func(t *testing.T, conn *mcp.ClientSession) []string {
		tools, err := conn.ListTools(t.Context(), &mcp.ListToolsParams{})
		require.NoError(t, err)
		var names []string
		for _, tool := range tools.Tools {
			names = append(names, tool.Name)
		}
		return names
	}
	require.Subset(t, toolNames(t, conn), []string{ai.ListSkillsName, ai.LoadSkillName})

	// A client without UseAI is not offered the skill tools
	token, err := auth.NewDevToken(nil, []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics})
	require.NoError(t, err)
	httpClient := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.Header.Set("Authorization", "Bearer "+token)
		return http.DefaultTransport.RoundTrip(req)
	})}
	conn, err = mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL, HTTPClient: httpClient}, nil)
	require.NoError(t, err)
	defer conn.Close()
	require.NotSubset(t, toolNames(t, conn), []string{ai.ListSkillsName})
	require.NotSubset(t, toolNames(t, conn), []string{ai.LoadSkillName})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// TestMCPSessionContinuity asserts that all the requests of one MCP client connection share one AI session,
// and that separate connections get separate sessions.
// The server runs the go-sdk's streamable HTTP handler in stateless mode, which since go-sdk v1.8 no longer issues the Mcp-Session-Id header itself.
func TestMCPSessionContinuity(t *testing.T) {
	// Clients on the 2025 protocol versions (Claude, ChatGPT, Cursor, ...) do an initialize handshake.
	// Clients on the 2026-07-28 protocol version (the go-sdk client's default) do not: the server only ever sees their tool calls.
	for _, protocolVersion := range []string{"2025-06-18", "2026-07-28"} {
		t.Run(protocolVersion, func(t *testing.T) {
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files: map[string]string{"rill.yaml": ""},
			})
			testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

			srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
			require.NoError(t, err)

			httpSrv := httptest.NewServer(auth.HTTPMiddleware(srv.aud, srv.mcpHandler()))
			defer httpSrv.Close()

			catalog, release, err := rt.Catalog(t.Context(), instanceID)
			require.NoError(t, err)
			defer release()
			sessions := func(t *testing.T) []*drivers.AISession {
				res, err := catalog.FindAISessions(t.Context(), "", "")
				require.NoError(t, err)
				return res
			}
			require.Empty(t, sessions(t))

			// The first request is answered with a session ID, which the client echoes on every subsequent request.
			mcpClient := mcp.NewClient(&mcp.Implementation{Name: "mcp/test", Version: "1.0.0"}, nil)
			connOpts := &mcp.ClientSessionOptions{ProtocolVersion: protocolVersion}
			conn, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, connOpts)
			require.NoError(t, err)
			defer conn.Close()
			require.Equal(t, protocolVersion, conn.InitializeResult().ProtocolVersion)
			require.NotEmpty(t, conn.ID())

			_, err = conn.ListTools(t.Context(), &mcp.ListToolsParams{})
			require.NoError(t, err)
			_, err = conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.ListMetricsViewsName})
			require.NoError(t, err)
			_, err = conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.ListMetricsViewsName})
			require.NoError(t, err)

			// All the requests above land in a single session.
			res := sessions(t)
			require.Len(t, res, 1)
			require.Equal(t, conn.ID(), res[0].ID)
			// Note the user agent is the HTTP User-Agent header, not the client info from the initialize handshake:
			// a stateless server handles the initialized notification in a session synthesized without the initialize params, so InitializedHandler never sees the client info.
			require.NotEmpty(t, res[0].UserAgent)
			msgs, err := catalog.FindAIMessages(t.Context(), conn.ID())
			require.NoError(t, err)
			require.Len(t, msgs, 4) // two tool calls, each with a call and a result message

			// A second connection gets its own session.
			conn2, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, connOpts)
			require.NoError(t, err)
			defer conn2.Close()
			require.NotEmpty(t, conn2.ID())
			require.NotEqual(t, conn.ID(), conn2.ID())
			_, err = conn2.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.ListMetricsViewsName})
			require.NoError(t, err)
			require.Len(t, sessions(t), 2)

			// Requests other than POST are rejected without creating a session.
			for _, method := range []string{http.MethodGet, http.MethodDelete} {
				req, err := http.NewRequestWithContext(t.Context(), method, httpSrv.URL, nil)
				require.NoError(t, err)
				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				res.Body.Close()
				require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, method)
			}
			require.Len(t, sessions(t), 2)
		})
	}
}
