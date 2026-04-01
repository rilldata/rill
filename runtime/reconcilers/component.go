package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindComponent, newComponentReconciler)
}

type ComponentReconciler struct {
	C *runtime.Controller
}

func newComponentReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ComponentReconciler{C: c}, nil
}

func (r *ComponentReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ComponentReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetComponent()
	b := to.GetComponent()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ComponentReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetComponent()
	b := to.GetComponent()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ComponentReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetComponent().State = &runtimev1.ComponentState{}
	return nil
}

func (r *ComponentReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	c := self.GetComponent()
	if c == nil {
		return runtime.ReconcileResult{Err: errors.New("not a component")}
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

	// Validate all refs
	validateErr := checkRefs(ctx, r.C, self.Meta.Refs)

	// Get valid metrics view refs.
	// NOTE: The validateErr may be non-nil if a metrics view has a reconcile error, but the same metrics view may still be returned here if its ValidSpec is non-nil (e.g. it might be serving previously valid state).
	mvs, allMetricsValid, dataRefreshedOn, err := r.referencedMetricsViews(ctx, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Validate the renderer properties (only if all metrics view refs have a ValidSpec).
	var rendererErr error
	if allMetricsValid {
		rendererErr = r.validateRendererProperties(c.Spec.Renderer, c.Spec.RendererProperties.AsMap(), mvs)
	} else {
		rendererErr = errors.New("one or more referenced metrics views are invalid")
	}
	if validateErr == nil { // Gives precedence to refs errors over renderer errors, since the ref error may have caused the renderer error.
		validateErr = rendererErr
	}

	// Update the state according to the validation result.
	// Even if the validation result is unchanged, we always update the state to ensure the state version is incremented.
	// When StageChanges is enabled, we want to make a best effort to serve the canvas anyway.
	// Specifically, if the renderer properties are valid (which also implies the metrics view(s) referenced by the component have a ValidSpec), we'll serve a ValidSpec (but still return the validation error so it gets surfaced).
	if validateErr == nil || (cfg.StageChanges && rendererErr == nil) {
		err = r.updateState(ctx, self, c.Spec, dataRefreshedOn, validateErr)
		return runtime.ReconcileResult{Err: err}
	}

	// Validation failed and we can't serve valid state.
	// So we clear out the ValidSpec.
	err = r.updateState(ctx, self, nil, nil, validateErr)
	return runtime.ReconcileResult{Err: err}
}

func (r *ComponentReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetComponent() == nil {
		return nil, fmt.Errorf("not a component resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// updateState is a helper for updating a component's state.
// If an error is provided, it will be returned after the state update, allowing simple returns.
func (r *ComponentReconciler) updateState(ctx context.Context, self *runtimev1.Resource, validSpec *runtimev1.ComponentSpec, dataRefreshedOn *timestamppb.Timestamp, basedOnErr error) error {
	// Don't update state for ctx errors.
	if basedOnErr != nil && errors.Is(basedOnErr, ctx.Err()) {
		return basedOnErr
	}

	// Update the state.
	c := self.GetComponent()
	c.State.ValidSpec = validSpec
	c.State.DataRefreshedOn = dataRefreshedOn
	err := r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return err
	}

	// Return the original error (if any) as per the docstring.
	return basedOnErr
}

// referencedMetricsViews returns the valid metrics view specs for the given refs. If any referenced metrics view is invalid, it is not included in the result, and the returned boolean will be false.
// It only returns an error if there was an issue checking the refs, not if a ref was simply invalid.
func (r *ComponentReconciler) referencedMetricsViews(ctx context.Context, refs []*runtimev1.ResourceName) (map[string]*runtimev1.MetricsViewSpec, bool, *timestamppb.Timestamp, error) {
	mvs := make(map[string]*runtimev1.MetricsViewSpec)
	allMetricsValid := true
	var dataRefreshedOn *timestamppb.Timestamp
	for _, ref := range refs {
		if ref.Kind != runtime.ResourceKindMetricsView {
			continue
		}

		res, err := r.C.Get(ctx, ref, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				allMetricsValid = false
				continue
			}
			return nil, false, nil, err
		}

		mv := res.GetMetricsView()
		if mv.State.ValidSpec != nil {
			mvs[ref.Name] = mv.State.ValidSpec
		} else {
			allMetricsValid = false
		}

		t := res.GetMetricsView().State.DataRefreshedOn
		if dataRefreshedOn == nil {
			dataRefreshedOn = t
		} else if t != nil && t.AsTime().After(dataRefreshedOn.AsTime()) {
			dataRefreshedOn = t
		}
	}
	return mvs, allMetricsValid, dataRefreshedOn, nil
}

// validateRendererProperties validates the renderer properties for a component.
// The provided metricsViews contains every valid metrics view referenced by the component (as determined in the parser).
// If the renderer properties reference a metrics view not in metricsViews, assume the metrics view is invalid or does not exist (don't look it up separately in the catalog).
// Note that metrics views referenced through markdown or metrics_sql cannot be validated here, and using rendererRefs to find them is not safe (because the parser doesn't add refs to them, so they may not have reconciled yet).
func (r *ComponentReconciler) validateRendererProperties(renderer string, props map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	switch renderer {
	case "line_chart":
		mvn, ok := pathutil.GetPathString(props, "metrics_view")
		if !ok {
			return errors.New("renderer properties must include a string 'metrics_view' property")
		}
		mv := metricsViews[mvn]
		if mv == nil {
			return fmt.Errorf("referenced metrics view %q is invalid", mvn)
		}

		xField, ok := pathutil.GetPathString(props, "x.field")
		if !ok {
			return errors.New("renderer properties for line_chart must include a string 'x.field' property")
		}
		if !metricsViewHasDimension(mv, xField) {
			return fmt.Errorf("referenced x.field %q is not a dimension in metrics view %q", xField, mvn)
		}

		yField, ok := pathutil.GetPathString(props, "y.field")
		if !ok {
			return errors.New("renderer properties for line_chart must include a string 'y.field' property")
		}
		if !metricsViewHasMeasure(mv, yField) {
			return fmt.Errorf("referenced y.field %q is not a measure in metrics view %q", yField, mvn)
		}

		// TODO: Any other validation for line charts?
	case "stacked_bar":
		// TODO: Implement
	case "bar_chart":
		// TODO: Implement
	case "stacked_bar_normalized":
		// TODO: Implement
	case "area_chart":
		// TODO: Implement
	case "donut_chart":
		// TODO: Implement
	case "pie_chart":
		// TODO: Implement
	case "heatmap":
		// TODO: Implement
	case "funnel_chart":
		// TODO: Implement
	case "combo_chart":
		// TODO: Implement
	case "scatter_plot":
		// TODO: Implement
	case "markdown":
		// TODO: Implement
	case "kpi":
		// TODO: Implement
	case "kpi_grid":
		// TODO: Implement
	case "image":
		// TODO: Implement
	case "table":
		// TODO: Implement
	case "pivot":
		// TODO: Implement
	case "leaderboard":
		// TODO: Implement
	default:
		return fmt.Errorf("unsupported renderer %q", renderer)
	}

	return nil
}

// metricsViewHasDimension returns true if the metrics view has a dimension with the given name.
func metricsViewHasDimension(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, d := range mv.Dimensions {
		if d.Name == fieldName {
			return true
		}
	}
	return false
}

// metricsViewHasMeasure returns true if the metrics view has a measure with the given name.
func metricsViewHasMeasure(mv *runtimev1.MetricsViewSpec, fieldName string) bool {
	for _, m := range mv.Measures {
		if m.Name == fieldName {
			return true
		}
	}
	return false
}
