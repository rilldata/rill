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

type ServeOptions struct {
	Port     int
	CertPath string
	KeyPath  string
}

// ServeHTTP serves a HTTP server and performs a graceful shutdown if/when ctx is cancelled.
func ServeHTTP(ctx context.Context, server *http.Server, options ServeOptions) error {
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

	// Channel to signal server has stopped
	serveErrCh := make(chan error)
	// Start server in a goroutine
	go func() {
		if options.CertPath != "" && options.KeyPath != "" {
			// Use HTTPS if cert and key are provided
			err := server.ServeTLS(lis, options.CertPath, options.KeyPath)
			serveErrCh <- err
		} else {
			// Otherwise use HTTP
			err := server.Serve(lis)
			serveErrCh <- err
		}
	}()

	// Wait for context cancellation or server stopped
	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()

		return server.Shutdown(ctx)
	}
}
