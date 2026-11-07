package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var alertReportFiles = map[string]string{
	"rill.yaml": "",
	"m.sql":     "SELECT 'US' AS country, 'web' AS channel, 100 AS revenue, TIMESTAMP '2024-01-15 00:00:00' AS event_time UNION ALL SELECT 'ES', 'web', 50, TIMESTAMP '2024-01-10 00:00:00'",
	"mv.yaml": `
type: metrics_view
model: m
timeseries: event_time
dimensions:
- column: country
- column: channel
measures:
- name: revenue
  expression: SUM(revenue)
- name: orders
  expression: COUNT(*)
explore:
  skip: true
`,
	"sales.yaml": `
type: explore
metrics_view: mv
dimensions: '*'
measures: '*'
defaults:
  time_range: P14D
`,
	"costs.yaml": `
type: metrics_view
model: m
timeseries: event_time
dimensions:
- column: country
measures:
- name: cost
  expression: SUM(revenue) / 2
explore:
  skip: true
`,
	"overview.yaml": `
type: canvas
rows:
- items:
  - kpi_grid:
      metrics_view: mv
      measures:
      - revenue
  - kpi_grid:
      metrics_view: costs
      measures:
      - cost
`,
}

func TestAlertReportToolsCheckAccess(t *testing.T) {
	user := map[string]any{"email": "user@example.com"}
	service := map[string]any{"name": "my-service"}

	tests := []struct {
		name        string
		environment string
		userID      string
		attrs       map[string]any
		userAgent   string
		permissions []runtime.Permission
		want        bool
	}{
		{name: "user with AI in the web chat", userID: uuid.NewString(), attrs: user, userAgent: "rill", permissions: []runtime.Permission{runtime.UseAI}, want: true},
		{name: "no user", userAgent: "rill", permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "service token", userID: uuid.NewString(), attrs: service, userAgent: "rill", permissions: []runtime.Permission{runtime.UseAI}, want: false},
		{name: "missing AI permission", userID: uuid.NewString(), attrs: user, userAgent: "rill", want: false},
		{name: "non-prod deployment", environment: "dev", userID: uuid.NewString(), attrs: user, userAgent: "rill", permissions: []runtime.Permission{runtime.UseAI}, want: false},
		// AI reports and alerts carry a user when they run as one, but nobody is there to confirm what the AI saves.
		{name: "AI report run as a user", userID: uuid.NewString(), attrs: user, userAgent: "rill/report", permissions: []runtime.Permission{runtime.UseAI}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := tt.environment
			if environment == "" {
				environment = "prod"
			}
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{"rill.yaml": ""}, Environment: environment})
			s := newAlertReportSession(t, rt, instanceID, tt.userID, tt.attrs, tt.userAgent, tt.permissions...)
			ctx := ai.WithSession(t.Context(), s)

			for _, name := range []string{ai.SaveAlertName, ai.ListAlertsName, ai.SaveReportName, ai.ListReportsName} {
				tool, ok := s.Tool(name)
				require.True(t, ok, name)
				allowed, err := tool.CheckAccess(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.want, allowed, name)
			}
		})
	}
}

