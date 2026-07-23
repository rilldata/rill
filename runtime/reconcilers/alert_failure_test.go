package reconcilers

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCalculateAlertExecutionTimesHonorsConfiguredLimit(t *testing.T) {
	// The interval limit counts actual checks. An off-by-one here can fan out one
	// extra query and notification every time a delayed alert catches up.
	alert := &runtimev1.Alert{Spec: &runtimev1.AlertSpec{
		IntervalsIsoDuration: "P1D",
		IntervalsLimit:       2,
	}}
	watermark := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	previous := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	got, err := calculateAlertExecutionTimes(alert, watermark, previous)
	require.NoError(t, err)
	require.Equal(t, []time.Time{
		time.Date(2026, time.January, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
	}, got)
}

func TestCalculateAlertExecutionTimesFailureBoundaries(t *testing.T) {
	watermark := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)

	// Reconciler inputs can bypass parser validation during recovery or upgrades,
	// so malformed durations and time zones must return errors rather than panic.
	tests := []struct {
		name     string
		interval string
		timezone string
		want     string
	}{
		{name: "malformed duration", interval: "not-a-duration", want: "failed to parse interval duration"},
		{name: "nonstandard duration", interval: "inf", want: "is not a standard ISO 8601 duration"},
		{name: "invalid timezone", interval: "P1D", timezone: "Mars/Olympus", want: "failed to load time zone"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each malformed persisted scheduling value must produce a stable,
			// actionable error instead of panicking the reconciler.
			alert := &runtimev1.Alert{Spec: &runtimev1.AlertSpec{
				IntervalsIsoDuration: tt.interval,
				RefreshSchedule:      &runtimev1.Schedule{TimeZone: tt.timezone},
			}}
			_, err := calculateAlertExecutionTimes(alert, watermark, time.Time{})
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestLatestAlertWarningsUsesNewestValidHistoryEntry(t *testing.T) {
	// Histories are newest-first. Reading the tail exposes stale warnings, while
	// a partially persisted nil result must not crash reconciliation.
	alert := &runtimev1.Alert{State: &runtimev1.AlertState{ExecutionHistory: []*runtimev1.AlertExecution{
		{Result: &runtimev1.AssertionResult{Warnings: []string{"newest warning"}}},
		{Result: &runtimev1.AssertionResult{Warnings: []string{"oldest warning"}}},
	}}}
	require.Equal(t, []string{"newest warning"}, latestAlertWarnings(alert))
	require.Nil(t, latestAlertWarnings(&runtimev1.Alert{State: &runtimev1.AlertState{ExecutionHistory: []*runtimev1.AlertExecution{{}}}}))
}

func TestPrepareAlertNotificationRenotifyBoundaries(t *testing.T) {
	// Renotify uses the last actual send time, and the configured delay is
	// inclusive: 59 seconds is suppressed while exactly 60 seconds is sent.
	lastSent := time.Date(2026, time.July, 23, 10, 0, 0, 0, time.UTC)
	alert := &runtimev1.Alert{
		Spec: &runtimev1.AlertSpec{
			DisplayName:          "Orders failed",
			NotifyOnFail:         true,
			Renotify:             true,
			RenotifyAfterSeconds: 60,
		},
		State: &runtimev1.AlertState{ExecutionHistory: []*runtimev1.AlertExecution{{
			Result:            alertFailureResult(t),
			ExecutionTime:     timestamppb.New(lastSent),
			SentNotifications: true,
		}}},
	}

	suppressed := &runtimev1.AlertExecution{
		Result:        alertFailureResult(t),
		ExecutionTime: timestamppb.New(lastSent.Add(59 * time.Second)),
	}
	msg, err := prepareAlertNotification(alert, suppressed)
	require.NoError(t, err)
	require.Nil(t, msg)
	require.Equal(t, lastSent, suppressed.SuppressedSince.AsTime())

	atBoundary := &runtimev1.AlertExecution{
		Result:        alertFailureResult(t),
		ExecutionTime: timestamppb.New(lastSent.Add(60 * time.Second)),
	}
	msg, err = prepareAlertNotification(alert, atBoundary)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, msg.Status)
}

func TestPrepareAlertNotificationRenotifyModes(t *testing.T) {
	// Zero seconds has different meaning when renotify is enabled, so exercise
	// both modes without relying on elapsed-time rounding.
	lastSent := time.Date(2026, time.July, 23, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		renotify   bool
		after      uint32
		wantNotify bool
	}{
		{name: "disabled suppresses unchanged status", wantNotify: false},
		{name: "zero delay resends unchanged status", renotify: true, wantNotify: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Renotify policy must distinguish a disabled policy from an enabled
			// zero-second policy even though both store a zero duration.
			alert := alertForNotification(&runtimev1.AlertSpec{
				NotifyOnFail:         true,
				Renotify:             tt.renotify,
				RenotifyAfterSeconds: tt.after,
			}, []*runtimev1.AlertExecution{{
				Result:            alertFailureResult(t),
				ExecutionTime:     timestamppb.New(lastSent),
				SentNotifications: true,
			}})
			msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, lastSent.Add(time.Second)))
			require.NoError(t, err)
			require.Equal(t, tt.wantNotify, msg != nil)
		})
	}
}

