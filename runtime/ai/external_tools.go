package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// This file integrates external MCP servers (connectors with driver "mcp") into AI sessions.
// The servers' tools are discovered once per session for the session's user, gated by the connector's security policy,
// and compiled into CompiledTools that the agents can call like built-in tools.
// External tools are never registered with Rill's own MCP server: doing so would turn Rill into a token-forwarding proxy.

const (
	// ExternalToolPrefix prefixes the names of tools contributed by external MCP servers.
	ExternalToolPrefix = "mcp__"
	// ErrorCodeMCPAuthorizationRequired is the error code in the JSON error emitted when a user must authorize a server.
	ErrorCodeMCPAuthorizationRequired = "mcp_authorization_required"
	// maxExternalCallsPerTurn caps calls to external servers within one top-level call, such as one chat message.
	maxExternalCallsPerTurn = 25
	// maxExternalDescriptionLen caps the length of tool descriptions from external servers before they enter the prompt.
	maxExternalDescriptionLen = 2000
	// maxToolNameLen is the longest tool name accepted by the LLM providers.
	maxToolNameLen = 64
	// mcpDriverName is the driver name of connectors for external MCP servers.
	mcpDriverName = "mcp"
)

// ExternalServer describes an external MCP server from the perspective of a specific user.
type ExternalServer struct {
	Connector    string
	DisplayName  string
	Instructions string
	URL          string
	// Agents the server is scoped to. Empty means all agents.
	Agents []string
	// OAuth is true if the server requires a per-user OAuth authorization.
	OAuth bool
	// Connected is true if the server can be used by the user. It is always true for reachable servers without OAuth.
	Connected bool
	// Tools the server exposes to the user, after applying the connector's security policy.
	Tools []*drivers.MCPTool
	// Compiled are the Tools compiled for use by the agents, keyed by their prefixed name.
	Compiled map[string]*CompiledTool
	// Err is set if the server could not be reached or its tools could not be listed.
	Err error

	res   *runtimev1.Resource
	token string // the user's OAuth access token in hosted deployments
	mu    sync.Mutex
}

// AllowsAgent reports whether the server may be used by the named agent.
func (s *ExternalServer) AllowsAgent(agent string) bool {
	return len(s.Agents) == 0 || slices.Contains(s.Agents, agent)
}

// ExternalAuthorizationRequiredError is returned by an external tool when the user must (re)authorize the server.
// Its message is a JSON envelope: the chat UI recognizes it to show a Connect card, and the LLM can read the explanation.
type ExternalAuthorizationRequiredError struct {
	Connector   string `json:"connector"`
	DisplayName string `json:"display_name"`
}

func (e *ExternalAuthorizationRequiredError) Error() string {
	data, _ := json.Marshal(map[string]any{
		"code":         ErrorCodeMCPAuthorizationRequired,
		"connector":    e.Connector,
		"display_name": e.DisplayName,
		"message":      fmt.Sprintf("The user must connect %s before its tools can be used. Ask them to click Connect in the chat, then try again once they confirm.", e.DisplayName),
	})
	return string(data)
}

// ExternalToolResult is the result of an external tool call as seen by the LLM and the chat UI.
type ExternalToolResult struct {
	Content           string          `json:"content,omitempty"`
	StructuredContent json.RawMessage `json:"structured_content,omitempty"`
	Truncated         bool            `json:"truncated,omitempty"`
}

// externalToolsState holds a session's discovered external servers.
type externalToolsState struct {
	once    sync.Once
	servers []*ExternalServer
	tools   map[string]*CompiledTool
}

// ExternalServers returns the external MCP servers available to the session's user, discovering them on first use.
// Discovery is best-effort: a server that cannot be reached is returned with Err set and no tools.
func (s *BaseSession) ExternalServers(ctx context.Context) []*ExternalServer {
	s.external.once.Do(func() {
		servers, err := s.runner.DiscoverExternalServers(ctx, s.instanceID, s.claims)
		if err != nil {
			if !errors.Is(err, ctx.Err()) {
				s.logger.Warn("failed to discover external MCP servers", zap.Error(err), observability.ZapCtx(ctx))
			}
			return
		}
		tools := make(map[string]*CompiledTool)
		for _, srv := range servers {
			for name, t := range srv.Compiled {
				tools[name] = t
			}
		}
		s.external.servers = servers
		s.external.tools = tools
	})
	return s.external.servers
}

