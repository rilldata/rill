package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ListReportsName = "list_reports"

// ListReports lists the current user's own scheduled reports, so the agent can link to one or delete it.
type ListReports struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListReportsArgs, *ListReportsResult] = (*ListReports)(nil)

type ListReportsArgs struct{}

type ListReportsResult struct {
	Reports []*ReportInfo `json:"reports"`
}

// ReportInfo describes one of the user's reports.
type ReportInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	// Dashboard is the explore or canvas the report is on, if any.
	Dashboard string `json:"dashboard,omitempty"`
	// AI is true for an analysis written by the AI, which its page can't edit.
	AI       bool   `json:"ai,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	// TimeZone is the time zone of the schedule's cron expression.
	TimeZone string `json:"time_zone,omitempty"`
	URL      string `json:"url,omitempty"`
}

func (t *ListReports) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListReportsName,
		Title:       "List my scheduled reports",
		Description: "Lists the current user's own scheduled reports, with the URL of the page where the user can see and delete each one, and edit it unless it's an AI report (ai is true). Use it when the user asks what reports they have, wants to change one (give them its URL; to change an AI report, create a new one and delete the old one) or delete one (with delete_report). Reports of other users and reports written in the project's code are not listed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing reports...",
			"openai/toolInvocation/invoked":  "Listed reports",
		},
	}
}

func (t *ListReports) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *ListReports) Handler(ctx context.Context, args *ListReportsArgs) (*ListReportsResult, error) {
	owned, err := ownedResources(ctx, t.Runtime, runtime.ResourceKindReport)
	if err != nil {
		return nil, err
	}
	res := &ListReportsResult{Reports: []*ReportInfo{}}
	for _, r := range owned {
		spec := r.GetReport().Spec
		info := &ReportInfo{
			Name:        r.Meta.Name.Name,
			DisplayName: spec.DisplayName,
			Dashboard:   spec.Annotations["explore"],
			AI:          spec.Resolver == "ai",
			Schedule:    spec.RefreshSchedule.GetCron(),
			TimeZone:    spec.RefreshSchedule.GetTimeZone(),
		}
		if info.Dashboard == "" {
			info.Dashboard = spec.Annotations["canvas"]
		}
		// A cron without a time zone runs in UTC.
		if info.Schedule != "" && info.TimeZone == "" {
			info.TimeZone = "UTC"
		}
		info.URL, err = resourcePageURL(ctx, t.Runtime, "reports", info.Name)
		if err != nil {
			return nil, err
		}
		res.Reports = append(res.Reports, info)
	}
	return res, nil
}
