package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindMetricsView, newMetricsViewReconciler)
}

type MetricsViewReconciler struct {
	C *runtime.Controller
}

func newMetricsViewReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &MetricsViewReconciler{C: c}, nil
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

	// Get instance config
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// If the spec references a model, try resolving it to a table before validating it.
	// For backwards compatibility, the model may actually be a source or external table.
	// So if a model is not found, we optimistically use the model name as the table and proceed to validation
	if mv.Spec.Model != "" {
		res, err := r.C.Get(ctx, &runtimev1.ResourceName{Name: mv.Spec.Model, Kind: runtime.ResourceKindModel}, false)
		if err == nil && res.GetModel().State.ResultTable != "" {
			mv.Spec.Table = res.GetModel().State.ResultTable
			mv.Spec.Connector = res.GetModel().State.ResultConnector
		} else {
			mv.Spec.Table = mv.Spec.Model
		}
	}

	// Find out if the metrics view has a ref to a source or model in the same project.
	hasInternalRef := false
	for _, ref := range self.Meta.Refs {
		if ref.Kind == runtime.ResourceKindSource || ref.Kind == runtime.ResourceKindModel {
			hasInternalRef = true
		}
	}

	// NOTE: In other reconcilers, state like spec_hash and refreshed_on is used to avoid redundant reconciles.
	// We don't do that here because none of the operations below are particularly expensive.
	// So it doesn't really matter if they run a bit more often than necessary ¯\_(ツ)_/¯.

	// NOTE: Not checking refs for errors since they may still be valid even if they have errors. Instead, we just validate the metrics view against the table name.

	// Validate the metrics view and update ValidSpec
	e, err := metricsview.NewExecutor(ctx, r.C.Runtime, r.C.InstanceID, mv.Spec, !hasInternalRef, runtime.ResolvedSecurityOpen, 0)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to create metrics view executor: %w", err)}
	}
	defer e.Close()
	validateResult, validateErr := e.ValidateMetricsView(ctx)
	if validateErr == nil {
		validateErr = validateResult.Error()
	}
	if ctx.Err() != nil { // May not be handled in all validation implementations
		return runtime.ReconcileResult{Err: ctx.Err()}
	}
	if validateErr != nil {
		// When not staging changes, clear the previously valid spec.
		// Otherwise, we keep serving the previously valid spec.
		if !cfg.StageChanges {
			mv.State.ValidSpec = nil
			mv.State.Streaming = false
			err = r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}

		// Return the validation error
		return runtime.ReconcileResult{Err: validateErr}
	}

	// Capture the spec, which we now know to be valid.
	mv.State.ValidSpec = mv.Spec
	// If there's no internal ref, we assume the metrics view is based on an externally managed table and set the streaming state to true.
	mv.State.Streaming = !hasInternalRef
	// Update the state. Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
