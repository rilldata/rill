package ai_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSaveReport(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)
	weekly := ai.Schedule{Cron: "0 9 * * 1", TimeZone: "Europe/Madrid"}

	t.Run("creates an explore data report that opens with each recipient's permissions", func(t *testing.T) {
		admin.reset()
		var res *ai.SaveReportResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, &res, &ai.SaveReportArgs{
			DisplayName: "Weekly sales",
			Kind:        "explore_data",
			Explore:     ptr("sales"),
			Dimensions:  []string{"country"},
			Measures:    []string{"revenue"},
			TimeRange:   ptr("7D as of latest/D"),
			Where:       []ai.DimensionFilter{{Dimension: "channel", Values: []string{"web"}, Exclude: true}},
			Export:      &ai.ReportExport{Format: "xlsx", Limit: 500, IncludeHeader: true},
			Schedule:    &weekly,
		})
		require.NoError(t, err)
		require.Equal(t, "report-1", res.Name)

		call := admin.lastReport(t)
		require.Equal(t, userID, call.userID)
		require.Empty(t, call.name)
		require.Equal(t, "recipient", call.opts.WebOpenMode)
		require.Equal(t, "sales", call.opts.Explore)
		require.Equal(t, "MetricsViewAggregation", call.opts.QueryName)
		require.Equal(t, runtimev1.ExportFormat_EXPORT_FORMAT_XLSX, call.opts.ExportFormat)
		require.Equal(t, uint64(500), call.opts.ExportLimit)
		require.True(t, call.opts.ExportIncludeHeader)
		require.False(t, call.opts.ExportAttach)
		require.Equal(t, "0 9 * * 1", call.opts.RefreshCron)
		require.Equal(t, "Europe/Madrid", call.opts.RefreshTimeZone)
		require.Equal(t, []string{"user@example.com"}, call.opts.EmailRecipients)

		q := parseAggregation(t, call.opts.QueryArgsJSON)
		require.Equal(t, "mv", q.MetricsView)
		require.Equal(t, "country", q.Dimensions[0].Name)
		require.Equal(t, "revenue", q.Measures[0].Name)
		require.Equal(t, "7D as of latest/D", q.TimeRange.Expression)
		require.Equal(t, "Europe/Madrid", q.TimeRange.TimeZone)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_NIN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}}`, q.Where)

		requireQueryRuns(t, rt, instanceID, call.opts.QueryArgsJSON)
	})

	t.Run("creates an AI report", func(t *testing.T) {
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, &ai.SaveReportArgs{
			DisplayName:             "Weekly insights",
			Kind:                    "ai",
			Explore:                 ptr("sales"),
			Prompt:                  ptr("Explain what drove revenue last week"),
			TimeRange:               ptr("7D as of latest/D"),
			CompareToPreviousPeriod: ptr(true),
			Schedule:                &weekly,
		})
		require.NoError(t, err)

		call := admin.lastReport(t)
		require.Equal(t, "ai", call.opts.Resolver)
		require.Empty(t, call.opts.QueryName)
		require.Equal(t, "recipient", call.opts.WebOpenMode)
		require.Equal(t, map[string]any{
			"prompt":                "Explain what drove revenue last week",
			"explore":               "sales",
			"time_range":            map[string]any{"expression": "7D as of latest/D"},
			"comparison_time_range": map[string]any{"expression": "7D as of latest/D offset -1P"},
			"time_zone":             "Europe/Madrid",
		}, call.opts.ResolverProperties)
	})

	t.Run("creates a canvas PDF report with the canvas filtered", func(t *testing.T) {
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, &ai.SaveReportArgs{
			DisplayName:   "Overview PDF",
			Kind:          "canvas_pdf",
			Canvas:        "Overview",
			TimeRange:     ptr("7D as of latest/D"),
			CanvasFilters: []ai.CanvasFilters{{MetricsView: "mv", Filters: []ai.DimensionFilter{{Dimension: "country", Values: []string{"US", "O'Hara"}}}}},
			Schedule:      &weekly,
		})
		require.NoError(t, err)

		call := admin.lastReport(t)
		require.Equal(t, "overview", call.opts.Canvas, "the canvas keeps its declared name")
		require.Empty(t, call.opts.Explore)
		require.Equal(t, runtimev1.ExportFormat_EXPORT_FORMAT_PDF, call.opts.ExportFormat)
		require.Len(t, call.opts.MetricsViewFilters, 1)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"country"},{"val":"US"},{"val":"O'Hara"}]}}]}}`, call.opts.MetricsViewFilters["mv"])

		// The canvas state the PDF renders and the report dialog edits.
		state, err := url.ParseQuery(call.opts.WebOpenState)
		require.NoError(t, err)
		require.Equal(t, "7D as of latest/D", state.Get("tr"))
		require.Equal(t, "Europe/Madrid", state.Get("tz"), "the canvas renders in the schedule's time zone")
		require.Equal(t, `country IN ('US','O\'Hara')`, state.Get("f.mv"))
		require.Equal(t, "true", state.Get("pdf_include_filters"))
		require.Equal(t, "true", state.Get("pdf_all_tabs"))
	})

	t.Run("opens with the creator's permissions when asked", func(t *testing.T) {
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, &ai.SaveReportArgs{
			DisplayName: "Weekly insights",
			Kind:        "ai",
			Schedule:    &weekly,
			Recipients:  &ai.Recipients{Emails: []string{"partner@example.org"}, SlackChannels: []string{"#sales"}},
			OpenMode:    "creator",
		})
		require.NoError(t, err)
		call := admin.lastReport(t)
		require.Equal(t, "creator", call.opts.WebOpenMode)
		require.Equal(t, []string{"partner@example.org"}, call.opts.EmailRecipients)
		require.Equal(t, []string{"#sales"}, call.opts.SlackChannels)
	})

	t.Run("invalid arguments are rejected before saving", func(t *testing.T) {
		// Each case starts from a valid report of its kind and breaks one thing, so it fails on the validation it names.
		exploreData := func() *ai.SaveReportArgs {
			return &ai.SaveReportArgs{DisplayName: "A", Kind: "explore_data", Explore: ptr("sales"), Measures: []string{"revenue"}, TimeRange: ptr("7D as of latest/D"), Schedule: &weekly}
		}
		aiReport := func() *ai.SaveReportArgs {
			return &ai.SaveReportArgs{DisplayName: "A", Kind: "ai", Schedule: &weekly}
		}
		canvasPDF := func() *ai.SaveReportArgs {
			return &ai.SaveReportArgs{DisplayName: "A", Kind: "canvas_pdf", Canvas: "overview", Schedule: &weekly}
		}
		for _, valid := range []func() *ai.SaveReportArgs{exploreData, aiReport, canvasPDF} {
			admin.reset()
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, valid())
			require.NoError(t, err, "the base of the cases must be valid")
		}

		cases := []struct {
			name   string
			args   *ai.SaveReportArgs
			change func(a *ai.SaveReportArgs)
			err    string
		}{
			{"unknown kind", aiReport(), func(a *ai.SaveReportArgs) { a.Kind = "query" }, `invalid kind "query"`},
			{"no schedule", aiReport(), func(a *ai.SaveReportArgs) { a.Schedule = nil }, "schedule.cron is required"},
			{"explore data without measures", exploreData(), func(a *ai.SaveReportArgs) { a.Measures = nil }, "needs at least one measure"},
			{"unknown export format", exploreData(), func(a *ai.SaveReportArgs) { a.Export = &ai.ReportExport{Format: "pdf"} }, `invalid export format "pdf"`},
			{"unknown canvas", canvasPDF(), func(a *ai.SaveReportArgs) { a.Canvas = "missing" }, `canvas "missing" not found`},
			{"filter on a view the canvas doesn't use", canvasPDF(), func(a *ai.SaveReportArgs) {
				a.CanvasFilters = []ai.CanvasFilters{{MetricsView: "other", Filters: []ai.DimensionFilter{{Dimension: "country", Values: []string{"x"}}}}}
			}, `metrics view "other" is not used by canvas "overview"`},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				admin.reset()
				c.change(c.args)
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, c.args)
				require.ErrorContains(t, err, c.err)
				require.Empty(t, admin.reports, "an invalid report reached the admin service")
			})
		}
	})
}

