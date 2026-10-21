package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/ai/mcpapp"
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
		ai.QueryMetricsViewName,
		ai.QueryMetricsViewSummaryName,
		ai.CreateChartName,
		ai.ChartDataName,
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

// TestMCPRequestStartsSession covers the classification of sessionless MCP requests, and that the body reaches the MCP handler intact.
func TestMCPRequestStartsSession(t *testing.T) {
	check := func(t *testing.T, method, body string, want bool) {
		t.Helper()
		var r io.Reader
		if body != "" {
			r = strings.NewReader(body)
		}
		req := httptest.NewRequest(method, "/mcp", r)
		require.Equal(t, want, mcpRequestStartsSession(req))
		got, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(got), "the body must be left intact for the MCP handler")
	}

	check(t, http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`, true)
	check(t, http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_metrics_views","arguments":{}}}`, true)
	check(t, http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"resources/list","params":{}}`, false)
	check(t, http.MethodPost, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, false)
	check(t, http.MethodGet, "", false)

	// A body larger than the peeked prefix cannot be classified and starts a session, so a large write_file call is never misclassified.
	large := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/models/m.sql","contents":"` + strings.Repeat("x", mcpRequestPeekBytes) + `"}}}`
	check(t, http.MethodPost, large, true)
	// Neither can a body that is not JSON.
	check(t, http.MethodPost, "not json", true)
}

// TestMCPApps checks the MCP Apps surface when the widget bundle is available:
// the widget resource and capability are advertised, create_chart carries the _meta.ui link, chart_data is app-only,
// create_chart returns markdown content plus structured chart data, and sessionless protocol requests do not persist sessions.
func TestMCPApps(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "widget.html"), []byte("<!doctype html><html><body>rill widget</body></html>"), 0o644))
	t.Setenv(mcpapp.DirEnv, dir)
	widget, ok := mcpapp.Widget()
	require.True(t, ok)

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: mcpChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)
	httpSrv := httptest.NewServer(auth.HTTPMiddleware(srv.aud, srv.mcpHandler()))
	defer httpSrv.Close()

	countSessions := func(t *testing.T) int {
		catalog, release, err := rt.Catalog(t.Context(), instanceID)
		require.NoError(t, err)
		defer release()
		sessions, err := catalog.FindAISessions(t.Context(), "", "")
		require.NoError(t, err)
		return len(sessions)
	}

	// A client that supports the MCP Apps UI extension
	clientOpts := &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}}
	clientOpts.Capabilities.AddExtension(mcpapp.UIExtension, map[string]any{"mimeTypes": []any{mcpapp.MIMEType}})
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "claude-ai", Version: "1.0.0"}, clientOpts)
	// Claude and ChatGPT speak protocol 2025-06-18, whose handshake is an "initialize" request that starts a session.
	// The SDK client defaults to a newer protocol that discovers the server without one.
	conn, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, &mcp.ClientSessionOptions{ProtocolVersion: "2025-06-18"})
	require.NoError(t, err)
	defer conn.Close()
	require.Contains(t, conn.InitializeResult().Capabilities.Extensions, mcpapp.UIExtension)
	require.NotNil(t, conn.InitializeResult().Capabilities.Resources)
	sessionsAfterInit := countSessions(t)
	require.GreaterOrEqual(t, sessionsAfterInit, 1)

	// create_chart links to the widget and chart_data is only visible to the widget
	tools, err := conn.ListTools(t.Context(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	toolsByName := make(map[string]*mcp.Tool)
	for _, tool := range tools.Tools {
		toolsByName[tool.Name] = tool
	}
	require.Contains(t, toolsByName, ai.CreateChartName)
	require.Contains(t, toolsByName, ai.ChartDataName)
	require.Equal(t, widget.URI, toolsByName[ai.CreateChartName].Meta["ui"].(map[string]any)["resourceUri"])
	require.Equal(t, "Creating chart...", toolsByName[ai.CreateChartName].Meta["openai/toolInvocation/invoking"], "existing meta is kept")
	require.Equal(t, []any{"app"}, toolsByName[ai.ChartDataName].Meta["ui"].(map[string]any)["visibility"])
	require.Nil(t, toolsByName[ai.QueryMetricsViewName].Meta["ui"])

	// The widget resource is listed and readable
	resources, err := conn.ListResources(t.Context(), &mcp.ListResourcesParams{})
	require.NoError(t, err)
	require.Len(t, resources.Resources, 1)
	require.Equal(t, widget.URI, resources.Resources[0].URI)
	require.Equal(t, "text/html;profile=mcp-app", resources.Resources[0].MIMEType)
	require.Equal(t, true, resources.Resources[0].Meta["ui"].(map[string]any)["prefersBorder"])
	read, err := conn.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: widget.URI})
	require.NoError(t, err)
	require.Len(t, read.Contents, 1)
	require.Equal(t, "text/html;profile=mcp-app", read.Contents[0].MIMEType)
	require.Equal(t, widget.HTML, read.Contents[0].Text)

	// create_chart takes the Flint shape externally and returns a markdown fallback as content and the chart data as structured content.
	// The previous shape is still accepted, since hosts cache a connector's tool list: it renders through the widget's legacy path.
	legacy, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.CreateChartName, Arguments: mcpLegacyChartArgs()})
	require.NoError(t, err)
	require.False(t, legacy.IsError, "the previous shape is still accepted")
	legacyOut := legacy.StructuredContent.(map[string]any)
	require.Contains(t, legacyOut, "spec")
	require.NotContains(t, legacyOut, "input")
	res, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.CreateChartName, Arguments: mcpChartArgs()})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)
	text := res.Content[0].(*mcp.TextContent).Text
	require.True(t, strings.HasPrefix(text, "**Bar Chart: Total Bids by Country** (2024-01-01 to 2024-02-01)"), text)
	require.Contains(t, text, "| Country | Total Bids |")
	require.Contains(t, text, "| US | 10 |")
	require.Contains(t, text, "_2 of 2 rows shown_")
	structured, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok, "structuredContent should be an object, got %T", res.StructuredContent)
	require.Equal(t, "chart", structured["widget"])
	require.Equal(t, "Bar Chart", structured["chart_type"])
	require.NotContains(t, structured, "version")
	require.NotContains(t, structured, "spec")
	require.Len(t, structured["data"], 2)
	require.Equal(t, []any{"US", "CA"}, structured["domain_values"].(map[string]any)["country"])
	require.Contains(t, structured["fields"], "total_bids")
	require.Equal(t, "Category", structured["semantic_types"].(map[string]any)["country"])
	require.Equal(t, "Total Bids", structured["field_display_names"].(map[string]any)["total_bids"])
	input := structured["input"].(map[string]any)
	require.Equal(t, "Bar Chart", input["chart_spec"].(map[string]any)["chartType"])
	require.EqualValues(t, 10, input["chart_spec"].(map[string]any)["encodings"].(map[string]any)["x"].(map[string]any)["limit"], "the input carries the applied limit")
	require.Equal(t, "mv1", input["data"].(map[string]any)["metrics_view"])

	// chart_data is callable by the widget and returns the data fields only
	res, err = conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.ChartDataName, Arguments: mcpChartArgs()})
	require.NoError(t, err)
	require.False(t, res.IsError)
	structured, ok = res.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Len(t, structured["data"], 2)
	require.NotContains(t, structured, "spec")
	require.NotContains(t, structured, "input")

	// A protocol request without a session ID gets a transient session: the catalog is unchanged.
	sessionsBefore := countSessions(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"resources/list","params":{}}`
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, httpSrv.URL, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, string(respBody))
	require.Contains(t, string(respBody), widget.URI)
	require.Equal(t, sessionsBefore, countSessions(t), "sessionless resources/list must not persist a session")
}

// TestMCPAppsWithoutWidget asserts that without the widget bundle create_chart is still served and external clients get the chart
// as markdown and data, while the widget surface is absent: no capability, no resource, no resourceUri on create_chart,
// and no chart_data, which only the widget may call.
func TestMCPAppsWithoutWidget(t *testing.T) {
	// An override directory without a bundle hides the copy embedded in the binary.
	t.Setenv(mcpapp.DirEnv, t.TempDir())

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: mcpChartTestFiles(""),
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)
	httpSrv := httptest.NewServer(auth.HTTPMiddleware(srv.aud, srv.mcpHandler()))
	defer httpSrv.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "mcp/test", Version: "1.0.0"}, nil)
	conn, err := mcpClient.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: httpSrv.URL}, nil)
	require.NoError(t, err)
	defer conn.Close()

	require.NotContains(t, conn.InitializeResult().Capabilities.Extensions, mcpapp.UIExtension)
	require.Nil(t, conn.InitializeResult().Capabilities.Resources)

	// create_chart is listed without a link to the widget; chart_data is not listed at all
	tools, err := conn.ListTools(t.Context(), &mcp.ListToolsParams{})
	require.NoError(t, err)
	toolsByName := make(map[string]*mcp.Tool)
	for _, tool := range tools.Tools {
		toolsByName[tool.Name] = tool
		require.Nil(t, tool.Meta["ui"], "tool %q", tool.Name)
	}
	require.Contains(t, toolsByName, ai.CreateChartName)
	require.NotContains(t, toolsByName, ai.ChartDataName)
	require.Equal(t, "Creating chart...", toolsByName[ai.CreateChartName].Meta["openai/toolInvocation/invoking"])

	// An external client gets the chart as markdown and data
	res, err := conn.CallTool(t.Context(), &mcp.CallToolParams{Name: ai.CreateChartName, Arguments: mcpChartArgs()})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "| Country | Total Bids |")
	structured, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Len(t, structured["data"], 2)
}

// mcpChartTestFiles is a project with a small time series metrics view for the chart tools.
func mcpChartTestFiles(rillYAML string) map[string]string {
	return map[string]string{
		"rill.yaml": rillYAML,
		"m.sql": `
