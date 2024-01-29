package queries

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnNumericHistogram struct {
	TableName  string
	ColumnName string
	Method     runtimev1.HistogramMethod
	Threshold  int
	Result     []*runtimev1.NumericHistogramBins_Bin
}

var _ runtime.Query = &ColumnNumericHistogram{}

func (q *ColumnNumericHistogram) Key() string {
	return fmt.Sprintf("ColumnNumericHistogram:%s:%s:%s:%d", q.TableName, q.ColumnName, q.Method.String(), q.Threshold)
}

func (q *ColumnNumericHistogram) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnNumericHistogram) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result) > 0 {
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *ColumnNumericHistogram) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.NumericHistogramBins_Bin)
	if !ok {
		return fmt.Errorf("ColumnNumericHistogram: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnNumericHistogram) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	if q.Method == runtimev1.HistogramMethod_HISTOGRAM_METHOD_FD {
		err := q.calculateFDMethod(ctx, rt, instanceID, priority)
		if err != nil {
			return err
		}
	} else if q.Method == runtimev1.HistogramMethod_HISTOGRAM_METHOD_DIAGNOSTIC {
		err := q.calculateDiagnosticMethod(ctx, rt, instanceID, priority)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown histogram method %v", q.Method)
	}

	return nil
}

func (q *ColumnNumericHistogram) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}

