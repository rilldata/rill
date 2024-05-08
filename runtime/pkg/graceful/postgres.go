package graceful

import (
	"context"
	"fmt"
	"net"
	"strings"

	wire "github.com/jeroenrinzema/psql-wire"
	"go.uber.org/zap"
)

// ServePostgres serves a Postgres server and performs a graceful shutdown if/when ctx is cancelled.
func ServePostgres(ctx context.Context,
	queryHandler func(ctx context.Context, query string) (wire.PreparedStatements, error),
	authHandler func(ctx context.Context, username, password string) (context.Context, bool, error),
	port int,
	logger *zap.Logger) error {

	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
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

	server, err := wire.NewServer(queryHandler)
	if err != nil {
		return err
	}

	return server.Serve(lis)
}
