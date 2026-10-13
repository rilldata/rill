package ai

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/chartquery"
	"github.com/rilldata/rill/runtime/ai/mcpapp"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/protobuf/encoding/protojson"
)

const CreateChartName = "create_chart"

// chartWidget identifies the widget mode of the result, shared with the MCP Apps widget.
const chartWidget = "chart"

// chartQueryTimeout bounds the data queries of a chart (the same as query_metrics_view).
const chartQueryTimeout = 2 * time.Minute

// chartMarkdownRows is the number of rows included in the markdown fallback of an MCP result.
const chartMarkdownRows = 20

type CreateChart struct {
	Runtime *runtime.Runtime
}

var _ Tool[CreateChartArgs, *CreateChartResult] = (*CreateChart)(nil)

var _ MCPResultDecorator[*CreateChartResult] = (*CreateChart)(nil)

var _ MCPAppTool = (*CreateChart)(nil)

type CreateChartArgs map[string]any

// CreateChartResult is the result of create_chart.
// The in-app chat renders the chart from the spec and fetches the data itself, so it only receives the first four fields:
// the widget mode, which is always present, plus the chart type, spec and message it received before.
// External MCP clients cannot query Rill, so for them the result also carries the chart-ready data (see ChartDataResult),
// the field metadata needed to format it, the project theme and a link that opens the chart in Rill,
// and the spec is returned with the effective top-N limit of each channel filled in (see specWithLimits).
// Those fields are omitted when empty, so the in-app result keeps its old shape plus widget.
type CreateChartResult struct {
	// Widget selects the mode of the MCP Apps widget bundle.
	Widget    string         `json:"widget"`
	ChartType string         `json:"chart_type"`
	Spec      map[string]any `json:"spec"`
	Message   string         `json:"message"`

	Title       string            `json:"title,omitempty"`
	MetricsView *ChartMetricsView `json:"metrics_view,omitempty"`
	*ChartDataResult
	// Theme is the protojson form of the project's default theme spec, when the project has one.
	Theme   map[string]any `json:"theme,omitempty"`
	OpenURL string         `json:"open_url,omitempty"`
}

// ChartMetricsView summarizes the metrics view a chart is built on.
type ChartMetricsView struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name,omitempty"`
	TimeDimension     string `json:"time_dimension,omitempty"`
	SmallestTimeGrain string `json:"smallest_time_grain,omitempty"`
}

// ChartDataResult is the chart-ready data shared by create_chart (for external clients) and chart_data.
// Rows are keyed by field name; comparison values are under <measure>_prev and temporal values are RFC3339 UTC strings.
// Data is always present when the data was queried, as [] when there are no rows, so that a renderer can tell an empty chart
// from a result without data. It is tagged omitzero rather than omitempty for that: a nil slice (never produced by chartquery)
// is omitted while an empty one is kept, and the inferred output schema keeps it optional, which the MCP server's output validation
// needs for in-app results, whose embedded ChartDataResult is nil.
type ChartDataResult struct {
	TimeGrain                   string                 `json:"time_grain,omitempty"`
	Data                        []map[string]any       `json:"data,omitzero"`
	Fields                      map[string]*ChartField `json:"fields,omitempty"`
	DomainValues                map[string][]any       `json:"domain_values,omitempty"`
	Totals                      map[string]any         `json:"totals,omitempty"`
	ResolvedTimeRange           *chartquery.TimeRange  `json:"resolved_time_range,omitempty"`
	ResolvedComparisonTimeRange *chartquery.TimeRange  `json:"resolved_comparison_time_range,omitempty"`
	// Truncated reports any truncation (top-N or rows); RowsTruncated reports that rows were cut,
	// in which case a temporal chart keeps its newest buckets (see chartquery.Output).
	Truncated         bool   `json:"truncated,omitempty"`
	RowsTruncated     bool   `json:"rows_truncated,omitempty"`
	TruncationWarning string `json:"truncation_warning,omitempty"`
}

// ChartField describes a metrics view field used by a chart, with what a renderer needs to label and format it.
type ChartField struct {
	Name string `json:"name"`
	// Kind is "measure", "dimension" or "time".
	Kind           string         `json:"kind"`
	DisplayName    string         `json:"display_name,omitempty"`
	Description    string         `json:"description,omitempty"`
	Type           string         `json:"type,omitempty"`
	FormatPreset   string         `json:"format_preset,omitempty"`
	FormatD3       string         `json:"format_d3,omitempty"`
	FormatD3Locale map[string]any `json:"format_d3_locale,omitempty"`
	LowerIsBetter  bool           `json:"lower_is_better,omitempty"`
}

func (t *CreateChart) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        CreateChartName,
		Title:       "Create chart",
		Description: createChartDescription,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating chart...",
			"openai/toolInvocation/invoked":  "Created chart",
		},
		InputSchema: chartsInputSchema(),
	}
}

func (t *CreateChart) CheckAccess(ctx context.Context) (bool, error) {
	// Must be able to query metrics and have the chat_charts feature flag.
	// The tool is offered to every client: Rill's own clients render the chart themselves, and external MCP clients get the chart data,
	// which the MCP Apps widget renders when the server serves it (see Session.MCPServer) and other clients get as a markdown table.
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.ReadMetrics) {
		return false, nil
	}
	ff, err := s.FeatureFlags(ctx)
	if err != nil {
		return false, err
	}
	return ff["chat_charts"], nil
}

func (t *CreateChart) Handler(ctx context.Context, args CreateChartArgs) (*CreateChartResult, error) {
	s := GetSession(ctx)

	req, err := parseChartRequest(ctx, t.Runtime, s, args)
	if err != nil {
		return nil, err
	}

	res := &CreateChartResult{
		Widget:    chartWidget,
		ChartType: req.chartType,
		Spec:      req.spec,
		Message:   fmt.Sprintf("Chart created successfully: %s", req.chartType),
	}

	// The in-app chat fetches the data itself from the spec. External clients cannot, so run the chart's queries for them.
	if isRillUserAgent(s) {
		return res, nil
	}

	instance, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	var limits map[string]int
	res.ChartDataResult, limits, err = queryChartData(ctx, t.Runtime, s, instance, req)
	if err != nil {
		return nil, err
	}
	res.Spec = specWithLimits(req.spec, limits)
	res.Title = chartTitle(req.parsed, res.Fields)
	res.MetricsView = &ChartMetricsView{
		Name:          req.parsed.MetricsView,
		DisplayName:   req.mv.DisplayName,
		TimeDimension: req.mv.TimeDimension,
	}
	if req.mv.SmallestTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		res.MetricsView.SmallestTimeGrain = req.mv.SmallestTimeGrain.String()
	}
	res.Theme, err = chartTheme(ctx, t.Runtime, instance)
	if err != nil {
		return nil, err
	}
	res.OpenURL, err = aiMessageOpenURL(instance.FrontendURL, s.ID(), s.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate open URL: %w", err)
	}

	return res, nil
}

// MCPAppUI links create_chart to the MCP Apps widget, which renders the chart from the result.
func (t *CreateChart) MCPAppUI() *mcpapp.ToolUI {
	return &mcpapp.ToolUI{}
}

