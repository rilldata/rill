package observability

import (
	"context"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
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

var _ oteltrace.SpanProcessor = (*RequestScopedSpanProcessor)(nil)

// RequestScopedSpanProcessor is a trace.SpanProcessor that captures spans and routes them to request-scoped collectors.
// This only works when a collector is present in the parent context, which is only set when req.trace is set on runtimev1 query requests.
type RequestScopedSpanProcessor struct {
	mu         sync.Mutex
	collectors map[trace.SpanID]*RequestScopedCollector
}

// NewRequestScopedSpanProcessor creates a new RequestScopedSpanProcessor.
func NewRequestScopedSpanProcessor() *RequestScopedSpanProcessor {
	return &RequestScopedSpanProcessor{
		collectors: make(map[trace.SpanID]*RequestScopedCollector),
	}
}

// OnStart extracts the collector from the parent context and stores it keyed by spanID.
func (p *RequestScopedSpanProcessor) OnStart(parent context.Context, s oteltrace.ReadWriteSpan) {
	collector, ok := RequestScopedCollectorFromContext(parent)
	if !ok {
		return
	}

	p.mu.Lock()
	p.collectors[s.SpanContext().SpanID()] = collector
	p.mu.Unlock()
}

// OnEnd processes completed spans, building a generic Span proto and recording it to the collector.
func (p *RequestScopedSpanProcessor) OnEnd(s oteltrace.ReadOnlySpan) {
	spanID := s.SpanContext().SpanID()

	p.mu.Lock()
	collector, ok := p.collectors[spanID]
	if ok {
		delete(p.collectors, spanID)
	}
	p.mu.Unlock()

	if !ok {
		return
	}

	// Build attributes map from span attributes
	attrs := make(map[string]string)
	for _, attr := range s.Attributes() {
		attrs[string(attr.Key)] = attributeValueToString(attr.Value)
	}

	// Determine parent span ID
	parentSpanID := ""
	if s.Parent().HasSpanID() {
		parentSpanID = s.Parent().SpanID().String()
	}

	collector.Record(&runtimev1.Span{
		Name:         s.Name(),
		SpanId:       spanID.String(),
		ParentSpanId: parentSpanID,
		StartTime:    timestamppb.New(s.StartTime()),
		DurationMs:   s.EndTime().Sub(s.StartTime()).Milliseconds(),
		Attributes:   attrs,
	})
}

// Shutdown is a no-op.
func (p *RequestScopedSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush is a no-op.
func (p *RequestScopedSpanProcessor) ForceFlush(ctx context.Context) error {
	return nil
}

// attributeValueToString converts an OTel attribute value to its string representation.
func attributeValueToString(v attribute.Value) string {
	switch v.Type() {
	case attribute.STRING:
		return v.AsString()
	case attribute.BOOL:
		return fmt.Sprintf("%t", v.AsBool())
	case attribute.INT64:
		return fmt.Sprintf("%d", v.AsInt64())
	case attribute.FLOAT64:
		return fmt.Sprintf("%g", v.AsFloat64())
	default:
		return v.Emit()
	}
}
