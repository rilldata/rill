package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const httpShutdownTimeout = 15 * time.Second

type Options struct {
	Port     int
	CertPath string
	KeyPath  string
}

// ServeHTTP serves a HTTP server and performs a graceful shutdown if/when ctx is cancelled.
func ServeHTTP(ctx context.Context, server *http.Server, options Options) error {
	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", options.Port))
	if err == nil {
		lis.Close()
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", options.Port))
	}
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("http port %d is in use by another process. Either kill that process or pass `--port PORT` to run Rill on another port", options.Port)
		}
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	var serveErr error
	go func() {
		if options.CertPath != "" && options.KeyPath != "" {

			serveErr = server.ServeTLS(lis, options.CertPath, options.KeyPath)
		} else {
			serveErr = server.Serve(lis)
		}

		cancel()
	}()

	<-cctx.Done()
	if serveErr == nil {
		// server.Serve always returns a non-nil err, so this must be a cancel on the parent ctx.
		// We perform a graceful shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()
		serveErr = server.Shutdown(ctx)
	}

	return serveErr
}
