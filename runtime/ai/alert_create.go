package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const CreateAlertName = "create_alert"

// CreateAlert creates an alert for the current user through the admin service, as the alert dialog creates it:
// the alert is owned by the user, listed in their alerts, and they can edit or delete it in the dialog.
type CreateAlert struct {
	Runtime *runtime.Runtime
}

var _ Tool[*CreateAlertArgs, *CreatedResource] = (*CreateAlert)(nil)

type CreateAlertArgs struct {
	DisplayName          string            `json:"display_name" jsonschema:"Display name of the alert."`
	Explore              string            `json:"explore" jsonschema:"Name of the explore dashboard the alert watches. Its metrics view provides the measure and dimensions, and the alert opens on it."`
	Condition            string            `json:"condition,omitempty" jsonschema:"For an AI alert: the condition the AI analyst checks on every check, in the user's words, for example 'sales drop unusually compared to recent weeks'. Don't pass it together with criteria."`
	Measure              string            `json:"measure,omitempty" jsonschema:"Name of the measure the alert checks. Required for criteria; for an AI alert, it optionally narrows what the AI looks at."`
	Dimension            string            `json:"dimension,omitempty" jsonschema:"Optional dimension to split the measure by, so the criteria are checked for each of its values (for an AI alert, it narrows what the AI looks at). Required for criteria of type 'percent_of_total'."`
	Criteria             []AlertCriterion  `json:"criteria,omitempty" jsonschema:"For a formula alert: the conditions on the measure that trigger the alert."`
	CriteriaOperator     string            `json:"criteria_operator,omitempty" jsonschema:"For a formula alert: how the criteria combine, 'and' (default) or 'or'."`
	Comparison           string            `json:"comparison,omitempty" jsonschema:"For criteria of type 'absolute_change' or 'percent_change': the period the time range is compared with, 'previous_period' (default, the period of the same length right before), 'previous_day', 'previous_week', 'previous_month', 'previous_quarter' or 'previous_year'."`
	TimeRange            string            `json:"time_range,omitempty" jsonschema:"Time range of the data on every check, as a Rill time expression relative to the latest data, for example '7D as of latest/D' for the last 7 complete days or '1M as of latest/M' for the last complete month. Required for criteria; an AI alert defaults to the explore's default time range."`
	TimeZone             string            `json:"time_zone,omitempty" jsonschema:"IANA time zone the time range is evaluated in, for example 'Europe/Madrid'. Defaults to the schedule's time zone, or UTC."`
	Where                []DimensionFilter `json:"where,omitempty" jsonschema:"Optional filters on dimensions of the explore."`
	Schedule             *Schedule         `json:"schedule,omitempty" jsonschema:"When to check the alert. A formula alert without a schedule is checked whenever the data refreshes. An AI alert needs a schedule."`
	Renotify             string            `json:"renotify,omitempty" jsonschema:"While the alert keeps triggering: 'every_check' (default) notifies on every check; 'after_interval' notifies again after renotify_after_seconds; 'never' notifies only when it starts triggering."`
	RenotifyAfterSeconds uint32            `json:"renotify_after_seconds,omitempty" jsonschema:"With renotify 'after_interval', seconds to wait before notifying again, for example 86400 for daily."`
	Recipients           *Recipients       `json:"recipients,omitempty" jsonschema:"Who to notify. Defaults to the user."`
}

// AlertCriterion is a condition on the alert's measure.
type AlertCriterion struct {
	Type     string  `json:"type,omitempty" jsonschema:"What is compared: 'value' (default) the measure itself; 'absolute_change' or 'percent_change' its change from the period of the comparison argument; 'percent_of_total' each value of the dimension argument's share of the total, from 0 to 100."`
	Operator string  `json:"operator" jsonschema:"One of 'gt', 'gte', 'lt', 'lte', 'eq' or 'neq'."`
	Value    float64 `json:"value" jsonschema:"Threshold. For percent types, in percent: -10 for a 10% drop."`
}

func (t *CreateAlert) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  CreateAlertName,
		Title: "Create alert",
		Description: "Creates an alert for the current user on the data of an explore dashboard. " +
			"Use it when the user asks to be alerted, notified or warned about their data, for example 'alert me if revenue drops below 30k' or 'tell me if sales drop unusually'. " +
			"An alert has a formula condition (criteria: a threshold on a measure or its change, checked exactly and at no AI cost) or an AI condition (condition: a description the AI analyst checks). " +
			"Prefer criteria whenever the condition is a threshold on a measure; use an AI condition only for what a threshold can't express, like 'unusual', 'unexpected' or patterns across dimensions. " +
			"Each check of an AI alert runs the AI analyst once, so it's only checked on a schedule: tell the user in the preview how often it's checked. " +
			"The alert belongs to the user, and they can edit or delete it on its page. Only create a file in the project's code instead when the user explicitly asks for that. " +
			confirmBeforeCreating +
			"This tool can't change alerts: to change one, give the user the URL of its page from list_alerts, where they can edit it. " +
			"Give the user the alert's URL from the response.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(true), // It notifies people
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating alert...",
			"openai/toolInvocation/invoked":  "Created alert",
		},
	}
}

