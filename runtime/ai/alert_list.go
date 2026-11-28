package ai

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/encoding/protojson"
)

const ListAlertsName = "list_alerts"

// ListAlerts lists the current user's own alerts, so the agent can link to one or delete it.
type ListAlerts struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListAlertsArgs, *ListAlertsResult] = (*ListAlerts)(nil)

type ListAlertsArgs struct{}

type ListAlertsResult struct {
	Alerts []*AlertInfo `json:"alerts"`
}

// AlertInfo describes one of the user's alerts.
type AlertInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Explore     string `json:"explore,omitempty"`
	// Condition is the AI condition of an AI alert, or the criteria of a formula alert.
	Condition string `json:"condition,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (t *ListAlerts) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListAlertsName,
		Title:       "List my alerts",
		Description: "Lists the current user's own alerts, with the URL of the page where the user can see, edit and delete each one. Use it when the user asks what alerts they have, wants to change one (give them its URL) or delete one (with delete_alert). Alerts of other users and alerts written in the project's code are not listed.",
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
	owned, err := ownedResources(ctx, t.Runtime, runtime.ResourceKindAlert)
	if err != nil {
		return nil, err
	}
	res := &ListAlertsResult{Alerts: []*AlertInfo{}}
	for _, r := range owned {
		spec := r.GetAlert().Spec
		info := &AlertInfo{
			Name:        r.Meta.Name.Name,
			DisplayName: spec.DisplayName,
			Explore:     strings.TrimPrefix(spec.Annotations["web_open_path"], "/explore/"),
			Condition:   alertCondition(spec),
		}
		info.URL, err = resourcePageURL(ctx, t.Runtime, "alerts", info.Name)
		if err != nil {
			return nil, err
		}
		res.Alerts = append(res.Alerts, info)
	}
	return res, nil
}

// alertCondition describes an alert's condition in one line: the prompt of an AI alert, or the criteria of a formula alert as SQL.
func alertCondition(spec *runtimev1.AlertSpec) string {
	props := spec.ResolverProperties.AsMap()
	if spec.Resolver == "ai" {
		prompt, _ := props["prompt"].(string)
		return prompt
	}
	argsJSON, _ := props["query_args_json"].(string)
	q := &runtimev1.MetricsViewAggregationRequest{}
	if err := protojson.Unmarshal([]byte(argsJSON), q); err != nil || q.Having == nil {
		return ""
	}
	sql, err := metricsview.ExpressionToSQL(metricsview.NewExpressionFromProto(q.Having))
	if err != nil {
		return ""
	}
	return sql
}
