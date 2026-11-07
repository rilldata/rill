package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const SaveAlertName = "save_alert"

// SaveAlert creates a KPI alert for the current user, or edits one of the user's own, through the admin service.
// What it saves is what the alert dialog saves: the alert is owned by the user, listed in their alerts and editable in the dialog.
// An edit only changes the arguments it sends, so whatever the dialog saved that the arguments can't express is kept.
type SaveAlert struct {
	Runtime *runtime.Runtime
}

var _ Tool[*SaveAlertArgs, *SaveAlertResult] = (*SaveAlert)(nil)

type SaveAlertArgs struct {
	Name                 string            `json:"name,omitempty" jsonschema:"Name of one of the user's own alerts to edit, as returned by list_alerts. Leave empty to create a new alert."`
	Version              string            `json:"version,omitempty" jsonschema:"Version of the alert to edit, as returned by list_alerts or by the last save_alert of it. Required to edit."`
	DisplayName          string            `json:"display_name,omitempty" jsonschema:"Display name of the alert. Required to create one."`
	Explore              string            `json:"explore,omitempty" jsonschema:"Name of the explore dashboard the alert watches. Its metrics view provides the measure and dimensions, and the alert opens on it. The name of a metrics view also works if it has a single explore; if not, the error lists the explores to choose from. Required to create one."`
	Measure              string            `json:"measure,omitempty" jsonschema:"Name of the measure the alert checks. Required to create a formula alert; for an AI alert, it optionally narrows what the AI looks at, and empty removes it."`
	Dimension            *string           `json:"dimension,omitempty" jsonschema:"Optional dimension to split the measure by, so the criteria are checked for each of its values (for an AI alert, it narrows what the AI looks at). Null removes it."`
	TimeRange            *string           `json:"time_range,omitempty" jsonschema:"Time range of the data on every check, as a Rill time expression relative to the latest data, for example '7D as of latest/D' for the last 7 complete days or '1M as of latest/M' for the last complete month. Required to create a formula alert; an AI alert defaults to the explore's default time range, and null removes it."`
	Criteria             []AlertCriterion  `json:"criteria,omitempty" jsonschema:"For a formula alert: conditions on the measure that trigger the alert. At least one to create one; passing them replaces all the current criteria, and turns an AI alert into a formula alert."`
	CriteriaOperator     string            `json:"criteria_operator,omitempty" jsonschema:"For a formula alert: how the criteria combine, 'and' (default) or 'or'."`
	Comparison           string            `json:"comparison,omitempty" jsonschema:"For a formula alert with 'absolute_change' or 'percent_change' criteria: the period the time range is compared with, 'previous_period' (default, the period of the same length right before), 'previous_day', 'previous_week', 'previous_month', 'previous_quarter' or 'previous_year'. Omitted in an edit, the current comparison is kept, also when the time range changes."`
	Condition            string            `json:"condition,omitempty" jsonschema:"For an AI alert: the condition the AI judges on every check, in the user's words, for example 'sales drop unusually compared to recent weeks'. Passing it makes the alert an AI alert. Don't pass it together with criteria."`
	TimeZone             *string           `json:"time_zone,omitempty" jsonschema:"IANA time zone the time range is evaluated in, for example 'Europe/Madrid'. Defaults to the schedule's time zone, or UTC."`
	Where                []DimensionFilter `json:"where,omitempty" jsonschema:"Optional filters on dimensions of the explore. Passing them replaces the current dimension filters, and null removes them; filters saved otherwise, like the dashboards' measure filters, are kept."`
	Schedule             *Schedule         `json:"schedule,omitempty" jsonschema:"When to check the alert. Without a schedule, or with null, it's checked whenever the data refreshes, which is the default."`
	Renotify             string            `json:"renotify,omitempty" jsonschema:"While the alert keeps triggering: 'every_check' (default) notifies on every check; 'after_interval' notifies again after renotify_after_seconds; 'never' notifies only when it starts triggering."`
	RenotifyAfterSeconds uint32            `json:"renotify_after_seconds,omitempty" jsonschema:"With renotify 'after_interval', seconds to wait before notifying again, for example 86400 for daily."`
	Recipients           *Recipients       `json:"recipients,omitempty" jsonschema:"Who to notify. Defaults to the user's own email when creating. Each kind of recipient passed replaces the current ones of that kind. Recipients outside the project see the alert's data with the user's permissions."`

	sent sentArgs
}

func (a *SaveAlertArgs) UnmarshalJSON(data []byte) error {
	type plain SaveAlertArgs
	if err := json.Unmarshal(data, (*plain)(a)); err != nil {
		return err
	}
	return a.sent.parse(data)
}

