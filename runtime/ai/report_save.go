package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const SaveReportName = "save_report"

// SaveReport creates a scheduled report for the current user, or edits one of the user's own, through the admin service.
// What it saves is what the report dialog saves: the report is owned by the user, listed in their reports and editable in the dialog.
// An edit only changes the arguments it sends, so whatever the dialog saved that the arguments can't express is kept.
type SaveReport struct {
	Runtime *runtime.Runtime
}

var _ Tool[*SaveReportArgs, *SaveReportResult] = (*SaveReport)(nil)

type SaveReportArgs struct {
	Name        string  `json:"name,omitempty" jsonschema:"Name of one of the user's own reports to edit, as returned by list_reports. Leave empty to create a new report."`
	Version     string  `json:"version,omitempty" jsonschema:"Version of the report to edit, as returned by list_reports or by the last save_report of it. Required to edit."`
	DisplayName string  `json:"display_name,omitempty" jsonschema:"Display name of the report. Required to create one."`
	Kind        string  `json:"kind,omitempty" jsonschema:"What the report sends: 'explore_data' an export of an explore dashboard's data (CSV or XLSX); 'ai' a written analysis by the AI with its insights; 'canvas_pdf' a canvas dashboard as a PDF. Required to create one; it can't change."`
	Explore     *string `json:"explore,omitempty" jsonschema:"Name of the explore dashboard. The name of a metrics view also works if it has a single explore; if not, the error lists the explores to choose from. Required for 'explore_data'; optional for 'ai', to focus the analysis on it, and null removes it."`
	Canvas      string  `json:"canvas,omitempty" jsonschema:"Name of the canvas dashboard. Required for 'canvas_pdf'."`
	// Data of explore_data reports
	Dimensions []string      `json:"dimensions,omitempty" jsonschema:"For 'explore_data': dimensions to break the data down by. Passing them replaces the current ones."`
	Measures   []string      `json:"measures,omitempty" jsonschema:"For 'explore_data': measures to export. At least one to create one; passing them replaces the current ones."`
	Export     *ReportExport `json:"export,omitempty" jsonschema:"For 'explore_data': the export's format and options. Each option passed replaces the current one, and null resets them all to their defaults."`
	// Data of ai reports
	Prompt                  *string `json:"prompt,omitempty" jsonschema:"For 'ai': what the analysis should focus on. Without it, or with null, it's a general analysis of the key metrics."`
	CompareToPreviousPeriod *bool   `json:"compare_to_previous_period,omitempty" jsonschema:"For 'ai': compare the time range with an earlier period, by default the previous period of the same length."`
	Comparison              string  `json:"comparison,omitempty" jsonschema:"For 'ai': the period the time range is compared with, 'previous_period' (default), 'previous_day', 'previous_week', 'previous_month', 'previous_quarter' or 'previous_year'. Passing it turns the comparison on. Omitted in an edit, the current comparison is kept, also when the time range changes."`
	// Data of canvas_pdf reports
	CanvasFilters     []CanvasFilters `json:"canvas_filters,omitempty" jsonschema:"For 'canvas_pdf': filters of the canvas, per metrics view of the canvas. Each metrics view passed gets these filters instead of its current ones (none removes them); the others keep theirs. Null removes the filters of every metrics view."`
	PDFIncludeFilters *bool           `json:"pdf_include_filters,omitempty" jsonschema:"For 'canvas_pdf': whether the PDF shows the filters applied. Defaults to true."`
	PDFAllTabs        *bool           `json:"pdf_all_tabs,omitempty" jsonschema:"For 'canvas_pdf': whether the PDF includes every tab of the canvas, or only the selected ones. Defaults to true."`
	Tabs              []CanvasTab     `json:"tabs,omitempty" jsonschema:"For 'canvas_pdf': the tab selected in each tab group of the canvas. Groups not listed show their first tab; null selects the first tab of every group."`
	// Shared
	TimeRange  *string           `json:"time_range,omitempty" jsonschema:"Time range of the data, as a Rill time expression relative to the latest data, for example '7D as of latest/D' for the last 7 complete days. Required for 'explore_data'; for 'ai' and 'canvas_pdf', null or nothing uses the dashboard's default."`
	Where      []DimensionFilter `json:"where,omitempty" jsonschema:"For 'explore_data' and 'ai': filters on dimensions of the explore. Passing them replaces the current dimension filters, and null removes them; filters saved otherwise, like the dashboards' measure filters, are kept."`
	Schedule   *Schedule         `json:"schedule,omitempty" jsonschema:"When to send the report. Required to create one."`
	Recipients *Recipients       `json:"recipients,omitempty" jsonschema:"Who to send it to. Defaults to the user's own email when creating. Each kind of recipient passed replaces the current ones of that kind."`
	OpenMode   string            `json:"open_mode,omitempty" jsonschema:"Whose permissions the report opens with: 'recipient' (default) each recipient's own, so only project members can open it, after logging in; 'creator' the user's own, so every recipient sees the data the user can see, without logging in. Only use 'creator' when the user chooses it knowing this; it's required to send to people outside the project or to Slack, or to attach the export."`

	sent sentArgs
}

