package reconcilers

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _defaultModelTimeout = 60 * time.Minute

const _modelSyncSplitsBatchSize = 1000

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
	b.State = a.State
	return nil
}

func (r *ModelReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetModel().State = &runtimev1.ModelState{}
	return nil
}

func (r *ModelReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	model := self.GetModel()
	if model == nil {
		return runtime.ReconcileResult{Err: errors.New("not a model")}
	}

	// If the model's state indicates that the last execution produced valid output, create a manager for the previous result
	var prevManager drivers.ModelManager
	var prevResult *drivers.ModelResult
	if model.State.ResultConnector != "" {
		conn, release, err := r.C.AcquireConn(ctx, model.State.ResultConnector)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		defer release()

		m, ok := conn.AsModelManager(r.C.InstanceID)
		if !ok {
			return runtime.ReconcileResult{Err: fmt.Errorf("connector %q does not support model management", model.State.ResultConnector)}
		}
		prevManager = m

		prevResult = &drivers.ModelResult{
			Connector:  model.State.ResultConnector,
			Properties: model.State.ResultProperties.AsMap(),
			Table:      model.State.ResultTable,
		}
	}

	// Fetch contextual config
	modelEnv, err := r.newModelEnv(ctx)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Handle deletion
	if self.Meta.DeletedOn != nil {
		if prevManager != nil {
			err = prevManager.Delete(ctx, prevResult)
			return runtime.ReconcileResult{Err: err}
		}
		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		if prevManager != nil {
			renameRes, err := prevManager.Rename(ctx, prevResult, self.Meta.Name.Name, modelEnv)
			if err == nil {
				err = r.updateStateWithResult(ctx, self, renameRes)
			}
			if err != nil {
				r.C.Logger.Warn("failed to rename model", zap.String("model", n.Name), zap.String("renamed_from", self.Meta.RenamedFrom.Name), zap.Error(err))
			}

			// Note: Not exiting early. We may need to retrigger the model in some cases. We also need to set the correct retrigger time.
		}
	}

	// Exit early if disabled
	if model.Spec.RefreshSchedule != nil && model.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// Check refs - stop if any of them are invalid
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		// If not staging changes, we need to drop the previous output (if any) before returning
		if !modelEnv.StageChanges && prevManager != nil {
			err2 := prevManager.Delete(ctx, prevResult)
			if err2 != nil {
				r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err2))
			}

			err2 = r.updateStateClear(ctx, self)
			if err2 != nil {
				r.C.Logger.Warn("refs check: failed to update state", zap.Any("error", err2))
			}
		}

		return runtime.ReconcileResult{Err: err}
	}

	// Compute hashes to determine if something has changes.
	// If the specHash changes, a full model reset is required (because the config changed).
	// If the refsHash changes, an incremental model run is sufficient (because the refs only went through a regular refresh).
	specHash, err := r.executionSpecHash(ctx, self.Meta.Refs, model.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute spec hash: %w", err)}
	}
	refsHash, err := r.refsStateHash(ctx, self.Meta.Refs, model.Spec)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute refs hash: %w", err)}
	}

	// Compute next time to refresh based on the RefreshSchedule (if any)
	var refreshOn time.Time
	if model.State.RefreshedOn != nil {
		refreshOn, err = nextRefreshTime(model.State.RefreshedOn.AsTime(), model.Spec.RefreshSchedule)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Check if the output still exists (might have been corrupted/lost somehow)
	var exists bool
	if prevManager != nil {
		exists, err = prevManager.Exists(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to check if model output exists", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Decide if we should trigger a reset
	triggerReset := model.State.ResultConnector == "" // If its nil, ResultProperties/ResultTable will also be nil
	triggerReset = triggerReset || model.State.RefreshedOn == nil
	triggerReset = triggerReset || model.State.SpecHash != specHash
	triggerReset = triggerReset || !exists

	// Decide if we should trigger
	trigger := triggerReset
	trigger = trigger || model.Spec.Trigger
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn)
	trigger = trigger || model.State.RefsHash != refsHash

	// Reschedule if we're not triggering
	if !trigger {
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// If the output connector has changed, drop data in the old output connector (if any).
	// If only the output properties have changed, the executor will handle dropping existing data (to comply with StageChanges).
	if prevManager != nil && model.State.ResultConnector != model.Spec.OutputConnector {
		err = prevManager.Delete(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Prepare the incremental state to pass to the executor
	incrementalRun := false
	incrementalState := map[string]any{}
	if !triggerReset && model.Spec.Incremental && prevResult != nil {
		// This is an incremental run!
		incrementalRun = true
		if model.State.IncrementalState != nil {
			incrementalState = model.State.IncrementalState.AsMap()
		}
	}
	incrementalState["incremental"] = incrementalRun // The incremental flag is hard-coded by convention

	// Build log message
	args := []zap.Field{zap.String("name", n.Name)}
	if incrementalRun {
		args = append(args, zap.String("run_type", "incremental"))
	} else {
		args = append(args, zap.String("run_type", "reset"))
	}
	if model.Spec.InputConnector == model.Spec.OutputConnector {
		args = append(args, zap.String("connector", model.Spec.InputConnector))
	} else {
		args = append(args, zap.String("input_connector", model.Spec.InputConnector), zap.String("output_connector", model.Spec.OutputConnector))
	}
	if model.Spec.StageConnector != "" {
		args = append(args, zap.String("stage_connector", model.Spec.StageConnector))
	}
	r.C.Logger.Debug("Building model output", args...)

	// Prepare the new execution options
	inputProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, model.Spec.InputConnector, model.Spec.InputProperties.AsMap())
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	outputProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, model.Spec.OutputConnector, model.Spec.OutputProperties.AsMap())
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	opts := &drivers.ModelExecutorOptions{
		Env:              modelEnv,
		ModelName:        self.Meta.Name.Name,
		InputConnector:   model.Spec.InputConnector,
		InputProperties:  inputProps,
		OutputConnector:  model.Spec.OutputConnector,
		OutputProperties: outputProps,
		Incremental:      model.Spec.Incremental,
		IncrementalRun:   incrementalRun,
		PreviousResult:   prevResult,
	}

	// Apply the timeout to the ctx
	timeout := _defaultModelTimeout
	if model.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(model.Spec.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// For safety, double check the ctx before executing the model (there may be some code paths where it's not checked)
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: ctx.Err()}
	}

	// Open executor for the new output and build the output
	var (
		executorConnector string
		execRes           *drivers.ModelResult
		execErr           error
	)
	if model.Spec.StageConnector != "" {
		stageProps, err := r.resolveTemplatedProps(ctx, self, incrementalState, model.Spec.StageConnector, model.Spec.StageProperties.AsMap())
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		executorConnector, execRes, execErr = r.executeWithStage(ctx, model.Spec.StageConnector, stageProps, opts)
	} else {
		executorConnector, execRes, execErr = r.execute(ctx, opts)
	}
	if execErr != nil {
		var err *modelBuildError
		if !errors.As(execErr, &err) {
			return runtime.ReconcileResult{Err: execErr}
		}
		// model build errors are handled later
	}

	// After the model has executed successfully, we re-evaluate the model's incremental state (not to be confused with the resource state)
	var newIncrementalState *structpb.Struct
	var newIncrementalStateSchema *runtimev1.StructType
	if execErr == nil {
		newIncrementalState, newIncrementalStateSchema, execErr = r.resolveIncrementalState(ctx, model)
	}

	// If the build succeeded, update the model's state accodingly
	if execErr == nil {
		model.State.ExecutorConnector = executorConnector
		model.State.SpecHash = specHash
		model.State.RefsHash = refsHash
		model.State.RefreshedOn = timestamppb.Now()
		model.State.IncrementalState = newIncrementalState
		model.State.IncrementalStateSchema = newIncrementalStateSchema
		err := r.updateStateWithResult(ctx, self, execRes)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// If the build failed, clear the state only if we're not staging changes
	if execErr != nil {
		if !modelEnv.StageChanges {
			err := r.updateStateClear(ctx, self)
			if err != nil {
				return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
			}
		}
	}

	// If the context was cancelled, we return now since we don't want to clear the trigger or set a next refresh time.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: errors.Join(ctx.Err(), execErr)}
	}

	// Reset spec.Trigger
	if model.Spec.Trigger {
		err := r.updateTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(time.Now(), model.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: errors.Join(err, execErr)}
	}

	// Note: If the build failed, this is where we return the error.
	return runtime.ReconcileResult{Err: execErr, Retrigger: refreshOn}
}

// executionSpecHash computes a hash of those model properties that impact execution.
// It also incorporates the spec hashes of the model's refs.
// If the spec hash changes, it means the model should be reset and fully re-executed.
func (r *ModelReconciler) executionSpecHash(ctx context.Context, refs []*runtimev1.ResourceName, spec *runtimev1.ModelSpec) (string, error) {
	hash := md5.New()

	for _, ref := range refs { // Refs are always sorted
		// Write name
		_, err := hash.Write([]byte(ref.Kind))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(ref.Name))
		if err != nil {
			return "", err
		}

		if ref.Kind != runtime.ResourceKindSource && ref.Kind != runtime.ResourceKindModel {
			continue
		}

		r, err := r.C.Get(ctx, ref, false)
		if err != nil {
			continue
		}

		var refSpechHash string
		switch ref.Kind {
		case runtime.ResourceKindSource:
			refSpechHash = r.GetSource().State.SpecHash
		case runtime.ResourceKindModel:
			refSpechHash = r.GetModel().State.SpecHash
		}

		_, err = hash.Write([]byte(refSpechHash))
		if err != nil {
			return "", err
		}
	}

	err := binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.Incremental)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.IncrementalStateResolver))
	if err != nil {
		return "", err
	}

	if spec.IncrementalStateResolverProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.IncrementalStateResolverProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.IncrementalStateResolverProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	_, err = hash.Write([]byte(spec.InputConnector))
	if err != nil {
		return "", err
	}

	if spec.InputProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.InputProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.InputProperties.AsMap())
		if err != nil {
			return "", err
		}
		err = hashWriteMapOrdered(hash, res)
		if err != nil {
			return "", err
		}
	}

	_, err = hash.Write([]byte(spec.OutputConnector))
	if err != nil {
		return "", err
	}

	if spec.OutputProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.OutputProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.OutputProperties.AsMap())
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

