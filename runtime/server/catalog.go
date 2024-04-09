package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/maps"
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

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
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
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	since := time.Now()

	err = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.InvalidArgument, ctx.Err().Error())
	case <-time.After(500 * time.Millisecond):
		// Give it 0.5s to create the derived resources
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return s.controllerToLegacyReconcileStatus(ctx, ctrl, since)
}

// PutFileAndReconcile implements RuntimeService.
func (s *Server) PutFileAndReconcile(ctx context.Context, req *runtimev1.PutFileAndReconcileRequest) (*runtimev1.PutFileAndReconcileResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	since := time.Now()

	err = s.runtime.PutFile(ctx, req.InstanceId, req.Path, strings.NewReader(req.Blob), req.Create, req.CreateOnly)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.InvalidArgument, ctx.Err().Error())
	case <-time.After(500 * time.Millisecond):
		// Give the watcher 0.5s to pick up the updated file
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := s.controllerToLegacyReconcileStatus(ctx, ctrl, since)
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
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	since := time.Now()

	err = s.runtime.RenameFile(ctx, req.InstanceId, req.FromPath, req.ToPath)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.InvalidArgument, ctx.Err().Error())
	case <-time.After(500 * time.Millisecond):
		// Give the watcher 0.5s to pick up the updated file
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := s.controllerToLegacyReconcileStatus(ctx, ctrl, since)
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
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditRepo) || !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	since := time.Now()

	err = s.runtime.DeleteFile(ctx, req.InstanceId, req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := s.controllerToLegacyReconcileStatus(ctx, ctrl, since)
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
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rs, err := ctrl.List(ctx, runtime.ResourceKindSource, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var names []*runtimev1.ResourceName
	for _, r := range rs {
		for _, p := range r.Meta.FilePaths {
			if p == req.Path {
				names = append(names, r.Meta.Name)
				break
			}
		}
	}

	since := time.Now()

	trgName := &runtimev1.ResourceName{
		Kind: runtime.ResourceKindRefreshTrigger,
		Name: fmt.Sprintf("trigger_adhoc_%s", time.Now().Format("200601021504059999")),
	}

	err = ctrl.Create(ctx, trgName, nil, nil, nil, true, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					OnlyNames: names,
				},
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := s.controllerToLegacyReconcileStatus(ctx, ctrl, since)
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
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	trgName := &runtimev1.ResourceName{
		Kind: runtime.ResourceKindRefreshTrigger,
		Name: fmt.Sprintf("trigger_adhoc_%s", time.Now().Format("200601021504059999")),
	}

	err = ctrl.Create(ctx, trgName, nil, nil, nil, true, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					OnlyNames: []*runtimev1.ResourceName{{Kind: runtime.ResourceKindSource, Name: req.Name}},
				},
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = ctrl.WaitUntilIdle(ctx, true)
	if ctx.Err() != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	olap, release, err := s.runtime.OLAP(ctx, instanceID, "")
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
		t, err := olap.InformationSchema().Lookup(ctx, "", "", src.State.Table)
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
		t, err := olap.InformationSchema().Lookup(ctx, "", "", mdl.State.Table)
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
				Column:      d.Expression,
			})
		}
		var ms []*runtimev1.MetricsView_Measure
		for _, m := range mv.State.ValidSpec.Measures {
			ms = append(ms, &runtimev1.MetricsView_Measure{
				Name:                m.Name,
				Label:               m.Label,
				Expression:          m.Expression,
				Description:         m.Description,
				Format:              m.FormatPreset,
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
				Model:              mv.State.ValidSpec.Table,
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

func (s *Server) controllerToLegacyReconcileStatus(ctx context.Context, ctrl *runtime.Controller, since time.Time) (*runtimev1.ReconcileResponse, error) {
	rs, err := ctrl.List(ctx, "", false)
	if err != nil {
		return nil, err
	}

	affectedPaths := make(map[string]bool)
	var errs []*runtimev1.ReconcileError

	for _, r := range rs {
		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			pp := r.GetProjectParser()
			for _, perr := range pp.State.ParseErrors {
				errs = append(errs, &runtimev1.ReconcileError{
					Code:     runtimev1.ReconcileError_CODE_SYNTAX,
					Message:  perr.Message,
					FilePath: perr.FilePath,
				})
			}
			continue
		}

		switch r.Meta.Name.Kind {
		case runtime.ResourceKindSource, runtime.ResourceKindModel, runtime.ResourceKindMetricsView:
		default:
			continue
		}

		if r.Meta.SpecUpdatedOn.AsTime().Before(since) && r.Meta.StateUpdatedOn.AsTime().Before(since) {
			continue
		}

		if r.Meta.ReconcileError != "" {
			for _, p := range r.Meta.FilePaths {
				affectedPaths[p] = true
			}
			errs = append(errs, &runtimev1.ReconcileError{
				Code:    runtimev1.ReconcileError_CODE_UNSPECIFIED,
				Message: r.Meta.ReconcileError,
			})
		}
	}

	return &runtimev1.ReconcileResponse{
		AffectedPaths: maps.Keys(affectedPaths),
		Errors:        errs,
	}, nil
}
