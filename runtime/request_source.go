package runtime

import "context"

// RequestSource identifies the origin of a request. It is used to tag billable usage metrics so that
// downstream billing can distinguish interactive dashboard usage (not billed) from programmatic access.
type RequestSource string

const (
	// RequestSourceUI is interactive dashboard traffic (and direct gRPC query traffic, which shares the same path).
	RequestSourceUI RequestSource = "ui"
	// RequestSourceAPI is the REST custom API.
	RequestSourceAPI RequestSource = "api"
	// RequestSourceMCP is the MCP server.
	RequestSourceMCP RequestSource = "mcp"
	// RequestSourceAlert is alert execution.
	RequestSourceAlert RequestSource = "alert"
	// RequestSourceReport is report execution.
	RequestSourceReport RequestSource = "report"
	// RequestSourceChat is conversational AI (chat) completion, over both gRPC and the SSE HTTP handler.
	RequestSourceChat RequestSource = "chat"
)

type requestSourceKey struct{}

// WithRequestSource returns a context tagged with the request's source.
func WithRequestSource(ctx context.Context, src RequestSource) context.Context {
	return context.WithValue(ctx, requestSourceKey{}, src)
}

// RequestSourceFromContext returns the request source set on the context, or "" if none was set.
func RequestSourceFromContext(ctx context.Context) RequestSource {
	src, _ := ctx.Value(requestSourceKey{}).(RequestSource)
	return src
}
