package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
)

// mcpInstructions are the instructions advertised to clients of the admin service's MCP server.
// They are a combination of the admin service's own instructions and the instructions from the runtime's AI tools that we proxy to.
const mcpInstructions = `
# Rill Cloud MCP Server
This server provides access to several Rill projects.

1. **List projects:** Use "list_projects" to discover the projects you have access to.
2. **Target a project:** Pass the "project" argument on every other tool call. Do this before the workflow below.

` + ai.MCPInstructions

const (
	// mcpProjectArg is the tool argument the MCP server uses to route a tool call to a project.
	mcpProjectArg        = "project"
	mcpListProjectsLimit = 100

	// mcpToolCallTimeout is deliberately higher than the runtime's own tool timeout,
	// so the runtime's timeout applies first and returns a more useful message.
	mcpToolCallTimeout = 6 * time.Minute
)

// mcpForwardedTools are the runtime tools exposed on the MCP server.
// The runtime exposes more tools than these, but the rest require runtime.EditRepo,
// which is only granted for editable deployments, and a project's primary deployment is never editable.
var mcpForwardedTools = []string{
	ai.ListMetricsViewsName,
	ai.GetMetricsViewName,
	ai.QueryMetricsViewSummaryName,
	ai.QueryMetricsViewName,
}

// mcpSessionNamespace is a UUIDv5 namespace for the per-project session IDs derived in callRuntimeTool.
// The value is a randomly generated UUID with no meaning: UUIDv5 requires a namespace and any fixed value will do.
// It must not be changed, since that would orphan the sessions derived using the previous value.
var mcpSessionNamespace = uuid.MustParse("6f4a1c94-3d1e-4f6b-9a02-1c0f6b7d5e8a")

// mcpHandler creates a handler for the MCP server.
// Unlike the per-project runtime proxy, it serves a single endpoint for all the projects a caller has access to:
// it answers the MCP protocol methods itself and forwards tool calls to the runtime of the project named in the tool arguments.
func (s *Server) mcpHandler() (http.Handler, error) {
	adminTools := s.mcpAdminTools()
	forwardedTools, err := mcpForwardedToolSpecs()
	if err != nil {
		return nil, err
	}
	for _, at := range adminTools {
		for _, ft := range forwardedTools {
			if at.spec.Name == ft.Name {
				return nil, fmt.Errorf("mcp: tool %q is both served by the admin service and forwarded to the runtime", ft.Name)
			}
		}
	}

	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		srv := mcp.NewServer(
			&mcp.Implementation{
				Name:    "rill",
				Title:   "Rill Cloud MCP Server",
				Version: s.admin.Version.String(),
			},
			&mcp.ServerOptions{
				Instructions: mcpInstructions,
				HasTools:     true,
				// Issue a session ID so clients pass it back on subsequent requests, enabling us to keep one AI session per project.
				GetSessionID: uuid.NewString,
			},
		)

		srv.AddReceivingMiddleware(observability.MCPMiddleware())

		// Separate timeout middleware is required since the SDK strips the request's deadline and cancellation before calling the tool handler.
		srv.AddReceivingMiddleware(middleware.TimeoutMCPMiddleware(func(method, tool string) time.Duration {
			return mcpToolCallTimeout
		}))

		client := &mcpClient{
			sessionID:       r.Header.Get("Mcp-Session-Id"),
			userAgent:       r.UserAgent(),
			protocolVersion: r.Header.Get("Mcp-Protocol-Version"),
		}

		for _, t := range adminTools {
			srv.AddTool(t.spec, t.handler)
		}
		for _, t := range forwardedTools {
			srv.AddTool(t, s.mcpForwardToolCall(t.Name, client))
		}

		return srv
	}, &mcp.StreamableHTTPOptions{
		Stateless: true, // Required since the server serves multiple users.
	}), nil
}