func (q *ColumnNumericHistogram) calculateBucketSize(ctx context.Context, olap drivers.OLAPStore, instanceID string, priority int) (float64, error) {
	sanitizedColumnName := safeName(q.ColumnName)
	var querySQL string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		querySQL = fmt.Sprintf(
			"SELECT (approx_quantile(%s, 0.75)-approx_quantile(%s, 0.25))::DOUBLE AS iqr, approx_count_distinct(%s) AS count, (max(%s) - min(%s))::DOUBLE AS range FROM %s",
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			safeName(q.TableName),
		)
	case drivers.DialectClickHouse:
		// assumes that column exists otherwise cast to double fails in clickhouse
		querySQL = fmt.Sprintf(
			"SELECT (quantileTDigest(0.75)(%s)-quantileTDigest(0.25)(%s))::DOUBLE AS iqr, uniq(%s) AS count, (max(%s) - min(%s))::DOUBLE AS range FROM %s",
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			sanitizedColumnName,
			safeName(q.TableName),
		)
	default:
		return 0, fmt.Errorf("unsupported dialect %v", olap.Dialect())
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            querySQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var iqr, rangeVal sql.NullFloat64
	var count float64
	if rows.Next() {
		err = rows.Scan(&iqr, &count, &rangeVal)
		if err != nil {
			return 0, err
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, err
	}

	if !iqr.Valid || !rangeVal.Valid || rangeVal.Float64 == 0.0 {
		return 0, nil
	}

	var bucketSize float64
	if count < 40 {
		// Use cardinality if unique count less than 40
		bucketSize = count
	} else {
		// Use Freedman–Diaconis rule for calculating number of bins
		bucketWidth := (2 * iqr.Float64) / math.Cbrt(count)
		FDEstimatorBucketSize := math.Ceil(rangeVal.Float64 / bucketWidth)
		bucketSize = math.Min(40, FDEstimatorBucketSize)
	}
	return bucketSize, nil
}

func (q *ColumnNumericHistogram) calculateFDMethod(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	min, max, rng, err := getMinMaxRange(ctx, olap, q.ColumnName, q.TableName, priority)
	if err != nil {
		return err
	}
	if min == nil || max == nil || rng == nil {
		return nil
	}

	sanitizedColumnName := safeName(q.ColumnName)
	bucketSize, err := q.calculateBucketSize(ctx, olap, instanceID, priority)
	if err != nil {
		return err
	}

	if bucketSize == 0 {
		return nil
	}

	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
	var histogramSQL string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		histogramSQL = fmt.Sprintf(
			`
			  WITH data_table AS (
				SELECT %[1]s as %[2]s 
				FROM %[3]s
				WHERE %[2]s IS NOT NULL
			  ), values AS (
				SELECT %[2]s as value from data_table
				WHERE %[2]s IS NOT NULL
			  ), buckets AS (
				SELECT
				  range as bucket,
				  (range) * (%[7]v) / %[4]v + (%[5]v) as low,
				  (range + 1) * (%[7]v) / %[4]v + (%[5]v) as high
				FROM range(0, %[4]v, 1)
			  ),
			  -- bin the values
			  binned_data AS (
				SELECT 
				  FLOOR((value - (%[5]v)) / (%[7]v) * %[4]v) as bucket
				from values
			  ),
			  -- join the bucket set with the binned values to generate the histogram
			  histogram_stage AS (
			  SELECT
				  buckets.bucket,
				  low,
				  high,
				  SUM(CASE WHEN binned_data.bucket = buckets.bucket THEN 1 ELSE 0 END) as count
				FROM buckets
				LEFT JOIN binned_data ON binned_data.bucket = buckets.bucket
				GROUP BY buckets.bucket, low, high
				ORDER BY buckets.bucket
			  ),
			  -- calculate the right edge, sine in histogram_stage we don't look at the values that
			  -- might be the largest.
			  right_edge AS (
				SELECT count(*) as c from values WHERE value = %[6]v
			  )
			  SELECT 
				bucket,
				low,
				high,
				-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
				CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
				FROM histogram_stage
				ORDER BY bucket
			  `,
			selectColumn,
			sanitizedColumnName,
			safeName(q.TableName),
			bucketSize,
			*min,
			*max,
			*rng,
		)
	case drivers.DialectClickHouse:
		histogramSQL = fmt.Sprintf(
			`
			  WITH data_table AS (
				SELECT %[1]s as %[2]s 
				FROM %[3]s
				WHERE %[2]s IS NOT NULL
			  ), values AS (
				SELECT %[2]s as value from data_table
				WHERE %[2]s IS NOT NULL
			  ), buckets AS (
				SELECT
				  number::DOUBLE as bucket,
				  ((number) * (%[7]v) / %[4]v + (%[5]v))::DOUBLE as low,
				  ((number + 1) * (%[7]v) / %[4]v + (%[5]v))::DOUBLE as high
				FROM numbers(%[4]v)
			  ),
			  -- bin the values
			  binned_data AS (
				SELECT 
				  FLOOR((value - (%[5]v)) / (%[7]v) * %[4]v) as bucket
				from values
			  ),
			  -- join the bucket set with the binned values to generate the histogram
			  histogram_stage AS (
			  SELECT
				  buckets.bucket,
				  low,
				  high,
				  SUM(CASE WHEN binned_data.bucket = buckets.bucket THEN 1 ELSE 0 END) as count
				FROM buckets
				LEFT JOIN binned_data ON binned_data.bucket = buckets.bucket
				GROUP BY buckets.bucket, low, high
				ORDER BY buckets.bucket
			  ),
			  -- calculate the right edge, sine in histogram_stage we don't look at the values that
			  -- might be the largest.
			  right_edge AS (
				SELECT count(*) as c from values WHERE value = %[6]v
			  )
			  SELECT 
			  	ifNull(bucket, 0)::Float64 as bucket,
				ifNull(low, 0)::Float64 as low,
				ifNull(high, 0)::Float64 as high,
				-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
				ifNull(CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END, 0)::Float64 AS count
				FROM histogram_stage
				ORDER BY bucket
			  `,
			selectColumn,
			sanitizedColumnName,
			safeName(q.TableName),
			bucketSize,
			*min,
			*max,
			*rng,
		)
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}
	histogramRows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            histogramSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}

	defer histogramRows.Close()

	histogramBins := make([]*runtimev1.NumericHistogramBins_Bin, 0)
	for histogramRows.Next() {
		bin := &runtimev1.NumericHistogramBins_Bin{}
		err = histogramRows.Scan(&bin.Bucket, &bin.Low, &bin.High, &bin.Count)
		if err != nil {
			return err
		}
		histogramBins = append(histogramBins, bin)
	}

	err = histogramRows.Err()
	if err != nil {
		return err
	}

	q.Result = histogramBins

	return nil
}

