package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	resp := &api.PingResponse{
		Message: "Pong",
	}
	return resp, nil
}

// CreateInstance implements RuntimeService
func (s *Server) CreateInstance(ctx context.Context, req *api.CreateInstanceRequest) (*api.CreateInstanceResponse, error) {
	instance, err := s.runtime.CreateInstance(req.Driver, req.Dsn)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp := &api.CreateInstanceResponse{
		InstanceId: instance.ID.String(),
	}

	return resp, nil
}

// QueryDirect implements RuntimeService
func (s *Server) QueryDirect(ctx context.Context, req *api.QueryDirectRequest) (*api.QueryDirectResponse, error) {
	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    req.Sql,
		Args:     args,
		DryRun:   req.DryRun,
		Priority: int(req.Priority),
	})
	if err != nil {
		// TODO: Parse error to determine error code
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.DryRun {
		// NOTE: Currently, dry run queries return nil rows
		return &api.QueryDirectResponse{}, nil
	}

	defer rows.Close()

	meta, err := rowsToMeta(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	data, err := rowsToData(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.QueryDirectResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

// MetricsViewMeta implements RuntimeService
func (s *Server) MetricsViewMeta(ctx context.Context, req *api.MetricsViewMetaRequest) (*api.MetricsViewMetaResponse, error) {
	// NOTE: Mock implementation

	dimensions := []*api.MetricsViewDimension{
		{Name: "time", Type: "TIMESTAMP", PrimaryTime: true},
		{Name: "foo", Type: "VARCHAR"},
	}

	measures := []*api.MetricsViewMeasure{
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
func (s *Server) MetricsViewToplist(ctx context.Context, req *api.MetricsViewToplistRequest) (*api.MetricsViewToplistResponse, error) {
	// NOTE: Mock implementation

	sql := `
		SELECT
			TIMESTAMP '1992-09-20 11:30:00' AS time,
			'hello' AS foo,
			3.14 AS bar,
			314 AS baz
		LIMIT ? OFFSET ?
	`

	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: sql,
		Args:  []any{req.Limit, req.Offset},
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	meta, err := rowsToMeta(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	data, err := rowsToData(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.MetricsViewToplistResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

// MetricsViewTimeSeries implements RuntimeService
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *api.MetricsViewTimeSeriesRequest) (*api.MetricsViewTimeSeriesResponse, error) {
	// NOTE: Mock implementation

	sql, args, err := buildMetricsTimeSeriesSQL(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error building query: %s", err.Error())
	}

	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: sql,
		Args:  args,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	meta, err := rowsToMeta(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	data, err := rowsToData(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return resp, nil
}

func (s *Server) query(ctx context.Context, instanceID string, stmt *drivers.Statement) (*sqlx.Rows, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, fmt.Errorf("invalid instance_id")
	}

	instance := s.runtime.InstanceFromID(id)
	if err = instance.Load(); err != nil {
		return nil, err
	}

	return instance.Query(ctx, stmt)
}

func rowsToMeta(rows *sqlx.Rows) ([]*api.SchemaColumn, error) {
	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	meta := make([]*api.SchemaColumn, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		meta[i] = &api.SchemaColumn{
			Name:     ct.Name(),
			Type:     ct.DatabaseTypeName(),
			Nullable: nullable,
		}
	}

	return meta, nil
}

func rowsToData(rows *sqlx.Rows) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		// Note: structpb only supports JSON types (e.g. not time.Time)
		// For now, we're doing a JSON round-trip for convenience

		json, err := json.Marshal(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct := &structpb.Struct{}
		err = rowStruct.UnmarshalJSON(json)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func buildMetricsTimeSeriesSQL(req *api.MetricsViewTimeSeriesRequest) (string, []any, error) {
	timeField := "time"
	timeCol := fmt.Sprintf("DATE_TRUNC(%s, %s) AS %s", timeField, req.TimeGranularity, timeField)
	selectCols := append([]string{timeCol}, req.MeasureNames...)

	whereClause := "time >= ? AND time < ? "
	args := []any{time.UnixMilli(req.TimeStart), time.UnixMilli(req.TimeEnd)}

	if req.Filter != nil && req.Filter.Include != nil {
		clause, clauseArgs, err := buildFilterClause(req.Filter.Include, "IN")
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	if req.Filter != nil && req.Filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClause(req.Filter.Exclude, "NOT IN")
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s LIMIT 1000", strings.Join(selectCols, ", "), req.MetricsViewName, whereClause, timeField)
	return sql, args, nil
}

func buildFilterClause(conds []*api.MetricsViewFilterCond, operator string) (string, []any, error) {
	args := []any{}
	clause := ""
	for _, cond := range conds {
		questionMarks := strings.Join(repeatString("?", len(cond.Values)), ",")
		clause += fmt.Sprintf("AND %s %s (%s) ", cond.Name, operator, questionMarks)
		for _, val := range cond.Values {
			arg, err := protobufValueToAny(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %s", err.Error())
			}
			args = append(args, arg)
		}
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
