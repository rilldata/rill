package server

import (
	"testing"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// TestYAMLForManagedReportOwner tests that managed reports only run "for" an explicit identity when created by an embed user.
func TestYAMLForManagedReportOwner(t *testing.T) {
	tt := []struct {
		name        string
		ownerUserID string
		ownerEmail  string
		webOpenMode WebOpenMode
		wantErr     bool
	}{
		{name: "user creator mode", ownerUserID: "u1", webOpenMode: WebOpenModeCreator},
		{name: "user recipient mode", ownerUserID: "u1", webOpenMode: WebOpenModeRecipient},
		{name: "user none mode", ownerUserID: "u1", webOpenMode: WebOpenModeNone},
		{name: "user default mode", ownerUserID: "u1"},
		{name: "embed creator mode", ownerEmail: "embed@example.com", webOpenMode: WebOpenModeCreator},
		{name: "embed recipient mode", ownerEmail: "embed@example.com", webOpenMode: WebOpenModeRecipient, wantErr: true},
		{name: "embed none mode", ownerEmail: "embed@example.com", webOpenMode: WebOpenModeNone, wantErr: true},
		{name: "embed default mode", ownerEmail: "embed@example.com", wantErr: true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			opts := &adminv1.ReportOptions{
				DisplayName:     "Test",
				RefreshCron:     "0 * * * *",
				QueryName:       "MetricsViewAggregation",
				QueryArgsJson:   "{}",
				EmailRecipients: []string{"embed@example.com"},
				WebOpenMode:     string(tc.webOpenMode),
			}
			data, err := (&Server{}).yamlForManagedReport(opts, tc.ownerUserID, tc.ownerEmail, "", nil)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var res reportYAML
			require.NoError(t, yaml.Unmarshal(data, &res))
			require.Equal(t, tc.ownerUserID, res.Annotations.AdminOwnerUserID)
			require.Equal(t, tc.ownerEmail, res.Annotations.AdminOwnerUserEmail)
			require.Empty(t, res.For.UserID)
			require.Empty(t, res.Query.For.UserID)
			require.Equal(t, tc.ownerEmail, res.For.UserEmail)
			require.Equal(t, tc.ownerEmail, res.Query.For.UserEmail)
		})
	}
}

func TestYAMLForManagedReportExportAttach(t *testing.T) {
	s := &Server{}
	opts := &adminv1.ReportOptions{
		DisplayName:     "Weekly report",
		RefreshCron:     "0 9 * * 1",
		QueryName:       "MetricsViewAggregation", // nolint:staticcheck // backwards compatibility
		QueryArgsJson:   `{"metrics_view":"mv1"}`, // nolint:staticcheck // backwards compatibility
		ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		ExportAttach:    true,
		EmailRecipients: []string{"user_1@example.com"},
		Explore:         "e1",
		WebOpenMode:     string(WebOpenModeCreator),
	}

	// The export is attached for reports run as the creator
	data, err := s.yamlForManagedReport(opts, "owner", "", "", nil)
	require.NoError(t, err)
	var res reportYAML
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.True(t, res.Export.Attach)

	// The property is omitted when the export is not attached, so runtimes that predate it can still parse the YAML
	opts.ExportAttach = false
	data, err = s.yamlForManagedReport(opts, "owner", "", "", nil)
	require.NoError(t, err)
	require.NotContains(t, string(data), "attach")
	opts.ExportAttach = true

	// Attachments are not supported for reports run as the recipient
	opts.WebOpenMode = string(WebOpenModeRecipient)
	_, err = s.yamlForManagedReport(opts, "owner", "", "", nil)
	require.ErrorContains(t, err, "only supported for reports run as the creator")
	opts.WebOpenMode = string(WebOpenModeCreator)

	// The export is always attached for reports created from an embed, since the email has no download link
	opts.ExportAttach = false
	opts.EmailRecipients = []string{"embed@example.com"}
	data, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.NoError(t, err)
	res = reportYAML{}
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.True(t, res.Export.Attach)
	opts.ExportAttach = true
	opts.EmailRecipients = []string{"user_1@example.com"}

	// Attachments are not supported for canvas PDF reports
	_, err = s.yamlForManagedReport(&adminv1.ReportOptions{
		DisplayName:     "Canvas report",
		RefreshCron:     "0 9 * * 1",
		ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_PDF,
		ExportAttach:    true,
		EmailRecipients: []string{"user_1@example.com"},
		Canvas:          "c1",
		WebOpenMode:     string(WebOpenModeCreator),
	}, "owner", "", "", nil)
	require.ErrorContains(t, err, "only supported for query-based reports")
}