// ExternalToolNames returns the names of the external tools available to the given agent, sorted for stable prompts.
func (s *BaseSession) ExternalToolNames(ctx context.Context, agent string) []string {
	var names []string
	for _, srv := range s.ExternalServers(ctx) {
		if !srv.Connected || !srv.AllowsAgent(agent) {
			continue
		}
		for name := range srv.Compiled {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// ExternalServerPromptData returns template data describing the external servers for the given agent's prompt.
func (s *BaseSession) ExternalServerPromptData(ctx context.Context, agent string) map[string]any {
	var connected, unconnected []map[string]any
	for _, srv := range s.ExternalServers(ctx) {
		if !srv.AllowsAgent(agent) {
			continue
		}
		entry := map[string]any{
			"display_name": srv.DisplayName,
			"instructions": strings.TrimSpace(srv.Instructions),
			"prefix":       ExternalToolName(srv.Connector, ""),
		}
		if srv.Connected && len(srv.Compiled) > 0 {
			connected = append(connected, entry)
		} else if srv.OAuth && !srv.Connected {
			unconnected = append(unconnected, entry)
		}
	}
	return map[string]any{
		"connected":   connected,
		"unconnected": unconnected,
	}
}

// externalTool returns a compiled external tool by its prefixed name, if it has been discovered.
func (s *BaseSession) externalTool(name string) (*CompiledTool, bool) {
	t, ok := s.external.tools[name]
	return t, ok
}

// DiscoverExternalServers lists the external MCP servers the user may use, listing and compiling their tools.
// Servers the user is denied by the connector's security policy are omitted entirely.
func (r *Runner) DiscoverExternalServers(ctx context.Context, instanceID string, claims *runtime.SecurityClaims) ([]*ExternalServer, error) {
	ctrl, err := r.Runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	resources, err := ctrl.List(ctx, runtime.ResourceKindConnector, "", false)
	if err != nil {
		return nil, err
	}
	var candidates []*runtimev1.Resource
	for _, res := range resources {
		if res.GetConnector().GetSpec().GetDriver() == mcpDriverName && res.Meta.DeletedOn == nil {
			candidates = append(candidates, res)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	slices.SortFunc(candidates, func(a, b *runtimev1.Resource) int {
		return strings.Compare(a.Meta.Name.Name, b.Meta.Name.Name)
	})

	inst, err := r.Runtime.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)

	servers := make([]*ExternalServer, len(candidates))
	grp, gctx := errgroup.WithContext(ctx)
	grp.SetLimit(4)
	for i, res := range candidates {
		grp.Go(func() error {
			srv, err := r.discoverExternalServer(gctx, instanceID, inst.Environment, vars, claims, res)
			if err != nil {
				return err
			}
			servers[i] = srv
			return nil
		})
	}
	if err := grp.Wait(); err != nil {
		return nil, err
	}

	res := make([]*ExternalServer, 0, len(servers))
	for _, srv := range servers {
		if srv != nil {
			res = append(res, srv)
		}
	}
	return res, nil
}

// discoverExternalServer builds the ExternalServer for one connector. It returns nil if the user may not use the server.
// Only context errors are returned as errors; every other failure is recorded on the server so one broken connector cannot hide the rest.
func (r *Runner) discoverExternalServer(ctx context.Context, instanceID, environment string, vars map[string]string, claims *runtime.SecurityClaims, res *runtimev1.Resource) (*ExternalServer, error) {
	spec := res.GetConnector().Spec
	props := spec.Properties.AsMap()
	srv := &ExternalServer{
		Connector: res.Meta.Name.Name,
		Agents:    spec.Agents,
		res:       res,
	}
	srv.DisplayName, _ = props["display_name"].(string)
	if srv.DisplayName == "" {
		srv.DisplayName = srv.Connector
	}
	srv.Instructions, _ = props["ai_instructions"].(string)
	auth, _ := props["auth"].(string)
	srv.OAuth = auth == "oauth"
	// The URL may be templated ({{ .env.MCP_URL }}); hosted tokens are bound to the resolved URL, so resolve it the way the driver does.
	cfg, err := r.Runtime.ConnectorConfig(ctx, instanceID, srv.Connector)
	if err != nil {
		srv.Err = fmt.Errorf("failed to resolve the connector's configuration: %w", err)
		return srv, nil
	}
	srv.URL, _ = cfg.Resolve()["url"].(string)

	// Apply the connector's access policy: denied servers are invisible.
	sec, err := r.Runtime.ResolveConnectorSecurity(ctx, instanceID, claims, res, nil)
	if err != nil {
		return externalServerPolicyError(ctx, srv, claims, err)
	}
	if !sec.CanAccess() {
		return nil, nil
	}

	// OAuth servers need an identified user to authorize as.
	if srv.OAuth && claims.UserID == "" && !claims.SkipChecks {
		return nil, nil
	}

	if res.Meta.ReconcileError != "" {
		srv.Err = errors.New(res.Meta.ReconcileError)
		return srv, nil
	}

	// Resolve the user's OAuth token in hosted deployments. Rill Developer keeps it in the connector's local state instead.
	if srv.OAuth {
		token, hosted, err := r.externalOAuthToken(ctx, instanceID, srv, claims.UserID, false)
		if err != nil {
			if errors.Is(err, ctx.Err()) {
				return nil, err
			}
			srv.Err = err
			return srv, nil
		}
		if hosted && token == "" {
			// The user has not authorized the server yet.
			return srv, nil
		}
		srv.token = token
	}

	opts := &drivers.MCPCallOptions{
		Environment:    environment,
		Variables:      vars,
		UserAttributes: claims.UserAttributes,
		OAuthToken:     srv.token,
	}
	tools, err := r.listExternalTools(ctx, instanceID, srv, claims.UserID, opts)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			return nil, err
		}
		if errors.Is(err, drivers.ErrMCPAuthorizationRequired) {
			// Not connected (or the token was rejected). The Connect flow will fix it; not an error.
			return srv, nil
		}
		srv.Err = err
		return srv, nil
	}
	srv.Connected = true

	// Apply the connector's per-tool rules, with the tool names playing the role of fields.
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name)
	}
	sec, err = r.Runtime.ResolveConnectorSecurity(ctx, instanceID, claims, res, names)
	if err != nil {
		return externalServerPolicyError(ctx, srv, claims, err)
	}
	srv.Compiled = make(map[string]*CompiledTool)
	for _, t := range tools {
		if !sec.CanAccessField(t.Name) {
			continue
		}
		srv.Tools = append(srv.Tools, t)
		ct := r.compileExternalTool(instanceID, environment, vars, srv, t)
		srv.Compiled[ct.Name] = ct
	}
	return srv, nil
}

