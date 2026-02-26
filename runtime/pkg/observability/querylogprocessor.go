package observability

import (
	"context"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
)

var _ oteltrace.SpanProcessor = (*QueryLogSpanProcessor)(nil)

// QueryLogSpanProcessor is a trace.SpanProcessor that captures spans and routes them to request-scoped query trace collectors.
type QueryLogSpanProcessor struct {
	collectors sync.Map // spanID (trace.SpanID) → *Collector
}

// NewQueryLogSpanProcessor creates a new QueryLogSpanProcessor.
func NewQueryLogSpanProcessor() *QueryLogSpanProcessor {
	return &QueryLogSpanProcessor{}
}

// OnStart extracts the collector from the parent context and stores it keyed by spanID.
func (p *QueryLogSpanProcessor) OnStart(parent context.Context, s oteltrace.ReadWriteSpan) {
	collector, ok := CollectorFromContext(parent)
	if !ok {
		return
	}

	p.collectors.Store(s.SpanContext().SpanID(), collector)
}

// OnEnd processes completed spans, building a generic Span proto and recording it to the collector.
func (p *QueryLogSpanProcessor) OnEnd(s oteltrace.ReadOnlySpan) {
	spanID := s.SpanContext().SpanID()
	val, ok := p.collectors.LoadAndDelete(spanID)
	if !ok {
		return
	}
	collector := val.(*Collector)

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

	// Build error info
	failed := s.Status().Code == codes.Error
	var errMsg string
	if failed {
		errMsg = s.Status().Description
	}

	collector.Record(&runtimev1.Span{
		Name:            s.Name(),
		SpanId:          spanID.String(),
		ParentSpanId:    parentSpanID,
		StartTimeUnixMs: s.StartTime().UnixMilli(),
		DurationMs:      s.EndTime().Sub(s.StartTime()).Milliseconds(),
		Attributes:      attrs,
		Failed:          failed,
		Error:           errMsg,
	})
}

// Shutdown is a no-op.
func (p *QueryLogSpanProcessor) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush is a no-op.
func (p *QueryLogSpanProcessor) ForceFlush(ctx context.Context) error {
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
