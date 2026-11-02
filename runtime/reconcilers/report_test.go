package reconcilers_test

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestReportAIExploreWatermarkInherit(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/explores/e1.yaml": `
type: explore
metrics_view: mv1
`,
		// AI report that inherits its watermark through the explore
		"/reports/r1.yaml": `
type: report
display_name: AI Report
refresh:
  cron: 0 8 * * *
watermark: inherit
data:
  ai:
    prompt: Analyze key metrics
    explore: e1
notify:
  email:
    recipients:
      - somebody@example.com
`,
		// Same report without watermark: inherit, which should use the trigger time
		"/reports/r2.yaml": `
type: report
display_name: AI Report
refresh:
  cron: 0 8 * * *
data:
  ai:
    prompt: Analyze key metrics
    explore: e1
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)

	r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.Meta.Refs, 1)
	require.Equal(t, runtime.ResourceKindExplore, r1.Meta.Refs[0].Kind)
	require.Equal(t, "e1", r1.Meta.Refs[0].Name)

	// Trigger both reports.
	// The execution's report time is recorded before the report is sent, so it is available regardless of whether the AI session could be created.
	triggerTime := time.Now()
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"})

	// r1 should have resolved the watermark of the explore's metrics view.
	// The watermark is max(__time) plus one second, since it's used as an exclusive upper bound (see ResourceWatermark).
	r1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.GetReport().State.ExecutionHistory, 1)
	require.Equal(t, time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC), r1.GetReport().State.ExecutionHistory[0].ReportTime.AsTime())

	// r2 should have used the trigger time
	r2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r2")
	require.Len(t, r2.GetReport().State.ExecutionHistory, 1)
	require.WithinDuration(t, triggerTime, r2.GetReport().State.ExecutionHistory[0].ReportTime.AsTime(), time.Minute)
}

