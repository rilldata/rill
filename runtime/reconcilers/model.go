package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
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

	// If the model's state indicates that the last execution produced valid output, create an executor for the previous output
	var prevExecutor drivers.ModelExecutor
	var prevResult *drivers.ModelExecuteResult
	if model.State.ExecutorConnector != "" {
		conn, release, err := r.C.AcquireConn(ctx, model.State.ExecutorConnector)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		defer release()

		executor, ok := conn.AsModelExecutor()
		if !ok {
			return runtime.ReconcileResult{Err: fmt.Errorf("connector %q no longer supports model execution", model.State.ExecutorConnector)}
		}
		prevExecutor = executor

		prevResult = &drivers.ModelExecuteResult{
			Connector:  model.State.ResultConnector,
			Properties: model.State.ResultProperties.AsMap(),
		}
	}

	// Fetch contextual config
	executorEnv, err := r.newExecutorEnv(ctx)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Handle deletion
	if self.Meta.DeletedOn != nil {
		if prevExecutor != nil {
			err = prevExecutor.Delete(ctx, prevResult)
			return runtime.ReconcileResult{Err: err}
		}
		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		if prevExecutor != nil {
			renameRes, err := prevExecutor.Rename(ctx, &drivers.ModelRenameOptions{
				NewName:        self.Meta.Name.Name,
				PreviousName:   self.Meta.RenamedFrom.Name,
				PreviousResult: prevResult,
				Env:            executorEnv,
			})
			if err != nil {
				r.C.Logger.Warn("failed to rename model", zap.String("model", n.Name), zap.String("renamed_from", self.Meta.RenamedFrom.Name), zap.Error(err))
			} else {
				resultProps, err := structpb.NewStruct(renameRes.Properties)
				if err != nil {
					r.C.Logger.Warn("failed to build result properties after rename", zap.String("model", n.Name), zap.Error(err))
				} else {
					model.State.ResultConnector = renameRes.Connector
					model.State.ResultProperties = resultProps
					model.State.ResultTable = renameRes.Table
					model.State.RefreshedOn = timestamppb.Now()
					err = r.C.UpdateState(ctx, self.Meta.Name, self)
					if err != nil {
						return runtime.ReconcileResult{Err: err}
					}
				}
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
		// If not staging changes, we need to drop the current output (if any)
		if !executorEnv.StageChanges && prevExecutor != nil {
			err = prevExecutor.Delete(ctx, prevResult)
			if err != nil {
				r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err))
			}

			model.State.ExecutorConnector = ""
			model.State.ResultConnector = ""
			model.State.ResultProperties = nil
			model.State.ResultTable = ""
			model.State.SpecHash = ""
			model.State.RefreshedOn = nil
			subErr := r.C.UpdateState(ctx, self.Meta.Name, self)
			if subErr != nil {
				r.C.Logger.Error("refs check: failed to update state", zap.Any("error", subErr))
			}
		}

		return runtime.ReconcileResult{Err: err}
	}

	// Use a hash of execution-related fields from the spec to determine if something has changed
	hash, err := r.executionSpecHash(ctx, self.Meta.Refs, model.Spec)
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

	// Check if the output still exists (might have been corrupted/lost somehow)
	var exists bool
	if prevExecutor != nil {
		exists, err = prevExecutor.Exists(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to check if model output exists", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Decide if we should trigger an update
	trigger := model.Spec.Trigger
	trigger = trigger || model.State.ResultConnector == "" // If its nil, ExecutorConnector/ResultProperties/ResultTable will also be nil
	trigger = trigger || model.State.RefreshedOn == nil
	trigger = trigger || model.State.SpecHash != hash
	trigger = trigger || !exists
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn)

	// Reschedule if we're not triggering
	if !trigger {
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// If the output connector has changed, drop data in the old output connector (if any).
	// If only the output properties have changed, the executor will handle dropping existing data (to comply with StageChanges).
	if prevExecutor != nil && model.State.ResultConnector != model.Spec.OutputConnector {
		err = prevExecutor.Delete(ctx, prevResult)
		if err != nil {
			r.C.Logger.Warn("failed to delete model output", zap.String("model", n.Name), zap.Error(err))
		}
	}

	// Prepare the new execution options
	incremental := model.Spec.Incremental && prevResult != nil // TODO: Not if resetting
	state := map[string]any{"incremental": incremental}
	inputProps, err := r.resolveTemplatedProps(ctx, self, model.Spec.InputConnector, model.Spec.InputProperties.AsMap(), state)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	outputProps, err := r.resolveTemplatedProps(ctx, self, model.Spec.OutputConnector, model.Spec.OutputProperties.AsMap(), state)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	opts := &drivers.ModelExecuteOptions{
		ModelName:        self.Meta.Name.Name,
		Env:              executorEnv,
		PreviousResult:   prevResult,
		Incremental:      incremental,
		InputConnector:   model.Spec.InputConnector,
		InputProperties:  inputProps,
		OutputConnector:  model.Spec.OutputConnector,
		OutputProperties: outputProps,
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

	// Build the output
	execRes, execErr := executor.Run(ctx, opts)
	if execErr != nil {
		execErr = fmt.Errorf("failed to build output: %w", execErr)
	}

	// If the build succeeded, update the model's state
	var update bool
	if execErr == nil {
		resultProps, err := structpb.NewStruct(execRes.Properties)
		if err != nil {
			execErr = fmt.Errorf("executor returned non-serializable output properties: %w", err)
		} else {
			model.State.ExecutorConnector = executorConnector
			model.State.ResultConnector = execRes.Connector
			model.State.ResultProperties = resultProps
			model.State.ResultTable = execRes.Table
			model.State.SpecHash = hash
			model.State.RefreshedOn = timestamppb.Now()
			update = true
		}
	}

	// If the build failed, clear the state only if we're not staging changes
	if execErr != nil {
		if !executorEnv.StageChanges {
			model.State.ExecutorConnector = ""
			model.State.ResultConnector = ""
			model.State.ResultProperties = nil
			model.State.ResultTable = ""
			model.State.SpecHash = ""
			model.State.RefreshedOn = nil
			update = true
		}
	}

	// Update state
	if update {
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// If the context was cancelled, we return now since we don't want to clear the trigger or set a next refresh time.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: errors.Join(ctx.Err(), execErr)}
	}

	// Reset spec.Trigger
	if model.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
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

// executionSpecHash computes a hash of only those model properties that impact execution.
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

		// Incorporate the ref's state info in the hash if and only if we are supposed to trigger when a ref has refreshed (denoted by RefreshSchedule.RefUpdate).
		if spec.RefreshSchedule != nil && spec.RefreshSchedule.RefUpdate {
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

// setTriggerFalse sets the model's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *ModelReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
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

// acquireExecutor acquires a ModelExecutor capable of executing a model with the given execution options.
func (r *ModelReconciler) acquireExecutor(ctx context.Context, opts *drivers.ModelExecuteOptions) (string, drivers.ModelExecutor, func(), error) {
	ic, release, err := r.C.AcquireConn(ctx, opts.InputConnector)
	if err != nil {
		return "", nil, nil, err
	}

	if e, ok := ic.AsModelExecutor(); ok {
		ok, err := e.Supports(ctx, opts)
		if err != nil {
			release()
			return "", nil, nil, err
		}

		if ok {
			return opts.InputConnector, e, release, nil
		}
	}

	release()

	if opts.InputConnector == opts.OutputConnector {
		return "", nil, nil, fmt.Errorf("connector %q is not capable of executing models", opts.InputConnector)
	}

	oc, release, err := r.C.AcquireConn(ctx, opts.OutputConnector)
	if err != nil {
		return "", nil, nil, err
	}

	if e, ok := oc.AsModelExecutor(); ok {
		ok, err := e.Supports(ctx, opts)
		if err != nil {
			release()
			return "", nil, nil, err
		}

		if ok {
			return opts.OutputConnector, e, release, nil
		}
	}

	release()

	return "", nil, nil, fmt.Errorf("cannot execute model: input connector %q and output connector %q are not compatible", opts.InputConnector, opts.OutputConnector)
}

// newExecutorEnv makes ModelExecutorEnv configured using the current instance.
func (r *ModelReconciler) newExecutorEnv(ctx context.Context) (*drivers.ModelExecutorEnv, error) {
	cfg, err := r.C.Runtime.InstanceConfig(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access instance config: %w", err)
	}

	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to access repo: %w", err)
	}
	defer release()

	return &drivers.ModelExecutorEnv{
		AllowHostAccess:    r.C.Runtime.AllowHostAccess(),
		RepoRoot:           repo.Root(),
		StageChanges:       cfg.StageChanges,
		DefaultMaterialize: cfg.ModelDefaultMaterialize,
		AcquireConnector:   r.C.AcquireConn,
	}, nil
}

// resolveTemplatedProps resolves template tags in strings nested in the provided props.
// Passing a connector is optional. If a connector is provided, it will be used to inform how values are escaped.
func (r *ModelReconciler) resolveTemplatedProps(ctx context.Context, self *runtimev1.Resource, connector string, props, state map[string]any) (map[string]any, error) {
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

// analyzeTemplatedVariables analyzes strings nested in the provided props for template tags that reference variables.
// The keys of the returned map are the variable names, and the values are empty strings (optimization to enable re-using the map in upstream code).
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
