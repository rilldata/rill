package runtime

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/parser"
	"google.golang.org/protobuf/proto"
)

// Built-in resource kinds
const (
	ResourceKindProjectParser  string = "rill.runtime.v1.ProjectParser"
	ResourceKindSource         string = "rill.runtime.v1.Source"
	ResourceKindModel          string = "rill.runtime.v1.Model"
	ResourceKindMetricsView    string = "rill.runtime.v1.MetricsView"
	ResourceKindExplore        string = "rill.runtime.v1.Explore"
	ResourceKindMigration      string = "rill.runtime.v1.Migration"
	ResourceKindReport         string = "rill.runtime.v1.Report"
	ResourceKindAlert          string = "rill.runtime.v1.Alert"
	ResourceKindRefreshTrigger string = "rill.runtime.v1.RefreshTrigger"
	ResourceKindTheme          string = "rill.runtime.v1.Theme"
	ResourceKindComponent      string = "rill.runtime.v1.Component"
	ResourceKindCanvas         string = "rill.runtime.v1.Canvas"
	ResourceKindAPI            string = "rill.runtime.v1.API"
	ResourceKindConnector      string = "rill.runtime.v1.Connector"
)

// ResourceKindFromPretty converts a user-friendly resource kind to a runtime resource kind.
// If the kind doesn't match a known shorthand, it is returned as-is.
func ResourceKindFromShorthand(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "projectparser", "project_parser":
		return ResourceKindProjectParser
	case "source":
		return ResourceKindSource
	case "model":
		return ResourceKindModel
	case "metricsview", "metrics_view":
		return ResourceKindMetricsView
	case "explore":
		return ResourceKindExplore
	case "migration":
		return ResourceKindMigration
	case "report":
		return ResourceKindReport
	case "alert":
		return ResourceKindAlert
	case "refreshtrigger", "refresh_trigger":
		return ResourceKindRefreshTrigger
	case "theme":
		return ResourceKindTheme
	case "component":
		return ResourceKindComponent
	case "canvas":
		return ResourceKindCanvas
	case "api":
		return ResourceKindAPI
	case "connector":
		return ResourceKindConnector
	default:
		return kind
	}
}

// ResourceKindFromParser converts a parser resource kind to a runtime resource kind.
func ResourceKindFromParser(kind parser.ResourceKind) string {
	switch kind {
	case parser.ResourceKindSource:
		return ResourceKindSource
	case parser.ResourceKindModel:
		return ResourceKindModel
	case parser.ResourceKindMetricsView:
		return ResourceKindMetricsView
	case parser.ResourceKindExplore:
		return ResourceKindExplore
	case parser.ResourceKindMigration:
		return ResourceKindMigration
	case parser.ResourceKindReport:
		return ResourceKindReport
	case parser.ResourceKindAlert:
		return ResourceKindAlert
	case parser.ResourceKindTheme:
		return ResourceKindTheme
	case parser.ResourceKindComponent:
		return ResourceKindComponent
	case parser.ResourceKindCanvas:
		return ResourceKindCanvas
	case parser.ResourceKindAPI:
		return ResourceKindAPI
	case parser.ResourceKindConnector:
		return ResourceKindConnector
	default:
		panic(fmt.Errorf("unknown parser resource type %q", kind))
	}
}

// ResourceKindToParser converts a runtime resource kind to a parser resource kind.
func ResourceKindToParser(kind string) parser.ResourceKind {
	switch kind {
	case ResourceKindSource:
		return parser.ResourceKindSource
	case ResourceKindModel:
		return parser.ResourceKindModel
	case ResourceKindMetricsView:
		return parser.ResourceKindMetricsView
	case ResourceKindExplore:
		return parser.ResourceKindExplore
	case ResourceKindMigration:
		return parser.ResourceKindMigration
	case ResourceKindReport:
		return parser.ResourceKindReport
	case ResourceKindAlert:
		return parser.ResourceKindAlert
	case ResourceKindTheme:
		return parser.ResourceKindTheme
	case ResourceKindComponent:
		return parser.ResourceKindComponent
	case ResourceKindCanvas:
		return parser.ResourceKindCanvas
	case ResourceKindAPI:
		return parser.ResourceKindAPI
	case ResourceKindConnector:
		return parser.ResourceKindConnector
	case ResourceKindProjectParser, ResourceKindRefreshTrigger:
		panic(fmt.Errorf("unsupported resource type %q", kind))
	default:
		panic(fmt.Errorf("unknown resource type %q", kind))
	}
}

// ResourceNameFromParser converts a parser resource name to a runtime resource name.
func ResourceNameFromParser(name parser.ResourceName) *runtimev1.ResourceName {
	return &runtimev1.ResourceName{Kind: ResourceKindFromParser(name.Kind), Name: name.Name}
}

// ResourceNameToParser converts a runtime resource name to a parser resource name.
func ResourceNameToParser(name *runtimev1.ResourceName) parser.ResourceName {
	return parser.ResourceName{Kind: ResourceKindToParser(name.Kind), Name: name.Name}
}

// PrettifyResourceKind returns the resource kind in a user-friendly format suitable for printing.
func PrettifyResourceKind(k string) string {
	k = strings.TrimPrefix(k, "rill.runtime.v1.")
	k = strings.TrimSuffix(k, "V2")
	return k
}

// PrettifyReconcileStatus returns the reconcile status in a user-friendly format suitable for printing.
func PrettifyReconcileStatus(s runtimev1.ReconcileStatus) string {
	switch s {
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_UNSPECIFIED:
		return "Unknown"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
		return "Idle"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
		return "Pending"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
		return "Running"
	default:
		panic(fmt.Errorf("unknown reconcile status: %s", s.String()))
	}
}