// MarshalJSON keeps the arguments as they were sent, so an argument sent as null still clears its setting when the call is replayed.
func (a *SaveAlertArgs) MarshalJSON() ([]byte, error) {
	if a.sent.raw != nil {
		return a.sent.raw, nil
	}
	type plain SaveAlertArgs
	return json.Marshal((*plain)(a))
}

// AlertCriterion is a condition on the alert's measure.
type AlertCriterion struct {
	Type     string  `json:"type,omitempty" jsonschema:"What is compared: 'value' (default) the measure itself; 'absolute_change' or 'percent_change' its change from the period of the comparison argument, by default the previous period of the same length; 'percent_of_total' each dimension value's share of the total."`
	Operator string  `json:"operator" jsonschema:"One of 'gt', 'gte', 'lt', 'lte', 'eq' or 'neq'."`
	Value    float64 `json:"value" jsonschema:"Threshold. For percent types, in percent: -10 for a 10% drop."`
}

type SaveAlertResult struct {
	Name           string        `json:"name" jsonschema:"Name of the saved alert. Pass it to save_alert to edit the same alert."`
	Version        string        `json:"version" jsonschema:"Version of the saved alert. Pass it to save_alert to edit it again."`
	URL            string        `json:"url" jsonschema:"URL where the user can see and edit the alert."`
	Changes        []FieldChange `json:"changes,omitempty" jsonschema:"For an edit, the settings it changed, with their old and new values. Tell the user what changed."`
	ReconcileError string        `json:"reconcile_error,omitempty" jsonschema:"Error the alert has, if any. Tell the user and offer to fix it."`
}

func (t *SaveAlert) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  SaveAlertName,
		Title: "Save alert",
		Description: "Creates an alert for the current user on the data of an explore dashboard, or edits one of the user's own alerts. " +
			"Use it when the user asks to be alerted, notified or warned about their data, for example 'alert me if revenue drops below 30k' or 'tell me if sales drop unusually'. " +
			"An alert has a formula condition (criteria: a threshold on a measure or its change, checked exactly and at no AI cost) or an AI condition (condition: a description the AI analyst judges on every check). " +
			"Prefer criteria whenever the condition is a threshold on a measure; use an AI condition only for what a threshold can't express, like 'unusual', 'unexpected' or patterns across dimensions. " +
			"Each check of an AI alert runs the AI analyst once: tell the user in the preview, together with how often it's checked. " +
			"The alert belongs to the user, is listed in their alerts, and they can keep editing it in the alert dialog. Only create a file in the project's code instead when the user explicitly asks for that. " +
			confirmBeforeSaving +
			"Alerts can notify anyone, including people outside the project, but every recipient sees the alert's data with the user's permissions: if a recipient may be outside the project, say so in the preview. " +
			"To edit an alert, find its name, version and current definition with list_alerts, even in a new conversation; if you can't find the alert the user means, ask them instead of creating a new one. " +
			"An edit only changes the arguments you pass: every other setting keeps its current value, and passing null clears an optional one. If the alert changed since it was listed, the edit fails: list it again and reapply the change. " +
			"It can't delete alerts: tell the user to delete them from the Alerts page. Give the user the alert's URL from the response, and for an edit, the changes it lists.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(true), // It notifies people
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Saving alert...",
			"openai/toolInvocation/invoked":  "Saved alert",
		},
	}
}

func (t *SaveAlert) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *SaveAlert) Handler(ctx context.Context, args *SaveAlertArgs) (*SaveAlertResult, error) {
	if err := args.sent.ensure(args); err != nil {
		return nil, err
	}
	s := GetSession(ctx)

	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	// An edit starts from the alert as it's stored, at the version it was listed with.
	var prev *drivers.ManagedAlert
	base := &drivers.AlertOptions{QueryName: "MetricsViewAggregation", Renotify: true}
	if args.Name != "" {
		if args.Version == "" {
			return nil, errors.New("invalid input: version is required to edit an alert: call list_alerts for its current version")
		}
		alerts, err := admin.ListAlerts(ctx, s.Claims().UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to list alerts: %w", err)
		}
		for _, a := range alerts {
			if a.Name == args.Name {
				prev = a
			}
		}
		if prev == nil {
			return nil, fmt.Errorf("alert %q not found: call list_alerts for the names of the user's alerts", args.Name)
		}
		if prev.Version != args.Version {
			return nil, editConflictError("alert", args.Name, ListAlertsName)
		}
		base = prev.Options
	}

	opts, err := t.applyArgs(ctx, args, base, prev == nil)
	if err != nil {
		return nil, err
	}

	res, err := admin.SaveAlert(ctx, s.Claims().UserID, args.Name, args.Version, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save alert: %w", err)
	}

	reconcileErr, err := managedResourceError(ctx, t.Runtime, runtime.ResourceKindAlert, res.Name)
	if err != nil {
		return nil, err
	}

	result := &SaveAlertResult{
		Name:           res.Name,
		Version:        res.Version,
		URL:            res.URL,
		ReconcileError: reconcileErr,
	}
	if prev != nil {
		result.Changes = diffFields(alertFields(prev.Options), alertFields(opts))
	}
	return result, nil
}