func TestSaveAlert(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	t.Run("creates a KPI alert owned by the user, as the alert dialog does", func(t *testing.T) {
		admin.reset()
		var res *ai.SaveAlertResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, &res, &ai.SaveAlertArgs{
			DisplayName: "Revenue below 30k",
			Explore:     "sales",
			Measure:     "revenue",
			Dimension:   ptr("country"),
			TimeRange:   ptr("7D as of latest/D"),
			Criteria:    []ai.AlertCriterion{{Operator: "lt", Value: 30000}},
			Where:       []ai.DimensionFilter{{Dimension: "channel", Values: []string{"web"}}},
		})
		require.NoError(t, err)
		require.Equal(t, "alert-1", res.Name)
		require.Equal(t, "https://ui.example.com/-/alerts/alert-1", res.URL)

		call := admin.lastAlert(t)
		require.Equal(t, userID, call.userID)
		require.Empty(t, call.name)
		require.Equal(t, "Revenue below 30k", call.opts.DisplayName)
		require.Equal(t, "mv", call.opts.MetricsViewName)
		require.Equal(t, "MetricsViewAggregation", call.opts.QueryName)
		require.Equal(t, "/explore/sales", call.opts.WebOpenPath)
		// Without a schedule, it's checked when the data refreshes, and like the dialog by default, it notifies on every check.
		require.Empty(t, call.opts.RefreshCron)
		require.True(t, call.opts.Renotify)
		require.Zero(t, call.opts.RenotifyAfterSeconds)
		require.Equal(t, []string{"user@example.com"}, call.opts.EmailRecipients)

		q := parseAggregation(t, call.opts.QueryArgsJSON)
		require.Equal(t, "mv", q.MetricsView)
		require.Equal(t, "revenue", q.Measures[0].Name)
		require.Equal(t, "country", q.Dimensions[0].Name)
		require.Equal(t, "7D as of latest/D", q.TimeRange.Expression)
		require.Nil(t, q.ComparisonTimeRange)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}}`, q.Where)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}`, q.Having)
	})

	t.Run("change criteria compare with the previous period", func(t *testing.T) {
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, &ai.SaveAlertArgs{
			DisplayName:          "Revenue drops 10%",
			Explore:              "sales",
			Measure:              "revenue",
			TimeRange:            ptr("7D as of latest/D"),
			Criteria:             []ai.AlertCriterion{{Type: "percent_change", Operator: "lte", Value: -10}},
			Schedule:             &ai.Schedule{Cron: "0 8 * * 1", TimeZone: "Europe/Madrid"},
			Recipients:           &ai.Recipients{Emails: []string{"boss@example.com"}, SlackChannels: []string{"#sales"}},
			Renotify:             "after_interval",
			RenotifyAfterSeconds: 86400,
		})
		require.NoError(t, err)

		call := admin.lastAlert(t)
		require.Equal(t, "0 8 * * 1", call.opts.RefreshCron)
		require.Equal(t, "Europe/Madrid", call.opts.RefreshTimeZone)
		require.True(t, call.opts.Renotify)
		require.Equal(t, uint32(86400), call.opts.RenotifyAfterSeconds)
		require.Equal(t, []string{"boss@example.com"}, call.opts.EmailRecipients)
		require.Equal(t, []string{"#sales"}, call.opts.SlackChannels)

		q := parseAggregation(t, call.opts.QueryArgsJSON)
		require.Equal(t, "7D as of latest/D offset -1P", q.ComparisonTimeRange.Expression)
		require.Equal(t, []string{"revenue", "revenue_delta", "revenue_delta_perc"}, []string{q.Measures[0].Name, q.Measures[1].Name, q.Measures[2].Name})
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LTE","exprs":[{"ident":"revenue_delta_perc"},{"val":-0.1}]}}]}}`, q.Having)
	})

	t.Run("a new alert's time range is evaluated in the schedule's time zone unless one is given", func(t *testing.T) {
		admin.reset()
		saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D","criteria":[{"type":"percent_change","operator":"lt","value":-10}],"schedule":{"cron":"0 8 * * 1","time_zone":"Europe/Madrid"}}`)
		q := parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
		require.Equal(t, "Europe/Madrid", q.TimeRange.TimeZone)
		require.Equal(t, "Europe/Madrid", q.ComparisonTimeRange.TimeZone)

		saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D","criteria":[{"operator":"lt","value":1}],"schedule":{"cron":"0 8 * * 1","time_zone":"Europe/Madrid"},"time_zone":"America/New_York"}`)
		q = parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
		require.Equal(t, "America/New_York", q.TimeRange.TimeZone)
	})

	t.Run("the saved queries run", func(t *testing.T) {
		admin.reset()
		for _, c := range []ai.AlertCriterion{{Operator: "lt", Value: 30000}, {Type: "absolute_change", Operator: "lt", Value: 0}, {Type: "percent_change", Operator: "lt", Value: -10}, {Type: "percent_of_total", Operator: "gt", Value: 50}} {
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, &ai.SaveAlertArgs{
				DisplayName: "A", Explore: "sales", Measure: "revenue", Dimension: ptr("country"), TimeRange: ptr("7D as of latest/D"),
				Criteria: []ai.AlertCriterion{c}, Where: []ai.DimensionFilter{{Dimension: "channel", Values: []string{"web"}}},
			})
			require.NoError(t, err)
			requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)
		}

		// Without a dimension, the query has the empty list of dimensions the dialog saves, and reads dimensions[0] back from when editing.
		saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D","criteria":[{"operator":"lt","value":1}]}`)
		var q map[string]any
		require.NoError(t, json.Unmarshal([]byte(admin.lastAlert(t).opts.QueryArgsJSON), &q))
		require.Equal(t, []any{}, q["dimensions"])
		requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)
	})

	t.Run("invalid arguments are rejected before saving", func(t *testing.T) {
		valid := func() *ai.SaveAlertArgs {
			return &ai.SaveAlertArgs{DisplayName: "A", Explore: "sales", Measure: "revenue", TimeRange: ptr("7D as of latest/D"), Criteria: []ai.AlertCriterion{{Operator: "lt", Value: 1}}}
		}
		cases := map[string]func(a *ai.SaveAlertArgs){
			"unknown explore":    func(a *ai.SaveAlertArgs) { a.Explore = "missing" },
			"unknown measure":    func(a *ai.SaveAlertArgs) { a.Measure = "profit" },
			"unknown filter":     func(a *ai.SaveAlertArgs) { a.Where = []ai.DimensionFilter{{Dimension: "city", Values: []string{"x"}}} },
			"no criteria":        func(a *ai.SaveAlertArgs) { a.Criteria = nil },
			"invalid time range": func(a *ai.SaveAlertArgs) { a.TimeRange = ptr("last week please") },
			"invalid cron":       func(a *ai.SaveAlertArgs) { a.Schedule = &ai.Schedule{Cron: "every monday"} },
		}
		for name, change := range cases {
			t.Run(name, func(t *testing.T) {
				admin.reset()
				args := valid()
				change(args)
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, args)
				require.Error(t, err)
				require.Empty(t, admin.alerts, "an invalid alert reached the admin service")
			})
		}
	})
}