// refsStateHash computes a hash of the state of the model's refs.
// It is used to check if the model's refs have been updated, which should trigger an (incremental) model execution.
// (Note that the refs state hash identifies when to trigger incremental runs, whereas the the execution spec hash identifies when to trigger full resets.)
func (r *ModelReconciler) refsStateHash(ctx context.Context, refs []*runtimev1.ResourceName, spec *runtimev1.ModelSpec) (string, error) {
	if spec.RefreshSchedule == nil || !spec.RefreshSchedule.RefUpdate {
		return "", nil
	}

	hash := md5.New()

	for _, ref := range refs {
		_, err := hash.Write([]byte(ref.Kind))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(ref.Name))
		if err != nil {
			return "", err
		}

		// Note: Only writing the state info to the hash, not spec version, because it doesn't matter whether the spec/meta changes, only whether the state changes.
		// Note: Also using StateUpdatedOn because the state version is reset when the resource is deleted and recreated.
		r, err := r.C.Get(ctx, ref, false)
		var stateVersion, stateUpdatedOn int64
		if err == nil {
			stateVersion = r.Meta.StateVersion
			stateUpdatedOn = r.Meta.StateUpdatedOn.Seconds
		} else {
			stateVersion = -1
		}
		err = binary.Write(hash, binary.BigEndian, stateVersion)
		if err != nil {
			return "", err
		}
		err = binary.Write(hash, binary.BigEndian, stateUpdatedOn)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// updateStateWithResult updates the model resource's state with the result of a model execution.
// It only updates the result-related fields. If changing other fields, such as RefreshedOn and SpecHash, they must be assigned before calling this function.
func (r *ModelReconciler) updateStateWithResult(ctx context.Context, self *runtimev1.Resource, res *drivers.ModelResult) error {
	mdl := self.GetModel()

	props, err := structpb.NewStruct(res.Properties)
	if err != nil {
		return fmt.Errorf("failed to serialize result properties: %w", err)
	}

	mdl.State.ResultConnector = res.Connector
	mdl.State.ResultProperties = props
	mdl.State.ResultTable = res.Table

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// updateStateClear clears the model resource's state.
func (r *ModelReconciler) updateStateClear(ctx context.Context, self *runtimev1.Resource) error {
	mdl := self.GetModel()

	mdl.State.ExecutorConnector = ""
	mdl.State.ResultConnector = ""
	mdl.State.ResultProperties = nil
	mdl.State.ResultTable = ""
	mdl.State.SpecHash = ""
	mdl.State.RefsHash = ""
	mdl.State.RefreshedOn = nil
	mdl.State.IncrementalState = nil
	mdl.State.IncrementalStateSchema = nil

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// updateTriggerFalse sets the model's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *ModelReconciler) updateTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, false)
	if err != nil {
		return err
	}

	model := self.GetModel()
	if model == nil {
		return fmt.Errorf("not a model")
	}

	model.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// resolveIncrementalState resolves the incremental state of a model using its configured incremental state resolver.
// Note the ambiguity around "state" in models – all resources have a "spec" and a "state",
// but models also have a resolver for "incremental state" that enables incremental/stateful computation by persisting data from the previous execution.
// It returns nil results if an incremental state resolver is not configured or does not return any data.
func (r *ModelReconciler) resolveIncrementalState(ctx context.Context, mdl *runtimev1.ModelV2) (*structpb.Struct, *runtimev1.StructType, error) {
	if !mdl.Spec.Incremental {
		return nil, nil, nil
	}

	if mdl.Spec.IncrementalStateResolver == "" {
		return nil, nil, nil
	}

	res, err := r.C.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.C.InstanceID,
		Resolver:           mdl.Spec.IncrementalStateResolver,
		ResolverProperties: mdl.Spec.IncrementalStateResolverProperties.AsMap(),
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		return nil, nil, err
	}
	defer res.Close()

	row, err := res.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Not returning any rows will clear the state
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to read state resolver output: %w", err)
	}

	state, err := structpb.NewStruct(row)
	if err != nil {
		return nil, nil, fmt.Errorf("state resolver produced invalid output: %w", err)
	}

	return state, res.Schema(), nil
}

// resolveAndSyncSplits resolves the model's splits using its configured splits resolver and inserts or updates them in the catalog.
func (r *ModelReconciler) resolveAndSyncSplits(ctx context.Context, mdl *runtimev1.ModelV2, incrementalState map[string]any) error {
	if mdl.Spec.SplitsResolver == "" {
		return nil
	}

	// Resolve split rows
	res, err := r.C.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.C.InstanceID,
		Resolver:           mdl.Spec.SplitsResolver,
		ResolverProperties: mdl.Spec.SplitsResolverProperties.AsMap(),
		Args:               map[string]any{"state": incrementalState},
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	if err != nil {
		return err
	}
	defer res.Close()

	// Consume the rows and sync them in batches
	var batch []map[string]any
	var batchStartIdx int
	for {
		// Read a row
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to read splits resolver output: %w", err)
		}
		batch = append(batch, row)

		// Flush a batch of rows
		if len(batch) >= _modelSyncSplitsBatchSize {
			// Sync the splits
			err = r.syncSplits(ctx, mdl, batchStartIdx, batch)
			if err != nil {
				return err
			}

			// Track the row index of the first row in the batch
			batchStartIdx += len(batch)

			// Reset the batch without reallocating
			for i := range batch {
				batch[i] = nil
			}
			batch = batch[:0]
		}
	}

	// Flush the remaining rows not handled in the loop
	return r.syncSplits(ctx, mdl, batchStartIdx, batch)
}

