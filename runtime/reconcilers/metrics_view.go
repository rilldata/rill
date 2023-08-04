package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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

func (r *MetricsViewReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	mv := self.GetMetricsView()

	validateErr := r.validate(ctx, mv.Spec)

	if errors.Is(validateErr, ctx.Err()) {
		return runtime.ReconcileResult{Err: validateErr}
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

func (r *MetricsViewReconciler) validate(ctx context.Context, mv *runtimev1.MetricsViewSpec) error {
	olap, release, err := r.C.AcquireOLAP(ctx, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	// Check underlying table exists
	t, err := olap.InformationSchema().Lookup(ctx, mv.Table)
	if err != nil {
		return fmt.Errorf("table %q does not exist", mv.Table)
	}

	fields := make(map[string]*runtimev1.StructType_Field, len(t.Schema.Fields))
	for _, f := range t.Schema.Fields {
		fields[strings.ToLower(f.Name)] = f
	}

	// Check time dimension exists
	if mv.TimeDimension != "" {
		_, ok := fields[strings.ToLower(mv.TimeDimension)]
		if !ok {
			return fmt.Errorf("timeseries %q is not a column in table %q", mv.TimeDimension, mv.Table)
		}
	}

	var errs []error

	// Check dimension columns exist
	for _, d := range mv.Dimensions {
		if _, ok := fields[strings.ToLower(d.Column)]; !ok {
			errs = append(errs, fmt.Errorf("dimension column %q not found in table %q", d.Column, mv.Table))
		}
	}

	// Check measure expressions are valid
	for _, d := range mv.Measures {
		err := validateMeasure(ctx, olap, t, d)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid expression for measure %q: %w", d.Name, err))
		}
	}

	return errors.Join(errs...)
}

func validateMeasure(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s from %s", m.Expression, safeSQLName(t.Name)),
		DryRun: true,
	})
	return err
}
