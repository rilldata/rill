package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rs, err := ctrl.List(ctx, req.Kind, req.Path, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		an := a.Meta.Name
		bn := b.Meta.Name
		if an.Kind < bn.Kind {
			return -1
		}
		if an.Kind > bn.Kind {
			return 1
		}
		return strings.Compare(an.Name, bn.Name)
	})

	i := 0
	for i < len(rs) {
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
		i++
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

	ctrl, err := s.runtime.Controller(ss.Context(), req.InstanceId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Replay {
		rs, err := ctrl.List(ss.Context(), req.Kind, "", false)
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
		if r != nil { // r is nil for deletion events
			var access bool
			var err error
			r, access, err = s.applySecurityPolicy(ss.Context(), req.InstanceId, r)
			if err != nil {
				s.logger.Info("failed to apply security policy", zap.String("name", n.Name), zap.Error(err))
				return
			}
			if !access {
				return
			}
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

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
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

	var kind string
	r := &runtimev1.Resource{}

	switch trg := req.Trigger.(type) {
	case *runtimev1.CreateTriggerRequest_PullTriggerSpec:
		kind = runtime.ResourceKindPullTrigger
		r.Resource = &runtimev1.Resource_PullTrigger{PullTrigger: &runtimev1.PullTrigger{Spec: trg.PullTriggerSpec}}
	case *runtimev1.CreateTriggerRequest_RefreshTriggerSpec:
		kind = runtime.ResourceKindRefreshTrigger
		r.Resource = &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{Spec: trg.RefreshTriggerSpec}}
	}

	n := &runtimev1.ResourceName{
		Kind: kind,
		Name: fmt.Sprintf("trigger_adhoc_%s", time.Now().Format("200601021504059999")),
	}

	err = ctrl.Create(ctx, n, nil, nil, nil, true, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to create trigger: %w", err).Error())
	}

	return &runtimev1.CreateTriggerResponse{}, nil
}

// applySecurityPolicy applies relevant security policies to the resource.
// The input resource will not be modified in-place (so no need to set clone=true when obtaining it from the catalog).
func (s *Server) applySecurityPolicy(ctx context.Context, instID string, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	switch r.Resource.(type) {
	case *runtimev1.Resource_MetricsView:
		return s.applySecurityPolicyMetricsView(ctx, instID, r)
	case *runtimev1.Resource_Report:
		return s.applySecurityPolicyReport(ctx, r)
	case *runtimev1.Resource_Alert:
		return s.applySecurityPolicyAlert(ctx, r)
	default:
		return r, true, nil
	}
}

// applySecurityPolicyMetricsView applies relevant security policies to a metrics view.
func (s *Server) applySecurityPolicyMetricsView(ctx context.Context, instID string, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	ctx, span := tracer.Start(ctx, "applySecurityPolicyMetricsView", trace.WithAttributes(attribute.String("instance_id", instID), attribute.String("kind", r.Meta.Name.Kind), attribute.String("name", r.Meta.Name.Name)))
	defer span.End()

	mv := r.GetMetricsView()
	if mv.State.ValidSpec == nil || mv.State.ValidSpec.Security == nil {
		// Allow if it doesn't have a valid security policy
		return r, true, nil
	}

	security, err := s.runtime.ResolveMetricsViewSecurity(auth.GetClaims(ctx).Attributes(), instID, mv.State.ValidSpec, r.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return nil, false, err
	}

	if !security.Access {
		return nil, false, nil
	}

	mv, changed := s.applySecurityPolicyMetricsViewIncludesAndExcludes(mv, security)
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
func (s *Server) applySecurityPolicyMetricsViewIncludesAndExcludes(mv *runtimev1.MetricsViewV2, policy *runtime.ResolvedMetricsViewSecurity) (*runtimev1.MetricsViewV2, bool) {
	if policy == nil {
		return mv, false
	}
	mv = proto.Clone(mv).(*runtimev1.MetricsViewV2)

	if policy.ExcludeAll {
		mv.Spec.Measures = make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
		mv.Spec.Dimensions = make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
		if mv.State.ValidSpec != nil {
			mv.State.ValidSpec.Measures = make([]*runtimev1.MetricsViewSpec_MeasureV2, 0)
			mv.State.ValidSpec.Dimensions = make([]*runtimev1.MetricsViewSpec_DimensionV2, 0)
		}
		return mv, true
	}

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

// applySecurityPolicyReport applies security policies to a report.
// TODO: This implementation is very specific to properties currently set by the admin server. Consider refactoring to a more generic implementation.
func (s *Server) applySecurityPolicyReport(ctx context.Context, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	report := r.GetReport()
	claims := auth.GetClaims(ctx)

	// Allow if the owner is accessing the report
	if report.Spec.Annotations != nil && claims.Subject() == report.Spec.Annotations["admin_owner_user_id"] {
		return r, true, nil
	}

	// Extract admin attributes
	var email string
	admin := true // If no attributes are set, assume it's an admin
	if attrs := claims.Attributes(); len(attrs) != 0 {
		email, _ = attrs["email"].(string)
		admin, _ = attrs["admin"].(bool)
	}

	// Allow if the user is an admin
	if admin {
		return r, true, nil
	}

	// Allow if the user is a recipient
	for _, notifier := range report.Spec.Notifiers {
		switch notifier.Connector {
		case "email":
			recipients := pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
			for _, recipient := range recipients {
				if recipient == email {
					return r, true, nil
				}
			}
		case "slack":
			props, err := slack.DecodeProps(notifier.Properties.AsMap())
			if err != nil {
				return nil, false, err
			}
			for _, user := range props.Users {
				if user == email {
					return r, true, nil
				}
			}
		}
	}

	// Don't allow
	return nil, false, nil
}

// applySecurityPolicyAlert applies security policies to an alert.
// TODO: This implementation is very specific to properties currently set by the admin server. Consider refactoring to a more generic implementation.
func (s *Server) applySecurityPolicyAlert(ctx context.Context, r *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	alert := r.GetAlert()
	claims := auth.GetClaims(ctx)

	// Allow if the owner is accessing the alert
	if alert.Spec.Annotations != nil && claims.Subject() == alert.Spec.Annotations["admin_owner_user_id"] {
		return r, true, nil
	}

	// Extract admin attributes
	var email string
	admin := true // If no attributes are set, assume it's an admin
	if attrs := claims.Attributes(); len(attrs) != 0 {
		email, _ = attrs["email"].(string)
		admin, _ = attrs["admin"].(bool)
	}

	// Allow if the user is an admin
	if admin {
		return r, true, nil
	}

	// Allow if the user is an email recipient
	for _, notifier := range alert.Spec.Notifiers {
		switch notifier.Connector {
		case "email":
			recipients := pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
			for _, recipient := range recipients {
				if recipient == email {
					return r, true, nil
				}
			}
		case "slack":
			props, err := slack.DecodeProps(notifier.Properties.AsMap())
			if err != nil {
				return nil, false, err
			}
			for _, user := range props.Users {
				if user == email {
					return r, true, nil
				}
			}
		}
	}

	// Don't allow
	return nil, false, nil
}
