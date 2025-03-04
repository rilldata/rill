package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/admin/database"
)

// NewTx starts a new database transaction. See database.Tx for details.
func (c *connection) NewTx(ctx context.Context, allowNested bool) (context.Context, database.Tx, error) {
	// Check if there's already a tx in the context
	if txFromContext(ctx) != nil {
		if !allowNested {
			panic("postgres: NewTx called in an existing transaction")
		}
		// We return a no-op tx because the actual tx must be committed/rolled back by the outermost acquirer.
		return ctx, noopTx{}, nil
	}

	// Start a new tx
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	// Wrap the tx
	return contextWithTx(ctx, tx), tx, nil
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

// noopTx implements a database.Tx that does nothing on Commit/Rollback.
// It is used in nested transactions to avoid committing too early.
type noopTx struct{}

func (noopTx) Commit() error {
	return nil
}

func (noopTx) Rollback() error {
	return nil
}