// externalServerPolicyError handles a security policy that could not be evaluated for the user.
// Access cannot be decided, so the server stays hidden from the user rather than failing the whole discovery.
// Admins see the server with the error, so the policy gets fixed instead of silently excluding people.
func externalServerPolicyError(ctx context.Context, srv *ExternalServer, claims *runtime.SecurityClaims, err error) (*ExternalServer, error) {
	if errors.Is(err, ctx.Err()) {
		return nil, err
	}
	if !claims.Admin() {
		return nil, nil
	}
	srv.Connected = false
	srv.Tools = nil
	srv.Compiled = nil
	srv.Err = fmt.Errorf("failed to evaluate the connector's security policy: %w", err)
	return srv, nil
}

// listExternalTools lists a server's tools, refreshing a rejected hosted token once.
func (r *Runner) listExternalTools(ctx context.Context, instanceID string, srv *ExternalServer, userID string, opts *drivers.MCPCallOptions) ([]*drivers.MCPTool, error) {
	handle, release, err := r.Runtime.AcquireHandle(ctx, instanceID, srv.Connector)
	if err != nil {
		return nil, err
	}
	defer release()
	svc, ok := handle.AsMCP()
	if !ok {
		return nil, fmt.Errorf("connector %q is not an MCP server", srv.Connector)
	}
	tools, err := svc.ListTools(ctx, opts)
	if err != nil && errors.Is(err, drivers.ErrMCPAuthorizationRequired) && opts.OAuthToken != "" {
		if token, err2 := r.refreshExternalOAuthToken(ctx, instanceID, srv, userID); err2 == nil && token != "" {
			opts.OAuthToken = token
			tools, err = svc.ListTools(ctx, opts)
		}
	}
	return tools, err
}

