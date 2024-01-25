package queries

import (
	"context"
	"fmt"
	"io"

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

func (q *ColumnDescriptiveStatistics) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnDescriptiveStatistics) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
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
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	sanitizedColumnName := safeName(q.ColumnName)
	var descriptiveStatisticsSQL string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		descriptiveStatisticsSQL = fmt.Sprintf("SELECT "+
			"min(%s)::DOUBLE as min, "+
			"approx_quantile(%s, 0.25)::DOUBLE as q25, "+
			"approx_quantile(%s, 0.5)::DOUBLE as q50, "+
			"approx_quantile(%s, 0.75)::DOUBLE as q75, "+
			"max(%s)::DOUBLE as max, "+
			"avg(%s)::DOUBLE as mean, "+
			"stddev_pop(%s)::DOUBLE as sd "+
			"FROM %s",
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			safeName(q.TableName))
	case drivers.DialectClickHouse:
		descriptiveStatisticsSQL = fmt.Sprintf("SELECT "+
			"min(%s)::DOUBLE as min, "+
			"quantileTDigest(0.25)(%s)::DOUBLE as q25, "+
			"quantileTDigest(0.5)(%s)::DOUBLE as q50, "+
			"quantileTDigest(0.75)(%s)::DOUBLE as q75, "+
			"max(%s)::DOUBLE as max, "+
			"avg(%s)::DOUBLE as mean, "+
			"stddevSamp(%s)::DOUBLE as sd "+
			"FROM %s",
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			safeName(q.TableName))
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            descriptiveStatisticsSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	stats := new(runtimev1.NumericStatistics)
	var min, q25, q50, q75, max, mean, sd *float64
	if rows.Next() {
		err = rows.Scan(&min, &q25, &q50, &q75, &max, &mean, &sd)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if min != nil {
		stats.Min = *min
		stats.Max = *max
		stats.Q25 = *q25
		stats.Q50 = *q50
		stats.Q75 = *q75
		stats.Mean = *mean
		stats.Sd = *sd
		q.Result = stats
	}

	return nil
}

func (q *ColumnDescriptiveStatistics) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