// DecorateMCPResult renders the chart as markdown for MCP clients that cannot render the widget:
// a title line with the time range, a table of the first rows, the row count and the link to open the chart in Rill.
func (t *CreateChart) DecorateMCPResult(ctx context.Context, out *CreateChartResult, res *mcp.CallToolResult) {
	text := out.Message
	if out.ChartDataResult != nil {
		text = chartMarkdown(out)
	}
	res.Content = []mcp.Content{&mcp.TextContent{Text: text}}
}

// chartRequest is a validated create_chart or chart_data call.
type chartRequest struct {
	chartType string
	spec      map[string]any
	parsed    *chartquery.Spec
	mv        *runtimev1.MetricsViewSpec
}

// chartsInputSchema returns the JSON schema of the chart arguments shared by create_chart and chart_data.
func chartsInputSchema() *jsonschema.Schema {
	var schema *jsonschema.Schema
	err := json.Unmarshal([]byte(metricsview.ChartsJSONSchema), &schema)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal input schema: %w", err))
	}
	return schema
}

// parseChartRequest validates the chart arguments against the metrics view the caller has access to.
func parseChartRequest(ctx context.Context, rt *runtime.Runtime, s *Session, args map[string]any) (*chartRequest, error) {
	chartType, ok := args["chart_type"].(string)
	if !ok || chartType == "" {
		return nil, fmt.Errorf("chart_type is required and must be a string")
	}

	spec, ok := args["spec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("spec is required and must be an object")
	}

	// Validate that metrics_view is specified
	metricsView, ok := spec["metrics_view"].(string)
	if !ok || metricsView == "" {
		return nil, fmt.Errorf("spec must contain a 'metrics_view' field")
	}

	// Validate that time_range is specified
	_, hasTimeRange := spec["time_range"]
	if !hasTimeRange {
		return nil, fmt.Errorf("spec must contain a 'time_range' field with 'start' and 'end' properties")
	}

	// Optional: Validate where clause structure if present
	if whereClause, hasWhere := spec["where"]; hasWhere {
		whereMap, ok := whereClause.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("'where' must be an object with a 'cond' property")
		}
		if _, hasCond := whereMap["cond"]; !hasCond {
			return nil, fmt.Errorf("'where' must contain a 'cond' property with 'op' and 'exprs'")
		}
	}

	// Optional: Validate time_grain if present
	if timeGrain, hasTimeGrain := spec["time_grain"]; hasTimeGrain {
		timeGrainStr, ok := timeGrain.(string)
		if !ok {
			return nil, fmt.Errorf("'time_grain' must be a string (e.g., 'TIME_GRAIN_DAY', 'TIME_GRAIN_MONTH')")
		}
		// Validate it's a valid time grain
		validTimeGrains := []string{
			"TIME_GRAIN_MILLISECOND", "TIME_GRAIN_SECOND", "TIME_GRAIN_MINUTE",
			"TIME_GRAIN_HOUR", "TIME_GRAIN_DAY", "TIME_GRAIN_WEEK",
			"TIME_GRAIN_MONTH", "TIME_GRAIN_QUARTER", "TIME_GRAIN_YEAR",
		}
		if !slices.Contains(validTimeGrains, timeGrainStr) {
			return nil, fmt.Errorf("'time_grain' must be one of: %v", validTimeGrains)
		}
	}

	// Optional: Validate comparison_time_range if present
	if comparisonTimeRange, hasComparisonTimeRange := spec["comparison_time_range"]; hasComparisonTimeRange {
		comparisonMap, ok := comparisonTimeRange.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("'comparison_time_range' must be an object with 'start' and 'end' properties")
		}
		if _, hasStart := comparisonMap["start"]; !hasStart {
			return nil, fmt.Errorf("'comparison_time_range' must contain a 'start' property")
		}
		if _, hasEnd := comparisonMap["end"]; !hasEnd {
			return nil, fmt.Errorf("'comparison_time_range' must contain an 'end' property")
		}

		// Only allow comparison_time_range for cartesian chart types
		validComparisonChartTypes := map[string]bool{
			"bar_chart":              true,
			"line_chart":             true,
			"stacked_bar":            true,
			"stacked_bar_normalized": true,
		}
		if !validComparisonChartTypes[chartType] {
			return nil, fmt.Errorf("'comparison_time_range' is only supported for: bar_chart, line_chart, stacked_bar, stacked_bar_normalized")
		}
	}

	// Validate that the metrics view exists
	ctrl, err := rt.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricsView}, false)
	if err != nil {
		return nil, fmt.Errorf("metrics view %q not found: %w", metricsView, err)
	}
	r, access, err := rt.ApplySecurityPolicy(ctx, s.InstanceID(), s.Claims(), r)
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, fmt.Errorf("resource not found")
	}
	mvState := r.GetMetricsView().State
	if mvState.ValidSpec == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", metricsView)
	}

	// Validate that all field references in the chart spec exist in the metrics view
	if err := validateChartFields(chartType, spec, mvState.ValidSpec); err != nil {
		return nil, fmt.Errorf("field validation failed: %w", err)
	}

	parsed, err := chartquery.ParseSpec(chartType, spec)
	if err != nil {
		return nil, err
	}

	return &chartRequest{
		chartType: chartType,
		spec:      spec,
		parsed:    parsed,
		mv:        mvState.ValidSpec,
	}, nil
}