// syncSplits syncs a batch of split rows to the catalog.
// If a split doesn't exist, it is inserted and marked for execution.
// If a split already exists, it will be ignored unless its watermark field has advanced, in which case it will be marked for execution.
//
// The startIdx should be the index of the first row in the batch in the full splits dataset.
// Split indexes only inform the order that splits are executed in, so they don't need to be very consistent across invocations.
func (r *ModelReconciler) syncSplits(ctx context.Context, mdl *runtimev1.ModelV2, startIdx int, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}

	catalog, release, err := r.C.Runtime.Catalog(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}
	defer release()

	// Build ModelSplit objects indexed by their Key
	splits := make(map[string]drivers.ModelSplit, len(rows))
	for i, row := range rows {
		// If a watermark field is configured, we extract and remove it from the map.
		// It is necessary to remove it to ensure the key is deterministic.
		var watermark *time.Time
		if mdl.Spec.SplitsWatermarkField != "" {
			if v, ok := row[mdl.Spec.SplitsWatermarkField]; ok {
				t, ok := v.(time.Time)
				if !ok {
					return fmt.Errorf(`expected a timestamp for split watermark field %q, got type %T`, mdl.Spec.SplitsWatermarkField, v)
				}

				watermark = &t
				delete(row, mdl.Spec.SplitsWatermarkField)
			}
		}

		// Marshal the rest of the row
		rowJSON, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("failed to marshal split row at index %d: %w", i, err)
		}

		// JSON serialization is deterministic, so we can hash it to get a key
		key, err := secureHash(rowJSON)
		if err != nil {
			return fmt.Errorf("failed to hash split row at index %d: %w", i, err)
		}

		splits[key] = drivers.ModelSplit{
			Key:       key,
			DataJSON:  rowJSON,
			Index:     startIdx + i,
			Watermark: watermark,
		}
	}

	// Find those splits that already exist in the catalog
	keys := make([]string, 0, len(splits))
	for key := range splits {
		keys = append(keys, key)
	}
	existing, err := catalog.FindModelSplitsByKeys(ctx, mdl.State.ModelId, keys)
	if err != nil {
		return fmt.Errorf("failed to find existing splits: %w", err)
	}

	// Handle the existing skips by skipping or updating them.
	// We remove the handled splits from the splits map. The ones that remain are new and should be inserted.
	for _, old := range existing {
		// Pop the matching split from the map
		split := splits[old.Key]
		delete(splits, old.Key)

		// If the watermark hasn't advanced, there's nothing to do
		if split.Watermark == nil {
			continue
		}
		if old.Watermark != nil && !old.Watermark.Before(*split.Watermark) {
			continue
		}

		// Update the split (since the new ExecutedOn will be nil, it will be marked for execution)
		err = catalog.UpdateModelSplit(ctx, mdl.State.ModelId, split)
		if err != nil {
			return fmt.Errorf("failed to update existing split: %w", err)
		}
	}

	// The remaining splits are new and should be inserted
	for _, split := range splits {
		err = catalog.InsertModelSplit(ctx, mdl.State.ModelId, split)
		if err != nil {
			return fmt.Errorf("failed to insert new split: %w", err)
		}
	}
	return nil
}

