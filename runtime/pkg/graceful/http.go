package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// ServeHTTP serves a HTTP server and performs a graceful shutdown if/when ctx is cancelled
func ServeHTTP(ctx context.Context, server *http.Server, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	var serveErr error
	go func() {
		serveErr = server.Serve(lis)
		cancel()
	}()

	<-cctx.Done()
	if serveErr == nil {
		// server.Serve always returns a non-nil err, so this must be a cancel on the parent ctx.
		// We perform a graceful shutdown.
		serveErr = server.Shutdown(context.Background())
	}

	return serveErr
}
