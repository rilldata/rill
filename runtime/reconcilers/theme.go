package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindTheme, newThemeReconciler)
}

type ThemeReconciler struct {
	C *runtime.Controller
}

func newThemeReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ThemeReconciler{C: c}, nil
}

func (r *ThemeReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ThemeReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetTheme()
	b := to.GetTheme()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ThemeReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetTheme()
	b := to.GetTheme()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ThemeReconciler) ResetState(res *runtimev1.Resource) error {
	return nil
}

func (r *ThemeReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetTheme()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a theme")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	return runtime.ReconcileResult{}
}

func (r *ThemeReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetTheme() == nil {
		return nil, fmt.Errorf("not a theme resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}
