package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) newMCPHandler(basePath string) http.Handler {
	mcpServer := server.NewMCPServer("rill-runtime", "1.0.0", server.WithToolCapabilities(false))

	tool := mcp.NewTool("echo",
		mcp.WithDescription("Echoes the message back to the caller"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("The message to echo"),
		),
	)

	mcpServer.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		claims := auth.GetClaims(ctx)
		return mcp.NewToolResultText(fmt.Sprintf("Echo to subject %q: %v", claims.Subject(), req.GetArguments()["message"])), nil
	})

	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithStaticBasePath(basePath),
		server.WithUseFullURLForMessageEndpoint(false),
	)

	return sseServer
}
