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
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/queries"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const alertExecutionHistoryLimit = 25

const alertDefaultIntervalsLimit = 25

const alertQueryPriority = 1

const alertCheckDefaultTimeout = 5 * time.Minute

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

	// Exit early if disabled
	if a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// TODO: Comment
	specHash, err := r.executionSpecHash(ctx, a.Spec, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}
	refsHash, err := r.refsStateHash(ctx, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Determine whether to trigger
	adhocTrigger := a.Spec.Trigger
	specHashTrigger := a.State.SpecHash != specHash
	refsHashTrigger := a.State.RefsHash != refsHash
	scheduleTrigger := a.State.NextRunOn != nil && !a.State.NextRunOn.AsTime().After(time.Now())
	trigger := adhocTrigger || specHashTrigger || refsHashTrigger || scheduleTrigger

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

	// If the spec hash changed, clear all alert state
	if specHashTrigger {
		a.State.SpecHash = specHash
		a.State.RefsHash = ""
		a.State.NextRunOn = nil
		a.State.CurrentExecution = nil
		a.State.ExecutionHistory = nil
		a.State.ExecutionCount = 0
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Evaluate the trigger time of the alert. If triggered by schedule, we use the "clean" scheduled time.
	// Note: Correction for watermarks and intervals is done in checkAlert.
	var triggerTime time.Time
	if scheduleTrigger && !adhocTrigger && !specHashTrigger && !refsHashTrigger {
		triggerTime = a.State.NextRunOn.AsTime()
	} else {
		triggerTime = time.Now()
	}

	// Run alert queries and send emails
	retry, alertErr := r.executeAlert(ctx, self, a, triggerTime, adhocTrigger)

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

	// Update refs hash
	if refsHashTrigger {
		a.State.RefsHash = refsHash
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
// NOTE: Unlike other resources, we don't include the refs' state version in the hash since it's managed separately using refsStateHash.
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
	}

	if spec.RefreshSchedule != nil {
		_, err := hash.Write([]byte(spec.RefreshSchedule.TimeZone))
		if err != nil {
			return "", err
		}
	}

	err := binary.Write(hash, binary.BigEndian, spec.WatermarkInherit)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.IntervalsIsoDuration))
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.IntervalsCheckUnclosed)
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.QueryName))
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

