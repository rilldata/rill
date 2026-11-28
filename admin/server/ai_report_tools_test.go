package server_test

import (
	"encoding/json"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestAIReportTools tests creating, listing and deleting scheduled reports with the AI tools, as the chat does:
// through the admin service, which writes them like the report dialog, and the runtime, which parses and reconciles them.
func TestAIReportTools(t *testing.T) {
	f := newForUserFixture(t)
	ctx := t.Context()

	runner := ai.NewRunner(f.fix.Runtime, activity.NewNoopClient())
	s, err := runner.Session(ctx, &ai.SessionOptions{
		InstanceID:        f.instance.ID,
		CreateIfNotExists: true,
		Claims: &runtime.SecurityClaims{
			UserID:         f.viewerUser.ID,
			UserAttributes: map[string]any{"email": f.viewerUser.Email},
			Permissions:    []runtime.Permission{runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics},
		},
		UserAgent: "rill/test",
	})
	require.NoError(t, err)

	createReport := func(args *ai.CreateReportArgs) (*ai.CreatedResource, error) {
		var res *ai.CreatedResource
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateReportName, &res, args)
		return res, err
	}
	weekly := &ai.Schedule{Cron: "0 9 * * 1", TimeZone: "Europe/Madrid"}

	var dataReport, aiReport string
	t.Run("creates an explore data report like the dialog", func(t *testing.T) {
		res, err := createReport(&ai.CreateReportArgs{
			DisplayName: "Revenue by country",
			Kind:        "explore_data",
			Explore:     "e1",
			Measures:    []string{"revenue"},
			Dimensions:  []string{"country"},
			TimeRange:   "7D as of latest/D",
			Where:       []ai.DimensionFilter{{Dimension: "country", Values: []string{"Denmark"}}},
			Export:      &ai.ReportExport{Format: "xlsx", IncludeHeader: true},
			Schedule:    weekly,
		})
		require.NoError(t, err)
		require.Empty(t, res.ReconcileError)
		dataReport = res.Name

		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindReport, res.Name).GetReport().Spec
		require.Equal(t, runtimev1.ExportFormat_EXPORT_FORMAT_XLSX, spec.ExportFormat)
		require.True(t, spec.ExportIncludeHeader)
		require.Equal(t, "recipient", spec.Annotations["web_open_mode"])
		require.Equal(t, "e1", spec.Annotations["explore"])
		require.Equal(t, []any{f.viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])
		var query map[string]any
		require.NoError(t, json.Unmarshal([]byte(spec.ResolverProperties.AsMap()["query_args_json"].(string)), &query))
		require.JSONEq(t, `{
			"metricsView": "mv1",
			"measures": [{"name": "revenue"}],
			"dimensions": [{"name": "country"}],
			"where": {"cond": {"op": "OPERATION_AND", "exprs": [{"cond": {"op": "OPERATION_IN", "exprs": [{"ident": "country"}, {"val": "Denmark"}]}}]}},
			"timeRange": {"expression": "7D as of latest/D", "timeZone": "Europe/Madrid"},
			"sort": [{"name": "revenue", "desc": true}]
		}`, mustJSON(t, query))
	})

	t.Run("creates an AI report", func(t *testing.T) {
		res, err := createReport(&ai.CreateReportArgs{
			DisplayName: "Weekly insights",
			Kind:        "ai",
			Explore:     "e1",
			Prompt:      "Focus on the countries that changed the most",
			TimeRange:   "7D as of latest/D",
			Comparison:  "previous_week",
			Schedule:    weekly,
		})
		require.NoError(t, err)
		require.Empty(t, res.ReconcileError)
		aiReport = res.Name

		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindReport, res.Name).GetReport().Spec
		require.Equal(t, "ai", spec.Resolver)
		require.Equal(t, map[string]any{
			"prompt":                "Focus on the countries that changed the most",
			"explore":               "e1",
			"time_range":            map[string]any{"expression": "7D as of latest/D"},
			"comparison_time_range": map[string]any{"expression": "7D as of latest/D offset -1W"},
			"time_zone":             "Europe/Madrid",
		}, spec.ResolverProperties.AsMap())
	})

	t.Run("only opens with the user's permissions when asked to", func(t *testing.T) {
		args := &ai.CreateReportArgs{
			DisplayName: "Revenue",
			Kind:        "explore_data",
			Explore:     "e1",
			Measures:    []string{"revenue"},
			TimeRange:   "7D as of latest/D",
			Export:      &ai.ReportExport{Attach: true},
			Schedule:    weekly,
		}
		_, err := createReport(args)
		require.ErrorContains(t, err, "open_mode 'creator'")

		args.Export = &ai.ReportExport{Format: "parquet", IncludeHeader: true}
		_, err = createReport(args)
		require.ErrorContains(t, err, "Parquet export can't include a header")

		args.Export = nil
		args.Recipients = &ai.Recipients{Emails: []string{f.outsiderUser.Email}}
		args.OpenMode = "creator"
		_, err = createReport(args)
		require.ErrorContains(t, err, "not a member of the project")
	})

	t.Run("lists and deletes the user's own reports", func(t *testing.T) {
		var list *ai.ListReportsResult
		_, err := s.CallTool(ctx, ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
		require.NoError(t, err)
		require.Len(t, list.Reports, 2)
		for _, r := range list.Reports {
			require.Contains(t, []string{dataReport, aiReport}, r.Name)
			require.Equal(t, "e1", r.Dashboard)
			require.Equal(t, "0 9 * * 1", r.Schedule)
		}

		deleteReport := func(confirmed bool) *ai.DeleteResourceResult {
			var res *ai.DeleteResourceResult
			_, err := s.CallTool(ctx, ai.RoleUser, ai.DeleteReportName, &res, &ai.DeleteResourceArgs{Name: aiReport, Confirmed: confirmed})
			require.NoError(t, err)
			return res
		}
		require.Equal(t, "confirmation_required", deleteReport(false).Status)
		require.Equal(t, "deleted", deleteReport(true).Status)

		_, err = s.CallTool(ctx, ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
		require.NoError(t, err)
		require.Len(t, list.Reports, 1)
	})
}
