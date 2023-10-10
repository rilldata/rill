package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindReport, newReportReconciler)
}

type ReportReconciler struct {
	C *runtime.Controller
}

func newReportReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ReportReconciler{C: c}
}

func (r *ReportReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ReportReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ReportReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ReportReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetReport().State = &runtimev1.ReportState{}
	return nil
}

func (r *ReportReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	rep := self.GetReport()
	if rep == nil {
		return runtime.ReconcileResult{Err: errors.New("not a report")}
	}

	// NOTE: refs not supported for reports.
	// Not supporting them simplifies report spec generation, improves performance (there may be many reports),
	// and it's anyway acceptable (maybe even expected) that a report fails with an execution error if the underlying operation errors.

	r.C.Logger.Info("Running report", "name", rep.Spec.Title)

	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