// execute executes a model with the given execution options.
func (r *ModelReconciler) execute(ctx context.Context, opts *drivers.ModelExecutorOptions) (string, *drivers.ModelResult, error) {
	executorName, e, release, err := r.acquireExecutor(ctx, opts)
	if err != nil {
		return "", nil, err
	}
	defer release()

	res, err := e.Execute(ctx)
	if err != nil {
		return "", nil, &modelBuildError{err: err}
	}
	return executorName, res, nil
}

// executeWithStage executes a model with a stage connector by first running an executor from the input connector to the stage connector,
// and then running an executor from the stage connector to the output connector.
func (r *ModelReconciler) executeWithStage(ctx context.Context, stageConnector string, stageProps map[string]any, opts *drivers.ModelExecutorOptions) (string, *drivers.ModelResult, error) {
	// we want to determine whether stage connector and output connector are compatible
	// so we get executor but do not execute it since we need to use result of stage 1
	//
	// Build model option for stage 2
	stage2Opts := &drivers.ModelExecutorOptions{
		InputConnector:   stageConnector,
		InputProperties:  stageProps,
		OutputConnector:  opts.OutputConnector,
		OutputProperties: opts.OutputProperties,
	}
	_, _, tempRel, err := r.acquireExecutor(ctx, stage2Opts)
	if err != nil {
		return "", nil, err
	}
	tempRel()

	// Build model options for stage 1
	// set stage props as outputprops
	stage1Opts := &drivers.ModelExecutorOptions{
		Env:              opts.Env,
		ModelName:        opts.ModelName,
		InputConnector:   opts.InputConnector,
		InputProperties:  opts.InputProperties,
		OutputConnector:  stageConnector,
		OutputProperties: stageProps,
	}
	_, executor, rel, err := r.acquireExecutor(ctx, stage1Opts)
	if err != nil {
		return "", nil, err
	}

	// execute stage 1
	res1, err := executor.Execute(ctx)
	if err != nil {
		rel()
		return "", nil, &modelBuildError{err: err}
	}
	rel()

	sc, sr, err := r.C.AcquireConn(ctx, res1.Connector)
	if err != nil {
		return "", nil, err
	}
	defer sr()

	// Build model option for stage 2
	// Use result's connector and result properties as input connector
	// Typically the result connector will be same as stage connector
	// but the properties can change
	stage2Opts = &drivers.ModelExecutorOptions{
		Env:              opts.Env,
		ModelName:        opts.ModelName,
		InputConnector:   res1.Connector,
		InputProperties:  res1.Properties,
		OutputConnector:  opts.OutputConnector,
		OutputProperties: opts.OutputProperties,
		Incremental:      opts.Incremental,
		IncrementalRun:   opts.IncrementalRun,
		PreviousResult:   opts.PreviousResult,
	}
	name, executor, rel, err := r.acquireExecutor(ctx, stage2Opts)
	if err != nil {
		// cleanup stage1 result data
		// This is done in same context. Can leak stage data in case of ctx cancellations.
		if mm, ok := sc.AsModelManager(r.C.InstanceID); ok {
			return "", nil, errors.Join(err, mm.Delete(ctx, res1))
		}
		return "", nil, err
	}

	// the final cleanup should also cleanup stage1 result data
	defer func() {
		rel()
		rc, rr, err := r.C.AcquireConn(ctx, res1.Connector)
		if err != nil {
			return
		}
		defer rr()
		if mm, ok := rc.AsModelManager(r.C.InstanceID); ok {
			// This is done in same context. Can leak stage data in case of ctx cancellations.
			err = mm.Delete(ctx, res1)
			if err != nil {
				r.C.Logger.Warn("failed to clean up stage output", zap.Error(err))
			}
		}
	}()

	res2, err := executor.Execute(ctx)
	if err != nil {
		return "", nil, &modelBuildError{err: err}
	}
	return name, res2, nil
}

