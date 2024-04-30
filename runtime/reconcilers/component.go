package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindComponent, newComponentReconciler)
}

type ComponentReconciler struct {
	C *runtime.Controller
}

func newComponentReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ComponentReconciler{C: c}
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
	t := self.GetComponent()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a Component")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	err = checkRefs(ctx, r.C, self.Meta.Refs)

	return runtime.ReconcileResult{Err: err}
}
