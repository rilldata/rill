package reconcilers

import (
	"context"
	"crypto/md5"
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

const _defaultModelTimeout = 15 * time.Minute

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

	// Prepare the state to pass to the executor
	incrementalRun := false
	state := map[string]any{}
	if !triggerReset && model.Spec.Incremental && prevResult != nil {
		// This is an incremental run!
		incrementalRun = true
		if model.State.State != nil {
			state = model.State.State.AsMap()
		}
	}
	state["incremental"] = incrementalRun // The incremental flag is hard-coded in the state by convention

	// Prepare the new execution options
	inputProps, err := r.resolveTemplatedProps(ctx, self, state, model.Spec.InputConnector, model.Spec.InputProperties.AsMap())
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	outputProps, err := r.resolveTemplatedProps(ctx, self, state, model.Spec.OutputConnector, model.Spec.OutputProperties.AsMap())
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

	// Open executor for the new output
	executorConnector, executor, release, err := r.acquireExecutor(ctx, opts)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	defer release()

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

	// Build the output
	execRes, execErr := executor.Execute(ctx)
	if execErr != nil {
		execErr = fmt.Errorf("failed to build output: %w", execErr)
	}

	// After the model has executed successfully, we re-evaluate the model state (not to be confused with the resource state)
	var newState *structpb.Struct
	var newStateSchema *runtimev1.StructType
	if execErr == nil {
		newState, newStateSchema, execErr = r.resolveState(ctx, model)
	}

	// If the build succeeded, update the model's state accodingly
	if execErr == nil {
		model.State.ExecutorConnector = executorConnector
		model.State.SpecHash = specHash
		model.State.RefsHash = refsHash
		model.State.RefreshedOn = timestamppb.Now()
		model.State.State = newState
		model.State.StateSchema = newStateSchema
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

	_, err = hash.Write([]byte(spec.StateResolver))
	if err != nil {
		return "", err
	}

	if spec.StateResolverProperties != nil {
		err = pbutil.WriteHash(structpb.NewStructValue(spec.StateResolverProperties), hash)
		if err != nil {
			return "", err
		}

		res, err := r.analyzeTemplatedVariables(ctx, spec.StateResolverProperties.AsMap())
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
	mdl.State.State = nil
	mdl.State.StateSchema = nil

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

// resolveState resolves the state of a model using its configured state resolver.
// Note the ambiguity around "state" in models – all resources have a "spec" and a "state",
// but models also have a "state" resolver that enables incremental/stateful computation by persisting data from the previous execution.
// It returns nil results if a state resolver is not configured or does not return any data.
func (r *ModelReconciler) resolveState(ctx context.Context, mdl *runtimev1.ModelV2) (*structpb.Struct, *runtimev1.StructType, error) {
	if mdl.Spec.StateResolver == "" {
		return nil, nil, nil
	}

	res, err := r.C.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.C.InstanceID,
		Resolver:           mdl.Spec.StateResolver,
		ResolverProperties: mdl.Spec.StateResolverProperties.AsMap(),
	})
	if err != nil {
		return nil, nil, err
	}

	var tmp []map[string]any
	err = json.Unmarshal(res.Data, &tmp)
	if err != nil {
		return nil, nil, fmt.Errorf("state resolver produced invalid JSON: %w", err)
	}

	if len(tmp) == 0 {
		// Not returning any rows will clear the state
		return nil, nil, nil
	}

	if len(tmp) > 1 {
		return nil, nil, fmt.Errorf("state resolver produced more than one row")
	}

	state, err := structpb.NewStruct(tmp[0])
	if err != nil {
		return nil, nil, fmt.Errorf("state resolver produced invalid output: %w", err)
	}

	return state, res.Schema, nil
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
func (r *ModelReconciler) resolveTemplatedProps(ctx context.Context, self *runtimev1.Resource, state map[string]any, connector string, props map[string]any) (map[string]any, error) {
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
		State:       state,
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

	val, err := resolveTemplatedValue(td, props)
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

// resolveTemplatedValue resolves template tags nested in strings in the provided value.
func resolveTemplatedValue(td compilerv1.TemplateData, val any) (any, error) {
	switch val := val.(type) {
	case string:
		return compilerv1.ResolveTemplate(val, td)
	case map[string]any:
		for k, v := range val {
			v, err := resolveTemplatedValue(td, v)
			if err != nil {
				return nil, err
			}
			val[k] = v
		}
		return val, nil
	case []any:
		for i, v := range val {
			v, err := resolveTemplatedValue(td, v)
			if err != nil {
				return nil, err
			}
			val[i] = v
		}
		return val, nil
	default:
		return val, nil
	}
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
