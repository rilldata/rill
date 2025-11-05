package server

import (
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

// mcpHandler creates an MCP server handler.
// It uses a new implementation that replaces the logic in mcp_server.go.
func (s *Server) mcpHandler() http.Handler {
	runner := ai.NewRunner(s.runtime, s.activity)

	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		// Extract instance ID from the request path
		instanceID := r.PathValue("instance_id")
		if instanceID == "" {
			// We also mount the MCP server on <root>/mcp to make it easier to use in Rill Developer (on localhost).
			// In those settings, we pick the default instance ID.
			// This is safe because if there is no default instance, it'll just be the empty string and requests will error with "not found".
			instanceID, _ = s.runtime.DefaultInstanceID()
		}

		// Get session ID (will be empty if it's the first request)
		sessionID := r.Header.Get("Mcp-Session-Id")

		// Create session
		sess, err := runner.Session(r.Context(), &ai.SessionOptions{
			InstanceID:        instanceID,
			SessionID:         sessionID,
			CreateIfNotExists: true,
			Claims:            auth.GetClaims(r.Context(), instanceID),
			UserAgent:         "mcp/unknown", // It's just preliminary: the MCP server updates it with the actual user agent after the initialization handshake.
		})
		if err != nil {
			s.logger.Error("failed to create AI session for MCP", zap.String("instance_id", instanceID), zap.String("session_id", sessionID), zap.Error(err))
			return nil
		}

		// Create MCP server for the session with middleware.
		srv := sess.MCPServer(r.Context())
		srv.AddReceivingMiddleware(observability.MCPMiddleware())
		srv.AddReceivingMiddleware(middleware.TimeoutMCPMiddleware(func(method, tool string) time.Duration {
			// Sets an upper limit, but note that some tools enforce shorter timeouts in their implementation.
			return 5 * time.Minute
		}))

		return srv
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
}
