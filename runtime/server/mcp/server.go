package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the MCP server
type Server struct {
	mcpServer *server.DefaultServer
}

// NewServer creates a new MCP server
func NewServer() *Server {
	s := &Server{
		mcpServer: server.NewDefaultServer("rill-mcp", "0.1.0"),
	}

	// Register tool handlers
	s.mcpServer.HandleCallTool(s.handleCallTool)
	s.mcpServer.HandleListTools(s.handleListTools)

	return s
}

// Start starts the MCP server
func (s *Server) Start() error {
	return server.ServeStdio(s.mcpServer)
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	// TODO: Implement graceful shutdown
	return nil
}

// handleCallTool handles tool calls
func (s *Server) handleCallTool(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// TODO: Implement tool dispatch logic
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Tool called: " + name,
			},
		},
	}, nil
}

// handleListTools returns the list of available tools
func (s *Server) handleListTools(ctx context.Context, cursor *string) (*mcp.ListToolsResult, error) {
	tools := []mcp.Tool{
		{
			Name:        "rill",
			Description: "Rill data tool",
			InputSchema: mcp.ToolInputSchema{Type: "object"},
		},
		{
			Name:        "visualization",
			Description: "Visualization tool",
			InputSchema: mcp.ToolInputSchema{Type: "object"},
		},
	}
	return &mcp.ListToolsResult{Tools: tools}, nil
}
