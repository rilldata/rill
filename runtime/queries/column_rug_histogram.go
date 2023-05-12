package queries

import (
	"context"
	"fmt"

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

func (q *ColumnRugHistogram) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnRugHistogram) MarshalResult() *runtime.CacheObject {
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

func (q *ColumnRugHistogram) UnmarshalResult(v any) error {
	res, ok := v.([]*runtimev1.NumericOutliers_Outlier)
	if !ok {
		return fmt.Errorf("ColumnRugHistogram: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnRugHistogram) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	sanitizedColumnName := safeName(q.ColumnName)
	outlierPseudoBucketSize := 500
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)

	rugSQL := fmt.Sprintf(`WITH data_table AS (
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
	  ), 
	  buckets AS (
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
	  WHERE present=true`, selectColumn, sanitizedColumnName, safeName(q.TableName), outlierPseudoBucketSize)

	outlierResults, err := olap.Execute(ctx, &drivers.Statement{
		Query:    rugSQL,
		Priority: priority,
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
