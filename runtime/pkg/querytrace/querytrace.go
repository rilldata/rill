package querytrace

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Collector accumulates query traces during a request.
// NOTE: Not thread-safe. Add a sync.Mutex if concurrent SQL queries are ever fired within a single request.
type Collector struct {
	entries []*runtimev1.QueryTrace
}

// Record adds a query trace entry.
func (c *Collector) Record(e *runtimev1.QueryTrace) {
	c.entries = append(c.entries, e)
}

// ToProto returns the collected traces.
func (c *Collector) ToProto() []*runtimev1.QueryTrace {
	return c.entries
}

type contextKey struct{}

// WithCollector returns a new context with the given collector.
func WithCollector(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

// FromContext extracts the collector from the context.
func FromContext(ctx context.Context) (*Collector, bool) {
	c, ok := ctx.Value(contextKey{}).(*Collector)
	return c, ok
}
