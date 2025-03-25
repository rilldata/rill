package graceful

import (
	"context"
	"errors"
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
	go func() {
		var serveErr error
		if options.CertPath != "" && options.KeyPath != "" {
			serveErr = server.ServeTLS(lis, options.CertPath, options.KeyPath)
		} else {
			serveErr = server.Serve(lis)
		}
		serveErrCh <- serveErr
	}()

	// Wait for context to be cancelled or failure to serve
	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
		// Create a separate context for shutdown with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()

		// Initiate graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		// Wait for server to complete shutdown
		select {
		case err := <-serveErrCh:
			if err != http.ErrServerClosed {
				return err
			}
			return nil
		case <-shutdownCtx.Done():
			return errors.New("http graceful shutdown timed out")
		}
	}
}
