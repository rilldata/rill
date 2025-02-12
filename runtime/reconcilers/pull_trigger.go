package reconcilers

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindPullTrigger, newPullTriggerReconciler)
}

// PullTriggerReconciler reconciles a PullTrigger.
// When a PullTrigger is created, the reconciler will retrigger the global project parser resource, causing it to pull and reparse the project.
// It will then delete the PullTrigger resource.
type PullTriggerReconciler struct {
	C *runtime.Controller
}

func newPullTriggerReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &PullTriggerReconciler{C: c}, nil
}

func (r *PullTriggerReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *PullTriggerReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetPullTrigger()
	b := to.GetPullTrigger()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *PullTriggerReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetPullTrigger()
	b := to.GetPullTrigger()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *PullTriggerReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetPullTrigger().State = &runtimev1.PullTriggerState{}
	return nil
}

func (r *PullTriggerReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	err = r.C.Reconcile(ctx, runtime.GlobalProjectParserName)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	err = r.C.Delete(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
