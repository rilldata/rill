package reconcilers_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/embedopen"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// embedAlertMetricsFiles are project files with a metrics view that alerts can query.
var embedAlertMetricsFiles = map[string]string{
	"models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, 'Denmark' AS country
`,
	"metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- name: total
  expression: count(*)
`,
	"metrics/mv2.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- name: total
  expression: count(*)
`,
	"explores/e1.yaml": `
type: explore
metrics_view: mv1
`,
}

// embedAlertYAML returns an alert created from an embed that always fails, so it always notifies.
func embedAlertYAML(displayName, metricsView string) string {
	return fmt.Sprintf(`
type: alert
display_name: %s
refs:
- type: MetricsView
  name: %s
query:
  name: MetricsViewAggregation
  args:
    metrics_view: %s
    dimensions:
    - name: country
    measures:
    - name: total
notify:
  email:
    recipients:
    - creator@other.com
    - recipient@other.com
annotations:
  admin_managed: true
  admin_embed_owner_id: ext_abc
  web_open_path: /explore/e1
`, displayName, metricsView, metricsView)
}

func TestAlertEmbedOpenLink(t *testing.T) {
	files := maps.Clone(embedAlertMetricsFiles)
	// Alerts created from an embed require the feature flag
	files["rill.yaml"] = `
features:
  alerts: true
embeds:
  open_url: https://app.example.org/open?tenant=foo
`
	files["alerts/a1.yaml"] = embedAlertYAML("Alert with creator email", "mv1")
	files["alerts/a2.yaml"] = embedAlertYAML("Alert without creator email", "mv1")
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AdminService: &embedAlertAdmin{
			embedAttrs: map[string]map[string]any{
				"a1": {"email": "creator@other.com"},
				"a2": {"email": "creator@other.com"},
			},
			// The creator of a1 saved their email (in a different case than the recipient), and the creator of a2 didn't.
			embedOwnerEmails: map[string]string{"a1": "Creator@other.com"},
		},
		Files: files,
	})
	testruntime.RequireReconcileState(t, rt, id, 7, 0, 0)

	emails := rt.Email.Sender.(*email.TestSender).Emails
	a1Emails, a2Emails := emails[:0:0], emails[:0:0]
	for _, e := range emails {
		switch {
		case strings.HasPrefix(e.Subject, "Alert with creator email"):
			a1Emails = append(a1Emails, e)
		case strings.HasPrefix(e.Subject, "Alert without creator email"):
			a2Emails = append(a2Emails, e)
		default:
			require.Fail(t, "unexpected email", e.Subject)
		}
	}

	// Only the creator's email has the open link to the embedder's app
	require.Len(t, a1Emails, 2)
	for _, e := range a1Emails {
		// No links to Rill Cloud
		require.NotContains(t, e.Body, "https://example.com/open")
		require.NotContains(t, e.Body, "https://example.com/edit")

		if e.ToEmail != "creator@other.com" {
			require.Equal(t, "recipient@other.com", e.ToEmail)
			require.Contains(t, e.Body, "https://example.com/unsubscribe")
			require.NotContains(t, e.Body, "https://app.example.org/open")
			continue
		}
		// The creator can't unsubscribe, so they get no unsubscribe link
		require.NotContains(t, e.Body, "https://example.com/unsubscribe")
		require.Contains(t, e.Body, "https://app.example.org/open?")
		require.Contains(t, e.Body, "tenant=foo")
		p := requireEmbedOpenPayload(t, e.Body)
		require.Equal(t, embedopen.KindAlert, p.Kind)
		require.Equal(t, "a1", p.Name)
		require.False(t, p.ExecutionTime.IsZero())
	}

	// Without a saved creator email, no email has the open link
	require.Len(t, a2Emails, 2)
	for _, e := range a2Emails {
		require.Contains(t, e.Body, "https://example.com/unsubscribe")
		require.NotContains(t, e.Body, "https://app.example.org/open")
		// The "Open in browser" button is hidden
		require.Contains(t, e.Body, `class="hidden"`)
	}
}

func TestAlertEmbedFeatureFlag(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%v", enabled), func(t *testing.T) {
			files := maps.Clone(embedAlertMetricsFiles)
			files["rill.yaml"] = ""
			if enabled {
				files["rill.yaml"] = "features:\n  alerts: true\n"
			}
			files["alerts/a1.yaml"] = embedAlertYAML("Embed alert", "mv1")
			rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				AdminService: &embedAlertAdmin{embedAttrs: map[string]map[string]any{"a1": {"email": "creator@other.com"}}},
				Files:        files,
			})

			// Alerts created from an embed only run if the "alerts" feature flag is enabled for embed users
			a1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindAlert, "a1").GetAlert()
			require.NotEmpty(t, a1.State.ExecutionHistory)
			emails := rt.Email.Sender.(*email.TestSender).Emails
			if enabled {
				testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
				require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, a1.State.ExecutionHistory[0].Result.Status)
				require.Len(t, emails, 2)
			} else {
				testruntime.RequireReconcileState(t, rt, id, 6, 1, 0)
				require.Contains(t, a1.State.ExecutionHistory[0].Result.ErrorMessage, `"alerts" feature flag`)
				require.Empty(t, emails)
			}
		})
	}
}

func TestAlertEmbedFailsClosed(t *testing.T) {
	files := maps.Clone(embedAlertMetricsFiles)
	files["rill.yaml"] = "features:\n  alerts: true\n"
	// The admin service fails to resolve the identity of the embed user who created a1 (e.g. because they are inactive)
	files["alerts/a1.yaml"] = embedAlertYAML("Alert without identity", "mv1")
	// a2 is restricted to the explore e1, but queries a metrics view outside of it
	files["alerts/a2.yaml"] = embedAlertYAML("Alert outside of resources", "mv2")
	// a3 is restricted to the explore e1, and queries its metrics view
	files["alerts/a3.yaml"] = embedAlertYAML("Alert inside of resources", "mv1")
	e1 := []*runtimev1.ResourceName{{Kind: runtime.ResourceKindExplore, Name: "e1"}}
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AdminService: &embedAlertAdmin{
			embedAttrs: map[string]map[string]any{"a2": {"email": "creator@other.com"}, "a3": {"email": "creator@other.com"}},
			resources:  map[string][]*runtimev1.ResourceName{"a2": e1, "a3": e1},
		},
		Files: files,
	})
	testruntime.RequireReconcileState(t, rt, id, 8, 1, 0)

	// The alert doesn't run without an identity, and doesn't notify
	a1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindAlert, "a1").GetAlert()
	require.NotEmpty(t, a1.State.ExecutionHistory)
	require.Contains(t, a1.State.ExecutionHistory[0].Result.ErrorMessage, "failed to get alert metadata")
	require.False(t, a1.State.ExecutionHistory[0].SentNotifications)

	// The alert's data is limited to the resources of the embed it was created from
	a2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindAlert, "a2").GetAlert()
	require.NotEmpty(t, a2.State.ExecutionHistory)
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR, a2.State.ExecutionHistory[0].Result.Status)
	a3 := testruntime.GetResource(t, rt, id, runtime.ResourceKindAlert, "a3").GetAlert()
	require.NotEmpty(t, a3.State.ExecutionHistory)
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, a3.State.ExecutionHistory[0].Result.Status)

	for _, e := range rt.Email.Sender.(*email.TestSender).Emails {
		require.False(t, strings.HasPrefix(e.Subject, "Alert without identity"), "unexpected email: %s", e.Subject)
	}
}

// embedAlertAdmin is an admin service that returns alert metadata the way the admin service does for alerts created from an embed:
// the alert runs with the attributes of the embed user, marked as an embed user's,
// and only the creator's saved email is marked as the creator's; it gets no unsubscribe link.
// It fails for alerts without embed attributes, like the admin service does for alerts whose embed user is missing or inactive.
type embedAlertAdmin struct {
	drivers.AdminService                                      // Not implemented; calls to methods that are not overridden panic
	embedAttrs           map[string]map[string]any            // Alert name -> attributes of the embed user who created the alert
	embedOwnerEmails     map[string]string                    // Alert name -> saved email of the embed user who created the alert
	resources            map[string][]*runtimev1.ResourceName // Alert name -> resources the alert's data is restricted to
}

func (a *embedAlertAdmin) GetConfig(ctx context.Context) (*drivers.Config, error) {
	return &drivers.Config{}, nil
}

func (a *embedAlertAdmin) GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	embedAttrs, ok := a.embedAttrs[alertName]
	if !ok {
		return nil, errors.New("the alert is paused because its owner hasn't opened an embedded dashboard in 90 days")
	}
	attrs := maps.Clone(embedAttrs)
	attrs["embed"] = true

	resources, restrict := a.resources[alertName]
	urls := make(map[string]drivers.AlertURLs)
	for _, recipient := range emailRecipients {
		owner := a.embedOwnerEmails[alertName] != "" && strings.EqualFold(recipient, a.embedOwnerEmails[alertName])
		u := drivers.AlertURLs{EmbedOwner: owner}
		if !owner {
			u.UnsubscribeURL = "https://example.com/unsubscribe"
		}
		urls[recipient] = u
	}
	return &drivers.AlertMetadata{
		RecipientURLs:      urls,
		QueryForAttributes: attrs,
		RestrictResources:  restrict,
		Resources:          resources,
	}, nil
}
