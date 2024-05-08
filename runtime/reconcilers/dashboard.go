package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindDashboard, newDashboardReconciler)
}

type DashboardReconciler struct {
	C *runtime.Controller
}

func newDashboardReconciler(c *runtime.Controller) runtime.Reconciler {
	return &DashboardReconciler{C: c}
}

func (r *DashboardReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *DashboardReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetDashboard()
	b := to.GetDashboard()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *DashboardReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetDashboard()
	b := to.GetDashboard()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *DashboardReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetDashboard().State = &runtimev1.DashboardState{}
	return nil
}

func (r *DashboardReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetDashboard()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a dashboard")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	err = checkRefs(ctx, r.C, self.Meta.Refs)

	return runtime.ReconcileResult{Err: err}
}