// applyArgs applies the arguments a call sent to an alert's options, which are empty when creating one,
// and validates the result. The query is the MetricsViewAggregation the alert dialog saves:
// the measure, split by the dimension, with the criteria as its having clause.
func (t *SaveAlert) applyArgs(ctx context.Context, args *SaveAlertArgs, base *drivers.AlertOptions, create bool) (*drivers.AlertOptions, error) {
	sent := &args.sent
	opts := cloneAlertOptions(base)

	if sent.has("display_name") {
		if args.DisplayName == "" {
			return nil, errors.New("invalid input: display_name can't be empty")
		}
		opts.DisplayName = args.DisplayName
	}
	if opts.DisplayName == "" {
		return nil, errors.New("invalid input: display_name is required")
	}

	// The data of the alert: a formula over a MetricsViewAggregation query, or an AI condition the ai resolver judges.
	// Like in the alert dialog, switching between them keeps the data context (explore, measure, dimension, filters and time range).
	if sent.hasAny("explore", "measure", "dimension", "time_range", "time_zone", "criteria", "criteria_operator", "comparison", "where", "condition") {
		if sent.has("condition") && sent.has("criteria") {
			return nil, errors.New("invalid input: pass either criteria, for a formula alert, or a condition, for an AI alert, not both")
		}
		isAI := opts.Resolver == "ai"
		switch {
		case sent.has("condition") || (isAI && !sent.has("criteria")):
			if err := t.applyAIConditionArgs(ctx, args, opts, create); err != nil {
				return nil, err
			}
		case isAI || opts.QueryName == "MetricsViewAggregation":
			if isAI {
				if err := t.queryFromAIProperties(ctx, opts); err != nil {
					return nil, err
				}
			}
			if err := t.applyQueryArgs(ctx, args, opts, create); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("the data of this alert was saved in a form this tool can't change, so only its name, schedule, notifications and recipients can be changed here: change the rest in the alert dialog")
		}
	}

	if sent.has("schedule") {
		if sent.null("schedule") || args.Schedule == nil {
			opts.RefreshCron, opts.RefreshTimeZone = "", ""
		} else {
			if err := checkSchedule(args.Schedule); err != nil {
				return nil, err
			}
			opts.RefreshCron, opts.RefreshTimeZone = args.Schedule.Cron, args.Schedule.TimeZone
		}
	}

	if sent.hasAny("renotify", "renotify_after_seconds") {
		mode := args.Renotify
		if mode == "" {
			mode = "after_interval"
		}
		switch mode {
		case "never":
			opts.Renotify, opts.RenotifyAfterSeconds = false, 0
		case "every_check":
			opts.Renotify, opts.RenotifyAfterSeconds = true, 0
		case "after_interval":
			if sent.has("renotify_after_seconds") {
				opts.RenotifyAfterSeconds = args.RenotifyAfterSeconds
			}
			if opts.RenotifyAfterSeconds == 0 {
				return nil, errors.New("invalid input: renotify 'after_interval' needs renotify_after_seconds")
			}
			opts.Renotify = true
		default:
			return nil, fmt.Errorf("invalid renotify %q: use 'every_check', 'after_interval' or 'never'", args.Renotify)
		}
	}

	var r Recipients
	if args.Recipients != nil {
		r = *args.Recipients
	}
	patchRecipients(sent, r, &opts.EmailRecipients, &opts.SlackUsers, &opts.SlackChannels, &opts.SlackWebhooks)
	if create {
		selfIfNoRecipients(ctx, &opts.EmailRecipients, opts.SlackUsers, opts.SlackChannels)
	} else if len(opts.EmailRecipients)+len(opts.SlackUsers)+len(opts.SlackChannels)+len(opts.SlackWebhooks) == 0 {
		return nil, errors.New("invalid input: an alert needs at least one recipient")
	}

	return opts, nil
}