func TestSaveAIAlert(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	// An AI alert as the alert dialog saves it (getAlertResolverOptions), for a condition on the explore "sales",
	// narrowed to a measure, a dimension and a filter, keeping the explore's default time range (P14D), in Madrid.
	dialog := func() *drivers.AlertOptions {
		return &drivers.AlertOptions{
			DisplayName:     "Sales drop unusually",
			MetricsViewName: "mv",
			Resolver:        "ai",
			ResolverProperties: map[string]any{
				"prompt":     "Sales drop unusually compared to recent weeks",
				"explore":    "sales",
				"measures":   []any{"revenue"},
				"dimensions": []any{"country"},
				"where":      map[string]any{"cond": map[string]any{"op": "and", "exprs": []any{map[string]any{"cond": map[string]any{"op": "in", "exprs": []any{map[string]any{"name": "channel"}, map[string]any{"val": "web"}}}}}}},
				"time_range": map[string]any{"expression": "P14D"},
				"time_zone":  "Europe/Madrid",
			},
			Renotify:        true,
			EmailRecipients: []string{"user@example.com"},
			WebOpenPath:     "/explore/sales",
		}
	}

	t.Run("creates the AI alert the dialog creates for the same inputs, with the explore's default time range written explicitly", func(t *testing.T) {
		admin.reset()
		saveAlert(t, s, `{"display_name":"Sales drop unusually","condition":"Sales drop unusually compared to recent weeks","explore":"sales","measure":"revenue","dimension":"country","where":[{"dimension":"channel","values":["web"]}],"time_zone":"Europe/Madrid"}`)
		requireSameAlert(t, dialog(), admin.lastAlert(t).opts)
	})

	t.Run("lists the AI alert in the arguments of save_alert", func(t *testing.T) {
		version := admin.storeAlert(userID, "listed-ai-alert", dialog())
		var res *ai.ListAlertsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &res, &ai.ListAlertsArgs{})
		require.NoError(t, err)
		a := asJSON(t, findByName(t, res.Alerts, "listed-ai-alert"))
		require.Equal(t, version, a["version"])
		require.Equal(t, "ai", a["condition_type"])
		require.Equal(t, "Sales drop unusually compared to recent weeks", a["condition"])
		require.Equal(t, "sales", a["explore"])
		require.Equal(t, "revenue", a["measure"])
		require.Equal(t, "country", a["dimension"])
		require.Equal(t, []any{map[string]any{"dimension": "channel", "values": []any{"web"}}}, a["where"])
		require.Equal(t, "P14D", a["time_range"])
		require.Equal(t, "Europe/Madrid", a["time_zone"])
	})

	t.Run("an edit of only the condition keeps the rest, and drops the stale dashboard state", func(t *testing.T) {
		stored := dialog()
		stored.ResolverProperties["comparison_time_range"] = map[string]any{"expression": "P14D offset -1P"} // Set through the API
		stored.WebOpenState = "CgVzYWxlcw=="
		version := admin.storeAlert(userID, "ai-1", stored)
		res := saveAlert(t, s, `{"name":"ai-1","version":"`+version+`","condition":"Revenue falls more than usual"}`)

		want := dialog()
		want.ResolverProperties["comparison_time_range"] = map[string]any{"expression": "P14D offset -1P"}
		want.ResolverProperties["prompt"] = "Revenue falls more than usual"
		requireSameAlert(t, want, admin.lastAlert(t).opts)
		require.Contains(t, changedFields(res.Changes), "resolver_properties.prompt")
	})

	t.Run("a formula alert becomes an AI alert keeping its data context and dropping its query", func(t *testing.T) {
		kpi := &drivers.AlertOptions{
			DisplayName:     "Revenue below 30k",
			MetricsViewName: "mv",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJSON:   `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[{"name":"country"}],"where":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}},"timeRange":{"isoDuration":"P14D","timeZone":"Europe/Madrid"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}}`,
			Renotify:        true,
			EmailRecipients: []string{"user@example.com"},
			WebOpenPath:     "/explore/sales",
			WebOpenState:    "CgVzYWxlcw==",
		}
		version := admin.storeAlert(userID, "kpi-to-ai", kpi)
		saveAlert(t, s, `{"name":"kpi-to-ai","version":"`+version+`","display_name":"Sales drop unusually","condition":"Sales drop unusually compared to recent weeks"}`)
		requireSameAlert(t, dialog(), admin.lastAlert(t).opts)
	})

	t.Run("a formula alert's complete-period range becomes the preset it stands for, as the dialog writes it", func(t *testing.T) {
		kpi := func(timeRange string) *drivers.AlertOptions {
			return &drivers.AlertOptions{
				DisplayName:     "Revenue below 30k",
				MetricsViewName: "mv",
				QueryName:       "MetricsViewAggregation",
				QueryArgsJSON:   `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[],"timeRange":` + timeRange + `,"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}}`,
				Renotify:        true,
				EmailRecipients: []string{"user@example.com"},
				WebOpenPath:     "/explore/sales",
			}
		}
		version := admin.storeAlert(userID, "previous-month", kpi(`{"isoDuration":"P1M","roundToGrain":"TIME_GRAIN_MONTH","timeZone":"Europe/Madrid"}`))
		saveAlert(t, s, `{"name":"previous-month","version":"`+version+`","condition":"Revenue looks off"}`)
		props := admin.lastAlert(t).opts.ResolverProperties
		require.Equal(t, map[string]any{"expression": "rill-PMC"}, props["time_range"])
		require.Equal(t, "Europe/Madrid", props["time_zone"])

		// A range the dialog's presets can't express is rejected, not replaced by another one.
		version = admin.storeAlert(userID, "offset", kpi(`{"isoDuration":"P7D","isoOffset":"P7D","timeZone":"Europe/Madrid"}`))
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(`{"name":"offset","version":"`+version+`","condition":"Revenue looks off"}`))
		require.ErrorContains(t, err, "time_range")
		require.Empty(t, admin.alerts)

		// Unless the call sets the time range itself.
		saveAlert(t, s, `{"name":"offset","version":"`+version+`","condition":"Revenue looks off","time_range":"7D as of latest/D"}`)
		require.Equal(t, map[string]any{"expression": "7D as of latest/D"}, admin.lastAlert(t).opts.ResolverProperties["time_range"])
		require.Equal(t, "Europe/Madrid", admin.lastAlert(t).opts.ResolverProperties["time_zone"], "the alert's time zone is kept with the new range")
	})

	t.Run("an AI alert becomes a formula alert keeping its data context and dropping its AI properties", func(t *testing.T) {
		version := admin.storeAlert(userID, "ai-to-kpi", dialog())
		saveAlert(t, s, `{"name":"ai-to-kpi","version":"`+version+`","display_name":"Revenue below 30k","criteria":[{"operator":"lt","value":30000}]}`)

		got := admin.lastAlert(t).opts
		require.Empty(t, got.Resolver)
		require.Nil(t, got.ResolverProperties)
		require.Equal(t, "MetricsViewAggregation", got.QueryName)
		q := parseAggregation(t, got.QueryArgsJSON)
		require.Equal(t, "revenue", q.Measures[0].Name)
		require.Equal(t, "country", q.Dimensions[0].Name)
		require.Equal(t, "P14D", q.TimeRange.Expression)
		require.Equal(t, "Europe/Madrid", q.TimeRange.TimeZone)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}}`, q.Where)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}`, q.Having)
		requireQueryRuns(t, rt, instanceID, got.QueryArgsJSON)
	})

	t.Run("an AI alert with a fixed time range becomes a formula alert over the same range", func(t *testing.T) {
		stored := dialog()
		stored.ResolverProperties["time_range"] = map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}
		version := admin.storeAlert(userID, "fixed-range", stored)
		saveAlert(t, s, `{"name":"fixed-range","version":"`+version+`","criteria":[{"operator":"lt","value":30000}]}`)

		q := parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
		require.Equal(t, "2024-01-01T00:00:00Z", q.TimeRange.Start.AsTime().Format(time.RFC3339))
		require.Equal(t, "2024-02-01T00:00:00Z", q.TimeRange.End.AsTime().Format(time.RFC3339))
		require.Equal(t, "Europe/Madrid", q.TimeRange.TimeZone)
		requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)
	})

	t.Run("a condition and criteria can't go together", func(t *testing.T) {
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(`{"display_name":"A","condition":"x","explore":"sales","measure":"revenue","time_range":"7D as of latest/D","criteria":[{"operator":"lt","value":1}]}`))
		require.ErrorContains(t, err, "either")
	})
}

