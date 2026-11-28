package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

// This file has the helpers shared by the tools that create, list and delete the user's alerts and reports.
// The tools only create and delete: to change an alert or report, the user edits it in its dialog.

// confirmBeforeCreating is the part of the create tools' descriptions that makes the model confirm with the user.
const confirmBeforeCreating = "Before calling this tool, show the user a short preview of what you will create (name, data, condition or content, schedule and recipients), then end your turn and wait for them to confirm. " +
	"It can only send to the user and other members of the project, by email or to their Slack user, and sends to the user if you pass no recipients. "

// DimensionFilter filters the data of an alert or report to some values of a dimension, or to the values that contain a text.
type DimensionFilter struct {
	Dimension string   `json:"dimension" jsonschema:"Name of the dimension to filter by."`
	Values    []string `json:"values,omitempty" jsonschema:"Values of the dimension to keep. Pass either values or contains."`
	Contains  string   `json:"contains,omitempty" jsonschema:"Text the values to keep contain, as the dashboard's Contains filter matches them. Pass either values or contains."`
	Exclude   bool     `json:"exclude,omitempty" jsonschema:"If true, exclude these values instead of keeping them."`
}

// Schedule is when a report is sent or an alert is checked.
type Schedule struct {
	Cron     string `json:"cron" jsonschema:"Cron expression at a minute M and hour H of the day: 'M H * * *' daily, 'M H * * 1-5' on weekdays, 'M H * * D' weekly on day D (0 is Sunday, 6 Saturday) or 'M H 1 * *' monthly on the 1st. For example '0 9 * * 1' for Mondays at 9:00. Other schedules are not available."`
	TimeZone string `json:"time_zone,omitempty" jsonschema:"IANA time zone of the cron expression, for example 'Europe/Madrid'. Defaults to UTC; use the user's time zone if you know it."`
}

// Recipients are who an alert or report is sent to.
// They can only be members of the project (see drivers.AdminService.CreateAlertForUser).
type Recipients struct {
	Emails     []string `json:"emails,omitempty" jsonschema:"Emails of project members to send to. If no recipient is given, it is sent to the user."`
	SlackUsers []string `json:"slack_users,omitempty" jsonschema:"Emails of project members to send to in Slack, if the project has Slack configured."`
}

// CreatedResource is the result of the tools that create an alert or report.
type CreatedResource struct {
	Name           string `json:"name" jsonschema:"Name of the alert or report."`
	URL            string `json:"url" jsonschema:"URL of its page, where the user can see, edit and delete it. Give it to the user."`
	ReconcileError string `json:"reconcile_error,omitempty" jsonschema:"Error it has, if any. Tell the user."`
}

// checkAlertReportAccess checks that the session belongs to a user who may use AI in a prod deployment and is present to confirm what is created.
// Alerts and reports created from the UI are stored in the prod environment, so other deployments (like editable branch deployments) can't create them.
// Users are told apart from services by their email attribute, which only user JWTs carry.
// AI reports and alerts run as a user, but nobody is there to confirm what the AI creates on their behalf.
// The admin service additionally checks that the user may create alerts or reports in the project.
func checkAlertReportAccess(ctx context.Context, rt *runtime.Runtime) (bool, error) {
	s := GetSession(ctx)
	claims := s.Claims()
	if !claims.Can(runtime.UseAI) || claims.UserID == "" {
		return false, nil
	}
	if email, _ := claims.UserAttributes["email"].(string); email == "" {
		return false, nil
	}
	if s.CatalogSession().UserAgent == reportUserAgent {
		return false, nil
	}

	inst, err := rt.Instance(ctx, s.InstanceID())
	if err != nil {
		return false, err
	}
	return inst.Environment == "prod", nil
}

// validExplore returns the valid spec of an explore the user can access.
func validExplore(ctx context.Context, rt *runtime.Runtime, name string) (*runtimev1.ExploreSpec, error) {
	if name == "" {
		return nil, errors.New("invalid input: explore is required")
	}
	explore, _, err := (&AnalystAgent{Runtime: rt}).getValidExploreAndMetricsView(ctx, name)
	if err != nil {
		return nil, err
	}
	return explore.GetExplore().State.ValidSpec, nil
}

// checkExploreFields checks that the measures and dimensions are available in the explore.
func checkExploreFields(spec *runtimev1.ExploreSpec, explore string, measures, dimensions []string) error {
	for _, m := range measures {
		if !slices.Contains(spec.Measures, m) {
			return fmt.Errorf("measure %q is not available in explore %q, which has the measures: %s", m, explore, strings.Join(spec.Measures, ", "))
		}
	}
	for _, d := range dimensions {
		if !slices.Contains(spec.Dimensions, d) {
			return fmt.Errorf("dimension %q is not available in explore %q, which has the dimensions: %s", d, explore, strings.Join(spec.Dimensions, ", "))
		}
	}
	return nil
}