// refsStateHash computes a hash of the refs and their state versions.
func (r *AlertReconciler) refsStateHash(ctx context.Context, refs []*runtimev1.ResourceName) (string, error) {
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
	a.State.ExecutionCount++

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

// executeAlert runs queries and (maybe) sends emails for the alert. It also adds entries to a.State.ExecutionHistory.
// By default, an alert is checked once for the current watermark, but if a.Spec.IntervalsIsoDuration is set, it will be checked *for each* interval that has elapsed since the previous execution watermark.
func (r *AlertReconciler) executeAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, triggerTime time.Time, adhocTrigger bool) (bool, error) {
	// Enforce timeout
	timeout := alertCheckDefaultTimeout
	if a.Spec.TimeoutSeconds > 0 {
		timeout = time.Duration(a.Spec.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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
		// If !ok, no watermark could be computed. So we'll just stick to triggerTime.
	}

	// Evaluate and invoke intervals and add to execution history
	dirty := false
	if executionErr == nil {
		// Evaluate previous watermark (if any)
		var previousWatermark time.Time
		for _, e := range a.State.ExecutionHistory {
			if e.ExecutionTime != nil {
				previousWatermark = e.ExecutionTime.AsTime()
				break
			}
		}

		// Evaluate watermarks to run for
		ts, err := calculateExecutionTimes(self, a, watermark, previousWatermark)
		if err != nil {
			// TODO: Okay to return? More than usually unexpected error.
			return false, err
		}

		if len(ts) == 0 {
			// TODO: Debug log
			r.C.Logger.Info("Skipped alert check because watermark has not advanced by a full interval", zap.String("name", self.Meta.Name.Name), zap.Time("current_watermark", watermark), zap.Time("previous_watermark", previousWatermark), zap.String("interval", a.Spec.IntervalsIsoDuration))
		}

		for _, t := range ts {
			dirty, err = r.executeSingleAlert(ctx, self, a, t, adhocTrigger)
			if err != nil {
				executionErr = err
				break
			}
		}
	}

	// If executionErr is nil, we're done.
	if executionErr == nil {
		return false, nil
	}

	// If executionErr is a non-dirty cancellation, we're also done.
	if errors.Is(executionErr, context.Canceled) && !dirty {
		// TODO: Might CurrentExecution be non-nil here?
		return false, executionErr
	}

	// There was an execution error. Add it to the execution history.
	if a.State.CurrentExecution == nil {
		// CurrentExecution will only be nil if we never made it to the point of checking the alert query.
		a.State.CurrentExecution = &runtimev1.AlertExecution{
			Adhoc:         adhocTrigger,
			ExecutionTime: timestamppb.New(triggerTime), // TODO: What to put here? triggerTime? watermark? nil? (the most recently tried interval?)
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

// executeSingleAlert runs the alert query and maybe sends emails for a single execution time.
// It returns true if an error occurred after some or all emails were sent.
func (r *AlertReconciler) executeSingleAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, executionTime time.Time, adhocTrigger bool) (bool, error) {
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
	res, dirtyErr, alertErr := r.checkAlert(ctx, self, a, executionTime)

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

	// There was an error. By returning it, the caller will add it to the execution history.
	if errors.Is(alertErr, context.Canceled) {
		alertErr = fmt.Errorf("Alert check was interrupted after some emails were sent. The alert will not automatically retry.")
	} else {
		alertErr = fmt.Errorf("Alert check failed: %v", alertErr.Error())
	}
	return dirtyErr, alertErr
}

// checkAlert runs the alert query and maybe sends emails.
// It returns true if an error occurred after some or all emails were sent.
func (r *AlertReconciler) checkAlert(ctx context.Context, self *runtimev1.Resource, a *runtimev1.Alert, executionTime time.Time) (*runtimev1.AssertionResult, bool, error) {
	// Log
	// TODO: Turn into debug log
	r.C.Logger.Info("Checking alert", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))

	// Build query proto
	qpb, err := buildProtoQuery(a.Spec.QueryName, a.Spec.QueryArgsJson, executionTime)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse query: %w", err)
	}

	// Connect to admin service
	var openURL, editURL string
	var queryForAttrs map[string]any
	admin, release, err := r.C.Runtime.Admin(ctx, r.C.InstanceID)
	if err != nil && !errors.Is(err, runtime.ErrAdminNotConfigured) {
		return nil, false, fmt.Errorf("failed to get admin client: %w", err)
	}
	if err == nil {
		// Connected successfully
		defer release()

		// Get alert metadata
		meta, err := admin.GetAlertMetadata(ctx, a.Spec.QueryName, a.Spec.Annotations, a.Spec.GetQueryForUserId(), a.Spec.GetQueryForUserEmail())
		if err != nil {
			return nil, false, fmt.Errorf("failed to get alert metadata: %w", err)
		}
		queryForAttrs = meta.QueryForAttributes
		openURL = meta.OpenURL
		editURL = meta.EditURL
	}

	// Let explicit queryForAttributes take precedence
	if a.Spec.GetQueryForAttributes() != nil {
		queryForAttrs = a.Spec.GetQueryForAttributes().AsMap()
	}

	// Create and execute query
	q, err := buildRuntimeQuery(qpb, queryForAttrs)
	if err != nil {
		return nil, false, fmt.Errorf("failed to build query: %w", err)
	}
	err = r.C.Runtime.Query(ctx, r.C.InstanceID, q, alertQueryPriority)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute query: %w", err)
	}

	// Extract result row
	row, ok, err := extractQueryResultFirstRow(q)
	if err != nil {
		return nil, false, fmt.Errorf("failed to extract query result: %w", err)
	}
	if !ok {
		r.C.Logger.Info("Alert passed", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))
		return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_PASS}, false, nil
	}

	r.C.Logger.Info("Alert failed", zap.String("name", self.Meta.Name.Name), zap.Time("execution_time", executionTime))

	// Send emails
	for _, recipient := range a.Spec.EmailRecipients {
		err := r.C.Runtime.Email.SendAlert(&email.Alert{
			ToEmail:       recipient,
			ToName:        "",
			Title:         a.Spec.Title,
			ExecutionTime: executionTime,
			FailRow:       row,
			OpenLink:      openURL,
			EditLink:      editURL,
		})
		if err != nil {
			return nil, true, fmt.Errorf("failed to send email to %q: %w", recipient, err)
		}
	}

	// Return fail row
	failRow, err := structpb.NewStruct(row)
	if err != nil {
		return nil, true, fmt.Errorf("failed to convert fail row to proto: %w", err)
	}
	return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, FailRow: failRow}, false, nil
}

