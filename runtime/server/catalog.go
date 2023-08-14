package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListCatalogEntries implements RuntimeService.
func (s *Server) ListCatalogEntries(ctx context.Context, req *connect.Request[runtimev1.ListCatalogEntriesRequest]) (*connect.Response[runtimev1.ListCatalogEntriesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.type", req.Msg.Type.String()),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	entries, err := s.runtime.ListCatalogEntries(ctx, req.Msg.InstanceId, pbToObjectType(req.Msg.Type))
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	pbs := make([]*runtimev1.CatalogEntry, len(entries))
	for i, obj := range entries {
		var err error
		pbs[i], err = catalogObjectToPB(obj)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return connect.NewResponse(&runtimev1.ListCatalogEntriesResponse{Entries: pbs}), nil
}

// GetCatalogEntry implements RuntimeService.
func (s *Server) GetCatalogEntry(ctx context.Context, req *connect.Request[runtimev1.GetCatalogEntryRequest]) (*connect.Response[runtimev1.GetCatalogEntryResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.name", req.Msg.Name),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	entry, err := s.runtime.GetCatalogEntry(ctx, req.Msg.InstanceId, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	pb, err := catalogObjectToPB(entry)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return connect.NewResponse(&runtimev1.GetCatalogEntryResponse{Entry: pb}), nil
}

// Reconcile implements RuntimeService.
func (s *Server) Reconcile(ctx context.Context, req *connect.Request[runtimev1.ReconcileRequest]) (*connect.Response[runtimev1.ReconcileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	res, err := s.runtime.Reconcile(ctx, req.Msg.InstanceId, req.Msg.ChangedPaths, req.Msg.ForcedPaths, req.Msg.Dry, req.Msg.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.ReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}), nil
}

// PutFileAndReconcile implements RuntimeService.
func (s *Server) PutFileAndReconcile(ctx context.Context, req *connect.Request[runtimev1.PutFileAndReconcileRequest]) (*connect.Response[runtimev1.PutFileAndReconcileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.Msg.InstanceId, auth.EditRepo) || !claims.CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.Msg.InstanceId, req.Msg.Path, strings.NewReader(req.Msg.Blob), req.Msg.Create, req.Msg.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Msg.Path}
	res, err := s.runtime.Reconcile(ctx, req.Msg.InstanceId, changedPaths, nil, req.Msg.Dry, req.Msg.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.PutFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}), nil
}

// RenameFileAndReconcile implements RuntimeService.
func (s *Server) RenameFileAndReconcile(ctx context.Context, req *connect.Request[runtimev1.RenameFileAndReconcileRequest]) (*connect.Response[runtimev1.RenameFileAndReconcileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.Msg.InstanceId, auth.EditRepo) || !claims.CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.Msg.InstanceId, req.Msg.FromPath, req.Msg.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Msg.FromPath, req.Msg.ToPath}
	res, err := s.runtime.Reconcile(ctx, req.Msg.InstanceId, changedPaths, nil, req.Msg.Dry, req.Msg.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.RenameFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}), nil
}

// DeleteFileAndReconcile implements RuntimeService.
func (s *Server) DeleteFileAndReconcile(ctx context.Context, req *connect.Request[runtimev1.DeleteFileAndReconcileRequest]) (*connect.Response[runtimev1.DeleteFileAndReconcileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.Msg.InstanceId, auth.EditRepo) || !claims.CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.Msg.InstanceId, req.Msg.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Msg.Path}
	res, err := s.runtime.Reconcile(ctx, req.Msg.InstanceId, changedPaths, nil, req.Msg.Dry, req.Msg.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.DeleteFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}), nil
}

// RefreshAndReconcile implements RuntimeService.
func (s *Server) RefreshAndReconcile(ctx context.Context, req *connect.Request[runtimev1.RefreshAndReconcileRequest]) (*connect.Response[runtimev1.RefreshAndReconcileResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	changedPaths := []string{req.Msg.Path}
	res, err := s.runtime.Reconcile(ctx, req.Msg.InstanceId, changedPaths, changedPaths, req.Msg.Dry, req.Msg.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.RefreshAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}), nil
}

// TriggerRefresh implements RuntimeService.
func (s *Server) TriggerRefresh(ctx context.Context, req *connect.Request[runtimev1.TriggerRefreshRequest]) (*connect.Response[runtimev1.TriggerRefreshResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RefreshSource(ctx, req.Msg.InstanceId, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return connect.NewResponse(&runtimev1.TriggerRefreshResponse{}), nil
}

// TriggerSync implements RuntimeService.
func (s *Server) TriggerSync(ctx context.Context, req *connect.Request[runtimev1.TriggerSyncRequest]) (*connect.Response[runtimev1.TriggerSyncResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	err := s.runtime.SyncExistingTables(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Done
	// TODO: This should return stats about synced tables. However, it will be refactored into reconcile, so no need to fix this now.
	return connect.NewResponse(&runtimev1.TriggerSyncResponse{}), nil
}

func pbToObjectType(in runtimev1.ObjectType) drivers.ObjectType {
	switch in {
	case runtimev1.ObjectType_OBJECT_TYPE_UNSPECIFIED:
		return drivers.ObjectTypeUnspecified
	case runtimev1.ObjectType_OBJECT_TYPE_TABLE:
		return drivers.ObjectTypeTable
	case runtimev1.ObjectType_OBJECT_TYPE_SOURCE:
		return drivers.ObjectTypeSource
	case runtimev1.ObjectType_OBJECT_TYPE_MODEL:
		return drivers.ObjectTypeModel
	case runtimev1.ObjectType_OBJECT_TYPE_METRICS_VIEW:
		return drivers.ObjectTypeMetricsView
	}
	panic(fmt.Errorf("unhandled object type %s", in))
}

func catalogObjectToPB(obj *drivers.CatalogEntry) (*runtimev1.CatalogEntry, error) {
	catalog := &runtimev1.CatalogEntry{
		Name:        obj.Name,
		Path:        obj.Path,
		Embedded:    obj.Embedded,
		Parents:     obj.Parents,
		Children:    obj.Children,
		CreatedOn:   timestamppb.New(obj.CreatedOn),
		UpdatedOn:   timestamppb.New(obj.UpdatedOn),
		RefreshedOn: timestamppb.New(obj.RefreshedOn),
	}

	switch obj.Type {
	case drivers.ObjectTypeTable:
		catalog.Object = &runtimev1.CatalogEntry_Table{
			Table: obj.GetTable(),
		}
	case drivers.ObjectTypeSource:
		catalog.Object = &runtimev1.CatalogEntry_Source{
			Source: obj.GetSource(),
		}
	case drivers.ObjectTypeModel:
		catalog.Object = &runtimev1.CatalogEntry_Model{
			Model: obj.GetModel(),
		}
	case drivers.ObjectTypeMetricsView:
		catalog.Object = &runtimev1.CatalogEntry_MetricsView{
			MetricsView: obj.GetMetricsView(),
		}
	default:
		panic("not implemented")
	}

	return catalog, nil
}
