package clickhouse

import (
	"context"
	"maps"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/pkg/observability"
)

// connCtxKey is used as the key when saving a connection in a context
type connCtxKey struct{}

// contextWithConnection returns a wrapped context that contains the connection
func contextWithConn(ctx context.Context, conn *sqlx.Conn) context.Context {
	return context.WithValue(ctx, connCtxKey{}, conn)
}

// connFromContext retrieves a connection wrapped with contextWithConn.
// If no connection is in the context, it returns nil.
func connFromContext(ctx context.Context) *sqlx.Conn {
	conn := ctx.Value(connCtxKey{})
	if conn != nil {
		return conn.(*sqlx.Conn)
	}
	return nil
}

// sessionAwareContext sets a session_id in context which is used to tie queries to a certain session.
// This is used to use certain session aware features like temporary tables.
func (c *connection) sessionAwareContext(ctx context.Context) context.Context {
	if c.opts.Protocol == clickhouse.HTTP {
		var settings map[string]any
		if len(c.opts.Settings) == 0 {
			settings = make(map[string]any)
		} else {
			settings = maps.Clone(c.opts.Settings)
		}
		settings["session_id"] = uuid.New().String()
		return clickhouse.Context(ctx, clickhouse.WithSettings(settings))
	}
	// native protocol already has session context
	return ctx
}

func contextWithQueryID(ctx context.Context) context.Context {
	traceID := observability.DatadogTraceID(ctx)
	if traceID == "" {
		return ctx
	}
	return clickhouse.Context(ctx, clickhouse.WithQueryID(traceID))
}