func TestEditReport(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	t.Run("explore data report", func(t *testing.T) {
		// As the report dialog saves it: dimensions pivoted into columns, an attached XLSX export, opening as the creator.
		dialog := func() *drivers.ReportOptions {
			return &drivers.ReportOptions{
				DisplayName:         "Weekly sales",
				RefreshCron:         "0 9 * * 1",
				RefreshTimeZone:     "Europe/Madrid",
				QueryName:           "MetricsViewAggregation",
				QueryArgsJSON:       `{"metricsView":"mv","dimensions":[{"name":"country"},{"name":"channel"}],"measures":[{"name":"revenue"}],"pivotOn":["channel"],"timeRange":{"isoDuration":"P7D","timeZone":"Europe/Madrid"},"sort":[{"name":"revenue","desc":true}]}`,
				ExportFormat:        runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
				ExportLimit:         500,
				ExportIncludeHeader: true,
				ExportAttach:        true,
				EmailRecipients:     []string{"user@example.com"},
				WebOpenState:        "CgVzYWxlcw==",
				WebOpenMode:         "creator",
				Explore:             "sales",
			}
		}

		version := admin.storeReport(userID, "weekly-sales", dialog())
		var list *ai.ListReportsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
		require.NoError(t, err)
		r := asJSON(t, findByName(t, list.Reports, "weekly-sales"))
		require.Equal(t, "explore_data", r["kind"])
		require.Equal(t, []any{"country", "channel"}, r["dimensions"])
		require.Equal(t, map[string]any{"format": "xlsx", "limit": 500.0, "include_header": true, "attach": true}, r["export"])

		// Changing the schedule keeps the rest exactly.
		res := saveReport(t, s, `{"name":"weekly-sales","version":"`+version+`","schedule":{"cron":"0 8 * * 1","time_zone":"Europe/Madrid"}}`)
		want := dialog()
		want.RefreshCron = "0 8 * * 1"
		require.Equal(t, want, admin.lastReport(t).opts)

		// Null resets the export options to their defaults.
		version = admin.storeReport(userID, "export-null", dialog())
		saveReport(t, s, `{"name":"export-null","version":"`+version+`","export":null}`)
		got := admin.lastReport(t).opts
		require.Equal(t, runtimev1.ExportFormat_EXPORT_FORMAT_CSV, got.ExportFormat)
		require.Zero(t, got.ExportLimit)
		require.False(t, got.ExportIncludeHeader)
		require.False(t, got.ExportAttach)

		// Changing the measures keeps the pivot and sort, and drops the stale dashboard state.
		saveReport(t, s, `{"name":"weekly-sales","version":"`+res.Version+`","measures":["revenue"],"dimensions":["channel"]}`)
		got = admin.lastReport(t).opts
		require.Empty(t, got.WebOpenState)
		q := parseAggregation(t, got.QueryArgsJSON)
		require.Equal(t, []string{"channel"}, q.PivotOn)
		require.Len(t, q.Sort, 1)
		require.Equal(t, "P7D", q.TimeRange.IsoDuration)
		require.True(t, got.ExportAttach)
		require.Equal(t, "creator", got.WebOpenMode)
	})

	t.Run("canvas PDF report", func(t *testing.T) {
		// As the report dialog saves it: per-view filters, a selected tab, PDF options, and other canvas state the tool doesn't set.
		dialog := func() *drivers.ReportOptions {
			return &drivers.ReportOptions{
				DisplayName:     "Overview PDF",
				RefreshCron:     "0 9 * * 1",
				ExportFormat:    runtimev1.ExportFormat_EXPORT_FORMAT_PDF,
				EmailRecipients: []string{"user@example.com"},
				WebOpenState:    "tr=7D&compare_tr=rill-PP&f.mv=country+IN+%28%27US%27%29&f.costs=country+IN+%28%27ES%27%29&tabs=main.costs&pdf_include_filters=false&pdf_all_tabs=false",
				WebOpenMode:     "recipient",
				Canvas:          "overview",
				MetricsViewFilters: map[string]*runtimev1.Expression{
					"mv":    expressionpb.And([]*runtimev1.Expression{expressionpb.In(expressionpb.Identifier("country"), []*runtimev1.Expression{expressionpb.String("US")})}),
					"costs": expressionpb.And([]*runtimev1.Expression{expressionpb.In(expressionpb.Identifier("country"), []*runtimev1.Expression{expressionpb.String("ES")})}),
				},
			}
		}

		version := admin.storeReport(userID, "overview-pdf", dialog())
		var list *ai.ListReportsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
		require.NoError(t, err)
		r := asJSON(t, findByName(t, list.Reports, "overview-pdf"))
		require.Equal(t, "canvas_pdf", r["kind"])
		require.Equal(t, false, r["pdf_include_filters"])
		require.Equal(t, false, r["pdf_all_tabs"])
		require.Equal(t, []any{map[string]any{"group": "main", "tab": "costs"}}, r["tabs"])
		require.Equal(t, []any{
			map[string]any{"metrics_view": "costs", "filters": []any{map[string]any{"dimension": "country", "values": []any{"ES"}}}},
			map[string]any{"metrics_view": "mv", "filters": []any{map[string]any{"dimension": "country", "values": []any{"US"}}}},
		}, r["canvas_filters"])

		// Changing only the schedule's cron keeps the canvas state and filters exactly.
		res := saveReport(t, s, `{"name":"overview-pdf","version":"`+version+`","schedule":{"cron":"0 7 * * 1"}}`)
		want := dialog()
		want.RefreshCron = "0 7 * * 1"
		requireSameReport(t, want, admin.lastReport(t).opts)

		// Changing the filters of one view keeps the other view's, the tab and the PDF options.
		res = saveReport(t, s, `{"name":"overview-pdf","version":"`+res.Version+`","canvas_filters":[{"metrics_view":"mv","filters":[{"dimension":"country","values":["FR"]}]}]}`)
		got := admin.lastReport(t).opts
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"country"},{"val":"ES"}]}}]}}`, got.MetricsViewFilters["costs"])
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"country"},{"val":"FR"}]}}]}}`, got.MetricsViewFilters["mv"])
		state, err := url.ParseQuery(got.WebOpenState)
		require.NoError(t, err)
		require.Equal(t, "country IN ('FR')", state.Get("f.mv"))
		require.Equal(t, "country IN ('ES')", state.Get("f.costs"))
		require.Equal(t, "main.costs", state.Get("tabs"))
		require.Equal(t, "false", state.Get("pdf_include_filters"))
		require.Equal(t, "false", state.Get("pdf_all_tabs"))
		require.Equal(t, "rill-PP", state.Get("compare_tr"))
		require.Equal(t, "7D", state.Get("tr"))

		// The PDF options and the tab can be changed, and null selects the first tabs.
		saveReport(t, s, `{"name":"overview-pdf","version":"`+res.Version+`","pdf_include_filters":true,"tabs":null}`)
		state, err = url.ParseQuery(admin.lastReport(t).opts.WebOpenState)
		require.NoError(t, err)
		require.Equal(t, "true", state.Get("pdf_include_filters"))
		require.Equal(t, "false", state.Get("pdf_all_tabs"))
		require.False(t, state.Has("tabs"))
		require.Equal(t, "country IN ('ES')", state.Get("f.costs"))
	})

	t.Run("AI report", func(t *testing.T) {
		// With resolver properties the tool's arguments don't express.
		dialog := func() *drivers.ReportOptions {
			return &drivers.ReportOptions{
				DisplayName:        "Weekly insights",
				RefreshCron:        "0 9 * * 1",
				RefreshTimeZone:    "Europe/Madrid",
				Resolver:           "ai",
				ResolverProperties: map[string]any{"prompt": "Summarize", "explore": "sales", "time_range": map[string]any{"expression": "7D as of latest/D"}, "time_zone": "Europe/Madrid", "measures": []any{"revenue"}},
				EmailRecipients:    []string{"user@example.com"},
				WebOpenMode:        "recipient",
			}
		}
		version := admin.storeReport(userID, "weekly-insights", dialog())
		res := saveReport(t, s, `{"name":"weekly-insights","version":"`+version+`","prompt":"Explain revenue by country"}`)
		want := dialog()
		want.ResolverProperties["prompt"] = "Explain revenue by country"
		require.Equal(t, want, admin.lastReport(t).opts)
		require.Equal(t, []ai.FieldChange{{Field: "resolver_properties.prompt", Old: "Summarize", New: "Explain revenue by country"}}, res.Changes)

		// Null clears the time range, and what depends on it.
		saveReport(t, s, `{"name":"weekly-insights","version":"`+res.Version+`","time_range":null}`)
		require.Equal(t, map[string]any{"prompt": "Explain revenue by country", "explore": "sales", "measures": []any{"revenue"}}, admin.lastReport(t).opts.ResolverProperties)
	})

	t.Run("an edit based on a stale version fails, telling to list again", func(t *testing.T) {
		opts := &drivers.ReportOptions{DisplayName: "A", RefreshCron: "0 9 * * 1", Resolver: "ai", EmailRecipients: []string{"user@example.com"}, WebOpenMode: "recipient"}
		version := admin.storeReport(userID, "stale", opts)
		admin.storeReport(userID, "stale", opts)
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, json.RawMessage(`{"name":"stale","version":"`+version+`","display_name":"B"}`))
		require.ErrorContains(t, err, "call list_reports again")
		require.Empty(t, admin.reports, "a stale edit reached the admin service")
	})

	t.Run("the kind of a report can't change", func(t *testing.T) {
		opts := &drivers.ReportOptions{DisplayName: "A", RefreshCron: "0 9 * * 1", Resolver: "ai", EmailRecipients: []string{"user@example.com"}, WebOpenMode: "recipient"}
		version := admin.storeReport(userID, "kind", opts)
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, json.RawMessage(`{"name":"kind","version":"`+version+`","kind":"canvas_pdf"}`))
		require.ErrorContains(t, err, "can't change")
	})
}