func TestEditAlert(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	// A KPI alert as the alert dialog saves it: an ISO duration time range, notifying on every check (snooze off), with the dashboard state it opens with.
	dialog := func() *drivers.AlertOptions {
		return &drivers.AlertOptions{
			DisplayName:     "Revenue below 30k",
			MetricsViewName: "mv",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJSON:   `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[{"name":"country"}],"where":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}},"timeRange":{"isoDuration":"P7D","timeZone":"Europe/Madrid"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":30000}]}}]}}}`,
			Renotify:        true,
			EmailRecipients: []string{"user@example.com"},
			SlackChannels:   []string{"#sales"},
			WebOpenPath:     "/explore/sales",
			WebOpenState:    "CgVzYWxlcw==",
		}
	}

	t.Run("lists the alert in the arguments of save_alert", func(t *testing.T) {
		version := admin.storeAlert(userID, "revenue-below-30k", dialog())
		admin.storeAlert(uuid.NewString(), "theirs", dialog())

		var res *ai.ListAlertsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &res, &ai.ListAlertsArgs{})
		require.NoError(t, err)
		require.Len(t, res.Alerts, 1, "only the user's own alerts are listed")
		a := asJSON(t, res.Alerts[0])
		require.Equal(t, "revenue-below-30k", a["name"])
		require.Equal(t, version, a["version"])
		require.Equal(t, "every_check", a["renotify"])
		require.Equal(t, "sales", a["explore"])
		require.Equal(t, "revenue", a["measure"])
		require.Equal(t, "country", a["dimension"])
		require.Equal(t, []any{map[string]any{"type": "value", "operator": "lt", "value": 30000.0}}, a["criteria"])
		require.Equal(t, []any{map[string]any{"dimension": "channel", "values": []any{"web"}}}, a["where"])
		require.Equal(t, map[string]any{"isoDuration": "P7D", "timeZone": "Europe/Madrid"}, a["time_range"])
	})

	t.Run("an edit that doesn't touch the data keeps everything else, the dashboard state included", func(t *testing.T) {
		version := admin.storeAlert(userID, "a1", dialog())
		res := saveAlert(t, s, `{"name":"a1","version":"`+version+`","display_name":"Revenue alert"}`)

		want := dialog()
		want.DisplayName = "Revenue alert"
		require.Equal(t, want, admin.lastAlert(t).opts)
		require.Equal(t, []ai.FieldChange{{Field: "display_name", Old: "Revenue below 30k", New: "Revenue alert"}}, res.Changes)
	})

	t.Run("an edit of the threshold keeps the time range, filters, notifications and recipients, and drops the stale dashboard state", func(t *testing.T) {
		version := admin.storeAlert(userID, "a2", dialog())
		res := saveAlert(t, s, `{"name":"a2","version":"`+version+`","criteria":[{"operator":"lt","value":25000}]}`)

		got := admin.lastAlert(t).opts
		want := dialog()
		want.QueryArgsJSON = got.QueryArgsJSON
		want.WebOpenState = ""
		require.Equal(t, want, got)

		q := parseAggregation(t, got.QueryArgsJSON)
		require.Equal(t, "P7D", q.TimeRange.IsoDuration)
		require.Equal(t, "Europe/Madrid", q.TimeRange.TimeZone)
		require.Equal(t, "country", q.Dimensions[0].Name)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}}]}}`, q.Where)
		requireExpr(t, `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":25000}]}}]}}`, q.Having)

		fields := map[string]bool{}
		for _, c := range res.Changes {
			fields[c.Field] = true
		}
		require.Equal(t, map[string]bool{"query.having": true, "dashboard_state": true}, fields)
	})

	t.Run("null clears an optional setting", func(t *testing.T) {
		opts := dialog()
		opts.RefreshCron, opts.RefreshTimeZone = "0 8 * * 1", "Europe/Madrid"
		version := admin.storeAlert(userID, "a4", opts)
		saveAlert(t, s, `{"name":"a4","version":"`+version+`","dimension":null,"where":null,"schedule":null}`)

		got := admin.lastAlert(t).opts
		q := parseAggregation(t, got.QueryArgsJSON)
		require.Empty(t, q.Dimensions)
		require.Nil(t, q.Where)
		require.Empty(t, got.RefreshCron)
		require.Equal(t, "P7D", q.TimeRange.IsoDuration)
	})

	t.Run("null recipients clear every kind of recipient, so the alert is left with none and isn't saved", func(t *testing.T) {
		opts := dialog()
		opts.SlackWebhooks = []string{"https://hooks.slack.com/x"}
		version := admin.storeAlert(userID, "a7", opts)
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(`{"name":"a7","version":"`+version+`","recipients":null}`))
		require.ErrorContains(t, err, "at least one recipient")
		require.Empty(t, admin.alerts)
	})

	t.Run("an edit based on a stale version fails, telling to list again", func(t *testing.T) {
		version := admin.storeAlert(userID, "a6", dialog())
		admin.storeAlert(userID, "a6", dialog()) // Changed since, e.g. in the dialog
		admin.reset()
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(`{"name":"a6","version":"`+version+`","display_name":"Stale"}`))
		require.ErrorContains(t, err, "call list_alerts again")
		require.Empty(t, admin.alerts, "a stale edit reached the admin service")

		_, err = s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(`{"name":"a6","display_name":"No version"}`))
		require.ErrorContains(t, err, "version is required")
	})
}

