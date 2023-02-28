package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetTopK(ctx context.Context, req *runtimev1.GetTopKRequest) (*runtimev1.GetTopKResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	agg := "count(*)"
	if req.Agg != "" {
		agg = req.Agg
	}

	k := 50
	if req.K != 0 {
		k = int(req.K)
	}

	q := &queries.ColumnTopK{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
		Agg:        agg,
		K:          k,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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

func (s *Server) GetNullCount(ctx context.Context, req *runtimev1.GetNullCountRequest) (*runtimev1.GetNullCountResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnNullCount{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.GetNullCountResponse{
		Count: q.Result,
	}, nil
}

func (s *Server) GetDescriptiveStatistics(ctx context.Context, req *runtimev1.GetDescriptiveStatisticsRequest) (*runtimev1.GetDescriptiveStatisticsResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnDescriptiveStatistics{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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

func (s *Server) EstimateSmallestTimeGrain(ctx context.Context, req *runtimev1.EstimateSmallestTimeGrainRequest) (*runtimev1.EstimateSmallestTimeGrainResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeGrain{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.EstimateSmallestTimeGrainResponse{
		TimeGrain: q.Result,
	}, nil
}

func (s *Server) GetNumericHistogram(ctx context.Context, req *runtimev1.GetNumericHistogramRequest) (*runtimev1.GetNumericHistogramResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnNumericHistogram{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
		Method:     req.HistogramMethod,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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

func (s *Server) GetRugHistogram(ctx context.Context, req *runtimev1.GetRugHistogramRequest) (*runtimev1.GetRugHistogramResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnRugHistogram{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.GetRugHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericOutliers{
				NumericOutliers: &runtimev1.NumericOutliers{
					Outliers: q.Result,
				},
			},
		},
	}, nil
}

func (s *Server) GetTimeRangeSummary(ctx context.Context, req *runtimev1.GetTimeRangeSummaryRequest) (*runtimev1.GetTimeRangeSummaryResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeRange{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.GetTimeRangeSummaryResponse{
		TimeRangeSummary: q.Result,
	}, nil
}

func (s *Server) GetCardinalityOfColumn(ctx context.Context, req *runtimev1.GetCardinalityOfColumnRequest) (*runtimev1.GetCardinalityOfColumnResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnCardinality{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return &runtimev1.GetCardinalityOfColumnResponse{
		CategoricalSummary: &runtimev1.CategoricalSummary{
			Case: &runtimev1.CategoricalSummary_Cardinality{
				Cardinality: q.Result,
			},
		},
	}, nil
}
