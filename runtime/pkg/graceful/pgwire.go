package graceful

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/rilldata/rill/runtime/pkg/pgwire"
	"go.uber.org/zap"
)

// ServePGWire serves a PostgreSQL wire-compatible server and shuts it down when
// ctx is cancelled.
func ServePGWire(ctx context.Context, server *pgwire.Server, port int, logger *zap.Logger) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err == nil {
		_ = listener.Close()
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("psql port %d is in use by another process; either kill that process or configure another pgwire port", port)
		}
		return err
	}
	if logger != nil {
		logger.Info("serving PostgreSQL wire protocol", zap.Int("port", port))
	}
	return server.Serve(ctx, listener)
}
