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

	// Validate refs
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)

	//
	if validateErr == nil {
		validateErr = r.validateMetricsViewTimeConsistency(ctx, self.Meta.Refs)
	}

	// Capture the valid spec in the state
	if validateErr == nil {
		c.State.ValidSpec = c.Spec
	} else if !cfg.StageChanges {
		c.State.ValidSpec = nil
	} else {
		// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
		// If any of the components referenced by the spec have a ValidSpec, we'll try to serve the canvas.
		validComponents, err := r.checkAnyComponentHasValidSpec(ctx, self.Meta.Refs)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if validComponents {
			c.State.ValidSpec = c.Spec
		} else {
			c.State.ValidSpec = nil
		}
	}

	// Update state. Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}

func (r *CanvasReconciler) checkAnyComponentHasValidSpec(ctx context.Context, refs []*runtimev1.ResourceName) (bool, error) {
	for _, ref := range refs {
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return false, nil
			}
			return false, err
		}
		if res.GetComponent().State.ValidSpec != nil {
			// Found component ref with a valid spec
			return true, nil
		}
	}
	return false, nil
}

// validateMetricsViewTimeConsistency checks that all the metrics views referenced by the canvas' components have the same first_day_of_week and first_month_of_year.
func (r *CanvasReconciler) validateMetricsViewTimeConsistency(ctx context.Context, refs []*runtimev1.ResourceName) error {
	metricsViews := make(map[string]*runtimev1.Resource)
	for _, ref := range refs {
		// Skip non-component refs
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		component, err := r.C.Get(ctx, ref, false)
		if err != nil {
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
				if errors.Is(err, drivers.ErrResourceNotFound) {
					return fmt.Errorf("component %q: metrics view %q not found", ref.Name, ref.Name)
				}
				return err
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
		var first bool = false
		var firstDayOfWeek uint32
		var firstMonthOfYear uint32
		var firstViewName string

		for mvName, mv := range metricsViews {
			mvSpec := mv.GetMetricsView().State.ValidSpec
			if mvSpec == nil {
				return status.Errorf(codes.Internal, "metrics view %q in valid spec not found", mvName)
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