func TestAlertComparison(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)
	change := `"criteria":[{"type":"percent_change","operator":"lt","value":-10}]`

	// The comparison time ranges the alert dialog saves for each comparison option (mapSelectedComparisonTimeRangeToV1TimeRange in the UI):
	// a Rill time expression is offset by the option's rillTimeOffset; an ISO duration, or a legacy expression, gets the option's name as its ISO offset,
	// except for the previous period, whose offset is the duration itself.
	cases := []struct {
		comparison string
		expression string
		isoOffset  string
	}{
		{comparison: "", expression: "7D as of latest/D offset -1P", isoOffset: "P7D"},
		{comparison: "previous_period", expression: "7D as of latest/D offset -1P", isoOffset: "P7D"},
		{comparison: "previous_day", expression: "7D as of latest/D offset -1D", isoOffset: "rill-PD"},
		{comparison: "previous_week", expression: "7D as of latest/D offset -1W", isoOffset: "rill-PW"},
		{comparison: "previous_month", expression: "7D as of latest/D offset -1M", isoOffset: "rill-PM"},
		{comparison: "previous_quarter", expression: "7D as of latest/D offset -1Q", isoOffset: "rill-PQ"},
		{comparison: "previous_year", expression: "7D as of latest/D offset -1Y", isoOffset: "rill-PY"},
	}
	for _, c := range cases {
		t.Run("comparison "+c.comparison, func(t *testing.T) {
			arg := ""
			if c.comparison != "" {
				arg = `,"comparison":"` + c.comparison + `"`
			}
			want := c.comparison
			if want == "" {
				want = "previous_period"
			}

			// A Rill time expression.
			res := saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D",`+change+arg+`}`)
			q := parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
			requireProtoJSON(t, `{"expression":"`+c.expression+`"}`, q.ComparisonTimeRange)
			requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)

			var list *ai.ListAlertsResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &list, &ai.ListAlertsArgs{})
			require.NoError(t, err)
			require.Equal(t, want, findByName(t, list.Alerts, res.Name)["comparison"])

			// A legacy ISO duration expression, which the dialog compares as an ISO duration.
			saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"P7D",`+change+arg+`}`)
			q = parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
			requireProtoJSON(t, `{"isoDuration":"P7D","isoOffset":"`+c.isoOffset+`"}`, q.ComparisonTimeRange)
			requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)

			// An ISO duration time range, as the dialog saves its presets.
			opts := &drivers.AlertOptions{
				DisplayName: "A", MetricsViewName: "mv", QueryName: "MetricsViewAggregation", Renotify: true, EmailRecipients: []string{"user@example.com"}, WebOpenPath: "/explore/sales",
				QueryArgsJSON: `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[],"timeRange":{"isoDuration":"P7D"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":1}]}}]}}}`,
			}
			version := admin.storeAlert(userID, "iso-"+want, opts)
			saveAlert(t, s, `{"name":"iso-`+want+`","version":"`+version+`",`+change+arg+`}`)
			q = parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
			requireProtoJSON(t, `{"isoDuration":"P7D","isoOffset":"`+c.isoOffset+`"}`, q.ComparisonTimeRange)
			requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)
		})
	}

	t.Run("an edit keeps the comparison, also when the time range changes, unless it sends one", func(t *testing.T) {
		res := saveAlert(t, s, `{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D",`+change+`,"comparison":"previous_week"}`)
		res = saveAlert(t, s, `{"name":"`+res.Name+`","version":"`+res.Version+`","criteria":[{"type":"absolute_change","operator":"lt","value":0}]}`)
		require.Equal(t, "7D as of latest/D offset -1W", parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).ComparisonTimeRange.Expression)

		res = saveAlert(t, s, `{"name":"`+res.Name+`","version":"`+res.Version+`","time_range":"14D as of latest/D"}`)
		require.Equal(t, "14D as of latest/D offset -1W", parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).ComparisonTimeRange.Expression)

		saveAlert(t, s, `{"name":"`+res.Name+`","version":"`+res.Version+`","comparison":"previous_year"}`)
		require.Equal(t, "14D as of latest/D offset -1Y", parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).ComparisonTimeRange.Expression)

		// The dialog's previous-month comparison of an ISO duration is kept when the time range changes to another legacy duration.
		opts := &drivers.AlertOptions{
			DisplayName: "A", MetricsViewName: "mv", QueryName: "MetricsViewAggregation", Renotify: true, EmailRecipients: []string{"user@example.com"}, WebOpenPath: "/explore/sales",
			QueryArgsJSON: `{"metricsView":"mv","measures":[{"name":"revenue"},{"name":"revenue_delta_perc","comparisonRatio":{"measure":"revenue"}}],"dimensions":[],"timeRange":{"isoDuration":"P7D"},"comparisonTimeRange":{"isoDuration":"P7D","isoOffset":"rill-PM"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue_delta_perc"},{"val":-0.1}]}}]}}}`,
		}
		version := admin.storeAlert(userID, "dialog-month", opts)
		saveAlert(t, s, `{"name":"dialog-month","version":"`+version+`","time_range":"P14D"}`)
		requireProtoJSON(t, `{"isoDuration":"P14D","isoOffset":"rill-PM"}`, parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).ComparisonTimeRange)
	})

	t.Run("a comparison is rejected where it doesn't apply", func(t *testing.T) {
		admin.reset()
		for _, args := range []string{
			`{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D","criteria":[{"operator":"lt","value":1}],"comparison":"previous_week"}`,
			`{"display_name":"A","explore":"sales","measure":"revenue","time_range":"7D as of latest/D",` + change + `,"comparison":"last_week"}`,
			`{"display_name":"A","explore":"sales","condition":"revenue drops unusually","comparison":"previous_week"}`,
		} {
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, json.RawMessage(args))
			require.Error(t, err, args)
		}
		require.Empty(t, admin.alerts)
	})
}

func TestEditKeepsFiltersTheToolsCantExpress(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	userID := uuid.NewString()
	s := newAlertReportSession(t, rt, instanceID, userID, map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	// The dashboards' filters: a dimension filter and a measure filter, which is a subquery on a dimension (createSubQueryExpression in the UI).
	measureFilter := `{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"country"},{"subquery":{"dimension":"country","measures":["revenue"],"having":{"cond":{"op":"OPERATION_GT","exprs":[{"ident":"revenue"},{"val":10}]}}}}]}}`
	dialogWhere := `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"channel"},{"val":"web"}]}},` + measureFilter + `]}}`
	editedWhere := `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_IN","exprs":[{"ident":"country"},{"val":"US"}]}},` + measureFilter + `]}}`
	onlyMeasureFilter := `{"cond":{"op":"OPERATION_AND","exprs":[` + measureFilter + `]}}`
	listed := []any{map[string]any{"dimension": "channel", "values": []any{"web"}}}
	whereProps := func() map[string]any {
		e := &runtimev1.Expression{}
		require.NoError(t, protojson.Unmarshal([]byte(dialogWhere), e))
		m, err := metricsview.NewExpressionFromProto(e).AsMap()
		require.NoError(t, err)
		return m
	}
	propsWhere := func(props map[string]any) *runtimev1.Expression {
		data, err := json.Marshal(props["where"])
		require.NoError(t, err)
		e := &metricsview.Expression{}
		require.NoError(t, json.Unmarshal(data, e))
		return metricsview.ExpressionToProto(e)
	}

	t.Run("formula alert", func(t *testing.T) {
		version := admin.storeAlert(userID, "measure-filter", &drivers.AlertOptions{
			DisplayName: "A", MetricsViewName: "mv", QueryName: "MetricsViewAggregation", Renotify: true, EmailRecipients: []string{"user@example.com"}, WebOpenPath: "/explore/sales",
			QueryArgsJSON: `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[],"where":` + dialogWhere + `,"timeRange":{"isoDuration":"P7D"},"having":{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LT","exprs":[{"ident":"revenue"},{"val":1}]}}]}}}`,
		})

		var list *ai.ListAlertsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &list, &ai.ListAlertsArgs{})
		require.NoError(t, err)
		a := asJSON(t, findByName(t, list.Alerts, "measure-filter"))
		require.Equal(t, listed, a["where"])
		require.Equal(t, []any{asJSONValue(t, json.RawMessage(measureFilter))}, a["where_saved_otherwise"])

		res := saveAlert(t, s, `{"name":"measure-filter","version":"`+version+`","where":[{"dimension":"country","values":["US"]}]}`)
		requireExpr(t, editedWhere, parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).Where)
		requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)

		saveAlert(t, s, `{"name":"measure-filter","version":"`+res.Version+`","where":null}`)
		requireExpr(t, onlyMeasureFilter, parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON).Where)
	})

	t.Run("AI alert", func(t *testing.T) {
		version := admin.storeAlert(userID, "ai-measure-filter", &drivers.AlertOptions{
			DisplayName: "A", MetricsViewName: "mv", Renotify: true, EmailRecipients: []string{"user@example.com"}, WebOpenPath: "/explore/sales",
			Resolver: "ai", ResolverProperties: map[string]any{"prompt": "revenue drops", "explore": "sales", "where": whereProps()},
		})

		var list *ai.ListAlertsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &list, &ai.ListAlertsArgs{})
		require.NoError(t, err)
		a := asJSON(t, findByName(t, list.Alerts, "ai-measure-filter"))
		require.Equal(t, listed, a["where"])
		require.Len(t, a["where_saved_otherwise"], 1)

		saveAlert(t, s, `{"name":"ai-measure-filter","version":"`+version+`","where":[{"dimension":"country","values":["US"]}]}`)
		requireExpr(t, editedWhere, propsWhere(admin.lastAlert(t).opts.ResolverProperties))
	})

	t.Run("explore data report", func(t *testing.T) {
		version := admin.storeReport(userID, "measure-filter", &drivers.ReportOptions{
			DisplayName: "A", RefreshCron: "0 9 * * 1", Explore: "sales", QueryName: "MetricsViewAggregation", ExportFormat: runtimev1.ExportFormat_EXPORT_FORMAT_CSV, EmailRecipients: []string{"user@example.com"}, WebOpenMode: "recipient",
			QueryArgsJSON: `{"metricsView":"mv","measures":[{"name":"revenue"}],"dimensions":[{"name":"country"}],"where":` + dialogWhere + `,"timeRange":{"isoDuration":"P7D"}}`,
		})

		var list *ai.ListReportsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListReportsName, &list, &ai.ListReportsArgs{})
		require.NoError(t, err)
		r := asJSON(t, findByName(t, list.Reports, "measure-filter"))
		require.Equal(t, listed, r["where"])
		require.Len(t, r["where_saved_otherwise"], 1)

		saveReport(t, s, `{"name":"measure-filter","version":"`+version+`","where":[{"dimension":"country","values":["US"]}]}`)
		requireExpr(t, editedWhere, parseAggregation(t, admin.lastReport(t).opts.QueryArgsJSON).Where)
		requireQueryRuns(t, rt, instanceID, admin.lastReport(t).opts.QueryArgsJSON)
	})

	t.Run("AI report", func(t *testing.T) {
		version := admin.storeReport(userID, "ai-measure-filter", &drivers.ReportOptions{
			DisplayName: "A", RefreshCron: "0 9 * * 1", EmailRecipients: []string{"user@example.com"}, WebOpenMode: "recipient",
			Resolver: "ai", ResolverProperties: map[string]any{"prompt": "Summarize", "explore": "sales", "where": whereProps()},
		})
		saveReport(t, s, `{"name":"ai-measure-filter","version":"`+version+`","where":[{"dimension":"country","values":["US"]}]}`)
		requireExpr(t, editedWhere, propsWhere(admin.lastReport(t).opts.ResolverProperties))
	})
}

