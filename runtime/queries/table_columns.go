package queries

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

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

func (q *TableColumns) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
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
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() == drivers.DialectDuckDB {
		return olap.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, _ *sql.Conn) error {
			// views return duplicate column names, so we need to create a temporary table
			temporaryTableName := tempName("profile_columns_")
			err = olap.Exec(ctx, &drivers.Statement{
				Query:            fmt.Sprintf(`CREATE TEMPORARY TABLE "%s" AS (SELECT * FROM %s LIMIT 1)`, temporaryTableName, safeName(q.TableName)),
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
				// TODO: Find a better way to handle this, this is ugly
				if strings.Contains(pc.Type, "ENUM") {
					pc.Type = "VARCHAR"
				}
				pcs = append(pcs, &pc)
				i++
			}

			q.Result = pcs[0:i]
			return nil
		})
	}

	if olap.Dialect() == drivers.DialectDruid {
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query: fmt.Sprintf(`
			SELECT COLUMN_NAME AS name, DATA_TYPE AS type
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_NAME = '%s'`, q.TableName),
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

			if strings.Contains(pc.Type, "COMPLEX") {
				pc.Type = "VARCHAR"
			}
			pcs = append(pcs, &pc)
			i++
		}

		q.Result = pcs[0:i]
		return nil
	}

	return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
}

func (q *TableColumns) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
