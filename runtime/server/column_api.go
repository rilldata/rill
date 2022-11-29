package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/marcboeker/go-duckdb"
	"google.golang.org/protobuf/types/known/structpb"
	"math"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultK = 50
const defaultAgg = "count(*)"

func (s *Server) GetTopK(ctx context.Context, topKRequest *runtimev1.GetTopKRequest) (*runtimev1.GetTopKResponse, error) {
	agg := defaultAgg
	k := int32(defaultK)
	if topKRequest.Agg != "" {
		agg = topKRequest.Agg
	}
	if topKRequest.K != 0 {
		k = topKRequest.K
	}
	topKSql := fmt.Sprintf("SELECT %s as value, %s AS count from %s GROUP BY %s ORDER BY count desc LIMIT %d",
		quoteName(topKRequest.ColumnName),
		agg,
		topKRequest.TableName,
		quoteName(topKRequest.ColumnName),
		k,
	)
	rows, err := s.query(ctx, topKRequest.InstanceId, &drivers.Statement{
		Query: topKSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	topKResponse := runtimev1.TopK{
		Entries: make([]*runtimev1.TopK_Entry, 0),
	}
	for rows.Next() {
		var topKEntry runtimev1.TopK_Entry
		var value sql.NullString
		err = rows.Scan(&value, &topKEntry.Count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if value.Valid {
			topKEntry.Value = structpb.NewStringValue(value.String)
		} else {
			topKEntry.Value = structpb.NewNullValue()
		}
		topKResponse.Entries = append(topKResponse.Entries, &topKEntry)
	}

	return &runtimev1.GetTopKResponse{
		CategoricalSummary: &runtimev1.CategoricalSummary{
			Case: &runtimev1.CategoricalSummary_TopK{
				TopK: &topKResponse,
			},
		},
	}, nil
}

func (s *Server) GetNullCount(ctx context.Context, nullCountRequest *runtimev1.GetNullCountRequest) (*runtimev1.GetNullCountResponse, error) {
	nullCountSql := fmt.Sprintf("SELECT count(*) as count from %s WHERE %s IS NULL",
		nullCountRequest.TableName,
		quoteName(nullCountRequest.ColumnName),
	)
	rows, err := s.query(ctx, nullCountRequest.InstanceId, &drivers.Statement{
		Query: nullCountSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var count float64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &runtimev1.GetNullCountResponse{
		Count: count,
	}, nil
}

func (s *Server) GetDescriptiveStatistics(ctx context.Context, request *runtimev1.GetDescriptiveStatisticsRequest) (*runtimev1.GetDescriptiveStatisticsResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	descriptiveStatisticsSql := fmt.Sprintf("SELECT "+
		"min(%s) as min, "+
		"approx_quantile(%s, 0.25) as q25, "+
		"approx_quantile(%s, 0.5)  as q50, "+
		"approx_quantile(%s, 0.75) as q75, "+
		"max(%s) as max, "+
		"avg(%s)::FLOAT as mean, "+
		"stddev_pop(%s) as sd "+
		"FROM %s",
		sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, request.TableName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: descriptiveStatisticsSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	stats := new(runtimev1.NumericStatistics)
	for rows.Next() {
		err = rows.Scan(&stats.Min, &stats.Q25, &stats.Q50, &stats.Q75, &stats.Max, &stats.Mean, &stats.Sd)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	resp := &runtimev1.NumericSummary{
		Case: &runtimev1.NumericSummary_NumericStatistics{
			NumericStatistics: stats,
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
	sampleSize := int64(500000)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf("SELECT count(*) as c FROM %s", request.TableName),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var totalRows int64
	for rows.Next() {
		err := rows.Scan(&totalRows)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	rows.Close()
	var useSample string
	if sampleSize > totalRows {
		useSample = ""
	} else {
		useSample = fmt.Sprintf("USING SAMPLE %d ROWS", sampleSize)
	}

	estimateSql := fmt.Sprintf(`
      WITH cleaned_column AS (
          SELECT %s as cd
          from %s
          %s
      ),
      time_grains as (
      SELECT 
          approx_count_distinct(extract('years' from cd)) as year,
          approx_count_distinct(extract('months' from cd)) as month,
          approx_count_distinct(extract('dayofyear' from cd)) as dayofyear,
          approx_count_distinct(extract('dayofmonth' from cd)) as dayofmonth,
          min(cd = last_day(cd)) = TRUE as lastdayofmonth,
          approx_count_distinct(extract('weekofyear' from cd)) as weekofyear,
          approx_count_distinct(extract('dayofweek' from cd)) as dayofweek,
          approx_count_distinct(extract('hour' from cd)) as hour,
          approx_count_distinct(extract('minute' from cd)) as minute,
          approx_count_distinct(extract('second' from cd)) as second,
          approx_count_distinct(extract('millisecond' from cd) - extract('seconds' from cd) * 1000) as ms
      FROM cleaned_column
      )
      SELECT 
        COALESCE(
            case WHEN ms > 1 THEN 'milliseconds' else NULL END,
            CASE WHEN second > 1 THEN 'seconds' else NULL END,
            CASE WHEN minute > 1 THEN 'minutes' else null END,
            CASE WHEN hour > 1 THEN 'hours' else null END,
            -- cases above, if equal to 1, then we have some candidates for
            -- bigger time grains. We need to reverse from here
            -- years, months, weeks, days.
            CASE WHEN dayofyear = 1 and year > 1 THEN 'years' else null END,
            CASE WHEN (dayofmonth = 1 OR lastdayofmonth) and month > 1 THEN 'months' else null END,
            CASE WHEN dayofweek = 1 and weekofyear > 1 THEN 'weeks' else null END,
            CASE WHEN hour = 1 THEN 'days' else null END
        ) as estimatedSmallestTimeGrain
      FROM time_grains
      `, quoteName(request.ColumnName), request.TableName, useSample)
	rows, err = s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: estimateSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var timeGrainString string
	for rows.Next() {
		err := rows.Scan(&timeGrainString)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	var timeGrain *runtimev1.EstimateSmallestTimeGrainResponse
	switch timeGrainString {
	case "milliseconds":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND,
		}
	case "seconds":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		}
	case "minutes":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		}
	case "hours":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		}
	case "days":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		}
	case "weeks":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		}
	case "months":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		}
	case "years":
		timeGrain = &runtimev1.EstimateSmallestTimeGrainResponse{
			TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		}
	}
	return timeGrain, nil
}

func (s *Server) GetNumericHistogram(ctx context.Context, request *runtimev1.GetNumericHistogramRequest) (*runtimev1.GetNumericHistogramResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	sql := fmt.Sprintf("SELECT approx_quantile(%s, 0.75)-approx_quantile(%s, 0.25) as IQR, approx_count_distinct(%s) as count, max(%s) - min(%s) as range FROM %s",
		sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, sanitizedColumnName, request.TableName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: sql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	var iqr, count, rangeVal float64
	for rows.Next() {
		err = rows.Scan(&iqr, &count, &rangeVal)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	var bucketSize float64
	if count < 40 {
		// Use cardinality if unique count less than 40
		bucketSize = count
	} else {
		// Use Freedman–Diaconis rule for calculating number of bins
		bucketWidth := (2 * iqr) / math.Cbrt(count)
		FDEstimatorBucketSize := math.Ceil(rangeVal / bucketWidth)
		bucketSize = math.Min(40, FDEstimatorBucketSize)
	}
	selectColumn := fmt.Sprintf("%s::DOUBLE", sanitizedColumnName)

	histogramSql := fmt.Sprintf(`
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
	      `, selectColumn, sanitizedColumnName, request.TableName, bucketSize)
	histogramRows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: histogramSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer histogramRows.Close()
	histogramBins := make([]*runtimev1.NumericHistogramBins_Bin, 0)
	for histogramRows.Next() {
		bin := &runtimev1.NumericHistogramBins_Bin{}
		err = histogramRows.Scan(&bin.Bucket, &bin.Low, &bin.High, &bin.Count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		histogramBins = append(histogramBins, bin)
	}
	return &runtimev1.GetNumericHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericHistogramBins{
				NumericHistogramBins: &runtimev1.NumericHistogramBins{
					Bins: histogramBins,
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
            CASE WHEN count>0 THEN true ELSE false END AS present
          FROM histrogram_with_edge
          WHERE present=true`, selectColumn, sanitizedColumnName, request.TableName, outlierPseudoBucketSize)

	outlierResults, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: rugSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer outlierResults.Close()

	outlierBins := make([]*runtimev1.NumericOutliers_Outlier, 0)
	for outlierResults.Next() {
		outlier := &runtimev1.NumericOutliers_Outlier{}
		err = outlierResults.Scan(&outlier.Bucket, &outlier.Low, &outlier.High, &outlier.Present)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
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
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	if rows.Next() {
		summary := &runtimev1.TimeRangeSummary{}
		rowMap := make(map[string]any)
		err = rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}
		summary.Min = timestamppb.New(rowMap["min"].(time.Time))
		summary.Max = timestamppb.New(rowMap["max"].(time.Time))
		summary.Interval, err = handleInterval(rowMap["interval"])
		if err != nil {
			return nil, err
		}
		return &runtimev1.GetTimeRangeSummaryResponse{
			TimeRangeSummary: summary,
		}, nil
	}
	return nil, status.Error(codes.Internal, "no rows returned")
}

func handleInterval(interval any) (*runtimev1.TimeRangeSummary_Interval, error) {
	switch interval.(type) {
	case duckdb.Interval:
		duckDbInterval := interval.(duckdb.Interval)
		var result = new(runtimev1.TimeRangeSummary_Interval)
		result.Days = duckDbInterval.Days
		result.Months = duckDbInterval.Months
		result.Micros = duckDbInterval.Micros
		return result, nil
	case int64:
		days := interval.(int64)
		var result = new(runtimev1.TimeRangeSummary_Interval)
		result.Days = int32(days)
		return result, nil
	}
	return nil, fmt.Errorf("cannot handle interval type %T", interval)
}

func (s *Server) GetCardinalityOfColumn(ctx context.Context, request *runtimev1.GetCardinalityOfColumnRequest) (*runtimev1.GetCardinalityOfColumnResponse, error) {
	sanitizedColumnName := quoteName(request.ColumnName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf("SELECT approx_count_distinct(%s) as count from %s", sanitizedColumnName, request.TableName),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	var count float64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
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