// computeInheritedWatermark computes the inherited watermark for the alert.
// It returns false if the watermark could not be computed.
func (r *AlertReconciler) computeInheritedWatermark(ctx context.Context, refs []*runtimev1.ResourceName) (time.Time, bool, error) {
	var t time.Time
	for _, ref := range refs {
		q := &queries.ResourceWatermark{
			ResourceKind: ref.Kind,
			ResourceName: ref.Name,
		}
		err := r.C.Runtime.Query(ctx, r.C.InstanceID, q, alertQueryPriority)
		if err != nil {
			return t, false, fmt.Errorf("failed to resolve watermark for %s/%s: %w", ref.Kind, ref.Name, err)
		}

		if q.Result != nil && (t.IsZero() || q.Result.Before(t)) {
			t = *q.Result
		}
	}

	return t, !t.IsZero(), nil
}

// calculateExecutionTimes calculates the execution times for an alert, taking into consideration the alert's intervals configuration and previous executions.
// If the alert is not configured to run on intervals, it will return a slice containing only the current watermark.
func calculateExecutionTimes(self *runtimev1.Resource, a *runtimev1.Alert, watermark, previousWatermark time.Time) ([]time.Time, error) {
	// If the alert is not configured to run on intervals, check it just for the current watermark.
	if a.Spec.IntervalsIsoDuration == "" {
		return []time.Time{watermark}, nil
	}

	// Note: The watermark and previousWatermark may be unaligned with the alert's interval duration.

	// Parse the interval duration
	// The YAML parser validates it as a StandardDuration, so this shouldn't fail.
	di, err := duration.ParseISO8601(a.Spec.IntervalsIsoDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval duration: %w", err)
	}
	d, ok := di.(duration.StandardDuration)
	if !ok {
		return nil, fmt.Errorf("interval duration %q is not a standard ISO 8601 duration", a.Spec.IntervalsIsoDuration)
	}

	// Extract time zone
	tz := time.UTC
	if a.Spec.RefreshSchedule != nil && a.Spec.RefreshSchedule.TimeZone != "" {
		tz, err = time.LoadLocation(a.Spec.RefreshSchedule.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("failed to load time zone %q: %w", a.Spec.RefreshSchedule.TimeZone, err)
		}
	}

	// Compute the last end time (rounded to the interval duration)
	// TODO: Find a way to incorporate first day of week and first month of year?
	end := watermark.In(tz)
	if a.Spec.IntervalsCheckUnclosed {
		// Ceil
		t := d.Truncate(end, 1, 1)
		if !t.Equal(end) {
			end = d.Add(t)
		}
	} else {
		// Floor
		end = d.Truncate(end, 1, 1)
	}

	// If there isn't a previous watermark, we'll just check the current watermark.
	if previousWatermark.IsZero() {
		return []time.Time{end}, nil
	}

	// Skip if end isn't past the previous watermark (unless we're supposed to check unclosed intervals)
	if !a.Spec.IntervalsCheckUnclosed && !end.After(previousWatermark) {
		return nil, nil
	}

	// Set a limit on the number of intervals to check
	limit := int(a.Spec.IntervalsLimit)
	if limit <= 0 {
		limit = alertDefaultIntervalsLimit
	}

	// Calculate the execution times
	ts := []time.Time{end}
	for i := 0; i < limit; i++ {
		t := ts[len(ts)-1]
		t = d.Sub(t)
		if !t.After(previousWatermark) {
			break
		}
		ts = append(ts, t)
	}

	// Reverse execution times so we run them in chronological order
	slices.Reverse(ts)

	return ts, nil
}

