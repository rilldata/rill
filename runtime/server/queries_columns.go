package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ColumnTopK(ctx context.Context, req *runtimev1.ColumnTopKRequest) (*runtimev1.ColumnTopKResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.k", int(req.K)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
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
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
		Agg:            agg,
		K:              k,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.ColumnTopKResponse{
		CategoricalSummary: &runtimev1.CategoricalSummary{
			Case: &runtimev1.CategoricalSummary_TopK{
				TopK: q.Result,
			},
		},
	}, nil
}

func (s *Server) ColumnNullCount(ctx context.Context, req *runtimev1.ColumnNullCountRequest) (*runtimev1.ColumnNullCountResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnNullCount{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.ColumnNullCountResponse{
		Count: q.Result,
	}, nil
}

func (s *Server) ColumnDescriptiveStatistics(ctx context.Context, req *runtimev1.ColumnDescriptiveStatisticsRequest) (*runtimev1.ColumnDescriptiveStatisticsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnDescriptiveStatistics{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// ColumnDescriptiveStatistics may return an empty result
	if q.Result == nil {
		return &runtimev1.ColumnDescriptiveStatisticsResponse{}, nil
	}

	return &runtimev1.ColumnDescriptiveStatisticsResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericStatistics{
				NumericStatistics: q.Result,
			},
		},
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

func (s *Server) ColumnTimeGrain(ctx context.Context, req *runtimev1.ColumnTimeGrainRequest) (*runtimev1.ColumnTimeGrainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeGrain{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.ColumnTimeGrainResponse{
		TimeGrain: q.Result,
	}, nil
}

func (s *Server) ColumnNumericHistogram(ctx context.Context, req *runtimev1.ColumnNumericHistogramRequest) (*runtimev1.ColumnNumericHistogramResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.table", req.TableName),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.column", req.ColumnName),
		attribute.String("args.histogram", req.HistogramMethod.String()),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnNumericHistogram{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
		Method:         req.HistogramMethod,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// NOTE: q.Result may be nil if there were no bins. The below will output it as an empty histogram.

	return &runtimev1.ColumnNumericHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericHistogramBins{
				NumericHistogramBins: &runtimev1.NumericHistogramBins{
					Bins: q.Result,
				},
			},
		},
	}, nil
}

func (s *Server) ColumnRugHistogram(ctx context.Context, req *runtimev1.ColumnRugHistogramRequest) (*runtimev1.ColumnRugHistogramResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnRugHistogram{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.ColumnRugHistogramResponse{
		NumericSummary: &runtimev1.NumericSummary{
			Case: &runtimev1.NumericSummary_NumericOutliers{
				NumericOutliers: &runtimev1.NumericOutliers{
					Outliers: q.Result,
				},
			},
		},
	}, nil
}

func (s *Server) ColumnTimeRange(ctx context.Context, req *runtimev1.ColumnTimeRangeRequest) (*runtimev1.ColumnTimeRangeResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeRange{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return &runtimev1.ColumnTimeRangeResponse{
		TimeRangeSummary: q.Result,
	}, nil
}

func (s *Server) ColumnCardinality(ctx context.Context, req *runtimev1.ColumnCardinalityRequest) (*runtimev1.ColumnCardinalityResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnCardinality{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		ColumnName:     req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return &runtimev1.ColumnCardinalityResponse{
		CategoricalSummary: &runtimev1.CategoricalSummary{
			Case: &runtimev1.CategoricalSummary_Cardinality{
				Cardinality: q.Result,
			},
		},
	}, nil
}
