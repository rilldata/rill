package server

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultK = 50
const defaultAgg = "count(*)"

func (s *Server) GetTopK(ctx context.Context, topKRequest *api.TopKRequest) (*api.CategoricalSummary, error) {
	agg := defaultAgg
	k := int32(defaultK)
	if topKRequest.Agg != nil {
		agg = *topKRequest.Agg
	}
	if topKRequest.K != nil {
		k = *topKRequest.K
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

	topKResponse := api.TopKResponse{
		Entries: make([]*api.TopKResponse_TopKEntry, 0),
	}
	for rows.Next() {
		var topKEntry api.TopKResponse_TopKEntry
		err := rows.Scan(&topKEntry.Value, &topKEntry.Count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		topKResponse.Entries = append(topKResponse.Entries, &topKEntry)
	}
	return &api.CategoricalSummary{
		TopKResponse: &topKResponse,
	}, nil
}

func (s *Server) GetNullCount(ctx context.Context, nullCountRequest *api.NullCountRequest) (*api.NullCountResponse, error) {
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

	var count int64
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := &api.NullCountResponse{
		Count: count,
	}
	return resp, nil
}

func (s *Server) GetDescriptiveStatistics(ctx context.Context, request *api.DescriptiveStatisticsRequest) (*api.NumericSummary, error) {
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

	stats := new(api.NumericStatistics)
	for rows.Next() {
		err := rows.Scan(&stats.Min, &stats.Q25, &stats.Q50, &stats.Q75, &stats.Max, &stats.Mean, &stats.Sd)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	resp := &api.NumericSummary{
		NumericStatistics: stats,
	}
	return resp, nil
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

func (s *Server) EstimateSmallestTimeGrain(ctx context.Context, request *api.EstimateSmallestTimeGrainRequest) (*api.EstimateSmallestTimeGrainResponse, error) {
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
	defer rows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var timeGrainString string
	for rows.Next() {
		err := rows.Scan(&timeGrainString)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	var timeGrain *api.EstimateSmallestTimeGrainResponse
	switch timeGrainString {
	case "milliseconds":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_MILLISECONDS,
		}
	case "seconds":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_SECONDS,
		}
	case "minutes":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_MINUTES,
		}
	case "hours":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_HOURS,
		}
	case "days":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_DAYS,
		}
	case "weeks":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_WEEKS,
		}
	case "months":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_MONTHS,
		}
	case "years":
		timeGrain = &api.EstimateSmallestTimeGrainResponse{
			TimeGrain: api.EstimateSmallestTimeGrainResponse_YEARS,
		}
	}
	return timeGrain, nil
}

func (s *Server) GetNumericHistogram(ctx context.Context, request *api.NumericHistogramRequest) (*api.NumericSummary, error) {
	return &api.NumericSummary{
		NumericHistogram: &api.NumericHistogram{
			Buckets: []*api.NumericHistogram_Bucket{},
		},
	}, nil
}

func (s *Server) GetRugHistogram(ctx context.Context, request *api.RugHistogramRequest) (*api.NumericSummary, error) {
	return &api.NumericSummary{
		NumericOutliers: &api.NumericOutliers{
			Outliers: []*api.NumericOutliers_Outlier{},
		},
	}, nil
}

func (s *Server) GetTimeRangeSummary(ctx context.Context, request *api.TimeRangeSummaryRequest) (*api.TimeRangeSummary, error) {
	return &api.TimeRangeSummary{}, nil
}

func (s *Server) GetCardinalityOfColumn(ctx context.Context, request *api.CardinalityOfColumnRequest) (*api.CategoricalSummary, error) {
	return &api.CategoricalSummary{
		Cardinality: new(int64),
	}, nil
}

func quoteName(columnName string) string {
	return fmt.Sprintf("\"%s\"", columnName)
}
