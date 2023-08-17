package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _defaultMaterializeTimeout = 15 * time.Minute

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindModel, newModelReconciler)
}

type ModelReconciler struct {
	C *runtime.Controller
}

func newModelReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ModelReconciler{C: c}
}

func (r *ModelReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ModelReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetModel()
	b := to.GetModel()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ModelReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetModel()
	b := to.GetModel()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ModelReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	model := self.GetModel()

	// The view/table name is derived from the resource name.
	// We only set src.State.Table after it has been created,
	tableName := self.Meta.Name.Name

	// Handle deletion
	if self.Meta.DeletedOn != nil {
		if t, ok := olapTableInfo(ctx, r.C, model.State.Connector, model.State.Table); ok {
			olapDropTableIfExists(ctx, r.C, model.State.Connector, model.State.Table, t.View)
		}
		if t, ok := olapTableInfo(ctx, r.C, model.State.Connector, r.stagingTableName(tableName)); ok {
			olapDropTableIfExists(ctx, r.C, model.State.Connector, t.Name, t.View)
		}
		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		if t, ok := olapTableInfo(ctx, r.C, model.State.Connector, model.State.Table); ok {
			// Clear any existing table with the new name
			if t2, ok := olapTableInfo(ctx, r.C, model.State.Connector, tableName); ok {
				olapDropTableIfExists(ctx, r.C, model.State.Connector, t2.Name, t2.View)
			}

			// Rename and update state
			err = olapRenameTable(ctx, r.C, model.State.Connector, model.State.Table, tableName, t.View)
			if err != nil {
				return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename model: %w", err)}
			}
			model.State.Table = tableName
			err = r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}
		// Note: Not exiting early. It might need to be created/materialized., and we need to set the correct retrigger time based on the refresh schedule.
	}

	// TODO: Exit if refs have errors

	// TODO: Incorporate changes to refs in hash – track if refs have changed (deleted, added, or state updated)

	// Use a hash of execution-related fields from the spec to determine if something has changed
	hash, err := r.executionSpecHash(model.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Compute next time to refresh based on the RefreshSchedule (if any)
	var refreshOn time.Time
	if model.State.RefreshedOn != nil {
		refreshOn, err = nextRefreshTime(model.State.RefreshedOn.AsTime(), model.Spec.RefreshSchedule)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Determine if we should materialize
	var materialize bool
	if model.Spec.Materialize != nil {
		materialize = true
	}

	// Check if the model still exists (might have been corrupted/lost somehow)
	t, exists := olapTableInfo(ctx, r.C, model.State.Connector, model.State.Table)

	// Decide if we should trigger an update
	trigger := model.Spec.Trigger
	trigger = trigger || model.State.Table == ""
	trigger = trigger || model.State.RefreshedOn == nil
	trigger = trigger || model.State.SpecHash != hash
	trigger = trigger || !exists
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn)

	// We support "delayed materialization" for models. It will materialize a model if it stays unchanged for a duration of time.
	// This is useful to support keystroke-by-keystroke editing.
	var delayedMaterializeOn time.Time
	var delayedMaterialize bool
	if !trigger && materialize && t.View {
		delayedMaterializeOn, delayedMaterialize = r.delayedMaterializeTime(model.Spec, model.State.RefreshedOn.AsTime())
	}

	// Reschedule if we're not triggering
	if !trigger && !delayedMaterialize {
		// In theory there are some more cases to cover here, but we assume materialize delays are shorter than refresh schedules.
		if !delayedMaterializeOn.IsZero() {
			return runtime.ReconcileResult{Retrigger: delayedMaterializeOn}
		}
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// If the Connector was changed, drop data in the old connector
	if model.State.Table != "" && model.State.Connector != model.Spec.Connector {
		if t, ok := olapTableInfo(ctx, r.C, model.State.Connector, model.State.Table); ok {
			olapDropTableIfExists(ctx, r.C, model.State.Connector, model.State.Table, t.View)
		}
		if t, ok := olapTableInfo(ctx, r.C, model.State.Connector, r.stagingTableName(model.State.Table)); ok {
			olapDropTableIfExists(ctx, r.C, model.State.Connector, t.Name, t.View)
		}
	}

	// Always stage changes if running a delayed materialization
	stage := model.Spec.StageChanges || delayedMaterialize
	stagingTableName := tableName
	if stage {
		stagingTableName = r.stagingTableName(tableName)
	}

	// Determine if we should delay materialization (note difference between "delayedMaterialize" and "delayingMaterialize")
	delayingMaterialize := false
	if materialize && model.State.SpecHash != hash && model.Spec.MaterializeDelaySeconds > 0 {
		delayingMaterialize = true
		materialize = false
	}

	// Should NEVER happen
	if delayedMaterialize && delayingMaterialize {
		return runtime.ReconcileResult{Err: fmt.Errorf("internal error: delayed and delaying materialization")}
	}

	// Drop the staging table if it exists
	connector := model.Spec.Connector
	if t, ok := olapTableInfo(ctx, r.C, connector, stagingTableName); ok {
		olapDropTableIfExists(ctx, r.C, connector, t.Name, t.View)
	}

	// Create the model
	createErr := r.createModel(ctx, self, stagingTableName, !materialize)
	if createErr != nil {
		createErr = fmt.Errorf("failed to create model: %w", createErr)
	}
	if createErr == nil && stage {
		// Drop the main view/table
		if t, ok := olapTableInfo(ctx, r.C, connector, tableName); ok {
			olapDropTableIfExists(ctx, r.C, connector, t.Name, t.View)
		}
		// Rename the staging table to main view/table
		err = olapRenameTable(ctx, r.C, connector, stagingTableName, tableName, !materialize)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename staged model: %w", err)}
		}
	}

	// How we handle ingestErr depends on several things:
	// If ctx was cancelled, we cleanup and exit
	// If StageChanges is true, we retain the existing table, but still return the error.
	// If StageChanges is false, we clear the existing table and return the error.

	// ctx will only be cancelled in cases where the Controller guarantees a new call to Reconcile.
	// We just clean up temp tables and state, then return.
	cleanupCtx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		cleanupCtx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
	}

	// Update state
	update := false
	if createErr == nil {
		// Successful ingestion
		update = true
		model.State.Connector = connector
		model.State.Table = tableName
		model.State.SpecHash = hash
		model.State.RefreshedOn = timestamppb.Now()
	} else if model.Spec.StageChanges {
		// Failed ingestion to staging table
		olapDropTableIfExists(cleanupCtx, r.C, connector, stagingTableName, !materialize)
	} else {
		// Failed ingestion to main table
		update = true
		olapDropTableIfExists(cleanupCtx, r.C, connector, tableName, !materialize)
		model.State.Connector = ""
		model.State.Table = ""
		model.State.SpecHash = ""
		model.State.RefreshedOn = nil
	}
	if update {
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// See earlier note – essential cleanup is done, we can return now.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: createErr}
	}

	// Reset spec.Trigger
	if model.Spec.Trigger {
		model.Spec.Trigger = false
		err = r.C.UpdateSpec(ctx, self.Meta.Name, self.Meta.Refs, self.Meta.Owner, self.Meta.FilePaths, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// If we're delaying materialization, we need to retrigger after the delay
	if createErr == nil && delayingMaterialize {
		t, ok := r.delayedMaterializeTime(model.Spec, time.Now())
		if ok {
			return runtime.ReconcileResult{Retrigger: t}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(time.Now(), model.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: createErr, Retrigger: refreshOn}
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func (r *ModelReconciler) stagingTableName(table string) string {
	return "__rill_tmp_model_" + table
}

// delayedMaterializeTime derives the timestamp (if any) to materialize a model with delayed materialization configured.
func (r *ModelReconciler) delayedMaterializeTime(spec *runtimev1.ModelSpec, since time.Time) (time.Time, bool) {
	if spec.MaterializeDelaySeconds == 0 {
		return time.Time{}, false
	}
	return since.Add(time.Duration(spec.MaterializeDelaySeconds) * time.Second), true
}

// executionSpecHash computes a hash of only those model spec properties that impact execution.
func (r *ModelReconciler) executionSpecHash(spec *runtimev1.ModelSpec) (string, error) {
	hash := md5.New()

	_, err := hash.Write([]byte(spec.Connector))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.Sql))
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.Materialize)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.UsesTemplating)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createModel creates or updates the model in the OLAP connector.
func (r *ModelReconciler) createModel(ctx context.Context, self *runtimev1.Resource, tableName string, view bool) error {
	inst, err := r.C.Runtime.FindInstance(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}

	spec := self.Resource.(*runtimev1.Resource_Model).Model.Spec
	state := self.Resource.(*runtimev1.Resource_Model).Model.State

	var sql string
	if spec.UsesTemplating {
		sql, err = compilerv1.ResolveTemplate(spec.Sql, compilerv1.TemplateData{
			Claims:    map[string]interface{}{},
			Variables: inst.ResolveVariables(),
			Self: compilerv1.TemplateResource{
				Meta:  self.Meta,
				Spec:  spec,
				State: state,
			},
			Resolve: func(ref compilerv1.ResourceName) (string, error) {
				return safeSQLName(ref.Name), nil
			},
			Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
				if name.Kind == compilerv1.ResourceKindUnspecified {
					return compilerv1.TemplateResource{}, fmt.Errorf("can't resolve name %q without kind specified", name.Name)
				}
				res, err := r.C.Get(ctx, resourceNameFromCompiler(name))
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
	} else {
		sql = spec.Sql
	}

	olap, release, err := r.C.AcquireOLAP(ctx, spec.Connector)
	if err != nil {
		return err
	}
	defer release()

	// If materializing, set timeout on ctx
	if !view {
		timeout := _defaultMaterializeTimeout
		if spec.TimeoutSeconds > 0 {
			timeout = time.Duration(spec.TimeoutSeconds) * time.Second
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE %s %s AS (%s)", typ, safeSQLName(tableName), sql),
		Priority: 100,
	})
}
