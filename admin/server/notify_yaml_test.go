package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Runtimes that predate the webhook notifier parse alerts and reports with KnownFields(true),
// so the generated YAML must only carry notify.webhook when webhook URLs are set.
func TestNotifyWebhookOmittedWhenEmpty(t *testing.T) {
	s := &Server{}
	alertOpts := func(urls []string) *adminv1.AlertOptions {
		return &adminv1.AlertOptions{
			DisplayName:     "Revenue alert",
			MetricsViewName: "mv",
			EmailRecipients: []string{"user@example.com"},
			WebhookUrls:     urls,
		}
	}
	reportOpts := func(urls []string) *adminv1.ReportOptions {
		return &adminv1.ReportOptions{
			DisplayName:     "Weekly report",
			RefreshCron:     "0 9 * * 1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   `{"metrics_view":"mv"}`,
			ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			EmailRecipients: []string{"user@example.com"},
			WebhookUrls:     urls,
		}
	}

	generators := map[string]func(urls []string) ([]byte, error){
		"managed alert":    func(urls []string) ([]byte, error) { return s.yamlForManagedAlert(alertOpts(urls), "owner") },
		"committed alert":  func(urls []string) ([]byte, error) { return s.yamlForCommittedAlert(alertOpts(urls)) },
		"managed report":   func(urls []string) ([]byte, error) { return s.yamlForManagedReport(reportOpts(urls), "owner") },
		"committed report": func(urls []string) ([]byte, error) { return s.yamlForCommittedReport(reportOpts(urls)) },
	}
	for name, gen := range generators {
		t.Run(name, func(t *testing.T) {
			for _, urls := range [][]string{nil, {}} {
				data, err := gen(urls)
				require.NoError(t, err)
				require.NotContains(t, notifyKeys(t, data), "webhook")
			}

			data, err := gen([]string{"https://example.com/hook"})
			require.NoError(t, err)
			require.Contains(t, notifyKeys(t, data), "webhook")
		})
	}
}

func notifyKeys(t *testing.T, data []byte) []string {
	var doc struct {
		Notify map[string]any `yaml:"notify"`
	}
	require.NoError(t, yaml.Unmarshal(data, &doc))
	keys := make([]string, 0, len(doc.Notify))
	for k := range doc.Notify {
		keys = append(keys, k)
	}
	return keys
}
