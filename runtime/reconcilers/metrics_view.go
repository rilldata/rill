package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindMetricsView, newMetricsViewReconciler)
}

type MetricsViewReconciler struct {
	C *runtime.Controller
}

func newMetricsViewReconciler(c *runtime.Controller) runtime.Reconciler {
	return &MetricsViewReconciler{C: c}
}

func (r *MetricsViewReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *MetricsViewReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetMetricsView()
	b := to.GetMetricsView()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *MetricsViewReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetMetricsView()
	b := to.GetMetricsView()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *MetricsViewReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetMetricsView().State = &runtimev1.MetricsViewState{}
	return nil
}

func (r *MetricsViewReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	mv := self.GetMetricsView()
	if mv == nil {
		return runtime.ReconcileResult{Err: errors.New("not a metrics view")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// NOTE: Not checking refs here since refs may still be valid even if they have errors (in case of staged changes).
	// Instead, we just validate against the table name.

	validateResult, validateErr := r.C.Runtime.ValidateMetricsView(ctx, r.C.InstanceID, mv.Spec)
	if validateErr == nil {
		validateErr = validateResult.Error()
	}

	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: errors.Join(validateErr, ctx.Err())}
	}

	if validateErr == nil {
		mv.State.ValidSpec = mv.Spec
	} else {
		mv.State.ValidSpec = nil
	}

	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}
