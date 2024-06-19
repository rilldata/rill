package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"
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
	config := &net.ListenConfig{Control: noReusePort}
	lis, err := config.Listen(ctx, "tcp", fmt.Sprintf(":%d", options.Port))
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

func noReusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 0)
	})
}