// queryChartData runs the chart's queries with the session's claims and the instance's AI query limits.
// It returns the chart-ready data and the effective top-N limit of each spec channel (see chartquery.Output.Limits).
// The result always carries the field metadata and the resolved time ranges, even when the queries return no rows,
// so that a renderer can show an empty state for the chart's fields and time range.
func queryChartData(ctx context.Context, rt *runtime.Runtime, s *Session, instance *drivers.Instance, req *chartRequest) (*ChartDataResult, map[string]int, error) {
	cfg, err := instance.Config()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get instance config: %w", err)
	}

	plan, err := chartquery.Build(req.parsed)
	if err != nil {
		return nil, nil, err
	}

	// Apply a timeout across the chart's queries to prevent runaway queries
	ctx, cancel := context.WithTimeout(ctx, chartQueryTimeout)
	defer cancel()

	out, err := chartquery.Execute(ctx, plan, &metricsQueryRunner{
		rt:         rt,
		instanceID: s.InstanceID(),
		claims:     s.Claims(),
		limits: metricsview.QueryLimits{
			RequireTimeRange: cfg.AIRequireTimeRange,
			MaxTimeRangeDays: cfg.AIMaxTimeRangeDays,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return &ChartDataResult{
		TimeGrain:                   out.TimeGrain,
		Data:                        out.Data,
		Fields:                      chartFields(req.parsed, req.mv, out.Types),
		DomainValues:                out.DomainValues,
		Totals:                      out.Totals,
		ResolvedTimeRange:           out.ResolvedTimeRange,
		ResolvedComparisonTimeRange: out.ResolvedComparisonTimeRange,
		Truncated:                   out.Truncated,
		RowsTruncated:               out.RowsTruncated,
		TruncationWarning:           out.TruncationWarning,
	}, out.Limits, nil
}

// specWithLimits returns a copy of the spec with the effective top-N limit of each channel filled in,
// so that a renderer can show the limit that was applied where the spec left it to the default. The input maps are not modified.
func specWithLimits(spec map[string]any, limits map[string]int) map[string]any {
	if len(limits) == 0 {
		return spec
	}
	res := maps.Clone(spec)
	for channel, limit := range limits {
		field, ok := res[channel].(map[string]any)
		if !ok {
			continue
		}
		field = maps.Clone(field)
		field["limit"] = limit
		res[channel] = field
	}
	return res
}

// metricsQueryRunner runs chart queries through the runtime's metrics resolver.
type metricsQueryRunner struct {
	rt         *runtime.Runtime
	instanceID string
	claims     *runtime.SecurityClaims
	limits     metricsview.QueryLimits
}

func (r *metricsQueryRunner) Query(ctx context.Context, q *metricsview.Query) (*chartquery.Result, error) {
	q.QueryLimits = &r.limits
	props, err := q.AsMap()
	if err != nil {
		return nil, err
	}

	res, _, err := r.rt.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         r.instanceID,
		Resolver:           "metrics",
		ResolverProperties: props,
		Claims:             r.claims,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	schema, rows, err := resolverResultToRows(res)
	if err != nil {
		return nil, err
	}
	types := make(map[string]string, len(schema))
	for _, f := range schema {
		types[f.Name] = f.Type
	}

	tr, ctr, err := resolveTimeRanges(res)
	if err != nil {
		return nil, err
	}

	return &chartquery.Result{
		Rows:                rows,
		Types:               types,
		TimeRange:           chartTimeRange(tr),
		ComparisonTimeRange: chartTimeRange(ctr),
	}, nil
}

func chartTimeRange(tr *metricsview.TimeRange) *chartquery.TimeRange {
	if tr == nil {
		return nil
	}
	return &chartquery.TimeRange{Start: tr.Start, End: tr.End}
}

// chartFields describes the metrics view fields the chart references, with the types of the result columns.
func chartFields(parsed *chartquery.Spec, mv *runtimev1.MetricsViewSpec, types map[string]string) map[string]*ChartField {
	temporal := make(map[string]bool)
	for _, f := range []*chartquery.Field{parsed.X, parsed.Y, parsed.Dimension} {
		if f != nil && f.Type == chartquery.TypeTemporal {
			temporal[f.Field] = true
		}
	}

	fields := make(map[string]*ChartField)
	for _, name := range parsed.Fields() {
		field := &ChartField{Name: name, Type: types[name]}
		if i := slices.IndexFunc(mv.Measures, func(m *runtimev1.MetricsViewSpec_Measure) bool { return m.Name == name }); i >= 0 {
			m := mv.Measures[i]
			field.Kind = "measure"
			field.DisplayName = m.DisplayName
			field.Description = m.Description
			field.FormatPreset = m.FormatPreset
			field.FormatD3 = m.FormatD3
			field.LowerIsBetter = m.LowerIsBetter
			if m.FormatD3Locale != nil {
				field.FormatD3Locale = m.FormatD3Locale.AsMap()
			}
			if field.Type == "" && m.DataType != nil {
				field.Type = strings.TrimPrefix(m.DataType.Code.String(), "CODE_")
			}
		} else if i := slices.IndexFunc(mv.Dimensions, func(d *runtimev1.MetricsViewSpec_Dimension) bool { return d.Name == name }); i >= 0 {
			d := mv.Dimensions[i]
			field.Kind = "dimension"
			field.DisplayName = d.DisplayName
			field.Description = d.Description
			if field.Type == "" && d.DataType != nil {
				field.Type = strings.TrimPrefix(d.DataType.Code.String(), "CODE_")
			}
		} else {
			continue
		}
		// The time dimension is always among the dimensions (the parser adds it), so it is found above and retyped here.
		if temporal[name] || name == mv.TimeDimension {
			field.Kind = "time"
			if field.Type == "" {
				field.Type = "TIMESTAMP"
			}
		}
		if field.DisplayName == "" {
			field.DisplayName = name
		}
		fields[name] = field
	}
	return fields
}

// chartTitle builds the chart title the same way the frontend chart providers do.
func chartTitle(parsed *chartquery.Spec, fields map[string]*ChartField) string {
	label := func(f *chartquery.Field) string {
		if f == nil || f.Field == "" {
			return ""
		}
		if f.Type == chartquery.TypeTemporal {
			return "Time"
		}
		if field, ok := fields[f.Field]; ok && field.DisplayName != "" {
			return field.DisplayName
		}
		return f.Field
	}
	labels := func(names []string) string {
		res := make([]string, len(names))
		for i, n := range names {
			res[i] = n
			if field, ok := fields[n]; ok && field.DisplayName != "" {
				res[i] = field.DisplayName
			}
		}
		return strings.Join(res, ", ")
	}
	preposition := func(x *chartquery.Field, over, other string) string {
		if x != nil && x.Type == chartquery.TypeTemporal {
			return over
		}
		return other
	}

	switch parsed.ChartType {
	case chartquery.BarChart, chartquery.LineChart, chartquery.AreaChart, chartquery.StackedBar, chartquery.StackedBarNormalized:
		if parsed.Y != nil && len(parsed.Y.Fields) > 0 {
			return fmt.Sprintf("%s %s %s", labels(parsed.Y.Fields), preposition(parsed.X, "over", "by"), label(parsed.X))
		}
		title := fmt.Sprintf("%s %s %s", label(parsed.Y), preposition(parsed.X, "over", "per"), label(parsed.X))
		if parsed.Color != nil && parsed.Color.Field != "" && parsed.Color.Field != chartquery.MeasuresField {
			title += " split by " + label(parsed.Color)
		}
		return title
	case chartquery.DonutChart, chartquery.PieChart:
		if parsed.Color != nil && parsed.Color.Field != "" {
			return fmt.Sprintf("%s split by %s", label(parsed.Measure), label(parsed.Color))
		}
		return label(parsed.Measure)
	case chartquery.FunnelChart:
		if parsed.Measure == nil || (parsed.Measure.Field == "" && len(parsed.Measure.Fields) == 0) {
			return "Funnel"
		}
		if parsed.BreakdownMode == "measures" || len(parsed.Measure.Fields) > 0 {
			names := parsed.Measure.Fields
			if len(names) == 0 {
				names = []string{parsed.Measure.Field}
			}
			return labels(names) + " funnel"
		}
		if parsed.Stage != nil && parsed.Stage.Field != "" {
			return fmt.Sprintf("%s funnel by %s", label(parsed.Measure), label(parsed.Stage))
		}
		return label(parsed.Measure)
	case chartquery.Heatmap:
		return fmt.Sprintf("%s by %s and %s", label(parsed.Color), label(parsed.X), label(parsed.Y))
	case chartquery.ComboChart:
		measures := label(parsed.Y1)
		if l := label(parsed.Y2); l != "" {
			measures += " & " + l
		}
		return fmt.Sprintf("%s %s %s", measures, preposition(parsed.X, "over", "per"), label(parsed.X))
	case chartquery.ScatterPlot:
		title := fmt.Sprintf("%s vs %s", label(parsed.X), label(parsed.Y))
		if l := label(parsed.Dimension); l != "" {
			title += " for " + l
		}
		if l := label(parsed.Color); l != "" {
			title += " split by " + l
		}
		if l := label(parsed.Size); l != "" {
			title += " sized by " + l
		}
		return title
	}
	return ""
}

// chartTheme returns the project's default theme spec as protojson, or nil when the project has none.
func chartTheme(ctx context.Context, rt *runtime.Runtime, instance *drivers.Instance) (map[string]any, error) {
	if instance.Theme == "" {
		return nil, nil
	}

	ctrl, err := rt.Controller(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: instance.Theme}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			// A dangling theme reference is a project misconfiguration that should not fail the chart.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get theme %q: %w", instance.Theme, err)
	}
	spec := r.GetTheme().GetSpec()
	if spec == nil {
		return nil, nil
	}

	data, err := protojson.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal theme: %w", err)
	}
	var theme map[string]any
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("failed to unmarshal theme: %w", err)
	}
	return theme, nil
}

