package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindCanvas, newCanvasReconciler)
}

type CanvasReconciler struct {
	C *runtime.Controller
}

func newCanvasReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &CanvasReconciler{C: c}, nil
}

func (r *CanvasReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *CanvasReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetCanvas()
	b := to.GetCanvas()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *CanvasReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetCanvas()
	b := to.GetCanvas()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *CanvasReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetCanvas().State = &runtimev1.CanvasState{}
	return nil
}

func (r *CanvasReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	c := self.GetCanvas()
	if c == nil {
		return runtime.ReconcileResult{Err: errors.New("not a canvas")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Get instance config
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Get all referenced components.
	// If a referenced component is not found, we still add it to the map with a nil value.
	components := make(map[string]*runtimev1.Resource)
	for _, ref := range self.Meta.Refs {
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: err}
			}
			components[ref.Name] = nil // Component not found, add it to the map with nil value
		} else {
			components[ref.Name] = res
		}
	}

	// Find most recent data refresh time across all components.
	var dataRefreshedOn *timestamppb.Timestamp
	for _, c := range components {
		if c == nil {
			continue
		}
		t := c.GetComponent().State.DataRefreshedOn
		if t == nil {
			continue
		}
		if dataRefreshedOn == nil || t.AsTime().After(dataRefreshedOn.AsTime()) {
			dataRefreshedOn = t
		}
	}

	// Validate refs
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)
	if validateErr == nil {
		validateErr = r.validateMetricsViewTimeConsistency(ctx, components)
	}

	// Capture the valid spec in the state
	if validateErr == nil {
		c.State.ValidSpec = c.Spec
		c.State.DataRefreshedOn = dataRefreshedOn
	} else if cfg.StageChanges && r.checkAnyComponentHasValidSpec(components) {
		// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
		// If any of the components referenced by the spec have a ValidSpec, we'll try to serve the canvas.
		c.State.ValidSpec = c.Spec
		c.State.DataRefreshedOn = dataRefreshedOn
	} else {
		c.State.ValidSpec = nil
		c.State.DataRefreshedOn = nil
	}

	// Update state. Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}

// checkAnyComponentHasValidSpec returns true if one or more components have a valid spec.
func (r *CanvasReconciler) checkAnyComponentHasValidSpec(components map[string]*runtimev1.Resource) bool {
	for _, res := range components {
		if res == nil {
			// Component not found, skip it
			continue
		}
		if res.GetComponent().State.ValidSpec != nil {
			// Found component ref with a valid spec
			return true
		}
	}
	return false
}

func (r *CanvasReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName
	refs := &rendererRefs{
		rt:           r.C.Runtime,
		instanceID:   r.C.InstanceID,
		claims:       claims,
		metricsViews: make(map[string]bool),
	}

	canvas := res.GetCanvas()
	if canvas == nil {
		return nil, fmt.Errorf("resource is not a canvas")
	}
	spec := canvas.GetState().GetValidSpec()
	if spec == nil {
		spec = canvas.GetSpec() // Fallback to spec if ValidSpec is not available
	}
	if spec == nil {
		return nil, fmt.Errorf("canvas spec is nil")
	}

	// explicitly allow access to the canvas itself
	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, runtime.ResourceKindTheme)

	// Get controller to fetch components
	ctr, err := r.C.Runtime.Controller(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}

	// Collect all component names referenced by the canvas
	componentNames := make(map[string]bool)
	for _, row := range spec.Rows {
		for _, item := range row.Items {
			componentNames[item.Component] = true
		}
	}

	// Process each component
	for componentName := range componentNames {
		componentRef := &runtimev1.ResourceName{
			Kind: runtime.ResourceKindComponent,
			Name: componentName,
		}
		// Allow access to the component itself
		conditionResources = append(conditionResources, componentRef)

		// Get component resource
		componentRes, err := ctr.Get(ctx, componentRef, false)
		if err != nil {
			// If component is not found, skip it but still allow access to the component name
			continue
		}

		// Get component spec to extract renderer properties
		componentSpec := componentRes.GetComponent().State.ValidSpec
		if componentSpec == nil {
			componentSpec = componentRes.GetComponent().Spec
		}
		if componentSpec.RendererProperties == nil {
			continue
		}

		// Track refs.
		// We silently ignore parse errors because some components may be malformed, which we don't want to block access to the entire canvas.
		// Hopefully the parse errors were caught in normal validation; otherwise, the UI will probably fail the component at render time.
		_ = refs.populateRendererRefs(ctx, componentSpec.Renderer, componentSpec.RendererProperties.AsMap())
		if ctx.Err() != nil { // Return ctx errors immediately.
			return nil, ctx.Err()
		}
	}

	// Add the discovered refs to the condition resources.
	conditionResources = append(conditionResources, refs.result()...)

	// Now build security rules based on the collected references.
	if len(conditionKinds) > 0 || len(conditionResources) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionKinds:     conditionKinds,
					ConditionResources: conditionResources,
					Allow:              true,
					Exclusive:          true,
				},
			},
		})
	}

	return rules, nil
}