// applyQueryArgs applies the data arguments a call sent to the alert's MetricsViewAggregation query.
// It keeps what the arguments don't change, like a time range or comparison the dialog saved in a form the arguments can't express,
// and it clears the dashboard state the alert opens with when the query changes, so the alert doesn't open on a view of other data.
func (t *SaveAlert) applyQueryArgs(ctx context.Context, args *SaveAlertArgs, opts *drivers.AlertOptions, create bool) error {
	sent := &args.sent
	q, err := parseAggregation(opts.QueryArgsJSON)
	if err != nil {
		return err
	}

	exploreName := exploreFromPath(opts.WebOpenPath)
	if sent.has("explore") {
		exploreName = args.Explore
	}
	exploreName, explore, err := validExplore(ctx, t.Runtime, exploreName)
	if err != nil {
		return err
	}
	if sent.has("explore") {
		q.MetricsView = explore.MetricsView
		opts.MetricsViewName = explore.MetricsView
		opts.WebOpenPath = "/explore/" + exploreName
	}

	if sent.has("time_range") {
		if args.TimeRange == nil || *args.TimeRange == "" {
			return errors.New("invalid input: a formula alert needs a time_range")
		}
		if err := checkTimeRange(*args.TimeRange); err != nil {
			return err
		}
		tz := opts.RefreshTimeZone
		if q.TimeRange != nil && q.TimeRange.TimeZone != "" {
			tz = q.TimeRange.TimeZone
		}
		q.TimeRange = &runtimev1.TimeRange{Expression: *args.TimeRange, TimeZone: tz}
		// Keep the kind of comparison, or the previous period for a comparison saved in a form the comparison argument can't express.
		if q.ComparisonTimeRange != nil {
			q.ComparisonTimeRange, err = comparisonFor(q.TimeRange, comparisonKindOf(q.ComparisonTimeRange))
			if err != nil {
				return err
			}
		}
	}
	if q.TimeRange == nil {
		return errors.New("invalid input: time_range is required")
	}
	if sent.has("time_zone") {
		q.TimeRange.TimeZone = ""
		if args.TimeZone != nil {
			if err := checkTimeZone(*args.TimeZone); err != nil {
				return err
			}
			q.TimeRange.TimeZone = *args.TimeZone
		}
		if q.ComparisonTimeRange != nil {
			q.ComparisonTimeRange.TimeZone = q.TimeRange.TimeZone
		}
	}

	// The measure and the criteria go together: the criteria compare the measure or measures computed from it.
	var baseMeasure *runtimev1.MetricsViewAggregationMeasure
	if len(q.Measures) > 0 {
		baseMeasure = q.Measures[0]
	}
	if sent.hasAny("measure", "criteria", "criteria_operator", "comparison") {
		criteria, joiner := args.Criteria, args.CriteriaOperator
		if !sent.has("criteria_operator") && q.Having.GetCond().GetOp() == runtimev1.Operation_OPERATION_OR {
			joiner = "or"
		}
		if !sent.has("criteria") && baseMeasure != nil {
			current, _, ok := criteriaFromHaving(q.Having, baseMeasure.Name)
			if !ok {
				return errors.New("the criteria of this alert were saved in a form this tool can't read: pass the criteria again")
			}
			criteria = current
		}
		measure := args.Measure
		if !sent.has("measure") && baseMeasure != nil {
			measure = baseMeasure.Name
		}
		if measure == "" {
			return errors.New("invalid input: measure is required")
		}
		if len(criteria) == 0 {
			return errors.New("invalid input: at least one criterion is required")
		}

		having, comparison, percentOfTotal, err := criteriaHaving(measure, criteria, joiner)
		if err != nil {
			return err
		}
		// Keep the measure as saved if it's the same, since the dialog may have saved it with a definition of its own.
		measures := []*runtimev1.MetricsViewAggregationMeasure{{Name: measure}}
		if baseMeasure != nil && baseMeasure.Name == measure {
			measures[0] = baseMeasure
		}
		if comparison {
			measures = append(measures,
				&runtimev1.MetricsViewAggregationMeasure{Name: measure + "_delta", Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{Measure: measure}}},
				&runtimev1.MetricsViewAggregationMeasure{Name: measure + "_delta_perc", Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonRatio{ComparisonRatio: &runtimev1.MetricsViewAggregationMeasureComputeComparisonRatio{Measure: measure}}},
			)
			// Keep a comparison the alert already has unless the call sends one.
			if q.ComparisonTimeRange == nil || sent.has("comparison") {
				q.ComparisonTimeRange, err = comparisonFor(q.TimeRange, args.Comparison)
				if err != nil {
					return err
				}
			}
		} else {
			if sent.has("comparison") && args.Comparison != "" {
				return errors.New("invalid input: comparison only applies to criteria of type 'absolute_change' or 'percent_change'")
			}
			q.ComparisonTimeRange = nil
		}
		if percentOfTotal {
			measures = append(measures, &runtimev1.MetricsViewAggregationMeasure{Name: measure + "_percent_of_total", Compute: &runtimev1.MetricsViewAggregationMeasure_PercentOfTotal{PercentOfTotal: &runtimev1.MetricsViewAggregationMeasureComputePercentOfTotal{Measure: measure}}})
		}
		q.Measures = measures
		q.Having = having
	}
	if len(q.Measures) == 0 {
		return errors.New("invalid input: measure is required")
	}

	if sent.has("dimension") {
		q.Dimensions = nil
		if args.Dimension != nil && *args.Dimension != "" {
			q.Dimensions = []*runtimev1.MetricsViewAggregationDimension{{Name: *args.Dimension}}
		}
	}
	if sent.has("where") {
		q.Where, err = replaceFilters(q.Where, args.Where)
		if err != nil {
			return err
		}
	}

	var dims []string
	for _, d := range q.Dimensions {
		dims = append(dims, d.Name)
	}
	err = checkExploreFields(explore, exploreName, []string{q.Measures[0].Name}, append(dims, filterDimensions(args.Where)...))
	if err != nil {
		return err
	}
	if create && args.Schedule != nil && !sent.has("time_zone") {
		q.TimeRange.TimeZone = args.Schedule.TimeZone
		if q.ComparisonTimeRange != nil {
			q.ComparisonTimeRange.TimeZone = args.Schedule.TimeZone
		}
	}

	queryArgs, err := aggregationQueryArgs(q)
	if err != nil {
		return err
	}
	if create || !sameJSON(queryArgs, opts.QueryArgsJSON) {
		if !create {
			opts.WebOpenState = ""
		}
		opts.QueryArgsJSON = queryArgs
	}
	return nil
}

