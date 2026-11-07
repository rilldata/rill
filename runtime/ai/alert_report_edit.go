package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// sentArgs records which arguments a tool call sent, so an edit only changes those:
// an argument the call omits keeps its stored value, and one it sends as null clears it.
// Keys of objects are recorded one level deep, as "recipients.emails".
type sentArgs struct {
	raw  json.RawMessage
	keys map[string]json.RawMessage
}

func (a *sentArgs) parse(data []byte) error {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return err
	}
	a.raw = append(json.RawMessage(nil), data...)
	a.keys = make(map[string]json.RawMessage, len(top))
	for k, v := range top {
		a.keys[k] = v
		var nested map[string]json.RawMessage
		if bytes.HasPrefix(bytes.TrimSpace(v), []byte("{")) && json.Unmarshal(v, &nested) == nil {
			for nk, nv := range nested {
				a.keys[k+"."+nk] = nv
			}
		}
	}
	return nil
}

// has reports whether the call sent the argument, null included.
func (a *sentArgs) has(key string) bool {
	_, ok := a.keys[key]
	return ok
}

// null reports whether the call sent the argument as null.
func (a *sentArgs) null(key string) bool {
	v, ok := a.keys[key]
	return ok && string(bytes.TrimSpace(v)) == "null"
}

// hasAny reports whether the call sent any of the arguments.
func (a *sentArgs) hasAny(keys ...string) bool {
	return slices.ContainsFunc(keys, a.has)
}

// ensure records the non-empty fields of args as sent, for arguments built in Go rather than unmarshaled from a tool call.
func (a *sentArgs) ensure(args any) error {
	if a.keys != nil {
		return nil
	}
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return a.parse(data)
}

// patchRecipients applies the recipients a call sent to stored ones: each kind of recipient it sends replaces the stored ones,
// and null clears them all, including Slack webhooks, which the arguments can't set but the dialogs or API may have.
func patchRecipients(sent *sentArgs, r Recipients, emails, slackUsers, slackChannels, slackWebhooks *[]string) {
	if sent.null("recipients") {
		*emails, *slackUsers, *slackChannels, *slackWebhooks = nil, nil, nil, nil
		return
	}
	if sent.has("recipients.emails") {
		*emails = r.Emails
	}
	if sent.has("recipients.slack_users") {
		*slackUsers = r.SlackUsers
	}
	if sent.has("recipients.slack_channels") {
		*slackChannels = r.SlackChannels
	}
}

// selfIfNoRecipients sets the user's own email as the recipient when there is none, as the dialogs default to.
func selfIfNoRecipients(ctx context.Context, emails *[]string, others ...[]string) {
	if len(*emails) > 0 || slices.ContainsFunc(others, func(s []string) bool { return len(s) > 0 }) {
		return
	}
	email, _ := GetSession(ctx).Claims().UserAttributes["email"].(string)
	*emails = []string{email}
}

// editConflictError is returned when an edit is based on a version that is no longer current.
func editConflictError(kind, name, tool string) error {
	return fmt.Errorf("%s %q changed since it was listed: call %s again and reapply the change to its current version", kind, name, tool)
}

// parseAggregation parses the query args of an alert or report saved as a MetricsViewAggregation query.
func parseAggregation(queryArgsJSON string) (*runtimev1.MetricsViewAggregationRequest, error) {
	q := &runtimev1.MetricsViewAggregationRequest{}
	if queryArgsJSON == "" {
		return q, nil
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(queryArgsJSON), q); err != nil {
		return nil, fmt.Errorf("failed to parse the saved query: %w", err)
	}
	return q, nil
}

