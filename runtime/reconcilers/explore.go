package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/fieldselectorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindExplore, newExploreReconciler)
}

type ExploreReconciler struct {
	C *runtime.Controller
}

func newExploreReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ExploreReconciler{C: c}, nil
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

	// Get instance config
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Validate and rewrite
	validSpec, dataRefreshedOn, validateErr := r.validateAndRewrite(ctx, self, e.Spec)

	// If spec validation failed and StageChanges is enabled, we will keep the old valid spec if its parent metrics view is still valid.
	// This is not perfect, but increases the chance of keeping the dashboard working in many cases.
	if validSpec == nil && cfg.StageChanges && e.State.ValidSpec != nil {
		// Get the metrics view referenced by the old valid spec.
		mvn := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: e.State.ValidSpec.MetricsView}
		mv, err := r.C.Get(ctx, mvn, false)
		if err == nil && mv.GetMetricsView().State.ValidSpec != nil {
			// Keep the old valid spec
			validSpec = e.State.ValidSpec
			dataRefreshedOn = mv.GetMetricsView().State.DataRefreshedOn
		}
	}

	// We update the state even if the validation result is unchanged to ensure the state version is incremented.
	e.State.ValidSpec = validSpec
	e.State.DataRefreshedOn = dataRefreshedOn
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: validateErr}
}

func (r *ExploreReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName

	explore := res.GetExplore()
	if explore == nil {
		return nil, fmt.Errorf("resource is not an explore")
	}

	spec := explore.GetState().GetValidSpec()
	if spec == nil {
		return nil, fmt.Errorf("explore valid spec is nil")
	}

	if spec.MetricsView == "" {
		return nil, fmt.Errorf("explore does not reference a metrics view")
	}

	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, runtime.ResourceKindTheme)

	// give access to the underlying metrics view
	if spec.MetricsView != "" {
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: spec.MetricsView})
	}

	if len(conditionKinds) > 0 || len(conditionResources) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionKinds:     conditionKinds,
					ConditionResources: conditionResources,
					Allow:              true,
					Exclusive:          true,
				},
			},
		})
	}

	return rules, nil
}

// validateAndRewrite validates the explore spec and rewrites it with the following rules:
//   - The dimensions_exclude and measures_exclude flags will be resolved using the parent metrics view's fields, and set to false.
//   - The parent metrics view's access and field access security rules will be copied to the explore spec's security rules.
//
// The provided spec will be modified in place, so it must be a deep clone.
func (r *ExploreReconciler) validateAndRewrite(ctx context.Context, self *runtimev1.Resource, spec *runtimev1.ExploreSpec) (*runtimev1.ExploreSpec, *timestamppb.Timestamp, error) {
	err := checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return nil, nil, err
	}

	// Check the theme exists
	if spec.Theme != "" {
		_, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: spec.Theme}, false)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find theme %q: %w", spec.Theme, err)
		}
	}

	// Get the parent metrics view's valid spec
	mvn := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: spec.MetricsView}
	mvr, err := r.C.Get(ctx, mvn, false)
	if err != nil {
		return nil, nil, fmt.Errorf("could not find metrics view %q: %w", spec.MetricsView, err)
	}
	mv := mvr.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, nil, fmt.Errorf("parent metrics view %q is invalid", spec.MetricsView)
	}

	if len(spec.SecurityRules) == 0 && len(mv.SecurityRules) > 0 {
		for _, rule := range mv.SecurityRules {
			if rule.GetAccess() != nil || rule.GetFieldAccess() != nil {
				spec.SecurityRules = append(spec.SecurityRules, rule)
			}
		}
	} else {
		for _, rule := range spec.SecurityRules {
			if rule.GetAccess() == nil {
				return nil, nil, fmt.Errorf("security rule %v is not an access rule", rule)
			}
		}

		// Merge access rules into a single rule
		access := mergeAccessRules(slices.Concat(mv.SecurityRules, spec.SecurityRules))
		if access != nil {
			spec.SecurityRules = []*runtimev1.SecurityRule{access}
		}

		// Copy field access rules
		for _, rule := range mv.SecurityRules {
			if rule.GetFieldAccess() != nil {
				spec.SecurityRules = append(spec.SecurityRules, rule)
			}
		}
	}

	// Validate and rewrite dimensions
	timeDims := make(map[string]struct{})
	allDims := make([]string, 0, len(mv.Dimensions))
	for _, d := range mv.Dimensions {
		if d.Type == runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME {
			timeDims[d.Name] = struct{}{}
		}
		allDims = append(allDims, d.Name)
	}
	spec.Dimensions, err = fieldselectorpb.ResolveFields(spec.Dimensions, spec.DimensionsSelector, allDims)
	if err != nil {
		return nil, nil, err
	}
	spec.DimensionsSelector = nil

	// Validate and rewrite measures
	allMeasures := make([]string, 0, len(mv.Measures))
	for _, m := range mv.Measures {
		allMeasures = append(allMeasures, m.Name)
	}
	spec.Measures, err = fieldselectorpb.ResolveFields(spec.Measures, spec.MeasuresSelector, allMeasures)
	if err != nil {
		return nil, nil, err
	}
	spec.MeasuresSelector = nil

	// Validate and rewrite presets, now in the context of the explore's dimensions and measures resolved above.
	if spec.DefaultPreset != nil {
		p := spec.DefaultPreset

		dims, err := fieldselectorpb.ResolveFields(p.Dimensions, p.DimensionsSelector, spec.Dimensions)
		if err != nil {
			return nil, nil, err
		}
		p.Dimensions = dims
		p.DimensionsSelector = nil

		measures, err := fieldselectorpb.ResolveFields(p.Measures, p.MeasuresSelector, spec.Measures)
		if err != nil {
			return nil, nil, err
		}
		p.Measures = measures
		p.MeasuresSelector = nil
	}

	// Filter out all time dimensions from the explore's spec.
	// TODO: Remove when the UI supports multiple time dimensions.
	spec.Dimensions = slices.DeleteFunc(spec.Dimensions, func(v string) bool {
		_, isTime := timeDims[v]
		return isTime
	})
	if spec.DefaultPreset != nil {
		spec.DefaultPreset.Dimensions = slices.DeleteFunc(spec.DefaultPreset.Dimensions, func(v string) bool {
			_, isTime := timeDims[v]
			return isTime
		})
	}

	// Done with rewriting
	return spec, mvr.GetMetricsView().State.DataRefreshedOn, nil
}

// mergeAccessRules combines Access rule conditions into a single rule
func mergeAccessRules(rules []*runtimev1.SecurityRule) *runtimev1.SecurityRule {
	ruleCount := len(rules)
	// If there are no rules, return nil
	if ruleCount == 0 {
		return nil
	}

	// If there is only one rule, return it
	if ruleCount == 1 {
		return rules[0]
	}

	// If there are multiple rules, merge their conditions into a single condition with AND operator
	var condition strings.Builder
	for i, rule := range rules {
		access := rule.GetAccess()
		if access == nil {
			// Skip rules without Access field or log an error
			continue
		}

		if i > 0 && condition.Len() > 0 {
			condition.WriteString(" AND ")
		}
		condition.WriteString("(")
		condition.WriteString(access.ConditionExpression)
		condition.WriteString(")")
	}

	// If no valid conditions were found
	if condition.Len() == 0 {
		return nil
	}

	return &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				ConditionExpression: condition.String(),
				Allow:               true,
			},
		},
	}
}
