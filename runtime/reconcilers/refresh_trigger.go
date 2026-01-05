package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
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

func newRefreshTriggerReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &RefreshTriggerReconciler{C: c}, nil
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

	// For some resource types, it is sufficient to call r.C.Reconcile without updating the spec.
	// This applies for resources that run a full reconcile on every invocation (i.e. doesn't cache state).
	// We handle these resources here instead of in the loop below since calling r.C.Reconcile directly must be done outside of a catalog lock.
	for i, rn := range trigger.Spec.Resources {
		// Apply to all resources except those that have a dedicated Trigger property.
		var skip bool
		switch rn.Kind {
		case runtime.ResourceKindSource, runtime.ResourceKindModel, runtime.ResourceKindAlert, runtime.ResourceKindReport:
			skip = true
		}
		if skip {
			continue
		}

		err = r.C.Reconcile(ctx, rn)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		// To not modify the list as we iterate over it, we just set the resource to nil and skip it in the loop below.
		trigger.Spec.Resources[i] = nil
	}

	// Get the catalog in case we need to update model partitions
	catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	defer release()

	// Lock the catalog, so we delay any reconciles from starting until we've set all the triggers.
	// This will remove the chance of fast cancellations if resources that are connected in the DAG are getting triggered.
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	// Handle model triggers
	for _, mt := range trigger.Spec.Models {
		mr, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: mt.Model}, true)
		if err != nil {
			// Skip triggers for non-existent models
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: err}
			}
			r.C.Logger.Warn("Skipped trigger for non-existent model", zap.String("model", mt.Model), observability.ZapCtx(ctx))
			continue
		}

		triggerPartitions := len(mt.Partitions) > 0 || mt.AllErroredPartitions
		if triggerPartitions {
			mdl := mr.GetModel()
			modelID := mdl.State.PartitionsModelId
			if !mdl.Spec.Incremental {
				r.C.Logger.Warn("Skipped partitions trigger for model because it is not incremental", zap.String("model", mt.Model), observability.ZapCtx(ctx))
				continue
			}
			if modelID == "" {
				r.C.Logger.Warn("Skipped partitions trigger for model because no partitions have been ingested yet", zap.String("model", mt.Model), observability.ZapCtx(ctx))
				continue
			}

			err = catalog.UpdateModelPartitionsTriggered(ctx, modelID, mt.Partitions, mt.AllErroredPartitions)
			if err != nil {
				return runtime.ReconcileResult{Err: fmt.Errorf("failed to update partitions as triggered for model %s: %w", mt.Model, err)}
			}
		}

		triggerNormal := !triggerPartitions && !mt.Full
		err = r.UpdateModelTrigger(ctx, mr, triggerNormal, mt.Full, triggerPartitions)
		if err != nil {
			// Not handling deletion race conditions because we hold a lock.
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to update trigger for model %q: %w", mt.Model, err)}
		}
	}

	// Handle generic resource triggers
	for _, rn := range trigger.Spec.Resources {
		if rn == nil { // Skip resources that were handled above
			continue
		}

		res, err := r.C.Get(ctx, rn, true)
		if err != nil {
			// Skip triggers for non-existent resources
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: err}
			}
			r.C.Logger.Warn("Skipped trigger for non-existent resource", zap.String("kind", rn.Kind), zap.String("name", rn.Name), observability.ZapCtx(ctx))
			continue
		}

		err = r.UpdateTriggerTrue(ctx, res)
		if err != nil {
			// Not handling deletion race conditions because we hold a lock.
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to update trigger for resource %q: %w", rn.Name, err)}
		}
	}

	// Delete self now that all triggers have been set
	err = r.C.Delete(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}

func (r *RefreshTriggerReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetRefreshTrigger() == nil {
		return nil, fmt.Errorf("not a refresh trigger resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// UpdateTriggerTrue sets the Trigger spec property of the resource to true.
// If you're refreshing a model, consider using UpdateModelTrigger directly for more granular control.
// NOTE: If you edit this logic, also update the checks in newResourceIfModified in project_parser.go accordingly (they need to incorporate triggers in their modified checks).
func (r *RefreshTriggerReconciler) UpdateTriggerTrue(ctx context.Context, res *runtimev1.Resource) error {
	switch res.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		source := res.GetSource()
		if source.Spec.Trigger {
			return nil
		}
		source.Spec.Trigger = true
	case runtime.ResourceKindModel:
		return r.UpdateModelTrigger(ctx, res, true, false, false)
	case runtime.ResourceKindAlert:
		alert := res.GetAlert()
		if alert.Spec.Trigger {
			return nil
		}
		alert.Spec.Trigger = true
	case runtime.ResourceKindReport:
		report := res.GetReport()
		if report.Spec.Trigger {
			return nil
		}
		report.Spec.Trigger = true
	default:
		// Nothing to do
		r.C.Logger.Warn("Attempted to trigger a resource type that is not triggerable", zap.String("kind", res.Meta.Name.Kind), zap.String("name", res.Meta.Name.Name), observability.ZapCtx(ctx))
		return nil
	}

	return r.C.UpdateSpec(ctx, res.Meta.Name, res)
}

// UpdateModelTrigger sets the Trigger, TriggerFull, or TriggerPartitions spec properties of the model resource based on the provided flags.
// Note the function only updates to truthy values; only the model reconciler can set these back to false.
// NOTE: If you edit this logic, also update the checks in newResourceIfModified in project_parser.go accordingly (they need to incorporate triggers in their modified checks).
func (r *RefreshTriggerReconciler) UpdateModelTrigger(ctx context.Context, res *runtimev1.Resource, normal, full, partitions bool) error {
	model := res.GetModel()
	if model == nil {
		return fmt.Errorf("not a model resource")
	}

	updated := false
	if full && !model.Spec.TriggerFull {
		model.Spec.TriggerFull = true
		updated = true
	}
	if normal && !model.Spec.Trigger && !model.Spec.TriggerFull {
		model.Spec.Trigger = true
		updated = true
	}
	if partitions && !model.Spec.TriggerPartitions && !model.Spec.Trigger && !model.Spec.TriggerFull {
		model.Spec.TriggerPartitions = true
		updated = true
	}

	if updated {
		return r.C.UpdateSpec(ctx, res.Meta.Name, res)
	}
	return nil
}
