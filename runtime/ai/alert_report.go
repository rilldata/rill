package ai

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/robfig/cron/v3"
)

// confirmBeforeSaving is the part of the save tools' descriptions that makes the model confirm with the user.
// The save tools act on the user's behalf and notify other people, so the user must see what will be saved first.
const confirmBeforeSaving = "Before calling this tool, show the user a short preview of what you will save (name, data, condition or content, schedule, recipients and, for reports, who it opens as), then end your turn and wait for them to confirm; only call it after they confirm, and save again with their corrections if they ask for changes. "

// DimensionFilter filters the data of an alert or report to some values of a dimension, or to the values that contain a text.
type DimensionFilter struct {
	Dimension string   `json:"dimension" jsonschema:"Name of the dimension to filter by."`
	Values    []string `json:"values,omitempty" jsonschema:"Values of the dimension to keep. Pass either values or contains."`
	Contains  string   `json:"contains,omitempty" jsonschema:"Text the values to keep contain, as the dashboard's Contains filter matches them. Pass either values or contains."`
	Exclude   bool     `json:"exclude,omitempty" jsonschema:"If true, exclude these values instead of keeping them."`
}

// Schedule is when a report is sent or an alert is checked.
type Schedule struct {
	Cron     string `json:"cron" jsonschema:"Standard 5-field cron expression, for example '0 9 * * 1' for Mondays at 9:00."`
	TimeZone string `json:"time_zone,omitempty" jsonschema:"IANA time zone of the cron expression, for example 'Europe/Madrid'. Defaults to UTC; use the user's time zone if you know it."`
}

// Recipients are who an alert or report notifies.
type Recipients struct {
	Emails        []string `json:"emails,omitempty" jsonschema:"Email addresses to notify. If no recipient is given, the user's own email is used."`
	SlackUsers    []string `json:"slack_users,omitempty" jsonschema:"Emails of Slack users to notify, if the project has Slack configured."`
	SlackChannels []string `json:"slack_channels,omitempty" jsonschema:"Slack channels to notify, for example '#sales', if the project has Slack configured."`
}

// checkAlertReportAccess checks that the session belongs to a user who may use AI in a prod deployment and is present to confirm what is saved.
// Managed alerts and reports are stored in the prod environment, so other deployments (like editable branch deployments) can't save them.
// Users are told apart from services by their email attribute, which only user JWTs carry.
// AI reports run as a user, but nobody is there to confirm what the AI saves on their behalf.
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

// validExplore returns the name and valid spec of an explore the user can access.
// The name may also be that of a metrics view, since the analyst's tools list metrics views rather than explores:
// then it's the one explore the user can access on that metrics view, and an error names the candidates if there are none or several.
func validExplore(ctx context.Context, rt *runtime.Runtime, name string) (string, *runtimev1.ExploreSpec, error) {
	if name == "" {
		return "", nil, errors.New("invalid input: explore is required")
	}
	analyst := &AnalystAgent{Runtime: rt}
	explore, _, exploreErr := analyst.getValidExploreAndMetricsView(ctx, name)
	if exploreErr == nil {
		return name, explore.GetExplore().State.ValidSpec, nil
	}

	ctrl, err := rt.Controller(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return "", nil, err
	}
	all, err := ctrl.List(ctx, runtime.ResourceKindExplore, "", false)
	if err != nil {
		return "", nil, err
	}
	var accessible, onMetricsView []string
	specs := map[string]*runtimev1.ExploreSpec{}
	for _, r := range all {
		// The same checks as for an explore passed by name, so a candidate is one that would be accepted.
		e, _, err := analyst.getValidExploreAndMetricsView(ctx, r.Meta.Name.Name)
		if err != nil {
			continue
		}
		spec := e.GetExplore().State.ValidSpec
		accessible = append(accessible, r.Meta.Name.Name)
		if spec.MetricsView == name {
			onMetricsView = append(onMetricsView, r.Meta.Name.Name)
			specs[r.Meta.Name.Name] = spec
		}
	}
	slices.Sort(accessible)
	slices.Sort(onMetricsView)
	switch {
	case len(onMetricsView) == 1:
		return onMetricsView[0], specs[onMetricsView[0]], nil
	case len(onMetricsView) > 1:
		return "", nil, fmt.Errorf("%q is a metrics view with several explores: pass one of them as the explore: %s", name, strings.Join(onMetricsView, ", "))
	case len(accessible) == 0:
		return "", nil, fmt.Errorf("explore %q not found, and the user can't access any explore: %w", name, exploreErr)
	default:
		return "", nil, fmt.Errorf("explore %q not found, and no explore uses a metrics view of that name: pass one of the explores: %s: %w", name, strings.Join(accessible, ", "), exploreErr)
	}
}