func TestSaveAlertReportExploreFromMetricsView(t *testing.T) {
	// The analyst's tools list metrics views, not explores, so the model may pass a metrics view's name as the explore.
	// mv has two explores, costs one the user can access and one they can't, and items none.
	files := maps.Clone(alertReportFiles)
	files["sales_eu.yaml"] = "type: explore\nmetrics_view: mv\n"
	files["costs_explore.yaml"] = "type: explore\nmetrics_view: costs\n"
	files["costs_hidden.yaml"] = "type: explore\nmetrics_view: costs\nsecurity:\n  access: false\n"
	files["items.yaml"] = "type: metrics_view\nmodel: m\ndimensions:\n- column: country\nmeasures:\n- name: revenue\n  expression: SUM(revenue)\nexplore:\n  skip: true\n"
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: files, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 12, 0, 0)
	s := newAlertReportSession(t, rt, instanceID, uuid.NewString(), map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)
	weekly := &ai.Schedule{Cron: "0 9 * * 1"}

	// Each tool saves the explore it resolves in its own place.
	tools := []struct {
		name    string
		save    func(explore string) error
		explore func(t *testing.T) string
	}{
		{
			name: "formula alert",
			save: func(explore string) error {
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, &ai.SaveAlertArgs{DisplayName: "A", Explore: explore, Measure: "cost", TimeRange: ptr("7D as of latest/D"), Criteria: []ai.AlertCriterion{{Operator: "lt", Value: 1}}})
				return err
			},
			explore: func(t *testing.T) string { return strings.TrimPrefix(admin.lastAlert(t).opts.WebOpenPath, "/explore/") },
		},
		{
			name: "AI alert",
			save: func(explore string) error {
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, &ai.SaveAlertArgs{DisplayName: "A", Explore: explore, Condition: "costs grow unusually"})
				return err
			},
			explore: func(t *testing.T) string {
				opts := admin.lastAlert(t).opts
				require.Equal(t, opts.ResolverProperties["explore"], strings.TrimPrefix(opts.WebOpenPath, "/explore/"))
				return opts.ResolverProperties["explore"].(string)
			},
		},
		{
			name: "explore data report",
			save: func(explore string) error {
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, &ai.SaveReportArgs{DisplayName: "A", Kind: "explore_data", Explore: &explore, Measures: []string{"cost"}, TimeRange: ptr("7D as of latest/D"), Schedule: weekly})
				return err
			},
			explore: func(t *testing.T) string { return admin.lastReport(t).opts.Explore },
		},
		{
			name: "AI report",
			save: func(explore string) error {
				_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveReportName, nil, &ai.SaveReportArgs{DisplayName: "A", Kind: "ai", Explore: &explore, Schedule: weekly})
				return err
			},
			explore: func(t *testing.T) string { return admin.lastReport(t).opts.ResolverProperties["explore"].(string) },
		},
	}
	cases := []struct {
		name    string
		explore string
		want    string
		err     string
	}{
		{name: "an explore's name", explore: "costs_explore", want: "costs_explore"},
		{name: "a metrics view with one explore the user can access", explore: "costs", want: "costs_explore"},
		{name: "a metrics view with several explores", explore: "mv", err: `"mv" is a metrics view with several explores: pass one of them as the explore: sales, sales_eu`},
		{name: "a metrics view without explores", explore: "items", err: `explore "items" not found, and no explore uses a metrics view of that name: pass one of the explores: costs_explore, sales, sales_eu`},
		{name: "an unknown name", explore: "missing", err: "pass one of the explores: costs_explore, sales, sales_eu:"},
		{name: "an explore the user can't access", explore: "costs_hidden", err: `explore "costs_hidden" not found`},
	}
	for _, tool := range tools {
		for _, c := range cases {
			t.Run(tool.name+"/"+c.name, func(t *testing.T) {
				admin.reset()
				err := tool.save(c.explore)
				if c.err != "" {
					require.ErrorContains(t, err, c.err)
					require.Empty(t, admin.alerts)
					require.Empty(t, admin.reports)
					return
				}
				require.NoError(t, err)
				require.Equal(t, c.want, tool.explore(t))
			})
		}
	}
}

