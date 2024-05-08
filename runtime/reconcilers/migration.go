package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindMigration, newMigrationReconciler)
}

type MigrationReconciler struct {
	C *runtime.Controller
}

func newMigrationReconciler(c *runtime.Controller) runtime.Reconciler {
	return &MigrationReconciler{C: c}
}

func (r *MigrationReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *MigrationReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetMigration()
	b := to.GetMigration()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *MigrationReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetMigration()
	b := to.GetMigration()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *MigrationReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetMigration().State = &runtimev1.MigrationState{}
	return nil
}

func (r *MigrationReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	mig := self.GetMigration()
	if mig == nil {
		return runtime.ReconcileResult{Err: errors.New("not a migration")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Check refs - stop if any of them are invalid
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	from := mig.State.Version
	to := mig.Spec.Version

	if to-from > 100 {
		return runtime.ReconcileResult{Err: fmt.Errorf("difference between migration versions %d and %d is too large", from, to)}
	}

	for v := from; v <= to; v++ {
		err := r.executeMigration(ctx, self, v)
		if err != nil {
			if v != 0 {
				err = fmt.Errorf("failed to execute version %d: %w", v, err)
			}
			return runtime.ReconcileResult{Err: err}
		}

		mig.State.Version = v
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	return runtime.ReconcileResult{}
}

func (r *MigrationReconciler) executeMigration(ctx context.Context, self *runtimev1.Resource, version uint32) error {
	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}

	spec := self.Resource.(*runtimev1.Resource_Migration).Migration.Spec
	state := self.Resource.(*runtimev1.Resource_Migration).Migration.State

	olap, release, err := r.C.AcquireOLAP(ctx, spec.Connector)
	if err != nil {
		return err
	}
	defer release()

	sql, err := compilerv1.ResolveTemplate(spec.Sql, compilerv1.TemplateData{
		Environment: inst.Environment,
		User:        map[string]interface{}{},
		Variables:   inst.ResolveVariables(),
		ExtraProps: map[string]interface{}{
			"version": version,
		},
		Self: compilerv1.TemplateResource{
			Meta:  self.Meta,
			Spec:  spec,
			State: state,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			return olap.Dialect().EscapeIdentifier(ref.Name), nil
		},
		Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
			if name.Kind == compilerv1.ResourceKindUnspecified {
				return compilerv1.TemplateResource{}, fmt.Errorf("can't resolve name %q without type specified", name.Name)
			}
			res, err := r.C.Get(ctx, runtime.ResourceNameFromCompiler(name), false)
			if err != nil {
				return compilerv1.TemplateResource{}, err
			}
			return compilerv1.TemplateResource{
				Meta:  res.Meta,
				Spec:  res.Resource.(*runtimev1.Resource_Model).Model.Spec,
				State: res.Resource.(*runtimev1.Resource_Model).Model.State,
			}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to resolve template: %w", err)
	}

	return olap.Exec(ctx, &drivers.Statement{
		Query:       sql,
		Priority:    100,
		LongRunning: true,
	})
}
