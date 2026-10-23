package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLForManagedAlertEmbed(t *testing.T) {
	s := &Server{}
	newOpts := func() *adminv1.AlertOptions {
		return &adminv1.AlertOptions{
			DisplayName:     "Embed alert",
			MetricsViewName: "mv1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   `{"metrics_view":"mv1"}`,
			EmailRecipients: []string{"embed@example.com"},
			WebOpenPath:     "/explore/e1",
		}
	}

	// The alert runs for the embed user, and the open URL of the embed is copied into the annotations
	data, err := s.yamlForManagedAlert(newOpts(), "", "embed@example.com", "https://app.acme.com/open?tenant=foo")
	require.NoError(t, err)
	var res alertYAML
	require.NoError(t, yaml.Unmarshal(data, &res))
	require.Equal(t, "embed@example.com", res.For.UserEmail)
	require.Equal(t, "embed@example.com", res.Annotations.AdminOwnerUserEmail)
	require.Equal(t, "https://app.acme.com/open?tenant=foo", res.Annotations.AdminEmbedOpenURL)
	require.Equal(t, "https://app.acme.com/open?tenant=foo", parseAlertAnnotations(map[string]string{"admin_embed_open_url": "https://app.acme.com/open?tenant=foo"}).AdminEmbedOpenURL)

	// The annotation is omitted when the embed has no open URL
	data, err = s.yamlForManagedAlert(newOpts(), "", "embed@example.com", "")
	require.NoError(t, err)
	require.NotContains(t, string(data), "admin_embed_open_url")

	// The open URL only applies to alerts created from an embed
	data, err = s.yamlForManagedAlert(newOpts(), "owner", "", "https://app.acme.com/open")
	require.NoError(t, err)
	require.NotContains(t, string(data), "admin_embed_open_url")

	// Only the creator can be a recipient
	opts := newOpts()
	opts.EmailRecipients = []string{"other@example.com"}
	_, err = s.yamlForManagedAlert(opts, "", "embed@example.com", "")
	require.ErrorContains(t, err, "can only be sent to the creator")
	opts.EmailRecipients = []string{"embed@example.com", "other@example.com"}
	_, err = s.yamlForManagedAlert(opts, "", "embed@example.com", "")
	require.ErrorContains(t, err, "can only be sent to the creator")

	// Slack is not supported
	opts = newOpts()
	opts.SlackChannels = []string{"#general"}
	_, err = s.yamlForManagedAlert(opts, "", "embed@example.com", "")
	require.ErrorContains(t, err, "can't be sent to Slack")
	opts = newOpts()
	opts.SlackUsers = []string{"embed@example.com"}
	_, err = s.yamlForManagedAlert(opts, "", "embed@example.com", "")
	require.ErrorContains(t, err, "can't be sent to Slack")
	opts = newOpts()
	opts.SlackWebhooks = []string{"https://hooks.slack.com/x"}
	_, err = s.yamlForManagedAlert(opts, "", "embed@example.com", "")
	require.ErrorContains(t, err, "can't be sent to Slack")

	// The restrictions don't apply to alerts created by Rill users
	opts = newOpts()
	opts.EmailRecipients = []string{"other@example.com", "another@example.com"}
	opts.SlackChannels = []string{"#general"}
	_, err = s.yamlForManagedAlert(opts, "owner", "", "")
	require.NoError(t, err)
}
