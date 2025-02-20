package queries

import (
	"context"
	"fmt"
	"io"
	"math"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnNumericHistogram struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	ColumnName     string
	Method         runtimev1.HistogramMethod
	Threshold      int
	Result         []*runtimev1.NumericHistogramBins_Bin
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
		return fmt.Errorf("unknown histogram method %q", q.Method)
	}

	return nil
}

func (q *ColumnNumericHistogram) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}

func (q *ColumnNumericHistogram) calculateBucketSize(ctx context.Context, olap drivers.OLAPStore, priority int) (float64, error) {
	sanitizedColumnName := safeName(q.ColumnName)
	var qryString string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		qryString = "SELECT (approx_quantile(%s, 0.75)-approx_quantile(%s, 0.25))::DOUBLE AS iqr, approx_count_distinct(%s) AS count, (max(%s) - min(%s))::DOUBLE AS range FROM %s"
	case drivers.DialectClickHouse:
		qryString = "SELECT (quantileTDigest(0.75)(%s)-quantileTDigest(0.25)(%s)) AS iqr, uniq(%s) AS count, (max(%s) - min(%s)) AS range FROM %s"
	default:
		return 0, fmt.Errorf("unsupported dialect %v", olap.Dialect())
	}
	querySQL := fmt.Sprintf(qryString,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            querySQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var iqr, rangeVal *float64
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

	if iqr == nil || rangeVal == nil || *rangeVal == 0.0 {
		return 0, nil
	}

	var bucketSize float64
	if count < 40 {
		// Use cardinality if unique count less than 40
		bucketSize = count
	} else {
		// Use Freedman–Diaconis rule for calculating number of bins
		bucketWidth := (2 * *iqr) / math.Cbrt(count)
		FDEstimatorBucketSize := math.Ceil(*rangeVal / bucketWidth)
		bucketSize = math.Min(40, FDEstimatorBucketSize)
	}
	return bucketSize, nil
}

func (q *ColumnNumericHistogram) calculateFDMethod(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect %q", olap.Dialect())
	}

	if olap.Dialect() == drivers.DialectClickHouse {
		// Returning early with empty results because this query tends to hang on ClickHouse.
		return nil
	}

	minVal, maxVal, rng, err := getMinMaxRange(ctx, olap, q.ColumnName, q.Database, q.DatabaseSchema, q.TableName, priority)
	if err != nil {
		return err
	}
	if minVal == nil || maxVal == nil || rng == nil {
		return nil
	}

	sanitizedColumnName := safeName(q.ColumnName)
	bucketSize, err := q.calculateBucketSize(ctx, olap, priority)
	if err != nil {
		return err
	}

	if bucketSize == 0 {
		return nil
	}

	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
	histogramSQL := fmt.Sprintf(
		`
          WITH data_table AS (
            SELECT %[1]s as %[2]s 
            FROM %[3]s
            WHERE `+isNonNullFinite(olap.Dialect(), sanitizedColumnName)+`
          ), values AS (
            SELECT %[2]s as value from data_table
            WHERE `+isNonNullFinite(olap.Dialect(), sanitizedColumnName)+`
          ), buckets AS (
            SELECT
              `+rangeNumbersCol(olap.Dialect())+`::DOUBLE as bucket,
              (bucket) * (%[7]v) / %[4]v + (%[5]v) as low,
              (bucket + 1) * (%[7]v) / %[4]v + (%[5]v) as high
            FROM `+rangeNumbers(olap.Dialect())+`(0, %[4]v)
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
		  	ifNull(bucket, 0) AS bucket,
		  	ifNull(low, 0) AS low,
		  	ifNull(high, 0) AS high,
            -- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
            ifNull(CASE WHEN high = (SELECT max(high) FROM histogram_stage) THEN count + (SELECT c FROM right_edge) ELSE count END, 0) AS count
            FROM histogram_stage
            ORDER BY bucket
	      `,
		selectColumn,
		sanitizedColumnName,
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
		bucketSize,
		*minVal,
		*maxVal,
		*rng,
	)

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
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if olap.Dialect() == drivers.DialectClickHouse {
		// Returning early with empty results because this query tends to hang on ClickHouse.
		return nil
	}

	minVal, maxVal, rng, err := getMinMaxRange(ctx, olap, q.ColumnName, q.Database, q.DatabaseSchema, q.TableName, priority)
	if err != nil {
		return err
	}
	if minVal == nil || maxVal == nil || rng == nil {
		return nil
	}

	ticks := 40.0
	if *rng < ticks {
		ticks = *rng
	}

	startTick, endTick, gap := NiceAndStep(*minVal, *maxVal, ticks)
	bucketCount := int(math.Ceil((endTick - startTick) / gap))
	if gap == 1 {
		bucketCount++
	}

	sanitizedColumnName := safeName(q.ColumnName)
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
	histogramSQL := fmt.Sprintf(
		`
		WITH data_table AS (
			SELECT %[1]s as %[2]s
			FROM %[3]s
			WHERE `+isNonNullFinite(olap.Dialect(), sanitizedColumnName)+`
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
				`+rangeNumbersCol(olap.Dialect())+`::FLOAT as bucket,
				(bucket * %[7]f::FLOAT + %[5]f) as low,
				(bucket * %[7]f::FLOAT + %7f::FLOAT / 2 + %[5]f) as midpoint,
				((bucket + 1) * %[7]f::FLOAT + %[5]f) as high
			FROM `+rangeNumbers(olap.Dialect())+`(0, %[4]d)
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
			LEFT JOIN binned_data ON binned_data.bucket = buckets.bucket
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
		olap.Dialect().EscapeTable(q.Database, q.DatabaseSchema, q.TableName),
		bucketCount,
		startTick,
		endTick,
		gap,
		endTick-startTick,
	)

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

// getMinMaxRange get min, max and range of values for a given column. This is needed since nesting it in query is throwing error in 0.9.x
func getMinMaxRange(ctx context.Context, olap drivers.OLAPStore, columnName, database, databaseSchema, tableName string, priority int) (*float64, *float64, *float64, error) {
	sanitizedColumnName := safeName(columnName)
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)

	minMaxSQL := fmt.Sprintf(
		`
			SELECT
				min(%[2]s) AS min,
				max(%[2]s) AS max,
				max(%[2]s) - min(%[2]s) AS range
			FROM %[1]s
			WHERE `+isNonNullFinite(olap.Dialect(), sanitizedColumnName)+`
		`,
		olap.Dialect().EscapeTable(database, databaseSchema, tableName),
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
	var minVal, maxVal, rng *float64
	if minMaxRow.Next() {
		err = minMaxRow.Scan(&minVal, &maxVal, &rng)
		if err != nil {
			minMaxRow.Close()
			return nil, nil, nil, err
		}
	}

	minMaxRow.Close()

	return minVal, maxVal, rng, nil
}

func isNonNullFinite(d drivers.Dialect, floatCol string) string {
	switch d {
	case drivers.DialectClickHouse:
		return fmt.Sprintf("%s IS NOT NULL AND isFinite(%s)", floatCol, floatCol)
	case drivers.DialectDuckDB:
		return fmt.Sprintf("%s IS NOT NULL AND NOT isinf(%s)", floatCol, floatCol)
	default:
		return "1=1"
	}
}
