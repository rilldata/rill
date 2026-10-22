package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// newExternalMCPServer serves an in-process MCP server with two tools. Requests must carry the given bearer token if it is not empty.
func newExternalMCPServer(t *testing.T, token string) *httptest.Server {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "ext", Version: "1.0.0"}, nil)
	type searchArgs struct {
		Query string `json:"query"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "search_issues", Description: "Search issues"}, func(ctx context.Context, req *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "results for " + args.Query + " as " + req.Extra.Header.Get("X-User")}}}, nil, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "delete_issue", Description: "Delete an issue"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, nil, nil
	})
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
	ext := newExternalMCPServer(t, "svc-token")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Variables: map[string]string{"ext_token": "svc-token"},
		Files: map[string]string{
			"rill.yaml": "",
			// A server with project-scoped credentials and no security policy: open to every user with AI access.
			"connectors/static.yaml": fmt.Sprintf(`
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
			// A server whose headers identify the user, restricted by a security policy.
			"connectors/tracker.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
display_name: Tracker
ai_instructions: Use it for issues.
headers:
  Authorization: "Bearer {{ .env.ext_token }}"
  X-User: "{{ .user.email }}"
security:
  access: "'{{ .user.domain }}' = 'rilldata.com'"
  agents: [analyst]
  exclude:
    - if: "NOT {{ .user.admin }}"
      names: [delete_issue]
`, ext.URL),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

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
	require.Equal(t, "static", servers[0].Connector)
	require.Equal(t, "static", servers[0].DisplayName)
	require.True(t, servers[0].Connected)
	require.Len(t, servers[0].Tools, 2)

	// A non-admin in the domain sees the restricted server without the excluded tool, and not the admin-only server
	servers, err = runner.DiscoverExternalServers(ctx, instanceID, newTestClaims("u@rilldata.com", "rilldata.com", false))
	require.NoError(t, err)
	require.Len(t, servers, 2)
	srv := servers[1]
	require.Equal(t, "tracker", srv.Connector)
	require.Equal(t, "Tracker", srv.DisplayName)
	require.Equal(t, []string{"analyst"}, srv.Agents)
	require.True(t, srv.Connected)
	require.False(t, srv.OAuth)
	require.NoError(t, srv.Err)
	require.Len(t, srv.Tools, 1)
	require.Equal(t, "search_issues", srv.Tools[0].Name)
	toolName := ai.ExternalToolName("tracker", "search_issues")
	require.Equal(t, "mcp__tracker__search_issues", toolName)
	require.Contains(t, srv.Compiled, toolName)
	require.Equal(t, "Tracker: search_issues", srv.Compiled[toolName].Spec.Title)

	// An admin sees all three servers and both tools of the restricted server
	servers, err = runner.DiscoverExternalServers(ctx, instanceID, newTestClaims("a@rilldata.com", "rilldata.com", true))
	require.NoError(t, err)
	require.Len(t, servers, 3)
	require.Equal(t, "broken", servers[0].Connector)
	require.True(t, servers[0].Connected)
	require.Len(t, servers[2].Tools, 2)

	// Through a session, the tools are scoped to the allowed agent, and calls resolve headers for the user.
	session, err := runner.Session(ctx, &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("u@rilldata.com", "rilldata.com", false), UserAgent: "rill/test"})
	require.NoError(t, err)
	staticTools := []string{ai.ExternalToolName("static", "delete_issue"), ai.ExternalToolName("static", "search_issues")}
	require.Equal(t, append(append([]string{}, staticTools...), toolName), session.ExternalToolNames(ctx, "analyst"))
	require.Equal(t, staticTools, session.ExternalToolNames(ctx, "developer"), "the restricted server is scoped to the analyst agent")
	promptData := session.ExternalServerPromptData(ctx, "analyst")
	require.Len(t, promptData["connected"], 2)
	require.Empty(t, promptData["unconnected"])

	_, ok := session.Tool(toolName)
	require.True(t, ok)
	_, ok = session.Tool(ai.ExternalToolName("tracker", "delete_issue"))
	require.False(t, ok, "excluded tools are not compiled")

	var out ai.ExternalToolResult
	_, err = session.CallTool(ctx, ai.RoleUser, toolName, &out, map[string]any{"query": "bugs"})
	require.NoError(t, err)
	require.Equal(t, "results for bugs as u@rilldata.com", out.Content)

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

func TestExternalToolsAuthorizationRequired(t *testing.T) {
	ext := newExternalMCPServer(t, "user-token")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"connectors/tracker.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
auth: oauth
`, ext.URL),
		},
	})
	// An OAuth server without a local token still reconciles: tools are listed per user
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	runner := ai.NewRunner(rt, activity.NewNoopClient())
	claims := newTestClaims("u@x.com", "x.com", false)
	servers, err := runner.DiscoverExternalServers(t.Context(), instanceID, claims)
	require.NoError(t, err)
	require.Len(t, servers, 1)
	require.True(t, servers[0].OAuth)
	require.False(t, servers[0].Authorized)
	require.False(t, servers[0].Connected)
	require.NoError(t, servers[0].Err)
	require.Empty(t, servers[0].Tools)

	// The unconnected server is described to the agent so it can ask the user to connect
	session, err := runner.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: claims, UserAgent: "rill/test"})
	require.NoError(t, err)
	require.Empty(t, session.ExternalToolNames(t.Context(), "analyst"))
	promptData := session.ExternalServerPromptData(t.Context(), "analyst")
	require.Empty(t, promptData["connected"])
	require.Len(t, promptData["unconnected"], 1)

	// Anonymous users never see OAuth servers
	servers, err = runner.DiscoverExternalServers(t.Context(), instanceID, &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.UseAI}})
	require.NoError(t, err)
	require.Empty(t, servers)

	// The authorization error is a JSON envelope the UI can recognize
	authErr := &ai.ExternalAuthorizationRequiredError{Connector: "tracker", DisplayName: "Tracker"}
	var envelope map[string]any
	require.NoError(t, json.Unmarshal([]byte(authErr.Error()), &envelope))
	require.Equal(t, ai.ErrorCodeMCPAuthorizationRequired, envelope["code"])
	require.Equal(t, "tracker", envelope["connector"])
}

