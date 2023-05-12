package queries

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

type TableHead struct {
	TableName string
	Limit     int
	Result    []*structpb.Struct
}

var _ runtime.Query = &TableHead{}

func (q *TableHead) Key() string {
	return fmt.Sprintf("TableHead:%s:%d", q.TableName, q.Limit)
}

func (q *TableHead) Deps() []string {
	return []string{q.TableName}
}

func (q *TableHead) MarshalResult() *runtime.CacheObject {
	var size int64
	if len(q.Result) > 0 {
		// approx
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}

	return &runtime.CacheObject{
		Result:      q.Result,
		SizeInBytes: size,
	}
}

func (q *TableHead) UnmarshalResult(v any) error {
	res, ok := v.([]*structpb.Struct)
	if !ok {
		return fmt.Errorf("TableHead: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableHead) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT * FROM %s LIMIT %d", safeName(q.TableName), q.Limit),
		Priority: priority,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return err
	}

	q.Result = data
	return nil
}
