package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindConnector, newConnectorReconciler)
}

type ConnectorReconciler struct {
	C *runtime.Controller
}

func newConnectorReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ConnectorReconciler{C: c}, nil
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
		// Drop a corresponding ClickHouse named collection if applicable. We do this BEFORE
		// removing the source connector from the instance because the drop only depends on the
		// OLAP (ClickHouse) connector, not on the source connector — but ordering here keeps the
		// failure mode obvious if the OLAP connection has issues.
		if err := r.syncClickHouseNamedCollection(ctx, t.Spec.Driver, self.Meta.Name.Name, true); err != nil {
			return runtime.ReconcileResult{Err: err}
		}
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
		// The connector configuration should be tested even when spec has not changed since connector errors can be transient (e.g. cluster temporarily down)
		err := r.testConnector(ctx, self.Meta.Name.Name)
		return runtime.ReconcileResult{Err: err}
	}

	// Update instance connectors
	err = r.C.Runtime.UpdateInstanceConnector(ctx, r.C.InstanceID, self.Meta.Name.Name, t.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Test the connector configuration
	err = r.testConnector(ctx, self.Meta.Name.Name)
	// update state even if test fails because the instance connectors have already been updated
	t.State.SpecHash = specHash

	// Sync the ClickHouse named collection for this connector if applicable. This runs only when
	// the project's OLAP engine is ClickHouse and the source connector driver is one of the
	// supported types (s3, gcs, azure, mysql, postgres). Failures here are surfaced as reconcile
	// errors so the user gets a clear signal (e.g. missing `named_collection_admin`).
	if err == nil {
		err = r.syncClickHouseNamedCollection(ctx, t.Spec.Driver, self.Meta.Name.Name, false)
	}

	err = errors.Join(err, r.C.UpdateState(ctx, self.Meta.Name, self))
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}

func (r *ConnectorReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetConnector() == nil {
		return nil, fmt.Errorf("not a connector resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

func (r *ConnectorReconciler) executionSpecHash(ctx context.Context, spec *runtimev1.ConnectorSpec) (string, error) {
	hash := md5.New()

	_, err := hash.Write([]byte(spec.Driver))
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.Provision)
	if err != nil {
		return "", err
	}

	if spec.ProvisionArgs != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.ProvisionArgs), hash)
		if err != nil {
			return "", err
		}
	}
	if spec.Properties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.Properties), hash)
		if err != nil {
			return "", err
		}
		res, err := analyzeTemplatedVariables(ctx, r.C, spec.Properties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (r *ConnectorReconciler) testConnector(ctx context.Context, connectorName string) error {
	// Get the connector configuration
	handle, release, err := r.C.Runtime.AcquireHandle(ctx, r.C.InstanceID, connectorName)
	if err != nil {
		return err
	}
	defer release()

	return handle.Ping(ctx)
}

// syncClickHouseNamedCollection creates or drops a ClickHouse named collection that mirrors the
// given connector's credentials, when the project's OLAP engine is ClickHouse and the connector
// driver is one of the supported types (s3, gcs, azure, mysql, postgres).
//
// For drivers that don't qualify, this is a no-op. For projects whose OLAP is not ClickHouse,
// this is also a no-op — DuckDB-as-OLAP projects use TEMPORARY SECRETS during model execution
// instead.
//
// If drop is true, the collection is dropped (used during connector deletion). Otherwise it is
// created or replaced.
func (r *ConnectorReconciler) syncClickHouseNamedCollection(ctx context.Context, sourceDriver, connectorName string, drop bool) error {
	if !clickhouse.IsSupportedNamedCollectionDriver(sourceDriver) {
		return nil
	}

	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to load instance: %w", err)
	}
	olapName := inst.ResolveOLAPConnector()
	if olapName == "" {
		return nil
	}

	// Acquire the OLAP connector handle. If it's not ClickHouse, this is a no-op.
	olapHandle, release, err := r.C.Runtime.AcquireHandle(ctx, r.C.InstanceID, olapName)
	if err != nil {
		return fmt.Errorf("failed to acquire OLAP connector %q: %w", olapName, err)
	}
	defer release()
	if olapHandle.Driver() != "clickhouse" {
		return nil
	}
	chConn, ok := olapHandle.(*clickhouse.Connection)
	if !ok {
		// Defensive: the only registered driver under "clickhouse" is *clickhouse.Connection.
		return fmt.Errorf("internal: OLAP connector %q has driver=clickhouse but unexpected handle type %T", olapName, olapHandle)
	}

	if drop {
		return chConn.DropNamedCollection(ctx, connectorName)
	}

	// Verify the user has named-collection-admin rights up front, so we surface a clear error
	// at connector reconcile rather than a confusing failure later during model resolution.
	if err := chConn.CheckNamedCollectionAdmin(ctx); err != nil {
		return err
	}

	// Resolve the source connector's config, then build the named-collection params.
	cfg, err := r.C.Runtime.ConnectorConfig(ctx, r.C.InstanceID, connectorName)
	if err != nil {
		return fmt.Errorf("failed to load connector config: %w", err)
	}
	params, err := clickhouse.BuildNamedCollectionParams(sourceDriver, cfg.Resolve())
	if err != nil {
		// GCS-with-native-creds and other "no usable creds" cases are expected for some users;
		// we surface them as warnings via the reconcile error chain rather than silent skips so
		// the user can act on them.
		return fmt.Errorf("connector %q: %w", connectorName, err)
	}
	return chConn.CreateOrReplaceNamedCollection(ctx, connectorName, params)
}