// buildRuntimeQuery builds a runtime query from a proto query and security attributes.
func buildRuntimeQuery(q *runtimev1.Query, attrs map[string]any) (runtime.Query, error) {
	one := int64(1)

	// NOTE: Pending refactors, this implementation is replicated from handlers in runtime/server.
	switch r := q.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		req := r.MetricsViewAggregationRequest

		tr := req.TimeRange
		if req.TimeStart != nil || req.TimeEnd != nil {
			tr = &runtimev1.TimeRange{
				Start: req.TimeStart,
				End:   req.TimeEnd,
			}
		}

		return &queries.MetricsViewAggregation{
			MetricsViewName:    req.MetricsView,
			Dimensions:         req.Dimensions,
			Measures:           req.Measures,
			Sort:               req.Sort,
			TimeRange:          tr,
			Where:              req.Where,
			Having:             req.Having,
			Filter:             req.Filter,
			Limit:              &one, // Alerts never inspect more than one row // TODO: Maybe put higher limit and return the minimum number of matching rows?
			Offset:             req.Offset,
			PivotOn:            req.PivotOn,
			SecurityAttributes: attrs,
		}, nil
	case *runtimev1.Query_MetricsViewComparisonRequest:
		req := r.MetricsViewComparisonRequest
		return &queries.MetricsViewComparison{
			MetricsViewName:     req.MetricsViewName,
			DimensionName:       req.Dimension.Name,
			Measures:            req.Measures,
			ComparisonMeasures:  req.ComparisonMeasures,
			TimeRange:           req.TimeRange,
			ComparisonTimeRange: req.ComparisonTimeRange,
			Limit:               req.Limit,
			Offset:              req.Offset,
			Sort:                req.Sort,
			Where:               req.Where,
			Having:              req.Having,
			Filter:              req.Filter,
			Exact:               req.Exact,
			SecurityAttributes:  attrs,
		}, nil
	default:
		return nil, fmt.Errorf("query type %T not supported for alerts", r)
	}
}

// extractQueryResultFirstRow extracts the first row from a query result.
// TODO: This should function more like an export, i.e. use dimension/measure labels instead of names.
func extractQueryResultFirstRow(q runtime.Query) (map[string]any, bool, error) {
	switch q := q.(type) {
	case *queries.MetricsViewAggregation:
		if q.Result != nil && len(q.Result.Data) > 0 {
			row := q.Result.Data[0]
			return row.AsMap(), true, nil
		}
		return nil, false, nil
	case *queries.MetricsViewComparison:
		if q.Result != nil && len(q.Result.Rows) > 0 {
			row := q.Result.Rows[0]
			res := make(map[string]any)
			res[q.DimensionName] = row.DimensionValue
			for _, v := range row.MeasureValues {
				res[v.MeasureName] = v.BaseValue.AsInterface()
				if v.ComparisonValue != nil {
					res[v.MeasureName+" (prev)"] = v.ComparisonValue.AsInterface()
				}
				if v.DeltaAbs != nil {
					res[v.MeasureName+" (Δ)"] = v.DeltaAbs.AsInterface()
				}
				if v.DeltaRel != nil {
					res[v.MeasureName+" (Δ%)"] = v.DeltaRel.AsInterface()
				}
			}
			return res, true, nil
		}
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("query type %T not supported for alerts", q)
	}
}