func TestDimensionFilterContains(t *testing.T) {
	admin := &fakeAlertReportAdmin{}
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: alertReportFiles, Environment: "prod", AdminService: admin})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)
	s := newAlertReportSession(t, rt, instanceID, uuid.NewString(), map[string]any{"email": "user@example.com"}, "rill", runtime.UseAI, runtime.ReadObjects, runtime.ReadMetrics)

	// The expression and filter param of the dashboard's Contains mode (DimensionFilterManager and convertExpressionToFilterParam in the UI).
	cases := []struct {
		name   string
		filter ai.DimensionFilter
		expr   string
		param  string
	}{
		{
			name:   "contains",
			filter: ai.DimensionFilter{Dimension: "country", Contains: "U"},
			expr:   `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_LIKE","exprs":[{"ident":"country"},{"val":"%U%"}]}}]}}`,
			param:  `country LIKE '%U%'`,
		},
		{
			name:   "doesn't contain",
			filter: ai.DimensionFilter{Dimension: "country", Contains: "O'H", Exclude: true},
			expr:   `{"cond":{"op":"OPERATION_AND","exprs":[{"cond":{"op":"OPERATION_NLIKE","exprs":[{"ident":"country"},{"val":"%O'H%"}]}}]}}`,
			param:  `country NOT LIKE '%O\'H%'`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := saveAlert(t, s, argsJSON(t, &ai.SaveAlertArgs{DisplayName: "A", Explore: "sales", Measure: "revenue", TimeRange: ptr("7D as of latest/D"), Criteria: []ai.AlertCriterion{{Operator: "lt", Value: 1}}, Where: []ai.DimensionFilter{c.filter}}))
			q := parseAggregation(t, admin.lastAlert(t).opts.QueryArgsJSON)
			requireExpr(t, c.expr, q.Where)
			requireQueryRuns(t, rt, instanceID, admin.lastAlert(t).opts.QueryArgsJSON)

			var list *ai.ListAlertsResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &list, &ai.ListAlertsArgs{})
			require.NoError(t, err)
			require.Equal(t, asJSONValue(t, []ai.DimensionFilter{c.filter}), asJSONValue(t, findByName(t, list.Alerts, res.Name)["where"]))

			// An AI alert writes the same condition in its properties, and reads it back.
			res = saveAlert(t, s, argsJSON(t, &ai.SaveAlertArgs{DisplayName: "A", Explore: "sales", Condition: "revenue drops", Where: []ai.DimensionFilter{c.filter}}))
			_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ListAlertsName, &list, &ai.ListAlertsArgs{})
			require.NoError(t, err)
			require.Equal(t, asJSONValue(t, []ai.DimensionFilter{c.filter}), asJSONValue(t, findByName(t, list.Alerts, res.Name)["where"]))

			// A canvas PDF report writes it in its security rules and in the canvas state.
			saveReport(t, s, argsJSON(t, &ai.SaveReportArgs{DisplayName: "A", Kind: "canvas_pdf", Canvas: "overview", CanvasFilters: []ai.CanvasFilters{{MetricsView: "mv", Filters: []ai.DimensionFilter{c.filter}}}, Schedule: &ai.Schedule{Cron: "0 9 * * 1"}}))
			requireExpr(t, c.expr, admin.lastReport(t).opts.MetricsViewFilters["mv"])
			state, err := url.ParseQuery(admin.lastReport(t).opts.WebOpenState)
			require.NoError(t, err)
			require.Equal(t, c.param, state.Get("f.mv"))
		})
	}

	t.Run("a filter takes either values or a contains text", func(t *testing.T) {
		admin.reset()
		for _, f := range []ai.DimensionFilter{{Dimension: "country", Values: []string{"US"}, Contains: "U"}, {Dimension: "country"}} {
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, nil, &ai.SaveAlertArgs{DisplayName: "A", Explore: "sales", Measure: "revenue", TimeRange: ptr("7D as of latest/D"), Criteria: []ai.AlertCriterion{{Operator: "lt", Value: 1}}, Where: []ai.DimensionFilter{f}})
			require.ErrorContains(t, err, "either at least one value or a contains text")
		}
		require.Empty(t, admin.alerts)
	})
}

func TestSaveArgsKeepNullsWhenForwarded(t *testing.T) {
	// MCP clients' arguments are unmarshaled into the tool's arguments and marshaled again to call it, so a null must survive the trip.
	for _, args := range []any{&ai.SaveAlertArgs{}, &ai.SaveReportArgs{}} {
		in := `{"name":"x","version":"v1","where":null,"time_range":null}`
		require.NoError(t, json.Unmarshal([]byte(in), args))
		out, err := json.Marshal(args)
		require.NoError(t, err)
		require.JSONEq(t, in, string(out))
	}
}

// argsJSON marshals tool arguments, for the helpers that call a tool with them as JSON.
func argsJSON(t *testing.T, args any) string {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	return string(data)
}

func newAlertReportSession(t *testing.T, rt *runtime.Runtime, instanceID, userID string, attrs map[string]any, userAgent string, permissions ...runtime.Permission) *ai.Session {
	t.Helper()
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims: &runtime.SecurityClaims{
			UserID:         userID,
			UserAttributes: attrs,
			Permissions:    permissions,
		},
		UserAgent: userAgent,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, s.Flush(t.Context()))
	})
	return s
}

// requireQueryRuns runs a saved MetricsViewAggregation query, as alerts and reports run it.
func requireQueryRuns(t *testing.T, rt *runtime.Runtime, instanceID, queryArgsJSON string) {
	t.Helper()
	res, _, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "legacy_metrics",
		ResolverProperties: map[string]any{"query_name": "MetricsViewAggregation", "query_args_json": queryArgsJSON},
		Claims:             &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err, queryArgsJSON)
	require.NoError(t, res.Close())
}

func parseAggregation(t *testing.T, queryArgsJSON string) *runtimev1.MetricsViewAggregationRequest {
	t.Helper()
	q := &runtimev1.MetricsViewAggregationRequest{}
	require.NoError(t, protojson.Unmarshal([]byte(queryArgsJSON), q))
	return q
}

func requireExpr(t *testing.T, want string, got *runtimev1.Expression) {
	t.Helper()
	data, err := protojson.Marshal(got)
	require.NoError(t, err)
	require.JSONEq(t, want, string(data))
}