// filtersExpression converts dimension filters into the where expression the dialogs save: an AND of IN or NIN conditions,
// and for a contains filter, the LIKE or NLIKE condition of the dashboard's Contains mode, with the text between % wildcards.
// It also returns the dimensions the filters use, and nil if there are no filters.
func filtersExpression(filters []DimensionFilter) (*runtimev1.Expression, []string, error) {
	if len(filters) == 0 {
		return nil, nil, nil
	}
	var exprs []*runtimev1.Expression
	var dims []string
	for _, f := range filters {
		if f.Dimension == "" || (len(f.Values) == 0) == (f.Contains == "") {
			return nil, nil, errors.New("invalid input: every filter needs a dimension, and either at least one value or a contains text")
		}
		dims = append(dims, f.Dimension)
		if f.Contains != "" {
			pattern := expressionpb.String("%" + f.Contains + "%")
			if f.Exclude {
				exprs = append(exprs, expressionpb.NotLike(expressionpb.Identifier(f.Dimension), pattern))
			} else {
				exprs = append(exprs, expressionpb.Like(expressionpb.Identifier(f.Dimension), pattern))
			}
			continue
		}
		var vals []*runtimev1.Expression
		for _, v := range f.Values {
			vals = append(vals, expressionpb.String(v))
		}
		if f.Exclude {
			exprs = append(exprs, expressionpb.NotIn(expressionpb.Identifier(f.Dimension), vals))
		} else {
			exprs = append(exprs, expressionpb.In(expressionpb.Identifier(f.Dimension), vals))
		}
	}
	return expressionpb.And(exprs), dims, nil
}

// checkTimeRange checks that a time range is a valid Rill time expression.
func checkTimeRange(expr string) error {
	if _, err := rilltime.Parse(expr, rilltime.ParseOptions{}); err != nil {
		return fmt.Errorf("invalid time range %q: %w", expr, err)
	}
	return nil
}

// comparisonOffsets are the Rill time offsets of the dashboards' comparison options (TIME_COMPARISON in the UI), by the name the comparison arguments take.
var comparisonOffsets = map[string]string{
	"previous_period":  "-1P",
	"previous_day":     "-1D",
	"previous_week":    "-1W",
	"previous_month":   "-1M",
	"previous_quarter": "-1Q",
	"previous_year":    "-1Y",
}

// comparisonExpression returns the time range a Rill time expression is compared with, as the dialogs write it:
// the expression offset by a comparison option, which defaults to the previous period.
func comparisonExpression(expr, comparison string) (string, error) {
	if comparison == "" {
		comparison = "previous_period"
	}
	offset, ok := comparisonOffsets[comparison]
	if !ok {
		return "", fmt.Errorf("invalid comparison %q: use 'previous_period', 'previous_day', 'previous_week', 'previous_month', 'previous_quarter' or 'previous_year'", comparison)
	}
	cmp := expr + " offset " + offset
	if err := checkTimeRange(cmp); err != nil {
		return "", fmt.Errorf("can't compare time range %q with the %s: %w", expr, strings.ReplaceAll(comparison, "_", " "), err)
	}
	return cmp, nil
}

// dialogCron matches the cron expressions the alert and report dialogs can edit (see ScheduleForm.svelte):
// at one time of day, daily, on weekdays, weekly on one day, or monthly on the 1st.
// The dialogs read other expressions as the closest of these, so saving an edit from the dialog would move the schedule.
var dialogCron = regexp.MustCompile(`^([0-9]|[1-5][0-9]) ([0-9]|1[0-9]|2[0-3]) (\* \* (\*|1-5|[0-6])|1 \* \*)$`)

// checkSchedule checks that a schedule's cron expression is one the user can edit in the dialogs, and that its time zone is valid.
func checkSchedule(s *Schedule) error {
	if s.Cron == "" {
		return errors.New("invalid input: schedule.cron is required")
	}
	if !dialogCron.MatchString(s.Cron) {
		return fmt.Errorf("invalid cron expression %q: use 'M H * * *' for daily, 'M H * * 1-5' for weekdays, 'M H * * D' for weekly on day D from 0 (Sunday) to 6, or 'M H 1 * *' for monthly on the 1st", s.Cron)
	}
	return checkTimeZone(s.TimeZone)
}

// checkTimeZone checks that a time zone is empty or a valid IANA time zone.
func checkTimeZone(tz string) error {
	if tz == "" {
		return nil
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("invalid time zone %q: %w", tz, err)
	}
	return nil
}

// createdResource returns the result of creating an alert or report, with the reconcile error it already has, if any.
// The admin service returns once the project parser has created the resource, so a parse error has failed the creation already.
func createdResource(ctx context.Context, rt *runtime.Runtime, kind, name, url string) (*CreatedResource, error) {
	ctrl, err := rt.Controller(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return nil, err
	}
	res := &CreatedResource{Name: name, URL: url}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: kind, Name: name}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return res, nil
		}
		return nil, err
	}
	res.ReconcileError = r.Meta.ReconcileError
	return res, nil
}

