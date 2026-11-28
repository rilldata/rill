package ai

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/parser"
	"google.golang.org/protobuf/types/known/structpb"
)

const CreateReportName = "create_report"

// CreateReport creates a scheduled report for the current user through the admin service, as the report dialog creates it:
// the report is owned by the user, listed in their reports, and they can delete it on its page, and edit it there unless it's an AI report.
type CreateReport struct {
	Runtime *runtime.Runtime
}

var _ Tool[*CreateReportArgs, *CreatedResource] = (*CreateReport)(nil)

type CreateReportArgs struct {
	DisplayName string `json:"display_name" jsonschema:"Display name of the report."`
	Kind        string `json:"kind" jsonschema:"What the report sends: 'explore_data' an export of an explore dashboard's data (CSV, XLSX or Parquet); 'ai' an analysis written by the AI with its insights."`
	Explore     string `json:"explore,omitempty" jsonschema:"Name of the explore dashboard. Required for 'explore_data'; optional for 'ai', to focus the analysis on it."`
	// Data of explore_data reports
	Measures   []string      `json:"measures,omitempty" jsonschema:"For 'explore_data': measures to export. At least one."`
	Dimensions []string      `json:"dimensions,omitempty" jsonschema:"For 'explore_data': dimensions to break the data down by."`
	Export     *ReportExport `json:"export,omitempty" jsonschema:"For 'explore_data': the export's format and options."`
	// Data of ai reports
	Prompt     string `json:"prompt,omitempty" jsonschema:"For 'ai': what the analysis should focus on. Without it, it's a general analysis of the key metrics."`
	Comparison string `json:"comparison,omitempty" jsonschema:"For 'ai': compare the time range with an earlier period, 'previous_period' (the period of the same length right before), 'previous_day', 'previous_week', 'previous_month', 'previous_quarter' or 'previous_year'. Needs a time_range."`
	// Shared
	TimeRange  string            `json:"time_range,omitempty" jsonschema:"Time range of the data, as a Rill time expression relative to the latest data, for example '7D as of latest/D' for the last 7 complete days. Required for 'explore_data'."`
	Where      []DimensionFilter `json:"where,omitempty" jsonschema:"Filters on dimensions of the explore."`
	Schedule   *Schedule         `json:"schedule" jsonschema:"When to send the report."`
	Recipients *Recipients       `json:"recipients,omitempty" jsonschema:"Who to send it to. Defaults to the user."`
	OpenMode   string            `json:"open_mode,omitempty" jsonschema:"Whose permissions the report opens with: 'recipient' (default) each recipient's own, after logging in; 'creator' the user's own, so every recipient sees the data the user can see, without logging in. Only use 'creator' when the user chooses it knowing this; it's required to send to Slack or to attach the export."`
}

// ReportExport are the export options of an explore data report.
type ReportExport struct {
	Format        string `json:"format,omitempty" jsonschema:"'csv' (default), 'xlsx' or 'parquet'."`
	Limit         uint64 `json:"limit,omitempty" jsonschema:"Maximum number of rows. 0 for no limit."`
	IncludeHeader bool   `json:"include_header,omitempty" jsonschema:"For CSV and XLSX: add a header with the report's name, filters and time range above the data."`
	Attach        bool   `json:"attach,omitempty" jsonschema:"Attach the export to the email instead of only linking to it. Requires open_mode 'creator'. Slack messages only link to it, so when sending to Slack users it also needs an email recipient."`
}

