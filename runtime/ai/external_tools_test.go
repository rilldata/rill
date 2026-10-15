package ai_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// newExternalMCPServer serves an in-process MCP server with the given tools, which echo their name and query.
// Requests must carry the given bearer token if it is not empty.
func newExternalMCPServer(t *testing.T, token string, tools ...string) *httptest.Server {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "ext", Version: "1.0.0"}, nil)
	type args struct {
		Query string `json:"query"`
	}
	for _, name := range tools {
		mcp.AddTool(server, &mcp.Tool{Name: name, Description: "Tool " + name}, func(ctx context.Context, req *mcp.CallToolRequest, args args) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: name + ": " + args.Query}}}, nil, nil
		})
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestClaims(email, domain string, admin bool) *runtime.SecurityClaims {
	return &runtime.SecurityClaims{
		UserID:         email,
		UserAttributes: map[string]any{"email": email, "domain": domain, "admin": admin, "groups": []any{}},
		Permissions:    []runtime.Permission{runtime.UseAI, runtime.ReadMetrics, runtime.ReadObjects},
	}
}

func TestExternalTools(t *testing.T) {
	ext := newExternalMCPServer(t, "svc-token", "search_issues", "delete_issue")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Variables: map[string]string{"ext_token": "svc-token"},
		Files: map[string]string{
			"rill.yaml": "",
			// No security policy: only admins may use the server, since every user acts with its credentials.
			"connectors/admin_only.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
headers:
  Authorization: "Bearer {{ .env.ext_token }}"
`, ext.URL),
			// A server whose security policy cannot be evaluated for users whose claims lack the attribute it references.
			"connectors/broken.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
headers:
  Authorization: "Bearer {{ .env.ext_token }}"
security:
  access: "{{ .user.admin }}"
`, ext.URL),
			// A server explicitly opened to every user with AI access.
			"connectors/open.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
headers:
  Authorization: "Bearer {{ .env.ext_token }}"
security:
  access: true
`, ext.URL),
			// A server restricted by domain, scoped to the analyst, with a tool excluded for non-admins.
			"connectors/tracker.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
display_name: Tracker
ai_instructions: Use it for issues.
headers:
  Authorization: "Bearer {{ .env.ext_token }}"
security:
  access: "'{{ .user.domain }}' = 'rilldata.com'"
  agents: [analyst]
  exclude:
    - if: "NOT {{ .user.admin }}"
      names: [delete_issue]
`, ext.URL),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	runner := ai.NewRunner(rt, activity.NewNoopClient())
	ctx := t.Context()

	// A user outside the domain whose claims carry no "admin" attribute: the policies referencing it cannot be
	// evaluated for them, which hides those servers without breaking discovery of the open server.
	outsider := &runtime.SecurityClaims{
		UserID:         "x@example.com",
		UserAttributes: map[string]any{"email": "x@example.com", "domain": "example.com", "groups": []any{}},
		Permissions:    []runtime.Permission{runtime.UseAI, runtime.ReadMetrics, runtime.ReadObjects},
	}
	servers, err := runner.DiscoverExternalServers(ctx, instanceID, outsider)
	require.NoError(t, err)
	require.Len(t, servers, 1)
	require.Equal(t, "open", servers[0].Connector)
	require.Equal(t, "open", servers[0].DisplayName)
	require.True(t, servers[0].Connected)
	require.Len(t, servers[0].Tools, 2)

	// A non-admin in the domain sees the restricted server without the excluded tool, but neither admin-only server
	servers, err = runner.DiscoverExternalServers(ctx, instanceID, newTestClaims("u@rilldata.com", "rilldata.com", false))
	require.NoError(t, err)
	require.Len(t, servers, 2)
	require.Equal(t, "open", servers[0].Connector)
	srv := servers[1]
	require.Equal(t, "tracker", srv.Connector)
	require.Equal(t, "Tracker", srv.DisplayName)
	require.Equal(t, []string{"analyst"}, srv.Agents)
	require.True(t, srv.Connected)
	require.NoError(t, srv.Err)
	require.Len(t, srv.Tools, 1)
	require.Equal(t, "search_issues", srv.Tools[0].Name)
	toolName := ai.ExternalToolName("tracker", "search_issues")
	require.Equal(t, "mcp__tracker__search_issues", toolName)
	require.Contains(t, srv.Compiled, toolName)
	require.Equal(t, "Tracker: search_issues", srv.Compiled[toolName].Spec.Title)
	require.Equal(t, "tracker", srv.Compiled[toolName].Spec.Meta["rill/mcpServer"])

	// An admin sees all servers and both tools of the restricted server
	servers, err = runner.DiscoverExternalServers(ctx, instanceID, newTestClaims("a@rilldata.com", "rilldata.com", true))
	require.NoError(t, err)
	require.Len(t, servers, 4)
	require.Equal(t, "admin_only", servers[0].Connector)
	require.True(t, servers[0].Connected)
	require.Len(t, servers[0].Tools, 2)
	require.Equal(t, "broken", servers[1].Connector)
	require.True(t, servers[1].Connected)
	require.Len(t, servers[3].Tools, 2)

	// Through a session, the tools are scoped to the allowed agent
	session, err := runner.Session(ctx, &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("u@rilldata.com", "rilldata.com", false), UserAgent: "rill/test"})
	require.NoError(t, err)
	openTools := []string{ai.ExternalToolName("open", "delete_issue"), ai.ExternalToolName("open", "search_issues")}
	require.Equal(t, append(append([]string{}, openTools...), toolName), session.ExternalToolNames(ctx, "analyst"))
	require.Equal(t, openTools, session.ExternalToolNames(ctx, "developer"), "the restricted server is scoped to the analyst agent")
	promptData := session.ExternalServerPromptData(ctx, "analyst")
	require.Len(t, promptData["connected"], 2)

	_, ok := session.Tool(toolName)
	require.True(t, ok)
	_, ok = session.Tool(ai.ExternalToolName("tracker", "delete_issue"))
	require.False(t, ok, "excluded tools are not compiled")

	var out ai.ExternalToolResult
	_, err = session.CallTool(ctx, ai.RoleUser, toolName, &out, map[string]any{"query": "bugs"})
	require.NoError(t, err)
	require.Equal(t, "search_issues: bugs", out.Content)

	// External tools are not exposed through Rill's own MCP server (which serves external clients, hence the non-rill user agent)
	for _, tool := range ai.MCPToolSpecs() {
		require.False(t, strings.HasPrefix(tool.Name, ai.ExternalToolPrefix))
	}
	mcpSession, err := runner.Session(ctx, &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("u@rilldata.com", "rilldata.com", false), UserAgent: "mcp/test"})
	require.NoError(t, err)
	mcpSession.ExternalServers(ctx)
	require.NotEmpty(t, mcpSession.ExternalToolNames(ctx, "analyst"))
	require.NotNil(t, mcpSession.MCPServer(ctx))
}