func (a *SaveReportArgs) UnmarshalJSON(data []byte) error {
	type plain SaveReportArgs
	if err := json.Unmarshal(data, (*plain)(a)); err != nil {
		return err
	}
	return a.sent.parse(data)
}

// MarshalJSON keeps the arguments as they were sent, so an argument sent as null still clears its setting when the call is replayed.
func (a *SaveReportArgs) MarshalJSON() ([]byte, error) {
	if a.sent.raw != nil {
		return a.sent.raw, nil
	}
	type plain SaveReportArgs
	return json.Marshal((*plain)(a))
}

// ReportExport are the export options of an explore data report.
type ReportExport struct {
	Format        string `json:"format,omitempty" jsonschema:"'csv' (default), 'xlsx' or 'parquet'."`
	Limit         uint64 `json:"limit,omitempty" jsonschema:"Maximum number of rows. 0 for no limit."`
	IncludeHeader bool   `json:"include_header,omitempty" jsonschema:"Add a header with the report's name, filters and time range above the data."`
	Attach        bool   `json:"attach,omitempty" jsonschema:"Attach the export to the email instead of only linking to it. Requires open_mode 'creator'."`
}

// CanvasFilters are the filters of a canvas report on one of the canvas' metrics views.
type CanvasFilters struct {
	MetricsView string            `json:"metrics_view" jsonschema:"Name of a metrics view of the canvas."`
	Filters     []DimensionFilter `json:"filters,omitempty" jsonschema:"Filters on its dimensions. None removes its filters."`
}

// CanvasTab is the tab selected in a tab group of a canvas.
type CanvasTab struct {
	Group string `json:"group" jsonschema:"Name of the tab group."`
	Tab   string `json:"tab" jsonschema:"Name of the selected tab."`
}

type SaveReportResult struct {
	Name           string        `json:"name" jsonschema:"Name of the saved report. Pass it to save_report to edit the same report."`
	Version        string        `json:"version" jsonschema:"Version of the saved report. Pass it to save_report to edit it again."`
	URL            string        `json:"url" jsonschema:"URL where the user can see and edit the report."`
	Changes        []FieldChange `json:"changes,omitempty" jsonschema:"For an edit, the settings it changed, with their old and new values. Tell the user what changed."`
	ReconcileError string        `json:"reconcile_error,omitempty" jsonschema:"Error the report has, if any. Tell the user and offer to fix it."`
}

