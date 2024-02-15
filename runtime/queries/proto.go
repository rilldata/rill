package queries

import (
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToQuery builds a runtime query from a proto query and security attributes.
// NOTE: Pending refactors, this implementation is replicated from handlers in runtime/server.
func ProtoToQuery(q *runtimev1.Query, attrs map[string]any) (runtime.Query, error) {
	switch r := q.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		req := r.MetricsViewAggregationRequest

		tr := req.TimeRange
		if req.TimeStart != nil || req.TimeEnd != nil {
			tr = &runtimev1.TimeRange{
				Start: req.TimeStart,
				End:   req.TimeEnd,
			}
		}

		return &MetricsViewAggregation{
			MetricsViewName:    req.MetricsView,
			Dimensions:         req.Dimensions,
			Measures:           req.Measures,
			Sort:               req.Sort,
			TimeRange:          tr,
			Where:              req.Where,
			Having:             req.Having,
			Filter:             req.Filter,
			Offset:             req.Offset,
			PivotOn:            req.PivotOn,
			SecurityAttributes: attrs,
		}, nil
	case *runtimev1.Query_MetricsViewComparisonRequest:
		req := r.MetricsViewComparisonRequest
		return &MetricsViewComparison{
			MetricsViewName:     req.MetricsViewName,
			DimensionName:       req.Dimension.Name,
			Measures:            req.Measures,
			ComparisonMeasures:  req.ComparisonMeasures,
			TimeRange:           req.TimeRange,
			ComparisonTimeRange: req.ComparisonTimeRange,
			Limit:               req.Limit,
			Offset:              req.Offset,
			Sort:                req.Sort,
			Where:               req.Where,
			Having:              req.Having,
			Filter:              req.Filter,
			Exact:               req.Exact,
			SecurityAttributes:  attrs,
		}, nil
	default:
		return nil, fmt.Errorf("query type %T not supported for alerts", r)
	}
}

// ProtoFromJSON builds a proto query from a query name, JSON args, and optional execution time.
func ProtoFromJSON(qryName, qryArgsJSON string, executionTime *time.Time) (*runtimev1.Query, error) {
	qry := &runtimev1.Query{}
	switch qryName {
	case "MetricsViewAggregation":
		req := &runtimev1.MetricsViewAggregationRequest{}
		qry.Query = &runtimev1.Query_MetricsViewAggregationRequest{MetricsViewAggregationRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		if executionTime != nil {
			req.TimeRange = overrideTimeRange(req.TimeRange, *executionTime)
		}
	case "MetricsViewToplist":
		req := &runtimev1.MetricsViewToplistRequest{}
		qry.Query = &runtimev1.Query_MetricsViewToplistRequest{MetricsViewToplistRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
	case "MetricsViewRows":
		req := &runtimev1.MetricsViewRowsRequest{}
		qry.Query = &runtimev1.Query_MetricsViewRowsRequest{MetricsViewRowsRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
	case "MetricsViewTimeSeries":
		req := &runtimev1.MetricsViewTimeSeriesRequest{}
		qry.Query = &runtimev1.Query_MetricsViewTimeSeriesRequest{MetricsViewTimeSeriesRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
	case "MetricsViewComparison":
		req := &runtimev1.MetricsViewComparisonRequest{}
		qry.Query = &runtimev1.Query_MetricsViewComparisonRequest{MetricsViewComparisonRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		if executionTime != nil {
			req.TimeRange = overrideTimeRange(req.TimeRange, *executionTime)
		}
	default:
		return nil, fmt.Errorf("query %q not supported for reports", qryName)
	}

	return qry, nil
}

func overrideTimeRange(tr *runtimev1.TimeRange, t time.Time) *runtimev1.TimeRange {
	if tr == nil {
		tr = &runtimev1.TimeRange{}
	}

	tr.End = timestamppb.New(t)

	return tr
}
