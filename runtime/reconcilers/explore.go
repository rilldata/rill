package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"golang.org/x/exp/maps"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindExplore, newExploreReconciler)
}

type ExploreReconciler struct {
	C *runtime.Controller
}

func newExploreReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ExploreReconciler{C: c}
}

func (r *ExploreReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ExploreReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetExplore()
	b := to.GetExplore()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ExploreReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetExplore()
	b := to.GetExplore()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ExploreReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetExplore().State = &runtimev1.ExploreState{}
	return nil
}

func (r *ExploreReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	e := self.GetExplore()
	if e == nil {
		return runtime.ReconcileResult{Err: errors.New("not an explore")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Validate and rewrite
	validSpec, validateErr := r.validateAndRewrite(ctx, self, e.Spec)

	// Always capture the valid spec in the state, even if validation failed and it is nil.
	// We update the state even if the validation result is unchanged to ensure the state version is incremented.
	e.State.ValidSpec = validSpec
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}

// validateAndRewrite validates the explore spec and rewrites it with the following rules:
//   - The dimensions_exclude and measures_exclude flags will be resolved using the parent metrics view's fields, and set to false.
//   - The parent metrics view's access and field access security rules will be copied to the explore spec's security rules.
//
// The provided spec will be modified in place, so it must be a deep clone.
func (r *ExploreReconciler) validateAndRewrite(ctx context.Context, self *runtimev1.Resource, spec *runtimev1.ExploreSpec) (*runtimev1.ExploreSpec, error) {
	err := checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return nil, err
	}

	// Check the theme exists
	if spec.Theme != "" {
		_, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: spec.Theme}, false)
		if err != nil {
			return nil, fmt.Errorf("failed to find theme %q: %w", spec.Theme, err)
		}
	}

	// Get the parent metrics view's valid spec
	mvn := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: spec.MetricsView}
	mvr, err := r.C.Get(ctx, mvn, false)
	if err != nil {
		return nil, fmt.Errorf("could not find metrics view %q: %w", spec.MetricsView, err)
	}
	mv := mvr.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("parent metrics view %q is invalid", spec.MetricsView)
	}

	// Add the access and field access security rules from the parent metrics view.
	for _, rule := range mv.SecurityRules {
		if rule.GetAccess() != nil || rule.GetFieldAccess() != nil {
			spec.SecurityRules = append(spec.SecurityRules, rule)
		}
	}
	spec.SecurityRules = append(spec.SecurityRules, mv.SecurityRules...)

	// Validate and rewrite dimensions
	allDims := make(map[string]bool)
	for _, d := range mv.Dimensions {
		allDims[d.Name] = true
	}
	dims, err := r.resolveNames(allDims, spec.Dimensions, spec.DimensionsExclude)
	if err != nil {
		return nil, err
	}
	spec.Dimensions = dims
	spec.DimensionsExclude = false

	// Validate and rewrite measures
	allMeasures := make(map[string]bool)
	for _, m := range mv.Measures {
		allMeasures[m.Name] = true
	}
	measures, err := r.resolveNames(allMeasures, spec.Measures, spec.MeasuresExclude)
	if err != nil {
		return nil, err
	}
	spec.Measures = measures
	spec.MeasuresExclude = false

	// Validate and rewrite presets, now in the context of the explore's dimensions and measures resolved above.
	if len(spec.Presets) > 0 {
		allDims = make(map[string]bool)
		for _, d := range spec.Dimensions {
			allDims[d] = true
		}

		allMeasures := make(map[string]bool)
		for _, m := range spec.Measures {
			allMeasures[m] = true
		}

		for _, p := range spec.Presets {
			dims, err = r.resolveNames(allDims, p.Dimensions, p.DimensionsExclude)
			if err != nil {
				return nil, err
			}
			p.Dimensions = dims
			p.DimensionsExclude = false

			measures, err := r.resolveNames(allMeasures, p.Measures, p.MeasuresExclude)
			if err != nil {
				return nil, err
			}
			p.Measures = measures
			p.MeasuresExclude = false
		}
	}

	// Done with rewriting
	return spec, nil
}

func (r *ExploreReconciler) resolveNames(allNames map[string]bool, names []string, exclude bool) ([]string, error) {
	// Optimization for the wildcard case (exclude nothing = include everything)
	if len(names) == 0 && exclude {
		return maps.Keys(allNames), nil
	}

	// Check the provided names exist
	for _, n := range names {
		if !allNames[n] {
			return nil, fmt.Errorf("dimension or measure name %q not found in the parent metrics view", n)
		}
	}

	// If not excluding, return the provided names as is
	if !exclude {
		return names, nil
	}

	// Get all names not in the provided names
	var res []string
	for n := range allNames {
		if !slices.Contains(names, n) {
			res = append(res, n)
		}
	}
	return res, nil
}