// comparisonFor returns the comparison time range of a time range, as the dialogs save it for a comparison kind (mapSelectedComparisonTimeRangeToV1TimeRange in the UI):
// a Rill time expression offset by the kind, or for an ISO duration or a legacy expression, the duration with the kind's option as its ISO offset,
// except for the previous period, whose ISO offset is the duration itself.
// Unlike the dialogs, it also writes the time zone, which the query requires to be the time range's if set.
func comparisonFor(tr *runtimev1.TimeRange, kind string) (*runtimev1.TimeRange, error) {
	duration := tr.GetIsoDuration()
	switch {
	case tr == nil:
		return nil, errors.New("comparing with an earlier period needs a time_range")
	case tr.Expression != "":
		e, err := rilltime.Parse(tr.Expression, rilltime.ParseOptions{})
		if err != nil {
			return nil, fmt.Errorf("invalid time range %q: %w", tr.Expression, err)
		}
		legacy := strings.EqualFold(strings.TrimSpace(tr.Expression), "inf") || e.Interval != nil && (e.Interval.LegacyIso != nil || e.Interval.LegacyDax != nil)
		if !legacy {
			cmp, err := comparisonExpression(tr.Expression, kind)
			if err != nil {
				return nil, err
			}
			return &runtimev1.TimeRange{Expression: cmp, TimeZone: tr.TimeZone}, nil
		}
		// The dialogs compare a legacy expression as an ISO duration.
		duration = tr.Expression
	case tr.IsoDuration == "":
		return nil, errors.New("can't compare this time range with an earlier period: pass a time_range")
	}
	k, err := lookupComparisonKind(kind)
	if err != nil {
		return nil, err
	}
	offset := k.option
	if k.name == "previous_period" {
		offset = duration
	}
	return &runtimev1.TimeRange{IsoDuration: duration, IsoOffset: offset, TimeZone: tr.TimeZone}, nil
}

// comparisonKindOf returns the comparison kind of a comparison time range as the dialogs save it, or "" if it's none of them.
func comparisonKindOf(cmp *runtimev1.TimeRange) string {
	switch {
	case cmp == nil:
		return ""
	case cmp.Expression != "":
		return comparisonKindOfExpression(cmp.Expression)
	case cmp.IsoOffset != "" && cmp.IsoOffset == cmp.IsoDuration:
		return "previous_period"
	}
	for _, k := range comparisonKinds {
		if cmp.IsoOffset == k.option {
			return k.name
		}
	}
	return ""
}

// timeRangeValue presents a saved time range: its Rill time expression, or its fields if it was saved another way.
func timeRangeValue(tr *runtimev1.TimeRange) any {
	if tr == nil {
		return nil
	}
	if tr.Expression != "" && tr.IsoDuration == "" && tr.Start == nil {
		return tr.Expression
	}
	return protoValue(tr)
}

// protoValue converts a proto message to a JSON value.
func protoValue(m proto.Message) any {
	data, err := protojson.Marshal(m)
	if err != nil {
		return nil
	}
	var v any
	_ = json.Unmarshal(data, &v)
	return v
}

// filtersFromExpression decodes a where expression saved as an AND of IN or NIN conditions on dimensions,
// and of the LIKE or NLIKE conditions of the dashboard's Contains mode, into dimension filters.
// It returns false if the expression can't be expressed as dimension filters.
func filtersFromExpression(e *runtimev1.Expression) ([]DimensionFilter, bool) {
	filters, others := splitWhere(e)
	return filters, len(others) == 0
}

// splitWhere splits a where expression into the dimension filters the tools' arguments express
// and the other conditions of its AND, like the dashboards' measure filters, which an edit of the filters keeps.
// A where expression that isn't an AND is a single other condition.
func splitWhere(e *runtimev1.Expression) ([]DimensionFilter, []*runtimev1.Expression) {
	if e == nil {
		return nil, nil
	}
	conds := []*runtimev1.Expression{e}
	if c := e.GetCond(); c != nil && c.Op == runtimev1.Operation_OPERATION_AND {
		conds = c.Exprs
	}
	var filters []DimensionFilter
	var others []*runtimev1.Expression
	for _, ce := range conds {
		if f, ok := containsFilter(ce.GetCond()); ok {
			filters = append(filters, f)
		} else if f, ok := valuesFilter(ce.GetCond()); ok {
			filters = append(filters, f)
		} else {
			others = append(others, ce)
		}
	}
	return filters, others
}

