package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const reportExecutionHistoryLimit = 10

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindReport, newReportReconciler)
}

type ReportReconciler struct {
	C *runtime.Controller
}

func newReportReconciler(c *runtime.Controller) runtime.Reconciler {
	return &ReportReconciler{C: c}
}

func (r *ReportReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ReportReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ReportReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetReport()
	b := to.GetReport()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ReportReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetReport().State = &runtimev1.ReportState{}
	return nil
}

func (r *ReportReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	rep := self.GetReport()
	if rep == nil {
		return runtime.ReconcileResult{Err: errors.New("not a report")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// NOTE: refs not supported for reports.
	// Not supporting them simplifies report spec generation, improves performance (there may be many reports),
	// and it's anyway acceptable (maybe even expected) that a report fails with an execution error if the underlying operation errors.

	// If CurrentExecution is not nil, a catastrophic failure occurred during the last execution.
	// Clean up to ensure CurrentExecution is nil.
	if rep.State.CurrentExecution != nil {
		rep.State.CurrentExecution.ErrorMessage = "Internal: report execution was interrupted unexpectedly"
		rep.State.CurrentExecution.FinishedOn = timestamppb.Now()
		err = r.popCurrentExecution(ctx, self, rep)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Determine whether to trigger
	adhocTrigger := rep.Spec.Trigger
	scheduleTrigger := rep.State.NextRunOn != nil && !rep.State.NextRunOn.AsTime().After(time.Now())

	// If not triggering now, update NextRunOn and retrigger when it falls due
	if !adhocTrigger && !scheduleTrigger {
		err = r.updateNextRunOn(ctx, self, rep)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
		if rep.State.NextRunOn != nil {
			return runtime.ReconcileResult{Retrigger: rep.State.NextRunOn.AsTime()}
		}
		return runtime.ReconcileResult{}
	}

	// Determine time to evaluate the report relative to.
	// We use the "clean" scheduled time unless it's an ad-hoc trigger.
	var reportTime *timestamppb.Timestamp
	if scheduleTrigger && !adhocTrigger {
		reportTime = rep.State.NextRunOn
	} else {
		reportTime = timestamppb.Now()
	}

	// Create new execution and save in State.CurrentExecution
	rep.State.CurrentExecution = &runtimev1.ReportExecution{
		Adhoc:      rep.Spec.Trigger,
		ReportTime: reportTime,
		StartedOn:  timestamppb.Now(),
	}
	err = r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Execute report
	dirtyErr, reportErr := r.sendReport(ctx, self, rep, reportTime.AsTime())

	// Set execution error and determine whether to retry.
	// We're only going to retry on non-dirty cancellations.
	retry := false
	if reportErr != nil {
		if errors.Is(reportErr, context.Canceled) {
			if dirtyErr {
				rep.State.CurrentExecution.ErrorMessage = "Report run was interrupted after some emails were sent. The report will not automatically retry."
			} else {
				retry = true
				rep.State.CurrentExecution.ErrorMessage = "Report run was interrupted. It will automatically retry."
			}
		} else {
			rep.State.CurrentExecution.ErrorMessage = fmt.Sprintf("Report run failed: %v", reportErr.Error())
		}
		reportErr = fmt.Errorf("Last report run failed with error: %v", reportErr.Error())
	}

	// Commit CurrentExecution to history
	rep.State.CurrentExecution.FinishedOn = timestamppb.Now()
	err = r.popCurrentExecution(ctx, self, rep)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// If we want to retry, exit without advancing NextRunOn or clearing spec.Trigger.
	// NOTE: We don't set Retrigger here because we'll leave re-scheduling to whatever cancelled the reconciler.
	if retry {
		return runtime.ReconcileResult{Err: reportErr}
	}

	// Advance NextRunOn
	err = r.updateNextRunOn(ctx, self, rep)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Clear ad-hoc trigger
	if rep.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to clear trigger: %w", err)}
		}
	}

	// Done
	if rep.State.NextRunOn != nil {
		return runtime.ReconcileResult{Err: reportErr, Retrigger: rep.State.NextRunOn.AsTime()}
	}
	return runtime.ReconcileResult{Err: reportErr}
}

// updateNextRunOn evaluates the report's schedule relative to the current time, and updates the NextRunOn state accordingly.
// If the schedule is nil, it will set NextRunOn to nil.
func (r *ReportReconciler) updateNextRunOn(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report) error {
	next, err := nextRefreshTime(time.Now(), rep.Spec.RefreshSchedule)
	if err != nil {
		return err
	}

	var curr time.Time
	if rep.State.NextRunOn != nil {
		curr = rep.State.NextRunOn.AsTime()
	}

	if next == curr {
		return nil
	}

	if next.IsZero() {
		rep.State.NextRunOn = nil
	} else {
		rep.State.NextRunOn = timestamppb.New(next)
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// popCurrentExecution moves the current execution into the execution history, and persists the updated state.
// At a certain limit, it trims old executions from the history to prevent it from growing unboundedly.
func (r *ReportReconciler) popCurrentExecution(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report) error {
	if rep.State.CurrentExecution == nil {
		panic(fmt.Errorf("attempting to pop current execution when there is none"))
	}

	rep.State.ExecutionHistory = slices.Insert(rep.State.ExecutionHistory, 0, rep.State.CurrentExecution)
	rep.State.CurrentExecution = nil

	if len(rep.State.ExecutionHistory) > reportExecutionHistoryLimit {
		rep.State.ExecutionHistory = rep.State.ExecutionHistory[:reportExecutionHistoryLimit]
	}

	return r.C.UpdateState(ctx, self.Meta.Name, self)
}

// setTriggerFalse sets the report's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *ReportReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, false)
	if err != nil {
		return err
	}

	rep := self.GetReport()
	if rep == nil {
		return fmt.Errorf("not a report")
	}

	rep.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// sendReport composes and sends the actual report to the configured recipients.
// It returns true if an error occurred after some or all emails were sent.
func (r *ReportReconciler) sendReport(ctx context.Context, self *runtimev1.Resource, rep *runtimev1.Report, t time.Time) (bool, error) {
	// TODO: Make the magic happen
	r.C.Logger.Info("Sending report", "report", self.Meta.Name.Name, "time", t)

	return false, ctx.Err()
}
