package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rilldata/rill/runtime"
)

func (s *Server) newMCPClient(mcpServer *server.MCPServer) *client.Client {
	client, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		panic(err)
	}

	// Start the client with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		panic(fmt.Errorf("failed to start MCP client: %w", err))
	}

	// Try to initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "rill",
		Version: "0.0.1",
	}

	if _, err := client.Initialize(ctx, initRequest); err != nil {
		panic(fmt.Errorf("failed to initialize MCP client: %w", err))
	}

	return client
}

func (s *Server) mcpListTools(ctx context.Context, instanceID string) ([]runtime.Tool, error) {
	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, instanceID)

	tools, err := s.mcpClient.ListTools(ctxWithInstance, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	runtimeTools := make([]runtime.Tool, len(tools.Tools))
	for i := range tools.Tools {
		tool := &tools.Tools[i]
		runtimeTool := runtime.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}

		// Convert InputSchema to JSON string if present
		if schemaBytes, err := json.Marshal(tool.InputSchema); err == nil && string(schemaBytes) != "{}" && string(schemaBytes) != "null" {
			runtimeTool.InputSchema = string(schemaBytes)
		}

		runtimeTools[i] = runtimeTool
	}

	return runtimeTools, nil
}

func (s *Server) mcpExecuteTool(ctx context.Context, instanceID, toolName string, toolArgs map[string]any) (any, error) {
	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, instanceID)

	resp, err := s.mcpClient.CallTool(ctxWithInstance, mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: toolArgs,
		},
	})
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}