// replaceFilters returns a where expression with the given dimension filters in place of the ones where has,
// keeping its conditions the filters can't express.
func replaceFilters(where *runtimev1.Expression, filters []DimensionFilter) (*runtimev1.Expression, error) {
	expr, err := filtersExpression(filters)
	if err != nil {
		return nil, err
	}
	// TODO: The tools don't write measure filters yet, so an edit keeps them untouched;
	// supporting them would need a filter argument with a measure, operator, value and dimension.
	_, others := splitWhere(where)
	if len(others) == 0 {
		return expr, nil
	}
	return expressionpb.And(append(expr.GetCond().GetExprs(), others...)), nil
}

// valuesFilter decodes an IN or NIN condition on a dimension with string values into a filter on those values.
func valuesFilter(c *runtimev1.Condition) (DimensionFilter, bool) {
	if c == nil || (c.Op != runtimev1.Operation_OPERATION_IN && c.Op != runtimev1.Operation_OPERATION_NIN) || len(c.Exprs) < 2 || c.Exprs[0].GetIdent() == "" {
		return DimensionFilter{}, false
	}
	f := DimensionFilter{Dimension: c.Exprs[0].GetIdent(), Exclude: c.Op == runtimev1.Operation_OPERATION_NIN}
	for _, v := range c.Exprs[1:] {
		s, ok := v.GetVal().AsInterface().(string)
		if !ok {
			return DimensionFilter{}, false
		}
		f.Values = append(f.Values, s)
	}
	return f, true
}

// whereFromProperties returns the where expression of the ai resolver's properties, or nil if they have none.
func whereFromProperties(props map[string]any) (*runtimev1.Expression, error) {
	where, ok := props["where"].(map[string]any)
	if !ok {
		return nil, nil
	}
	data, err := json.Marshal(where)
	if err != nil {
		return nil, err
	}
	expr := &metricsview.Expression{}
	if err := json.Unmarshal(data, expr); err != nil {
		return nil, fmt.Errorf("failed to read the filters: %w", err)
	}
	return metricsview.ExpressionToProto(expr), nil
}

// setWhereProperty sets the where expression of the ai resolver's properties, or removes it if it's nil.
func setWhereProperty(props map[string]any, where *runtimev1.Expression) error {
	if where == nil {
		delete(props, "where")
		return nil
	}
	m, err := metricsview.NewExpressionFromProto(where).AsMap()
	if err != nil {
		return err
	}
	props["where"] = m
	return nil
}

// whereFields sets the where argument of a definition from a where expression,
// and the conditions it can't express as where_saved_otherwise, which an edit of where keeps.
func whereFields(def map[string]any, where *runtimev1.Expression) {
	filters, others := splitWhere(where)
	setField(def, "where", filters)
	if len(others) > 0 {
		var vals []any
		for _, o := range others {
			vals = append(vals, protoValue(o))
		}
		def["where_saved_otherwise"] = vals
	}
}

// containsFilter decodes the LIKE or NLIKE condition of the dashboard's Contains mode, a text between % wildcards, into a contains filter.
func containsFilter(c *runtimev1.Condition) (DimensionFilter, bool) {
	if c == nil || (c.Op != runtimev1.Operation_OPERATION_LIKE && c.Op != runtimev1.Operation_OPERATION_NLIKE) || len(c.Exprs) != 2 || c.Exprs[0].GetIdent() == "" {
		return DimensionFilter{}, false
	}
	pattern, ok := c.Exprs[1].GetVal().AsInterface().(string)
	if !ok || len(pattern) < 3 || !strings.HasPrefix(pattern, "%") || !strings.HasSuffix(pattern, "%") {
		return DimensionFilter{}, false
	}
	return DimensionFilter{Dimension: c.Exprs[0].GetIdent(), Contains: pattern[1 : len(pattern)-1], Exclude: c.Op == runtimev1.Operation_OPERATION_NLIKE}, true
}

