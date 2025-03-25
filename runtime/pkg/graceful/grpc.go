package graceful

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
)

const grpcShutdownTimeout = 15 * time.Second

// ServeGRPC serves a GRPC server and performs a graceful shutdown if/when ctx is cancelled.
func ServeGRPC(ctx context.Context, server *grpc.Server, port int) error {
	var lis net.Listener
	var err error

	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	lis, err = net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err == nil {
		lis.Close()
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("grpc port %d is in use by another process. Either kill that process or pass `--port-grpc PORT` to run Rill on another port", port)
		}
		return err
	}

	// Channel to signal server has stopped
	serveErrCh := make(chan error)
	go func() {
		err := server.Serve(lis)
		serveErrCh <- err
	}()

	// Wait for context to be cancelled or failure to serve
	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
		server.GracefulStop()
	}

	// Wait for graceful shutdown
	select {
	case err := <-serveErrCh:
		return err
	case <-time.After(grpcShutdownTimeout):
		server.Stop()
		return errors.New("grpc graceful shutdown timed out")
	}
}
