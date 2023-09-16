package server

import (
	"context"
	"errors"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NOTE: Lots of ugly code in here. It's just temporary backwards compatibility.

// ListCatalogEntries implements RuntimeService.
func (s *Server) ListCatalogEntries(ctx context.Context, req *runtimev1.ListCatalogEntriesRequest) (*runtimev1.ListCatalogEntriesResponse, error) {
	var kind string
	switch req.Type {
	case runtimev1.ObjectType_OBJECT_TYPE_UNSPECIFIED:
		kind = ""
	case runtimev1.ObjectType_OBJECT_TYPE_SOURCE:
		kind = runtime.ResourceKindSource
	case runtimev1.ObjectType_OBJECT_TYPE_MODEL:
		kind = runtime.ResourceKindModel
	case runtimev1.ObjectType_OBJECT_TYPE_METRICS_VIEW:
		kind = runtime.ResourceKindMetricsView
	default:
		return nil, errors.New("unsupported object type")
	}

	res, err := s.ListResources(ctx, &runtimev1.ListResourcesRequest{
		InstanceId: req.InstanceId,
		Kind:       kind,
	})
	if err != nil {
		return nil, err
	}

	var pbs []*runtimev1.CatalogEntry
	for _, r := range res.Resources {
		pb, err := s.resourceToEntry(ctx, req.InstanceId, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if pb != nil {
			pbs = append(pbs, pb)
		}
	}

	return &runtimev1.ListCatalogEntriesResponse{Entries: pbs}, nil
}

// GetCatalogEntry implements RuntimeService.
func (s *Server) GetCatalogEntry(ctx context.Context, req *runtimev1.GetCatalogEntryRequest) (*runtimev1.GetCatalogEntryResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name", req.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	kinds := []string{
		runtime.ResourceKindSource,
		runtime.ResourceKindModel,
		runtime.ResourceKindMetricsView,
	}

	for _, k := range kinds {
		r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: k, Name: req.Name}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		r, access, err := s.applySecurityPolicy(ctx, req.InstanceId, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if !access {
			return nil, status.Error(codes.NotFound, "resource not found")
		}

		pb, err := s.resourceToEntry(ctx, req.InstanceId, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &runtimev1.GetCatalogEntryResponse{Entry: pb}, nil
	}

	return nil, status.Error(codes.NotFound, "resource not found")
}

// Reconcile implements RuntimeService.
func (s *Server) Reconcile(ctx context.Context, req *runtimev1.ReconcileRequest) (*runtimev1.ReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	res, err := s.runtime.Reconcile(ctx, req.InstanceId, req.ChangedPaths, req.ForcedPaths, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.ReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// PutFileAndReconcile implements RuntimeService.
func (s *Server) PutFileAndReconcile(ctx context.Context, req *runtimev1.PutFileAndReconcileRequest) (*runtimev1.PutFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.PutFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// RenameFileAndReconcile implements RuntimeService.
func (s *Server) RenameFileAndReconcile(ctx context.Context, req *runtimev1.RenameFileAndReconcileRequest) (*runtimev1.RenameFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.FromPath, req.ToPath}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RenameFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// DeleteFileAndReconcile implements RuntimeService.
func (s *Server) DeleteFileAndReconcile(ctx context.Context, req *runtimev1.DeleteFileAndReconcileRequest) (*runtimev1.DeleteFileAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, nil, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteFileAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// RefreshAndReconcile implements RuntimeService.
func (s *Server) RefreshAndReconcile(ctx context.Context, req *runtimev1.RefreshAndReconcileRequest) (*runtimev1.RefreshAndReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	changedPaths := []string{req.Path}
	res, err := s.runtime.Reconcile(ctx, req.InstanceId, changedPaths, changedPaths, req.Dry, req.Strict)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.RefreshAndReconcileResponse{
		Errors:        res.Errors,
		AffectedPaths: res.AffectedPaths,
	}, nil
}

// TriggerRefresh implements RuntimeService.
func (s *Server) TriggerRefresh(ctx context.Context, req *runtimev1.TriggerRefreshRequest) (*runtimev1.TriggerRefreshResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	err := s.runtime.RefreshSource(ctx, req.InstanceId, req.Name)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &runtimev1.TriggerRefreshResponse{}, nil
}

func (s *Server) resourceToEntry(ctx context.Context, instanceID string, r *runtimev1.Resource) (*runtimev1.CatalogEntry, error) {
	var path string
	if len(r.Meta.FilePaths) != 0 {
		path = r.Meta.FilePaths[0]
	}

	var parents []string
	for _, ref := range r.Meta.Refs {
		parents = append(parents, ref.Name)
	}

	res := &runtimev1.CatalogEntry{
		Name:        r.Meta.Name.Name,
		Path:        path,
		Parents:     parents,
		CreatedOn:   r.Meta.CreatedOn,
		UpdatedOn:   r.Meta.SpecUpdatedOn,
		RefreshedOn: r.Meta.StateUpdatedOn,
	}

	olap, release, err := s.runtime.OLAP(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	switch r.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		src := r.GetSource()
		if src.State.Table == "" {
			return nil, nil
		}
		t, err := olap.InformationSchema().Lookup(ctx, src.State.Table)
		if err != nil {
			return nil, err
		}
		res.Object = &runtimev1.CatalogEntry_Source{
			Source: &runtimev1.Source{
				Name:           r.Meta.Name.Name,
				Connector:      src.Spec.SourceConnector,
				Properties:     src.Spec.Properties,
				Schema:         t.Schema,
				TimeoutSeconds: int32(src.Spec.TimeoutSeconds),
			},
		}
	case runtime.ResourceKindModel:
		mdl := r.GetModel()
		if mdl.State.Table == "" {
			return nil, nil
		}
		t, err := olap.InformationSchema().Lookup(ctx, mdl.State.Table)
		if err != nil {
			return nil, err
		}
		materialize := false
		if mdl.Spec.Materialize != nil {
			materialize = *mdl.Spec.Materialize
		}
		res.Object = &runtimev1.CatalogEntry_Model{
			Model: &runtimev1.Model{
				Name:        r.Meta.Name.Name,
				Sql:         mdl.Spec.Sql,
				Dialect:     runtimev1.Model_DIALECT_DUCKDB,
				Schema:      t.Schema,
				Materialize: materialize,
			},
		}
	case runtime.ResourceKindMetricsView:
		mv := r.GetMetricsView()
		if mv.State.ValidSpec == nil {
			return nil, nil
		}
		var dims []*runtimev1.MetricsView_Dimension
		for _, d := range mv.State.ValidSpec.Dimensions {
			dims = append(dims, &runtimev1.MetricsView_Dimension{
				Name:        d.Name,
				Label:       d.Label,
				Description: d.Description,
				Column:      d.Column,
			})
		}
		var ms []*runtimev1.MetricsView_Measure
		for _, m := range mv.State.ValidSpec.Measures {
			ms = append(ms, &runtimev1.MetricsView_Measure{
				Name:                m.Name,
				Label:               m.Label,
				Expression:          m.Expression,
				Description:         m.Description,
				Format:              m.Format,
				ValidPercentOfTotal: m.ValidPercentOfTotal,
			})
		}
		var security *runtimev1.MetricsView_Security
		if mv.State.ValidSpec.Security != nil {
			var includes []*runtimev1.MetricsView_Security_FieldCondition
			for _, inc := range mv.State.ValidSpec.Security.Include {
				includes = append(includes, &runtimev1.MetricsView_Security_FieldCondition{
					Condition: inc.Condition,
					Names:     inc.Names,
				})
			}
			var excludes []*runtimev1.MetricsView_Security_FieldCondition
			for _, exc := range mv.State.ValidSpec.Security.Exclude {
				excludes = append(excludes, &runtimev1.MetricsView_Security_FieldCondition{
					Condition: exc.Condition,
					Names:     exc.Names,
				})
			}
			security = &runtimev1.MetricsView_Security{
				Access:    mv.State.ValidSpec.Security.Access,
				RowFilter: mv.State.ValidSpec.Security.RowFilter,
				Include:   includes,
				Exclude:   excludes,
			}
		}
		res.Object = &runtimev1.CatalogEntry_MetricsView{
			MetricsView: &runtimev1.MetricsView{
				Name:               r.Meta.Name.Name,
				TimeDimension:      mv.State.ValidSpec.TimeDimension,
				Dimensions:         dims,
				Measures:           ms,
				Label:              mv.State.ValidSpec.Title,
				Description:        mv.State.ValidSpec.Description,
				SmallestTimeGrain:  mv.State.ValidSpec.SmallestTimeGrain,
				DefaultTimeRange:   mv.State.ValidSpec.DefaultTimeRange,
				AvailableTimeZones: mv.State.ValidSpec.AvailableTimeZones,
				Security:           security,
			},
		}
	default:
		// Don't need to support other types here
		return nil, nil
	}

	return res, nil
}
