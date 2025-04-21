package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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

	// Validate
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)

	// Capture the valid spec in the state
	if validateErr == nil {
		c.State.ValidSpec = c.Spec
	} else if !cfg.StageChanges {
		c.State.ValidSpec = nil
	} else {
		// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
		// If all the metrics view(s) referenced by the spec have a ValidSpec, we'll consider the component valid.
		validMetrics, err := r.checkMetricsViewsValidSpec(ctx, self.Meta.Refs)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if validMetrics {
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

// checkMetricsViewsValidSpec returns true if all the metrics views referenced by the component have a valid spec.
// Note that it returns false if no metrics views are referenced.
func (r *ComponentReconciler) checkMetricsViewsValidSpec(ctx context.Context, refs []*runtimev1.ResourceName) (bool, error) {
	var n int
	for _, ref := range refs {
		if ref.Kind != runtime.ResourceKindMetricsView {
			continue
		}
		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return false, nil
			}
			return false, err
		}
		if res.GetMetricsView().State.ValidSpec == nil {
			return false, nil
		}
		n++
	}
	return n > 0, nil
}