func TestReportAIResolveTransitiveAccess(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP as __time, 'Denmark' as country
`,
		"/metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- expression: count(*)
`,
		"/explores/e1.yaml": `
type: explore
metrics_view: mv1
`,
		// A minimal AI report as it may be written by hand: no agent (defaults to the analyst agent) and no prompt (reports have a default prompt).
		"/reports/r1.yaml": `
type: report
display_name: AI Report
refresh:
  cron: 0 8 * * *
data:
  ai:
    explore: e1
notify:
  email:
    recipients:
      - somebody@example.com
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	// Opening a report resolves transitive access through the report, which initializes its resolver.
	// This must apply the same defaults as executing the report, or the resolver rejects the properties and the report cannot be opened.
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"},
					},
				},
			},
		},
	}
	r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	sec, err := rt.ResolveSecurity(t.Context(), id, claims, r1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
}

func TestReportCanvasResolveTransitiveAccess(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create a canvas with components referencing mv1 and mv2, an unrelated metrics view mv3,
	// and a canvas PDF report referencing the canvas.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"m3.sql": `SELECT 'baz' as baz, 3 as z`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
`,
		"mv3.yaml": `
version: 1
type: metrics_view
model: m3
dimensions:
- column: baz
measures:
- name: z
  expression: sum(z)
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - custom_chart:
          metrics_sql: "SELECT bar, y FROM mv2"
`,
		"r1.yaml": `
type: report
display_name: My Canvas PDF Report
export:
  format: pdf
notify:
  email:
    recipients:
      - user_1@example.com
annotations:
  canvas: c1
  metrics_view_filters: '{"mv1":{"cond":{"op":"OPERATION_EQ","exprs":[{"ident":"foo"},{"val":"foo"}]}}}'
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 11, 0, 0)

	// Build claims with a transitive access rule on the report
	ctx := t.Context()
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: runtime.ResourceKindReport,
							Name: "r1",
						},
					},
				},
			},
		},
	}

	// The canvas referenced by the report should be accessible
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	sec, err := rt.ResolveSecurity(ctx, id, claims, c1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// The canvas's components should be accessible
	for _, ref := range c1.Meta.Refs {
		if ref.Kind != runtime.ResourceKindComponent {
			continue
		}
		comp := testruntime.GetResource(t, rt, id, ref.Kind, ref.Name)
		sec, err = rt.ResolveSecurity(ctx, id, claims, comp)
		require.NoError(t, err)
		require.True(t, sec.CanAccess())
	}

	// The metrics views referenced by the canvas's components should be accessible.
	// The filters captured in the report's "metrics_view_filters" annotation should be applied as a query filter.
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
	require.NotNil(t, sec.QueryFilter())

	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv2)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
	require.Nil(t, sec.QueryFilter())

	// A metrics view not referenced by the canvas should not be accessible
	mv3 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv3")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv3)
	require.NoError(t, err)
	require.False(t, sec.CanAccess())
}

func TestReportExportAttach(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AdminService: &creatorModeAdmin{
			userAttrs:  map[string]map[string]any{"owner": {"domain": "example.com"}},
			embedAttrs: map[string]map[string]any{"r3": {"email": "embed@other.com", "domain": "other.com", "admin": false}},
		},
		Files: map[string]string{
			"rill.yaml": "features:\n  reports: true\n", // Reports created from an embed require the feature flag
			"models/bar.sql": `
SELECT * FROM (VALUES
	('2024-01-01T00:00:00Z'::TIMESTAMP, 'Denmark', 'example.com'),
	('2024-01-01T00:00:00Z'::TIMESTAMP, 'Sweden', 'other.com')
) AS t(__time, country, domain)
`,
			// The row filter checks that the export is run with the report owner's attributes
			"metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
- column: domain
measures:
- name: total
  expression: count(*)
security:
  access: true
  row_filter: domain = '{{ .user.domain }}'
`,
			"explores/e1.yaml": `
type: explore
metrics_view: mv1
`,
			"reports/r1.yaml": `
type: report
display_name: Report with owner
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - user_1@example.com
    - user_2@example.com
annotations:
  admin_owner_user_id: owner
  web_open_mode: creator
  explore: e1
`,
			// Without an owner, there are no attributes to run the export with
			"reports/r2.yaml": `
type: report
display_name: Report without owner
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - user_1@example.com
annotations:
  web_open_mode: creator
`,
			// Reports created from an embed are run as the embed user, and only have an unsubscribe link
			"reports/r3.yaml": `
type: report
display_name: Report from embed
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@other.com
    - colleague@other.com
annotations:
  admin_embed_owner_id: ext_abc
  web_open_mode: creator
  explore: e1
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 7, 0, 0)

	for _, name := range []string{"r1", "r2", "r3"} {
		testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: name})
		if name == "r2" {
			continue
		}
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, name)
		require.Len(t, r.GetReport().State.ExecutionHistory, 1)
		require.Empty(t, r.GetReport().State.ExecutionHistory[0].ErrorMessage)
		require.Empty(t, r.GetReport().State.ExecutionHistory[0].Warnings)
	}

	r2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r2")
	require.Len(t, r2.GetReport().State.ExecutionHistory, 1)
	require.Empty(t, r2.GetReport().State.ExecutionHistory[0].ErrorMessage)
	require.Equal(t, []string{"The export was not attached to emails because the report has no owner or identity to run as"}, r2.GetReport().State.ExecutionHistory[0].Warnings)

	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 5)
	var recipients, embedRecipients []string
	for _, e := range emails {
		switch {
		case strings.HasPrefix(e.Subject, "Report with owner"):
			recipients = append(recipients, e.ToEmail)
			require.Contains(t, e.Body, "The CSV file is attached.")
			require.Len(t, e.Attachments, 1)
			require.Regexp(t, `^mv1_filtered_\d{14}\.csv$`, e.Attachments[0].Filename)
			require.Equal(t, "text/csv", e.Attachments[0].ContentType)
			require.Equal(t, "Country,Total\nDenmark,1\n", string(e.Attachments[0].Data))
		case strings.HasPrefix(e.Subject, "Report without owner"):
			require.Equal(t, "user_1@example.com", e.ToEmail)
			require.NotContains(t, e.Body, "The CSV file is attached.")
			require.Empty(t, e.Attachments)
		case strings.HasPrefix(e.Subject, "Report from embed"):
			embedRecipients = append(embedRecipients, e.ToEmail)
			require.Contains(t, e.Body, "is ready.")
			require.Contains(t, e.Body, "The CSV file is attached.")
			require.Len(t, e.Attachments, 1)
			require.Equal(t, "Country,Total\nSweden,1\n", string(e.Attachments[0].Data))
			require.Contains(t, e.Body, "https://example.com/unsubscribe")
			require.NotContains(t, e.Body, "https://example.com/open")
			require.NotContains(t, e.Body, "https://example.com/edit")
			require.NotContains(t, e.Body, "https://example.com/export")
		default:
			require.Fail(t, "unexpected email", e.Subject)
		}
	}
	require.ElementsMatch(t, []string{"user_1@example.com", "user_2@example.com"}, recipients)
	require.ElementsMatch(t, []string{"embed@other.com", "colleague@other.com"}, embedRecipients)
}

func TestReportExportAttachRestrictedResources(t *testing.T) {
	e1 := []*runtimev1.ResourceName{{Kind: runtime.ResourceKindExplore, Name: "e1"}}
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AdminService: &creatorModeAdmin{
			embedAttrs: map[string]map[string]any{"r1": {"email": "embed@example.com"}, "r2": {"email": "embed@example.com"}},
			resources:  map[string][]*runtimev1.ResourceName{"r1": e1, "r2": e1},
		},
		Files: map[string]string{
			"rill.yaml": "features:\n  reports: true\n", // Reports created from an embed require the feature flag
			"models/bar.sql": `
SELECT * FROM (VALUES
	('2024-01-01T00:00:00Z'::TIMESTAMP, 'Denmark')
) AS t(__time, country)
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
			"explores/e2.yaml": `
type: explore
metrics_view: mv2
`,
			// Queries the metrics view of the explore it is restricted to
			"reports/r1.yaml": `
type: report
display_name: Report on allowed metrics view
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_embed_owner_id: ext_abc
  web_open_mode: creator
  explore: e1
`,
			// Queries a metrics view outside the explore it is restricted to
			"reports/r2.yaml": `
type: report
display_name: Report on other metrics view
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv2
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_embed_owner_id: ext_abc
  web_open_mode: creator
  explore: e1
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"})
	r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.GetReport().State.ExecutionHistory, 1)
	require.Empty(t, r1.GetReport().State.ExecutionHistory[0].ErrorMessage)

	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"})
	r2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r2")
	require.NotEmpty(t, r2.GetReport().State.ExecutionHistory)
	require.Contains(t, r2.GetReport().State.ExecutionHistory[0].ErrorMessage, "failed to export attachment: action not allowed")

	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 1)
	require.True(t, strings.HasPrefix(emails[0].Subject, "Report on allowed metrics view"))
	require.Len(t, emails[0].Attachments, 1)
	require.Equal(t, "Country,Total\nDenmark,1\n", string(emails[0].Attachments[0].Data))
}

func TestReportExportAttachTooLarge(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AdminService: &creatorModeAdmin{
			userAttrs:  map[string]map[string]any{"owner": {"email": "owner@example.com"}},
			embedAttrs: map[string]map[string]any{"r2": {"email": "embed@example.com"}},
		},
		Files: map[string]string{
			"rill.yaml": "features:\n  reports: true\n", // Reports created from an embed require the feature flag
			// Exports to a CSV file of more than 10 MB
			"models/bar.sql": `
SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, range AS id, repeat('x', 100) AS padding FROM range(120000)
`,
			"metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: id
- column: padding
measures:
- name: total
  expression: count(*)
`,
			"reports/r1.yaml": `
type: report
display_name: Large report
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: id
    - name: padding
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - user_1@example.com
annotations:
  admin_owner_user_id: owner
`,
			// Reports created from an embed have no download link to fall back to
			"reports/r2.yaml": `
type: report
display_name: Large embed report
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: id
    - name: padding
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_embed_owner_id: ext_abc
  web_open_mode: creator
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"})

	// The report is still sent, but without the attachment
	r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
	require.Len(t, r1.GetReport().State.ExecutionHistory, 1)
	require.Empty(t, r1.GetReport().State.ExecutionHistory[0].ErrorMessage)
	require.Equal(t, []string{"The export was not attached to emails because it is larger than 10 MB"}, r1.GetReport().State.ExecutionHistory[0].Warnings)

	r2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r2")
	require.Len(t, r2.GetReport().State.ExecutionHistory, 1)
	require.Empty(t, r2.GetReport().State.ExecutionHistory[0].ErrorMessage)
	require.Equal(t, []string{"The export was not attached to emails because it is larger than 10 MB"}, r2.GetReport().State.ExecutionHistory[0].Warnings)

	emails := rt.Email.Sender.(*email.TestSender).Emails
	require.Len(t, emails, 2)
	for _, e := range emails {
		require.Empty(t, e.Attachments)
		switch {
		case strings.HasPrefix(e.Subject, "Large report"):
			require.Contains(t, e.Body, "The CSV file is too large to attach to this email, but you can download it below.")
			require.Contains(t, e.Body, "https://example.com/export")
		case strings.HasPrefix(e.Subject, "Large embed report"):
			require.Contains(t, e.Body, "The CSV file is too large to attach to this email. Try lowering the row limit of the report.")
			require.NotContains(t, e.Body, "https://example.com/export")
		default:
			require.Fail(t, "unexpected email", e.Subject)
		}
	}
}

func TestReportEmbedFeatureFlag(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%v", enabled), func(t *testing.T) {
			rillYAML := ""
			if enabled {
				rillYAML = "features:\n  reports: true\n"
			}
			rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				AdminService: &creatorModeAdmin{embedAttrs: map[string]map[string]any{"r1": {"email": "embed@example.com"}}},
				Files: map[string]string{
					"rill.yaml":      rillYAML,
					"models/bar.sql": `SELECT 'US' AS country`,
					"metrics/mv1.yaml": `
version: 1
type: metrics_view
model: bar
dimensions:
- column: country
measures:
- name: total
  expression: count(*)
`,
					"reports/r1.yaml": `
type: report
display_name: Embed report
query:
  name: MetricsViewAggregation
  args:
    metrics_view: mv1
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_embed_owner_id: ext_abc
  web_open_mode: creator
`,
				},
			})
			testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
			testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"})

			// Reports created from an embed only run if the "reports" feature flag is enabled for embed users
			r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
			require.NotEmpty(t, r1.GetReport().State.ExecutionHistory)
			emails := rt.Email.Sender.(*email.TestSender).Emails
			if enabled {
				require.Empty(t, r1.GetReport().State.ExecutionHistory[0].ErrorMessage)
				require.Len(t, emails, 1)
			} else {
				require.Contains(t, r1.GetReport().State.ExecutionHistory[0].ErrorMessage, `"reports" feature flag`)
				require.Empty(t, emails)
			}
		})
	}
}

// creatorModeAdmin is an admin service that returns report metadata the way the admin service does for reports run as the creator:
// every recipient's delivery has the user ID and attributes of the report owner,
// or for reports created from an embed, the attributes of the embed user marked as an embed user's.
type creatorModeAdmin struct {
	drivers.AdminService                                      // Not implemented; calls to methods that are not overridden panic
	userAttrs            map[string]map[string]any            // User ID -> attributes
	embedAttrs           map[string]map[string]any            // Report name -> attributes of the embed user who created the report
	resources            map[string][]*runtimev1.ResourceName // Report name -> resources the report's data is restricted to
}

func (a *creatorModeAdmin) GetConfig(ctx context.Context) (*drivers.Config, error) {
	return &drivers.Config{}, nil
}

func (a *creatorModeAdmin) GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	var userID string
	var attrs map[string]any
	if embedAttrs, ok := a.embedAttrs[reportName]; ok {
		attrs = maps.Clone(embedAttrs)
		attrs["embed"] = true
	} else if ownerID != "" {
		userID = ownerID
		attrs = a.userAttrs[ownerID]
	}

	resources, restrict := a.resources[reportName]
	delivery := make(map[string]drivers.ReportDelivery)
	for _, recipient := range emailRecipients {
		delivery[recipient] = drivers.ReportDelivery{
			OpenURL:           "https://example.com/open",
			ExportURL:         "https://example.com/export",
			EditURL:           "https://example.com/edit",
			UnsubscribeURL:    "https://example.com/unsubscribe",
			UserID:            userID,
			UserAttrs:         attrs,
			RestrictResources: restrict,
			Resources:         resources,
		}
	}
	return &drivers.ReportMetadata{ReportDelivery: delivery}, nil
}