// ownedResources returns the alerts or reports that the session's user created from the UI or the chat, from the catalog.
// Alerts and reports written in the project's code, and those of other users, are left out.
func ownedResources(ctx context.Context, rt *runtime.Runtime, kind string) ([]*runtimev1.Resource, error) {
	s := GetSession(ctx)
	ctrl, err := rt.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	all, err := ctrl.List(ctx, kind, "", false)
	if err != nil {
		return nil, err
	}
	var res []*runtimev1.Resource
	for _, r := range all {
		var annotations map[string]string
		if a := r.GetAlert(); a != nil {
			annotations = a.Spec.Annotations
		} else if rep := r.GetReport(); rep != nil {
			annotations = rep.Spec.Annotations
		}
		if annotations["admin_managed"] == "true" && annotations["admin_owner_user_id"] == s.Claims().UserID {
			res = append(res, r)
		}
	}
	slices.SortFunc(res, func(a, b *runtimev1.Resource) int { return strings.Compare(a.Meta.Name.Name, b.Meta.Name.Name) })
	return res, nil
}

// resourcePageURL returns the URL of an alert's or report's page (dir "alerts" or "reports"), or "" if the instance has no frontend URL.
func resourcePageURL(ctx context.Context, rt *runtime.Runtime, dir, name string) (string, error) {
	inst, err := rt.Instance(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return "", err
	}
	if inst.FrontendURL == "" {
		return "", nil
	}
	return url.JoinPath(inst.FrontendURL, "-", dir, name)
}

// DeleteResourceArgs are the arguments of the tools that delete one of the user's alerts or reports.
type DeleteResourceArgs struct {
	Name      string `json:"name" jsonschema:"Name of the alert or report to delete, as listed."`
	Confirmed bool   `json:"confirmed,omitempty" jsonschema:"Set it only once the user confirmed, in their message right after your first call for this name, that they want to delete it."`
}

// DeleteResourceResult is the result of the tools that delete one of the user's alerts or reports.
type DeleteResourceResult struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	// Status is "confirmation_required" until the user confirms, and then "deleted".
	Status string `json:"status"`
}

// deleteOwnedResource deletes an alert or report the user owns, once they confirmed it.
// The first call for a name only returns "confirmation_required". A call with confirmed=true deletes it,
// but only in the turn that answers that first call: what the tools read can't stand in for the user's confirmation,
// and a proposal the user moved on from doesn't stay open.
func deleteOwnedResource(ctx context.Context, rt *runtime.Runtime, tool, kind string, args *DeleteResourceArgs, del func(userID, name string) error) (*DeleteResourceResult, error) {
	owned, err := ownedResources(ctx, rt, kind)
	if err != nil {
		return nil, err
	}
	i := slices.IndexFunc(owned, func(r *runtimev1.Resource) bool { return r.Meta.Name.Name == args.Name })
	if i < 0 {
		return nil, fmt.Errorf("%q is not one of the user's own: list them to find its name", args.Name)
	}
	res := &DeleteResourceResult{Name: args.Name, DisplayName: resourceDisplayName(owned[i])}

	s := GetSession(ctx)
	if !args.Confirmed || !proposedInPreviousTurn(s, tool, args.Name) {
		res.Status = "confirmation_required"
		return res, nil
	}
	if err := del(s.Claims().UserID, args.Name); err != nil {
		return nil, err
	}
	res.Status = "deleted"
	return res, nil
}

// proposedInPreviousTurn returns whether a call of tool for the given name returned "confirmation_required" in the turn right before the current one,
// which means that the user's latest message answers it.
func proposedInPreviousTurn(s *Session, tool, name string) bool {
	current := s.RootID()
	turns := s.Messages(FilterByRoot(), FilterByType(MessageTypeCall))
	i := slices.IndexFunc(turns, func(m *Message) bool { return m.ID == current })
	if i < 1 {
		return false
	}
	previous := turns[i-1].ID
	for _, m := range s.Messages(FilterByTool(tool), FilterByType(MessageTypeResult)) {
		var res DeleteResourceResult
		if json.Unmarshal([]byte(m.Content), &res) != nil || res.Name != name || res.Status != "confirmation_required" {
			continue
		}
		root := m
		for root.ParentID != "" {
			parent, ok := s.Message(FilterByID(root.ParentID))
			if !ok {
				break
			}
			root = parent
		}
		if root.ID == previous {
			return true
		}
	}
	return false
}

// resourceDisplayName returns an alert's or report's display name.
func resourceDisplayName(r *runtimev1.Resource) string {
	if a := r.GetAlert(); a != nil {
		return a.Spec.DisplayName
	}
	return r.GetReport().GetSpec().GetDisplayName()
}
