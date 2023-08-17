package reconcilers

import (
	"context"
	"fmt"
	"time"

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

func newPullTriggerReconciler(c *runtime.Controller) runtime.Reconciler {
	return &PullTriggerReconciler{C: c}
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
	b.Spec = a.Spec
	return nil
}

func (r *PullTriggerReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	err = r.C.Retrigger(ctx, GlobalProjectParserName, time.Time{})
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	err = r.C.Delete(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
