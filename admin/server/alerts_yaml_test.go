package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

func TestYAMLForManagedAlert(t *testing.T) {
	s := &Server{}

	t.Run("resource status alert", func(t *testing.T) {
		props, err := structpb.NewStruct(map[string]any{"where_error": true, "where_resource": []any{"model/orders"}})
		require.NoError(t, err)
		data, err := s.yamlForManagedAlert(&adminv1.AlertOptions{
			DisplayName:        "Resource errors",
			RefreshCron:        "*/10 * * * *",
			Resolver:           "resource_status",
			ResolverProperties: props,
			EmailRecipients:    []string{"admin@example.com"},
			WebOpenPath:        "/-/status/resources?status=error",
		}, "owner")
		require.NoError(t, err)

		var res map[string]any
		require.NoError(t, yaml.Unmarshal(data, &res))
		require.Equal(t, map[string]any{"resource_status": map[string]any{"where_error": true, "where_resource": []any{"model/orders"}}}, res["data"])
		require.Equal(t, true, res["on_recover"])
		require.NotContains(t, res, "refs")
		require.NotContains(t, res, "query")
		require.NotContains(t, res, "for")
		require.NotContains(t, res, "watermark")

		annotations := res["annotations"].(map[string]any)
		require.Equal(t, true, annotations["admin_system"])
		require.Equal(t, true, annotations["admin_managed"])
		require.NotContains(t, annotations, "admin_owner_user_id")
	})

	t.Run("metrics view alert", func(t *testing.T) {
		data, err := s.yamlForManagedAlert(&adminv1.AlertOptions{
			DisplayName:     "Revenue drop",
			QueryName:       "MetricsViewAggregation", // nolint:staticcheck // backwards compatibility
			QueryArgsJson:   `{"metrics_view":"mv1"}`, // nolint:staticcheck // backwards compatibility
			MetricsViewName: "mv1",
			EmailRecipients: []string{"user@example.com"},
		}, "owner")
		require.NoError(t, err)

		var res map[string]any
		require.NoError(t, yaml.Unmarshal(data, &res))
		require.Equal(t, []any{"MetricsView/mv1"}, res["refs"])
		require.Equal(t, "inherit", res["watermark"])
		require.Equal(t, map[string]any{"user_id": "owner"}, res["for"])
		require.Equal(t, map[string]any{"name": "MetricsViewAggregation", "args_json": `{"metrics_view":"mv1"}`, "for": map[string]any{"user_id": "owner"}}, res["query"])
		require.NotContains(t, res, "on_recover")

		annotations := res["annotations"].(map[string]any)
		require.Equal(t, "owner", annotations["admin_owner_user_id"])
		require.NotContains(t, annotations, "admin_system")
	})
}

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