func TestExternalToolsRejectedCredentials(t *testing.T) {
	ext := newExternalMCPServer(t, "svc-token", "search_issues")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"connectors/tracker.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
headers:
  Authorization: "Bearer wrong"
`, ext.URL),
		},
	})
	// The reconciler connects to the server, so rejected credentials surface as a reconcile error
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 1, 0)

	runner := ai.NewRunner(rt, activity.NewNoopClient())
	servers, err := runner.DiscoverExternalServers(t.Context(), instanceID, newTestClaims("a@x.com", "x.com", true))
	require.NoError(t, err)
	require.Len(t, servers, 1)
	require.False(t, servers[0].Connected)
	require.Error(t, servers[0].Err)
	require.Empty(t, servers[0].Tools)
}

func TestExternalToolsLimit(t *testing.T) {
	var tools []string
	for i := range 85 {
		tools = append(tools, fmt.Sprintf("tool_%02d", i))
	}
	ext := newExternalMCPServer(t, "", tools...)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":             "",
			"connectors/many.yaml":  fmt.Sprintf("type: connector\ndriver: mcp\nurl: %s\n", ext.URL),
			"connectors/other.yaml": fmt.Sprintf("type: connector\ndriver: mcp\nurl: %s\n", ext.URL),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	runner := ai.NewRunner(rt, activity.NewNoopClient())
	session, err := runner.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("a@x.com", "x.com", true), UserAgent: "rill/test"})
	require.NoError(t, err)

	// Tools are kept in server order, so the first server keeps 80 tools and the second loses all of them
	names := session.ExternalToolNames(t.Context(), "analyst")
	require.Len(t, names, 80)
	for _, name := range names {
		require.True(t, strings.HasPrefix(name, ai.ExternalToolName("many", "")))
	}
	_, ok := session.Tool(ai.ExternalToolName("other", "tool_00"))
	require.False(t, ok)
	promptData := session.ExternalServerPromptData(t.Context(), "analyst")
	require.Len(t, promptData["connected"], 1, "a server without remaining tools is not described to the agent")
}

func TestExternalToolName(t *testing.T) {
	require.Equal(t, "mcp__linear__search_issues", ai.ExternalToolName("linear", "search_issues"))
	require.Equal(t, "mcp__linear__", ai.ExternalToolName("linear", ""))

	// Names that change under sanitizing get a hash so they cannot collide with names that do not
	lossy := ai.ExternalToolName("My-Server", "get.item")
	require.Regexp(t, `^mcp__my_server__get_item_[0-9a-f]{8}$`, lossy)
	require.NotEqual(t, lossy, ai.ExternalToolName("my_server", "get_item"))
	require.NotEqual(t, ai.ExternalToolName("github", "search-users"), ai.ExternalToolName("github", "search_users"))
	require.NotEqual(t, ai.ExternalToolName("a.b", "c"), ai.ExternalToolName("a", "b/c"))
	require.NotEqual(t, ai.ExternalToolName("a__b", "c"), ai.ExternalToolName("a", "b__c"))
	require.NotEqual(t, ai.ExternalToolName("srv", "Search"), ai.ExternalToolName("srv", "search"))
	long := ai.ExternalToolName("a_very_long_connector_name_for_testing", "an_even_longer_tool_name_that_goes_past_the_limit_for_sure")
	require.LessOrEqual(t, len(long), 64)
	require.True(t, strings.HasPrefix(long, "mcp__a_very_long_connector_name_for_testing__"))
	require.NotEqual(t, long, ai.ExternalToolName("a_very_long_connector_name_for_testing", "an_even_longer_tool_name_that_goes_past_the_limit_for_sure2"))
}
