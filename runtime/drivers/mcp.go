package drivers

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrMCPAuthorizationRequired is returned by an MCPService when the server requires a per-user OAuth authorization
// that the caller has not supplied, or that the server rejected.
var ErrMCPAuthorizationRequired = errors.New("authorization required")

// MCPService enables AI agents to discover and call tools on an external MCP (Model Context Protocol) server.
// Handles are shared by all users of an instance; anything specific to the calling user travels in MCPCallOptions.
type MCPService interface {
	// OAuthConfig returns the server's per-user OAuth configuration, with templating resolved, or nil if the server doesn't use per-user OAuth.
	// Authorizations are bound to its Hash: callers that hold a token must compare the hash before supplying the token.
	OAuthConfig() *MCPOAuthConfig
	// ListTools lists the tools the server exposes to the caller described by opts.
	// Implementations may cache the result for a short period.
	ListTools(ctx context.Context, opts *MCPCallOptions) ([]*MCPTool, error)
	// CallTool invokes a tool on the server on behalf of the caller described by opts.
	// A failure reported by the tool itself is returned as a result with IsError set, not as an error.
	CallTool(ctx context.Context, name string, args map[string]any, opts *MCPCallOptions) (*MCPToolResult, error)
}

// MCPOAuthConfig is the per-user OAuth configuration of an MCP server.
type MCPOAuthConfig struct {
	ServerURL    string
	ClientID     string
	ClientSecret string
	Scopes       []string
	// AuthorizationParams are extra query parameters for the authorization endpoint.
	AuthorizationParams map[string]string
	// Hash identifies the configuration. It changes when any other field changes, which invalidates existing authorizations.
	Hash string
}

// MCPCallOptions describes the caller of an MCP request.
type MCPCallOptions struct {
	// Environment and Variables resolve templating in the connector's headers, such as {{ .env.TOKEN }}.
	Environment string
	Variables   map[string]string
	// UserAttributes are the attributes of the user on whose behalf the request is made.
	// They are exposed to header templates as {{ .user }}.
	UserAttributes map[string]any
	// OAuthToken is the user's OAuth access token for servers that require per-user OAuth.
	// The driver holds no OAuth state of its own: without a token, such servers report ErrMCPAuthorizationRequired.
	OAuthToken string
}

// MCPTool describes a tool exposed by an external MCP server.
type MCPTool struct {
	Name         string
	Title        string
	Description  string
	InputSchema  map[string]any
	OutputSchema map[string]any
	// Annotations describe the tool's behavior as claimed by the server.
	Annotations MCPToolAnnotations
}

// MCPToolAnnotations are the behavior hints a server declares for a tool, with the protocol's defaults applied where it declares none.
// They are claims of the server, not guarantees: callers must only act on them for servers they trust.
type MCPToolAnnotations struct {
	// ReadOnly is true if the tool does not modify its environment.
	ReadOnly bool
	// Destructive is true if the tool may perform destructive updates. It is always false for a read-only tool.
	Destructive bool
	// Idempotent is true if calling the tool repeatedly with the same arguments has no additional effect.
	Idempotent bool
	// OpenWorld is true if the tool may interact with external entities outside a closed domain.
	OpenWorld bool
}

// MCPToolResult is the result of a tool call on an external MCP server.
type MCPToolResult struct {
	// Content is the unstructured result: text blocks concatenated, binary blocks replaced with placeholders.
	Content string
	// StructuredContent is the structured result, if the server returned one.
	StructuredContent json.RawMessage
	// IsError is true if the tool call failed on the server; Content then holds the error message.
	IsError bool
	// Truncated is true if Content was cut to the connector's max_result_bytes.
	Truncated bool
}