// acquireExecutor acquires a ModelExecutor capable of executing a model with the given execution options.
// It handles acquiring and setting opts.InputHandle and opts.OutputHandle.
func (r *ModelReconciler) acquireExecutor(ctx context.Context, opts *drivers.ModelExecutorOptions) (string, drivers.ModelExecutor, func(), error) {
	ic, ir, err := r.C.AcquireConn(ctx, opts.InputConnector)
	if err != nil {
		return "", nil, nil, err
	}

	if opts.InputConnector == opts.OutputConnector {
		opts.InputHandle = ic
		opts.OutputHandle = ic

		e, ok := ic.AsModelExecutor(r.C.InstanceID, opts)
		if !ok {
			return "", nil, nil, fmt.Errorf("connector %q is not capable of executing models", opts.InputConnector)
		}

		return opts.InputConnector, e, ir, nil
	}

	oc, or, err := r.C.AcquireConn(ctx, opts.OutputConnector)
	if err != nil {
		ir()
		return "", nil, nil, err
	}

	opts.InputHandle = ic
	opts.OutputHandle = oc

	executorName := opts.InputConnector
	e, ok := ic.AsModelExecutor(r.C.InstanceID, opts)
	if !ok {
		executorName = opts.OutputConnector
		e, ok = oc.AsModelExecutor(r.C.InstanceID, opts)
		if !ok {
			ir()
			or()
			return "", nil, nil, fmt.Errorf("cannot execute model: input connector %q and output connector %q are not compatible", opts.InputConnector, opts.OutputConnector)
		}
	}

	release := func() {
		ir()
		or()
	}

	return executorName, e, release, nil
}