// requireProtoJSON checks that a proto message has the given JSON form.
func requireProtoJSON(t *testing.T, want string, got proto.Message) {
	t.Helper()
	data, err := protojson.Marshal(got)
	require.NoError(t, err)
	require.JSONEq(t, want, string(data))
}

// requireSameAlert compares alert options, with their resolver properties compared in their JSON form.
func requireSameAlert(t *testing.T, want, got *drivers.AlertOptions) {
	t.Helper()
	w, g := *want, *got
	require.Equal(t, asJSONValue(t, w.ResolverProperties), asJSONValue(t, g.ResolverProperties))
	w.ResolverProperties, g.ResolverProperties = nil, nil
	require.Equal(t, w, g)
}

func saveAlert(t *testing.T, s *ai.Session, args string) *ai.SaveAlertResult {
	t.Helper()
	var res *ai.SaveAlertResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SaveAlertName, &res, json.RawMessage(args))
	require.NoError(t, err)
	return res
}

// asJSON converts a value to its JSON form, as the model reads it.
func asJSON(t *testing.T, v any) map[string]any {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal(data, &res))
	return res
}

func asJSONValue(t *testing.T, v any) any {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	var res any
	require.NoError(t, json.Unmarshal(data, &res))
	return res
}

func findByName(t *testing.T, items []map[string]any, name string) map[string]any {
	t.Helper()
	for _, item := range items {
		if item["name"] == name {
			return item
		}
	}
	require.Failf(t, "not listed", "%q", name)
	return nil
}

func changedFields(changes []ai.FieldChange) []string {
	var fields []string
	for _, c := range changes {
		fields = append(fields, c.Field)
	}
	return fields
}

// fakeAlertReportAdmin is an admin service that stores the alerts and reports saved through it, with versions, and records the saves.
type fakeAlertReportAdmin struct {
	drivers.AdminService // Not implemented; calls to methods that are not overridden panic
	mu                   sync.Mutex
	alerts               []savedAlert
	reports              []savedReport
	storedAlerts         map[string]*storedAlert
	storedReports        map[string]*storedReport
	versions             int
	err                  error
}

type savedAlert struct {
	userID, name, version string
	opts                  *drivers.AlertOptions
}

type savedReport struct {
	userID, name, version string
	opts                  *drivers.ReportOptions
}

type storedAlert struct {
	owner string
	alert drivers.ManagedAlert
}

type storedReport struct {
	owner  string
	report drivers.ManagedReport
}

// storeAlert stores an alert as if a user saved it from the dialog, and returns its version.
func (a *fakeAlertReportAdmin) storeAlert(owner, name string, opts *drivers.AlertOptions) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.storedAlerts == nil {
		a.storedAlerts = map[string]*storedAlert{}
	}
	a.versions++
	v := fmt.Sprintf("v%d", a.versions)
	a.storedAlerts[name] = &storedAlert{owner: owner, alert: drivers.ManagedAlert{Name: name, Version: v, Options: opts}}
	return v
}

// storeReport stores a report as if a user saved it from the dialog, and returns its version.
func (a *fakeAlertReportAdmin) storeReport(owner, name string, opts *drivers.ReportOptions) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.storedReports == nil {
		a.storedReports = map[string]*storedReport{}
	}
	a.versions++
	v := fmt.Sprintf("v%d", a.versions)
	a.storedReports[name] = &storedReport{owner: owner, report: drivers.ManagedReport{Name: name, Version: v, Options: opts}}
	return v
}

func (a *fakeAlertReportAdmin) ListAlerts(ctx context.Context, userID string) ([]*drivers.ManagedAlert, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var res []*drivers.ManagedAlert
	for _, s := range a.storedAlerts {
		if s.owner == userID {
			alert := s.alert
			res = append(res, &alert)
		}
	}
	return res, nil
}

func (a *fakeAlertReportAdmin) ListReports(ctx context.Context, userID string) ([]*drivers.ManagedReport, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var res []*drivers.ManagedReport
	for _, s := range a.storedReports {
		if s.owner == userID {
			report := s.report
			res = append(res, &report)
		}
	}
	return res, nil
}

func (a *fakeAlertReportAdmin) SaveAlert(ctx context.Context, userID, name, version string, opts *drivers.AlertOptions) (*drivers.SavedResource, error) {
	a.mu.Lock()
	a.alerts = append(a.alerts, savedAlert{userID: userID, name: name, version: version, opts: opts})
	err := a.err
	if err == nil && name != "" {
		if s, ok := a.storedAlerts[name]; !ok || s.owner != userID || s.alert.Version != version {
			err = status.Errorf(codes.Aborted, "alert %q changed since it was listed", name)
		}
	}
	if name == "" {
		name = fmt.Sprintf("alert-%d", len(a.alerts))
	}
	a.mu.Unlock()
	if err != nil {
		return nil, err
	}
	v := a.storeAlert(userID, name, opts)
	return &drivers.SavedResource{Name: name, URL: "https://ui.example.com/-/alerts/" + name, Version: v}, nil
}

func (a *fakeAlertReportAdmin) SaveReport(ctx context.Context, userID, name, version string, opts *drivers.ReportOptions) (*drivers.SavedResource, error) {
	a.mu.Lock()
	a.reports = append(a.reports, savedReport{userID: userID, name: name, version: version, opts: opts})
	err := a.err
	if err == nil && name != "" {
		if s, ok := a.storedReports[name]; !ok || s.owner != userID || s.report.Version != version {
			err = status.Errorf(codes.Aborted, "report %q changed since it was listed", name)
		}
	}
	if name == "" {
		name = fmt.Sprintf("report-%d", len(a.reports))
	}
	a.mu.Unlock()
	if err != nil {
		return nil, err
	}
	v := a.storeReport(userID, name, opts)
	return &drivers.SavedResource{Name: name, URL: "https://ui.example.com/-/reports/" + name, Version: v}, nil
}

func (a *fakeAlertReportAdmin) reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.alerts = nil
	a.reports = nil
}

func (a *fakeAlertReportAdmin) lastAlert(t *testing.T) savedAlert {
	a.mu.Lock()
	defer a.mu.Unlock()
	require.NotEmpty(t, a.alerts, "no alert reached the admin service")
	return a.alerts[len(a.alerts)-1]
}

func (a *fakeAlertReportAdmin) lastReport(t *testing.T) savedReport {
	a.mu.Lock()
	defer a.mu.Unlock()
	require.NotEmpty(t, a.reports, "no report reached the admin service")
	return a.reports[len(a.reports)-1]
}

func ptr[T any](v T) *T { return &v }
