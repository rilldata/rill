package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// MuxHandle is a wrapper around http.ServeMux.Handle that adds route tags to the handler.
// It does NOT wrap the handler with observability.Middleware. The caller is expected to add that on the ServeMux itself.
func MuxHandle(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, otelhttp.WithRouteTag(pattern, handler))
}
