package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

const mcpInstructions = `
# Rill MCP Server
This server exposes APIs for querying **metrics views**, which represent Rill's metrics layer.

## Workflow Overview
1. **List metrics views:** Use "list_metrics_views" to discover available metrics views in the project.
2. **Get metrics view spec:** Use "get_metrics_view" to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
3. **Query the summary:** Use "query_metrics_view_summary" to obtain the available time range for a metrics view and sample values with their data types for each dimension. This provides a richer context for understanding the data.
4. **Query the metrics:** Use "query_metrics_view" to run queries to get aggregated results.

In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
If a response contains an "ai_instructions" field, you should interpret it as additional instructions for how to behave in subsequent responses that relate to that tool call.
`

// MCPServer returns a new MCP server scoped to the current session.
// Since it is scoped to the session, a new MCP server should be created for each client connection.
// Using a separate MCP server for each client enables tailoring the server's instructions and available tools to the end user's claims.
func (s *Session) MCPServer(ctx context.Context) *mcp.Server {
	// Create the MCP server
	srv := mcp.NewServer(
		&mcp.Implementation{
			Name:    "rill",
			Title:   "Rill MCP Server",
			Version: s.runner.Runtime.Version().String(),
		},
		&mcp.ServerOptions{
			Instructions: mcpInstructions,
			InitializedHandler: func(ctx context.Context, r *mcp.InitializedRequest) {
				// Save user agent in the session
				clientInfo := r.Session.InitializeParams().ClientInfo
				if clientInfo != nil && clientInfo.Name != "" {
					userAgent := clientInfo.Name
					if clientInfo.Version != "" {
						userAgent += "/" + clientInfo.Version
					}

					err := s.UpdateUserAgent(ctx, userAgent)
					if err != nil && !errors.Is(err, ctx.Err()) {
						s.logger.Warn("failed to update user agent", zap.Error(err))
					}
				}
			},
			KeepAlive: 30 * time.Second,
			HasTools:  true,
			GetSessionID: func() string {
				return s.id
			},
		},
	)

	// Inject the Session before every request, and trigger a flush after each request is finished.
	srv.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			ctx = WithSession(ctx, s)
			res, err := next(ctx, method, req)
			flushErr := s.Flush(ctx)
			if flushErr != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to flush session: %w", flushErr))
			}
			return res, err
		}
	})

	// Add only the tools that the user has access to
	ctx = WithSession(ctx, s)
	for _, t := range s.runner.Tools {
		if !t.CheckAccess(ctx) {
			continue
		}
		t.RegisterWithMCPServer(srv)
	}

	return srv
}

// InternalError represents an internal error in a tool call.
// This is needed because by default, downstream logic (such as the MCP middleware) treats errors returned from tool handlers as user errors, not internal errors.
type InternalError struct {
	err error
}

// NewInternalError creates a new internal error. See InternalError for details.
func NewInternalError(err error) error {
	if err == nil {
		err = fmt.Errorf("nil error")
	}
	return InternalError{err: err}
}

// Error implements the error interface.
func (e InternalError) Error() string {
	return fmt.Sprintf("internal: %s", e.err.Error())
}
