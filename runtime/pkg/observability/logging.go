package observability

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapCtx returns a Zap field that adds "trace_id" and "span_id" fields to the log message
func ZapCtx(ctx context.Context) zap.Field {
	span := trace.SpanFromContext(ctx)
	sctx := span.SpanContext()
	if !sctx.IsValid() {
		return zap.Skip()
	}

	return zap.Inline(zapSpan{
		traceID: sctx.TraceID().String(),
		spanID:  sctx.SpanID().String(),
	})
}

type zapSpan struct {
	traceID string
	spanID  string
}

func (z zapSpan) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("trace_id", z.traceID)
	enc.AddString("span_id", z.spanID)
	enc.AddString("dd.trace_id", convertToDatadogID(z.traceID))
	enc.AddString("dd.span_id", convertToDatadogID(z.spanID))
	return nil
}

// convertToDatadogID returns a Datadog compatible 64bit unsigned int version of the OpenTelemetry 128bit unsigned int ID,
// adapted from https://docs.datadoghq.com/tracing/other_telemetry/connect_logs_and_traces/opentelemetry
func convertToDatadogID(otelID string) string {
	if len(otelID) < 16 {
		return ""
	}
	if len(otelID) > 16 {
		otelID = otelID[16:]
	}
	intValue, err := strconv.ParseUint(otelID, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(intValue, 10)
}
