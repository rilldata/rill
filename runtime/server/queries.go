package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

// Query implements RuntimeService
func (s *Server) Query(ctx context.Context, req *api.QueryRequest) (*api.QueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
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
		// TODO: Return a meta object for dry-run queries
		// NOTE: Currently, instance.Query return nil rows for succesful dry-run queries
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
		// For now, we're doing a JSON round-trip for convenience (not great for performance)

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
