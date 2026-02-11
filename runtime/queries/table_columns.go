package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type TableColumns struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	Result         *runtimev1.TableColumnsResponse
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
	if len(q.Result.ProfileColumns) > 0 {
		// approx
		size = sizeProtoMessage(q.Result.ProfileColumns[0]) * int64(len(q.Result.ProfileColumns))
	}
	if len(q.Result.UnsupportedColumns) > 0 {
		r, err := json.Marshal(q.Result.UnsupportedColumns)
		if err == nil { // ignore error
			size += int64(len(r))
		}
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *TableColumns) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TableColumnsResponse)
	if !ok {
		return fmt.Errorf("TableColumns: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableColumns) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if !supportedTableHeadDialects[olap.Dialect()] {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
	if olap.Dialect() == drivers.DialectDuckDB {
		return olap.WithConnection(ctx, priority, func(ctx context.Context, ensuredCtx context.Context) error {
			// views return duplicate column names, so we need to create a temporary table
			temporaryTableName := tempName("profile_columns_")
			err = olap.Exec(ctx, &drivers.Statement{
				Query:            fmt.Sprintf(`CREATE TEMPORARY TABLE %s AS (SELECT * FROM %s LIMIT 1)`, temporaryTableName, olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName)),
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

			rows, err := olap.Query(ctx, &drivers.Statement{
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
				if err := rows.Scan(&pc.Name, &pc.Type); err != nil {
					return err
				}
				// TODO: Find a better way to handle this, this is ugly
				if strings.Contains(pc.Type, "ENUM") {
					pc.Type = "VARCHAR"
				}
				pcs = append(pcs, &pc)
				i++
			}
			err = rows.Err()
			if err != nil {
				return err
			}
			q.Result = &runtimev1.TableColumnsResponse{
				ProfileColumns: pcs[0:i],
			}
			return nil
		})
	}
	tbl, err := olap.InformationSchema().Lookup(ctx, q.Database, q.DatabaseSchema, q.TableName)
	if err != nil {
		return err
	}
	q.Result = &runtimev1.TableColumnsResponse{
		ProfileColumns:     make([]*runtimev1.ProfileColumn, len(tbl.Schema.Fields)),
		UnsupportedColumns: tbl.UnsupportedCols,
	}
	for i := 0; i < len(tbl.Schema.Fields); i++ {
		q.Result.ProfileColumns[i] = &runtimev1.ProfileColumn{
			Name: tbl.Schema.Fields[i].Name,
			Type: tbl.Schema.Fields[i].Type.Code.String(),
		}
	}
	return nil
}

func (q *TableColumns) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
