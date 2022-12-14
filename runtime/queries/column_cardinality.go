package queries

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ColumnCardinality struct {
	TableName  string
	ColumnName string
	Result     float64
}

var _ runtime.Query = &ColumnCardinality{}

func (q *ColumnCardinality) Key() string {
	return fmt.Sprintf("ColumnCardinality:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnCardinality) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnCardinality) MarshalResult() any {
	return q.Result
}

func (q *ColumnCardinality) UnmarshalResult(v any) error {
	res, ok := v.(float64)
	if !ok {
		return fmt.Errorf("ColumnCardinality: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnCardinality) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	sanitizedColumnName := quoteName(q.ColumnName)
	requestSQL := fmt.Sprintf("SELECT approx_count_distinct(%s) as count from %s", sanitizedColumnName, quoteName(q.TableName))

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    requestSQL,
		Priority: priority,
	})

	if err != nil {
		return err
	}
	defer rows.Close()
	var count float64
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return err
		}
		q.Result = count
		return nil
	}
	return status.Error(codes.Internal, "no rows returned")
}