// ApplySecurityPolicy applies relevant security policies to the resource.
// The input resource will not be modified in-place (so no need to set clone=true when obtaining it from the catalog).
func (r *Runtime) ApplySecurityPolicy(ctx context.Context, instID string, claims *SecurityClaims, res *runtimev1.Resource) (*runtimev1.Resource, bool, error) {
	security, err := r.ResolveSecurity(ctx, instID, claims, res)
	if err != nil {
		return nil, false, err
	}

	if security == nil {
		return res, true, nil
	}

	if !security.CanAccess() {
		return nil, false, nil
	}

	// Some resources may need deeper checks than just access.
	switch res.Resource.(type) {
	case *runtimev1.Resource_MetricsView:
		// For metrics views, we need to remove fields excluded by the field access rules.
		return r.applyMetricsViewSecurity(res, security), true, nil
	case *runtimev1.Resource_Explore:
		// For explores, we need to remove fields excluded by the field access rules.
		return r.applyExploreSecurity(res, security), true, nil
	default:
		// The resource can be returned as is.
		return res, true, nil
	}
}

// applyMetricsViewSecurity rewrites a metrics view based on the field access conditions of a security policy.
func (r *Runtime) applyMetricsViewSecurity(res *runtimev1.Resource, security *ResolvedSecurity) *runtimev1.Resource {
	if security.CanAccessAllFields() {
		return res
	}

	mv := res.GetMetricsView()
	specDims, specMeasures, specChanged := r.applyMetricsViewSpecSecurity(mv.Spec, security)
	validSpecDims, validSpecMeasures, validSpecChanged := r.applyMetricsViewSpecSecurity(mv.State.ValidSpec, security)

	if !specChanged && !validSpecChanged {
		return res
	}

	mv = proto.Clone(mv).(*runtimev1.MetricsView)

	if specChanged {
		mv.Spec.Dimensions = specDims
		mv.Spec.Measures = specMeasures
	}

	if validSpecChanged {
		mv.State.ValidSpec.Dimensions = validSpecDims
		mv.State.ValidSpec.Measures = validSpecMeasures
	}

	// We mustn't modify the resource in-place
	return &runtimev1.Resource{
		Meta:     res.Meta,
		Resource: &runtimev1.Resource_MetricsView{MetricsView: mv},
	}
}

// applyMetricsViewSpecSecurity rewrites a metrics view spec based on the field access conditions of a security policy.
func (r *Runtime) applyMetricsViewSpecSecurity(spec *runtimev1.MetricsViewSpec, security *ResolvedSecurity) ([]*runtimev1.MetricsViewSpec_Dimension, []*runtimev1.MetricsViewSpec_Measure, bool) {
	if spec == nil {
		return nil, nil, false
	}

	var dims []*runtimev1.MetricsViewSpec_Dimension
	for _, dim := range spec.Dimensions {
		if security.CanAccessField(dim.Name) {
			dims = append(dims, dim)
		}
	}

	var ms []*runtimev1.MetricsViewSpec_Measure
	for _, m := range spec.Measures {
		if security.CanAccessField(m.Name) {
			ms = append(ms, m)
		}
	}

	if len(dims) == len(spec.Dimensions) && len(ms) == len(spec.Measures) {
		return nil, nil, false
	}

	return dims, ms, true
}

// applyExploreSecurity rewrites an explore based on the field access conditions of a security policy.
func (r *Runtime) applyExploreSecurity(res *runtimev1.Resource, security *ResolvedSecurity) *runtimev1.Resource {
	if security.CanAccessAllFields() {
		return res
	}

	// We only rewrite the ValidSpec at the moment.
	// In the future, to avoid leaking field names in the main spec (which is not really used outside of the reconciler),
	// we might consider not returning the spec at all for non-admins.
	spec := res.GetExplore().State.ValidSpec
	if spec == nil {
		return res
	}
	if spec.DimensionsSelector != nil || spec.MeasuresSelector != nil {
		// If the ValidSpec has dynamic selectors, we don't know what the available fields, so we can't filter it correctly.
		// This should never happen because the Explore reconciler should have resolved the fields and removed the exclude flags.
		panic(fmt.Errorf("the ValidSpec for an explore should not have exclude flags set"))
	}

	// Clone the spec so we can edit it in-place
	spec = proto.Clone(spec).(*runtimev1.ExploreSpec)

	// Filter the dimensions
	var dims []string
	for _, dim := range spec.Dimensions {
		if security.CanAccessField(dim) {
			dims = append(dims, dim)
		}
	}
	spec.Dimensions = dims

	// Filter the measures
	var ms []string
	for _, m := range spec.Measures {
		if security.CanAccessField(m) {
			ms = append(ms, m)
		}
	}
	spec.Measures = ms

	// Filter the dimensions and measures in the presets
	if spec.DefaultPreset != nil {
		p := spec.DefaultPreset

		var dims []string
		for _, dim := range p.Dimensions {
			if security.CanAccessField(dim) {
				dims = append(dims, dim)
			}
		}
		p.Dimensions = dims

		var ms []string
		for _, m := range p.Measures {
			if security.CanAccessField(m) {
				ms = append(ms, m)
			}
		}
		p.Measures = ms
	}

	// We mustn't modify the resource in-place
	return &runtimev1.Resource{
		Meta: res.Meta,
		Resource: &runtimev1.Resource_Explore{Explore: &runtimev1.Explore{
			Spec:  res.GetExplore().Spec,
			State: &runtimev1.ExploreState{ValidSpec: spec},
		}},
	}
}
