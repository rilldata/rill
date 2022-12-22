package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (q *TableColumns) MarshalResult() any {
	return q.Result
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

	temporaryTableName := "profile_columns_" + ReplaceHyphen(uuid.New().String())
	// views return duplicate column names, so we need to create a temporary table
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf(`CREATE TEMPORARY TABLE "%s" AS (SELECT * FROM "%s" LIMIT 1)`, temporaryTableName, q.TableName),
		Priority: priority,
	})
	if err != nil {
		return err
	}
	rows.Close()
	defer DropTempTable(olap, priority, temporaryTableName)

	rows, err = olap.Execute(ctx, &drivers.Statement{
		Query: fmt.Sprintf(`select column_name as name, data_type as type from information_schema.columns 
		where table_name = '%s' and table_schema = 'temp'`, temporaryTableName),
		Priority: priority,
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

	// Disabling this for now. we need to move this to a separate API
	// It adds a lot of response time to getting columns
	// for _, pc := range pcs[0:i] {
	//	columnName := EscapeDoubleQuotes(pc.Name)
	//	rows, err = s.query(ctx, req.InstanceId, &drivers.Statement{
	//		Query:    fmt.Sprintf(`select max(length("%s")) as max from %s`, columnName, req.TableName),
	//		Priority: int(req.Priority),
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//	for rows.Next() {
	//		var max sql.NullInt32
	//		if err := rows.Scan(&max); err != nil {
	//			return nil, err
	//		}
	//		if max.Valid {
	//			pc.LargestStringLength = int32(max.Int32)
	//		}
	//	}
	//	rows.Close()
	//}

	q.Result = pcs[0:i]
	return nil
}
