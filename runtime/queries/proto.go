package queries

import (
	"fmt"
	"slices"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToQuery builds a runtime query from a proto query and security attributes.
// NOTE: Pending refactors, this implementation is replicated from handlers in runtime/server.
func ProtoToQuery(q *runtimev1.Query, claims *runtime.SecurityClaims, executionTime *time.Time) (runtime.Query, error) {
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
			MetricsViewName:     req.MetricsView,
			Dimensions:          req.Dimensions,
			Measures:            req.Measures,
			Sort:                req.Sort,
			TimeRange:           tr,
			ComparisonTimeRange: req.ComparisonTimeRange,
			Where:               req.Where,
			Having:              req.Having,
			Filter:              req.Filter,
			Offset:              req.Offset,
			PivotOn:             req.PivotOn,
			SecurityClaims:      claims,
			ExecutionTime:       executionTime,
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
			SecurityClaims:      claims,
			ExecutionTime:       executionTime,
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
			if req.ComparisonTimeRange != nil {
				req.ComparisonTimeRange = overrideTimeRange(req.ComparisonTimeRange, *executionTime)
			}
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
			if req.ComparisonTimeRange != nil {
				req.ComparisonTimeRange = overrideTimeRange(req.ComparisonTimeRange, *executionTime)
			}
		}
	default:
		return nil, fmt.Errorf("query %q not supported for reports", qryName)
	}

	return qry, nil
}

// MetricsViewFromQuery extracts the metrics view name from a JSON query based on the query name.
func MetricsViewFromQuery(qryName, qryArgsJSON string) (string, error) {
	qry := &runtimev1.Query{}
	var metricsView string
	switch qryName {
	case "MetricsViewAggregation":
		req := &runtimev1.MetricsViewAggregationRequest{}
		qry.Query = &runtimev1.Query_MetricsViewAggregationRequest{MetricsViewAggregationRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		metricsView = req.MetricsView
	case "MetricsViewToplist":
		req := &runtimev1.MetricsViewToplistRequest{}
		qry.Query = &runtimev1.Query_MetricsViewToplistRequest{MetricsViewToplistRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		metricsView = req.MetricsViewName
	case "MetricsViewRows":
		req := &runtimev1.MetricsViewRowsRequest{}
		qry.Query = &runtimev1.Query_MetricsViewRowsRequest{MetricsViewRowsRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		metricsView = req.MetricsViewName
	case "MetricsViewTimeSeries":
		req := &runtimev1.MetricsViewTimeSeriesRequest{}
		qry.Query = &runtimev1.Query_MetricsViewTimeSeriesRequest{MetricsViewTimeSeriesRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		metricsView = req.MetricsViewName
	case "MetricsViewComparison":
		req := &runtimev1.MetricsViewComparisonRequest{}
		qry.Query = &runtimev1.Query_MetricsViewComparisonRequest{MetricsViewComparisonRequest: req}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}
		metricsView = req.MetricsViewName
	default:
		return "", fmt.Errorf("query %q not supported for reports", qryName)
	}

	return metricsView, nil
}

// SecurityFromQuery extracts security attributes like row filter, accessible fields like dimensions and measures from a JSON query.
func SecurityFromQuery(qryName, qryArgsJSON string) (string, []string, error) {
	if qryName == "" || qryArgsJSON == "" {
		return "", nil, nil
	}

	var rowFilter string
	var accessibleFields []string
	switch qryName {
	case "MetricsViewAggregation":
		req := &runtimev1.MetricsViewAggregationRequest{}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}

		rowFilter, err = rowFilterJSON(req.Where, req.WhereSql, req.Filter)
		if err != nil {
			return "", nil, err
		}
		for _, d := range req.Dimensions {
			accessibleFields = append(accessibleFields, d.Name)
		}
		for _, m := range req.Measures {
			accessibleFields = append(accessibleFields, m.Name)
		}
		if req.TimeRange != nil && req.TimeRange.TimeDimension != "" && !slices.Contains(accessibleFields, req.TimeRange.TimeDimension) {
			accessibleFields = append(accessibleFields, req.TimeRange.TimeDimension)
		}
		for _, s := range req.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case "MetricsViewToplist":
		req := &runtimev1.MetricsViewToplistRequest{}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}

		rowFilter, err = rowFilterJSON(req.Where, req.WhereSql, req.Filter)
		if err != nil {
			return "", nil, err
		}
		if req.DimensionName != "" {
			accessibleFields = append(accessibleFields, req.DimensionName)
		}
		accessibleFields = append(accessibleFields, req.MeasureNames...)
		for _, s := range req.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case "MetricsViewRows":
		req := &runtimev1.MetricsViewRowsRequest{}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}

		rowFilter, err = rowFilterJSON(req.Where, "", req.Filter)
		if err != nil {
			return "", nil, err
		}
		if req.TimeDimension != "" && !slices.Contains(accessibleFields, req.TimeDimension) {
			accessibleFields = append(accessibleFields, req.TimeDimension)
		}
		for _, s := range req.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case "MetricsViewTimeSeries":
		req := &runtimev1.MetricsViewTimeSeriesRequest{}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}

		rowFilter, err = rowFilterJSON(req.Where, req.WhereSql, req.Filter)
		if err != nil {
			return "", nil, err
		}
		accessibleFields = append(accessibleFields, req.MeasureNames...)
		if req.TimeDimension != "" && !slices.Contains(accessibleFields, req.TimeDimension) {
			accessibleFields = append(accessibleFields, req.TimeDimension)
		}
	case "MetricsViewComparison":
		req := &runtimev1.MetricsViewComparisonRequest{}
		err := protojson.Unmarshal([]byte(qryArgsJSON), req)
		if err != nil {
			return "", nil, fmt.Errorf("invalid properties for query %q: %w", qryName, err)
		}

		rowFilter, err = rowFilterJSON(req.Where, req.WhereSql, req.Filter)
		if err != nil {
			return "", nil, err
		}
		if req.Dimension != nil {
			accessibleFields = append(accessibleFields, req.Dimension.Name)
		}
		for _, m := range req.Measures {
			accessibleFields = append(accessibleFields, m.Name)
		}
		if req.TimeRange != nil && req.TimeRange.TimeDimension != "" && !slices.Contains(accessibleFields, req.TimeRange.TimeDimension) {
			accessibleFields = append(accessibleFields, req.TimeRange.TimeDimension)
		}
		for _, s := range req.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	default:
		return "", nil, fmt.Errorf("query %q not supported for reports", qryName)
	}

	return rowFilter, accessibleFields, nil
}