func TestExternalToolsPolicyChange(t *testing.T) {
	ext := newExternalMCPServer(t, "")
	connector := fmt.Sprintf(`
type: connector
driver: mcp
url: %s
`, ext.URL)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": "", "connectors/tracker.yaml": connector},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	runner := ai.NewRunner(rt, activity.NewNoopClient())
	claims := newTestClaims("u@x.com", "x.com", false)
	servers, err := runner.DiscoverExternalServers(t.Context(), instanceID, claims)
	require.NoError(t, err)
	require.Len(t, servers[0].Tools, 2)

	// Adding a policy to an existing connector takes effect without the connector's state changing
	testruntime.PutFiles(t, rt, instanceID, map[string]string{"connectors/tracker.yaml": connector + `
security:
  access: true
  include:
    - if: true
      names: [search_issues]
`})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	servers, err = runner.DiscoverExternalServers(t.Context(), instanceID, claims)
	require.NoError(t, err)
	require.Len(t, servers[0].Tools, 1)
	require.Equal(t, "search_issues", servers[0].Tools[0].Name)
}

func TestExternalToolsSearch(t *testing.T) {
	ext := newExternalMCPServer(t, "")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"connectors/tracker.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
display_name: Tracker
`, ext.URL),
			"connectors/devonly.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
security:
  access: true
  agents: [developer]
`, ext.URL),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	searchIssues := ai.ExternalToolName("tracker", "search_issues")
	deleteIssue := ai.ExternalToolName("tracker", "delete_issue")
	script := &scriptedAIService{turns: []turnFunc{
		// Calling an external tool before loading it is refused.
		toolCallTurn(deleteIssue, nil),
		toolCallTurn(ai.SearchExternalToolsName, map[string]any{"query": "search issues"}),
		// Tools of servers scoped to another agent are not found.
		toolCallTurn(ai.SearchExternalToolsName, map[string]any{"names": []any{ai.ExternalToolName("devonly", "search_issues")}}),
		toolCallTurn(searchIssues, map[string]any{"query": "bugs"}),
		textTurn("first"),
		textTurn("second"),
	}}
	runner := ai.NewRunner(rt, activity.NewNoopClient())
	session, err := runner.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("u@x.com", "x.com", false), UserAgent: "rill/test"})
	require.NoError(t, err)
	session.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
		return script, func() {}, nil
	})

	res, err := session.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "find bugs"})
	require.NoError(t, err)

	// External tools are listed by name in the prompt, but only the search tool is offered up front.
	var prompt strings.Builder
	for _, m := range script.inputs[0] {
		for _, b := range m.Content {
			prompt.WriteString(b.GetText())
		}
	}
	require.Contains(t, prompt.String(), "Tools: "+deleteIssue+", "+searchIssues)
	require.Contains(t, script.tools[0], ai.SearchExternalToolsName)
	require.False(t, slices.ContainsFunc(script.tools[0], isExternalToolName))

	// The search loads only the matching tool, which is offered from the next iteration on.
	require.NotContains(t, script.tools[1], searchIssues)
	require.Contains(t, script.tools[2], searchIssues)
	require.NotContains(t, script.tools[2], deleteIssue)

	results := callResults(session, res.Call.ID)
	require.Equal(t, ai.MessageContentTypeError, results[0].ContentType)
	require.Contains(t, results[0].Content, "not available")
	var found ai.SearchExternalToolsResult
	require.NoError(t, json.Unmarshal([]byte(results[1].Content), &found))
	require.Len(t, found.Tools, 1)
	require.Equal(t, searchIssues, found.Tools[0].Name)
	require.Equal(t, "Tracker", found.Tools[0].Server)
	require.NoError(t, json.Unmarshal([]byte(results[2].Content), &found))
	require.Empty(t, found.Tools)
	require.Equal(t, []string{ai.ExternalToolName("devonly", "search_issues")}, found.NotFound)
	require.Equal(t, ai.MessageContentTypeJSON, results[3].ContentType)
	require.Contains(t, results[3].Content, "results for bugs")

	// Loaded tools stay offered in later turns.
	_, err = session.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "again"})
	require.NoError(t, err)
	require.Contains(t, script.tools[5], searchIssues)
	require.NotContains(t, script.tools[5], deleteIssue)
}

