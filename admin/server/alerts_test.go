package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLForAlertExportAttach(t *testing.T) {
	s := &Server{}
	opts := &adminv1.AlertOptions{
		DisplayName:         "Alert",
		IntervalDuration:    "P1D",
		QueryName:           "MetricsViewAggregation",
		QueryArgsJson:       `{"metrics_view":"mv1"}`,
		MetricsViewName:     "mv1",
		ExportAttach:        true,
		ExportFormat:        runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
		ExportIncludeHeader: true,
		ExportLimit:         100,
		EmailRecipients:     []string{"user_1@example.com"},
	}

	for _, committed := range []bool{false, true} {
		yamlForAlert := func(opts *adminv1.AlertOptions) ([]byte, error) {
			if committed {
				return s.yamlForCommittedAlert(opts)
			}
			return s.yamlForManagedAlert(opts, "owner")
		}

		// The export options are set when attaching
		data, err := yamlForAlert(opts)
		require.NoError(t, err)
		var res alertYAML
		require.NoError(t, yaml.Unmarshal(data, &res))
		require.True(t, res.Export.Attach)
		require.Equal(t, "xlsx", res.Export.Format)
		require.True(t, res.Export.IncludeHeader)
		require.Equal(t, uint64(100), res.Export.Limit)

		// The export options are omitted when not attaching, so runtimes that predate them can still parse the YAML
		opts.ExportAttach = false
		data, err = yamlForAlert(opts)
		require.NoError(t, err)
		require.NotContains(t, string(data), "export")
		opts.ExportAttach = true

		// Attachments require a data export format
		opts.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_PDF
		_, err = yamlForAlert(opts)
		require.ErrorContains(t, err, "invalid export format")
		opts.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_XLSX

		// Attachments are not supported for resolver-based alerts
		_, err = yamlForAlert(&adminv1.AlertOptions{
			DisplayName:     "Alert",
			MetricsViewName: "mv1",
			Resolver:        "metrics_sql",
			ExportAttach:    true,
			ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		})
		require.ErrorContains(t, err, "only supported for query-based alerts")
	}
}

func TestYAMLForAlertNotifyRowsLimit(t *testing.T) {
	s := &Server{}
	opts := &adminv1.AlertOptions{
		DisplayName:      "Alert",
		IntervalDuration: "P1D",
		QueryName:        "MetricsViewAggregation",
		QueryArgsJson:    `{"metrics_view":"mv1"}`,
		MetricsViewName:  "mv1",
		NotifyRowsLimit:  10,
		EmailRecipients:  []string{"user_1@example.com"},
	}

	for _, committed := range []bool{false, true} {
		yamlForAlert := func(opts *adminv1.AlertOptions) ([]byte, error) {
			if committed {
				return s.yamlForCommittedAlert(opts)
			}
			return s.yamlForManagedAlert(opts, "owner")
		}

		// The limit is set
		data, err := yamlForAlert(opts)
		require.NoError(t, err)
		var res alertYAML
		require.NoError(t, yaml.Unmarshal(data, &res))
		require.Equal(t, uint32(10), res.Notify.RowsLimit)

		// The limit is omitted when unset, so runtimes that predate it can still parse the YAML
		opts.NotifyRowsLimit = 0
		data, err = yamlForAlert(opts)
		require.NoError(t, err)
		require.NotContains(t, string(data), "rows_limit")
		opts.NotifyRowsLimit = 10
	}
}