func (t *CreateReport) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  CreateReportName,
		Title: "Create scheduled report",
		Description: "Creates a scheduled report for the current user: an export of an explore dashboard's data, or an analysis written by the AI, sent by email or Slack on a schedule. " +
			"Use it when the user asks to be sent data or a summary regularly, for example 'send me the sales by country every Monday'. " +
			"The report belongs to the user, and they can delete it on its page, and edit an 'explore_data' report there. Only create a file in the project's code instead when the user explicitly asks for that. " +
			confirmBeforeCreating +
			"Reports open with each recipient's own permissions by default (open_mode 'recipient'). " +
			"Sending to Slack users or attaching the export needs open_mode 'creator': when the user asks for either, say so in the preview, explaining that with 'creator' every recipient sees the data with the user's permissions, without logging in, and only use it if they agree. " +
			"This tool can't change reports: to change an 'explore_data' report, give the user the URL of its page from list_reports, where they can edit it; to change an 'ai' report, which its page can't edit, create a new one and delete the old one. " +
			"Give the user the report's URL from the response.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(true), // It notifies people
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating report...",
			"openai/toolInvocation/invoked":  "Created report",
		},
	}
}

func (t *CreateReport) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *CreateReport) Handler(ctx context.Context, args *CreateReportArgs) (*CreatedResource, error) {
	if args.DisplayName == "" {
		return nil, errors.New("invalid input: display_name is required")
	}
	if args.Schedule == nil {
		return nil, errors.New("invalid input: schedule is required")
	}
	if err := checkSchedule(args.Schedule); err != nil {
		return nil, err
	}
	if args.TimeRange != "" {
		if err := checkTimeRange(args.TimeRange); err != nil {
			return nil, err
		}
	}

	opts := &adminv1.ReportOptions{
		DisplayName:     args.DisplayName,
		RefreshCron:     args.Schedule.Cron,
		RefreshTimeZone: args.Schedule.TimeZone,
		WebOpenMode:     args.OpenMode,
	}
	switch args.OpenMode {
	case "":
		opts.WebOpenMode = "recipient"
	case "recipient", "creator":
	default:
		return nil, fmt.Errorf("invalid open_mode %q: use 'recipient' or 'creator'", args.OpenMode)
	}
	if args.Recipients != nil {
		opts.EmailRecipients = args.Recipients.Emails
		opts.SlackUsers = args.Recipients.SlackUsers
		if len(opts.SlackUsers) > 0 && opts.WebOpenMode != "creator" {
			return nil, errors.New("invalid input: sending to Slack needs open_mode 'creator', since Slack recipients can't log in to open the report: ask the user")
		}
	}

	var err error
	switch args.Kind {
	case "explore_data":
		err = t.setExploreData(ctx, args, opts)
	case "ai":
		err = t.setAI(ctx, args, opts)
	default:
		err = fmt.Errorf("invalid kind %q: use 'explore_data' or 'ai'", args.Kind)
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
	name, url, err := admin.CreateReportForUser(ctx, s.Claims().UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}
	return createdResource(ctx, t.Runtime, runtime.ResourceKindReport, name, url)
}

// setExploreData sets the report's data to an export of an explore's data, as the report dialog saves it:
// a MetricsViewAggregation query of the measures broken down by the dimensions, sorted by the first measure.
func (t *CreateReport) setExploreData(ctx context.Context, args *CreateReportArgs, opts *adminv1.ReportOptions) error {
	if args.Prompt != "" || args.Comparison != "" {
		return errors.New("invalid input: prompt and comparison only apply to 'ai' reports")
	}
	explore, err := validExplore(ctx, t.Runtime, args.Explore)
	if err != nil {
		return err
	}
	if len(args.Measures) == 0 {
		return errors.New("invalid input: an 'explore_data' report needs at least one measure")
	}
	if args.TimeRange == "" {
		return errors.New("invalid input: an 'explore_data' report needs a time_range")
	}
	where, whereDims, err := filtersExpression(args.Where)
	if err != nil {
		return err
	}
	if err := checkExploreFields(explore, args.Explore, args.Measures, slices.Concat(args.Dimensions, whereDims)); err != nil {
		return err
	}

	q := &runtimev1.MetricsViewAggregationRequest{
		MetricsView: explore.MetricsView,
		Where:       where,
		TimeRange:   &runtimev1.TimeRange{Expression: args.TimeRange, TimeZone: args.Schedule.TimeZone},
		Sort:        []*runtimev1.MetricsViewAggregationSort{{Name: args.Measures[0], Desc: true}},
	}
	for _, m := range args.Measures {
		q.Measures = append(q.Measures, &runtimev1.MetricsViewAggregationMeasure{Name: m})
	}
	for _, d := range args.Dimensions {
		q.Dimensions = append(q.Dimensions, &runtimev1.MetricsViewAggregationDimension{Name: d})
	}

	export := args.Export
	if export == nil {
		export = &ReportExport{}
	}
	format := export.Format
	if format == "" {
		format = "csv"
	}
	opts.ExportFormat, err = parser.ParseExportFormat(format)
	if err != nil || opts.ExportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED || opts.ExportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_PDF {
		return fmt.Errorf("invalid export format %q: use 'csv', 'xlsx' or 'parquet'", export.Format)
	}
	// Like the report dialog, a Parquet export has no header.
	if export.IncludeHeader && opts.ExportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
		return errors.New("invalid input: a Parquet export can't include a header")
	}
	if export.Attach && opts.WebOpenMode != "creator" {
		return errors.New("invalid input: attaching the export needs open_mode 'creator', since the export is run once for every recipient: ask the user")
	}
	// The report only attaches the export to emails, and the user's email is only the default recipient when there are no others.
	if export.Attach && len(opts.EmailRecipients) == 0 && len(opts.SlackUsers) > 0 {
		return errors.New("invalid input: the export is only attached to emails, and Slack messages link to it: ask the user for an email recipient, or don't attach it")
	}
	opts.ExportLimit = export.Limit
	opts.ExportIncludeHeader = export.IncludeHeader
	opts.ExportAttach = export.Attach

	opts.Explore = args.Explore
	opts.QueryName = "MetricsViewAggregation"         // nolint:staticcheck // The report dialog saves explore reports as queries
	opts.QueryArgsJson, err = aggregationQueryArgs(q) // nolint:staticcheck // The report dialog saves explore reports as queries
	return err
}

