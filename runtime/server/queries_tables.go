package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

const _tableHeadDefaultLimit = 25

// Table level profiling APIs.
func (s *Server) TableCardinality(ctx context.Context, req *runtimev1.TableCardinalityRequest) (*runtimev1.TableCardinalityResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.TableCardinality{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
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
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.TableColumns{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) TableRows(ctx context.Context, req *runtimev1.TableRowsRequest) (*runtimev1.TableRowsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.TableName),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadProfiling) {
		return nil, ErrForbidden
	}

	limit := int(req.Limit)
	if limit == 0 {
		limit = _tableHeadDefaultLimit
	}

	q := &queries.TableHead{
		Connector:      req.Connector,
		Database:       req.Database,
		DatabaseSchema: req.DatabaseSchema,
		TableName:      req.TableName,
		Limit:          limit,
	}

	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.TableRowsResponse{
		Data: q.Result,
	}, nil
}
