package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

// Table level profiling APIs.
func (s *Server) GetTableCardinality(ctx context.Context, req *runtimev1.GetTableCardinalityRequest) (*runtimev1.GetTableCardinalityResponse, error) {
	q := &queries.TableCardinality{
		TableName: req.TableName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return &runtimev1.GetTableCardinalityResponse{
		Cardinality: q.Result,
	}, nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

func (s *Server) ProfileColumns(ctx context.Context, req *runtimev1.ProfileColumnsRequest) (*runtimev1.ProfileColumnsResponse, error) {
	q := &queries.TableColumns{
		TableName: req.TableName,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.ProfileColumnsResponse{
		ProfileColumns: q.Result,
	}, nil
}

func (s *Server) GetTableRows(ctx context.Context, req *runtimev1.GetTableRowsRequest) (*runtimev1.GetTableRowsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    fmt.Sprintf("select * from %s limit %d", req.TableName, req.Limit),
		Priority: int(req.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*structpb.Struct
	if data, err = rowsToData(rows); err != nil {
		return nil, err
	}

	return &runtimev1.GetTableRowsResponse{
		Data: data,
	}, nil
}
