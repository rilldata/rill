package queries

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
)

type ColumnTopK struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	ColumnName     string
	Agg            string
	K              int
	Result         *runtimev1.TopK
}

var _ runtime.Query = &ColumnTopK{}

func (q *ColumnTopK) Key() string {
	return fmt.Sprintf("ColumnTopK:%s:%s:%s:%d", q.TableName, q.ColumnName, q.Agg, q.K)
}

func (q *ColumnTopK) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnTopK) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *ColumnTopK) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TopK)
	if !ok {
		return fmt.Errorf("topk: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTopK) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Get OLAP connection
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	// Check dialect
	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	// Build SQL
	qry := fmt.Sprintf("SELECT %s AS value, %s AS count FROM %s GROUP BY %s ORDER BY count DESC, value ASC LIMIT %d",
		safeName(q.ColumnName),
		q.Agg,
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
		safeName(q.ColumnName),
		q.K,
	)

	// Run query
	rows, err := olap.Query(ctx, &drivers.Statement{
		Query:            qry,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan result
	res := &runtimev1.TopK{}
	for rows.Next() {
		entry := &runtimev1.TopK_Entry{}
		var val interface{}
		err = rows.Scan(&val, &entry.Count)
		if err != nil {
			return err
		}

		entry.Value, err = pbutil.ToValue(val, safeFieldType(rows.Schema, 0))
		if err != nil {
			return err
		}

		res.Entries = append(res.Entries, entry)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	// Save result
	q.Result = res
	return nil
}

func (q *ColumnTopK) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