// applyAIConditionArgs applies the data arguments a call sent to an AI alert, whose condition the ai resolver judges, as the alert dialog saves it:
// the condition as the resolver's prompt, the explore, the measure and dimension that narrow what the AI looks at, filters and a time range.
// A formula alert that becomes an AI alert keeps its data context and drops its query, and an AI alert keeps the properties the arguments don't set.
func (t *SaveAlert) applyAIConditionArgs(ctx context.Context, args *SaveAlertArgs, opts *drivers.AlertOptions, create bool) error {
	sent := &args.sent
	if sent.has("comparison") {
		return errors.New("invalid input: comparison only applies to formula alerts, with criteria of type 'absolute_change' or 'percent_change'")
	}
	wasAI := opts.Resolver == "ai"
	before := opts.ResolverProperties
	props := map[string]any{}
	if wasAI {
		props = maps.Clone(opts.ResolverProperties)
		if props == nil {
			props = map[string]any{}
		}
	} else if !create && opts.QueryArgsJSON != "" {
		var rangeErr, err error
		props, rangeErr, err = aiPropertiesFromQuery(opts)
		if err != nil {
			return err
		}
		if rangeErr != nil && !sent.has("time_range") {
			return rangeErr
		}
	}

	if sent.has("condition") {
		condition := strings.TrimSpace(args.Condition)
		if condition == "" {
			return errors.New("invalid input: the condition of an AI alert can't be empty")
		}
		props["prompt"] = condition
	}
	if prompt, _ := props["prompt"].(string); prompt == "" {
		return errors.New("invalid input: an AI alert needs a condition")
	}

	exploreName, _ := props["explore"].(string)
	if exploreName == "" {
		exploreName = exploreFromPath(opts.WebOpenPath)
	}
	if sent.has("explore") {
		exploreName = args.Explore
	}
	exploreName, explore, err := validExplore(ctx, t.Runtime, exploreName)
	if err != nil {
		return err
	}
	props["explore"] = exploreName
	opts.MetricsViewName = explore.MetricsView
	opts.WebOpenPath = "/explore/" + exploreName

	if sent.has("measure") {
		setFirstOfList(props, "measures", args.Measure)
	}
	if sent.has("dimension") {
		dim := ""
		if args.Dimension != nil {
			dim = *args.Dimension
		}
		setFirstOfList(props, "dimensions", dim)
	}
	if sent.has("where") {
		where, err := whereFromProperties(props)
		if err != nil {
			return err
		}
		where, err = replaceFilters(where, args.Where)
		if err != nil {
			return err
		}
		if err := setWhereProperty(props, where); err != nil {
			return err
		}
	}
	if sent.has("time_range") {
		delete(props, "time_range")
		if args.TimeRange != nil && *args.TimeRange != "" {
			if err := checkTimeRange(*args.TimeRange); err != nil {
				return err
			}
			props["time_range"] = map[string]any{"expression": *args.TimeRange}
		}
	} else if create {
		// Like the dialog, write the explore's default time range explicitly, so what the user sees is what is judged.
		if tr := explore.GetDefaultPreset().GetTimeRange(); tr != "" && tr != "inf" {
			props["time_range"] = map[string]any{"expression": tr}
		}
	}
	if sent.has("time_zone") {
		delete(props, "time_zone")
		if args.TimeZone != nil && *args.TimeZone != "" {
			if err := checkTimeZone(*args.TimeZone); err != nil {
				return err
			}
			props["time_zone"] = *args.TimeZone
		}
	}
	// Like the dialog, a time range is written with the time zone it's evaluated in.
	if props["time_range"] != nil && props["time_zone"] == nil {
		tz := opts.RefreshTimeZone
		if args.Schedule != nil && args.Schedule.TimeZone != "" {
			tz = args.Schedule.TimeZone
		}
		if tz == "" {
			tz = "UTC"
		}
		props["time_zone"] = tz
	}

	var measures, dims []string
	for _, k := range []string{"measures", "dimensions"} {
		list, _ := normalizeJSON(props[k]).([]any)
		for _, v := range list {
			if name, ok := v.(string); ok {
				if k == "measures" {
					measures = append(measures, name)
				} else {
					dims = append(dims, name)
				}
			}
		}
	}
	if err := checkExploreFields(explore, exploreName, measures, append(dims, filterDimensions(args.Where)...)); err != nil {
		return err
	}

	if !create && (!wasAI || !reflect.DeepEqual(normalizeJSON(props), normalizeJSON(before))) {
		opts.WebOpenState = ""
	}
	opts.Resolver = "ai"
	opts.ResolverProperties = props
	opts.QueryName, opts.QueryArgsJSON = "", ""
	return nil
}