func (t *CreateAlert) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *CreateAlert) Handler(ctx context.Context, args *CreateAlertArgs) (*CreatedResource, error) {
	if args.DisplayName == "" {
		return nil, errors.New("invalid input: display_name is required")
	}
	explore, err := validExplore(ctx, t.Runtime, args.Explore)
	if err != nil {
		return nil, err
	}
	where, whereDims, err := filtersExpression(args.Where)
	if err != nil {
		return nil, err
	}
	if err := checkTimeZone(args.TimeZone); err != nil {
		return nil, err
	}

	opts := &adminv1.AlertOptions{
		DisplayName:     args.DisplayName,
		MetricsViewName: explore.MetricsView,
		WebOpenPath:     "/explore/" + url.PathEscape(args.Explore),
	}
	if args.Recipients != nil {
		opts.EmailRecipients = args.Recipients.Emails
		opts.SlackUsers = args.Recipients.SlackUsers
	}

	// A time range is evaluated in the schedule's time zone unless the call passes another, as in the dialog.
	tz := args.TimeZone
	if args.Schedule != nil {
		if err := checkSchedule(args.Schedule); err != nil {
			return nil, err
		}
		opts.RefreshCron, opts.RefreshTimeZone = args.Schedule.Cron, args.Schedule.TimeZone
		if tz == "" {
			tz = args.Schedule.TimeZone
		}
	}

	switch args.Renotify {
	case "", "every_check":
		opts.Renotify = true
	case "after_interval":
		if args.RenotifyAfterSeconds == 0 {
			return nil, errors.New("invalid input: renotify 'after_interval' needs renotify_after_seconds")
		}
		opts.Renotify, opts.RenotifyAfterSeconds = true, args.RenotifyAfterSeconds
	case "never":
	default:
		return nil, fmt.Errorf("invalid renotify %q: use 'every_check', 'after_interval' or 'never'", args.Renotify)
	}

	if args.Condition != "" {
		err = t.setAICondition(args, explore, where, whereDims, tz, opts)
	} else {
		err = t.setCriteria(args, explore, where, whereDims, tz, opts)
	}
	if err != nil {
		return nil, err
	}

	s := GetSession(ctx)
	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()
	name, url, err := admin.CreateAlertForUser(ctx, s.Claims().UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}
	return createdResource(ctx, t.Runtime, runtime.ResourceKindAlert, name, url)
}

// setAICondition sets the alert's data to an AI condition the ai resolver checks, as the alert dialog saves it:
// the condition as the resolver's prompt, the explore, the measure and dimension that narrow what the AI looks at, filters and a time range.
func (t *CreateAlert) setAICondition(args *CreateAlertArgs, explore *runtimev1.ExploreSpec, where *runtimev1.Expression, whereDims []string, tz string, opts *adminv1.AlertOptions) error {
	if len(args.Criteria) > 0 || args.CriteriaOperator != "" || args.Comparison != "" {
		return errors.New("invalid input: pass either criteria, for a formula alert, or a condition, for an AI alert, not both")
	}
	if args.Schedule == nil {
		return errors.New("invalid input: an AI alert needs a schedule, since every check runs the AI analyst")
	}
	if err := parser.ValidateAIAlertSchedule(&parser.ScheduleYAML{Cron: args.Schedule.Cron, TimeZone: args.Schedule.TimeZone}); err != nil {
		return err
	}

	var measures, dims []string
	props := map[string]any{
		"prompt":  strings.TrimSpace(args.Condition),
		"explore": args.Explore,
	}
	if args.Measure != "" {
		measures = []string{args.Measure}
		props["measures"] = []any{args.Measure}
	}
	if args.Dimension != "" {
		dims = []string{args.Dimension}
		props["dimensions"] = []any{args.Dimension}
	}
	if err := checkExploreFields(explore, args.Explore, measures, append(dims, whereDims...)); err != nil {
		return err
	}
	if where != nil {
		w, err := metricsview.NewExpressionFromProto(where).AsMap()
		if err != nil {
			return err
		}
		props["where"] = w
	}
	// Like the dialog, write the explore's default time range explicitly, so what the user sees is what is checked.
	timeRange := args.TimeRange
	if timeRange == "" {
		if tr := explore.GetDefaultPreset().GetTimeRange(); tr != "inf" {
			timeRange = tr
		}
	}
	if timeRange != "" {
		if err := checkTimeRange(timeRange); err != nil {
			return err
		}
		if tz == "" {
			tz = "UTC"
		}
		props["time_range"] = map[string]any{"expression": timeRange}
		props["time_zone"] = tz
	}

	var err error
	opts.Resolver = "ai"
	opts.ResolverProperties, err = structpb.NewStruct(props)
	return err
}

