package graceful

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const httpShutdownTimeout = 5 * time.Second

type ServeOptions struct {
	Port     int
	GRPCPort int
	CertPath string
	KeyPath  string
	Logger   *zap.Logger
}

// ServeHTTP serves a HTTP server and performs a graceful shutdown if/when ctx is cancelled.
func ServeHTTP(ctx context.Context, handler http.Handler, opts ServeOptions) error {
	// By default, we serve HTTP with h2c on the main HTTP port, enabling both HTTP/1.1 and HTTP/2 cleartext on that port.
	// However, if a dedicated gRPC port is configured, we only use h2c on that port, not the main HTTP port.
	// If TLS is configured, we only serve HTTP/2.
	useTLS := opts.CertPath != "" && opts.KeyPath != ""
	useDedicatedGRPC := opts.GRPCPort != 0 && opts.Port != opts.GRPCPort
	servers := map[int]*http.Server{}
	if useTLS {
		// If TLS options are passed, configure a single HTTP/2 server.
		if useDedicatedGRPC {
			return errors.New("cannot combine TLS options with a dedicated gRPC port")
		}
		s := &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.Port),
			Handler: handler,
		}
		err := http2.ConfigureServer(s, &http2.Server{})
		if err != nil {
			return err
		}
		servers[opts.Port] = s
	} else if useDedicatedGRPC {
		// Use h2c on the gRPC port and normal HTTP on the main port.
		servers[opts.Port] = &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.Port),
			Handler: handler,
		}
		servers[opts.GRPCPort] = &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.GRPCPort),
			Handler: h2c.NewHandler(handler, &http2.Server{}),
		}
	} else {
		// Use h2c on the main HTTP port.
		servers[opts.Port] = &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.Port),
			Handler: h2c.NewHandler(handler, &http2.Server{}),
		}
	}

	// Start the servers in the background
	serveErrCh := make(chan error, len(servers)) // Channel to signal a server has stopped
	for port, srv := range servers {
		if opts.Logger != nil {
			opts.Logger.Sugar().Infof("serving HTTP on port:%v", port)
		}

		// Start the listener and server in the background
		go func() {
			// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
			// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
			lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
			if err == nil {
				lis.Close()
				lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
			}
			if err != nil {
				if strings.Contains(err.Error(), "address already in use") {
					err = fmt.Errorf("http port %d is in use by another process; either kill that process or pass `--port PORT` to run Rill on another port", port)
				}
				serveErrCh <- err
				return
			}

			// Serve the server
			if opts.CertPath != "" && opts.KeyPath != "" {
				err = srv.ServeTLS(lis, opts.CertPath, opts.KeyPath)
			} else {
				err = srv.Serve(lis)
			}
			serveErrCh <- err
		}()
	}

	// Wait for context cancellation or a server stopped
	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()
		for _, srv := range servers {
			err := srv.Shutdown(ctx)
			if err != nil && !errors.Is(err, ctx.Err()) {
				// Ignoring context errors because they are quite frequent with MCP streaming connections, which don't stop gracefully.
				return err
			}
		}
		return nil
	}
}