// setFirstOfList sets the first item of a list property, as the dialog does with an AI alert's measures and dimensions:
// the list is kept while its first item doesn't change, and an empty value removes it.
func setFirstOfList(props map[string]any, key, first string) {
	if first == "" {
		delete(props, key)
		return
	}
	if list, ok := normalizeJSON(props[key]).([]any); ok && len(list) > 0 && list[0] == first {
		return
	}
	props[key] = []any{first}
}

// aiPropertiesFromQuery returns the data context of a formula alert as the properties of an AI alert,
// as the dialog keeps its data tab when the condition changes from criteria to AI.
// rangeErr explains why the query's time range can't be carried over, if so; then the caller needs a time range of its own.
func aiPropertiesFromQuery(opts *drivers.AlertOptions) (props map[string]any, rangeErr, err error) {
	q, err := parseAggregation(opts.QueryArgsJSON)
	if err != nil {
		return nil, nil, err
	}
	props = map[string]any{"explore": exploreFromPath(opts.WebOpenPath)}
	if len(q.Measures) > 0 {
		props["measures"] = []any{q.Measures[0].Name}
	}
	if len(q.Dimensions) > 0 {
		props["dimensions"] = []any{q.Dimensions[0].Name}
	}
	if len(q.Where.GetCond().GetExprs()) > 0 {
		props["where"], err = metricsview.NewExpressionFromProto(q.Where).AsMap()
		if err != nil {
			return nil, nil, err
		}
	}
	tr, rangeErr := aiTimeRangeFromQuery(q.TimeRange)
	if tr != nil {
		props["time_range"] = tr
	}
	// Keep the time zone even if the range can't be carried over, so a range the call passes instead is judged in it too.
	if tz := q.TimeRange.GetTimeZone(); tz != "" {
		props["time_zone"] = tz
	}
	return props, rangeErr, nil
}

// aiTimeRangeFromQuery converts the time range of a formula alert's query into an ai resolver time range, as the dialog writes it:
// a rill time expression, which includes the dashboards' preset names, or a start and end. It returns nil for all time.
// It returns an error for a range the dialog's presets can't express, like an ISO duration with an offset, instead of changing what is judged.
func aiTimeRangeFromQuery(tr *runtimev1.TimeRange) (map[string]any, error) {
	switch {
	case tr == nil:
		return nil, nil
	case tr.Expression != "":
		return map[string]any{"expression": tr.Expression}, nil
	case tr.Start != nil && tr.End != nil:
		return map[string]any{"start": tr.Start.AsTime().Format(time.RFC3339Nano), "end": tr.End.AsTime().Format(time.RFC3339Nano)}, nil
	case tr.IsoDuration == "inf":
		return nil, nil
	case tr.IsoDuration != "":
		// The dialog's complete-period presets are saved as a duration rounded to a grain, and written by their names.
		for name, preset := range previousCompletePresets {
			if tr.IsoDuration == preset.duration && tr.RoundToGrain == preset.grain && tr.IsoOffset == "" {
				return map[string]any{"expression": name}, nil
			}
		}
		if tr.IsoOffset == "" && tr.RoundToGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			// The dialog writes a duration preset as its name, which rill time parses.
			return map[string]any{"expression": tr.IsoDuration}, nil
		}
	}
	return nil, fmt.Errorf("the alert's time range (%v) can't be written as the time range of an AI alert without changing it: pass a time_range", timeRangeValue(tr))
}

// previousCompletePresets are the dialog's complete-period presets, by name, with the duration and grain their queries are saved with
// (PreviousCompleteRangeMap in the UI).
var previousCompletePresets = map[string]struct {
	duration string
	grain    runtimev1.TimeGrain
}{
	"rill-PDC": {"P1D", runtimev1.TimeGrain_TIME_GRAIN_DAY},
	"rill-PWC": {"P1W", runtimev1.TimeGrain_TIME_GRAIN_WEEK},
	"rill-PMC": {"P1M", runtimev1.TimeGrain_TIME_GRAIN_MONTH},
	"rill-PQC": {"P3M", runtimev1.TimeGrain_TIME_GRAIN_QUARTER},
	"rill-PYC": {"P1Y", runtimev1.TimeGrain_TIME_GRAIN_YEAR},
}