// externalOAuthToken fetches the user's OAuth token for a server from the admin service.
// The token is bound to the server URL it was granted for, so a connector whose URL changed gets no token until the user connects again.
// hosted is false in Rill Developer, where tokens are persisted by the driver instead and the returned token is empty.
func (r *Runner) externalOAuthToken(ctx context.Context, instanceID string, srv *ExternalServer, userID string, forceRefresh bool) (token string, hosted bool, err error) {
	admin, release, err := r.Runtime.Admin(ctx, instanceID)
	if err != nil {
		// No admin connector: Rill Developer.
		return "", false, nil
	}
	defer release()
	tok, err := admin.GetMCPOAuthToken(ctx, srv.Connector, srv.URL, userID, forceRefresh)
	if err != nil {
		if errors.Is(err, drivers.ErrNotImplemented) {
			return "", false, nil
		}
		if errors.Is(err, drivers.ErrNotFound) {
			return "", true, nil
		}
		return "", true, fmt.Errorf("failed to get the user's authorization for %q: %w", srv.Connector, err)
	}
	return tok.AccessToken, true, nil
}

// refreshExternalOAuthToken force-refreshes the user's hosted token after the server rejected it.
func (r *Runner) refreshExternalOAuthToken(ctx context.Context, instanceID string, srv *ExternalServer, userID string) (string, error) {
	token, hosted, err := r.externalOAuthToken(ctx, instanceID, srv, userID, true)
	if err != nil || !hosted {
		return "", err
	}
	srv.mu.Lock()
	srv.token = token
	srv.mu.Unlock()
	return token, nil
}

// compileExternalTool compiles an external tool so agents can call it.
func (r *Runner) compileExternalTool(instanceID, environment string, vars map[string]string, srv *ExternalServer, t *drivers.MCPTool) *CompiledTool {
	name := ExternalToolName(srv.Connector, t.Name)
	title := t.Title
	if title == "" {
		title = t.Name
	}
	spec := &mcp.Tool{
		Name:        name,
		Title:       fmt.Sprintf("%s: %s", srv.DisplayName, title),
		Description: sanitizeExternalDescription(t.Description),
		InputSchema: normalizeInputSchema(t.InputSchema),
		Meta: map[string]any{
			"rill/mcpServer":            srv.Connector,
			"rill/mcpServerDisplayName": srv.DisplayName,
			"rill/mcpToolName":          t.Name,
		},
	}
	toolName := t.Name
	connector := srv.Connector

	return &CompiledTool{
		Name: name,
		Spec: spec,
		CheckAccess: func(ctx context.Context) (bool, error) {
			s := GetSession(ctx)
			if !s.Claims().Can(runtime.UseAI) {
				return false, nil
			}
			// Re-evaluate the policy on every call, so a user whose attributes changed loses access immediately.
			names := make([]string, 0, len(srv.Tools))
			for _, st := range srv.Tools {
				names = append(names, st.Name)
			}
			sec, err := r.Runtime.ResolveConnectorSecurity(ctx, s.InstanceID(), s.Claims(), srv.res, names)
			if err != nil {
				return false, err
			}
			return sec.CanAccessField(toolName), nil
		},
		UnmarshalArgs: func(content string) (any, error) {
			var args map[string]any
			if err := json.Unmarshal([]byte(content), &args); err != nil {
				return nil, err
			}
			return args, nil
		},
		UnmarshalResult: func(content string) (any, error) {
			var res ExternalToolResult
			if err := json.Unmarshal([]byte(content), &res); err != nil {
				return nil, err
			}
			return res, nil
		},
		JSONHandler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
			s := GetSession(ctx)
			if err := s.checkExternalCallBudget(); err != nil {
				return nil, err
			}
			var args map[string]any
			if len(input) > 0 {
				if err := json.Unmarshal(input, &args); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
			}

			handle, release, err := r.Runtime.AcquireHandle(ctx, instanceID, connector)
			if err != nil {
				return nil, err
			}
			defer release()
			svc, ok := handle.AsMCP()
			if !ok {
				return nil, fmt.Errorf("connector %q is not an MCP server", connector)
			}

			srv.mu.Lock()
			token := srv.token
			srv.mu.Unlock()
			opts := &drivers.MCPCallOptions{
				Environment:    environment,
				Variables:      vars,
				UserAttributes: s.Claims().UserAttributes,
				OAuthToken:     token,
			}
			res, err := svc.CallTool(ctx, toolName, args, opts)
			if err != nil && errors.Is(err, drivers.ErrMCPAuthorizationRequired) && token != "" {
				// The hosted token was rejected: refresh it once and retry.
				if newToken, err2 := r.refreshExternalOAuthToken(ctx, instanceID, srv, s.Claims().UserID); err2 == nil && newToken != "" {
					opts.OAuthToken = newToken
					res, err = svc.CallTool(ctx, toolName, args, opts)
				}
			}
			s.activity.Record(ctx, "log", "mcp_tool_call",
				attribute.String("connector", connector),
				attribute.String("tool", toolName),
				attribute.Bool("error", err != nil || (res != nil && res.IsError)),
			)
			if err != nil {
				if errors.Is(err, drivers.ErrMCPAuthorizationRequired) {
					return nil, &ExternalAuthorizationRequiredError{Connector: connector, DisplayName: srv.DisplayName}
				}
				return nil, fmt.Errorf("%s: %w", srv.DisplayName, err)
			}
			if res.IsError {
				return nil, fmt.Errorf("%s returned an error: %s", srv.DisplayName, res.Content)
			}
			return json.Marshal(&ExternalToolResult{
				Content:           res.Content,
				StructuredContent: res.StructuredContent,
				Truncated:         res.Truncated,
			})
		},
		// External tools are never exposed through Rill's own MCP server.
		RegisterWithMCPServer: func(srv *mcp.Server) {},
	}
}