func TestSendAlertNotificationsStopsBeforeCanceledDelivery(t *testing.T) {
	// A pre-existing context interruption must be detected before the first
	// transport call for both cancellation and deadline expiry.
	tests := []struct {
		name string
		ctx  func() context.Context
		err  error
	}{
		{
			name: "cancellation",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			err: context.Canceled,
		},
		{
			name: "deadline",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				cancel()
				return ctx
			},
			err: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// An interruption already visible before the first side effect must
			// leave the delivery clean and must not call the email transport.
			sender := &alertCallbackSender{}
			reconciler, alert, msg := newAlertDeliverySeam(t, sender, "first@example.com")
			sent, err := reconciler.sendAlertNotifications(tt.ctx(), alert, nil, msg, msg.ExecutionTime)
			require.ErrorIs(t, err, tt.err)
			require.False(t, sent)
			require.Empty(t, sender.recipients)
		})
	}
}

func TestSendAlertNotificationsMarksInterruptionAfterDeliveryDirty(t *testing.T) {
	// Once one recipient succeeds, later interruption must be reported as dirty
	// so callers can avoid replaying that successful side effect.
	t.Run("cancellation", func(t *testing.T) {
		// Cancellation observed after the first successful recipient is dirty;
		// the second recipient must not be attempted by this invocation.
		ctx, cancel := context.WithCancel(context.Background())
		sender := &alertCallbackSender{afterSend: func(call int) {
			if call == 1 {
				cancel()
			}
		}}
		reconciler, alert, msg := newAlertDeliverySeam(t, sender, "first@example.com", "second@example.com")
		sent, err := reconciler.sendAlertNotifications(ctx, alert, nil, msg, msg.ExecutionTime)
		require.ErrorIs(t, err, context.Canceled)
		require.True(t, sent)
		require.Equal(t, []string{"first@example.com"}, sender.recipients)
	})

	t.Run("deadline", func(t *testing.T) {
		// A deadline that expires immediately after the first transport success
		// has the same dirty semantics as cancellation.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		sender := &alertCallbackSender{afterSend: func(call int) {
			if call == 1 {
				<-ctx.Done()
			}
		}}
		reconciler, alert, msg := newAlertDeliverySeam(t, sender, "first@example.com", "second@example.com")
		sent, err := reconciler.sendAlertNotifications(ctx, alert, nil, msg, msg.ExecutionTime)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.True(t, sent)
		require.Equal(t, []string{"first@example.com"}, sender.recipients)
	})
}

func TestPrepareAlertNotificationStatusTransitions(t *testing.T) {
	// Notification policy distinguishes an initial pass from a real recovery,
	// while fail and error transitions retain their semantic payloads.
	executionTime := time.Date(2026, time.July, 23, 11, 0, 0, 0, time.UTC)
	t.Run("initial pass is silent", func(t *testing.T) {
		// A pass without a preceding non-pass status is not a recovery event.
		alert := alertForNotification(&runtimev1.AlertSpec{NotifyOnRecover: true}, nil)
		msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_PASS, executionTime))
		require.NoError(t, err)
		require.Nil(t, msg)
	})

	t.Run("recovery follows a failure", func(t *testing.T) {
		// A pass immediately after failure carries the recovery marker.
		previous := alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, executionTime.Add(-time.Minute))
		alert := alertForNotification(&runtimev1.AlertSpec{DisplayName: "Orders", NotifyOnRecover: true}, []*runtimev1.AlertExecution{previous})
		msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_PASS, executionTime))
		require.NoError(t, err)
		require.True(t, msg.IsRecover)
		require.Equal(t, "Orders", msg.DisplayName)
	})

	t.Run("cleanly failed recovery delivery retries", func(t *testing.T) {
		// A recovery send that reached no recipient remains recoverable, while
		// retaining PASS as the actual assertion status in persisted history.
		failedRecovery := alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_PASS, executionTime.Add(-time.Minute))
		failedRecovery.Result.ErrorMessage = "failed to send recovery email"
		previousFailure := alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, executionTime.Add(-2*time.Minute))
		alert := alertForNotification(&runtimev1.AlertSpec{NotifyOnRecover: true}, []*runtimev1.AlertExecution{failedRecovery, previousFailure})
		msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_PASS, executionTime))
		require.NoError(t, err)
		require.True(t, msg.IsRecover)
	})

	t.Run("error carries execution failure", func(t *testing.T) {
		// Evaluation errors retain their diagnostic text in the outgoing payload.
		alert := alertForNotification(&runtimev1.AlertSpec{NotifyOnError: true}, nil)
		current := alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR, executionTime)
		current.Result.ErrorMessage = "query unavailable"
		msg, err := prepareAlertNotification(alert, current)
		require.NoError(t, err)
		require.Equal(t, "query unavailable", msg.ExecutionError)
	})

	t.Run("disabled failure notification is silent", func(t *testing.T) {
		// A failing assertion must remain silent when fail notifications are off.
		alert := alertForNotification(&runtimev1.AlertSpec{NotifyOnFail: false}, nil)
		msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, executionTime))
		require.NoError(t, err)
		require.Nil(t, msg)
	})
}