// sameJSON reports whether two JSON documents have the same content, whatever the order of their keys.
func sameJSON(a, b string) bool {
	var va, vb any
	if json.Unmarshal([]byte(a), &va) != nil || json.Unmarshal([]byte(b), &vb) != nil {
		return a == b
	}
	return reflect.DeepEqual(va, vb)
}

// FieldChange is a setting an edit changed.
type FieldChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

// diffFields returns the fields whose values differ, sorted by name.
func diffFields(before, after map[string]any) []FieldChange {
	keys := slices.Collect(maps.Keys(before))
	for k := range after {
		if _, ok := before[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	var changes []FieldChange
	for _, k := range keys {
		if !reflect.DeepEqual(before[k], after[k]) {
			changes = append(changes, FieldChange{Field: k, Old: before[k], New: after[k]})
		}
	}
	return changes
}

// setField sets a field for diffFields, leaving out empty values so a field that is unset in one version and empty in the other isn't a change.
func setField(m map[string]any, key string, v any) {
	switch x := v.(type) {
	case string:
		if x == "" {
			return
		}
	case []string:
		if len(x) == 0 {
			return
		}
	case bool:
		if !x {
			return
		}
	case uint32:
		if x == 0 {
			return
		}
	case uint64:
		if x == 0 {
			return
		}
	case nil:
		return
	}
	m[key] = v
}

// setJSONFields sets the top-level keys of a JSON object as fields, as prefix.key.
func setJSONFields(m map[string]any, prefix, data string) {
	if data == "" {
		return
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		m[prefix] = data
		return
	}
	setMapFields(m, prefix, obj)
}

// setMapFields sets the keys of a map as fields, as prefix.key.
func setMapFields(m map[string]any, prefix string, obj map[string]any) {
	for k, v := range obj {
		m[prefix+"."+k] = normalizeJSON(v)
	}
}

// normalizeJSON converts a value to its JSON form, so values built in Go compare equal to values parsed from JSON.
func normalizeJSON(v any) any {
	data, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var res any
	_ = json.Unmarshal(data, &res)
	return res
}

// recipientsFields sets the fields of recipients for diffFields.
func recipientsFields(m map[string]any, emails, slackUsers, slackChannels, slackWebhooks []string) {
	setField(m, "recipients.emails", emails)
	setField(m, "recipients.slack_users", slackUsers)
	setField(m, "recipients.slack_channels", slackChannels)
	setField(m, "recipients.slack_webhooks", slackWebhooks)
}

// recipientsValue presents the recipients of an alert or report in the shape of Recipients.
func recipientsValue(emails, slackUsers, slackChannels, slackWebhooks []string) map[string]any {
	m := map[string]any{}
	setField(m, "emails", emails)
	setField(m, "slack_users", slackUsers)
	setField(m, "slack_channels", slackChannels)
	setField(m, "slack_webhooks", slackWebhooks)
	return m
}

// scheduleValue presents a schedule in the shape of Schedule, or nil if there is no cron.
func scheduleValue(cron, timeZone string) any {
	if cron == "" {
		return nil
	}
	m := map[string]any{"cron": cron}
	setField(m, "time_zone", timeZone)
	return m
}

// exploreFromPath returns the explore an alert opens on, from its web_open_path.
func exploreFromPath(path string) string {
	return strings.TrimPrefix(path, "/explore/")
}

// cloneAlertOptions returns a copy of alert options that can be changed without changing the original.
func cloneAlertOptions(o *drivers.AlertOptions) *drivers.AlertOptions {
	c := *o
	c.ResolverProperties = maps.Clone(o.ResolverProperties)
	return &c
}

// cloneReportOptions returns a copy of report options that can be changed without changing the original.
func cloneReportOptions(o *drivers.ReportOptions) *drivers.ReportOptions {
	c := *o
	c.ResolverProperties = maps.Clone(o.ResolverProperties)
	c.MetricsViewFilters = maps.Clone(o.MetricsViewFilters)
	return &c
}