// checkExploreFields checks that the measures and dimensions are available in the explore.
func checkExploreFields(spec *runtimev1.ExploreSpec, exploreName string, measures, dimensions []string) error {
	for _, m := range measures {
		if !slices.Contains(spec.Measures, m) {
			return fmt.Errorf("measure %q is not available in explore %q, which has the measures: %s", m, exploreName, strings.Join(spec.Measures, ", "))
		}
	}
	for _, d := range dimensions {
		if !slices.Contains(spec.Dimensions, d) {
			return fmt.Errorf("dimension %q is not available in explore %q, which has the dimensions: %s", d, exploreName, strings.Join(spec.Dimensions, ", "))
		}
	}
	return nil
}

// filterDimensions returns the dimensions the filters use.
func filterDimensions(filters []DimensionFilter) []string {
	var dims []string
	for _, f := range filters {
		dims = append(dims, f.Dimension)
	}
	return dims
}

// filtersExpression converts dimension filters into the where expression the dialogs save: an AND of IN or NIN conditions,
// and for a contains filter, the LIKE or NLIKE condition of the dashboard's Contains mode, with the text between % wildcards.
// It returns nil if there are no filters.
func filtersExpression(filters []DimensionFilter) (*runtimev1.Expression, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	var exprs []*runtimev1.Expression
	for _, f := range filters {
		if f.Dimension == "" || (len(f.Values) == 0) == (f.Contains == "") {
			return nil, errors.New("invalid input: every filter needs a dimension, and either at least one value or a contains text")
		}
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
	return expressionpb.And(exprs), nil
}

// checkTimeRange checks that a time range is a valid Rill time expression.
func checkTimeRange(expr string) error {
	if _, err := rilltime.Parse(expr, rilltime.ParseOptions{}); err != nil {
		return fmt.Errorf("invalid time range %q: %w", expr, err)
	}
	return nil
}

// comparisonKind is a comparison option of the dashboards' time controls.
type comparisonKind struct {
	name, option, offset string
}

// comparisonKinds are the comparison options of the dashboards' time controls (TIME_COMPARISON in the UI), by the name the comparison arguments take:
// the option's name, which the dialogs save as the ISO offset of a duration, and the offset they append to a Rill time expression.
var comparisonKinds = []comparisonKind{
	{"previous_period", "rill-PP", "-1P"},
	{"previous_day", "rill-PD", "-1D"},
	{"previous_week", "rill-PW", "-1W"},
	{"previous_month", "rill-PM", "-1M"},
	{"previous_quarter", "rill-PQ", "-1Q"},
	{"previous_year", "rill-PY", "-1Y"},
}

// lookupComparisonKind returns the comparison kind of a comparison argument, which defaults to the previous period.
func lookupComparisonKind(name string) (comparisonKind, error) {
	if name == "" {
		name = "previous_period"
	}
	var names []string
	for _, k := range comparisonKinds {
		if k.name == name {
			return k, nil
		}
		names = append(names, "'"+k.name+"'")
	}
	return comparisonKind{}, fmt.Errorf("invalid comparison %q: use one of %s", name, strings.Join(names, ", "))
}

// comparisonExpression returns the time range a Rill time expression is compared with, as the dialogs write it:
// the expression offset by the comparison kind, which defaults to the previous period.
func comparisonExpression(expr, kind string) (string, error) {
	k, err := lookupComparisonKind(kind)
	if err != nil {
		return "", err
	}
	cmp := expr + " offset " + k.offset
	if err := checkTimeRange(cmp); err != nil {
		return "", fmt.Errorf("can't compare time range %q with the %s: %w", expr, strings.ReplaceAll(k.name, "_", " "), err)
	}
	return cmp, nil
}

// comparisonKindOfExpression returns the comparison kind of a comparison saved as a Rill time expression, or "" if it's none of them.
func comparisonKindOfExpression(expr string) string {
	for _, k := range comparisonKinds {
		if strings.HasSuffix(expr, " offset "+k.offset) {
			return k.name
		}
	}
	return ""
}

// checkSchedule checks that a schedule's cron expression and time zone are valid.
func checkSchedule(s *Schedule) error {
	if s.Cron == "" {
		return errors.New("invalid input: schedule.cron is required")
	}
	if _, err := cron.ParseStandard(s.Cron); err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", s.Cron, err)
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
