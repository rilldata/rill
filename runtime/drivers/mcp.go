package drivers

import (
	"context"
	"encoding/json"
)

// MCPService enables AI agents to discover and call tools on an external MCP (Model Context Protocol) server.
type MCPService interface {
	// ListTools lists the tools the server exposes.
	// Implementations may cache the result for a short period.
	ListTools(ctx context.Context) ([]*MCPTool, error)
	// CallTool invokes a tool on the server.
	// A failure reported by the tool itself is returned as a result with IsError set, not as an error.
	CallTool(ctx context.Context, name string, args map[string]any) (*MCPToolResult, error)
}

// MCPTool describes a tool exposed by an external MCP server.
type MCPTool struct {
	Name         string
	Title        string
	Description  string
	InputSchema  map[string]any
	OutputSchema map[string]any
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
