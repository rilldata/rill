package queries

import (
	"context"
	"database/sql"
	"fmt"
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

func (q *ColumnNumericHistogram) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnNumericHistogram) MarshalResult() any {
	return q.Result
}

func (q *ColumnNumericHistogram) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.NumericHistogramBins_Bin)
	if !ok {
		return fmt.Errorf("ColumnNumericHistogram: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnNumericHistogram) calculateBucketSize(ctx context.Context, olap drivers.OLAPStore, instanceID string, priority int) (float64, error) {
	sanitizedColumnName := safeName(q.ColumnName)
	querySQL := fmt.Sprintf(
		"SELECT approx_quantile(%s, 0.75)-approx_quantile(%s, 0.25) AS iqr, approx_count_distinct(%s) AS count, max(%s) - min(%s) AS range FROM %s",
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		safeName(q.TableName),
	)

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    querySQL,
		Priority: priority,
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
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
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
	histogramSQL := fmt.Sprintf(
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
              range as bucket,
              (range) * (select range FROM S) / %[4]v + (select minVal from S) as low,
              (range + 1) * (select range FROM S) / %[4]v + (select minVal from S) as high
            FROM range(0, %[4]v, 1)
          ),
          -- bin the values
          binned_data AS (
            SELECT 
              FLOOR((value - (select minVal from S)) / (select range from S) * %[4]v) as bucket
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
            SELECT count(*) as c from values WHERE value = (select maxVal from S)
          )
          SELECT 
            bucket,
            low,
            high,
            -- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
            CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
            FROM histogram_stage
	      `,
		selectColumn,
		sanitizedColumnName,
		safeName(q.TableName),
		bucketSize,
	)

	histogramRows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    histogramSQL,
		Priority: priority,
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
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	sanitizedColumnName := safeName(q.ColumnName)
	minMaxSQL := fmt.Sprintf(
		`
			SELECT
				min(%[2]s) as min,
				max(%[2]s) as max,
				max(%[2]s) - min(%[2]s) as range
			FROM %[1]s 
			WHERE %[2]s IS NOT NULL
		`,
		safeName(q.TableName),
		sanitizedColumnName,
	)

	minMaxRow, err := olap.Execute(ctx, &drivers.Statement{
		Query:    minMaxSQL,
		Priority: priority,
	})
	if err != nil {
		return err
	}

	var min, max, rng float64
	if minMaxRow.Next() {
		err = minMaxRow.Scan(&min, &max, &rng)
		if err != nil {
			minMaxRow.Close()
			return err
		}
	}

	minMaxRow.Close()

	ticks := 40.0
	if rng < ticks {
		ticks = rng
	}
	niceResult := NiceAndStep(min, max, ticks)
	startTick := niceResult[0]
	endTick := niceResult[1]
	gap := niceResult[2]
	if gap < 0.0 {
		gap = 1 / -gap
	}
	bucketCount := int(math.Ceil((endTick - startTick) / gap))

	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)
	histogramSQL := fmt.Sprintf(
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
				range as bucket,
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
			CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
		FROM histogram_stage
	  `,
		selectColumn,
		sanitizedColumnName,
		safeName(q.TableName),
		bucketCount,
		startTick,
		endTick,
		gap,
		endTick-startTick,
	)

	histogramRows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    histogramSQL,
		Priority: priority,
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
		panic(fmt.Sprintf("Unknown histogram method %v", q.Method))
	}

	return nil
}