// setCriteria sets the alert's data to the MetricsViewAggregation query the alert dialog saves:
// the measure, split by the dimension, with the criteria as its having clause, sorted by the measure.
func (t *CreateAlert) setCriteria(args *CreateAlertArgs, explore *runtimev1.ExploreSpec, where *runtimev1.Expression, whereDims []string, tz string, opts *adminv1.AlertOptions) error {
	if args.Measure == "" || len(args.Criteria) == 0 {
		return errors.New("invalid input: pass a measure and criteria, for a formula alert, or a condition, for an AI alert")
	}
	if args.TimeRange == "" {
		return errors.New("invalid input: a formula alert needs a time_range")
	}
	if err := checkTimeRange(args.TimeRange); err != nil {
		return err
	}
	var dims []string
	if args.Dimension != "" {
		dims = []string{args.Dimension}
	}
	if err := checkExploreFields(explore, args.Explore, []string{args.Measure}, append(dims, whereDims...)); err != nil {
		return err
	}

	having, comparison, percentOfTotal, err := criteriaHaving(args.Measure, args.Criteria, args.CriteriaOperator)
	if err != nil {
		return err
	}
	// Without a split, the total's share of itself is always 100%, so the alert could never trigger, or always would.
	if percentOfTotal && args.Dimension == "" {
		return errors.New("invalid input: criteria of type 'percent_of_total' need a dimension, whose values' shares of the total they check")
	}
	q := &runtimev1.MetricsViewAggregationRequest{
		MetricsView: explore.MetricsView,
		Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: args.Measure}},
		Where:       where,
		Having:      having,
		TimeRange:   &runtimev1.TimeRange{Expression: args.TimeRange, TimeZone: tz},
		Sort:        []*runtimev1.MetricsViewAggregationSort{{Name: args.Measure, Desc: true}},
	}
	for _, d := range dims {
		q.Dimensions = append(q.Dimensions, &runtimev1.MetricsViewAggregationDimension{Name: d})
	}
	if comparison {
		cmp, err := comparisonExpression(args.TimeRange, args.Comparison)
		if err != nil {
			return err
		}
		q.ComparisonTimeRange = &runtimev1.TimeRange{Expression: cmp, TimeZone: tz}
		q.Measures = append(q.Measures,
			&runtimev1.MetricsViewAggregationMeasure{Name: args.Measure + "_delta", Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{Measure: args.Measure}}},
			&runtimev1.MetricsViewAggregationMeasure{Name: args.Measure + "_delta_perc", Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonRatio{ComparisonRatio: &runtimev1.MetricsViewAggregationMeasureComputeComparisonRatio{Measure: args.Measure}}},
		)
	} else if args.Comparison != "" {
		return errors.New("invalid input: comparison only applies to criteria of type 'absolute_change' or 'percent_change'")
	}
	if percentOfTotal {
		q.Measures = append(q.Measures, &runtimev1.MetricsViewAggregationMeasure{Name: args.Measure + "_percent_of_total", Compute: &runtimev1.MetricsViewAggregationMeasure_PercentOfTotal{PercentOfTotal: &runtimev1.MetricsViewAggregationMeasureComputePercentOfTotal{Measure: args.Measure}}})
	}

	opts.QueryName = "MetricsViewAggregation"
	opts.QueryArgsJson, err = aggregationQueryArgs(q)
	return err
}

// criteriaHaving converts criteria on a measure into the having clause the alert dialog saves, and reports the computed measures it needs.
func criteriaHaving(measure string, criteria []AlertCriterion, joiner string) (having *runtimev1.Expression, comparison, percentOfTotal bool, err error) {
	var join func([]*runtimev1.Expression) *runtimev1.Expression
	switch joiner {
	case "", "and":
		join = expressionpb.And
	case "or":
		join = expressionpb.Or
	default:
		return nil, false, false, fmt.Errorf("invalid criteria_operator %q: use 'and' or 'or'", joiner)
	}

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
			if value < 0 || value > 100 {
				return nil, false, false, fmt.Errorf("invalid criterion value %v: a 'percent_of_total' threshold is a percent from 0 to 100", c.Value)
			}
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
	return join(exprs), comparison, percentOfTotal, nil
}

var criterionOperators = map[string]runtimev1.Operation{
	"gt":  runtimev1.Operation_OPERATION_GT,
	"gte": runtimev1.Operation_OPERATION_GTE,
	"lt":  runtimev1.Operation_OPERATION_LT,
	"lte": runtimev1.Operation_OPERATION_LTE,
	"eq":  runtimev1.Operation_OPERATION_EQ,
	"neq": runtimev1.Operation_OPERATION_NEQ,
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
