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

type TableCardinality struct {
	Connector string
	TableName string
	Result    int64
}

var _ runtime.Query = &TableCardinality{}

func (q *TableCardinality) Key() string {
	return fmt.Sprintf("TableCardinality:%s", q.TableName)
}

func (q *TableCardinality) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *TableCardinality) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: int64(reflect.TypeOf(q.Result).Size()),
	}
}

func (q *TableCardinality) UnmarshalResult(v any) error {
	res, ok := v.(int64)
	if !ok {
		return fmt.Errorf("TableCardinality: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableCardinality) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	countSQL := fmt.Sprintf("SELECT count(*) AS count FROM %s",
		safeName(q.TableName),
	)

	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            countSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int64
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

func (q *TableCardinality) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