func (q *ColumnNumericHistogram) calculateDiagnosticMethod(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	histogramSQL, err := histogramDiagnosticMethodSQL(ctx, olap, q.ColumnName, q.TableName, priority)
	if err != nil {
		return err
	}
	if histogramSQL == "" {
		return nil
	}

	histogramRows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            histogramSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}

	defer histogramRows.Close()

	histogramBins := make([]*runtimev1.NumericHistogramBins_Bin, 0)
	for histogramRows.Next() {
		bin := &runtimev1.NumericHistogramBins_Bin{}
		err = histogramRows.Scan(&bin.Bucket, &bin.Low, &bin.High, &bin.Midpoint, &bin.Count)
		if err != nil {
			return err
		}
		histogramBins = append(histogramBins, bin)
	}

	err = histogramRows.Err()
	if err != nil {
		return err
	}

	q.Result = histogramBins

	return nil
}

func histogramDiagnosticMethodSQL(ctx context.Context, olap drivers.OLAPStore, colName, tblName string, priority int) (string, error) {
	min, max, rng, err := getMinMaxRange(ctx, olap, colName, tblName, priority)
	if err != nil {
		return "", err
	}
	if min == nil || max == nil || rng == nil {
		return "", nil
	}

	ticks := 40.0
	if *rng < ticks {
		ticks = *rng
	}

	startTick, endTick, gap := NiceAndStep(*min, *max, ticks)
	bucketCount := int(math.Ceil((endTick - startTick) / gap))
	if gap == 1 {
		bucketCount++
	}

	sanitizedColumnName := safeName(colName)
	if olap.Dialect() == drivers.DialectDuckDB {
		selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
		return fmt.Sprintf(
			`
			WITH data_table AS (
				SELECT %[1]s as %[2]s
				FROM %[3]s
				WHERE %[2]s IS NOT NULL
			), S AS (
				SELECT
					min(%[2]s) as minVal,
					max(%[2]s) as maxVal,
					(max(%[2]s) - min(%[2]s)) as range
					FROM data_table
			), values AS (
				SELECT %[2]s as value from data_table
				WHERE %[2]s IS NOT NULL
			), buckets AS (
				SELECT
					range::FLOAT as bucket,
					(range * %[7]f::FLOAT + %[5]f) as low,
					(range * %[7]f::FLOAT + %7f::FLOAT / 2 + %[5]f) as midpoint,
					((range + 1) * %[7]f::FLOAT + %[5]f) as high
				FROM range(0, %[4]d, 1)
			),
			-- bin the values
			binned_data AS (
				SELECT
					FLOOR(%[4]d::FLOAT * ((value::FLOAT - %[5]f) / %[8]f)) as bucket
				from values
			),
			-- join the bucket set with the binned values to generate the histogram
			histogram_stage AS (
				SELECT
					buckets.bucket,
					low,
					high,
					midpoint,
					SUM(CASE WHEN binned_data.bucket = buckets.bucket THEN 1 ELSE 0 END) as count
				FROM buckets
				LEFT JOIN binned_data ON buckets.bucket = binned_data.bucket
				GROUP BY buckets.bucket, low, high, midpoint
				ORDER BY buckets.bucket
			),
			-- calculate the right edge, sine in histogram_stage we don't look at the values that
			-- might be the largest.
			right_edge AS (
				SELECT count(*) as c from values WHERE value = %[6]f
			)
			SELECT
				bucket,
				low,
				high,
				midpoint,
			-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
				CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
			FROM histogram_stage
		ORDER BY bucket
			`,
			selectColumn,
			sanitizedColumnName,
			safeName(tblName),
			bucketCount,
			startTick,
			endTick,
			gap,
			endTick-startTick,
		), nil
	}
	if olap.Dialect() == drivers.DialectClickHouse {
		selectColumn := fmt.Sprintf("%s::Nullable(DOUBLE)", sanitizedColumnName)
		return fmt.Sprintf(
			`
			WITH data_table AS (
				SELECT %[1]s as %[2]s
				FROM %[3]s
				WHERE %[2]s IS NOT NULL
			), S AS (
				SELECT
					min(%[2]s) as minVal,
					max(%[2]s) as maxVal,
					(max(%[2]s) - min(%[2]s)) as range
					FROM data_table
			), values AS (
				SELECT %[2]s as value from data_table
				WHERE %[2]s IS NOT NULL
			), buckets AS (
				SELECT
					number::FLOAT as bucket,
					(number * %[7]f::FLOAT + %[5]f) as low,
					(number * %[7]f::FLOAT + %7f::FLOAT / 2 + %[5]f) as midpoint,
					((number + 1) * %[7]f::FLOAT + %[5]f) as high
				FROM numbers(0, %[4]d)
			),
			-- bin the values
			binned_data AS (
				SELECT
					FLOOR(%[4]d::FLOAT * ((value::FLOAT - %[5]f) / %[8]f)) as bucket
				from values
			),
			-- join the bucket set with the binned values to generate the histogram
			histogram_stage AS (
				SELECT
					buckets.bucket,
					low,
					high,
					midpoint,
					SUM(CASE WHEN binned_data.bucket = buckets.bucket THEN 1 ELSE 0 END) as count
				FROM buckets
				LEFT JOIN binned_data ON buckets.bucket = binned_data.bucket
				GROUP BY buckets.bucket, low, high, midpoint
				ORDER BY buckets.bucket
			),
			-- calculate the right edge, sine in histogram_stage we don't look at the values that
			-- might be the largest.
			right_edge AS (
				SELECT count(*) as c from values WHERE value = %[6]f
			)
			SELECT
				bucket,
				low,
				high,
				midpoint,
			-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
				ifNull(CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END, 0) AS count
			FROM histogram_stage
		ORDER BY bucket
			`,
			selectColumn,
			sanitizedColumnName,
			safeName(tblName),
			bucketCount,
			startTick,
			endTick,
			gap,
			endTick-startTick,
		), nil
	}
	return "", fmt.Errorf("unsupported dialect %s", olap.Dialect())
}

