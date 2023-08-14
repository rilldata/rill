package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

const _tableHeadDefaultLimit = 25

// Table level profiling APIs.
func (s *Server) TableCardinality(ctx context.Context, req *connect.Request[runtimev1.TableCardinalityRequest]) (*connect.Response[runtimev1.TableCardinalityResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.table", req.Msg.TableName),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.TableCardinality{
		TableName: req.Msg.TableName,
	}
	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&runtimev1.TableCardinalityResponse{
		Cardinality: q.Result,
	}), nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

func (s *Server) TableColumns(ctx context.Context, req *connect.Request[runtimev1.TableColumnsRequest]) (*connect.Response[runtimev1.TableColumnsResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.table", req.Msg.TableName),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.TableColumns{
		TableName: req.Msg.TableName,
	}

	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.TableColumnsResponse{
		ProfileColumns: q.Result,
	}), nil
}

func (s *Server) TableRows(ctx context.Context, req *connect.Request[runtimev1.TableRowsRequest]) (*connect.Response[runtimev1.TableRowsResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.table", req.Msg.TableName),
		attribute.Int("args.limit", int(req.Msg.Limit)),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	limit := int(req.Msg.Limit)
	if limit == 0 {
		limit = _tableHeadDefaultLimit
	}

	q := &queries.TableHead{
		TableName: req.Msg.TableName,
		Limit:     limit,
	}

	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.TableRowsResponse{
		Data: q.Result,
	}), nil
}