func (t *SaveReport) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  SaveReportName,
		Title: "Save scheduled report",
		Description: "Creates a scheduled report for the current user, or edits one of the user's own: an export of an explore dashboard's data, an analysis written by the AI, or a canvas dashboard as a PDF, sent by email or Slack on a schedule. " +
			"Use it when the user asks to be sent data, a summary or a dashboard regularly, for example 'send me the sales by country every Monday'. " +
			"The report belongs to the user, is listed in their reports, and they can keep editing it in the report dialog. Only create a file in the project's code instead when the user explicitly asks for that. " +
			confirmBeforeSaving +
			"Reports open with each recipient's own permissions by default (open_mode 'recipient'). " +
			"Sending to Slack users or channels, or attaching the export, needs open_mode 'creator': when the user asks for any of them, say so in the preview, explaining that with 'creator' every recipient sees the data with the user's permissions, without logging in, and only use it if they agree. " +
			"Sending by email to someone who isn't a member of the project needs it too, but only saving can tell: if saving fails for that reason, explain the same and only save with 'creator' if they agree. " +
			"To edit a report, find its name, version and current definition with list_reports, even in a new conversation; if you can't find the report the user means, ask them instead of creating a new one. " +
			"An edit only changes the arguments you pass: every other setting keeps its current value, and passing null clears an optional one. If the report changed since it was listed, the edit fails: list it again and reapply the change. " +
			"It can't delete reports: tell the user to delete them from the Reports page. Give the user the report's URL from the response, and for an edit, the changes it lists.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(true), // It notifies people
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Saving report...",
			"openai/toolInvocation/invoked":  "Saved report",
		},
	}
}

