package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// MetricsViewMeta implements RuntimeService
func (s *Server) MetricsViewMeta(
	ctx context.Context,
	req *api.MetricsViewMetaRequest,
) (*api.MetricsViewMetaResponse, error) {
	// NOTE: Mock implementation

	dimensions := []*api.MetricsView_Dimension{
		{Name: "time", Type: "TIMESTAMP", PrimaryTime: true},
		{Name: "foo", Type: "VARCHAR"},
	}

	measures := []*api.MetricsView_Measure{
		{Name: "bar", Type: "DOUBLE"},
		{Name: "baz", Type: "INTEGER"},
	}

	resp := &api.MetricsViewMetaResponse{
		MetricsViewName: req.MetricsViewName,
		Dimensions:      dimensions,
		Measures:        measures,
	}

	return resp, nil
}

// MetricsViewToplist implements RuntimeService
func (s *Server) MetricsViewToplist(
	ctx context.Context,
	req *api.MetricsViewToplistRequest,
) (*api.MetricsViewToplistResponse, error) {
	// NOTE: Mock implementation

	sql, args, err := buildMetricsTopListSql(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error building query: %s", err.Error())
	}

	meta, data, err := s.runQuery(ctx, req.InstanceId, sql, args)
	if err != nil {
		return nil, err
	}

	resp := &api.MetricsViewToplistResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

// MetricsViewTimeSeries implements RuntimeService
func (s *Server) MetricsViewTimeSeries(
	ctx context.Context,
	req *api.MetricsViewTimeSeriesRequest,
) (*api.MetricsViewTimeSeriesResponse, error) {
	// NOTE: Partially mocked - timestamp column is hardcoded

	sql, args, err := buildMetricsTimeSeriesSQL(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error building query: %s", err.Error())
	}

	meta, data, err := s.runQuery(ctx, req.InstanceId, sql, args)
	if err != nil {
		return nil, err
	}

	resp := &api.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

// MetricsViewTotals implements RuntimeService
func (s *Server) MetricsViewTotals(
	ctx context.Context,
	req *api.MetricsViewTotalsRequest,
) (*api.MetricsViewTotalsResponse, error) {
	sql, args, err := buildMetricsTotalsSql(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error building query: %s", err.Error())
	}

	meta, data, err := s.runQuery(ctx, req.InstanceId, sql, args)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, status.Errorf(codes.Internal, "no rows received from totals query")
	}

	resp := &api.MetricsViewTotalsResponse{
		Meta: meta,
		Data: data[0],
	}

	return resp, nil
}

func (s *Server) runQuery(
	ctx context.Context,
	instanceId string,
	sql string,
	args []any,
) ([]*api.MetricsViewColumn, []*structpb.Struct, error) {
	rows, err := s.query(ctx, instanceId, &drivers.Statement{
		Query: sql,
		Args:  args,
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

func buildMetricsTimeSeriesSQL(req *api.MetricsViewTimeSeriesRequest) (string, []any, error) {
	// TODO: get from Catalog
	timeField := "timestamp"
	timeCol := fmt.Sprintf("DATE_TRUNC('%s', %s) AS %s", req.TimeGranularity, timeField, timeField)
	selectCols := append([]string{timeCol}, req.MeasureNames...)

	whereClause := fmt.Sprintf("%s >= epoch_ms(?) AND %s < epoch_ms(?) ", timeField, timeField)
	args := []any{req.TimeStart, req.TimeEnd}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s GROUP BY %s LIMIT 1000",
		strings.Join(selectCols, ", "),
		req.MetricsViewName,
		whereClause,
		timeField,
	)
	return sql, args, nil
}

func buildMetricsTopListSql(req *api.MetricsViewToplistRequest) (string, []any, error) {
	// TODO: get from Catalog
	timeField := "timestamp"
	selectCols := append([]string{req.DimensionName}, req.MeasureNames...)
	whereClause := fmt.Sprintf("%s >= epoch_ms(?) AND %s < epoch_ms(?) ", timeField, timeField)
	args := []any{req.TimeStart, req.TimeEnd}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	if req.Sort != nil {
		// TODO
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s LIMIT %d",
		strings.Join(selectCols, ", "),
		req.MetricsViewName, whereClause, req.DimensionName, req.Limit)
	return sql, args, nil
}

func buildMetricsTotalsSql(req *api.MetricsViewTotalsRequest) (string, []any, error) {
	// TODO: get from Catalog
	timeField := "timestamp"
	whereClause := fmt.Sprintf("%s >= epoch_ms(?) AND %s < epoch_ms(?) ", timeField, timeField)
	args := []any{req.TimeStart, req.TimeEnd}

	if req.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(req.Filter)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(req.MeasureNames, ", "), req.MetricsViewName, whereClause)
	return sql, args, nil
}

// Builds clause and args for api.MetricsViewFilter
func buildFilterClauseForMetricsViewFilter(filter *api.MetricsViewFilter) (string, []any, error) {
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

func buildFilterClauseForConditions(conds []*api.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
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

func buildFilterClauseForCondition(cond *api.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
	var clauses []string
	var args []any

	var operatorPrefix string
	var conditionJoiner string
	if exclude {
		operatorPrefix = "NOT"
		conditionJoiner = "AND"
	} else {
		operatorPrefix = ""
		conditionJoiner = "OR"
	}

	if len(cond.In) > 0 {
		// null values should be added with IS NULL / IS NOT NULL
		nullCount := 0
		for _, val := range cond.In {
			if val == nil {
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
		clauses = append(clauses, fmt.Sprintf("%s %s IN (%s)", cond.Name, operatorPrefix, questionMarks))
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
		clause = fmt.Sprintf(" AND %s", strings.Join(clauses, conditionJoiner))
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

func structTypeToMetricsViewColumn(v *api.StructType) []*api.MetricsViewColumn {
	res := make([]*api.MetricsViewColumn, len(v.Fields))
	for i, f := range v.Fields {
		res[i] = &api.MetricsViewColumn{
			Name:     f.Name,
			Type:     f.Type.Code.String(),
			Nullable: f.Type.Nullable,
		}
	}
	return res
}
