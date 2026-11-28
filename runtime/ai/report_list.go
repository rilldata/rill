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
	Schedule  string `json:"schedule,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (t *ListReports) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListReportsName,
		Title:       "List my scheduled reports",
		Description: "Lists the current user's own scheduled reports, with the URL of the page where the user can see, edit and delete each one. Use it when the user asks what reports they have, wants to change one (give them its URL) or delete one (with delete_report). Reports of other users and reports written in the project's code are not listed.",
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
			Schedule:    spec.RefreshSchedule.GetCron(),
		}
		if info.Dashboard == "" {
			info.Dashboard = spec.Annotations["canvas"]
		}
		info.URL, err = resourcePageURL(ctx, t.Runtime, "reports", info.Name)
		if err != nil {
			return nil, err
		}
		res.Reports = append(res.Reports, info)
	}
	return res, nil
}