// newModelEnv makes a ModelEnv configured using the current instance.
func (r *ModelReconciler) newModelEnv(ctx context.Context) (*drivers.ModelEnv, error) {
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access instance config: %w", err)
	}

	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access repo: %w", err)
	}
	defer release()

	return &drivers.ModelEnv{
		AllowHostAccess:    r.C.Runtime.AllowHostAccess(),
		RepoRoot:           repo.Root(),
		StageChanges:       cfg.StageChanges,
		DefaultMaterialize: cfg.ModelDefaultMaterialize,
		AcquireConnector:   r.C.AcquireConn,
	}, nil
}

// resolveTemplatedProps resolves template tags in strings nested in the provided props.
// Passing a connector is optional. If a connector is provided, it will be used to inform how values are escaped.
func (r *ModelReconciler) resolveTemplatedProps(ctx context.Context, self *runtimev1.Resource, incrementalState map[string]any, connector string, props map[string]any) (map[string]any, error) {
	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return nil, err
	}

	// If we know the prop's connector AND it's an OLAP, we use its dialect to escape refs
	var dialect drivers.Dialect
	if connector != "" {
		olap, release, err := r.C.AcquireOLAP(ctx, connector)
		if err == nil {
			dialect = olap.Dialect()
			release()
		}
	}

	td := compilerv1.TemplateData{
		Environment: inst.Environment,
		User:        map[string]any{},
		Variables:   inst.ResolveVariables(),
		State:       incrementalState,
		Self: compilerv1.TemplateResource{
			Meta:  self.Meta,
			Spec:  self.GetModel().Spec,
			State: self.GetModel().State,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			if dialect == drivers.DialectUnspecified {
				return ref.Name, nil
			}
			return dialect.EscapeIdentifier(ref.Name), nil
		},
	}

	val, err := compilerv1.ResolveTemplateRecursively(props, td)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template: %w", err)
	}
	return val.(map[string]any), nil
}

