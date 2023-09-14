package server

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// GetLogs implements runtimev1.RuntimeServiceServer
func (s *Server) GetLogs(ctx context.Context, req *runtimev1.GetLogsRequest) (*runtimev1.GetLogsResponse, error) {
	panic("not implemented")
}

// WatchLogs implements runtimev1.RuntimeServiceServer
func (s *Server) WatchLogs(req *runtimev1.WatchLogsRequest, srv runtimev1.RuntimeService_WatchLogsServer) error {
	panic("not implemented")
}

// ListResources implements runtimev1.RuntimeServiceServer
func (s *Server) ListResources(ctx context.Context, req *runtimev1.ListResourcesRequest) (*runtimev1.ListResourcesResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rs, err := ctrl.List(ctx, req.Kind, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	for i := 0; i < len(rs); i++ {
		r := rs[i]
		r, access, err := s.applySecurityPolicy(ctx, req.InstanceId, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if !access {
			// Remove from the slice
			rs[i] = rs[len(rs)-1]
			rs[len(rs)-1] = nil
			rs = rs[:len(rs)-1]
			continue
		}
		rs[i] = r
	}

	return &runtimev1.ListResourcesResponse{Resources: rs}, nil
}

// WatchResources implements runtimev1.RuntimeServiceServer
func (s *Server) WatchResources(req *runtimev1.WatchResourcesRequest, ss runtimev1.RuntimeService_WatchResourcesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
	)

	if !auth.GetClaims(ss.Context()).CanInstance(req.InstanceId, auth.ReadObjects) {
		return ErrForbidden
	}

	ctrl, err := s.runtime.Controller(req.InstanceId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Replay {
		rs, err := ctrl.List(ss.Context(), req.Kind, false)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		for _, r := range rs {
			r, access, err := s.applySecurityPolicy(ss.Context(), req.InstanceId, r)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if !access {
				continue
			}

			err = ss.Send(&runtimev1.WatchResourcesResponse{
				Event:    runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE,
				Resource: r,
			})
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	return ctrl.Subscribe(ss.Context(), func(e runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource) {
		r, access, err := s.applySecurityPolicy(ss.Context(), req.InstanceId, r)
		if err != nil {
			s.logger.Info("failed to apply security policy", zap.String("name", n.Name), zap.Error(err))
			return
		}
		if !access {
			return
		}

		err = ss.Send(&runtimev1.WatchResourcesResponse{
			Event:    e,
			Name:     n,
			Resource: r,
		})
		if err != nil {
			s.logger.Info("failed to send resource event", zap.Error(err))
		}
	})
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name.kind", req.Name.Kind),
		attribute.String("args.name.name", req.Name.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, err := ctrl.Get(ctx, req.Name, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
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

	return &runtimev1.GetResourceResponse{Resource: r}, nil
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	panic("not implemented")
}

// applySecurityPolicy applies relevant security policies to the resource.
// The input resource will not be modified in-place (so no need to set clone=true when obtaining it from the catalog).
func (s *Server) applySecurityPolicy(ctx context.Context, instID string, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	mv := r.GetMetricsView()
	if mv == nil || mv.State.ValidSpec == nil || mv.State.ValidSpec.Security == nil {
		// Allow if it's not a metrics view or it doesn't have a valid security policy.
		return r, true, nil
	}

	security, err := s.runtime.ResolveMetricsViewSecurityV2(auth.GetClaims(ctx).Attributes(), instID, mv.State.ValidSpec, r.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return nil, false, err
	}

	if !security.Access {
		return nil, false, err
	}

	mv, changed := s.applySecurityPolicyIncludesAndExcludes(mv, security)
	if changed {
		// We mustn't modify the resource in-place
		r = &runtimev1.Resource{
			Meta:     r.Meta,
			Resource: &runtimev1.Resource_MetricsView{MetricsView: mv},
		}
	}

	return r, true, nil
}

// applySecurityPolicyIncludesAndExcludes rewrites a metrics view based on the include/exclude conditions of a security policy.
func (s *Server) applySecurityPolicyIncludesAndExcludes(mv *runtimev1.MetricsViewV2, policy *runtime.ResolvedMetricsViewSecurity) (*runtimev1.MetricsViewV2, bool) {
	if policy == nil || (len(policy.Include) == 0 && len(policy.Exclude) == 0) {
		return mv, false
	}

	mv = proto.Clone(mv).(*runtimev1.MetricsViewV2)

	if len(policy.Include) > 0 {
		allowed := make(map[string]bool)
		for _, include := range policy.Include {
			allowed[include] = true
		}

		dims := make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
		for _, dim := range mv.Spec.Dimensions {
			if allowed[dim.Name] {
				dims = append(dims, dim)
			}
		}
		mv.Spec.Dimensions = dims

		ms := make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
		for _, m := range mv.Spec.Measures {
			if allowed[m.Name] {
				ms = append(ms, m)
			}
		}
		mv.Spec.Measures = ms

		if mv.State.ValidSpec != nil {
			dims = make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
			for _, dim := range mv.State.ValidSpec.Dimensions {
				if allowed[dim.Name] {
					dims = append(dims, dim)
				}
			}
			mv.State.ValidSpec.Dimensions = dims

			ms = make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
			for _, m := range mv.State.ValidSpec.Measures {
				if allowed[m.Name] {
					ms = append(ms, m)
				}
			}
			mv.State.ValidSpec.Measures = ms
		}
	}

	if len(policy.Exclude) > 0 {
		restricted := make(map[string]bool)
		for _, exclude := range policy.Exclude {
			restricted[exclude] = true
		}

		dims := make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
		for _, dim := range mv.Spec.Dimensions {
			if !restricted[dim.Name] {
				dims = append(dims, dim)
			}
		}
		mv.Spec.Dimensions = dims

		ms := make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
		for _, m := range mv.Spec.Measures {
			if !restricted[m.Name] {
				ms = append(ms, m)
			}
		}
		mv.Spec.Measures = ms

		if mv.State.ValidSpec != nil {
			dims = make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
			for _, dim := range mv.State.ValidSpec.Dimensions {
				if !restricted[dim.Name] {
					dims = append(dims, dim)
				}
			}
			mv.State.ValidSpec.Dimensions = dims

			ms = make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
			for _, m := range mv.State.ValidSpec.Measures {
				if !restricted[m.Name] {
					ms = append(ms, m)
				}
			}
			mv.State.ValidSpec.Measures = ms
		}
	}

	return mv, true
}