SELECT * FROM (VALUES ('US', TIMESTAMP '2024-01-15 00:00:00', 10), ('CA', TIMESTAMP '2024-01-16 00:00:00', 5)) AS t(country, event_time, bids)
`,
		"mv1.yaml": `
type: metrics_view
model: m
timeseries: event_time
dimensions:
- column: country
measures:
- name: total_bids
  expression: SUM(bids)
explore:
  skip: true
`,
	}
}

// mcpChartArgs are create_chart (and chart_data) arguments for the metrics view of mcpChartTestFiles, in the Flint shape external clients use.
func mcpChartArgs() map[string]any {
	return map[string]any{
		"chart_spec": map[string]any{
			"chartType": "Bar Chart",
			"encodings": map[string]any{
				"x": map[string]any{"field": "country", "limit": 10, "sortBy": "y", "sortOrder": "descending"},
				"y": "total_bids",
			},
		},
		"data": map[string]any{
			"metrics_view": "mv1",
			"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
		},
	}
}

// mcpLegacyChartArgs are the same chart in the in-app shape, which the MCP server does not advertise.
func mcpLegacyChartArgs() map[string]any {
	return map[string]any{
		"chart_type": "bar_chart",
		"spec": map[string]any{
			"metrics_view": "mv1",
			"time_range":   map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
			"color":        "primary",
			"x":            map[string]any{"field": "country", "type": "nominal", "limit": 10, "sort": "-y"},
			"y":            map[string]any{"field": "total_bids", "type": "quantitative"},
		},
	}
}
