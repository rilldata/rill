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

	validateErr := r.validate(ctx, mv.Spec)

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

func (r *MetricsViewReconciler) validate(ctx context.Context, mv *runtimev1.MetricsViewSpec) error {
	olap, release, err := r.C.AcquireOLAP(ctx, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	// Check underlying table exists
	t, err := olap.InformationSchema().Lookup(ctx, mv.Table)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return fmt.Errorf("table %q does not exist", mv.Table)
		}
		return fmt.Errorf("could not find table %q: %w", mv.Table, err)
	}

	fields := make(map[string]*runtimev1.StructType_Field, len(t.Schema.Fields))
	for _, f := range t.Schema.Fields {
		fields[strings.ToLower(f.Name)] = f
	}

	// Check time dimension exists
	if mv.TimeDimension != "" {
		f, ok := fields[strings.ToLower(mv.TimeDimension)]
		if !ok {
			return fmt.Errorf("timeseries %q is not a column in table %q", mv.TimeDimension, mv.Table)
		}
		if f.Type.Code != runtimev1.Type_CODE_TIMESTAMP && f.Type.Code != runtimev1.Type_CODE_DATE {
			return fmt.Errorf("timeseries %q is not a TIMESTAMP column", mv.TimeDimension)
		}
	}

	var errs []error

	// Check dimension columns exist
	for _, d := range mv.Dimensions {
		err = validateDimension(ctx, olap, t, d, fields)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Check measure expressions are valid
	for _, d := range mv.Measures {
		err := validateMeasure(ctx, olap, t, d)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid expression for measure %q: %w", d.Name, err))
		}
	}

	if mv.DefaultTheme != "" {
		_, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: mv.DefaultTheme}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrNotFound) {
				return fmt.Errorf("theme %q does not exist", mv.DefaultTheme)
			}
			return fmt.Errorf("could not find theme %q: %w", mv.DefaultTheme, err)
		}
	}

	return errors.Join(errs...)
}

func validateDimension(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, d *runtimev1.MetricsViewSpec_DimensionV2, fields map[string]*runtimev1.StructType_Field) error {
	if d.Expression == "" {
		if d.Column == "" {
			return nil
		}
		if _, isColumn := fields[d.Column]; !isColumn {
			return fmt.Errorf("failed to validate dimension %q: column %q not found in table", d.Name, d.Column)
		}
		return nil
	}

	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s FROM %s GROUP BY 1", d.Expression, safeSQLName(t.Name)),
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate expression for dimension %q: %w", d.Name, err)
	}
	return nil
}

func validateMeasure(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s from %s", m.Expression, safeSQLName(t.Name)),
		DryRun: true,
	})
	return err
}
