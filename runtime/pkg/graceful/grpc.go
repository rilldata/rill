package graceful

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// ServeGRPC serves a GRPC server and performs a graceful shutdown if/when ctx is cancelled.
func ServeGRPC(ctx context.Context, server *grpc.Server, port int) error {
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
		server.GracefulStop()
	}

	return serveErr
}
