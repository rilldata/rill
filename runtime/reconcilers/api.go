package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindAPI, newAPIReconciler)
}

type APIReconciler struct {
	C *runtime.Controller
}

func newAPIReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &APIReconciler{C: c}, nil
}

func (r *APIReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *APIReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetApi()
	b := to.GetApi()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *APIReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetApi()
	b := to.GetApi()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *APIReconciler) ResetState(res *runtimev1.Resource) error {
	return nil
}

func (r *APIReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetApi()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not an API")}
	}

	// TODO: Add validation of the resolver properties when the resolver abstractions are implemented

	return runtime.ReconcileResult{}
}

func (r *APIReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetApi() == nil {
		return nil, fmt.Errorf("not an API resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}
