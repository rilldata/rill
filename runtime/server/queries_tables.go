package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

const _tableHeadDefaultLimit = 25

// Table level profiling APIs.
func (s *Server) TableCardinality(ctx context.Context, req *runtimev1.TableCardinalityRequest) (*runtimev1.TableCardinalityResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.table", req.TableName),
		attribute.Int("args.priority", int(req.Priority)),
	)

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
	observability.SetRequestAttributes(ctx,
		attribute.String("instance_id", req.InstanceId),
		attribute.String("table", req.TableName),
		attribute.Int("args.priority", int(req.Priority)),
	)

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

	return &runtimev1.TableColumnsResponse{
		ProfileColumns: q.Result,
	}, nil
}

func (s *Server) TableRows(ctx context.Context, req *runtimev1.TableRowsRequest) (*runtimev1.TableRowsResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.table", req.TableName),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int("args.priority", int(req.Priority)),
	)

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

	return &runtimev1.TableRowsResponse{
		Data: q.Result,
	}, nil
}