// chartMarkdown renders a chart result as markdown: a bold title with the time range, a table of the first rows,
// the row count and the link to open the chart in Rill.
func chartMarkdown(out *CreateChartResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "**%s: %s**", chartTypeLabel(out.ChartType), out.Title)
	if tr := out.ResolvedTimeRange; tr != nil {
		fmt.Fprintf(&b, " (%s to %s", formatChartTime(tr.Start), formatChartTime(tr.End))
		if ctr := out.ResolvedComparisonTimeRange; ctr != nil {
			fmt.Fprintf(&b, ", vs %s to %s", formatChartTime(ctr.Start), formatChartTime(ctr.End))
		}
		b.WriteString(")")
	}
	b.WriteString("\n\n")

	if len(out.Data) > 0 {
		columns := chartColumns(out)
		for i, c := range columns {
			if i > 0 {
				b.WriteString(" | ")
			} else {
				b.WriteString("| ")
			}
			b.WriteString(escapeMarkdownCell(c.title))
		}
		b.WriteString(" |\n|")
		for range columns {
			b.WriteString(" - |")
		}
		b.WriteString("\n")
		for _, row := range out.Data[:min(len(out.Data), chartMarkdownRows)] {
			b.WriteString("|")
			for _, c := range columns {
				b.WriteString(" ")
				b.WriteString(escapeMarkdownCell(formatChartValue(row[c.name])))
				b.WriteString(" |")
			}
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "\n_%d of %d rows shown_", min(len(out.Data), chartMarkdownRows), len(out.Data))
	} else {
		b.WriteString("_No data for this chart._")
	}
	if out.TruncationWarning != "" {
		b.WriteString("\n\n")
		b.WriteString(out.TruncationWarning)
	}
	if out.OpenURL != "" {
		fmt.Fprintf(&b, "\n\nOpen the interactive chart in Rill: %s", out.OpenURL)
	}

	return b.String()
}

type chartColumn struct {
	name  string
	title string
}

// chartColumns orders the result columns for the markdown table using the field metadata the result carries:
// time fields, then dimensions, then measures, each measure followed by its comparison value, and each group sorted by name;
// columns without field metadata come last.
func chartColumns(out *CreateChartResult) []chartColumn {
	rank := func(name string) int {
		switch f := out.Fields[name]; {
		case f == nil:
			return 3
		case f.Kind == "time":
			return 0
		case f.Kind == "dimension":
			return 1
		default:
			return 2
		}
	}
	names := slices.Sorted(maps.Keys(out.Data[0]))
	slices.SortStableFunc(names, func(a, b string) int { return cmp.Compare(rank(a), rank(b)) })

	var columns []chartColumn
	for _, name := range names {
		if base, ok := strings.CutSuffix(name, chartquery.PrevSuffix); ok && out.Fields[base] != nil {
			continue // Added right after its measure below
		}
		title := name
		if f := out.Fields[name]; f != nil && f.DisplayName != "" {
			title = f.DisplayName
		}
		columns = append(columns, chartColumn{name: name, title: title})
		if prev := name + chartquery.PrevSuffix; out.Fields[name] != nil && slices.Contains(names, prev) {
			columns = append(columns, chartColumn{name: prev, title: title + " (prev)"})
		}
	}
	return columns
}

func chartTypeLabel(chartType string) string {
	label := strings.ReplaceAll(chartType, "_", " ")
	if label == "" {
		return label
	}
	return strings.ToUpper(label[:1]) + label[1:]
}

func formatChartTime(t time.Time) string {
	t = t.UTC()
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format(time.DateOnly)
	}
	return t.Format(time.RFC3339)
}

func formatChartValue(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeMarkdownCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	return strings.ReplaceAll(s, "\n", " ")
}

