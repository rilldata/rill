package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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
	if validateErr == nil {
		validateErr = r.validateCanvasFiltersAgainstMetricsViews(ctx, c.Spec)
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

// validateCanvasFiltersAgainstMetricsViews validates that all dimensions and measures
// referenced in the canvas default filters exist in at least one metrics view.
func (r *CanvasReconciler) validateCanvasFiltersAgainstMetricsViews(ctx context.Context, spec *runtimev1.CanvasSpec) error {
	if spec == nil || spec.DefaultPreset == nil || spec.DefaultPreset.Filters == nil {
		return nil
	}

	filters := spec.DefaultPreset.Filters

	allDimensions := make(map[string]bool)
	allMeasures := make(map[string]bool)

	resources, err := r.C.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return fmt.Errorf("failed to list metrics views: %w", err)
	}

	for _, resource := range resources {
		mv := resource.GetMetricsView()
		if mv == nil || mv.State.ValidSpec == nil {
			continue
		}

		spec := mv.State.ValidSpec

	
		for _, dim := range spec.Dimensions {
			if dim.Name != "" {
				allDimensions[dim.Name] = true
			}
		}

	
		for _, measure := range spec.Measures {
			if measure.Name != "" {
				allMeasures[measure.Name] = true
			}
		}
	}

	for _, dimFilter := range filters.Dimensions {
		if dimFilter.Dimension != "" {
			if !allDimensions[dimFilter.Dimension] {
				return fmt.Errorf("dimension filter references unknown dimension %q - dimension must exist in at least one metrics view", dimFilter.Dimension)
			}
		}
	}

	
	for _, measureFilter := range filters.Measures {
		if measureFilter.Measure != "" {
			if !allMeasures[measureFilter.Measure] {
				return fmt.Errorf("measure filter references unknown measure %q - measure must exist in at least one metrics view", measureFilter.Measure)
			}
		}

		if measureFilter.ByDimension != "" {
			if !allDimensions[measureFilter.ByDimension] {
				return fmt.Errorf("measure filter %q references unknown by_dimension %q - dimension must exist in at least one metrics view", measureFilter.Measure, measureFilter.ByDimension)
			}
		}
	}

	return nil
}
