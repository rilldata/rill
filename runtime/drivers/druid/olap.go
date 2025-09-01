package druid

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid/druidsqldriver"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const (
	numRetries = 3
	retryWait  = 300 * time.Millisecond
)

var _ drivers.OLAPStore = &connection{}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDruid
}

func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return fmt.Errorf("druid: WithConnection not supported")
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	return res.Close()
}

func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("druid query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
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

	var queryCfg *druidsqldriver.QueryConfig
	if stmt.UseCache != nil {
		queryCfg = &druidsqldriver.QueryConfig{
			UseCache: stmt.UseCache,
		}
	}
	if stmt.PopulateCache != nil {
		if queryCfg == nil {
			queryCfg = &druidsqldriver.QueryConfig{}
		}
		queryCfg.PopulateCache = stmt.PopulateCache
	}
	if !c.config.SkipQueryPriority && stmt.Priority != 0 {
		if queryCfg == nil {
			queryCfg = &druidsqldriver.QueryConfig{}
		}
		queryCfg.Priority = stmt.Priority
	}

	if queryCfg != nil {
		ctx = druidsqldriver.WithQueryConfig(ctx, queryCfg)
	}

	var rows *sqlx.Rows
	var err error

	re := retrier.New(retrier.ExponentialBackoff(numRetries, retryWait), retryErrClassifier{})
	err = re.RunCtx(ctx, func(ctx2 context.Context) error {
		rows, err = c.db.QueryxContext(ctx2, stmt.Query, stmt.Args...)
		return err
	})
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

func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	query = fmt.Sprintf("SELECT * FROM (%s) LIMIT 0", query)

	res, err := c.Query(ctx, &drivers.Statement{
		Query:            query,
		Args:             args,
		ExecutionTimeout: drivers.DefaultQuerySchemaTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return res.Schema, nil
}

func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// retryErrClassifier classifies 429 errors as retryable and all other errors as non retryable
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	if strings.Contains(err.Error(), "QueryCapacityExceededException") {
		return retrier.Retry
	}

	return retrier.Fail
}
