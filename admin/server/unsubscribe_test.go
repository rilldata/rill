package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Unsubscribing the last email recipient must not delete an alert that still notifies a webhook.
func TestUnsubscribeKeepsAlertWithWebhook(t *testing.T) {
	managedAlert := func(webhookURLs []string) *alertYAML {
		data, err := (&Server{}).yamlForManagedAlert(&adminv1.AlertOptions{
			DisplayName:     "Revenue alert",
			MetricsViewName: "mv",
			EmailRecipients: []string{"user@example.com"},
			WebhookUrls:     webhookURLs,
		}, "owner")
		require.NoError(t, err)
		var alert alertYAML
		require.NoError(t, yaml.Unmarshal(data, &alert))
		return &alert
	}

	alert := managedAlert([]string{"https://example.com/hook"})
	found, remaining := alert.unsubscribe("USER@example.com", "")
	require.True(t, found)
	require.True(t, remaining, "the alert still notifies https://example.com/hook")
	require.Empty(t, alert.Notify.Email.Recipients)

	found, remaining = managedAlert(nil).unsubscribe("user@example.com", "")
	require.True(t, found)
	require.False(t, remaining)

	found, _ = managedAlert(nil).unsubscribe("other@example.com", "")
	require.False(t, found)
}

// Unsubscribing the last email recipient must not delete a report that still notifies a webhook.
func TestUnsubscribeKeepsReportWithWebhook(t *testing.T) {
	managedReport := func(webhookURLs []string) *reportYAML {
		data, err := (&Server{}).yamlForManagedReport(&adminv1.ReportOptions{
			DisplayName:     "Weekly report",
			RefreshCron:     "0 9 * * 1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   `{"metrics_view":"mv"}`,
			ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			EmailRecipients: []string{"user@example.com"},
			WebhookUrls:     webhookURLs,
		}, "owner")
		require.NoError(t, err)
		var report reportYAML
		require.NoError(t, yaml.Unmarshal(data, &report))
		return &report
	}

	report := managedReport([]string{"https://example.com/hook"})
	found, remaining := report.unsubscribe("USER@example.com", "")
	require.True(t, found)
	require.True(t, remaining, "the report still notifies https://example.com/hook")
	require.Empty(t, report.Notify.Email.Recipients)

	found, remaining = managedReport(nil).unsubscribe("user@example.com", "")
	require.True(t, found)
	require.False(t, remaining)

	found, _ = managedReport(nil).unsubscribe("other@example.com", "")
	require.False(t, found)
}
