package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
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
				EmailRecipients: []string{"recipient@example.com"},
				WebOpenMode:     string(tc.webOpenMode),
			}
			data, err := (&Server{}).yamlForManagedReport(opts, tc.ownerUserID, tc.ownerEmail)
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
	data, err := s.yamlForManagedReport(opts, "owner", "")
	require.NoError(t, err)
	var res reportYAML
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.True(t, res.Export.Attach)

	// The property is omitted when the export is not attached, so runtimes that predate it can still parse the YAML
	opts.ExportAttach = false
	data, err = s.yamlForManagedReport(opts, "owner", "")
	require.NoError(t, err)
	require.NotContains(t, string(data), "attach")
	opts.ExportAttach = true

	// Attachments are not supported for reports run as the recipient
	opts.WebOpenMode = string(WebOpenModeRecipient)
	_, err = s.yamlForManagedReport(opts, "owner", "")
	require.ErrorContains(t, err, "only supported for reports run as the creator")
	opts.WebOpenMode = string(WebOpenModeCreator)

	// Attachments are not supported for canvas PDF reports
	_, err = s.yamlForManagedReport(&adminv1.ReportOptions{
		DisplayName:     "Canvas report",
		RefreshCron:     "0 9 * * 1",
		ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_PDF,
		ExportAttach:    true,
		EmailRecipients: []string{"user_1@example.com"},
		Canvas:          "c1",
		WebOpenMode:     string(WebOpenModeCreator),
	}, "owner", "")
	require.ErrorContains(t, err, "only supported for query-based reports")
}
