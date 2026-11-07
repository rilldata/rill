package ai

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const ListReportsName = "list_reports"

// ListReports lists the current user's own scheduled reports with their versions, so the agent can find the one to edit instead of creating a duplicate,
// and edit it from its current definition.
type ListReports struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListReportsArgs, *ListReportsResult] = (*ListReports)(nil)

type ListReportsArgs struct{}

type ListReportsResult struct {
	Reports []map[string]any `json:"reports"`
}

func (t *ListReports) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListReportsName,
		Title:       "List my reports",
		Description: "Lists the current user's own scheduled reports, with their version and current definition in the arguments of save_report (kind 'explore_data', 'ai' or 'canvas_pdf'), plus how they were saved. Use it to find a report to edit with save_report, or when the user asks what reports they have. To edit one, pass its name and version to save_report with only the arguments to change. Reports of other users and reports written in the project's code are not listed.",
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
	s := GetSession(ctx)
	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	managed, err := admin.ListReports(ctx, s.Claims().UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	slices.SortFunc(managed, func(a, b *drivers.ManagedReport) int { return strings.Compare(a.Name, b.Name) })

	reports := []map[string]any{}
	for _, r := range managed {
		reports = append(reports, reportDefinition(r))
	}
	return &ListReportsResult{Reports: reports}, nil
}

// reportDefinition presents a managed report in the arguments of save_report, where they can express it.
func reportDefinition(r *drivers.ManagedReport) map[string]any {
	o := r.Options
	openMode := o.WebOpenMode
	if openMode == "" {
		openMode = "recipient" // As the admin service reads it
	}
	def := map[string]any{
		"name":       r.Name,
		"version":    r.Version,
		"kind":       reportKind(o),
		"recipients": recipientsValue(o.EmailRecipients, o.SlackUsers, o.SlackChannels, o.SlackWebhooks),
		"open_mode":  openMode,
	}
	setField(def, "display_name", o.DisplayName)
	setField(def, "schedule", scheduleValue(o.RefreshCron, o.RefreshTimeZone))

	switch def["kind"] {
	case "explore_data":
		def["explore"] = o.Explore
		def["export"] = map[string]any{
			"format":         exportFormatName(o.ExportFormat),
			"limit":          o.ExportLimit,
			"include_header": o.ExportIncludeHeader,
			"attach":         o.ExportAttach,
		}
		def["query"] = o.QueryArgsJSON
		q, err := parseAggregation(o.QueryArgsJSON)
		if err != nil {
			break
		}
		var dims, measures []string
		for _, d := range q.Dimensions {
			dims = append(dims, d.Name)
		}
		for _, m := range q.Measures {
			measures = append(measures, m.Name)
		}
		setField(def, "dimensions", dims)
		setField(def, "measures", measures)
		setField(def, "time_range", timeRangeValue(q.TimeRange))
		whereFields(def, q.Where)
	case "ai":
		props := o.ResolverProperties
		def["resolver_properties"] = props
		setField(def, "prompt", props["prompt"])
		setField(def, "explore", props["explore"])
		if tr, ok := props["time_range"].(map[string]any); ok {
			setField(def, "time_range", tr["expression"])
		}
		def["compare_to_previous_period"] = props["comparison_time_range"] != nil
		if cmp, ok := props["comparison_time_range"].(map[string]any); ok {
			expr, _ := cmp["expression"].(string)
			setField(def, "comparison", comparisonKindOfExpression(expr))
		}
	case "canvas_pdf":
		def["canvas"] = o.Canvas
		def["canvas_state"] = o.WebOpenState
		state, err := url.ParseQuery(o.WebOpenState)
		if err != nil {
			break
		}
		setField(def, "time_range", state.Get("tr"))
		def["pdf_include_filters"] = state.Get("pdf_include_filters") != "false"
		def["pdf_all_tabs"] = state.Get("pdf_all_tabs") != "false"
		if tabs := decodeTabs(state.Get(canvasTabsParam)); len(tabs) > 0 {
			def["tabs"] = tabs
		}
		var filters []CanvasFilters
		for _, mv := range slices.Sorted(maps.Keys(o.MetricsViewFilters)) {
			fs, ok := filtersFromExpression(o.MetricsViewFilters[mv])
			if !ok {
				def["canvas_filters_saved_otherwise"] = true
				continue
			}
			filters = append(filters, CanvasFilters{MetricsView: mv, Filters: fs})
		}
		if len(filters) > 0 {
			def["canvas_filters"] = filters
		}
	}
	return def
}
