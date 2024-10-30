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
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	ColumnName     string
	Result         *runtimev1.NumericStatistics
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
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	sanitizedColumnName := safeName(q.ColumnName)
	var descriptiveStatisticsSQL string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		descriptiveStatisticsSQL = fmt.Sprintf("SELECT "+
			"min(%[1]s)::DOUBLE as min, "+
			"approx_quantile(%[1]s, 0.25)::DOUBLE as q25, "+
			"approx_quantile(%[1]s, 0.5)::DOUBLE as q50, "+
			"approx_quantile(%[1]s, 0.75)::DOUBLE as q75, "+
			"max(%[1]s)::DOUBLE as max, "+
			"avg(%[1]s)::DOUBLE as mean, "+
			"'NaN'::DOUBLE as sd "+
			"FROM %[2]s WHERE NOT isinf(%[1]s) ",
			sanitizedColumnName,
			olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName))
	case drivers.DialectClickHouse:
		descriptiveStatisticsSQL = fmt.Sprintf(`SELECT 
			min(%[1]s)::DOUBLE as min, 
			quantileTDigest(0.25)(%[1]s)::DOUBLE as q25, 
			quantileTDigest(0.5)(%[1]s)::DOUBLE as q50, 
			quantileTDigest(0.75)(%[1]s)::DOUBLE as q75, 
			max(%[1]s)::DOUBLE as max, 
			avg(%[1]s)::DOUBLE as mean, 
			stddevSamp(%[1]s)::DOUBLE as sd 
			FROM %[2]s WHERE `+isNonNullFinite(olap.Dialect(), sanitizedColumnName)+``,
			sanitizedColumnName,
			olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName))
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
	// clickhouse driver can't scan into sql.Nullx when value is not a null
	var minVal, q25, q50, q75, maxVal, mean, sd *float64
	if rows.Next() {
		err = rows.Scan(&minVal, &q25, &q50, &q75, &maxVal, &mean, &sd)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	if minVal != nil {
		stats.Min = *minVal
		stats.Max = *maxVal
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
