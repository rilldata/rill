package queries

import (
	"context"
	"database/sql"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnDescriptiveStatistics struct {
	TableName  string
	ColumnName string
	Result     *runtimev1.NumericStatistics
}

var _ runtime.Query = &ColumnDescriptiveStatistics{}

func (q *ColumnDescriptiveStatistics) Key() string {
	return fmt.Sprintf("ColumnDescriptiveStatistics:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnDescriptiveStatistics) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnDescriptiveStatistics) MarshalResult() any {
	return q.Result
}

func (q *ColumnDescriptiveStatistics) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.NumericStatistics)
	if !ok {
		return fmt.Errorf("ColumnDescriptiveStatistics: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnDescriptiveStatistics) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	sanitizedColumnName := quoteName(q.ColumnName)
	descriptiveStatisticsSql := fmt.Sprintf("SELECT "+
		"min(%s) as min, "+
		"approx_quantile(%s, 0.25) as q25, "+
		"approx_quantile(%s, 0.5)  as q50, "+
		"approx_quantile(%s, 0.75) as q75, "+
		"max(%s) as max, "+
		"avg(%s)::FLOAT as mean, "+
		"stddev_pop(%s) as sd "+
		"FROM %s",
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		quoteName(q.TableName))

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    descriptiveStatisticsSql,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	stats := new(runtimev1.NumericStatistics)
	var min, q25, q50, q75, max, mean, sd sql.NullFloat64
	if rows.Next() {
		err = rows.Scan(&min, &q25, &q50, &q75, &max, &mean, &sd)
		if err != nil {
			return err
		}
	}
	if min.Valid {
		stats.Min = min.Float64
		stats.Max = max.Float64
		stats.Q25 = q25.Float64
		stats.Q50 = q50.Float64
		stats.Q75 = q75.Float64
		stats.Mean = mean.Float64
		stats.Sd = sd.Float64
		q.Result = stats
	}

	return nil
}
