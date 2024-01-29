package queries

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnRugHistogram struct {
	TableName  string
	ColumnName string
	Result     []*runtimev1.NumericOutliers_Outlier
}

var _ runtime.Query = &ColumnRugHistogram{}

func (q *ColumnRugHistogram) Key() string {
	return fmt.Sprintf("ColumnRugHistogram:%s:%s", q.TableName, q.ColumnName)
}

func (q *ColumnRugHistogram) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnRugHistogram) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result) > 0 {
		// approx
		size = sizeProtoMessage(q.Result[0]) * int64(len(q.Result))
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *ColumnRugHistogram) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.NumericOutliers_Outlier)
	if !ok {
		return fmt.Errorf("ColumnRugHistogram: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnRugHistogram) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	rugSQL, err := rugSQL(ctx, olap, q, priority)
	if err != nil {
		return err
	}
	if rugSQL == "" {
		return nil
	}

	outlierResults, err := olap.Execute(ctx, &drivers.Statement{
		Query:            rugSQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer outlierResults.Close()

	outlierBins := make([]*runtimev1.NumericOutliers_Outlier, 0)
	for outlierResults.Next() {
		outlier := &runtimev1.NumericOutliers_Outlier{}
		err = outlierResults.Scan(&outlier.Bucket, &outlier.Low, &outlier.High, &outlier.Present, &outlier.Count)
		if err != nil {
			return err
		}
		outlierBins = append(outlierBins, outlier)
	}

	err = outlierResults.Err()
	if err != nil {
		return err
	}

	q.Result = outlierBins

	return nil
}

func rugSQL(ctx context.Context, olap drivers.OLAPStore, q *ColumnRugHistogram, priority int) (string, error) {
	min, max, rng, err := getMinMaxRange(ctx, olap, q.ColumnName, q.TableName, priority)
	if err != nil {
		return "", err
	}
	if min == nil || max == nil || rng == nil {
		return "", nil
	}

	sanitizedColumnName := safeName(q.ColumnName)
	outlierPseudoBucketSize := 500
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)

	if olap.Dialect() == drivers.DialectDuckDB {
		return fmt.Sprintf(
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
		SELECT count(*) as c from values WHERE value = (%[6]v)
	), histrogram_with_edge AS (
	  SELECT
			bucket,
			low,
			high,
			-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
			CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
		FROM histogram_stage
	)
	SELECT
		bucket,
		low,
		high,
		CASE WHEN count>0 THEN true ELSE false END AS present,
		count
	  FROM histrogram_with_edge
	  WHERE present=true
`,
			selectColumn,
			sanitizedColumnName,
			safeName(q.TableName),
			outlierPseudoBucketSize,
			*min,
			*max,
			*rng,
		), nil
	} else if olap.Dialect() == drivers.DialectClickHouse {
		return fmt.Sprintf(
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
		  number::FLOAT as bucket,
		  (number) * (%[7]v) / %[4]v + (%[5]v) as low,
		  (number + 1) * (%[7]v) / %[4]v + (%[5]v) as high
		FROM numbers(0, %[4]v)
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
		SELECT count(*) as c from values WHERE value = (%[6]v)
	), histrogram_with_edge AS (
	  SELECT
			bucket,
			low,
			high,
			-- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
			CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
		FROM histogram_stage
	)
	SELECT
		bucket,
		low,
		high,
		CASE WHEN count>0 THEN true ELSE false END AS present,
		ifNull(count, 0)
	  FROM histrogram_with_edge
	  WHERE present=true
`,
			selectColumn,
			sanitizedColumnName,
			safeName(q.TableName),
			outlierPseudoBucketSize,
			*min,
			*max,
			*rng,
		), nil
	}
	return "", fmt.Errorf("unknown dialect '%s'", olap.Dialect())
}

func (q *ColumnRugHistogram) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
