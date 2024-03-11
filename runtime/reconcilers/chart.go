package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindChart, newChartReconciler)
}

type ChartReconciler struct {
	C *runtime.Controller
}

func newChartReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ChartReconciler{C: c}
}

func (r *ChartReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ChartReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetChart()
	b := to.GetChart()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ChartReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetChart()
	b := to.GetChart()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ChartReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetChart().State = &runtimev1.ChartState{}
	return nil
}

func (r *ChartReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetChart()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a chart")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	err = checkRefs(ctx, r.C, self.Meta.Refs)

	return runtime.ReconcileResult{Err: err}
}