// getMinMaxRange get min, max and range of values for a given column. This is needed since nesting it in query is throwing error in 0.9.x
func getMinMaxRange(ctx context.Context, olap drivers.OLAPStore, columnName, tableName string, priority int) (*float64, *float64, *float64, error) {
	sanitizedColumnName := safeName(columnName)
	var selectColumn string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		selectColumn = fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
	case drivers.DialectClickHouse:
		selectColumn = fmt.Sprintf("%s::Nullable(DOUBLE)", sanitizedColumnName)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported dialect %s", olap.Dialect())
	}

	minMaxSQL := fmt.Sprintf(
		`
			SELECT
				min(%[2]s) AS min,
				max(%[2]s) AS max,
				max(%[2]s) - min(%[2]s) AS range
			FROM %[1]s
			WHERE %[2]s IS NOT NULL
		`,
		safeName(tableName),
		selectColumn,
	)

	minMaxRow, err := olap.Execute(ctx, &drivers.Statement{
		Query:            minMaxSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// clickhouse does not support scanning non null values into sql.Nullx
	// issue : https://github.com/ClickHouse/clickhouse-go/issues/754
	var min, max, rng *float64
	if minMaxRow.Next() {
		err = minMaxRow.Scan(&min, &max, &rng)
		if err != nil {
			minMaxRow.Close()
			return nil, nil, nil, err
		}
	}

	minMaxRow.Close()

	return min, max, rng, nil
}
