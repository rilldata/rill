package queries

import (
	"fmt"
	"slices"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
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

// SecurityFromRuntimeQuery extracts security attributes like row filter, accessible fields like dimensions and measures from a runtime.Query.
func SecurityFromRuntimeQuery(query runtime.Query) (string, []string, error) {
	if query == nil {
		return "", nil, nil
	}

	var rowFilter string
	var accessibleFields []string
	var err error

	var filterFields []string

	switch q := query.(type) {
	case *MetricsViewAggregation:
		rowFilter, filterFields, err = rowFilterJSONAndFields(q.Where, q.WhereSQL, q.Filter)
		if err != nil {
			return "", nil, err
		}
		for _, d := range q.Dimensions {
			accessibleFields = append(accessibleFields, d.Name)
		}
		for _, m := range q.Measures {
			accessibleFields = append(accessibleFields, m.Name)
		}
		if q.TimeRange != nil && q.TimeRange.TimeDimension != "" && !slices.Contains(accessibleFields, q.TimeRange.TimeDimension) {
			accessibleFields = append(accessibleFields, q.TimeRange.TimeDimension)
		}
		for _, f := range filterFields {
			if !slices.Contains(accessibleFields, f) {
				accessibleFields = append(accessibleFields, f)
			}
		}
		for _, s := range q.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case *MetricsViewToplist:
		rowFilter, filterFields, err = rowFilterJSONAndFields(q.Where, q.WhereSQL, q.Filter)
		if err != nil {
			return "", nil, err
		}
		if q.DimensionName != "" {
			accessibleFields = append(accessibleFields, q.DimensionName)
		}
		accessibleFields = append(accessibleFields, q.MeasureNames...)
		for _, f := range filterFields {
			if !slices.Contains(accessibleFields, f) {
				accessibleFields = append(accessibleFields, f)
			}
		}
		for _, s := range q.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case *MetricsViewRows:
		rowFilter, filterFields, err = rowFilterJSONAndFields(q.Where, "", q.Filter)
		if err != nil {
			return "", nil, err
		}
		if q.TimeDimension != "" && !slices.Contains(accessibleFields, q.TimeDimension) {
			accessibleFields = append(accessibleFields, q.TimeDimension)
		}
		for _, f := range filterFields {
			if !slices.Contains(accessibleFields, f) {
				accessibleFields = append(accessibleFields, f)
			}
		}
		for _, s := range q.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	case *MetricsViewTimeSeries:
		rowFilter, filterFields, err = rowFilterJSONAndFields(q.Where, q.WhereSQL, q.Filter)
		if err != nil {
			return "", nil, err
		}
		accessibleFields = append(accessibleFields, q.MeasureNames...)
		if q.TimeDimension != "" && !slices.Contains(accessibleFields, q.TimeDimension) {
			accessibleFields = append(accessibleFields, q.TimeDimension)
		}
		for _, f := range filterFields {
			if !slices.Contains(accessibleFields, f) {
				accessibleFields = append(accessibleFields, f)
			}
		}
	case *MetricsViewComparison:
		rowFilter, filterFields, err = rowFilterJSONAndFields(q.Where, q.WhereSQL, q.Filter)
		if err != nil {
			return "", nil, err
		}
		if q.DimensionName != "" {
			accessibleFields = append(accessibleFields, q.DimensionName)
		}
		for _, m := range q.Measures {
			accessibleFields = append(accessibleFields, m.Name)
		}
		if q.TimeRange != nil && q.TimeRange.TimeDimension != "" && !slices.Contains(accessibleFields, q.TimeRange.TimeDimension) {
			accessibleFields = append(accessibleFields, q.TimeRange.TimeDimension)
		}
		for _, f := range filterFields {
			if !slices.Contains(accessibleFields, f) {
				accessibleFields = append(accessibleFields, f)
			}
		}
		for _, s := range q.Sort {
			if !slices.Contains(accessibleFields, s.Name) {
				accessibleFields = append(accessibleFields, s.Name)
			}
		}
	default:
		return "", nil, fmt.Errorf("query type %T not supported for security extraction", query)
	}

	return rowFilter, accessibleFields, nil
}

// rowFilterJSONAndFields builds a row filter JSON from a where expression, where SQL, and/or a filter. It also returns the fields referenced in the row filter expression.
func rowFilterJSONAndFields(where *runtimev1.Expression, whereSQL string, filter *runtimev1.MetricsViewFilter) (string, []string, error) {
	if filter != nil { // Backwards compatibility
		if where != nil {
			return "", nil, fmt.Errorf("both filter and where is provided")
		}
		where = convertFilterToExpression(filter)
	}
	var whereSQLExp *runtimev1.Expression
	if whereSQL != "" {
		mvExp, err := metricssql.ParseFilter(whereSQL)
		if err != nil {
			return "", nil, fmt.Errorf("invalid where SQL: %w", err)
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

	if where == nil {
		return "", nil, nil
	}

	b, err := protojson.Marshal(where)
	if err != nil {
		return "", nil, fmt.Errorf("invalid where expression: %w", err)
	}

	fields := metricsview.AnalyzeExpressionFields(metricsview.NewExpressionFromProto(where))

	return string(b), fields, nil
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
