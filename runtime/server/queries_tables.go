package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
)

const _tableHeadDefaultLimit = 25

// Table level profiling APIs.
func (s *Server) GetTableCardinality(ctx context.Context, req *runtimev1.GetTableCardinalityRequest) (*runtimev1.GetTableCardinalityResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

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
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

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
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	limit := int(req.Limit)
	if limit == 0 {
		limit = _tableHeadDefaultLimit
	}

	q := &queries.TableHead{
		TableName: req.TableName,
		Limit:     limit,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GetTableRowsResponse{
		Data: q.Result,
	}, nil
}
