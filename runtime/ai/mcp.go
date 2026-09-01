package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/pkg/jsonschemautil"
	"go.uber.org/zap"
)

// MCPInstructions are the instructions advertised by the MCP server.
// It is exported so the unified MCP server in the admin service can extend it instead of restating it.
const MCPInstructions = `
# Rill MCP Server
This server exposes APIs for querying **metrics views**, which represent Rill's metrics layer.

## Workflow Overview
1. **List metrics views:** Use "list_metrics_views" to discover available metrics views in the project.
2. **Get metrics view spec:** Use "get_metrics_view" to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
3. **Query the summary:** Use "query_metrics_view_summary" to obtain the available time range for a metrics view and sample values with their data types for each dimension. This provides a richer context for understanding the data.
4. **Query the metrics:** Use "query_metrics_view" to run queries to get aggregated results.

In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
If a response contains an "ai_instructions" field, you should interpret it as additional instructions for how to behave in subsequent responses that relate to that tool call.

## Project Development
If you have edit access, the server also exposes tools for inspecting and editing the project's source code, which consists of YAML and SQL files:
- **List files:** Use "list_files" to browse the files in the project.
- **Search files:** Use "search_files" to find files by name or content.
- **Read a file:** Use "read_file" to read the contents of a file.
- **Write a file:** Use "write_file" to create, update or delete a file. If the file declares a Rill resource, it returns the resource's status and any errors encountered after reconciliation.
`

// MCPToolSpecs returns the specs of all registered tools, keyed by name.
// The specs are freshly built and owned by the caller, so they may be mutated.
// It is used by the unified MCP server in the admin service, which advertises the tools without being able to run them.
func MCPToolSpecs() map[string]*mcp.Tool {
	// Safe to pass a nil runtime: Spec() does not use it (asserted by TestMCPToolSpecs).
	runner := NewRunner(nil, nil)
	specs := make(map[string]*mcp.Tool, len(runner.Tools))
	for name, t := range runner.Tools {
		specs[name] = t.Spec
	}
	return specs
}

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
			Instructions: MCPInstructions,
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

	// Tolerantly decode tool arguments where object/array-typed fields arrive as JSON-encoded strings.
	// This works around a serialization bug in some MCP clients (see https://github.com/anthropics/claude-code/issues/25865).
	// The coercion runs only as a fallback: if a tool call fails the SDK's input schema validation,
	// the arguments are coerced and the call is retried once; calls that validate as-is are never rewritten.
	// The internal LLM tool-call path (CallToolWithOptions) receives tool inputs as real JSON objects from the provider API,
	// so it does not need this; if that ever changes, the coercion should be applied there too.
	srv.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if method != "tools/call" || err == nil {
				return res, err
			}
			var jsonrpcErr *jsonrpc.Error
			if !errors.As(err, &jsonrpcErr) || jsonrpcErr.Code != jsonrpc.CodeInvalidParams {
				return res, err
			}
			if !s.coerceMCPToolArgs(req) {
				return res, err
			}
			if params, ok := req.GetParams().(*mcp.CallToolParamsRaw); ok {
				s.logger.Debug("tolerantly decoded stringified MCP tool call arguments; retrying the call", zap.String("tool", params.Name))
			}
			return next(ctx, method, req)
		}
	})

	// Add only the tools that the user has access to
	ctx = WithSession(ctx, s)
	for _, t := range s.runner.Tools {
		ok, err := t.CheckAccess(ctx)
		if err != nil {
			if !errors.Is(err, ctx.Err()) {
				s.logger.Error("failed to check tool access", zap.Error(err))
			}
			continue
		}
		if !ok {
			continue
		}
		t.RegisterWithMCPServer(srv)
	}

	return srv
}

// coerceMCPToolArgs rewrites a tool call's raw arguments,
// JSON-decoding any string values in positions where the tool's input schema expects an object or array.
// It returns whether the arguments were rewritten.
// It fails open: on any lookup or decode failure it leaves the arguments untouched,
// so the SDK's schema validation produces its normal error.
func (s *Session) coerceMCPToolArgs(req mcp.Request) bool {
	params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
	if !ok || len(params.Arguments) == 0 {
		return false
	}
	t, ok := s.runner.Tools[params.Name]
	if !ok || t.Spec == nil {
		return false
	}
	schema, ok := t.Spec.InputSchema.(*jsonschema.Schema)
	if !ok || schema == nil {
		return false
	}

	var args any
	dec := json.NewDecoder(bytes.NewReader(params.Arguments))
	dec.UseNumber() // preserve integer fidelity across the re-marshal
	if err := dec.Decode(&args); err != nil {
		return false
	}

	coerced, changed := jsonschemautil.CoerceStringifiedJSON(schema, args)
	if !changed {
		return false
	}
	data, err := json.Marshal(coerced)
	if err != nil {
		return false
	}
	params.Arguments = data
	return true
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
