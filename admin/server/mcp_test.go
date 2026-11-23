package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/ai/mcpapp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMCPForwardedToolSpecs(t *testing.T) {
	// Without the widget, create_chart is still forwarded (clients get its text fallback) but chart_data, which only the widget
	// may call, is left out; no tool carries UI metadata.
	tools, err := mcpForwardedToolSpecs(nil)
	require.NoError(t, err)
	require.Len(t, tools, len(mcpForwardedTools)-1)

	byName := make(map[string]*mcp.Tool)
	for _, tool := range tools {
		byName[tool.Name] = tool
		schema, ok := tool.InputSchema.(*jsonschema.Schema)
		require.True(t, ok, "tool %q", tool.Name)
		require.Contains(t, schema.Properties, mcpProjectArg, "tool %q", tool.Name)
		require.Contains(t, schema.Required, mcpProjectArg, "tool %q", tool.Name)
		require.Nil(t, tool.Meta["ui"], "tool %q", tool.Name)

		// query_metrics_view has a hand-written schema, which must be extended and not replaced.
		if tool.Name == ai.QueryMetricsViewName {
			require.Contains(t, schema.Properties, "metrics_view")
			require.NotEmpty(t, schema.Defs)
		}
	}
	require.Contains(t, byName, ai.CreateChartName)
	require.Equal(t, "Creating chart...", byName[ai.CreateChartName].Meta["openai/toolInvocation/invoking"], "the tool's own meta is kept")
	require.NotContains(t, byName, ai.ChartDataName)
	require.Contains(t, byName, ai.QueryMetricsViewName)

	// With the widget, create_chart links to it and chart_data is restricted to it.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "widget.html"), []byte("<html></html>"), 0o644))
	t.Setenv(mcpapp.DirEnv, dir)
	widget, ok := mcpapp.Widget()
	require.True(t, ok)

	tools, err = mcpForwardedToolSpecs(widget)
	require.NoError(t, err)
	require.Len(t, tools, len(mcpForwardedTools))
	byName = make(map[string]*mcp.Tool)
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	require.Equal(t, widget.URI, byName[ai.CreateChartName].Meta["ui"].(map[string]any)["resourceUri"])
	require.Equal(t, "Creating chart...", byName[ai.CreateChartName].Meta["openai/toolInvocation/invoking"], "existing meta is kept")
	require.Equal(t, []any{"app"}, byName[ai.ChartDataName].Meta["ui"].(map[string]any)["visibility"])
	require.Nil(t, byName[ai.QueryMetricsViewName].Meta["ui"])
	require.Contains(t, byName[ai.CreateChartName].InputSchema.(*jsonschema.Schema).Properties, mcpProjectArg)
	// chart_data's minimal schema gets the project argument too.
	chartData := byName[ai.ChartDataName].InputSchema.(*jsonschema.Schema)
	require.Contains(t, chartData.Properties, "chart_spec")
	require.Contains(t, chartData.Properties, mcpProjectArg)
	require.Equal(t, []string{mcpProjectArg}, chartData.Required)
}

// TestCallRuntimeToolPassthrough verifies that the content blocks, structured content and _meta a runtime tool returns
// reach the client unchanged, which the MCP Apps widget relies on.
func TestCallRuntimeToolPassthrough(t *testing.T) {
	type args struct {
		Name string `json:"name"`
	}
	type result struct {
		Name string `json:"name"`
	}
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		srv := mcp.NewServer(&mcp.Implementation{Name: "runtime"}, &mcp.ServerOptions{HasTools: true})
		mcp.AddTool(srv, &mcp.Tool{Name: "echo"}, func(ctx context.Context, req *mcp.CallToolRequest, a args) (*mcp.CallToolResult, result, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "**Chart: " + a.Name + "**"}},
				Meta:    map[string]any{"rill/chart": map[string]any{"rows": float64(1)}},
			}, result{Name: a.Name}, nil
		})
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	rt := httptest.NewServer(handler)
	t.Cleanup(rt.Close)

	s := &Server{}
	depl := &database.Deployment{RuntimeHost: rt.URL, RuntimeInstanceID: "default"}
	res, err := s.callRuntimeTool(t.Context(), depl, "jwt", &mcpClient{sessionID: "session", userAgent: "claude-ai"}, &mcp.CallToolParams{Name: "echo", Arguments: json.RawMessage(`{"name":"hello"}`)})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)
	require.Equal(t, "**Chart: hello**", res.Content[0].(*mcp.TextContent).Text)
	require.Equal(t, map[string]any{"name": "hello"}, res.StructuredContent)
	require.Equal(t, map[string]any{"rill/chart": map[string]any{"rows": float64(1)}}, map[string]any(res.Meta))
}

