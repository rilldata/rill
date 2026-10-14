package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

// mcpChartDataTimeout bounds chart_data calls, which the chart widget makes interactively.
const mcpChartDataTimeout = 45 * time.Second

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

		// Requests without a session ID normally start a conversation (initialize, or a tool call from a proxy that keeps its own sessions).
		// Any other sessionless request (resources/list, tools/list, ping, ...) is not part of a conversation,
		// so it gets a transient session that is never persisted instead of leaving an orphan session behind.
		transient := sessionID == "" && !mcpRequestStartsSession(r)

		// It's just preliminary: for direct connections, the MCP server updates it with the actual user agent after the initialization handshake.
		// Requests proxied by the admin service's unified MCP server don't carry that handshake, so the forwarded user agent is the only signal of the real client.
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "mcp/unknown"
		}

		// Create session
		sess, err := runner.Session(r.Context(), &ai.SessionOptions{
			InstanceID:        instanceID,
			SessionID:         sessionID,
			CreateIfNotExists: true,
			Transient:         transient,
			Claims:            auth.GetClaims(r.Context(), instanceID),
			UserAgent:         userAgent,
		})
		if err != nil {
			s.logger.Error("failed to create AI session for MCP", zap.String("instance_id", instanceID), zap.String("session_id", sessionID), zap.Error(err))
			return nil
		}

		// Create MCP server for the session with middleware.
		srv := sess.MCPServer(r.Context())
		srv.AddReceivingMiddleware(observability.MCPMiddleware())
		srv.AddReceivingMiddleware(mcpRequestSourceMiddleware())
		srv.AddReceivingMiddleware(middleware.TimeoutMCPMiddleware(func(method, tool string) time.Duration {
			// Sets an upper limit, but note that some tools enforce shorter timeouts in their implementation.
			if tool == ai.ChartDataName {
				return mcpChartDataTimeout
			}
			return 5 * time.Minute
		}))

		return srv
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
		// Respond with application/json instead of text/event-stream.
		// It only affects the response: as the spec requires, requests must still accept both content types.
		// It lets the admin service's unified MCP server forward tool calls with a plain HTTP request instead of parsing an SSE stream.
		JSONResponse: true,
	})
}

// mcpRequestPeekBytes bounds how much of a request body mcpRequestStartsSession reads to find the JSON-RPC method.
// Protocol requests are tiny; only tool calls (such as write_file with a large file) can exceed it.
const mcpRequestPeekBytes = 4 << 20

// mcpRequestStartsSession reports whether a request without a session ID should create a persistent AI session:
// an "initialize" handshake or a "tools/call" (which proxies such as the admin service's unified MCP server send without a handshake).
// It peeks at the JSON-RPC method in a bounded prefix of the body and stitches the body back together for the MCP handler.
// A GET request never starts a conversation. A body it cannot classify (larger than the prefix, malformed, or without a method)
// is treated as starting a session: a persistent session is at worst an orphan, whereas a transient one would lose the messages of a real tool call.
func mcpRequestStartsSession(r *http.Request) bool {
	if r.Method != http.MethodPost || r.Body == nil {
		return false
	}
	prefix, err := io.ReadAll(io.LimitReader(r.Body, mcpRequestPeekBytes))
	r.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(prefix), r.Body), r.Body}
	if err != nil {
		return true
	}
	var msg struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(prefix, &msg); err != nil || msg.Method == "" {
		return true
	}
	return msg.Method == "initialize" || msg.Method == "tools/call"
}

// mcpRequestSourceMiddleware tags MCP requests with the "mcp" source so billable queries they trigger
// are attributed to MCP (programmatic access).
func mcpRequestSourceMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceMCP)
			return next(ctx, method, req)
		}
	}
}