func rowFilterJSON(where *runtimev1.Expression, whereSQL string, filter *runtimev1.MetricsViewFilter) (string, error) {
	if filter != nil { // Backwards compatibility
		if where != nil {
			return "", fmt.Errorf("both filter and where is provided")
		}
		where = convertFilterToExpression(filter)
	}
	var whereSQLExp *runtimev1.Expression
	if whereSQL != "" {
		mvExp, err := metricssqlparser.ParseSQLFilter(whereSQL)
		if err != nil {
			return "", fmt.Errorf("invalid where SQL: %w", err)
		}
		whereSQLExp = metricsview.ExpressionToProto(mvExp)
	}

	if whereSQLExp != nil && where != nil {
		where = &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_AND,
					Exprs: []*runtimev1.Expression{
						{
							Expression: whereSQLExp.Expression,
						},
						{
							Expression: where.Expression,
						},
					},
				},
			},
		}
	} else if whereSQLExp != nil {
		where = whereSQLExp
	}

	b, err := protojson.Marshal(where)
	if err != nil {
		return "", fmt.Errorf("invalid where expression: %w", err)
	}

	return string(b), nil
}

func overrideTimeRange(tr *runtimev1.TimeRange, t time.Time) *runtimev1.TimeRange {
	if tr == nil {
		tr = &runtimev1.TimeRange{}
	}
	if tr.Expression != "" {
		// Do not add `end` for rill time expressions. Execution time will be passed through to executor.Query.
		return tr
	}

	tr.End = timestamppb.New(t)
	return tr
}
