package queries

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type TableColumns struct {
	TableName string
	Result    []*runtimev1.ProfileColumn
}

var _ runtime.Query = &TableColumns{}

func (q *TableColumns) Key() string {
	return fmt.Sprintf("TableColumns:%s", q.TableName)
}

func (q *TableColumns) Deps() []string {
	return []string{q.TableName}
}

func (q *TableColumns) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result) > 0 {
		// approx
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *TableColumns) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.ProfileColumn)
	if !ok {
		return fmt.Errorf("TableColumns: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableColumns) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return olap.WithConnection(ctx, priority, func(ctx context.Context, ensuredCtx context.Context) error {
		// views return duplicate column names, so we need to create a temporary table
		temporaryTableName := tempName("profile_columns_")
		err = olap.Exec(ctx, &drivers.Statement{
			Query:            fmt.Sprintf(`CREATE TEMPORARY TABLE "%s" AS (SELECT * FROM "%s" LIMIT 1)`, temporaryTableName, q.TableName),
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer func() {
			// NOTE: Using ensuredCtx
			_ = olap.Exec(ensuredCtx, &drivers.Statement{
				Query:            `DROP TABLE "` + temporaryTableName + `"`,
				Priority:         priority,
				ExecutionTimeout: defaultExecutionTimeout,
			})
		}()

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query: fmt.Sprintf(`
				SELECT column_name AS name, data_type AS type
				FROM information_schema.columns
				WHERE table_catalog = 'temp' AND table_name = '%s'`, temporaryTableName),
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer rows.Close()

		var pcs []*runtimev1.ProfileColumn
		i := 0
		for rows.Next() {
			pc := runtimev1.ProfileColumn{}
			if err := rows.StructScan(&pc); err != nil {
				return err
			}
			pcs = append(pcs, &pc)
			i++
		}

		q.Result = pcs[0:i]
		return nil
	})
}
