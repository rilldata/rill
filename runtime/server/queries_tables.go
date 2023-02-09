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
func (s *Server) TableCardinality(ctx context.Context, req *runtimev1.TableCardinalityRequest) (*runtimev1.TableCardinalityResponse, error) {
	q := &queries.TableCardinality{
		TableName: req.TableName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return &runtimev1.TableCardinalityResponse{
		Cardinality: q.Result,
	}, nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

func (s *Server) TableColumns(ctx context.Context, req *runtimev1.TableColumnsRequest) (*runtimev1.TableColumnsResponse, error) {
	q := &queries.TableColumns{
		TableName: req.TableName,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.TableColumnsResponse{
		ProfileColumns: q.Result,
	}, nil
}

func (s *Server) TableRows(ctx context.Context, req *runtimev1.TableRowsRequest) (*runtimev1.TableRowsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    fmt.Sprintf("select * from %q limit %d", req.TableName, req.Limit),
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

	return &runtimev1.TableRowsResponse{
		Data: data,
	}, nil
}
