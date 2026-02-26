package observability

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Collector accumulates span traces during a request.
// NOTE: Not thread-safe. Add a sync.Mutex if concurrent SQL queries are ever fired within a single request.
type Collector struct {
	entries []*runtimev1.Span
}

// Record adds a span trace entry.
func (c *Collector) Record(e *runtimev1.Span) {
	c.entries = append(c.entries, e)
}

// ToProto returns the collected traces as a TraceDetails proto.
func (c *Collector) ToProto() *runtimev1.Trace {
	if len(c.entries) == 0 {
		return nil
	}
	return &runtimev1.Trace{Spans: c.entries}
}

type collectorContextKey struct{}

// WithCollector returns a new context with the given collector.
func WithCollector(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, collectorContextKey{}, c)
}

// CollectorFromContext extracts the collector from the context.
func CollectorFromContext(ctx context.Context) (*Collector, bool) {
	c, ok := ctx.Value(collectorContextKey{}).(*Collector)
	return c, ok
}
