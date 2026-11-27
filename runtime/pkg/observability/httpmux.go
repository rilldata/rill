package observability

import (
	"net/http"

	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

// MuxHandle is a wrapper around http.ServeMux.Handle that sets the route attribute on the active span.
// It does NOT wrap the handler with observability.Middleware. The caller is expected to add that on the ServeMux itself.
//
// otelhttp derives the route from http.Request.Pattern for span names and metrics,
// but sets the span's route attribute when the span starts,
// which is before a nested ServeMux has matched the pattern.
func MuxHandle(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace.SpanFromContext(r.Context()).SetAttributes(semconv.HTTPRoute(pattern))
		handler.ServeHTTP(w, r)
	}))
}
