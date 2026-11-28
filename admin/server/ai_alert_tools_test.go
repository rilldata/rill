package server_test

import (
	"encoding/json"
	"fmt"
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestAIAlertTools tests creating, listing and deleting alerts with the AI tools, as the chat does:
// through the admin service, which writes them like the alert dialog, and the runtime, which parses and reconciles them.
func TestAIAlertTools(t *testing.T) {
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

	createAlert := func(args *ai.CreateAlertArgs) (*ai.CreatedResource, error) {
		var res *ai.CreatedResource
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateAlertName, &res, args)
		return res, err
	}

	var formulaAlert, aiAlert string
	t.Run("creates a formula alert like the dialog", func(t *testing.T) {
		res, err := createAlert(&ai.CreateAlertArgs{
			DisplayName: "Revenue drop",
			Explore:     "e1",
			Measure:     "revenue",
			Dimension:   "country",
			Criteria:    []ai.AlertCriterion{{Type: "percent_change", Operator: "lt", Value: -10}},
			TimeRange:   "7D as of latest/D",
			Schedule:    &ai.Schedule{Cron: "0 9 * * *", TimeZone: "Europe/Madrid"},
		})
		require.NoError(t, err)
		require.Empty(t, res.ReconcileError)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/alerts/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, res.Name), res.URL)
		formulaAlert = res.Name

		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindAlert, res.Name).GetAlert().Spec
		require.Equal(t, "0 9 * * *", spec.RefreshSchedule.Cron)
		require.Equal(t, "Europe/Madrid", spec.RefreshSchedule.TimeZone)
		require.Equal(t, []any{f.viewerUser.Email}, spec.Notifiers[0].Properties.AsMap()["recipients"])
		// The query is the one the alert dialog saves for the same alert
		var query map[string]any
		require.NoError(t, json.Unmarshal([]byte(spec.ResolverProperties.AsMap()["query_args_json"].(string)), &query))
		require.JSONEq(t, `{
			"metricsView": "mv1",
			"measures": [
				{"name": "revenue"},
				{"name": "revenue_delta", "comparisonDelta": {"measure": "revenue"}},
				{"name": "revenue_delta_perc", "comparisonRatio": {"measure": "revenue"}}
			],
			"dimensions": [{"name": "country"}],
			"having": {"cond": {"op": "OPERATION_AND", "exprs": [{"cond": {"op": "OPERATION_LT", "exprs": [{"ident": "revenue_delta_perc"}, {"val": -0.1}]}}]}},
			"timeRange": {"expression": "7D as of latest/D", "timeZone": "Europe/Madrid"},
			"comparisonTimeRange": {"expression": "7D as of latest/D offset -1P", "timeZone": "Europe/Madrid"},
			"sort": [{"name": "revenue", "desc": true}]
		}`, mustJSON(t, query))
	})

	t.Run("creates an AI alert checked at most once an hour", func(t *testing.T) {
		args := &ai.CreateAlertArgs{
			DisplayName: "Unusual revenue",
			Explore:     "e1",
			Condition:   "Revenue drops unusually compared to recent weeks",
			Schedule:    &ai.Schedule{Cron: "0 * * * *"},
		}
		res, err := createAlert(args)
		require.NoError(t, err)
		require.Empty(t, res.ReconcileError)
		aiAlert = res.Name
		spec := testruntime.GetResource(t, f.fix.Runtime, f.instance.ID, runtime.ResourceKindAlert, res.Name).GetAlert().Spec
		require.Equal(t, "ai", spec.Resolver)
		require.Equal(t, map[string]any{"prompt": args.Condition, "explore": "e1"}, spec.ResolverProperties.AsMap())
		require.Equal(t, f.viewerUser.ID, spec.GetQueryForUserId())

		args.Schedule = &ai.Schedule{Cron: "*/30 * * * *"}
		_, err = createAlert(args)
		require.ErrorContains(t, err, "at most once an hour")
		args.Schedule = nil
		_, err = createAlert(args)
		require.ErrorContains(t, err, "needs a schedule")
	})

	t.Run("only sends to project members", func(t *testing.T) {
		_, err := createAlert(&ai.CreateAlertArgs{
			DisplayName: "Revenue drop",
			Explore:     "e1",
			Measure:     "revenue",
			Criteria:    []ai.AlertCriterion{{Operator: "lt", Value: 0}},
			TimeRange:   "7D as of latest/D",
			Recipients:  &ai.Recipients{Emails: []string{f.outsiderUser.Email}},
		})
		require.ErrorContains(t, err, "not a member of the project")
	})

	t.Run("lists the user's own alerts", func(t *testing.T) {
		_, err := f.owner.CreateAlert(ctx, &adminv1.CreateAlertRequest{Org: f.org, Project: f.project, Options: &adminv1.AlertOptions{
			DisplayName:     "Someone else's alert",
			MetricsViewName: "mv1",
			QueryName:       "MetricsViewAggregation",
			QueryArgsJson:   `{"metrics_view":"mv1","measures":[{"name":"revenue"}]}`,
			WebOpenPath:     "/explore/e1",
		}})
		require.NoError(t, err)

		var res *ai.ListAlertsResult
		_, err = s.CallTool(ctx, ai.RoleUser, ai.ListAlertsName, &res, &ai.ListAlertsArgs{})
		require.NoError(t, err)
		require.Len(t, res.Alerts, 2)
		byName := map[string]*ai.AlertInfo{}
		for _, a := range res.Alerts {
			byName[a.Name] = a
		}
		require.Equal(t, "(revenue_delta_perc < -0.1)", byName[formulaAlert].Condition)
		require.Equal(t, "Revenue drops unusually compared to recent weeks", byName[aiAlert].Condition)
		require.Equal(t, "e1", byName[aiAlert].Explore)
		require.Equal(t, fmt.Sprintf("%s/%s/%s/-/alerts/%s", f.fix.Admin.URLs.Frontend(), f.org, f.project, aiAlert), byName[aiAlert].URL)
	})

	t.Run("deletes an alert once the user confirmed it in a later message", func(t *testing.T) {
		deleteAlert := func(s *ai.Session, confirmed bool) (*ai.DeleteResourceResult, *ai.CallResult) {
			var res *ai.DeleteResourceResult
			call, err := s.CallTool(ctx, ai.RoleUser, ai.DeleteAlertName, &res, &ai.DeleteResourceArgs{Name: aiAlert, Confirmed: confirmed})
			require.NoError(t, err)
			return res, call
		}

		// Without an earlier call for it, a confirmed call still needs a confirmation
		res, call := deleteAlert(s, true)
		require.Equal(t, "confirmation_required", res.Status)
		// In the same turn, it still needs a confirmation
		res, _ = deleteAlert(&ai.Session{BaseSession: s.BaseSession, ParentID: call.Call.ID}, true)
		require.Equal(t, "confirmation_required", res.Status)
		// In a later turn, it deletes the alert
		res, _ = deleteAlert(s, true)
		require.Equal(t, "deleted", res.Status)
		ctrl, err := f.fix.Runtime.Controller(ctx, f.instance.ID)
		require.NoError(t, err)
		_, err = ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: aiAlert}, false)
		require.ErrorIs(t, err, drivers.ErrResourceNotFound)
	})
}

func mustJSON(t *testing.T, v any) string {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return string(data)
}