func TestTakeMCPProjectArg(t *testing.T) {
	org, project, rest, err := takeMCPProjectArg(json.RawMessage(`{"project":"org/proj","metrics_view":"mv"}`))
	require.NoError(t, err)
	require.Equal(t, "org", org)
	require.Equal(t, "proj", project)
	require.JSONEq(t, `{"metrics_view":"mv"}`, string(rest))

	// The project argument is required, and must name both an organization and a project.
	for _, args := range []string{``, `{}`, `{"project":""}`, `{"project":"org"}`, `{"project":"org/"}`, `{"project":"/proj"}`, `{"project":123}`, `not json`} {
		_, _, _, err := takeMCPProjectArg(json.RawMessage(args))
		require.Error(t, err, "args %q", args)
	}
}

// TestCallRuntimeToolErrors verifies how errors from a runtime's MCP server are surfaced to the client.
// It runs a real MCP server with the runtime's transport options, so the errors are the ones the SDK actually emits.
func TestCallRuntimeToolErrors(t *testing.T) {
	type echoArgs struct {
		Name string `json:"name"`
	}
	type echoResult struct {
		Name string `json:"name"`
	}

	// The server serves a single tool, so that calls to any other tool exercise the runtime's unknown tool error.
	var runtimeSessionID string
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		runtimeSessionID = r.Header.Get("Mcp-Session-Id")
		srv := mcp.NewServer(&mcp.Implementation{Name: "runtime"}, &mcp.ServerOptions{HasTools: true})
		mcp.AddTool(srv, &mcp.Tool{Name: "echo"}, func(ctx context.Context, req *mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, echoResult, error) {
			return nil, echoResult{Name: args.Name}, nil
		})
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	rt := httptest.NewServer(handler)
	t.Cleanup(rt.Close)

	s := &Server{}
	depl := &database.Deployment{RuntimeHost: rt.URL, RuntimeInstanceID: "default"}
	client := &mcpClient{sessionID: "session", userAgent: "test"}

	// A successful call returns the runtime's result.
	res, err := s.callRuntimeTool(t.Context(), depl, "jwt", client, &mcp.CallToolParams{Name: "echo", Arguments: json.RawMessage(`{"name":"hello"}`)})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Equal(t, map[string]any{"name": "hello"}, res.StructuredContent)
	// The call carries a session ID derived from the client's, so the runtime keeps one AI session per client and project.
	require.Equal(t, uuid.NewSHA1(mcpSessionNamespace, []byte("default/session")).String(), runtimeSessionID)

	// Calling a tool the runtime does not serve returns a tool error the client can act on, not a protocol error.
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", client, &mcp.CallToolParams{Name: ai.ListSkillsName, Arguments: json.RawMessage(`{}`)})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Len(t, res.Content, 1)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "unknown tool")

	// Invalid arguments are also returned as a tool error.
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", client, &mcp.CallToolParams{Name: "echo", Arguments: json.RawMessage(`{"name":123}`)})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.NotEmpty(t, res.Content)

	// Calls from clients on protocol versions >= 2026-07-28 carry their protocol version and capabilities in _meta, which must reach the runtime.
	newProtocolClient := &mcpClient{sessionID: "session", userAgent: "test", protocolVersion: "2026-07-28"}
	meta := mcp.Meta{
		mcp.MetaKeyProtocolVersion:    "2026-07-28",
		mcp.MetaKeyClientCapabilities: map[string]any{},
		mcp.MetaKeyClientInfo:         map[string]any{"name": "test", "version": "1.0"},
	}
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", newProtocolClient, &mcp.CallToolParams{Meta: meta, Name: "echo", Arguments: json.RawMessage(`{"name":"hello"}`)})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Equal(t, map[string]any{"name": "hello"}, res.StructuredContent)
	// The runtime's server info is stripped, so clients see the admin service's server info on every result.
	require.NotContains(t, res.Meta, mcp.MetaKeyServerInfo)

	// Any other error from the runtime stays a protocol error.
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"error":   jsonrpc.Error{Code: jsonrpc.CodeInternalError, Message: "boom"},
		})
	}))
	t.Cleanup(internal.Close)
	_, err = s.callRuntimeTool(t.Context(), &database.Deployment{RuntimeHost: internal.URL, RuntimeInstanceID: "default"}, "jwt", client, &mcp.CallToolParams{Name: "echo", Arguments: json.RawMessage(`{}`)})
	require.ErrorContains(t, err, "boom")
}

// TestMCPHandlerSessionID asserts that the MCP server issues a session ID on a client's first request and keeps the one a client sends.
// The server is stateless, so the SDK does not do this itself; without it, every tool call would become its own AI session.
func TestMCPHandlerSessionID(t *testing.T) {
	s := &Server{admin: &admin.Service{}, logger: zap.NewNop()}
	handler, err := s.mcpHandler()
	require.NoError(t, err)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	post := func(t *testing.T, sessionID, body string) *http.Response {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if sessionID != "" {
			req.Header.Set("Mcp-Session-Id", sessionID)
		}
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { res.Body.Close() })
		require.Equal(t, http.StatusOK, res.StatusCode)
		return res
	}

	res := post(t, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`)
	sessionID := res.Header.Get("Mcp-Session-Id")
	_, err = uuid.Parse(sessionID)
	require.NoError(t, err)

	res = post(t, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	require.Equal(t, sessionID, res.Header.Get("Mcp-Session-Id"))
}
