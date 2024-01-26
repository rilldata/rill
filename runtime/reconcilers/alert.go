package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const alertExecutionHistoryLimit = 25

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindAlert, newAlertReconciler)
}

type AlertReconciler struct {
	C *runtime.Controller
}

func newAlertReconciler(c *runtime.Controller) runtime.Reconciler {
	return &AlertReconciler{C: c}
}

func (r *AlertReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *AlertReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetAlert()
	b := to.GetAlert()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *AlertReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetAlert()
	b := to.GetAlert()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *AlertReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetAlert().State = &runtimev1.AlertState{}
	return nil
}

func (r *AlertReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	a := self.GetAlert()
	if a == nil {
		return runtime.ReconcileResult{Err: errors.New("not an alert")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// If CurrentExecution is not nil, a catastrophic failure occurred during the last execution.
	// Clean up to ensure CurrentExecution is nil.
	if a.State.CurrentExecution != nil {
		a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
			Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
			ErrorMessage: "Internal: alert execution was interrupted unexpectedly",
		}
		err = r.popCurrentExecution(ctx, self, a)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Use a hash of execution-related fields from the spec to determine if something has changed
	hash, err := r.executionSpecHash(ctx, a.Spec, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Determine whether to trigger
	adhocTrigger := a.Spec.Trigger
	hashTrigger := a.State.SpecHash != hash
	scheduleTrigger := a.State.NextRunOn != nil && !a.State.NextRunOn.AsTime().After(time.Now())
	disabled := a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.Disable
	trigger := !disabled && (adhocTrigger || hashTrigger || scheduleTrigger)

	// If not triggering now, update NextRunOn and retrigger when it falls due
	if !trigger {
		err = r.updateNextRunOn(ctx, self, a)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if a.State.NextRunOn != nil {
			return runtime.ReconcileResult{Retrigger: a.State.NextRunOn.AsTime()}
		}
		return runtime.ReconcileResult{}
	}

	// Evaluate the trigger time of the alert. If triggered by schedule, we use the "clean" scheduled time.
	// Note: Correction for watermarks and intervals is done in checkAlert.
	var triggerTime time.Time
	if scheduleTrigger && !adhocTrigger && !hashTrigger {
		triggerTime = a.State.NextRunOn.AsTime()
	} else {
		triggerTime = time.Now()
	}

	// Run alert queries and send emails
	retry, alertErr := r.checkAlert(ctx, self, a, triggerTime, adhocTrigger)

	// If we were cancelled, exit without updating any other trigger-related state.
	// NOTE: We don't set Retrigger here because we'll leave re-scheduling to whatever cancelled the reconciler.
	if retry {
		return runtime.ReconcileResult{Err: alertErr}
	}

	// Advance NextRunOn
	err = r.updateNextRunOn(ctx, self, a)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Clear ad-hoc trigger
	if a.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
		}
	}

	// Update spec hash
	if hashTrigger {
		a.State.SpecHash = hash
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Done
	if a.State.NextRunOn != nil {
		return runtime.ReconcileResult{Err: alertErr, Retrigger: a.State.NextRunOn.AsTime()}
	}
	return runtime.ReconcileResult{Err: alertErr}
}

// executionSpecHash computes a hash of the alert properties that impact execution.
func (r *AlertReconciler) executionSpecHash(ctx context.Context, spec *runtimev1.AlertSpec, refs []*runtimev1.ResourceName) (string, error) {
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

		// Incorporate the ref's state version in the hash if and only if we are supposed to trigger when a ref has refreshed (denoted by RefreshSchedule.RefUpdate).
		if spec.RefreshSchedule != nil && spec.RefreshSchedule.RefUpdate {
			// Note: Only writing the state version to the hash, not spec version, because it doesn't matter whether the spec/meta changes, only whether the state changes.
			r, err := r.C.Get(ctx, ref, false)
			var stateVersion int64
			if err == nil {
				stateVersion = r.Meta.StateVersion
			} else {
				stateVersion = -1
			}
			err = binary.Write(hash, binary.BigEndian, stateVersion)
			if err != nil {
				return "", err
			}
		}
	}

	_, err := hash.Write([]byte(spec.QueryName))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.QueryArgsJson))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.GetQueryForUserId()))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.GetQueryForUserEmail()))
	if err != nil {
		return "", err
	}

	// TODO: Add spec.QueryForAttributes

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// updateNextRunOn evaluates the alert's schedule relative to the current time, and updates the NextRunOn state accordingly.
// If the schedule is nil, it will set NextRunOn to nil.
func (r *AlertReconciler) updateNextRunOn(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert) error {
	next, err := nextRefreshTime(time.Now(), a.Spec.RefreshSchedule)
	if err != nil {
		return err
	}

	var curr time.Time
	if a.State.NextRunOn != nil {
		curr = a.State.NextRunOn.AsTime()
	}

	if next == curr {
		return nil
	}

	if next.IsZero() {
		a.State.NextRunOn = nil
	} else {
		a.State.NextRunOn = timestamppb.New(next)
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// popCurrentExecution moves the current execution into the execution history, and persists the updated state.
// At a certain limit, it trims old executions from the history to prevent it from growing unboundedly.
func (r *AlertReconciler) popCurrentExecution(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert) error {
	if a.State.CurrentExecution == nil {
		panic(fmt.Errorf("attempting to pop current execution when there is none"))
	}

	a.State.CurrentExecution.FinishedOn = timestamppb.Now()
	a.State.ExecutionHistory = slices.Insert(a.State.ExecutionHistory, 0, a.State.CurrentExecution)
	a.State.CurrentExecution = nil

	if len(a.State.ExecutionHistory) > alertExecutionHistoryLimit {
		a.State.ExecutionHistory = a.State.ExecutionHistory[:alertExecutionHistoryLimit]
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// setTriggerFalse sets the alert's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *AlertReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, false)
	if err != nil {
		return err
	}

	a := self.GetAlert()
	if a == nil {
		return fmt.Errorf("not an alert")
	}

	a.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// checkAlert runs queries and (maybe) sends emails for the alert. It also adds entries to a.State.ExecutionHistory.
// By default, an alert is checked once for the current watermark, but if a.Spec.IntervalsIsoDuration is set, it will be checked *for each* interval that has elapsed since the previous execution watermark.
func (r *AlertReconciler) checkAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, triggerTime time.Time, adhocTrigger bool) (bool, error) {
	// Check refs
	executionErr := checkRefs(ctx, r.C, self.Meta.Refs)

	// Evaluate watermark unless refs check failed.
	watermark := triggerTime
	if executionErr == nil && a.Spec.WatermarkInherit {
		t, ok, err := r.computeInheritedWatermark(ctx, self.Meta.Refs)
		if err != nil {
			executionErr = err
		} else if ok {
			watermark = t
		}
		// If !ok, no watermark could be computed. So we'll just use the trigger time.
	}

	// Evaluate previous watermark (will be equal to watermark if there was no previous execution)
	previousWatermark := watermark
	if executionErr == nil && a.State.ExecutionHistory != nil {
		for _, e := range a.State.ExecutionHistory {
			previousWatermark = e.ExecutionTime.AsTime()
			break
		}
	}

	// Evaluate and invoke intervals and add to execution history
	dirty := false
	if executionErr == nil {
		log.Printf("HERE: %v", previousWatermark)
		// TODO: Evaluate intervals
		// TODO: Call checkSingleAlert for each interval
		// TODO: Add to execution history.
		// TODO: Update executionErr
	}

	// If executionErr is nil, we're done.
	if executionErr == nil {
		return false, nil
	}

	// If executionErr is a non-dirty cancellation, we're also done.
	if errors.Is(executionErr, context.Canceled) && !dirty {
		return false, executionErr
	}

	// There was an execution error. Add it to the execution history.
	if a.State.CurrentExecution == nil {
		// CurrentExecution will only be nil if we never made it to the point of checking the alert query.
		a.State.CurrentExecution = &runtimev1.AlertExecution{
			Adhoc:         adhocTrigger,
			ExecutionTime: nil, // TODO: What to put here? triggerTime? watermark? nil? (the most recently tried interval?)
			StartedOn:     timestamppb.Now(),
		}
	}
	a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
		Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
		ErrorMessage: executionErr.Error(),
	}
	a.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err := r.popCurrentExecution(ctx, self, a)
	if err != nil {
		return false, err
	}

	return !dirty, executionErr
}

// checkAlert runs the alert query and maybe sends emails.
// It returns true if an error occurred after some or all emails were sent.
func (r *AlertReconciler) checkSingleAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, executionTime time.Time, adhocTrigger bool) (bool, error) {
	r.C.Logger.Info("Checking alert", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))

	// Create new execution and save in State.CurrentExecution
	a.State.CurrentExecution = &runtimev1.AlertExecution{
		Adhoc:         adhocTrigger,
		ExecutionTime: timestamppb.New(executionTime),
		StartedOn:     timestamppb.Now(),
	}
	err := r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return false, err
	}

	// Query and email
	res, dirtyErr, alertErr := r.executeSingleAlert(ctx, self, a, executionTime)

	// If there was no error, we're done.
	if alertErr == nil {
		a.State.CurrentExecution.Result = res
		a.State.CurrentExecution.FinishedOn = timestamppb.Now()
		err = r.popCurrentExecution(ctx, self, a)
		if err != nil {
			return true, err
		}
		return false, nil
	}

	// If the error is a non-dirty cancellation, we're also done (will be retried)
	if errors.Is(alertErr, context.Canceled) && !dirtyErr {
		// Pretend the CurrentExecution never happened
		a.State.CurrentExecution = nil
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return false, err
		}

		return false, alertErr
	}

	// There was an error. Add it to the execution history.
	if errors.Is(alertErr, context.Canceled) {
		alertErr = fmt.Errorf("Alert check was interrupted after some emails were sent. The alert will not automatically retry.")
	} else {
		alertErr = fmt.Errorf("Alert check failed: %v", alertErr.Error())
	}
	a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
		Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
		ErrorMessage: alertErr.Error(),
	}

	// Commit CurrentExecution to history
	err = r.popCurrentExecution(ctx, self, a)
	if err != nil {
		return true, err
	}

	return dirtyErr, alertErr
}

// executeSingleAlert runs the alert query and maybe sends emails.
// It returns true if an error occurred after some or all emails were sent.
func (r *AlertReconciler) executeSingleAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, executionTime time.Time) (*runtimev1.AssertionResult, bool, error) {
	// TODO: Implement
	r.C.Logger.Info("Triggered alert", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))
	return nil, false, nil
}

// computeInheritedWatermark computes the inherited watermark for the alert.
// It returns false if the watermark could not be computed.
func (r *AlertReconciler) computeInheritedWatermark(ctx context.Context, refs []*runtimev1.ResourceName) (time.Time, bool, error) {
	var t time.Time
	for _, ref := range refs {
		rs, err := r.C.Get(ctx, ref, false)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("failed to get ref %v: %w", ref, err)
		}

		// Currently only metrics views have watermarks
		mv := rs.GetMetricsView()
		if mv == nil {
			continue
		}

		// TODO: Query for the watermark
	}

	return t, !t.IsZero(), nil
}