// setAI sets the report's data to an analysis by the AI: the properties of the ai resolver, which AI reports run.
func (t *CreateReport) setAI(ctx context.Context, args *CreateReportArgs, opts *adminv1.ReportOptions) error {
	if len(args.Measures) > 0 || len(args.Dimensions) > 0 || args.Export != nil {
		return errors.New("invalid input: measures, dimensions and export only apply to 'explore_data' reports")
	}
	props := map[string]any{}
	if prompt := strings.TrimSpace(args.Prompt); prompt != "" {
		props["prompt"] = prompt
	}
	if args.Explore != "" {
		explore, err := validExplore(ctx, t.Runtime, args.Explore)
		if err != nil {
			return err
		}
		props["explore"] = args.Explore
		opts.Explore = args.Explore
		where, whereDims, err := filtersExpression(args.Where)
		if err != nil {
			return err
		}
		if err := checkExploreFields(explore, args.Explore, nil, whereDims); err != nil {
			return err
		}
		if where != nil {
			props["where"], err = metricsview.NewExpressionFromProto(where).AsMap()
			if err != nil {
				return err
			}
		}
	} else if len(args.Where) > 0 {
		return errors.New("invalid input: filters of an 'ai' report need an explore")
	}
	if args.TimeRange != "" {
		props["time_range"] = map[string]any{"expression": args.TimeRange}
		if args.Schedule.TimeZone != "" {
			props["time_zone"] = args.Schedule.TimeZone
		}
	}
	if args.Comparison != "" {
		if args.TimeRange == "" {
			return errors.New("invalid input: comparing to an earlier period needs a time_range")
		}
		cmp, err := comparisonExpression(args.TimeRange, args.Comparison)
		if err != nil {
			return err
		}
		props["comparison_time_range"] = map[string]any{"expression": cmp}
	}

	var err error
	opts.Resolver = "ai"
	opts.ResolverProperties, err = structpb.NewStruct(props)
	return err
}