func TestExternalToolsLimit(t *testing.T) {
	ext := newBigExternalMCPServer(t, 130)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"connectors/big.yaml": fmt.Sprintf(`
type: connector
driver: mcp
url: %s
`, ext.URL),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	var names []any
	for i := range 130 {
		names = append(names, ai.ExternalToolName("big", fmt.Sprintf("tool_%d", i)))
	}
	script := &scriptedAIService{turns: []turnFunc{
		toolCallTurn(ai.SearchExternalToolsName, map[string]any{"names": names}),
		textTurn("done"),
	}}
	runner := ai.NewRunner(rt, activity.NewNoopClient())
	session, err := runner.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, Claims: newTestClaims("u@x.com", "x.com", false), UserAgent: "rill/test"})
	require.NoError(t, err)
	session.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
		return script, func() {}, nil
	})

	// More tools than the providers accept no longer fail the request: loading too many is a tool error the LLM can recover from.
	res, err := session.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{Prompt: "hi"})
	require.NoError(t, err)
	results := callResults(session, res.Call.ID)
	require.Equal(t, ai.MessageContentTypeError, results[0].ContentType)
	require.Contains(t, results[0].Content, "limit of 100 loaded external tools")
	require.False(t, slices.ContainsFunc(script.tools[1], isExternalToolName))
}

