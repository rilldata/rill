package motherduck

import (
	"context"

	"github.com/jmoiron/sqlx"
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
