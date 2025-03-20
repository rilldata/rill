package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
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
	var mu sync.Mutex
	var lis net.Listener
	var err error

	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	mu.Lock()
	lis, err = net.Listen("tcp4", fmt.Sprintf(":%d", options.Port))
	mu.Unlock()
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
	serverStopped := make(chan struct{})
	serverError := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		defer close(serverStopped)
		var err error
		if options.CertPath != "" && options.KeyPath != "" {
			// Use HTTPS if cert and key are provided
			err = server.ServeTLS(lis, options.CertPath, options.KeyPath)
		} else {
			// Otherwise use HTTP
			err = server.Serve(lis)
		}
		if err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	// Wait for context cancellation or server stopped
	select {
	case <-ctx.Done():
		// Shutdown the server with a timeout context
		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()

		mu.Lock()
		err = server.Shutdown(shutdownCtx)
		mu.Unlock()

		// Wait for the server to stop or the shutdown timeout
		select {
		case <-serverStopped:
		case <-shutdownCtx.Done():
		}
		return err
	case err := <-serverError:
		return err
	case <-serverStopped:
		return nil
	}
}