// validateMetricsViewTimeConsistency checks that all the metrics views referenced by the canvas' components have the same first_day_of_week and first_month_of_year.
func (r *CanvasReconciler) validateMetricsViewTimeConsistency(ctx context.Context, components map[string]*runtimev1.Resource) error {
	metricsViews := make(map[string]*runtimev1.Resource)
	for _, component := range components {
		if component == nil {
			// Component not found, skip it
			continue
		}

		// Skip non-metrics view refs
		for _, ref := range component.Meta.Refs {
			if ref.Kind != runtime.ResourceKindMetricsView {
				continue
			}
			// Skip if the metrics view is already in the map
			if _, ok := metricsViews[ref.Name]; ok {
				continue
			}
			// Get the metrics view
			mv, err := r.C.Get(ctx, ref, false)
			if err != nil {
				continue
			}

			// Skip if the metrics view is not valid
			mvSpec := mv.GetMetricsView().State.ValidSpec
			if mvSpec == nil {
				continue
			}

			metricsViews[ref.Name] = mv
		}
	}

	// Validate all metrics views have consistent first_day_of_week or first_month_of_year
	if len(metricsViews) > 0 {
		first := false
		var firstDayOfWeek uint32
		var firstMonthOfYear uint32
		var firstViewName string

		for mvName, mv := range metricsViews {
			mvSpec := mv.GetMetricsView().State.ValidSpec
			if mvSpec == nil {
				continue
			}

			if !first {
				// This is the first metrics view, store its settings as reference
				first = true
				firstDayOfWeek = mvSpec.FirstDayOfWeek
				firstMonthOfYear = mvSpec.FirstMonthOfYear
				firstViewName = mvName
			} else {
				// Compare subsequent views with the first one
				if firstDayOfWeek != mvSpec.FirstDayOfWeek {
					return status.Errorf(codes.InvalidArgument, "metrics views %q and %q have inconsistent first_day_of_week", firstViewName, mvName)
				} else if firstMonthOfYear != mvSpec.FirstMonthOfYear {
					return status.Errorf(codes.InvalidArgument, "metrics views %q and %q have inconsistent first_month_of_year", firstViewName, mvName)
				}
			}
		}
	}

	return nil
}

// rendererRefs tracks all metrics views found in canvas component renderer properties.
// It currently only tracks metrics views, but in the future we may want to add an option to also track metrics view fields and filters.
// We did that previously, but removed it since such granular security was considered too strict (it also impacts ability to filter by fields not present on the canvas).
// See this PR for details in case we want to reintroduce it: https://github.com/rilldata/rill/pull/8370
type rendererRefs struct {
	rt           *runtime.Runtime
	instanceID   string
	claims       *runtime.SecurityClaims
	metricsViews map[string]bool
}

// result returns the accumulated refs.
func (r *rendererRefs) result() []*runtimev1.ResourceName {
	refs := make([]*runtimev1.ResourceName, 0, len(r.metricsViews))
	for mv := range r.metricsViews {
		refs = append(refs, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: mv})
	}
	return refs
}

// populateRendererRefs discovers and tracks all metrics views referenced in the renderer properties.
func (r *rendererRefs) populateRendererRefs(ctx context.Context, renderer string, rendererProps map[string]any) error {
	// Check for a direct metrics_view reference.
	if mv, ok := pathutil.GetPath(rendererProps, "metrics_view"); ok {
		return r.metricsView(mv)
	}

	// Check for a metrics_sql reference; use a resolver to discover the referenced metrics views.
	if sql, ok := pathutil.GetPath(rendererProps, "metrics_sql"); ok {
		return r.metricsSQL(ctx, sql)
	}

	// For markdown renderers, analyze the content text for embedded metrics_sql references.
	if renderer == "markdown" {
		if content, ok := pathutil.GetPath(rendererProps, "content"); ok {
			return r.text(ctx, content)
		}
	}

	return nil
}

// metricsView registers a metrics view reference.
func (r *rendererRefs) metricsView(mv any) error {
	if mv, ok := mv.(string); ok {
		r.metricsViews[mv] = true
		return nil
	}
	return fmt.Errorf("metrics view field is not a string")
}

// text initializes a text resolver to discover metrics view references in a template string.
func (r *rendererRefs) text(ctx context.Context, content any) error {
	contentStr, ok := content.(string)
	if !ok {
		return fmt.Errorf("content field is not a string")
	}

	initializer, ok := runtime.ResolverInitializers["text"]
	if !ok {
		return fmt.Errorf("text resolver not registered")
	}
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    r.rt,
		InstanceID: r.instanceID,
		Properties: map[string]any{"text": contentStr},
		Claims: &runtime.SecurityClaims{
			UserID:         r.claims.UserID,
			UserAttributes: r.claims.UserAttributes,
			SkipChecks:     true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize text resolver: %w", err)
	}
	defer resolver.Close()

	for _, ref := range resolver.Refs() {
		if ref.Kind == runtime.ResourceKindMetricsView {
			r.metricsViews[ref.Name] = true
		}
	}
	return nil
}

// metricsSQL parses and registers metrics view references found in a metrics SQL string.
func (r *rendererRefs) metricsSQL(ctx context.Context, sql any) error {
	sqlStr, ok := sql.(string)
	if !ok {
		return fmt.Errorf("metrics_sql field is not a string")
	}

	initializer, ok := runtime.ResolverInitializers["metrics_sql"]
	if !ok {
		return fmt.Errorf("metrics_sql resolver not registered")
	}
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    r.rt,
		InstanceID: r.instanceID,
		Properties: map[string]any{"sql": sqlStr},
		Claims: &runtime.SecurityClaims{
			UserID:         r.claims.UserID,
			UserAttributes: r.claims.UserAttributes,
			SkipChecks:     true, // To avoid infinite recursion
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize metrics_sql resolver: %w", err)
	}
	defer resolver.Close()

	for _, ref := range resolver.Refs() {
		if ref.Kind == runtime.ResourceKindMetricsView {
			r.metricsViews[ref.Name] = true
		}
	}
	return nil
}
