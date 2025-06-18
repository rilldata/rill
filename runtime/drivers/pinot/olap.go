package pinot

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = &Connection{}

func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectPinot
}

func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

func (c *Connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return fmt.Errorf("pinot: WithConnection not supported")
}

func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	return res.Close()
}

func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.logQueries {
		c.logger.Info("pinot query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args), zap.Int64("timeoutMS", c.timeoutMS), observability.ZapCtx(ctx))
	}
	if stmt.DryRun {
		rows, err := c.db.QueryxContext(ctx, "EXPLAIN PLAN FOR "+stmt.Query, stmt.Args...)
		if err != nil {
			return nil, err
		}

		return nil, rows.Close()
	}

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	// add timeout if configured to the sql to propagate it to the Pinot server to override the cluster timeout
	if c.timeoutMS > 0 {
		stmt.Query = fmt.Sprintf("SET timeoutMS=%d; %s", c.timeoutMS, stmt.Query)
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		rows.Close()
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	r := &drivers.Result{Rows: rows, Schema: schema}
	r.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil
	})

	return r, nil
}
