package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	aiEvalExecutionHistoryLimit = 10
	aiEvalDefaultTimeout        = 30 * time.Minute
	aiEvalDefaultCaseTimeout    = 5 * time.Minute
	aiEvalDefaultConcurrency    = 2
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindAIEval, newAIEvalReconciler)
}

// AIEvalReconciler reconciles an AIEval.
// Reconciles are validate-only unless the ephemeral Spec.Trigger is set:
// eval runs make LLM calls per case, so they must never happen as a side effect of a file save or a restart.
// Triggered runs execute the selected cases, grade them, and store the results in the resource state's execution history.
type AIEvalReconciler struct {
	C *runtime.Controller
}

func newAIEvalReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &AIEvalReconciler{C: c}, nil
}

func (r *AIEvalReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *AIEvalReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetAiEval()
	b := to.GetAiEval()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *AIEvalReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetAiEval()
	b := to.GetAiEval()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *AIEvalReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetAiEval().State = &runtimev1.AIEvalState{}
	return nil
}

func (r *AIEvalReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	eval := self.GetAiEval()
	if eval == nil {
		return runtime.ReconcileResult{Err: errors.New("not an AI eval")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// If CurrentExecution is not nil, the previous run was interrupted (crash or cancellation).
	// Finalize it into the history so the UI never shows a stuck run and re-triggering works.
	if eval.State.CurrentExecution != nil {
		finalizeInterruptedAIEvalExecution(eval.State.CurrentExecution)
		err = r.popCurrentExecution(ctx, self, eval)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Surface broken refs (e.g. an errored metrics view a case expects) as a reconcile error even when not running.
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Validate-only unless explicitly triggered. This is the guarantee that plain saves never spend LLM tokens.
	// (A restart during a triggered run re-runs it, same as models; the crash-recovery block archives the interrupted run.)
	if !eval.Spec.Trigger {
		return runtime.ReconcileResult{}
	}

	// Eval runs execute with security checks skipped, so they only run in development contexts:
	// Rill Developer (host access, auth disabled anyway) or a cloud dev deployment (environment "dev",
	// only reachable by project editors). On a production deployment, skipped checks would bypass
	// row-level security and the transcripts would land in ownerless sessions readable by any user.
	// Production runs need run-as attributes first.
	if !r.C.Runtime.AllowHostAccess() {
		inst, instErr := r.C.Runtime.Instance(ctx, r.C.InstanceID)
		if instErr != nil {
			return runtime.ReconcileResult{Err: instErr}
		}
		if inst.Environment != "dev" {
			err = r.setTriggerFalse(ctx, n)
			if err != nil {
				return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
			}
			return runtime.ReconcileResult{Err: errors.New("AI evals can only run in development environments (Rill Developer or a cloud dev deployment)")}
		}
	}

	executeErr := r.executeAll(ctx, self, eval)

	// If the reconcile was canceled mid-run, return without clearing the trigger:
	// the controller re-invokes us and the crash-recovery block above finalizes the interrupted run.
	// (A cancellation caused by an explicit cancel trigger has already cleared the trigger spec.)
	if errors.Is(executeErr, context.Canceled) {
		return runtime.ReconcileResult{Err: executeErr}
	}

	// Clear the ad-hoc trigger
	err = r.setTriggerFalse(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
	}

	return runtime.ReconcileResult{Err: executeErr}
}

func (r *AIEvalReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetAiEval() == nil {
		return nil, fmt.Errorf("not an AI eval resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// executeAll runs the triggered cases with bounded concurrency and stores the graded results in the state.
// The state is updated after every case transition so the UI can render live progress through the resource watch.
func (r *AIEvalReconciler) executeAll(ctx context.Context, self *runtimev1.Resource, eval *runtimev1.AIEval) error {
	// Resolve the selected cases
	cases := selectAIEvalCases(eval.Spec.Cases, eval.Spec.TriggerCases)
	for _, unknown := range unknownAIEvalCases(eval.Spec.Cases, eval.Spec.TriggerCases) {
		r.C.Logger.Warn("Skipped unknown AI eval case in trigger", zap.String("case", unknown), observability.ZapCtx(ctx))
	}
	if len(cases) == 0 {
		return errors.New("no matching cases to run")
	}

	// Initialize the execution with all selected cases pending, so the UI shows the roster immediately.
	execution := &runtimev1.AIEvalExecution{
		Adhoc:        true,
		TriggerCases: eval.Spec.TriggerCases,
		StartedOn:    timestamppb.Now(),
	}
	for _, c := range cases {
		execution.CaseResults = append(execution.CaseResults, &runtimev1.AIEvalCaseResult{
			CaseName: c.Name,
			Verdict:  runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_PENDING,
		})
	}
	eval.State.CurrentExecution = execution

	// mu guards all mutations of the execution and state updates, since cases run concurrently.
	var mu sync.Mutex
	updateState := func() error {
		return r.C.UpdateState(ctx, self.Meta.Name, self)
	}
	mu.Lock()
	err := updateState()
	mu.Unlock()
	if err != nil {
		return err
	}

	timeout := aiEvalDefaultTimeout
	if eval.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(eval.Spec.TimeoutSeconds) * time.Second
	}
	caseTimeout := aiEvalDefaultCaseTimeout
	if eval.Spec.CaseTimeoutSeconds > 0 {
		caseTimeout = time.Duration(eval.Spec.CaseTimeoutSeconds) * time.Second
	}
	concurrency := aiEvalDefaultConcurrency
	if eval.Spec.Concurrency > 0 {
		concurrency = int(eval.Spec.Concurrency)
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Claims for eval runs. Evals currently only run in Rill Developer, where auth is disabled;
	// running them in Rill Cloud would require run-as attributes on the spec (out of scope for now).
	claims := &runtime.SecurityClaims{SkipChecks: true}

	runner := ai.NewRunner(r.C.Runtime, r.C.Runtime.Activity())

	grp, grpCtx := errgroup.WithContext(runCtx)
	grp.SetLimit(concurrency)
	for i, c := range cases {
		caseResult := execution.CaseResults[i]
		grp.Go(func() error {
			// Leave the case pending if the run has been canceled or timed out; it's marked skipped during finalization.
			if grpCtx.Err() != nil {
				return nil
			}

			mu.Lock()
			caseResult.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_RUNNING
			caseResult.StartedOn = timestamppb.Now()
			err := updateState()
			mu.Unlock()
			if err != nil {
				return err
			}

			caseCtx, caseCancel := context.WithTimeout(grpCtx, caseTimeout)
			defer caseCancel()
			res, err := runner.RunEvalCase(caseCtx, &ai.EvalCaseRunOptions{
				InstanceID:   r.C.InstanceID,
				Claims:       claims,
				Case:         c,
				SuiteNotes:   eval.Spec.Notes,
				SuiteExplore: eval.Spec.Explore,
				Agent:        eval.Spec.Agent,
			})
			// Report per-case timeouts with a readable reason instead of a raw context error.
			timedOut := errors.Is(caseCtx.Err(), context.DeadlineExceeded) && grpCtx.Err() == nil

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				caseResult.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR
				caseResult.FailureReasons = []string{err.Error()}
			} else {
				caseResult.Verdict = res.Verdict
				caseResult.FailureReasons = res.FailureReasons
				caseResult.JudgeReasoning = res.JudgeReasoning
				caseResult.SessionId = res.SessionID
				caseResult.AnswerPreview = res.AnswerPreview
			}
			if timedOut && caseResult.Verdict == runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR {
				caseResult.FailureReasons = []string{fmt.Sprintf("case timed out after %s", caseTimeout)}
			}
			caseResult.FinishedOn = timestamppb.Now()
			// Tolerate state-update failures caused by cancellation; finalization handles persistence.
			if err := updateState(); err != nil && ctx.Err() == nil {
				return err
			}
			return nil
		})
	}
	grpErr := grp.Wait()

	// Finalize: mark cases that never ran as skipped, compute counts, and pop into history.
	if runCtx.Err() != nil && ctx.Err() == nil {
		execution.ErrorMessage = fmt.Sprintf("run timed out after %s", timeout)
	}
	for _, cr := range execution.CaseResults {
		if cr.Verdict == runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_PENDING {
			cr.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_SKIPPED
			execution.Canceled = true
		}
	}
	setAIEvalExecutionCounts(execution)
	execution.FinishedOn = timestamppb.Now()

	err = r.popCurrentExecution(ctx, self, eval)
	if err != nil {
		if ctx.Err() != nil {
			// The reconcile was canceled; leave CurrentExecution dangling for the crash-recovery path.
			return ctx.Err()
		}
		return err
	}

	if grpErr != nil {
		return grpErr
	}
	return ctx.Err()
}

// popCurrentExecution moves the current execution into the execution history, and persists the updated state.
// At a certain limit, it trims old executions from the history to prevent it from growing unboundedly.
func (r *AIEvalReconciler) popCurrentExecution(ctx context.Context, self *runtimev1.Resource, eval *runtimev1.AIEval) error {
	if eval.State.CurrentExecution == nil {
		panic(fmt.Errorf("attempting to pop current execution when there is none"))
	}

	eval.State.ExecutionHistory = slices.Insert(eval.State.ExecutionHistory, 0, eval.State.CurrentExecution)
	eval.State.CurrentExecution = nil
	eval.State.ExecutionCount++

	if len(eval.State.ExecutionHistory) > aiEvalExecutionHistoryLimit {
		eval.State.ExecutionHistory = eval.State.ExecutionHistory[:aiEvalExecutionHistoryLimit]
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// setTriggerFalse clears the eval's ephemeral trigger properties.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *AIEvalReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return err
	}

	eval := self.GetAiEval()
	if eval == nil {
		return fmt.Errorf("not an AI eval")
	}

	eval.Spec.Trigger = false
	eval.Spec.TriggerCases = nil
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// finalizeInterruptedAIEvalExecution marks an interrupted execution's unfinished cases and closes it out.
func finalizeInterruptedAIEvalExecution(execution *runtimev1.AIEvalExecution) {
	for _, cr := range execution.CaseResults {
		switch cr.Verdict {
		case runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_RUNNING:
			cr.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR
			cr.FailureReasons = []string{"run was interrupted"}
			cr.FinishedOn = timestamppb.Now()
		case runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_PENDING:
			cr.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_SKIPPED
		}
	}
	if execution.ErrorMessage == "" {
		execution.Canceled = true
	}
	setAIEvalExecutionCounts(execution)
	if execution.FinishedOn == nil {
		execution.FinishedOn = timestamppb.Now()
	}
}

func setAIEvalExecutionCounts(execution *runtimev1.AIEvalExecution) {
	execution.PassedCount = 0
	execution.FailedCount = 0
	execution.ErroredCount = 0
	for _, cr := range execution.CaseResults {
		switch cr.Verdict {
		case runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_PASS:
			execution.PassedCount++
		case runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_FAIL:
			execution.FailedCount++
		case runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR:
			execution.ErroredCount++
		}
	}
}

// selectAIEvalCases returns the spec cases matching the trigger's case selection. An empty selection means all cases.
func selectAIEvalCases(cases []*runtimev1.AIEvalCase, selection []string) []*runtimev1.AIEvalCase {
	if len(selection) == 0 {
		return cases
	}
	selected := make(map[string]bool, len(selection))
	for _, s := range selection {
		selected[strings.ToLower(s)] = true
	}
	var res []*runtimev1.AIEvalCase
	for _, c := range cases {
		if selected[strings.ToLower(c.Name)] {
			res = append(res, c)
		}
	}
	return res
}

// unknownAIEvalCases returns trigger case selections that don't match any spec case.
func unknownAIEvalCases(cases []*runtimev1.AIEvalCase, selection []string) []string {
	if len(selection) == 0 {
		return nil
	}
	known := make(map[string]bool, len(cases))
	for _, c := range cases {
		known[strings.ToLower(c.Name)] = true
	}
	var res []string
	for _, s := range selection {
		if !known[strings.ToLower(s)] {
			res = append(res, s)
		}
	}
	return res
}
