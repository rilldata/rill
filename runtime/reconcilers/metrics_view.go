package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	// NOTE: In other reconcilers, state like spec_hash and refreshed_on is used to avoid redundant reconciles.
	// We don't do that here because none of the operations below are particularly expensive.
	// So it doesn't really matter if they run a bit more often than necessary ¯\_(ツ)_/¯.

	// NOTE: Not checking refs for errors since they may still be valid even if they have errors. Instead, we just validate the metrics view against the table name.

	// Validate the metrics view and update ValidSpec
	e, err := metricsview.NewExecutor(ctx, r.C.Runtime, r.C.InstanceID, mv.Spec, runtime.ResolvedSecurityOpen, 0)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to create metrics view executor: %w", err)}
	}
	defer e.Close()
	validateResult, validateErr := e.ValidateMetricsView(ctx)
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

	// Set the "streaming" state (see docstring in the proto for details).
	mv.State.Streaming = false
	if validateErr == nil {
		// Find out if the metrics view has a ref to a source or model in the same project.
		hasInternalRef := false
		for _, ref := range self.Meta.Refs {
			if ref.Kind == runtime.ResourceKindSource || ref.Kind == runtime.ResourceKindModel {
				hasInternalRef = true
			}
		}

		// If not, we assume the metrics view is based on an externally managed table and set the streaming state to true.
		mv.State.Streaming = !hasInternalRef
	}

	// set cache controls
	olap, release, err := r.C.AcquireOLAP(ctx, mv.Spec.Connector)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to get OLAP: %w", err)}
	}
	defer release()

	// Update state. Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	// Override the cache control with our calculated cache control
	mv.State.ValidSpec.Cache = metricsViewCacheControl(mv.Spec, mv.State.Streaming, self.Meta.StateUpdatedOn, olap.Dialect())

	return runtime.ReconcileResult{Err: validateErr}
}

func metricsViewCacheControl(spec *runtimev1.MetricsViewSpec, streaming bool, updatedOn *timestamppb.Timestamp, dialect drivers.Dialect) *runtimev1.MetricsViewSpec_Cache {
	var enabled *bool
	if spec.Cache != nil && spec.Cache.Enabled != nil {
		enabled = spec.Cache.Enabled
	} else {
		enabled = boolPtr(!streaming)
	}
	cache := &runtimev1.MetricsViewSpec_Cache{
		Enabled: enabled,
	}
	if spec.Cache != nil && spec.Cache.KeyTtlSeconds != 0 {
		cache.KeyTtlSeconds = spec.Cache.KeyTtlSeconds
	} else if streaming {
		cache.KeyTtlSeconds = 60
	}

	if spec.Cache != nil && spec.Cache.KeySql != "" {
		cache.KeySql = spec.Cache.KeySql
	} else {
		if streaming {
			cache.KeySql = fmt.Sprintf("SELECT %s FROM %s", spec.WatermarkExpression, dialect.EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
		} else {
			cache.KeySql = "SELECT " + fmt.Sprintf("'%d:%d'", updatedOn.GetSeconds(), updatedOn.GetNanos()/int32(time.Millisecond))
		}
	}
	return cache
}

func boolPtr(b bool) *bool {
	return &b
}
