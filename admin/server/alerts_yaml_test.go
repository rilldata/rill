package server

import (
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
		}, "owner", "", "", nil)
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
		}, "owner", "", "", nil)
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

func TestYAMLForManagedAlertEmbed(t *testing.T) {
	s := &Server{}
	newOpts := func() *adminv1.AlertOptions {
		return &adminv1.AlertOptions{
			DisplayName:      "Embed alert",
			IntervalDuration: "P1D",
			QueryName:        "MetricsViewAggregation",
			QueryArgsJson:    `{"metrics_view":"mv1"}`,
			MetricsViewName:  "mv1",
			EmailRecipients:  []string{"Embed@Example.com"},
		}
	}

	// Alerts created from an embed don't run for a Rill user; the admin service resolves the embed user's identity when they run
	opts := newOpts()
	opts.EmailRecipients = []string{"embed@example.com", "other@example.com"}
	data, err := s.yamlForManagedAlert(opts, "", "ext_abc", "", nil)
	require.NoError(t, err)
	var res alertYAML
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, opts.EmailRecipients, res.Notify.Email.Recipients)
	require.Nil(t, res.For)
	require.Empty(t, res.Query.For.UserID)
	require.Empty(t, res.Annotations.AdminOwnerUserID)
	require.Equal(t, "ext_abc", res.Annotations.AdminEmbedOwnerID)
	require.Empty(t, res.Annotations.AdminEmbedOwnerEmail)
	require.Empty(t, res.Annotations.AdminEmbedResources)
	require.True(t, res.Annotations.AdminManaged)

	// The creator's email is saved, and isn't added again if it's already a recipient (compared case-insensitively)
	data, err = s.yamlForManagedAlert(newOpts(), "", "ext_abc", "embed@example.com", nil)
	require.NoError(t, err)
	res = alertYAML{}
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, []string{"Embed@Example.com"}, res.Notify.Email.Recipients)
	require.Equal(t, "embed@example.com", res.Annotations.AdminEmbedOwnerEmail)

	// The creator's email is added to the recipients if missing
	opts = newOpts()
	opts.EmailRecipients = []string{"other@example.com"}
	data, err = s.yamlForManagedAlert(opts, "", "ext_abc", "embed@example.com", nil)
	require.NoError(t, err)
	res = alertYAML{}
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, []string{"other@example.com", "embed@example.com"}, res.Notify.Email.Recipients)
	require.Equal(t, []string{"other@example.com"}, opts.EmailRecipients)

	// The resources of the embed are saved
	resources := []database.ResourceName{{Type: runtime.ResourceKindExplore, Name: "e1"}}
	data, err = s.yamlForManagedAlert(newOpts(), "", "ext_abc", "", resources)
	require.NoError(t, err)
	res = alertYAML{}
	require.NoError(t, yaml.Unmarshal(data, &res))
	parsed, err := parseEmbedResources(res.Annotations.AdminEmbedResources)
	require.NoError(t, err)
	require.Equal(t, resources, parsed)

	// The annotations round trip through the runtime's annotations
	annotations := parseAlertAnnotations(map[string]string{
		"admin_embed_owner_id":    "ext_abc",
		"admin_embed_owner_email": "embed@example.com",
		"admin_embed_resources":   res.Annotations.AdminEmbedResources,
	})
	require.Equal(t, "ext_abc", annotations.AdminEmbedOwnerID)
	require.Equal(t, "embed@example.com", annotations.AdminEmbedOwnerEmail)
	require.Equal(t, res.Annotations.AdminEmbedResources, annotations.AdminEmbedResources)

	// Slack is not supported
	for _, set := range []func(o *adminv1.AlertOptions){
		func(o *adminv1.AlertOptions) { o.SlackChannels = []string{"#general"} },
		func(o *adminv1.AlertOptions) { o.SlackUsers = []string{"embed@example.com"} },
		func(o *adminv1.AlertOptions) { o.SlackWebhooks = []string{"https://hooks.slack.com/x"} },
	} {
		opts = newOpts()
		set(opts)
		_, err = s.yamlForManagedAlert(opts, "", "ext_abc", "", nil)
		require.ErrorContains(t, err, "can't be sent to Slack")
	}

	// Data resolvers are not supported, since they could access data the embed doesn't expose
	opts = newOpts()
	opts.QueryName = ""
	opts.QueryArgsJson = ""
	opts.Resolver = "sql"
	opts.ResolverProperties, _ = structpb.NewStruct(map[string]any{"sql": "SELECT * FROM secrets"})
	_, err = s.yamlForManagedAlert(opts, "", "ext_abc", "", nil)
	require.ErrorContains(t, err, "must use a query")

	// The restrictions don't apply to alerts created by Rill users, which run for their owner
	opts = newOpts()
	opts.SlackChannels = []string{"#general"}
	data, err = s.yamlForManagedAlert(opts, "owner", "", "", nil)
	require.NoError(t, err)
	res = alertYAML{}
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, "owner", res.For.UserID)
	require.Equal(t, "owner", res.Query.For.UserID)
	require.Equal(t, "owner", res.Annotations.AdminOwnerUserID)
	require.Empty(t, res.Annotations.AdminEmbedOwnerID)
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
			return s.yamlForManagedAlert(opts, "owner", "", "", nil)
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
			return s.yamlForManagedAlert(opts, "owner", "", "", nil)
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
