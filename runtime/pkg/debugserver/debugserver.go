package debugserver

import (
	"context"
	"net/http"

	"github.com/rilldata/rill/runtime/pkg/graceful"

	// Register /debug/pprof/* endpoints on http.DefaultServeMux
	_ "net/http/pprof"
)

const port = 6060

func ServeHTTP(ctx context.Context) error {
	srv := &http.Server{} // An empty server will serve http.DefaultServeMux
	return graceful.ServeHTTP(ctx, srv, port)
}
