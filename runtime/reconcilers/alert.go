package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const alertExecutionHistoryLimit = 10

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
		a.State.CurrentExecution.FinishedOn = timestamppb.Now()
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
	trigger := adhocTrigger || hashTrigger || scheduleTrigger

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

	// Determine time to evaluate the alert relative to.
	// We use the "clean" scheduled time unless it's an ad-hoc trigger.
	// TODO: We can incorporate watermarks here.
	var alertTime *timestamppb.Timestamp
	if scheduleTrigger && !adhocTrigger && !hashTrigger {
		alertTime = a.State.NextRunOn
	} else {
		alertTime = timestamppb.Now()
	}

	// Create new execution and save in State.CurrentExecution
	a.State.CurrentExecution = &runtimev1.AlertExecution{
		Adhoc:     adhocTrigger,
		AlertTime: alertTime,
		StartedOn: timestamppb.Now(),
	}
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Execute report
	res, dirtyErr, alertErr := r.checkAlert(ctx, self, a, alertTime.AsTime())

	// Update CurrentExecution
	a.State.CurrentExecution.Result = res
	a.State.CurrentExecution.FinishedOn = timestamppb.Now()

	// If the check failed, set CurrentExecution.Result to an error and determine whether to retry.
	// We're only going to retry on non-dirty cancellations.
	retry := false
	if alertErr != nil {
		var msg string
		if errors.Is(alertErr, context.Canceled) {
			if dirtyErr {
				msg = "Alert check was interrupted after some emails were sent. The alert will not automatically retry."
			} else {
				retry = true
				msg = "Alert check was interrupted. It will automatically retry."
			}
		} else {
			msg = fmt.Sprintf("Alert check failed: %v", alertErr.Error())
		}
		a.State.CurrentExecution.Result = &runtimev1.AssertionResult{
			Status:       runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
			ErrorMessage: msg,
		}
		alertErr = fmt.Errorf("Last alert check failed with error: %v", alertErr.Error())
	}

	// Log it
	if alertErr != nil {
		r.C.Logger.Error("Alert check failed", "alert", self.Meta.Name, "error", alertErr.Error())
	}

	// Commit CurrentExecution to history
	err = r.popCurrentExecution(ctx, self, a)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// If we want to retry, exit without updating any other trigger-related state.
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

		// Write state version (doesn't matter how the spec or meta has changed, only if/when state changes)
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

// checkAlert runs the alert query and maybe sends emails.
// It returns true if an error occurred after some or all emails were sent.
func (r *AlertReconciler) checkAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, t time.Time) (*runtimev1.AssertionResult, bool, error) {
	r.C.Logger.Info("Checking alert", "alert", self.Meta.Name.Name, "alert_time", t)

	// Check refs - stop if any of them are invalid
	err := checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR, ErrorMessage: err.Error()}, false, nil
	}

	// TODO: Implement

	return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS}, false, nil
}