// checkExternalCallBudget enforces the per-turn cap on external tool calls.
func (s *Session) checkExternalCallBudget() error {
	if s.ParentID == "" {
		return nil
	}
	var count int
	for _, m := range s.MessagesWithDescendents(FilterByID(s.RootID())) {
		if m.Type == MessageTypeCall && strings.HasPrefix(m.Tool, ExternalToolPrefix) {
			count++
		}
	}
	// The current call has already been recorded, so count includes it.
	if count > maxExternalCallsPerTurn {
		return fmt.Errorf("the limit of %d external tool calls per message has been reached; answer with the information gathered so far", maxExternalCallsPerTurn)
	}
	return nil
}

// ExternalToolName returns the prefixed name used for a tool of an external server.
// Names are limited to the characters and length LLM providers accept.
// Sanitizing is lossy ("get.item" and "get_item" both become "get_item"), so whenever either part changes under sanitizing,
// or contains the "__" separator, a hash of the original identifiers is appended to keep distinct tools distinct.
// Over-long names are shortened in front of the same hash.
// With an empty tool, it returns the prefix shared by all of the server's tools.
func ExternalToolName(connector, tool string) string {
	sanitizedConnector := sanitizeToolName(connector)
	name := ExternalToolPrefix + sanitizedConnector + "__"
	if tool == "" {
		return name
	}
	sanitizedTool := sanitizeToolName(tool)
	name += sanitizedTool

	lossy := sanitizedConnector != connector || sanitizedTool != tool || strings.Contains(connector, "__") || strings.Contains(tool, "__")
	if !lossy && len(name) <= maxToolNameLen {
		return name
	}
	// The NUL separator makes the hash unambiguous for identifiers that themselves contain the separator.
	sum := sha256.Sum256([]byte(connector + "\x00" + tool))
	suffix := "_" + hex.EncodeToString(sum[:])[:8]
	if len(name)+len(suffix) > maxToolNameLen {
		name = name[:maxToolNameLen-len(suffix)]
	}
	return name + suffix
}

// sanitizeToolName lowercases a name and replaces characters LLM providers do not accept in tool names.
func sanitizeToolName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

// sanitizeExternalDescription cleans a tool description from an external server before it enters the prompt.
// Descriptions are untrusted input, so control characters are stripped and the length is capped.
func sanitizeExternalDescription(desc string) string {
	var b strings.Builder
	for _, r := range desc {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	res := strings.TrimSpace(b.String())
	if len(res) > maxExternalDescriptionLen {
		res = res[:maxExternalDescriptionLen] + "…"
	}
	return res
}

// normalizeInputSchema makes an input schema acceptable to all LLM providers: an object with explicit properties.
func normalizeInputSchema(schema map[string]any) map[string]any {
	if schema == nil {
		schema = map[string]any{}
	}
	if _, ok := schema["type"]; !ok {
		schema["type"] = "object"
	}
	if _, ok := schema["properties"]; !ok {
		schema["properties"] = map[string]any{}
	}
	return schema
}