// mcpRequireAuth rejects unauthenticated requests to the MCP server.
// It wraps the MCP handler instead of being applied inside it, so that clients receive the challenge on the initialization request and can start an OAuth flow.
func (s *Server) mcpRequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unlike the per-project runtime proxy, we can't accept a runtime JWT here since it's scoped to a single deployment.
		if auth.GetClaims(r.Context()).OwnerType() == auth.OwnerTypeAnon {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", s.admin.URLs.OAuthProtectedResourceMetadata(r)))
			httputil.WriteError(w, httputil.Errorf(http.StatusUnauthorized, "not authenticated"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// mcpAdminTools returns the tools served by the admin service itself.
// These are tools for concepts the runtime is not aware of, such as organizations and projects.
func (s *Server) mcpAdminTools() []mcpAdminTool {
	return []mcpAdminTool{
		{
			spec: &mcp.Tool{
				Name:        "list_projects",
				Title:       "List Projects",
				Description: `List the Rill projects you have access to. Pass a returned value as the "project" argument of other tools. The list may be incomplete; if you already know a project, you can pass it directly.`,
				Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
				InputSchema: &jsonschema.Schema{Type: "object"},
			},
			handler: s.mcpListProjects,
		},
	}
}

// mcpAdminTool is a tool served by the admin service itself instead of being forwarded to a project's runtime.
type mcpAdminTool struct {
	spec    *mcp.Tool
	handler mcp.ToolHandler
}

// mcpForwardedToolSpecs returns the specs of the tools forwarded to a project's runtime, using the tool definitions in runtime/ai as the source of truth.
// It injects the "project" argument into the input schema of each tool, which we intercept to route calls to the correct runtime.
func mcpForwardedToolSpecs() ([]*mcp.Tool, error) {
	specs := ai.MCPToolSpecs()

	res := make([]*mcp.Tool, 0, len(mcpForwardedTools))
	for _, name := range mcpForwardedTools {
		spec, ok := specs[name]
		if !ok {
			return nil, fmt.Errorf("mcp: no spec for tool %q", name)
		}

		// Inject the "project" argument into the input schema.
		schema, ok := spec.InputSchema.(*jsonschema.Schema)
		if !ok {
			return nil, fmt.Errorf("mcp: tool %q does not have a JSON schema", name)
		}
		if schema.Properties == nil {
			schema.Properties = make(map[string]*jsonschema.Schema)
		}
		schema.Properties[mcpProjectArg] = &jsonschema.Schema{
			Type:        "string",
			Description: `The project to target, as returned by "list_projects".`,
		}
		schema.Required = append(schema.Required, mcpProjectArg)

		res = append(res, spec)
	}

	return res, nil
}

// mcpClient describes a proxy client making a request from the admin service to a runtime.
type mcpClient struct {
	sessionID       string
	userAgent       string
	protocolVersion string
}

// mcpForwardToolCall returns a handler that forwards a tool call to the runtime of the project named in its arguments.
func (s *Server) mcpForwardToolCall(name string, client *mcpClient) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Note we return tool errors instead of errors for anything the caller can act on,
		// since an error is returned to the client as a protocol error it can't recover from.
		org, project, args, err := takeMCPProjectArg(req.Params.Arguments)
		if err != nil {
			return mcpToolErrorf(`%s. Call "list_projects" to discover available projects.`, err), nil
		}

		proj, depl, err := s.resolveDeploymentForOrgAndProject(ctx, org, project, "")
		if err != nil {
			return mcpToolErrorf("failed to find project %q: %s", org+"/"+project, err), nil
		}

		jwt, err := s.issueEphemeralRuntimeToken(ctx, proj, depl, runtimeProxyAccessTokenTTL)
		if err != nil {
			if errors.Is(err, errNoProdAccess) {
				return mcpToolErrorf("you do not have access to project %q", org+"/"+project), nil
			}
			return nil, err
		}

		s.admin.Used.Deployment(depl.ID)

		return s.callRuntimeTool(ctx, depl, jwt, client, name, args)
	}
}

// callRuntimeTool calls a tool on a deployment's MCP server.
// It sends a plain JSON-RPC request instead of using the MCP SDK's client, which would add an initialization
// handshake to every tool call and would not let us pass through the session ID and user agent.
func (s *Server) callRuntimeTool(ctx context.Context, depl *database.Deployment, jwt string, client *mcpClient, name string, args json.RawMessage) (*mcp.CallToolResult, error) {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  mcp.CallToolParams{Name: name, Arguments: args},
	})
	if err != nil {
		return nil, err
	}

	reqURL, err := url.Parse(runtimeHTTPHost(depl.RuntimeHost))
	if err != nil {
		return nil, err
	}
	reqURL = reqURL.JoinPath("/v1/instances", depl.RuntimeInstanceID, "/mcp")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream") // The runtime requires both, even though it responds with JSON
	req.Header.Set("Authorization", "Bearer "+jwt)
	if client.protocolVersion != "" {
		req.Header.Set("Mcp-Protocol-Version", client.protocolVersion)
	}
	if client.userAgent != "" {
		req.Header.Set("User-Agent", client.userAgent)
	}
	if client.sessionID != "" {
		// Derive a session ID per project instead of passing the client's through.
		// A runtime may serve several projects from one catalog, where AI session IDs must be unique.
		sessionID := uuid.NewSHA1(mcpSessionNamespace, []byte(depl.RuntimeInstanceID+"/"+client.sessionID))
		req.Header.Set("Mcp-Session-Id", sessionID.String())
	}

	// Uses http.DefaultClient to get caching/pooling of TCP connections.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("runtime returned status %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}

	// The runtime responds with a single JSON-RPC message, or an array of them if it also emitted notifications.
	msgs := []json.RawMessage{data}
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		if err := json.Unmarshal(data, &msgs); err != nil {
			return nil, err
		}
	}
	for _, msg := range msgs {
		var rpc struct {
			Result *mcp.CallToolResult `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(msg, &rpc); err != nil {
			return nil, err
		}
		if rpc.Error != nil {
			return nil, errors.New(rpc.Error.Message)
		}
		if rpc.Result != nil {
			return rpc.Result, nil
		}
	}
	return nil, errors.New("runtime did not return a result")
}

// mcpListProjects lists the projects the caller has access to.
// It is a discovery convenience, not an access check: it may return fewer projects than the caller can access
// (for example, it does not cover public projects), and each tool call checks access for the project it targets.
func (s *Server) mcpListProjects(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	claims := auth.GetClaims(ctx)

	var projs []*database.Project
	switch claims.OwnerType() {
	case auth.OwnerTypeUser:
		var err error
		projs, err = s.admin.DB.FindProjectsForUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, err
		}
	case auth.OwnerTypeService:
		svc, err := s.admin.DB.FindService(ctx, claims.OwnerID())
		if err != nil {
			return nil, err
		}
		// A service belongs to exactly one organization.
		// If it can manage the organization's projects, it has access to all of them, not just those it's a member of.
		if claims.OrganizationPermissions(ctx, svc.OrgID).ManageProjects {
			projs, err = s.admin.DB.FindProjectsForOrganization(ctx, svc.OrgID, "", mcpListProjectsLimit)
			if err != nil {
				return nil, err
			}
			break
		}
		members, err := s.admin.DB.FindProjectMemberServicesForService(ctx, svc.ID)
		if err != nil {
			return nil, err
		}
		for _, m := range members {
			proj, err := s.admin.DB.FindProject(ctx, m.ProjectID)
			if err != nil {
				return nil, err
			}
			projs = append(projs, proj)
		}
	case auth.OwnerTypeMagicAuthToken:
		tkn, ok := claims.AuthTokenModel().(*database.MagicAuthToken)
		if !ok {
			return nil, errors.New("unexpected auth token model")
		}
		proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
		if err != nil {
			return nil, err
		}
		projs = append(projs, proj)
	default:
		return nil, fmt.Errorf("listing projects is not supported for %s tokens", claims.OwnerType())
	}

	if len(projs) > mcpListProjectsLimit {
		projs = projs[:mcpListProjectsLimit]
	}

	// Build the "<organization>/<project>" slugs used for the "project" tool argument.
	res := struct {
		Projects []string `json:"projects"`
		Note     string   `json:"note,omitempty"`
	}{Projects: make([]string, 0, len(projs))}
	orgNames := make(map[string]string)
	for _, proj := range projs {
		name, ok := orgNames[proj.OrganizationID]
		if !ok {
			org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
			if err != nil {
				return nil, err
			}
			name = org.Name
			orgNames[proj.OrganizationID] = name
		}
		res.Projects = append(res.Projects, name+"/"+proj.Name)
	}
	if len(res.Projects) == mcpListProjectsLimit {
		res.Note = fmt.Sprintf("Only the first %d projects are shown.", mcpListProjectsLimit)
	}

	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(data)}},
		StructuredContent: res,
	}, nil
}

// takeMCPProjectArg removes the project argument from raw tool arguments and returns the organization and project it names.
// It must be removed because the runtime's tools do not accept it.
func takeMCPProjectArg(args json.RawMessage) (org, project string, rest json.RawMessage, err error) {
	var m map[string]json.RawMessage
	if len(args) > 0 {
		if err := json.Unmarshal(args, &m); err != nil {
			return "", "", nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	raw, ok := m[mcpProjectArg]
	if !ok {
		return "", "", nil, fmt.Errorf("missing the %q argument", mcpProjectArg)
	}
	var slug string
	if err := json.Unmarshal(raw, &slug); err != nil {
		return "", "", nil, fmt.Errorf("invalid %q argument: %w", mcpProjectArg, err)
	}
	org, project, ok = strings.Cut(slug, "/")
	if !ok || org == "" || project == "" {
		return "", "", nil, fmt.Errorf("invalid %q argument %q: expected the format \"<organization>/<project>\"", mcpProjectArg, slug)
	}
	delete(m, mcpProjectArg)

	rest, err = json.Marshal(m)
	if err != nil {
		return "", "", nil, err
	}
	return org, project, rest, nil
}

// mcpToolErrorf returns a tool call result representing an error.
// Unlike an error returned from a tool handler, it is presented to the model as a result it can act on.
func mcpToolErrorf(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}
