package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindSkill, newSkillReconciler)
}

type SkillReconciler struct {
	C *runtime.Controller
}

func newSkillReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &SkillReconciler{C: c}, nil
}

func (r *SkillReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *SkillReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetSkill()
	b := to.GetSkill()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *SkillReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetSkill()
	b := to.GetSkill()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *SkillReconciler) ResetState(res *runtimev1.Resource) error {
	return nil
}

func (r *SkillReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	s := self.GetSkill()
	if s == nil {
		return runtime.ReconcileResult{Err: errors.New("not a skill")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Check that the metrics views the skill is scoped to exist.
	// We deliberately don't use checkRefs here: scoping is a relevance filter, not a hard dependency,
	// so a skill should stay valid while a referenced metrics view is refreshing or has a transient error.
	for _, mv := range s.Spec.MetricsViews {
		_, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: mv}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return runtime.ReconcileResult{Err: fmt.Errorf("metrics view %q referenced in metrics_views not found", mv)}
			}
			return runtime.ReconcileResult{Err: err}
		}
	}

	return runtime.ReconcileResult{}
}

func (r *SkillReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetSkill() == nil {
		return nil, fmt.Errorf("not a skill resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}
