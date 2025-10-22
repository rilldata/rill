package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

func (s *Server) newMCPClient(mcpServer *server.MCPServer) (*client.Client, error) {
	client, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		return nil, err
	}

	// Start the client with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Try to initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "rill",
		Version: "0.0.1",
	}

	if _, err := client.Initialize(ctx, initRequest); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	return client, nil
}

func (s *Server) mcpListTools(ctx context.Context, instanceID string) ([]*aiv1.Tool, error) {
	mcpServer, err := s.newMCPServer(ctx, instanceID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	mcpClient, err := s.newMCPClient(mcpServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer mcpClient.Close()

	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, instanceID)

	tools, err := mcpClient.ListTools(ctxWithInstance, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	aiTools := make([]*aiv1.Tool, len(tools.Tools))
	for i := range tools.Tools {
		tool := &tools.Tools[i]
		aiTool := &aiv1.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}

		// Convert InputSchema to JSON string if present
		if schemaBytes, err := json.Marshal(tool.InputSchema); err == nil && string(schemaBytes) != "{}" && string(schemaBytes) != "null" {
			aiTool.InputSchema = string(schemaBytes)
		}

		aiTools[i] = aiTool
	}

	return aiTools, nil
}

func (s *Server) mcpExecuteTool(ctx context.Context, instanceID, toolName string, toolArgs map[string]any) (string, error) {
	mcpServer, err := s.newMCPServer(ctx, instanceID, false)
	if err != nil {
		return "", fmt.Errorf("failed to create MCP server: %w", err)
	}
	mcpClient, err := s.newMCPClient(mcpServer)
	if err != nil {
		return "", fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer mcpClient.Close()

	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, instanceID)

	resp, err := mcpClient.CallTool(ctxWithInstance, mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: toolArgs,
		},
	})

	// Handle errors
	if err != nil {
		return "", err
	} else if len(resp.Content) == 0 {
		return "", nil
	} else if len(resp.Content) > 1 {
		return "", fmt.Errorf("multiple content items not supported, got %d items", len(resp.Content))
	}

	// Extract text content from MCP response
	switch content := resp.Content[0].(type) {
	case mcp.TextContent:
		return content.Text, nil
	default:
		return "", fmt.Errorf("unsupported content type: %T", content) // Future work: support other content types
	}
}
