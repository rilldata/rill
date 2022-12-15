package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// NOTE: The queries in here are generally not vetted or fully implemented. Use it as guidelines for the real implementation
// once the metrics view artifact representation is ready.

// MetricsViewToplist implements RuntimeService
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Filter:          req.Filter,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeSeries implements RuntimeService
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *runtimev1.MetricsViewTimeSeriesRequest) (*runtimev1.MetricsViewTimeSeriesResponse, error) {
	mv, err := s.lookupMetricsView(ctx, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	sql, args, err := buildMetricsTimeSeriesSQL(req, mv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error building query: %s", err.Error())
	}

	meta, data, err := s.metricsQuery(ctx, req.InstanceId, int(req.Priority), sql, args)
	if err != nil {
		return nil, err
	}

	resp := &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

// MetricsViewTotals implements RuntimeService
func (s *Server) MetricsViewTotals(ctx context.Context, req *runtimev1.MetricsViewTotalsRequest) (*runtimev1.MetricsViewTotalsResponse, error) {
	q := &queries.MetricsViewTotals{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

func (s *Server) lookupMetricsView(ctx context.Context, instanceID string, name string) (*runtimev1.MetricsView, error) {
	obj, err := s.runtime.GetCatalogEntry(ctx, instanceID, name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if obj.GetMetricsView() == nil {
		return nil, status.Errorf(codes.NotFound, "object named '%s' is not a metrics view", name)
	}

	return obj.GetMetricsView(), nil
}

func (s *Server) metricsQuery(ctx context.Context, instanceId string, priority int, sql string, args []any) ([]*runtimev1.MetricsViewColumn, []*structpb.Struct, error) {
	rows, err := s.query(ctx, instanceId, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: priority,
	})
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return structTypeToMetricsViewColumn(rows.Schema), data, nil
}

func buildMetricsTopListSql(req *runtimev1.MetricsViewToplistRequest, mv *runtimev1.MetricsView) (string, []any, error) {
	dimName := quoteName(req.DimensionName)
	selectCols := []string{dimName}
	for _, n := range req.MeasureNames {
		found := false
		for _, m := range mv.Measures {
			if m.Name == n {
				expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
				selectCols = append(selectCols, expr)
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	args := []any{}
	whereClause := "1=1"
	if mv.TimeDimension != "" {
		if req.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", mv.TimeDimension)
			args = append(args, req.TimeStart.AsTime())
		}
		if req.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", mv.TimeDimension)
			args = append(args, req.TimeEnd.AsTime())
		}
	}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	orderClause := "true"
	for _, s := range req.Sort {
		orderClause += ", "
		orderClause += s.Name
		if !s.Ascending {
			orderClause += " DESC"
		}
		orderClause += " NULLS LAST"
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s ORDER BY %s LIMIT %d",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
		dimName,
		orderClause,
		req.Limit,
	)

	return sql, args, nil
}

func buildMetricsTimeSeriesSQL(req *runtimev1.MetricsViewTimeSeriesRequest, mv *runtimev1.MetricsView) (string, []any, error) {
	timeCol := fmt.Sprintf("DATE_TRUNC('%s', %s) AS %s", req.TimeGranularity, mv.TimeDimension, mv.TimeDimension)
	selectCols := []string{timeCol}
	for _, n := range req.MeasureNames {
		found := false
		for _, m := range mv.Measures {
			if m.Name == n {
				expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
				selectCols = append(selectCols, expr)
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	whereClause := "1=1"
	args := []any{}
	if mv.TimeDimension != "" {
		if req.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", mv.TimeDimension)
			args = append(args, req.TimeStart.AsTime())
		}
		if req.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", mv.TimeDimension)
			args = append(args, req.TimeEnd.AsTime())
		}
	}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s GROUP BY 1 ORDER BY %s LIMIT 1000",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
		mv.TimeDimension,
	)
	return sql, args, nil
}

func buildMetricsTotalsSql(req *runtimev1.MetricsViewTotalsRequest, mv *runtimev1.MetricsView) (string, []any, error) {
	selectCols := []string{}
	for _, n := range req.MeasureNames {
		found := false
		for _, m := range mv.Measures {
			if m.Name == n {
				expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
				selectCols = append(selectCols, expr)
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	whereClause := "1=1"
	args := []any{}
	if mv.TimeDimension != "" {
		if req.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", mv.TimeDimension)
			args = append(args, req.TimeStart.AsTime())
		}
		if req.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", mv.TimeDimension)
			args = append(args, req.TimeEnd.AsTime())
		}
	}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)
	return sql, args, nil
}

// Builds clause and args for runtimev1.MetricsViewFilter
func buildFilterClauseForMetricsViewFilter(filter *runtimev1.MetricsViewFilter) (string, []any, error) {
	whereClause := ""
	var args []any

	if filter != nil && filter.Include != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Include, false)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	if filter != nil && filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Exclude, true)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	return whereClause, args, nil
}

func buildFilterClauseForConditions(conds []*runtimev1.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
	clause := ""
	var args []any

	for _, cond := range conds {
		condClause, condArgs, err := buildFilterClauseForCondition(cond, exclude)
		if err != nil {
			return "", nil, fmt.Errorf("filter error: %s", err.Error())
		}
		if condClause == "" {
			continue
		}
		clause += condClause
		args = append(args, condArgs...)
	}

	return clause, args, nil
}

func buildFilterClauseForCondition(cond *runtimev1.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
	var clauses []string
	var args []any

	var operatorPrefix string
	var conditionJoiner string
	if exclude {
		operatorPrefix = " NOT "
		conditionJoiner = ") AND ("
	} else {
		operatorPrefix = ""
		conditionJoiner = " OR "
	}

	if len(cond.In) > 0 {
		// null values should be added with IS NULL / IS NOT NULL
		nullCount := 0
		for _, val := range cond.In {
			if _, ok := val.Kind.(*structpb.Value_NullValue); ok {
				nullCount++
				continue
			}
			arg, err := protobufValueToAny(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %s", err.Error())
			}
			args = append(args, arg)
		}

		questionMarks := strings.Join(repeatString("?", len(cond.In)-nullCount), ",")
		// <dimension> (NOT) IN (?,?,...)
		if questionMarks != "" {
			clauses = append(clauses, fmt.Sprintf("%s %s IN (%s)", cond.Name, operatorPrefix, questionMarks))
		}
		if nullCount > 0 {
			// <dimension> IS (NOT) NULL
			clauses = append(clauses, fmt.Sprintf("%s IS %s NULL", cond.Name, operatorPrefix))
		}
	}

	if len(cond.Like) > 0 {
		for _, val := range cond.Like {
			arg, err := protobufValueToAny(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %s", err.Error())
			}
			args = append(args, arg)
			// <dimension> (NOT) ILIKE ?
			clauses = append(clauses, fmt.Sprintf("%s %s ILIKE ?", cond.Name, operatorPrefix))
		}
	}

	clause := ""
	if len(clauses) > 0 {
		clause = fmt.Sprintf(" AND (%s)", strings.Join(clauses, conditionJoiner))
	}
	return clause, args, nil
}

func repeatString(val string, n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = val
	}
	return res
}

func protobufValueToAny(val *structpb.Value) (any, error) {
	switch v := val.GetKind().(type) {
	case *structpb.Value_StringValue:
		return v.StringValue, nil
	case *structpb.Value_BoolValue:
		return v.BoolValue, nil
	case *structpb.Value_NumberValue:
		return v.NumberValue, nil
	case *structpb.Value_NullValue:
		return nil, nil
	default:
		return nil, fmt.Errorf("value not supported: %v", v)
	}
}

func structTypeToMetricsViewColumn(v *runtimev1.StructType) []*runtimev1.MetricsViewColumn {
	res := make([]*runtimev1.MetricsViewColumn, len(v.Fields))
	for i, f := range v.Fields {
		res[i] = &runtimev1.MetricsViewColumn{
			Name:     f.Name,
			Type:     f.Type.Code.String(),
			Nullable: f.Type.Nullable,
		}
	}
	return res
}

func quoteName(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}
