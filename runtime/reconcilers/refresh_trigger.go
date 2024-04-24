package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindRefreshTrigger, newRefreshTriggerReconciler)
}

// RefreshTriggerReconciler reconciles a RefreshTrigger.
// When a RefreshTrigger is created, the reconciler will refresh source and model by setting Trigger=true in their specs.
// After that, it will delete the RefreshTrigger resource.
type RefreshTriggerReconciler struct {
	C *runtime.Controller
}

func newRefreshTriggerReconciler(c *runtime.Controller) runtime.Reconciler {
	return &RefreshTriggerReconciler{C: c}
}

func (r *RefreshTriggerReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *RefreshTriggerReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetRefreshTrigger()
	b := to.GetRefreshTrigger()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *RefreshTriggerReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetRefreshTrigger()
	b := to.GetRefreshTrigger()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *RefreshTriggerReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetRefreshTrigger().State = &runtimev1.RefreshTriggerState{}
	return nil
}

func (r *RefreshTriggerReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	trigger := self.GetRefreshTrigger()
	if trigger == nil {
		return runtime.ReconcileResult{Err: errors.New("not a refresh trigger")}
	}

	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	resources, err := r.C.List(ctx, "", "", false)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	for _, res := range resources {
		// Check sources and models always; also check reports and alerts if OnlyNames is not empty (i.e. explicitly specified).
		switch res.Meta.Name.Kind {
		case runtime.ResourceKindSource, runtime.ResourceKindModel:
			// nothing to do
		case runtime.ResourceKindReport, runtime.ResourceKindAlert:
			if len(trigger.Spec.OnlyNames) == 0 {
				// skip
				continue
			}
		default:
			// skip
			continue
		}

		// Check if it's in OnlyNames
		if len(trigger.Spec.OnlyNames) > 0 {
			found := false
			for _, n := range trigger.Spec.OnlyNames {
				if strings.EqualFold(n.Name, res.Meta.Name.Name) {
					// If Kind is empty, match any kind
					if n.Kind == "" || n.Kind == res.Meta.Name.Kind {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		}

		// Set Trigger=true
		updated := true
		switch res.Meta.Name.Kind {
		case runtime.ResourceKindSource:
			source := res.GetSource()
			source.Spec.Trigger = true
		case runtime.ResourceKindModel:
			model := res.GetModel()
			model.Spec.Trigger = true
		case runtime.ResourceKindReport:
			report := res.GetReport()
			report.Spec.Trigger = true
		case runtime.ResourceKindAlert:
			alert := res.GetAlert()
			alert.Spec.Trigger = true
		default:
			updated = false
		}
		if updated {
			err = r.C.UpdateSpec(ctx, res.Meta.Name, res)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}
	}

	err = r.C.Delete(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