func TestExternalToolsLimitCountsUnavailableTools(t *testing.T) {
	ext := newBigExternalMCPServer(t, 60)
	connector := fmt.Sprintf(`
type: connector
driver: mcp
url: %s
`, ext.URL)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": "", "connectors/a.yaml": connector, "connectors/b.yaml": connector},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	names := func(connector string) []any {
		var res []any
		for i := range 60 {
			res = append(res, ai.ExternalToolName(connector, fmt.Sprintf("tool_%d", i)))
		}
		return res
	}
	script := &scriptedAIService{turns: []turnFunc{
		toolCallTurn(ai.SearchExternalToolsName, map[string]any{"names": names("a")}),
		textTurn("first"),
		toolCallTurn(ai.SearchExternalToolsName, map[string]any{"names": names("b")}),
		textTurn("second"),
	}}
	runner := ai.NewRunner(rt, activity.NewNoopClient())
	claims := newTestClaims("u@x.com", "x.com", false)
	newTurnSession := func(sessionID string) *ai.Session {
		session, err := runner.Session(t.Context(), &ai.SessionOptions{InstanceID: instanceID, SessionID: sessionID, CreateIfNotExists: sessionID == "", Claims: claims, UserAgent: "rill/test"})
		require.NoError(t, err)
		session.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
			return script, func() {}, nil
		})
		return session
	}

	// The first turn loads all of a's tools.
	session := newTurnSession("")
	res, err := session.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{Prompt: "first"})
	require.NoError(t, err)
	require.Equal(t, ai.MessageContentTypeJSON, callResults(session, res.Call.ID)[0].ContentType)
	require.NoError(t, session.Flush(t.Context()))

	// After a becomes unavailable, its loaded tools still count: otherwise loading b's tools would succeed,
	// and all 120 would be offered once a is available again.
	testruntime.PutFiles(t, rt, instanceID, map[string]string{"connectors/a.yaml": connector + `
security:
  access: false
`})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	session = newTurnSession(session.ID())
	res, err = session.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, ai.AnalystAgentArgs{Prompt: "second"})
	require.NoError(t, err)
	results := callResults(session, res.Call.ID)
	require.Equal(t, ai.MessageContentTypeError, results[0].ContentType)
	require.Contains(t, results[0].Content, "limit of 100 loaded external tools")
}

// newBigExternalMCPServer serves an in-process MCP server with n tools named tool_0 to tool_<n-1>.
func newBigExternalMCPServer(t *testing.T, n int) *httptest.Server {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "big", Version: "1.0.0"}, nil)
	for i := range n {
		mcp.AddTool(server, &mcp.Tool{Name: fmt.Sprintf("tool_%d", i)}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{}, nil, nil
		})
	}
	ext := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(ext.Close)
	return ext
}

// toolCallTurn makes the model call a tool.
func toolCallTurn(name string, input map[string]any) turnFunc {
	return func(_ *drivers.CompleteOptions) *aiv1.CompletionMessage {
		in, err := structpb.NewStruct(input)
		if err != nil {
			panic(err)
		}
		return &aiv1.CompletionMessage{
			Role:    "assistant",
			Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: uuid.NewString(), Name: name, Input: in}}}},
		}
	}
}

// callResults returns the results of the external and search_external_tools calls within the given agent call, in order.
func callResults(s *ai.Session, callID string) []*ai.Message {
	external := func(m *ai.Message) bool { return m.Tool == ai.SearchExternalToolsName || isExternalToolName(m.Tool) }
	var res []*ai.Message
	for _, call := range s.Messages(ai.FilterByParent(callID), ai.FilterByType(ai.MessageTypeCall), external) {
		if result, ok := s.Message(ai.FilterByParent(call.ID), ai.FilterByType(ai.MessageTypeResult)); ok {
			res = append(res, result)
		}
	}
	return res
}

func isExternalToolName(name string) bool {
	return strings.HasPrefix(name, ai.ExternalToolPrefix)
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