func TestReportComparison(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	s := newAlertReportSession(t, rt, instanceID, uuid.NewString(), map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	// An AI report compares with the chosen period, and keeps it when the time range changes.
	res := saveReport(t, s, `{"display_name":"A","kind":"ai","explore":"sales","time_range":"7D as of latest/D","comparison":"previous_month","schedule":{"cron":"0 9 * * 1"}}`)
	require.Equal(t, map[string]any{"expression": "7D as of latest/D offset -1M"}, admin.lastReport(t).opts.ResolverProperties["comparison_time_range"])

	res = saveReport(t, s, `{"name":"`+res.Name+`","version":"`+res.Version+`","time_range":"14D as of latest/D"}`)
	require.Equal(t, map[string]any{"expression": "14D as of latest/D offset -1M"}, admin.lastReport(t).opts.ResolverProperties["comparison_time_range"])

	var list *ai.ListReportsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
	require.NoError(t, err)
	listed := findByName(t, list.Reports, res.Name)
	require.Equal(t, "previous_month", listed["comparison"])
	require.Equal(t, true, listed["compare_to_previous_period"])

	// compare_to_previous_period alone keeps comparing with the previous period, and false removes the comparison.
	res = saveReport(t, s, `{"display_name":"B","kind":"ai","time_range":"7D as of latest/D","compare_to_previous_period":true,"schedule":{"cron":"0 9 * * 1"}}`)
	require.Equal(t, map[string]any{"expression": "7D as of latest/D offset -1P"}, admin.lastReport(t).opts.ResolverProperties["comparison_time_range"])
	saveReport(t, s, `{"name":"`+res.Name+`","version":"`+res.Version+`","compare_to_previous_period":false}`)
	require.NotContains(t, admin.lastReport(t).opts.ResolverProperties, "comparison_time_range")

	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, json.RawMessage(`{"display_name":"A","kind":"explore_data","explore":"sales","measures":["revenue"],"time_range":"7D as of latest/D","comparison":"previous_week","schedule":{"cron":"0 9 * * 1"}}`))
	require.ErrorContains(t, err, "comparison only applies to 'ai' reports")
}

// requireSameReport compares report options, with their filter expressions compared as protos.
func requireSameReport(t *testing.T, want, got *drivers.ReportOptions) {
	t.Helper()
	require.Len(t, got.MetricsViewFilters, len(want.MetricsViewFilters))
	for mv, expr := range want.MetricsViewFilters {
		require.True(t, proto.Equal(expr, got.MetricsViewFilters[mv]), "filter of %q: got %v", mv, got.MetricsViewFilters[mv])
	}
	w, g := *want, *got
	w.MetricsViewFilters, g.MetricsViewFilters = nil, nil
	require.Equal(t, w, g)
}

func saveReport(t *testing.T, s *ai.Session, args string) *ai.SaveReportResult {
	t.Helper()
	var res *ai.SaveReportResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, &res, json.RawMessage(args))
	require.NoError(t, err)
	return res
}
