package reconcilers

import (
	"context"
	"time"

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

func (r *PullTriggerReconciler) Reconcile(ctx context.Context, s *runtime.Signal) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, s.Name)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	if self.Meta.Deleted {
		return runtime.ReconcileResult{}
	}

	err = r.C.Retrigger(ctx, GlobalProjectParserName, time.Time{})
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	err = r.C.Delete(ctx, s.Name)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
