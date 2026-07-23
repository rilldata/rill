package reconcilers

import (
	"context"
	"errors"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestReportEmailDeliveryTracksDirtyStateAcrossNotifiers(t *testing.T) {
	// Automatic retry is safe only until the first successful external send.
	// Separate email notifiers model the same state transition as email then Slack.
	tests := []struct {
		name      string
		failAt    int
		wantDirty bool
		wantCalls []string
	}{
		{name: "first delivery fails cleanly", failAt: 1, wantDirty: false, wantCalls: []string{"first@example.com"}},
		{name: "second delivery fails after side effect", failAt: 2, wantDirty: true, wantCalls: []string{"first@example.com", "second@example.com"}},
		{name: "all deliveries succeed", wantDirty: true, wantCalls: []string{"first@example.com", "second@example.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &failNthReportSender{failAt: tt.failAt, err: context.Canceled}
			reconciler, self, report := newReportDeliveryFixture(t, sender, "first@example.com", "second@example.com")
			content := map[string]*notificationData{
				"first@example.com":  {openLink: "https://example.com/first"},
				"second@example.com": {openLink: "https://example.com/second"},
			}

			dirty, warnings, err := reconciler.sendReportNotifications(t.Context(), self, report, time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC), content)
			require.Equal(t, tt.wantDirty, dirty)
			require.Empty(t, warnings)
			require.Equal(t, tt.wantCalls, sender.recipients)
			if tt.failAt == 0 {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, context.Canceled)
			}
		})
	}
}

func TestReportEmailDeliverySurfacesMissingRecipientContent(t *testing.T) {
	// Recipient-mode metadata can omit a user. The run must record that omission
	// instead of silently appearing fully successful while sending nothing.
	sender := &failNthReportSender{}
	reconciler, self, report := newReportDeliveryFixture(t, sender, "missing@example.com")

	dirty, warnings, err := reconciler.sendReportNotifications(t.Context(), self, report, time.Now(), map[string]*notificationData{})
	require.NoError(t, err)
	require.False(t, dirty)
	require.Empty(t, sender.recipients)
	require.Equal(t, []string{`Skipped recipient "missing@example.com" because notification content was unavailable`}, warnings)
}

func TestClassifyReportErrorRetryInvariant(t *testing.T) {
	// Cancellation and deadline errors have identical retry semantics; ordinary
	// failures are recorded but never treated as controller interruptions.
	tests := []struct {
		name        string
		dirty       bool
		err         error
		wantRetry   bool
		wantMessage string
	}{
		{name: "clean cancellation retries", err: context.Canceled, wantRetry: true, wantMessage: "It will automatically retry"},
		{name: "dirty cancellation does not retry", dirty: true, err: context.Canceled, wantRetry: false, wantMessage: "will not automatically retry"},
		{name: "clean deadline retries", err: context.DeadlineExceeded, wantRetry: true, wantMessage: "It will automatically retry"},
		{name: "dirty deadline does not retry", dirty: true, err: context.DeadlineExceeded, wantRetry: false, wantMessage: "will not automatically retry"},
		{name: "ordinary failure does not retry", err: errors.New("export failed"), wantRetry: false, wantMessage: "Report run failed: export failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retry, message := classifyReportError(tt.dirty, tt.err)
			require.Equal(t, tt.wantRetry, retry)
			require.Contains(t, message, tt.wantMessage)
		})
	}
}

type failNthReportSender struct {
	failAt     int
	err        error
	recipients []string
}

func (s *failNthReportSender) Send(toEmail, _ string, _, _ string) error {
	s.recipients = append(s.recipients, toEmail)
	if s.failAt > 0 && len(s.recipients) == s.failAt {
		return s.err
	}
	return nil
}

func newReportDeliveryFixture(t *testing.T, sender email.Sender, recipients ...string) (*ReportReconciler, *runtimev1.Resource, *runtimev1.Report) {
	t.Helper()
	controller := &runtime.Controller{
		Runtime: &runtime.Runtime{Email: email.New(sender)},
		Logger:  zap.NewNop(),
	}
	reconciler := &ReportReconciler{C: controller}
	self := &runtimev1.Resource{Meta: &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "delivery-test"}}}
	report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{DisplayName: "Delivery test"}}
	for _, recipient := range recipients {
		props, err := structpb.NewStruct(map[string]any{"recipients": []any{recipient}})
		require.NoError(t, err)
		report.Spec.Notifiers = append(report.Spec.Notifiers, &runtimev1.Notifier{
			Connector:  "email",
			Properties: props,
		})
	}
	return reconciler, self, report
}
