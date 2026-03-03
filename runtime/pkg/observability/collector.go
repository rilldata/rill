package observability

import (
	"context"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// RequestScopedCollector accumulates span traces during a request.
type RequestScopedCollector struct {
	mu      sync.Mutex
	entries []*runtimev1.Span
}

// Record adds a span trace entry.
func (c *RequestScopedCollector) Record(e *runtimev1.Span) {
	c.mu.Lock()
	c.entries = append(c.entries, e)
	c.mu.Unlock()
}

// ToProto returns the collected traces as Trace proto.
func (c *RequestScopedCollector) ToProto() *runtimev1.Trace {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) == 0 {
		return nil
	}
	return &runtimev1.Trace{Spans: c.entries}
}

type collectorContextKey struct{}

// WithRequestScopedCollector returns a new context with the given collector.
func WithRequestScopedCollector(ctx context.Context, c *RequestScopedCollector) context.Context {
	return context.WithValue(ctx, collectorContextKey{}, c)
}

// RequestScopedCollectorFromContext extracts the collector from the context.
func RequestScopedCollectorFromContext(ctx context.Context) (*RequestScopedCollector, bool) {
	c, ok := ctx.Value(collectorContextKey{}).(*RequestScopedCollector)
	return c, ok
}
