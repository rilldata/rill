package queries

import (
	"context"
	"fmt"
	"io"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnNullCount struct {
	Connector  string
	TableName  string
	ColumnName string
	Result     float64
}

var _ runtime.Query = &ColumnNullCount{}

func (q *ColumnNullCount) Key() string {
	return fmt.Sprintf("ColumnNullCount:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnNullCount) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnNullCount) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: int64(reflect.TypeOf(q.Result).Size()),
	}
}

func (q *ColumnNullCount) UnmarshalResult(v any) error {
	res, ok := v.(float64)
	if !ok {
		return fmt.Errorf("ColumnNullCount: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnNullCount) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	nullCountSQL := fmt.Sprintf("SELECT count(*) AS count FROM %s WHERE %s IS NULL",
		safeName(q.TableName),
		safeName(q.ColumnName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            nullCountSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var count float64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	q.Result = count
	return nil
}

func (q *ColumnNullCount) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