// validateChartFields validates that all field references in the chart spec exist in the metrics view
func validateChartFields(chartType string, spec map[string]any, mvSpec *runtimev1.MetricsViewSpec) error {
	// Build a map of available fields (dimensions and measures)
	availableFields := make(map[string]bool)

	// Add time dimension if present
	if mvSpec.TimeDimension != "" {
		availableFields[mvSpec.TimeDimension] = true
	}

	// Add dimensions
	for _, dim := range mvSpec.Dimensions {
		if dim.Name != "" {
			availableFields[dim.Name] = true
		}
	}

	// Add measures
	for _, measure := range mvSpec.Measures {
		if measure.Name != "" {
			availableFields[measure.Name] = true
		}
	}

	// Validate fields based on chart type
	switch chartType {
	case "heatmap":
		if colorField, ok := pathutil.GetPath(spec, "color.field"); ok {
			if err := validateField(availableFields, colorField); err != nil {
				return fmt.Errorf("invalid color field: %w", err)
			}
		}
		if xField, ok := pathutil.GetPath(spec, "x.field"); ok {
			if err := validateField(availableFields, xField); err != nil {
				return fmt.Errorf("invalid x field: %w", err)
			}
		}
		if yField, ok := pathutil.GetPath(spec, "y.field"); ok {
			if err := validateField(availableFields, yField); err != nil {
				return fmt.Errorf("invalid y field: %w", err)
			}
		}

	case "funnel_chart":
		if stageField, ok := pathutil.GetPath(spec, "stage.field"); ok {
			if err := validateField(availableFields, stageField); err != nil {
				return fmt.Errorf("invalid stage field: %w", err)
			}
		}
		if measureField, ok := pathutil.GetPath(spec, "measure.field"); ok {
			if err := validateField(availableFields, measureField); err != nil {
				return fmt.Errorf("invalid measure field: %w", err)
			}
		}
		// Validate fields array if present
		if fields, ok := pathutil.GetPath(spec, "measure.fields"); ok {
			if err := validateFieldsArray(availableFields, fields); err != nil {
				return fmt.Errorf("invalid measure fields: %w", err)
			}
		}
		// Optional enum-valued funnel fields. Cross-validation between breakdownMode
		// and color (e.g. color "stage" requires dimension mode) is enforced by the
		// chart renderer, which falls back to sensible defaults; we only reject typos here.
		if breakdownMode, ok := spec["breakdownMode"]; ok {
			if err := validateEnum("breakdownMode", breakdownMode, []string{"dimension", "measures"}); err != nil {
				return err
			}
		}
		if mode, ok := spec["mode"]; ok {
			if err := validateEnum("mode", mode, []string{"width", "order"}); err != nil {
				return err
			}
		}
		if color, ok := spec["color"]; ok {
			if err := validateEnum("color", color, []string{"stage", "measure", "name", "value"}); err != nil {
				return err
			}
		}
		if percentMode, ok := spec["percentMode"]; ok {
			if err := validateEnum("percentMode", percentMode, []string{"top", "previous"}); err != nil {
				return err
			}
		}

	case "donut_chart", "pie_chart":
		if colorField, ok := pathutil.GetPath(spec, "color.field"); ok {
			if err := validateField(availableFields, colorField); err != nil {
				return fmt.Errorf("invalid color field: %w", err)
			}
		}
		if measureField, ok := pathutil.GetPath(spec, "measure.field"); ok {
			if err := validateField(availableFields, measureField); err != nil {
				return fmt.Errorf("invalid measure field: %w", err)
			}
		}

	case "bar_chart", "line_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		if colorField, ok := pathutil.GetPath(spec, "color.field"); ok {
			// Skip validation for special field "rill_measures"
			if fieldStr, ok := colorField.(string); !ok || fieldStr != "rill_measures" {
				if err := validateField(availableFields, colorField); err != nil {
					return fmt.Errorf("invalid color field: %w", err)
				}
			}
		}
		if xField, ok := pathutil.GetPath(spec, "x.field"); ok {
			if err := validateField(availableFields, xField); err != nil {
				return fmt.Errorf("invalid x field: %w", err)
			}
		}
		if yField, ok := pathutil.GetPath(spec, "y.field"); ok {
			if err := validateField(availableFields, yField); err != nil {
				return fmt.Errorf("invalid y field: %w", err)
			}
		}
		// Validate multi-measure fields arrays if present. Horizontal bar charts carry the measures on x.
		if fields, ok := pathutil.GetPath(spec, "y.fields"); ok {
			if err := validateFieldsArray(availableFields, fields); err != nil {
				return fmt.Errorf("invalid y fields: %w", err)
			}
		}
		if fields, ok := pathutil.GetPath(spec, "x.fields"); ok {
			if err := validateFieldsArray(availableFields, fields); err != nil {
				return fmt.Errorf("invalid x fields: %w", err)
			}
		}

	case "combo_chart":
		// For combo_chart, color.type must be "value"
		if colorType, ok := pathutil.GetPath(spec, "color.type"); ok {
			if typeStr, ok := colorType.(string); !ok || typeStr != "value" {
				return fmt.Errorf("combo_chart color type must be 'value', got %q", colorType)
			}
		}
		if xField, ok := pathutil.GetPath(spec, "x.field"); ok {
			if err := validateField(availableFields, xField); err != nil {
				return fmt.Errorf("invalid x field: %w", err)
			}
		}
		if y1Field, ok := pathutil.GetPath(spec, "y1.field"); ok {
			if err := validateField(availableFields, y1Field); err != nil {
				return fmt.Errorf("invalid y1 field: %w", err)
			}
		}
		if y2Field, ok := pathutil.GetPath(spec, "y2.field"); ok {
			if err := validateField(availableFields, y2Field); err != nil {
				return fmt.Errorf("invalid y2 field: %w", err)
			}
		}

	case "scatter_plot":
		if xField, ok := pathutil.GetPath(spec, "x.field"); ok {
			if err := validateField(availableFields, xField); err != nil {
				return fmt.Errorf("invalid x field: %w", err)
			}
		}
		if yField, ok := pathutil.GetPath(spec, "y.field"); ok {
			if err := validateField(availableFields, yField); err != nil {
				return fmt.Errorf("invalid y field: %w", err)
			}
		}
		if dimensionField, ok := pathutil.GetPath(spec, "dimension.field"); ok {
			if err := validateField(availableFields, dimensionField); err != nil {
				return fmt.Errorf("invalid dimension field: %w", err)
			}
		}
		if sizeField, ok := pathutil.GetPath(spec, "size.field"); ok {
			if err := validateField(availableFields, sizeField); err != nil {
				return fmt.Errorf("invalid size field: %w", err)
			}
		}
		if colorField, ok := pathutil.GetPath(spec, "color.field"); ok {
			if err := validateField(availableFields, colorField); err != nil {
				return fmt.Errorf("invalid color field: %w", err)
			}
		}
	}

	return nil
}

// validateField checks if a single field exists in the available fields
func validateField(availableFields map[string]bool, field any) error {
	fieldStr, ok := field.(string)
	if !ok {
		return fmt.Errorf("field is not a string")
	}
	if fieldStr == "" {
		return nil
	}

	if !availableFields[fieldStr] {
		return fmt.Errorf("field %q not found in metrics view", fieldStr)
	}
	return nil
}

// validateEnum checks that value is a string and one of the allowed options.
func validateEnum(name string, value any, allowed []string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q must be a string", name)
	}
	for _, a := range allowed {
		if str == a {
			return nil
		}
	}
	return fmt.Errorf("%q must be one of %v, got %q", name, allowed, str)
}

// validateFieldsArray validates an array of field names
func validateFieldsArray(availableFields map[string]bool, fields any) error {
	fieldsArray, ok := fields.([]any)
	if !ok {
		return fmt.Errorf("fields is not an array")
	}

	for i, field := range fieldsArray {
		if err := validateField(availableFields, field); err != nil {
			return fmt.Errorf("field at index %d: %w", i, err)
		}
	}
	return nil
}

