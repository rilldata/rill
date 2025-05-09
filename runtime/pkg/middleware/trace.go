package middleware

import (
	"net/http"

	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/trace"
)

// traceMiddleware adds trace ID headers to the response if a trace is in progress
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			w.Header().Set(observability.TracingHeader, traceID)
		}
		next.ServeHTTP(w, r)
	})
}
