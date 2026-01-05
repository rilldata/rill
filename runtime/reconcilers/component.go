package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindComponent, newComponentReconciler)
}

type ComponentReconciler struct {
	C *runtime.Controller
}

func newComponentReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ComponentReconciler{C: c}, nil
}

func (r *ComponentReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ComponentReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetComponent()
	b := to.GetComponent()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ComponentReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetComponent()
	b := to.GetComponent()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ComponentReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetComponent().State = &runtimev1.ComponentState{}
	return nil
}

func (r *ComponentReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	c := self.GetComponent()
	if c == nil {
		return runtime.ReconcileResult{Err: errors.New("not a component")}
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

	// Validate all refs
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)

	// Check metrics view refs specifically (even if validateErr != nil)
	validMetrics, dataRefreshedOn, err := r.checkMetricsViews(ctx, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Capture the valid spec in the state
	if validateErr == nil {
		c.State.ValidSpec = c.Spec
		c.State.DataRefreshedOn = dataRefreshedOn
	} else if cfg.StageChanges && validMetrics {
		// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
		// If all the metrics view(s) referenced by the spec have a ValidSpec, we'll consider the component valid.
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

func (r *ComponentReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetComponent() == nil {
		return nil, fmt.Errorf("not a component resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// checkMetricsViews returns true if all the metrics views referenced by the component have a valid spec.
// If all metrics views are valid, it also returns the most recent DataRefreshedOn timestamp across all referenced metrics views.
// Note that it returns false if no metrics views are referenced.
func (r *ComponentReconciler) checkMetricsViews(ctx context.Context, refs []*runtimev1.ResourceName) (bool, *timestamppb.Timestamp, error) {
	var n int
	var dataRefreshedOn *timestamppb.Timestamp
	for _, ref := range refs {
		if ref.Kind != runtime.ResourceKindMetricsView {
			continue
		}

		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return false, nil, nil
			}
			return false, nil, err
		}
		if res.GetMetricsView().State.ValidSpec == nil {
			return false, nil, nil
		}

		n++

		t := res.GetMetricsView().State.DataRefreshedOn
		if dataRefreshedOn == nil {
			dataRefreshedOn = t
		} else if t != nil && t.AsTime().After(dataRefreshedOn.AsTime()) {
			dataRefreshedOn = t
		}
	}
	return n > 0, dataRefreshedOn, nil
}
