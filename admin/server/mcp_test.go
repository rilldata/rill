package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/stretchr/testify/require"
)

func TestMCPForwardedToolSpecs(t *testing.T) {
	tools, err := mcpForwardedToolSpecs()
	require.NoError(t, err)
	require.Len(t, tools, len(mcpForwardedTools))

	for _, tool := range tools {
		schema, ok := tool.InputSchema.(*jsonschema.Schema)
		require.True(t, ok, "tool %q", tool.Name)
		require.Contains(t, schema.Properties, mcpProjectArg, "tool %q", tool.Name)
		require.Contains(t, schema.Required, mcpProjectArg, "tool %q", tool.Name)

		// query_metrics_view has a hand-written schema, which must be extended and not replaced.
		if tool.Name == ai.QueryMetricsViewName {
			require.Contains(t, schema.Properties, "metrics_view")
			require.NotEmpty(t, schema.Defs)
		}
	}
}

func TestMCPForwardsPersonalCanvasDeveloper(t *testing.T) {
	tools, err := mcpForwardedToolSpecs()
	require.NoError(t, err)

	var names []string
	for _, tool := range tools {
		names = append(names, tool.Name)
	}

	// Clients can create personal canvases, but only through the sub-agent that knows how to author them.
	require.Contains(t, names, ai.DevelopPersonalCanvasName)
	require.NotContains(t, names, ai.WritePersonalCanvasName)
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
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
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
	res, err := s.callRuntimeTool(t.Context(), depl, "jwt", client, "echo", json.RawMessage(`{"name":"hello"}`))
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Equal(t, map[string]any{"name": "hello"}, res.StructuredContent)

	// Calling a tool the runtime does not serve returns a tool error the client can act on, not a protocol error.
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", client, ai.ListSkillsName, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Len(t, res.Content, 1)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "is not available in this project")

	// For develop_personal_canvas, the error says the project doesn't enable personal canvases.
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", client, ai.DevelopPersonalCanvasName, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "Personal canvases are not enabled in this project")

	// Invalid arguments are also returned as a tool error.
	res, err = s.callRuntimeTool(t.Context(), depl, "jwt", client, "echo", json.RawMessage(`{"name":123}`))
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.NotEmpty(t, res.Content)

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
	_, err = s.callRuntimeTool(t.Context(), &database.Deployment{RuntimeHost: internal.URL, RuntimeInstanceID: "default"}, "jwt", client, "echo", json.RawMessage(`{}`))
	require.ErrorContains(t, err, "boom")
}