// analyzeTemplatedVariables analyzes strings nested in the provided props for template tags that reference instance variables.
// It returns a map of variable names referenced in the props mapped to their current value.
func (r *ModelReconciler) analyzeTemplatedVariables(ctx context.Context, props map[string]any) (map[string]string, error) {
	res := make(map[string]string)
	err := analyzeTemplatedVariables(props, res)
	if err != nil {
		return nil, err
	}

	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables()

	for k := range res {
		k2 := strings.TrimPrefix(k, "vars.")
		if len(k) == len(k2) {
			continue
		}

		res[k] = vars[k2]
	}

	return res, nil
}

// analyzeTemplatedVariables analyzes strings nested in the provided value for template tags that reference variables.
// Variables are added as keys to the provided map, with empty strings as values.
// The values are empty strings instead of booleans as an optimization to enable re-using the map in upstream code.
func analyzeTemplatedVariables(val any, res map[string]string) error {
	switch val := val.(type) {
	case string:
		meta, err := compilerv1.AnalyzeTemplate(val)
		if err != nil {
			return err
		}
		for _, k := range meta.Variables {
			res[k] = ""
		}
	case map[string]any:
		for _, v := range val {
			err := analyzeTemplatedVariables(v, res)
			if err != nil {
				return err
			}
		}
	case []any:
		for _, v := range val {
			err := analyzeTemplatedVariables(v, res)
			if err != nil {
				return err
			}
		}
	default:
		// Nothing to do
	}
	return nil
}

// hashWriteMapOrdered writes the keys and values of a map to the writer in a deterministic order.
func hashWriteMapOrdered(w io.Writer, m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err := w.Write([]byte(k))
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(m[k]))
		if err != nil {
			return err
		}
	}

	return nil
}

// secureHash returns a hex-encoded SHA-256 hash of the provided byte slice.
func secureHash(val []byte) (string, error) {
	hash := sha256.New()
	_, err := hash.Write(val)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type modelBuildError struct {
	err error
}

func (e *modelBuildError) Error() string {
	return e.err.Error()
}

func (e *modelBuildError) Unwrap() error { return e.err }
