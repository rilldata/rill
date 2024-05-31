package reconcilers

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindConnector, newConnectorReconciler)
}

type ConnectorReconciler struct {
	C *runtime.Controller
}

func newConnectorReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ConnectorReconciler{C: c}
}

func (r *ConnectorReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ConnectorReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetConnector()
	b := to.GetConnector()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ConnectorReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetConnector()
	b := to.GetConnector()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ConnectorReconciler) ResetState(res *runtimev1.Resource) error {
	return nil
}

func (r *ConnectorReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetConnector()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a connector")}
	}

	// Get and sync repo
	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to access repo: %w", err)}
	}
	defer release()
	err = repo.Sync(ctx)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to sync repo: %w", err)}
	}

	// Get instance
	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to find instance: %w", err)}
	}

	// Parse the project
	parser, err := rillv1.Parse(ctx, repo, r.C.InstanceID, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to parse: %w", err)}
	}

	// Update instance connectors
	err = r.C.Runtime.UpdateInstanceWithRillYAML(ctx, inst.ID, parser, false)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to update instance: %w", err)}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	return runtime.ReconcileResult{}
}