func TestYAMLForManagedReportEmbed(t *testing.T) {
	s := &Server{}
	newOpts := func() *adminv1.ReportOptions {
		return &adminv1.ReportOptions{
			DisplayName:     "Embed report",
			RefreshCron:     "0 9 * * 1",
			QueryName:       "MetricsViewAggregation", // nolint:staticcheck // backwards compatibility
			QueryArgsJson:   `{"metrics_view":"mv1"}`, // nolint:staticcheck // backwards compatibility
			ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			EmailRecipients: []string{"Embed@Example.com"},
			Explore:         "e1",
			WebOpenMode:     string(WebOpenModeCreator),
		}
	}

	// The open URL of the embed is copied into the annotations
	data, err := s.yamlForManagedReport(newOpts(), "", "embed@example.com", "https://app.acme.com/open?tenant=foo", nil)
	require.NoError(t, err)
	var res reportYAML
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, "https://app.acme.com/open?tenant=foo", res.Annotations.AdminEmbedOpenURL)
	require.Equal(t, "https://app.acme.com/open?tenant=foo", parseReportAnnotations(map[string]string{"admin_embed_open_url": "https://app.acme.com/open?tenant=foo"}).AdminEmbedOpenURL)

	// The annotation is omitted when the embed has no open URL
	data, err = s.yamlForManagedReport(newOpts(), "", "embed@example.com", "", nil)
	require.NoError(t, err)
	require.NotContains(t, string(data), "admin_embed_open_url")

	// The open URL only applies to reports created from an embed
	data, err = s.yamlForManagedReport(newOpts(), "owner", "", "https://app.acme.com/open", nil)
	require.NoError(t, err)
	require.NotContains(t, string(data), "admin_embed_open_url")

	// No recipients is allowed
	opts := newOpts()
	opts.EmailRecipients = nil
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.NoError(t, err)

	// Only the creator can be a recipient
	opts = newOpts()
	opts.EmailRecipients = []string{"other@example.com"}
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.ErrorContains(t, err, "can only be sent to the creator")
	opts.EmailRecipients = []string{"embed@example.com", "other@example.com"}
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.ErrorContains(t, err, "can only be sent to the creator")

	// Slack is not supported
	opts = newOpts()
	opts.SlackChannels = []string{"#general"}
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.ErrorContains(t, err, "can't be sent to Slack")
	opts = newOpts()
	opts.SlackUsers = []string{"embed@example.com"}
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.ErrorContains(t, err, "can't be sent to Slack")
	opts = newOpts()
	opts.SlackWebhooks = []string{"https://hooks.slack.com/x"}
	_, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.ErrorContains(t, err, "can't be sent to Slack")

	// The restrictions don't apply to reports created by Rill users
	opts = newOpts()
	opts.EmailRecipients = []string{"other@example.com", "another@example.com"}
	opts.SlackChannels = []string{"#general"}
	_, err = s.yamlForManagedReport(opts, "owner", "", "", nil)
	require.NoError(t, err)
}

// TestEmbedReportResources tests that reports created from an embed limited to specific resources must target one of its dashboards,
// and that the dashboards are saved in the report's annotations.
func TestEmbedReportResources(t *testing.T) {
	e1 := database.ResourceName{Type: runtime.ResourceKindExplore, Name: "e1"}
	c1 := database.ResourceName{Type: runtime.ResourceKindCanvas, Name: "c1"}
	r1 := database.ResourceName{Type: runtime.ResourceKindReport, Name: "r1"}

	// An embed that isn't limited can target any dashboard
	res, err := embedReportResources(nil, &adminv1.ReportOptions{Explore: "e2"})
	require.NoError(t, err)
	require.Empty(t, res)

	// The target must be one of the embed's dashboards, of the same kind
	res, err = embedReportResources([]database.ResourceName{e1}, &adminv1.ReportOptions{Explore: "e1"})
	require.NoError(t, err)
	require.Equal(t, []database.ResourceName{e1}, res)
	_, err = embedReportResources([]database.ResourceName{e1}, &adminv1.ReportOptions{Explore: "e2"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	_, err = embedReportResources([]database.ResourceName{e1}, &adminv1.ReportOptions{Canvas: "e1"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	_, err = embedReportResources([]database.ResourceName{e1}, &adminv1.ReportOptions{})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	res, err = embedReportResources([]database.ResourceName{c1}, &adminv1.ReportOptions{Canvas: "c1"})
	require.NoError(t, err)
	require.Equal(t, []database.ResourceName{c1}, res)

	// Reports are dropped, since the embed user controls their queries; for example, when the embed opens a report
	res, err = embedReportResources([]database.ResourceName{r1, e1}, &adminv1.ReportOptions{Explore: "e1"})
	require.NoError(t, err)
	require.Equal(t, []database.ResourceName{e1}, res)
	_, err = embedReportResources([]database.ResourceName{r1, e1}, &adminv1.ReportOptions{Explore: "r1"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	// The resources are saved in the annotations of reports created from an embed
	opts := &adminv1.ReportOptions{
		DisplayName:     "Embed report",
		RefreshCron:     "0 9 * * 1",
		QueryName:       "MetricsViewAggregation", // nolint:staticcheck // backwards compatibility
		QueryArgsJson:   `{"metrics_view":"mv1"}`, // nolint:staticcheck // backwards compatibility
		ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
		EmailRecipients: []string{"embed@example.com"},
		Explore:         "e1",
		WebOpenMode:     string(WebOpenModeCreator),
	}
	s := &Server{}
	data, err := s.yamlForManagedReport(opts, "", "embed@example.com", "", []database.ResourceName{e1})
	require.NoError(t, err)
	var rep reportYAML
	require.NoError(t, yaml.Unmarshal(data, &rep))
	parsed, err := parseEmbedResources(rep.Annotations.AdminEmbedResources)
	require.NoError(t, err)
	require.Equal(t, []database.ResourceName{e1}, parsed)

	// The annotation is omitted when the embed isn't limited
	data, err = s.yamlForManagedReport(opts, "", "embed@example.com", "", nil)
	require.NoError(t, err)
	require.NotContains(t, string(data), "admin_embed_resources")
}
