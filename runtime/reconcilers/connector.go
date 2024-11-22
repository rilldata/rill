package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
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

	err = binary.Write(hash, binary.BigEndian, spec.Provision)
	if err != nil {
		return "", err
	}

	err = pbutil.WriteHash(structpb.NewStructValue(spec.ProvisionArgs), hash)
	if err != nil {
		return "", err
	}

	// write properties to hash
	props, err := runtime.ResolveConnectorProperties(instance.Environment, vars, &runtimev1.Connector{
		Type:                spec.Driver,
		Config:              spec.Properties,
		TemplatedProperties: spec.TemplatedProperties,
		Provision:           spec.Provision,
		ProvisionArgs:       spec.ProvisionArgs,
		ConfigFromVariables: spec.PropertiesFromVariables,
	})
	if err != nil {
		return "", err
	}
	keys := maps.Keys(props)
	slices.Sort(keys)
	for _, k := range keys {
		_, err = hash.Write([]byte(k))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(props[k]))
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
