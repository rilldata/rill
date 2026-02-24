package observability

import (
	"context"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/querytrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
)

var _ oteltrace.SpanProcessor = (*QueryLogSpanProcessor)(nil)

// QueryLogSpanProcessor is a trace.SpanProcessor that captures sql.conn.query spans
// (created by otelsql) and routes them to request-scoped query trace collectors.
type QueryLogSpanProcessor struct {
	collectors sync.Map // spanID (trace.SpanID) → *querytrace.Collector
}

// NewQueryLogSpanProcessor creates a new QueryLogSpanProcessor.
func NewQueryLogSpanProcessor() *QueryLogSpanProcessor {
	return &QueryLogSpanProcessor{}
}

// OnStart extracts the collector from the parent context and stores it keyed by spanID.
// It also enriches the span with queue duration from the query origin context.
func (p *QueryLogSpanProcessor) OnStart(parent context.Context, s oteltrace.ReadWriteSpan) {
	if s.Name() != "sql.conn.query" {
		return
	}

	collector, ok := querytrace.FromContext(parent)
	if !ok {
		return
	}

	p.collectors.Store(s.SpanContext().SpanID(), collector)

	// Enrich span with queue duration from context (set by OLAP drivers after connection acquisition)
	queueDurationMs, ok := queueDurationFromContext(parent)
	if ok && queueDurationMs > 0 {
		s.SetAttributes(attribute.Int64("queue_duration_ms", queueDurationMs))
	}
}

// OnEnd processes completed spans, extracting query data and recording it to the collector.
func (p *QueryLogSpanProcessor) OnEnd(s oteltrace.ReadOnlySpan) {
	if s.Name() != "sql.conn.query" {
		return
	}

	spanID := s.SpanContext().SpanID()
	val, ok := p.collectors.LoadAndDelete(spanID)
	if !ok {
		return
	}
	collector := val.(*querytrace.Collector)

	var sql, errMsg string
	var queueDurationMs int64
	for _, attr := range s.Attributes() {
		switch string(attr.Key) {
		case "db.statement":
			sql = attr.Value.AsString()
		case "queue_duration_ms":
			queueDurationMs = attr.Value.AsInt64()
		}
	}

	failed := s.Status().Code == codes.Error
	if failed {
		errMsg = s.Status().Description
	}

	collector.Record(&runtimev1.QueryTrace{
		Sql:             sql,
		DurationMs:      s.EndTime().Sub(s.StartTime()).Milliseconds(),
		QueueDurationMs: queueDurationMs,
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

type queueDurationKey struct{}

// WithQueueDuration stores the queue duration (time spent waiting for a connection) in the context.
func WithQueueDuration(ctx context.Context, queueDurationMs int64) context.Context {
	return context.WithValue(ctx, queueDurationKey{}, queueDurationMs)
}

// queueDurationFromContext extracts the queue duration from the context.
func queueDurationFromContext(ctx context.Context) (int64, bool) {
	queueDurationMs, ok := ctx.Value(queueDurationKey{}).(int64)
	return queueDurationMs, ok
}
