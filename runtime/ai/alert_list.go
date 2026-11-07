package ai

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const ListAlertsName = "list_alerts"

// ListAlerts lists the current user's own alerts with their versions, so the agent can find the one to edit instead of creating a duplicate,
// and edit it from its current definition.
type ListAlerts struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListAlertsArgs, *ListAlertsResult] = (*ListAlerts)(nil)

type ListAlertsArgs struct{}

type ListAlertsResult struct {
	Alerts []map[string]any `json:"alerts"`
}

func (t *ListAlerts) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListAlertsName,
		Title:       "List my alerts",
		Description: "Lists the current user's own alerts, formula alerts (condition_type 'criteria') and AI alerts (condition_type 'ai'), with their version and current definition in the arguments of save_alert, plus how they were saved. Use it to find an alert to edit with save_alert, or when the user asks what alerts they have. To edit one, pass its name and version to save_alert with only the arguments to change. Alerts of other users and alerts written in the project's code are not listed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing alerts...",
			"openai/toolInvocation/invoked":  "Listed alerts",
		},
	}
}

func (t *ListAlerts) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *ListAlerts) Handler(ctx context.Context, args *ListAlertsArgs) (*ListAlertsResult, error) {
	s := GetSession(ctx)
	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	managed, err := admin.ListAlerts(ctx, s.Claims().UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	slices.SortFunc(managed, func(a, b *drivers.ManagedAlert) int { return strings.Compare(a.Name, b.Name) })

	alerts := []map[string]any{}
	for _, a := range managed {
		alerts = append(alerts, alertDefinition(a))
	}
	return &ListAlertsResult{Alerts: alerts}, nil
}

// alertDefinition presents a managed alert in the arguments of save_alert, where they can express it.
func alertDefinition(a *drivers.ManagedAlert) map[string]any {
	o := a.Options
	def := map[string]any{
		"name":       a.Name,
		"version":    a.Version,
		"recipients": recipientsValue(o.EmailRecipients, o.SlackUsers, o.SlackChannels, o.SlackWebhooks),
		"renotify":   renotifyValue(o),
	}
	setField(def, "display_name", o.DisplayName)
	setField(def, "explore", exploreFromPath(o.WebOpenPath))
	setField(def, "renotify_after_seconds", o.RenotifyAfterSeconds)
	setField(def, "schedule", scheduleValue(o.RefreshCron, o.RefreshTimeZone))
	if o.Resolver == "ai" {
		return aiAlertDefinition(def, o.ResolverProperties)
	}
	if o.Resolver != "" {
		def["resolver"] = o.Resolver
		def["resolver_properties"] = o.ResolverProperties
	}
	if o.QueryArgsJSON == "" {
		return def
	}
	def["condition_type"] = "criteria"
	def["query"] = o.QueryArgsJSON

	q, err := parseAggregation(o.QueryArgsJSON)
	if err != nil {
		return def
	}
	if len(q.Measures) > 0 {
		def["measure"] = q.Measures[0].Name
		if criteria, joiner, ok := criteriaFromHaving(q.Having, q.Measures[0].Name); ok {
			def["criteria"] = criteria
			def["criteria_operator"] = joiner
		}
	}
	setField(def, "comparison", comparisonKindOf(q.ComparisonTimeRange))
	if len(q.Dimensions) > 0 {
		def["dimension"] = q.Dimensions[0].Name
	}
	setField(def, "time_range", timeRangeValue(q.TimeRange))
	whereFields(def, q.Where)
	return def
}

// aiAlertDefinition adds an AI alert's resolver properties to its definition, in the arguments of save_alert where they can express them.
func aiAlertDefinition(def, props map[string]any) map[string]any {
	def["condition_type"] = "ai"
	def["resolver_properties"] = props
	setField(def, "condition", props["prompt"])
	if explore, _ := props["explore"].(string); explore != "" {
		def["explore"] = explore
	}
	for key, list := range map[string]string{"measure": "measures", "dimension": "dimensions"} {
		if items, ok := normalizeJSON(props[list]).([]any); ok && len(items) > 0 {
			def[key] = items[0]
		}
	}
	if where, err := whereFromProperties(props); err == nil {
		whereFields(def, where)
	}
	if tr, ok := props["time_range"].(map[string]any); ok {
		if expr, _ := tr["expression"].(string); expr != "" {
			def["time_range"] = expr
		} else {
			def["time_range"] = tr
		}
	}
	setField(def, "time_zone", props["time_zone"])
	return def
}