const createChartDescription = `# Chart Visualization Tool

Create visualization charts based on metrics views. This tool generates chart specifications that will be rendered in the chat interface.

Clients that cannot render charts (for example terminals or IDE chats) receive the chart's data as a table together with a link. Whenever the result includes an ` + "`open_url`" + `, surface it to the user so they can open the interactive chart in Rill.

## Required Parameters

All chart specifications must include:
- ` + "`chart_type`" + `: The type of visualization to create
- ` + "`spec`" + `: The chart specification object containing:
  - ` + "`metrics_view`" + `: The name of the metrics view to query
  - ` + "`time_range`" + `: **Required** - Time range for the data query
    - ` + "`start`" + `: ISO 8601 timestamp (inclusive)
    - ` + "`end`" + `: ISO 8601 timestamp (exclusive)
    - ` + "`time_zone`" + `: Optional time zone (defaults to "UTC")
    - Example: ` + "`\"start\": \"2024-01-01T00:00:00Z\", \"end\": \"2024-12-31T23:59:59Z\"`" + `

## Optional Parameters

- ` + "`time_grain`" + `: Time granularity for temporal aggregations (e.g., "TIME_GRAIN_DAY", "TIME_GRAIN_MONTH", "TIME_GRAIN_YEAR"). Defaults to "TIME_GRAIN_DAY" if not specified.
- ` + "`where`" + `: Filter expression to apply to the underlying data. Use the same structure as in query_metrics_view.
- ` + "`comparison_time_range`" + `: Optional time range for period-over-period comparison. Only supported for cartesian chart types: bar_chart, line_chart, stacked_bar, stacked_bar_normalized.
  - ` + "`start`" + `: ISO 8601 timestamp (inclusive) for the comparison period start
  - ` + "`end`" + `: ISO 8601 timestamp (exclusive) for the comparison period end
  - The comparison period should ideally match the duration of the primary time_range for meaningful comparisons
  - Example: To compare H2 2024 vs H1 2024, set time_range to Jul-Dec 2024 and comparison_time_range to Jan-Jun 2024

### Where Expression Structure
The where clause follows this structure:
` + "```json" + `
{
  "cond": {
    "op": "and",  // or "or", "eq", "neq", "in", "nin", "lt", "lte", "gt", "gte", "ilike", "nilike"
    "exprs": [
      {
        "cond": {
          "op": "eq",
          "exprs": [
            {"name": "dimension_name"},
            {"val": "value"}
          ]
        }
      }
    ]
  }
}
` + "```" + `

Example with country filter:
` + "```json" + `
{
  "where": {
    "cond": {
      "op": "in",
      "exprs": [
        {"name": "country"},
        {"val": ["US", "CA", "GB"]}
      ]
    }
  }
}
` + "```" + `

## Supported Chart Types

### 1. Bar Chart (` + "`bar_chart`" + `)
**Use for:** Comparing values across different categories

Example Specification: Plotting a bar chart of the top 20 advertisers by total bids.

Field details:
bids_metrics: metrics_view
advertiser_name: dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": "primary",
    "x": {
      "field": "advertiser_name",
      "limit": 20,
      "showNull": true,
      "type": "nominal",
      "sort": "-y"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

Example of a horizontal bar chart: the same chart with bars running left to right. ` + "`x`" + ` and ` + "`y`" + ` always name the field drawn on that axis, so put the measure (quantitative) under ` + "`x`" + ` and the dimension under ` + "`y`" + `. Sort values refer to axes, so ` + "`\"sort\": \"-x\"`" + ` on ` + "`y`" + ` sorts the categories by the measure. Prefer this for long category labels or many categories.

` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": "primary",
    "x": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    },
    "y": {
      "field": "advertiser_name",
      "limit": 20,
      "type": "nominal",
      "sort": "-x"
    }
  }
}
` + "```" + `

Example with filters: Bar chart showing top advertisers in specific countries

