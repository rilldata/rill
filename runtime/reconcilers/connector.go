package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
	res.GetConnector().State = &runtimev1.ConnectorState{}
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

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		err = r.C.Runtime.UpdateInstanceConnector(ctx, r.C.InstanceID, self.Meta.Name.Name, nil)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		return runtime.ReconcileResult{}
	}

	// Check if the spec has changed
	specHash, err := r.executionSpecHash(ctx, t.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	if specHash == t.State.SpecHash {
		return runtime.ReconcileResult{}
	}

	// Update instance connectors
	err = r.C.Runtime.UpdateInstanceConnector(ctx, r.C.InstanceID, self.Meta.Name.Name, t.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	t.State.SpecHash = specHash

	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}

func (r *ConnectorReconciler) executionSpecHash(ctx context.Context, spec *runtimev1.ConnectorSpec) (string, error) {
	instance, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return "", err
	}
	vars := instance.ResolveVariables(false)

	hash := md5.New()

	_, err = hash.Write([]byte(spec.Driver))
	if err != nil {
		return "", err
	}

	// sort properties by key
	keys := make([]string, 0, len(spec.Properties))
	for k := range spec.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// write properties to hash
	for _, k := range keys {
		_, err = hash.Write([]byte(k))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(spec.Properties[k]))
		if err != nil {
			return "", err
		}
	}

	// sort propertiesFromVariables by key
	keys = make([]string, 0, len(spec.PropertiesFromVariables))
	for k := range spec.PropertiesFromVariables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// write propertiesFromVariables and corresponding vars to hash
	for _, k := range keys {
		_, err = hash.Write([]byte(k))
		if err != nil {
			return "", err
		}
		name := spec.PropertiesFromVariables[k]
		_, err = hash.Write([]byte(name))
		if err != nil {
			return "", err
		}
		if value, ok := vars[name]; ok {
			_, err = hash.Write([]byte(value))
			if err != nil {
				return "", err
			}
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
