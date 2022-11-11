package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

// Table level profiling APIs
func (s *Server) RenameDatabaseObject(ctx context.Context, req *api.RenameDatabaseObjectRequest) (*api.RenameDatabaseObjectResponse, error) {
	return &api.RenameDatabaseObjectResponse{}, nil
}

func (s *Server) TableCardinality(ctx context.Context, req *api.CardinalityRequest) (*api.CardinalityResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: "select count(*) from " + quoteName(req.TableName),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	return &api.CardinalityResponse{
		Cardinality: count,
	}, nil
}

func (s *Server) ProfileColumns(ctx context.Context, req *api.ProfileColumnsRequest) (*api.ProfileColumnsResponse, error) {
	return &api.ProfileColumnsResponse{
		ProfileColumn: []*api.ProfileColumn{},
	}, nil
}

func (s *Server) TableRows(ctx context.Context, req *api.RowsRequest) (*api.RowsResponse, error) {
	rows := make([]*structpb.Struct, 1)
	rows[0] = &structpb.Struct{}
	return &api.RowsResponse{
		Data: rows,
	}, nil
}