Field details:
bids_metrics: metrics_view
advertiser_name: dimension
country: dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "where": {
      "cond": {
        "op": "in",
        "exprs": [
          {"name": "country"},
          {"val": ["US", "CA", "GB"]}
        ]
      }
    },
    "color": "primary",
    "x": {
      "field": "advertiser_name",
      "limit": 20,
      "type": "nominal",
      "sort": "-y"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

Example with time comparison: Bar chart comparing current period to previous period

Field details:
bids_metrics: metrics_view
advertiser_name: dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-07-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "comparison_time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-06-30T23:59:59Z"
    },
    "color": "primary",
    "x": {
      "field": "advertiser_name",
      "limit": 20,
      "type": "nominal",
      "sort": "-y"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

Example with delta sorting: Bar chart showing top advertisers by biggest change in bids (period-over-period). Use delta sorting when the user wants to see what changed most between two periods.

Field details:
bids_metrics: metrics_view
advertiser_name: dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-07-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "comparison_time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-06-30T23:59:59Z"
    },
    "color": "primary",
    "x": {
      "field": "advertiser_name",
      "limit": 20,
      "type": "nominal",
      "sort": "-y_delta"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 2. Line Chart (` + "`line_chart`" + `)
**Use for:** Showing trends over time

Example Specification: Line chart with monthly aggregation

Field details:
bids_metrics: metrics_view
device_os: dimension
__time: timestamp dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "line_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "time_grain": "TIME_GRAIN_MONTH",
    "color": {
      "field": "device_os",
      "limit": 3,
      "type": "nominal"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "sort": "-y",
      "type": "temporal"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

Example with filters and time grain: Daily trends for specific device types

Field details:
bids_metrics: metrics_view
device_os: dimension
__time: timestamp dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "line_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "time_grain": "TIME_GRAIN_DAY",
    "where": {
      "cond": {
        "op": "in",
        "exprs": [
          {"name": "device_os"},
          {"val": ["iOS", "Android"]}
        ]
      }
    },
    "color": {
      "field": "device_os",
      "type": "nominal"
    },
    "x": {
      "field": "__time",
      "type": "temporal"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

Example with time comparison: Line chart comparing current month trend to previous month

Field details:
bids_metrics: metrics_view
__time: timestamp dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "line_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-06-01T00:00:00Z",
      "end": "2024-06-30T23:59:59Z"
    },
    "comparison_time_range": {
      "start": "2024-05-01T00:00:00Z",
      "end": "2024-05-31T23:59:59Z"
    },
    "time_grain": "TIME_GRAIN_DAY",
    "color": "primary",
    "x": {
      "field": "__time",
      "type": "temporal"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 3. Area Chart (` + "`area_chart`" + `)
**Use for:** Showing magnitude of change over time with filled areas

Example Specification

Field details:
auction_metrics: metrics_view
app_or_site: dimension
__time: timestamp dimension
requests: measure

` + "```json" + `
{
  "chart_type": "area_chart",
  "spec": {
    "metrics_view": "auction_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "app_or_site",
      "type": "nominal"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "showNull": true,
      "type": "temporal"
    },
    "y": {
      "field": "requests",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 4. Stacked Bar Chart (` + "`stacked_bar`" + `)
**Use for:** Showing multiple data series stacked on top of each other.

Example Specification

Field details:
bids_metrics: metrics_view
rill_measures: special field
__time: timestamp dimension
clicks, video_starts, video_completes, ctr, ecpm, impressions: measures


` + "```json" + `
{
  "chart_type": "stacked_bar",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "rill_measures",
      "legendOrientation": "top",
      "type": "value"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "type": "temporal"
    },
    "y": {
      "field": "clicks",
      "fields": [
        "video_starts",
        "video_completes",
        "ctr",
        "clicks",
        "ecpm",
        "impressions"
      ],
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

**IMPORTANT** : The chart types bar_chart, area_chart, line_chart and stacked_bar follow the same schema definition.
Note that when charting out multiple fields using "fields" key, you must also add a "field" key with value being the first field in fields array

Horizontal bars: bar_chart, stacked_bar and stacked_bar_normalized may put the measure on ` + "`x`" + ` (` + "`\"type\": \"quantitative\"`" + `, with ` + "`fields`" + ` for multiple measures) and the dimension on ` + "`y`" + `; the bars then run left to right. Sort values name axes, so use ` + "`\"sort\": \"-x\"`" + ` on ` + "`y`" + ` to order categories by the measure. line_chart and area_chart always keep the dimension on x.


### 5. Normalized Stacked Bar Chart (` + "`stacked_bar_normalized`" + `)
**Use for:** Showing proportions instead of absolute values (100% stacked)

Example Specification

Field details:
rill_commits_metrics: metrics_view
username: dimension
date: timestamp dimension
number_of_commits: measure

` + "```json" + `
{
  "chart_type": "stacked_bar_normalized",
  "spec": {
    "metrics_view": "rill_commits_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "username",
      "limit": 3,
      "type": "nominal"
    },
    "x": {
      "field": "date",
      "limit": 20,
      "type": "temporal"
    },
    "y": {
      "field": "number_of_commits",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 6. Donut Chart (` + "`donut_chart`" + `)
**Use for:** Displaying data as segments of a circle with a hollow center

Example Specification

Field details:
rill_commits_metrics: metrics_view
username: dimension
number_of_commits: measure

` + "```json" + `
{
  "chart_type": "donut_chart",
  "spec": {
    "metrics_view": "rill_commits_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "username",
      "limit": 20,
      "type": "nominal"
    },
    "innerRadius": 50,
    "measure": {
      "field": "number_of_commits",
      "type": "quantitative",
      "showTotal": true
    }
  }
}
` + "```" + `

### 7. Funnel Chart (` + "`funnel_chart`" + `)
**Use for:** Showing flow through a process with decreasing values at each stage or measure

**Funnel-specific fields:**
- ` + "`breakdownMode`" + ` (required): ` + "`\"dimension\"`" + ` (one measure split by a stage dimension) or ` + "`\"measures\"`" + ` (multiple measures, one per stage)
- ` + "`mode`" + `: ` + "`\"width\"`" + ` (bar width proportional to value, default) or ` + "`\"order\"`" + ` (bars shrink by fixed ratio in rank order)
- ` + "`color`" + `:
  - In ` + "`dimension`" + ` mode: ` + "`\"stage\"`" + ` (distinct color per stage) or ` + "`\"measure\"`" + ` (gradient by value)
  - In ` + "`measures`" + ` mode: ` + "`\"name\"`" + ` (distinct color per measure) or ` + "`\"value\"`" + ` (gradient by value)
- ` + "`percentMode`" + `: ` + "`\"top\"`" + ` (default; on-bar % is relative to the top stage) or ` + "`\"previous\"`" + ` (on-bar % is relative to the prior stage). The tooltip always shows both, regardless of this setting. Choose ` + "`\"previous\"`" + ` when the user asks about stage-to-stage drop-off or step conversion; choose ` + "`\"top\"`" + ` when they ask about overall conversion from the entry point.

Example Specification with 1 dimension and 1 measure breakdown

Field details:
Funnel_Dataset_metrics: metrics_view
stage: dimension
total_users_measure: measure

` + "```json" + `
{
  "chart_type": "funnel_chart",
  "spec": {
    "metrics_view": "Funnel_Dataset_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "breakdownMode": "dimension",
    "color": "stage",
    "measure": {
      "field": "total_users_measure",
      "type": "quantitative"
    },
    "mode": "width",
    "percentMode": "top",
    "stage": {
      "field": "stage",
      "limit": 15,
      "type": "nominal"
    }
  }
}
` + "```" + `

Example Specification with multiple measures breakdown and stage-to-stage percentages

Field details:
bids: metrics_view
impressions, video_starts, video_completes: measures

` + "```json" + `
{
  "chart_type": "funnel_chart",
  "spec": {
    "breakdownMode": "measures",
		"time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": "name",
    "measure": {
      "field": "impressions",
      "type": "quantitative",
      "fields": [
        "impressions",
        "video_starts",
        "video_completes"
      ]
    },
    "metrics_view": "bids",
    "mode": "width",
    "percentMode": "previous"
  }
}
` + "```" + `

### 8. Heat Map (` + "`heatmap`" + `)
**Use for:** Visualizing data density using color intensity across two dimensions

Example Specification

Field details:
bids_metrics: metrics_view
day: dimension
hour: dimension
total_bids: measure

` + "```json" + `
{
  "chart_type": "heatmap",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "total_bids",
      "type": "quantitative"
    },
    "x": {
      "field": "day",
      "limit": 10,
      "type": "nominal",
      "sort": [
        "Sunday",
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday"
      ]
    },
    "y": {
      "field": "hour",
      "limit": 24,
      "type": "nominal",
      "sort": "-color"
    }
  }
}
` + "```" + `

### 9. Combo Chart (` + "`combo_chart`" + `)
**Use for:** Combining different chart types (like bars and lines) in a single visualization

Example Specification

Field details:
auction_metrics: metrics_view
__time: timestamp dimension
date: timestamp dimension
1d_qps: measure
requests: measure
rill_measures: special field

` + "```json" + `
{
  "chart_type": "combo_chart",
  "spec": {
    "metrics_view": "auction_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "rill_measures",
      "legendOrientation": "top",
      "type": "value"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "type": "temporal"
    },
    "y1": {
      "field": "1d_qps",
      "mark": "bar",
      "type": "quantitative",
      "zeroBasedOrigin": true
    },
    "y2": {
      "field": "requests",
      "mark": "line",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 10. Scatter Plot (` + "`scatter_plot`" + `)
**Use for:** Visualizing the relationship between two measures, with each point representing a data value. Points can be optionally colored by a dimension and sized by a measure. Useful for correlation analysis and identifying outliers across two numeric metrics.

**Scatter-specific fields:**
- ` + "`x`" + ` (required): Measure on the x-axis. Must be ` + "`type: quantitative`" + `.
- ` + "`y`" + ` (required): Measure on the y-axis. Must be ` + "`type: quantitative`" + `.
- ` + "`dimension`" + ` (optional): Nominal dimension; each distinct value renders as a separate point.
- ` + "`size`" + ` (optional): Measure that controls the point radius. Must be ` + "`type: quantitative`" + `.
- ` + "`color`" + ` (optional): Either a color string or a nominal dimension field object to color points by.

Example Specification

Field details:
bids_metrics: metrics_view
advertiser_name: dimension
device_os: dimension
total_bids: measure
total_clicks: measure
total_impressions: measure

` + "```json" + `
{
  "chart_type": "scatter_plot",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "x": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": false
    },
    "y": {
      "field": "total_clicks",
      "type": "quantitative",
      "zeroBasedOrigin": false
    },
    "dimension": {
      "field": "advertiser_name",
      "limit": 10,
      "type": "nominal"
    },
    "size": {
      "field": "total_impressions",
      "type": "quantitative"
    },
    "color": {
      "field": "device_os",
      "limit": 5,
      "type": "nominal",
      "legendOrientation": "top"
    }
  }
}
` + "```" + `

## Field Type Definitions

### Data Types
- **nominal**: Categorical data (e.g., categories, names, labels), use for dimensions
- **temporal**: Time-based data (dates, timestamps), use for time dimensions and timestamps
- **quantitative**: Numerical data (counts, amounts, measurements), use for measures
- **value**: Special type for multiple measures (used in color field)

### Common Field Properties
- **field**: The field name from the metrics view
- **type**: Data type (nominal, temporal, quantitative, value)
- **limit**: Maximum number of values to display for selected sort mode
- **showNull**: Include null values in the visualization (true/false)
- **sort**: Sorting order
  - ` + "`\"x\"`" + ` or ` + "`\"-x\"`" + `: Sort by x-axis values (ascending/descending)
  - ` + "`\"y\"`" + ` or ` + "`\"-y\"`" + `: Sort by y-axis values (ascending/descending)
  - ` + "`\"y_delta\"`" + ` or ` + "`\"-y_delta\"`" + `: Sort by absolute change (delta) between current and comparison period (ascending/descending). Only effective when ` + "`comparison_time_range`" + ` is provided; otherwise falls back to regular y-axis sort. Use this when the user asks about "what changed most", "biggest increase/decrease", or "top movers". For a horizontal bar chart (measure on x) use ` + "`\"x_delta\"`" + ` or ` + "`\"-x_delta\"`" + ` on the y field instead.
	- ` + "`\"color\"`" + ` or ` + "`\"-color\"`" + `: Sort by color field values (ascending/descending) Only used for heatmap charts
	- ` + "`\"measure\"`" + ` or ` + "`\"-measure\"`" + `: Sort by measure field values (ascending/descending) Only used for donut charts
  - Array of values for custom sort order (e.g., weekday names)
- **zeroBasedOrigin**: Start y-axis from zero (true/false)
- **showTotal**: Displays the measure total without any breakdown. Only used for donut chart to display totals in center

### Special Fields
- **rill_measures**: Special field for multiple measures in stacked charts and area charts. The field name is only used in color field object. DO NOT USE it for other keys except for "color" key in the field object.

## Color Configuration

Colors can be specified in three ways depending on the chart type and requirements:

### 1. Single Color String
For bar_chart, stacked_bar, line_chart, and area_chart types in single measure mode and only 1 dimensions is involved. That is, there is no color dimension, only the X-axis dimension is present:
- Named colors: "primary" or "secondary"
- CSS color values: "#FF5733", "rgb(255, 87, 51)", "hsl(12, 100%, 60%)"
- **Note**: If no color field object is provided, a color string MUST be included for the mentioned chart types

### 2. Special Values (Funnel Charts Only)
For funnel_chart type, use one of these special keywords:
In breakdown mode "dimension" - 
- "stage" - Colors each dimensional funnel segment with different color
- "measure" - Colors funnel segments with similar color based on value

In breakdown mode "measures" - 
- "name" - Colors each measure funnel segment with different color
- "value" - Colors measures with similar color based on value. Prefer this over "name" when possible.

### 3. Field-Based Color Object
For dynamic coloring based on data dimensions:
` + "```json" + `
{
  "field": "dimension_name|rill_measures",      // The data field to base colors on
  "type": "nominal|value", // Data type, use value only when field in "rill_measures"
  "limit": 10,                     // Limit denotes the maximum number of color categories
  "legendOrientation": "top|bottom|left|right" // Legend position (optional)
}
` + "```" + `

## Visualization Best Practices & Usage Guidelines

Choose the appropriate chart type based on your data and analysis goals:

### Time Series Analysis
- **` + "`line_chart`" + `**: Best for showing trends over time, especially with continuous data or multiple series
- **` + "`area_chart`" + `**: Ideal for cumulative trends or showing magnitude of change over time
- **Temporal axis**: Always use temporal encoding for time-based x-axis

### Categorical Comparisons
- **` + "`bar_chart`" + `**: Standard choice for comparing discrete categories or groups
- **` + "`stacked_bar`" + `**: Standard choice for comparing discrete categories or groups when split by dimension is involved
- **Nominal axis**: Use nominal encoding for categorical x-axis
- **Horizontal bars**: For bar_chart, stacked_bar or stacked_bar_normalized, put the measure on x (quantitative) and the dimension on y when category labels are long or there are many categories

### Part-to-Whole Relationships
- **` + "`donut_chart`" + `**: Shows composition of a whole
- **` + "`stacked_bar_normalized`" + `**: Compares part-to-whole across multiple groups
- **Consideration**: Avoid when precise value comparison is needed

### Multiple Dimensions
- **` + "`combo_chart`" + `**: Combines different chart types for metrics with different scales. Used when comparing 2 measures.
- **` + "`stacked_bar`" + `**: Shows cumulative values across categories (use for 2+ measures)
- **` + "`heatmap`" + `**: Reveals patterns across two categorical dimensions along with single measure
- **Color encoding**: Add a second dimension to bar, stacked bar, line and area charts through color mapping

### Specialized Use Cases
- **` + "`funnel_chart`" + `**: Visualizes conversion rates or stage-based processes
- **Distribution patterns**: Use ` + "`heatmap`" + ` for density or correlation analysis
- **Multi-measure comparison**: Prefer ` + "`stacked_bar`" + ` when comparing 3 or more related measures
- **Two-measure relationships**: Use ` + "`scatter_plot`" + ` to explore correlation, clusters, or outliers between two numeric measures across a dimension's values


## Important Chart Configuration Notes and Requirements

### Temporal Configuration
- **'time_range'** (REQUIRED for all generated specifications): Defines the temporal bounds for the chart
  - 'start': Inclusive timestamp
  - 'end': Exclusive timestamp
- **'time_grain'**: Controls temporal aggregation granularity
  - Default: "TIME_GRAIN_DAY"
  - Use to adjust time-based grouping (e.g., hour, day, week, month)
  - ALWAYS add a time_grain if a temporal dimension is mentioned in the chart spec
- **timestamp dimension field**: When referencing the time dimension, always set type to "temporal"

### Data Filtering
- **'where'**: Apply filters to chart data
  - Uses same filtering syntax as 'query_metrics_view'

### Field Configuration
- **Y-axis with multiple measures**: Use the 'fields' array in y-axis configuration
- **Field names**: Must exactly match field names in the metrics view (case-sensitive)
- **Metrics view name**: Must exactly match available view names

### Time Comparison
- **Supported chart types**: ` + "`bar_chart`" + `, ` + "`line_chart`" + `, ` + "`stacked_bar`" + `, ` + "`stacked_bar_normalized`" + ` support period-over-period comparison via the ` + "`comparison_time_range`" + ` parameter
- **How it works**: When ` + "`comparison_time_range`" + ` is provided, the chart renders the current period alongside the comparison period with visual differentiation (e.g., side-by-side bars, dimmed comparison lines)
- **Best practice**: Match the duration of the comparison period to the primary time range for meaningful comparisons

### Delta Sorting (Period-over-Period Change Sorting)
- **What it does**: Sorts dimensions by the **absolute change (delta)** between the current period and the comparison period, rather than by the current period's value alone
- **Sort values**: Use ` + "`\"y_delta\"`" + ` (ascending) or ` + "`\"-y_delta\"`" + ` (descending) in the x-axis ` + "`sort`" + ` field (` + "`\"x_delta\"`" + ` / ` + "`\"-x_delta\"`" + ` on the y field of a horizontal bar chart)
- **Requires**: ` + "`comparison_time_range`" + ` must be provided. If no comparison is active, delta sort automatically falls back to regular ` + "`\"y\"`" + ` / ` + "`\"-y\"`" + ` sorting
- **Supported chart types**: Only works on cartesian chart types with a nominal x-axis: ` + "`bar_chart`" + `, ` + "`stacked_bar`" + `, ` + "`stacked_bar_normalized`" + `, ` + "`line_chart`" + `
- **When to use**: When the user asks about "what changed the most", "biggest increase or decrease", "top movers", "which categories grew or declined", or any question about period-over-period ranking by change

### Limitations
- **Comparison measures**: Direct comparison measures like 'measure_name__delta_abs' or 'measure_name__delta_rel' are not allowed in the spec. Do not use such measures in the spec anywhere.
- **Comparison chart types**: ` + "`comparison_time_range`" + ` is only supported for cartesian chart types (bar_chart, line_chart, stacked_bar, stacked_bar_normalized). It is NOT supported for donut_chart, pie_chart, funnel_chart, heatmap, combo_chart, area_chart, or scatter_plot.`
