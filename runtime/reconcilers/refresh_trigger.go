package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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

	// As a special case, triggers for the global project parser should actually be handled just by triggering a reconcile on it.
	// We do this here instead of in the loop below since calling r.C.Reconcile directly must be done outside of a catalog lock.
	for i, rn := range trigger.Spec.Resources {
		if !equalResourceName(rn, runtime.GlobalProjectParserName) {
			continue
		}

		err = r.C.Reconcile(ctx, runtime.GlobalProjectParserName)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		trigger.Spec.Resources = slices.Delete(trigger.Spec.Resources, i, i+1)
		break
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
			r.C.Logger.Warn("Skipped trigger for non-existent model", zap.String("model", mt.Model))
			continue
		}

		if len(mt.Partitions) > 0 || mt.AllErroredPartitions {
			mdl := mr.GetModel()
			modelID := mdl.State.PartitionsModelId
			if !mdl.Spec.Incremental {
				r.C.Logger.Warn("Skipped partitions trigger for model because it is not incremental", zap.String("model", mt.Model))
				continue
			}
			if modelID == "" {
				r.C.Logger.Warn("Skipped partitions trigger for model because no partitions have been ingested yet", zap.String("model", mt.Model))
				continue
			}

			if mt.AllErroredPartitions {
				err := catalog.UpdateModelPartitionsPendingIfError(ctx, modelID)
				if err != nil {
					return runtime.ReconcileResult{Err: fmt.Errorf("failed to update partitions for model %s: %w", mt.Model, err)}
				}
			}

			for _, partition := range mt.Partitions {
				err := catalog.UpdateModelPartitionPending(ctx, modelID, partition)
				if err != nil {
					return runtime.ReconcileResult{Err: fmt.Errorf("failed to mark partition %q pending for model %q: %w", partition, mt.Model, err)}
				}
			}
		}

		err = r.UpdateTriggerTrue(ctx, mr, mt.Full)
		if err != nil {
			// Not handling deletion race conditions because we hold a lock.
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to update trigger for model %q: %w", mt.Model, err)}
		}
	}

	// Handle generic resource triggers
	for _, rn := range trigger.Spec.Resources {
		res, err := r.C.Get(ctx, rn, true)
		if err != nil {
			// Skip triggers for non-existent resources
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: err}
			}
			r.C.Logger.Warn("Skipped trigger for non-existent resource", zap.String("kind", rn.Kind), zap.String("name", rn.Name))
			continue
		}

		err = r.UpdateTriggerTrue(ctx, res, false)
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

func (r *RefreshTriggerReconciler) UpdateTriggerTrue(ctx context.Context, res *runtimev1.Resource, full bool) error {
	switch res.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		source := res.GetSource()
		if source.Spec.Trigger {
			return nil
		}
		source.Spec.Trigger = true
	case runtime.ResourceKindModel:
		model := res.GetModel()
		if full {
			if model.Spec.TriggerFull {
				return nil
			}
			model.Spec.TriggerFull = true
		} else {
			if model.Spec.Trigger || model.Spec.TriggerFull {
				return nil
			}
			model.Spec.Trigger = true
		}
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
	case runtime.ResourceKindMetricsView:
		metricsView := res.GetMetricsView()
		if metricsView.Spec.Trigger {
			return nil
		}
	default:
		// Nothing to do
		r.C.Logger.Warn("Attempted to trigger a resource type that is not triggerable", zap.String("kind", res.Meta.Name.Kind), zap.String("name", res.Meta.Name.Name))
		return nil
	}

	return r.C.UpdateSpec(ctx, res.Meta.Name, res)
}
