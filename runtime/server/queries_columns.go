package server

import (
	"context"
	"fmt"
	"time"

	"github.com/marcboeker/go-duckdb"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetTopK(ctx context.Context, request *runtimev1.GetTopKRequest) (*runtimev1.GetTopKResponse, error) {
	agg := "count(*)"
	if request.Agg != "" {
		agg = request.Agg
	}

	k := 50
	if request.K != 0 {
		k = int(request.K)
	}

	q := &queries.ColumnTopK{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
		Agg:        agg,
		K:          k,
	}

	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.GetTopKResponse{
		CategoricalSummary: &runtimev1.CategoricalSummary{
			Case: &runtimev1.CategoricalSummary_TopK{
				TopK: q.Result,
			},
		},
	}, nil
}

func (s *Server) GetNullCount(ctx context.Context, request *runtimev1.GetNullCountRequest) (*runtimev1.GetNullCountResponse, error) {
	q := &queries.ColumnNullCount{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}

	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.GetNullCountResponse{
		Count: q.Result,
	}, nil

}

func (s *Server) GetDescriptiveStatistics(ctx context.Context, request *runtimev1.GetDescriptiveStatisticsRequest) (*runtimev1.GetDescriptiveStatisticsResponse, error) {
	q := &queries.ColumnDescriptiveStatistics{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	resp := &runtimev1.NumericSummary{
		Case: &runtimev1.NumericSummary_NumericStatistics{
			NumericStatistics: q.Result,
		},
	}
	return &runtimev1.GetDescriptiveStatisticsResponse{
		NumericSummary: resp,
	}, nil
}

/**
 * Estimates the smallest time grain present in the column.
 * The "smallest time grain" is the smallest value that we believe the user
 * can reliably roll up. In other words, if the data is reported daily, this
 * action will return "day", since that's the smallest rollup grain we can
 * rely on.
 *
 * This function can only focus on some common time grains. It will operate on
 * - ms
 * - second
 * - minute
 * - hour
 * - day
 * - week
 * - month
 * - year
 *
 * It will not estimate any more nuanced or difficult-to-measure time grains, such as
 * quarters, once-a-month, etc.
 *
 * It accomplishes its goal by sampling 500k values of a column and then estimating the cardinality
 * of each. If there are < 500k samples, the action will use all of the column's data.
 * We're not sure all the ways this heuristic will fail, but it seems pretty resilient to the tests
 * we've thrown at it.
 */

func (s *Server) EstimateSmallestTimeGrain(ctx context.Context, request *runtimev1.EstimateSmallestTimeGrainRequest) (*runtimev1.EstimateSmallestTimeGrainResponse, error) {
	q := &queries.ColumnTimeGrain{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.EstimateSmallestTimeGrainResponse{
		TimeGrain: q.Result,
	}, nil
}

func (s *Server) GetNumericHistogram(ctx context.Context, request *runtimev1.GetNumericHistogramRequest) (*runtimev1.GetNumericHistogramResponse, error) {
	q := &queries.ColumnNumericHistogram{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.GetNumericHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericHistogramBins{
				NumericHistogramBins: &runtimev1.NumericHistogramBins{
					Bins: q.Result,
				},
			},
		},
	}, nil
}

func (s *Server) GetRugHistogram(ctx context.Context, request *runtimev1.GetRugHistogramRequest) (*runtimev1.GetRugHistogramResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	outlierPseudoBucketSize := 500
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)

	rugSql := fmt.Sprintf(`WITH data_table AS (
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
          WHERE present=true`, selectColumn, sanitizedColumnName, request.TableName, outlierPseudoBucketSize)

	outlierResults, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query:    rugSql,
		Priority: int(request.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer outlierResults.Close()

	outlierBins := make([]*runtimev1.NumericOutliers_Outlier, 0)
	for outlierResults.Next() {
		outlier := &runtimev1.NumericOutliers_Outlier{}
		err = outlierResults.Scan(&outlier.Bucket, &outlier.Low, &outlier.High, &outlier.Present, &outlier.Count)
		if err != nil {
			return nil, err
		}
		outlierBins = append(outlierBins, outlier)
	}

	return &runtimev1.GetRugHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericOutliers{
				NumericOutliers: &runtimev1.NumericOutliers{
					Outliers: outlierBins,
				},
			},
		},
	}, nil
}

func (s *Server) GetTimeRangeSummary(ctx context.Context, request *runtimev1.GetTimeRangeSummaryRequest) (*runtimev1.GetTimeRangeSummaryResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf("SELECT min(%[1]s) as min, max(%[1]s) as max, max(%[1]s) - min(%[1]s) as interval FROM %[2]s",
			sanitizedColumnName, request.TableName),
		Priority: int(request.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}
		if v := rowMap["min"]; v != nil {
			summary.Min = timestamppb.New(v.(time.Time))
			summary.Max = timestamppb.New(rowMap["max"].(time.Time))
			summary.Interval, err = handleInterval(rowMap["interval"])
			if err != nil {
				return nil, err
			}
		}
		return &runtimev1.GetTimeRangeSummaryResponse{
			TimeRangeSummary: summary,
		}, nil
	}
	return nil, status.Error(codes.Internal, "no rows returned")
}

func handleInterval(interval any) (*runtimev1.TimeRangeSummary_Interval, error) {
	switch i := interval.(type) {
	case duckdb.Interval:
		var result = new(runtimev1.TimeRangeSummary_Interval)
		result.Days = i.Days
		result.Months = i.Months
		result.Micros = i.Micros
		return result, nil
	case int64:
		// for date type column interval is difference in num days for two dates
		var result = new(runtimev1.TimeRangeSummary_Interval)
		result.Days = int32(i)
		return result, nil
	}
	return nil, fmt.Errorf("cannot handle interval type %T", interval)
}

func (s *Server) GetCardinalityOfColumn(ctx context.Context, request *runtimev1.GetCardinalityOfColumnRequest) (*runtimev1.GetCardinalityOfColumnResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT approx_count_distinct(%s) as count from %s", sanitizedColumnName, request.TableName),
		Priority: int(request.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count float64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return nil, err
		}
		return &runtimev1.GetCardinalityOfColumnResponse{
			CategoricalSummary: &runtimev1.CategoricalSummary{
				Case: &runtimev1.CategoricalSummary_Cardinality{
					Cardinality: count,
				},
			},
		}, nil
	}
	return nil, status.Error(codes.Internal, "no rows returned")
}

func quoteName(columnName string) string {
	return fmt.Sprintf("\"%s\"", columnName)
}