// queryFromAIProperties turns an AI alert into a formula alert's MetricsViewAggregation query with its data context,
// as the dialog keeps its data tab when the condition changes from AI to criteria, and drops its AI properties.
// The criteria are applied afterwards.
func (t *SaveAlert) queryFromAIProperties(ctx context.Context, opts *drivers.AlertOptions) error {
	props := opts.ResolverProperties
	exploreName, _ := props["explore"].(string)
	if exploreName == "" {
		exploreName = exploreFromPath(opts.WebOpenPath)
	}
	exploreName, explore, err := validExplore(ctx, t.Runtime, exploreName)
	if err != nil {
		return err
	}
	q := &runtimev1.MetricsViewAggregationRequest{MetricsView: explore.MetricsView}
	if list, ok := normalizeJSON(props["measures"]).([]any); ok && len(list) > 0 {
		if name, ok := list[0].(string); ok {
			q.Measures = []*runtimev1.MetricsViewAggregationMeasure{{Name: name}}
		}
	}
	if list, ok := normalizeJSON(props["dimensions"]).([]any); ok && len(list) > 0 {
		if name, ok := list[0].(string); ok {
			q.Dimensions = []*runtimev1.MetricsViewAggregationDimension{{Name: name}}
		}
	}
	if where, ok := props["where"].(map[string]any); ok {
		expr := &metricsview.Expression{}
		data, err := json.Marshal(where)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, expr); err != nil {
			return fmt.Errorf("failed to read the AI alert's filters: %w", err)
		}
		q.Where = metricsview.ExpressionToProto(expr)
	}
	if tr, ok := props["time_range"].(map[string]any); ok {
		tz, _ := props["time_zone"].(string)
		expr, _ := tr["expression"].(string)
		start, _ := tr["start"].(string)
		end, _ := tr["end"].(string)
		switch {
		case expr != "":
			q.TimeRange = &runtimev1.TimeRange{Expression: expr, TimeZone: tz}
		case start != "" && end != "":
			// A fixed range, as the dialog writes a custom one.
			s, err := time.Parse(time.RFC3339Nano, start)
			if err != nil {
				return fmt.Errorf("invalid start of the AI alert's time range: %w", err)
			}
			e, err := time.Parse(time.RFC3339Nano, end)
			if err != nil {
				return fmt.Errorf("invalid end of the AI alert's time range: %w", err)
			}
			q.TimeRange = &runtimev1.TimeRange{Start: timestamppb.New(s), End: timestamppb.New(e), TimeZone: tz}
		}
	}

	opts.MetricsViewName = explore.MetricsView
	opts.WebOpenPath = "/explore/" + exploreName
	opts.QueryName = "MetricsViewAggregation"
	opts.QueryArgsJSON, err = aggregationQueryArgs(q)
	if err != nil {
		return err
	}
	opts.Resolver, opts.ResolverProperties = "", nil
	return nil
}

// criteriaHaving converts criteria on a measure into the having clause the alert dialog saves, and reports the computed measures it needs.
func criteriaHaving(measure string, criteria []AlertCriterion, joiner string) (*runtimev1.Expression, bool, bool, error) {
	var op runtimev1.Operation
	switch joiner {
	case "", "and":
		op = runtimev1.Operation_OPERATION_AND
	case "or":
		op = runtimev1.Operation_OPERATION_OR
	default:
		return nil, false, false, fmt.Errorf("invalid criteria_operator %q: use 'and' or 'or'", joiner)
	}

	var comparison, percentOfTotal bool
	var exprs []*runtimev1.Expression
	for _, c := range criteria {
		cop, ok := criterionOperators[c.Operator]
		if !ok {
			return nil, false, false, fmt.Errorf("invalid criterion operator %q: use one of 'gt', 'gte', 'lt', 'lte', 'eq' or 'neq'", c.Operator)
		}
		name, value := measure, c.Value
		switch c.Type {
		case "", "value":
		case "absolute_change":
			name += "_delta"
			comparison = true
		case "percent_change":
			name += "_delta_perc"
			value /= 100
			comparison = true
		case "percent_of_total":
			name += "_percent_of_total"
			value /= 100
			percentOfTotal = true
		default:
			return nil, false, false, fmt.Errorf("invalid criterion type %q: use 'value', 'absolute_change', 'percent_change' or 'percent_of_total'", c.Type)
		}
		exprs = append(exprs, &runtimev1.Expression{Expression: &runtimev1.Expression_Cond{Cond: &runtimev1.Condition{
			Op:    cop,
			Exprs: []*runtimev1.Expression{expressionpb.Identifier(name), expressionpb.Number(value)},
		}}})
	}
	return &runtimev1.Expression{Expression: &runtimev1.Expression_Cond{Cond: &runtimev1.Condition{Op: op, Exprs: exprs}}}, comparison, percentOfTotal, nil
}