func TestPrepareAlertNotificationMalformedPersistedState(t *testing.T) {
	// Reconciliation can encounter partially persisted state after a crash, so
	// missing results and fail rows must return errors instead of panicking.
	alert := alertForNotification(&runtimev1.AlertSpec{NotifyOnFail: true}, nil)
	_, err := prepareAlertNotification(alert, &runtimev1.AlertExecution{})
	require.ErrorContains(t, err, "result is missing")

	current := alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, time.Now())
	current.Result.FailRow = nil
	_, err = prepareAlertNotification(alert, current)
	require.ErrorContains(t, err, "fail row")

	current = alertExecution(runtimev1.AssertionStatus(99), time.Now())
	_, err = prepareAlertNotification(alert, current)
	require.ErrorContains(t, err, "unexpected assertion result status")

	// A malformed history entry is treated as a status boundary rather than
	// dereferenced, allowing the current valid failure to notify safely.
	alert.State.ExecutionHistory = []*runtimev1.AlertExecution{nil}
	msg, err := prepareAlertNotification(alert, alertExecution(runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, time.Now()))
	require.NoError(t, err)
	require.NotNil(t, msg)
}

func TestAddExecutionTimePreservesAndValidatesURL(t *testing.T) {
	// Recipient links retain existing query parameters and fragments, and a
	// malformed encoded query is rejected before any notification is sent.
	executionTime := time.Date(2026, time.July, 23, 12, 30, 0, 0, time.FixedZone("offset", 5*60*60+30*60))
	got, err := addExecutionTime("https://example.com/explore?foo=bar#chart", executionTime)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/explore?execution_time=2026-07-23T07%3A00%3A00Z&foo=bar#chart", got)

	_, err = addExecutionTime("https://example.com/explore?bad=%zz", executionTime)
	require.Error(t, err)
}

func alertFailureResult(t *testing.T) *runtimev1.AssertionResult {
	t.Helper()
	row, err := structpb.NewStruct(map[string]any{"orders": 1})
	require.NoError(t, err)
	return &runtimev1.AssertionResult{Status: runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, FailRow: row}
}

func alertExecution(status runtimev1.AssertionStatus, executionTime time.Time) *runtimev1.AlertExecution {
	result := &runtimev1.AssertionResult{Status: status}
	if status == runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL {
		result.FailRow, _ = structpb.NewStruct(map[string]any{"orders": 1})
	}
	return &runtimev1.AlertExecution{Result: result, ExecutionTime: timestamppb.New(executionTime)}
}

func alertForNotification(spec *runtimev1.AlertSpec, history []*runtimev1.AlertExecution) *runtimev1.Alert {
	return &runtimev1.Alert{Spec: spec, State: &runtimev1.AlertState{ExecutionHistory: history}}
}

type alertCallbackSender struct {
	recipients []string
	afterSend  func(call int)
}

func (s *alertCallbackSender) Send(toEmail, _, _, _ string) error {
	s.recipients = append(s.recipients, toEmail)
	if s.afterSend != nil {
		s.afterSend(len(s.recipients))
	}
	return nil
}

func newAlertDeliverySeam(t *testing.T, sender email.Sender, recipients ...string) (*AlertReconciler, *runtimev1.Alert, *drivers.AlertStatus) {
	t.Helper()
	values := make([]any, len(recipients))
	for i, recipient := range recipients {
		values[i] = recipient
	}
	properties, err := structpb.NewStruct(map[string]any{"recipients": values})
	require.NoError(t, err)
	reconciler := &AlertReconciler{C: &runtime.Controller{Runtime: &runtime.Runtime{Email: email.New(sender)}}}
	alert := &runtimev1.Alert{Spec: &runtimev1.AlertSpec{Notifiers: []*runtimev1.Notifier{{Connector: "email", Properties: properties}}}}
	msg := &drivers.AlertStatus{
		DisplayName:   "Delivery seam",
		ExecutionTime: time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
		FailRow:       map[string]any{"orders": 1},
	}
	return reconciler, alert, msg
}
