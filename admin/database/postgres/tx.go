package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"
)

// NewTx starts a new database transaction. See database.Tx for details.
func (c *connection) NewTx(ctx context.Context) (context.Context, database.Tx, error) {
	// Check there's not already a tx in the context
	if txFromContext(ctx) != nil {
		panic("postgres: NewTx called in an existing transaction")
	}

	// Start a new tx
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	// Wrap the tx
	return contextWithTx(ctx, tx), transaction{tx: tx}, nil
}

// transaction implements database.Tx.
type transaction struct {
	tx *sqlx.Tx
}

func (t transaction) Commit() error {
	return t.tx.Commit()
}

func (t transaction) Rollback() error {
	return t.tx.Rollback()
}

// txCtxKey is used for saving a DB transaction in a context.
type txCtxKey struct{}

// contextWithTx returns a wrapped context that contains a DB transaction.
func contextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// txFromContext retrieves a DB transaction wrapped with contextWithTx.
// If no transaction is in the context, it returns nil.
func txFromContext(ctx context.Context) *sqlx.Tx {
	conn := ctx.Value(txCtxKey{})
	if conn != nil {
		return conn.(*sqlx.Tx)
	}
	return nil
}

// dbHandle provides a common interface for sqlx.DB and sqlx.Tx.
type dbHandle interface {
	sqlx.QueryerContext
	sqlx.PreparerContext
	sqlx.ExecerContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// getDB either returns the current tx (if one is present) or c.db.
func (c *connection) getDB(ctx context.Context) dbHandle {
	tx := txFromContext(ctx)
	if tx == nil {
		return c.db
	}
	return tx
}