func (t *SaveReport) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *SaveReport) Handler(ctx context.Context, args *SaveReportArgs) (*SaveReportResult, error) {
	if err := args.sent.ensure(args); err != nil {
		return nil, err
	}
	s := GetSession(ctx)

	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	// An edit starts from the report as it's stored, at the version it was listed with.
	var prev *drivers.ManagedReport
	var base *drivers.ReportOptions
	kind := args.Kind
	if args.Name == "" {
		base = &drivers.ReportOptions{WebOpenMode: "recipient"}
		switch kind {
		case "explore_data":
			base.QueryName = "MetricsViewAggregation"
			base.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_CSV
		case "ai":
			base.Resolver = "ai"
		case "canvas_pdf":
			base.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_PDF
			base.WebOpenState = "pdf_all_tabs=true&pdf_include_filters=true"
		default:
			return nil, fmt.Errorf("invalid kind %q: use 'explore_data', 'ai' or 'canvas_pdf'", args.Kind)
		}
	} else {
		if args.Version == "" {
			return nil, errors.New("invalid input: version is required to edit a report: call list_reports for its current version")
		}
		reports, err := admin.ListReports(ctx, s.Claims().UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to list reports: %w", err)
		}
		for _, r := range reports {
			if r.Name == args.Name {
				prev = r
			}
		}
		if prev == nil {
			return nil, fmt.Errorf("report %q not found: call list_reports for the names of the user's reports", args.Name)
		}
		if prev.Version != args.Version {
			return nil, editConflictError("report", args.Name, ListReportsName)
		}
		base = prev.Options
		kind = reportKind(base)
		if args.sent.has("kind") && args.Kind != kind {
			return nil, fmt.Errorf("invalid input: the kind of a report can't change, and this one is %q: create a new report instead", kind)
		}
	}

	opts, err := t.applyArgs(ctx, args, base, kind, prev == nil)
	if err != nil {
		return nil, err
	}

	res, err := admin.SaveReport(ctx, s.Claims().UserID, args.Name, args.Version, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	reconcileErr, err := managedResourceError(ctx, t.Runtime, runtime.ResourceKindReport, res.Name)
	if err != nil {
		return nil, err
	}

	result := &SaveReportResult{
		Name:           res.Name,
		Version:        res.Version,
		URL:            res.URL,
		ReconcileError: reconcileErr,
	}
	if prev != nil {
		result.Changes = diffFields(reportFields(prev.Options), reportFields(opts))
	}
	return result, nil
}

// applyArgs applies the arguments a call sent to a report's options, which only have defaults for its kind when creating one, and validates the result.
func (t *SaveReport) applyArgs(ctx context.Context, args *SaveReportArgs, base *drivers.ReportOptions, kind string, create bool) (*drivers.ReportOptions, error) {
	sent := &args.sent
	opts := cloneReportOptions(base)

	if sent.has("display_name") {
		if args.DisplayName == "" {
			return nil, errors.New("invalid input: display_name can't be empty")
		}
		opts.DisplayName = args.DisplayName
	}
	if opts.DisplayName == "" {
		return nil, errors.New("invalid input: display_name is required")
	}

	if sent.has("schedule") {
		if args.Schedule == nil {
			return nil, errors.New("invalid input: a report needs a schedule")
		}
		if err := checkSchedule(args.Schedule); err != nil {
			return nil, err
		}
		opts.RefreshCron, opts.RefreshTimeZone = args.Schedule.Cron, args.Schedule.TimeZone
	}
	if opts.RefreshCron == "" {
		return nil, errors.New("invalid input: schedule.cron is required")
	}
	if sent.has("time_range") && args.TimeRange != nil {
		if err := checkTimeRange(*args.TimeRange); err != nil {
			return nil, err
		}
	}

	// The admin service checks the recipients and options that need the creator's permissions, since it knows the project's members.
	if sent.has("open_mode") {
		switch args.OpenMode {
		case "recipient", "creator":
			opts.WebOpenMode = args.OpenMode
		default:
			return nil, fmt.Errorf("invalid open_mode %q: use 'recipient' or 'creator'", args.OpenMode)
		}
	}

	if sent.has("comparison") && kind != "ai" {
		return nil, errors.New("invalid input: comparison only applies to 'ai' reports")
	}

	// The data of the report is only touched when the call sends some of it, so an edit of anything else keeps it exactly as saved.
	var err error
	switch {
	case kind == "explore_data" && (create || sent.hasAny("explore", "dimensions", "measures", "time_range", "where", "export")):
		err = t.applyExploreDataArgs(ctx, args, opts, create)
	case kind == "ai" && (create || sent.hasAny("explore", "prompt", "time_range", "compare_to_previous_period", "comparison", "where")):
		err = t.applyAIArgs(ctx, args, opts, create)
	case kind == "canvas_pdf" && (create || sent.hasAny("canvas", "time_range", "canvas_filters", "pdf_include_filters", "pdf_all_tabs", "tabs", "schedule.time_zone")):
		err = t.applyCanvasArgs(ctx, args, opts)
	case kind == "" && sent.hasAny("explore", "canvas", "dimensions", "measures", "export", "prompt", "compare_to_previous_period", "canvas_filters", "pdf_include_filters", "pdf_all_tabs", "tabs", "time_range", "where"):
		err = errors.New("the data of this report was saved in a form this tool can't change, so only its name, schedule, recipients and open mode can be changed here: change the rest in the report dialog")
	}
	if err != nil {
		return nil, err
	}

	var r Recipients
	if args.Recipients != nil {
		r = *args.Recipients
	}
	patchRecipients(sent, r, &opts.EmailRecipients, &opts.SlackUsers, &opts.SlackChannels, &opts.SlackWebhooks)
	if create {
		selfIfNoRecipients(ctx, &opts.EmailRecipients, opts.SlackUsers, opts.SlackChannels)
	} else if len(opts.EmailRecipients)+len(opts.SlackUsers)+len(opts.SlackChannels)+len(opts.SlackWebhooks) == 0 {
		return nil, errors.New("invalid input: a report needs at least one recipient")
	}
	return opts, nil
}

// applyExploreDataArgs applies the data arguments a call sent to a report that exports an explore's data:
// a MetricsViewAggregation query, as the report dialog saves it. It keeps what the arguments don't change,
// like which dimensions the dialog pivoted into columns, and it clears the dashboard state the report opens with when the query changes.
func (t *SaveReport) applyExploreDataArgs(ctx context.Context, args *SaveReportArgs, opts *drivers.ReportOptions, create bool) error {
	sent := &args.sent
	q, err := parseAggregation(opts.QueryArgsJSON)
	if err != nil {
		return err
	}

	exploreName := opts.Explore
	if sent.has("explore") {
		if args.Explore == nil || *args.Explore == "" {
			return errors.New("invalid input: an 'explore_data' report needs an explore")
		}
		exploreName = *args.Explore
	}
	exploreName, explore, err := validExplore(ctx, t.Runtime, exploreName)
	if err != nil {
		return err
	}
	q.MetricsView = explore.MetricsView
	opts.Explore = exploreName

	if sent.has("dimensions") {
		var dims []*runtimev1.MetricsViewAggregationDimension
		for _, name := range args.Dimensions {
			// Keep a dimension as saved if it's still there, since the dialog may have saved it with options of its own.
			i := slices.IndexFunc(q.Dimensions, func(d *runtimev1.MetricsViewAggregationDimension) bool { return d.Name == name })
			if i >= 0 {
				dims = append(dims, q.Dimensions[i])
			} else {
				dims = append(dims, &runtimev1.MetricsViewAggregationDimension{Name: name})
			}
		}
		q.Dimensions = dims
		if q.PivotOn != nil {
			q.PivotOn = slices.DeleteFunc(q.PivotOn, func(p string) bool { return !slices.Contains(args.Dimensions, p) })
		}
	}
	if sent.has("measures") {
		var measures []*runtimev1.MetricsViewAggregationMeasure
		for _, name := range args.Measures {
			i := slices.IndexFunc(q.Measures, func(m *runtimev1.MetricsViewAggregationMeasure) bool { return m.Name == name })
			if i >= 0 {
				measures = append(measures, q.Measures[i])
			} else {
				measures = append(measures, &runtimev1.MetricsViewAggregationMeasure{Name: name})
			}
		}
		q.Measures = measures
	}
	if len(q.Measures) == 0 {
		return errors.New("invalid input: an 'explore_data' report needs at least one measure")
	}
	if sent.has("time_range") {
		if args.TimeRange == nil || *args.TimeRange == "" {
			return errors.New("invalid input: an 'explore_data' report needs a time_range")
		}
		tz := opts.RefreshTimeZone
		if q.TimeRange != nil && q.TimeRange.TimeZone != "" && !create {
			tz = q.TimeRange.TimeZone
		}
		q.TimeRange = &runtimev1.TimeRange{Expression: *args.TimeRange, TimeZone: tz}
		// Keep the kind of comparison the dialog saved, or the previous period for one saved in a form the comparison argument can't express.
		if q.ComparisonTimeRange != nil {
			q.ComparisonTimeRange, err = comparisonFor(q.TimeRange, comparisonKindOf(q.ComparisonTimeRange))
			if err != nil {
				return err
			}
		}
	}
	if q.TimeRange == nil {
		return errors.New("invalid input: an 'explore_data' report needs a time_range")
	}
	if sent.has("where") {
		q.Where, err = replaceFilters(q.Where, args.Where)
		if err != nil {
			return err
		}
	}

	var measures, dims []string
	for _, m := range q.Measures {
		if m.Compute == nil {
			measures = append(measures, m.Name)
		}
	}
	for _, d := range q.Dimensions {
		dims = append(dims, d.Name)
	}
	if err := checkExploreFields(explore, exploreName, measures, append(dims, filterDimensions(args.Where)...)); err != nil {
		return err
	}

	if sent.null("export") {
		// Null resets the export options to their defaults.
		opts.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_CSV
		opts.ExportLimit, opts.ExportIncludeHeader, opts.ExportAttach = 0, false, false
	}
	if sent.has("export.format") {
		format := args.Export.Format
		if format == "" {
			format = "csv"
		}
		f, ok := dataExportFormats[format]
		if !ok {
			return fmt.Errorf("invalid export format %q: use 'csv', 'xlsx' or 'parquet'", args.Export.Format)
		}
		opts.ExportFormat = f
	}
	if sent.has("export.limit") {
		opts.ExportLimit = args.Export.Limit
	}
	if sent.has("export.include_header") {
		opts.ExportIncludeHeader = args.Export.IncludeHeader
	}
	if sent.has("export.attach") {
		opts.ExportAttach = args.Export.Attach
	}

	queryArgs, err := aggregationQueryArgs(q)
	if err != nil {
		return err
	}
	opts.QueryName = "MetricsViewAggregation"
	if create || !sameJSON(queryArgs, opts.QueryArgsJSON) {
		if !create {
			opts.WebOpenState = ""
		}
		opts.QueryArgsJSON = queryArgs
	}
	return nil
}

// applyAIArgs applies the data arguments a call sent to the properties of the ai resolver of an AI report, keeping the properties they don't change.
// It clears the dashboard state the report opens with when the properties change.
func (t *SaveReport) applyAIArgs(ctx context.Context, args *SaveReportArgs, opts *drivers.ReportOptions, create bool) error {
	sent := &args.sent
	// Change a copy, so the result can be compared with the properties as saved.
	props := maps.Clone(opts.ResolverProperties)
	if props == nil {
		props = map[string]any{}
	}

	if sent.has("prompt") {
		if args.Prompt == nil || *args.Prompt == "" {
			delete(props, "prompt")
		} else {
			props["prompt"] = *args.Prompt
		}
	}
	if sent.has("explore") {
		if args.Explore == nil || *args.Explore == "" {
			delete(props, "explore")
		} else {
			props["explore"] = *args.Explore
		}
	}
	exploreName, _ := props["explore"].(string)
	if exploreName != "" {
		name, explore, err := validExplore(ctx, t.Runtime, exploreName)
		if err != nil {
			return err
		}
		exploreName = name
		props["explore"] = name
		if err := checkExploreFields(explore, exploreName, nil, filterDimensions(args.Where)); err != nil {
			return err
		}
	}
	if sent.has("where") {
		if len(args.Where) > 0 && exploreName == "" {
			return errors.New("invalid input: filters of an 'ai' report need an explore")
		}
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
		if args.TimeRange == nil || *args.TimeRange == "" {
			delete(props, "time_range")
			delete(props, "comparison_time_range")
			delete(props, "time_zone")
		} else {
			props["time_range"] = map[string]any{"expression": *args.TimeRange}
			if opts.RefreshTimeZone != "" {
				props["time_zone"] = opts.RefreshTimeZone
			}
		}
	}
	// Keep the kind of comparison the report has, or the previous period for one saved in a form the comparison argument can't express.
	compare := props["comparison_time_range"] != nil
	cmp, _ := props["comparison_time_range"].(map[string]any)
	cmpExpr, _ := cmp["expression"].(string)
	kind := comparisonKindOfExpression(cmpExpr)
	if sent.has("compare_to_previous_period") {
		compare = args.CompareToPreviousPeriod != nil && *args.CompareToPreviousPeriod
	}
	if sent.has("comparison") && args.Comparison != "" {
		if sent.has("compare_to_previous_period") && !compare {
			return errors.New("invalid input: pass either a comparison or compare_to_previous_period false, not both")
		}
		kind, compare = args.Comparison, true
	}
	if compare && (sent.hasAny("time_range", "compare_to_previous_period", "comparison")) {
		tr, _ := props["time_range"].(map[string]any)
		expr, _ := tr["expression"].(string)
		if expr == "" {
			return errors.New("invalid input: comparing to an earlier period needs a time_range")
		}
		cmp, err := comparisonExpression(expr, kind)
		if err != nil {
			return err
		}
		props["comparison_time_range"] = map[string]any{"expression": cmp}
	} else if !compare {
		delete(props, "comparison_time_range")
	}

	if !create && !reflect.DeepEqual(normalizeJSON(props), normalizeJSON(opts.ResolverProperties)) {
		opts.WebOpenState = ""
	}
	opts.Resolver = "ai"
	opts.ResolverProperties = props
	return nil
}

// applyCanvasArgs applies the data arguments a call sent to a canvas PDF report: the canvas state the PDF renders, and the filters baked into the report's security rules.
// It changes the state the report dialog saved only in the parameters the arguments set, so other settings of the canvas are kept.
func (t *SaveReport) applyCanvasArgs(ctx context.Context, args *SaveReportArgs, opts *drivers.ReportOptions) error {
	sent := &args.sent
	state, err := url.ParseQuery(opts.WebOpenState)
	if err != nil {
		return fmt.Errorf("failed to parse the saved canvas state: %w", err)
	}
	before := maps.Clone(state)

	canvasName := opts.Canvas
	if sent.has("canvas") {
		if args.Canvas == "" {
			return errors.New("invalid input: a 'canvas_pdf' report needs a canvas")
		}
		canvasName = args.Canvas
	}
	if canvasName == "" {
		return errors.New("invalid input: a 'canvas_pdf' report needs a canvas")
	}
	canvas, metricsViews, err := (&AnalystAgent{Runtime: t.Runtime}).getValidCanvasAndMetricsViews(ctx, canvasName)
	if err != nil {
		return fmt.Errorf("canvas %q not found: call list_canvases to find the canvas dashboards: %w", canvasName, err)
	}
	// Resource names are case insensitive: keep the canvas name as declared, so the same canvas typed differently keeps its filters.
	canvasName = canvas.Meta.Name.Name
	if canvasName != opts.Canvas && opts.Canvas != "" {
		// The filters and tabs of another canvas don't apply.
		opts.MetricsViewFilters = nil
		for k := range state {
			if strings.HasPrefix(k, "f.") || k == canvasTabsParam {
				state.Del(k)
			}
		}
	}
	opts.Canvas = canvasName

	if sent.has("time_range") {
		if args.TimeRange == nil || *args.TimeRange == "" {
			state.Del("tr")
		} else {
			state.Set("tr", *args.TimeRange)
		}
	}
	if sent.has("schedule.time_zone") && opts.RefreshTimeZone != "" {
		state.Set("tz", opts.RefreshTimeZone)
	}
	if sent.null("canvas_filters") {
		// Null removes the filters of every metrics view, from the report's security rules and from the canvas state alike.
		opts.MetricsViewFilters = nil
		for k := range state {
			if strings.HasPrefix(k, "f.") {
				state.Del(k)
			}
		}
	}
	for _, cf := range args.CanvasFilters {
		mv, ok := metricsViews[cf.MetricsView]
		if !ok {
			return fmt.Errorf("metrics view %q is not used by canvas %q, which uses: %s", cf.MetricsView, canvasName, strings.Join(slices.Sorted(maps.Keys(metricsViews)), ", "))
		}
		for _, f := range cf.Filters {
			if !slices.ContainsFunc(mv.GetMetricsView().State.ValidSpec.Dimensions, func(d *runtimev1.MetricsViewSpec_Dimension) bool { return d.Name == f.Dimension }) {
				return fmt.Errorf("dimension %q is not available in metrics view %q of canvas %q", f.Dimension, cf.MetricsView, canvasName)
			}
		}
		if len(cf.Filters) == 0 {
			delete(opts.MetricsViewFilters, cf.MetricsView)
			state.Del("f." + cf.MetricsView)
			continue
		}
		expr, err := filtersExpression(cf.Filters)
		if err != nil {
			return err
		}
		if opts.MetricsViewFilters == nil {
			opts.MetricsViewFilters = map[string]*runtimev1.Expression{}
		}
		opts.MetricsViewFilters[cf.MetricsView] = expr
		state.Set("f."+cf.MetricsView, filtersURLParam(cf.Filters))
	}
	if sent.has("pdf_include_filters") {
		setBoolParam(state, "pdf_include_filters", args.PDFIncludeFilters)
	}
	if sent.has("pdf_all_tabs") {
		setBoolParam(state, "pdf_all_tabs", args.PDFAllTabs)
	}
	if sent.has("tabs") {
		var pairs []string
		for _, tab := range args.Tabs {
			pairs = append(pairs, encodeTabKey(tab.Group)+"."+encodeTabKey(tab.Tab))
		}
		if len(pairs) == 0 {
			state.Del(canvasTabsParam)
		} else {
			state.Set(canvasTabsParam, strings.Join(pairs, ","))
		}
	}

	// Keep the state as saved unless it changed, since encoding it again may reorder its parameters.
	if !reflect.DeepEqual(state, before) {
		opts.WebOpenState = state.Encode()
	}
	opts.ExportFormat = runtimev1.ExportFormat_EXPORT_FORMAT_PDF
	return nil
}

// canvasTabsParam is the canvas URL parameter with the tab selected in each tab group, as `group.tab` pairs separated by commas (CANVAS_TABS_URL_PARAM in the UI).
const canvasTabsParam = "tabs"

// setBoolParam sets a boolean URL parameter, or removes it for nil so it takes its default.
func setBoolParam(state url.Values, key string, v *bool) {
	if v == nil {
		state.Del(key)
		return
	}
	state.Set(key, fmt.Sprint(*v))
}

// encodeTabKey encodes a tab group or tab name as the UI does: with encodeURIComponent, and dots escaped too.
func encodeTabKey(name string) string {
	var b strings.Builder
	for _, c := range []byte(name) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte("-_!~*'()", c) >= 0 {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// decodeTabs decodes the tabs canvas URL parameter.
func decodeTabs(param string) []CanvasTab {
	var tabs []CanvasTab
	for _, pair := range strings.Split(param, ",") {
		group, tab, ok := strings.Cut(pair, ".")
		if !ok {
			continue
		}
		g, err1 := url.PathUnescape(group)
		n, err2 := url.PathUnescape(tab)
		if err1 == nil && err2 == nil && g != "" && n != "" {
			tabs = append(tabs, CanvasTab{Group: g, Tab: n})
		}
	}
	return tabs
}

// reportKind returns the kind of report the options describe, or "" if they describe none of the kinds this tool saves.
func reportKind(o *drivers.ReportOptions) string {
	switch {
	case o.Canvas != "":
		return "canvas_pdf"
	case o.Resolver == "ai":
		return "ai"
	case o.QueryName == "MetricsViewAggregation":
		return "explore_data"
	default:
		return ""
	}
}

// reportFields flattens a report's options into named fields, to report what an edit changed.
func reportFields(o *drivers.ReportOptions) map[string]any {
	m := map[string]any{}
	setField(m, "display_name", o.DisplayName)
	setField(m, "schedule.cron", o.RefreshCron)
	setField(m, "schedule.time_zone", o.RefreshTimeZone)
	setField(m, "open_mode", o.WebOpenMode)
	recipientsFields(m, o.EmailRecipients, o.SlackUsers, o.SlackChannels, o.SlackWebhooks)
	setField(m, "explore", o.Explore)
	setField(m, "canvas", o.Canvas)
	if o.ExportFormat != runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED {
		m["export.format"] = exportFormatName(o.ExportFormat)
	}
	setField(m, "export.limit", o.ExportLimit)
	setField(m, "export.include_header", o.ExportIncludeHeader)
	setField(m, "export.attach", o.ExportAttach)
	setJSONFields(m, "query", o.QueryArgsJSON)
	setMapFields(m, "resolver_properties", o.ResolverProperties)
	for mv, expr := range o.MetricsViewFilters {
		m["metrics_view_filters."+mv] = protoValue(expr)
	}
	if o.Canvas != "" {
		state, err := url.ParseQuery(o.WebOpenState)
		if err == nil {
			for k := range state {
				m["canvas_state."+k] = state.Get(k)
			}
		}
	} else {
		setField(m, "dashboard_state", o.WebOpenState)
	}
	return m
}

// dataExportFormats are the formats of a data report's export, by the name report YAML gives them, which the export argument takes.
var dataExportFormats = map[string]runtimev1.ExportFormat{
	"csv":     runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	"xlsx":    runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
	"parquet": runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET,
}

// exportFormatName returns the name of an export format as the export argument takes it, or "pdf" for a canvas report's.
func exportFormatName(f runtimev1.ExportFormat) string {
	if f == runtimev1.ExportFormat_EXPORT_FORMAT_PDF {
		return "pdf"
	}
	for name, df := range dataExportFormats {
		if df == f {
			return name
		}
	}
	return f.String()
}