// criteriaFromHaving decodes the having clause of an alert into criteria on its measure, and the operator that combines them.
// It returns false if the clause isn't in the form the alert dialog saves.
func criteriaFromHaving(having *runtimev1.Expression, measure string) ([]AlertCriterion, string, bool) {
	c := having.GetCond()
	if c == nil {
		return nil, "", false
	}
	joiner := "and"
	switch c.Op {
	case runtimev1.Operation_OPERATION_AND:
	case runtimev1.Operation_OPERATION_OR:
		joiner = "or"
	default:
		return nil, "", false
	}
	var criteria []AlertCriterion
	for _, e := range c.Exprs {
		ec := e.GetCond()
		if ec == nil || len(ec.Exprs) != 2 {
			return nil, "", false
		}
		var operator string
		for k, v := range criterionOperators {
			if v == ec.Op {
				operator = k
			}
		}
		val, ok := ec.Exprs[1].GetVal().AsInterface().(float64)
		if operator == "" || !ok {
			return nil, "", false
		}
		criterion := AlertCriterion{Operator: operator}
		switch ident := ec.Exprs[0].GetIdent(); ident {
		case measure:
			criterion.Type, criterion.Value = "value", val
		case measure + "_delta":
			criterion.Type, criterion.Value = "absolute_change", val
		case measure + "_delta_perc":
			criterion.Type, criterion.Value = "percent_change", val*100
		case measure + "_percent_of_total":
			criterion.Type, criterion.Value = "percent_of_total", val*100
		default:
			return nil, "", false
		}
		criteria = append(criteria, criterion)
	}
	return criteria, joiner, true
}

// alertFields flattens an alert's options into named fields, to report what an edit changed.
func alertFields(o *drivers.AlertOptions) map[string]any {
	m := map[string]any{}
	setField(m, "display_name", o.DisplayName)
	setField(m, "explore", exploreFromPath(o.WebOpenPath))
	setField(m, "schedule.cron", o.RefreshCron)
	setField(m, "schedule.time_zone", o.RefreshTimeZone)
	setField(m, "renotify", renotifyValue(o))
	setField(m, "renotify_after_seconds", o.RenotifyAfterSeconds)
	recipientsFields(m, o.EmailRecipients, o.SlackUsers, o.SlackChannels, o.SlackWebhooks)
	setJSONFields(m, "query", o.QueryArgsJSON)
	setField(m, "resolver", o.Resolver)
	setMapFields(m, "resolver_properties", o.ResolverProperties)
	setField(m, "dashboard_state", o.WebOpenState)
	return m
}

// renotifyValue presents an alert's renotify setting as the renotify argument.
func renotifyValue(o *drivers.AlertOptions) string {
	switch {
	case !o.Renotify:
		return "never"
	case o.RenotifyAfterSeconds == 0:
		return "every_check"
	default:
		return "after_interval"
	}
}

// aggregationQueryArgs serializes a MetricsViewAggregation query as the dialogs save it.
// Unlike protojson, the dialogs always include the list of dimensions, even if empty, and the alert dialog reads it back when editing.
func aggregationQueryArgs(query *runtimev1.MetricsViewAggregationRequest) (string, error) {
	data, err := protojson.Marshal(query)
	if err != nil {
		return "", err
	}
	var args map[string]any
	if err := json.Unmarshal(data, &args); err != nil {
		return "", err
	}
	if _, ok := args["dimensions"]; !ok {
		args["dimensions"] = []any{}
	}
	data, err = json.Marshal(args)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var criterionOperators = map[string]runtimev1.Operation{
	"gt":  runtimev1.Operation_OPERATION_GT,
	"gte": runtimev1.Operation_OPERATION_GTE,
	"lt":  runtimev1.Operation_OPERATION_LT,
	"lte": runtimev1.Operation_OPERATION_LTE,
	"eq":  runtimev1.Operation_OPERATION_EQ,
	"neq": runtimev1.Operation_OPERATION_NEQ,
}

// managedResourceError returns the reconcile error of a saved alert or report, if it has one already.
// The admin service returns once the project parser has created the resource, so a parse error has failed the save already.
func managedResourceError(ctx context.Context, rt *runtime.Runtime, kind, name string) (string, error) {
	ctrl, err := rt.Controller(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return "", err
	}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: kind, Name: name}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return "", nil
		}
		return "", err
	}
	return r.Meta.ReconcileError, nil
}
