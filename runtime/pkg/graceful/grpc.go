package graceful

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

const grpcShutdownTimeout = 15 * time.Second

// ServeGRPC serves a GRPC server and performs a graceful shutdown if/when ctx is cancelled.
func ServeGRPC(ctx context.Context, server *grpc.Server, port int) error {
	var mu sync.Mutex
	var lis net.Listener
	var err error

	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	mu.Lock()
	lis, err = net.Listen("tcp4", fmt.Sprintf(":%d", port))
	mu.Unlock()
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
	serverStopped := make(chan struct{})
	serverError := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		defer close(serverStopped)
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			serverError <- err
		}
	}()

	// Wait for context cancellation or server stopped
	select {
	case <-ctx.Done():
		// Gracefully stop the server
		mu.Lock()
		server.GracefulStop()
		mu.Unlock()

		// Wait for the server to actually stop
		select {
		case <-serverStopped:
		case <-time.After(grpcShutdownTimeout):
			mu.Lock()
			server.Stop()
			mu.Unlock()
			<-serverStopped
		}
		return ctx.Err()
	case err := <-serverError:
		return err
	case <-serverStopped:
		return nil
	}
}
