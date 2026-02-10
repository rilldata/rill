package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"golang.org/x/exp/slices"
)

const AnalystAgentName = "analyst_agent"

type AnalystAgent struct {
	Runtime *runtime.Runtime
}

var _ Tool[*AnalystAgentArgs, *AnalystAgentResult] = (*AnalystAgent)(nil)

type AnalystAgentArgs struct {
	Prompt     string   `json:"prompt"`
	Explore    string   `json:"explore,omitempty" yaml:"explore" jsonschema:"Optional explore dashboard name. If provided, the exploration will be limited to this dashboard."`
	Dimensions []string `json:"dimensions,omitempty" yaml:"dimensions" jsonschema:"Optional list of dimensions for queries. If provided, the queries will be limited to these dimensions."`
	Measures   []string `json:"measures,omitempty" yaml:"measures" jsonschema:"Optional list of measures for queries. If provided, the queries will be limited to these measures."`

	Canvas              string                             `json:"canvas,omitempty" yaml:"canvas" jsonschema:"Optional canvas name. If provided, the exploration will be limited to this canvas."`
	CanvasComponent     string                             `json:"canvas_component,omitempty" yaml:"canvas_component" jsonschema:"Optional canvas component name. If provided, the exploration will be limited to this canvas component."`
	WherePerMetricsView map[string]*metricsview.Expression `json:"where_per_metrics_view,omitempty" yaml:"where_per_metrics_view" jsonschema:"Optional filter for queries per metrics view. If provided, this filter will be applied to queries for each metrics view."`

	Where     *metricsview.Expression `json:"where,omitempty" yaml:"where" jsonschema:"Optional filter for queries. If provided, this filter will be applied to all queries."`
	TimeStart time.Time               `json:"time_start,omitempty" yaml:"time_start" jsonschema:"Optional start time for queries. time_end must be provided if time_start is provided."`
	TimeEnd   time.Time               `json:"time_end,omitempty" yaml:"time_end" jsonschema:"Optional end time for queries. time_start must be provided if time_end is provided."`
}

func (a *AnalystAgentArgs) ToLLM() *aiv1.ContentBlock {
	return &aiv1.ContentBlock{
		BlockType: &aiv1.ContentBlock_Text{
			Text: a.Prompt,
		},
	}
}

type AnalystAgentResult struct {
	Response string `json:"response"`
}

func (r *AnalystAgentResult) ToLLM() *aiv1.ContentBlock {
	return &aiv1.ContentBlock{
		BlockType: &aiv1.ContentBlock_Text{
			Text: r.Response,
		},
	}
}

func (t *AnalystAgent) Spec() *mcp.Tool {
	// It can't automatically infer schemas that use the metricsview.Expression type, so we manually do that here.
	inputSchema, err := jsonschema.For[*AnalystAgentArgs](&jsonschema.ForOptions{
		TypeSchemas: metricsview.TypeSchemas(),
	})
	if err != nil {
		panic(fmt.Errorf("failed to infer input schema: %w", err))
	}

	return &mcp.Tool{
		Name:        AnalystAgentName,
		Title:       "Analyst Agent",
		Description: "Agent that assists with data analysis tasks.",
		InputSchema: inputSchema,
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Analyzing...",
			"openai/toolInvocation/invoked":  "Analysis completed",
		},
	}
}

func (t *AnalystAgent) CheckAccess(ctx context.Context) (bool, error) {
	// Must be allowed to use AI and query metrics
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) || !s.Claims().Can(runtime.ReadMetrics) {
		return false, nil
	}

	// Only allow for rill user agents since it's not useful in MCP contexts.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}

	return true, nil
}

func (t *AnalystAgent) Handler(ctx context.Context, args *AnalystAgentArgs) (*AnalystAgentResult, error) {
	s := GetSession(ctx)

	// Determine if it's the first invocation of the agent in this session.
	first := len(s.Messages(FilterByType(MessageTypeCall), FilterByTool(AnalystAgentName))) == 1

	// If a specific dashboard is being explored, we pre-invoke some relevant tool calls for that dashboard.
	// TODO: This uses `first`, but that may not be safe if the user has navigated to another dashboard. We probably need some more sophisticated de-duplication here.
	var metricsViewNames []string
	if first {
		if args.Explore != "" {
			_, metricsView, err := t.getValidExploreAndMetricsView(ctx, args.Explore)
			if err != nil {
				return nil, err
			}
			metricsViewNames = append(metricsViewNames, metricsView.Meta.Name.Name)
		} else if args.Canvas != "" {
			// Pre-invoke the get_canvas tool to get the canvas definition.
			_, err := s.CallTool(ctx, RoleAssistant, GetCanvasName, nil, &GetCanvasArgs{
				Canvas: args.Canvas,
			})
			if err != nil {
				return nil, err
			}

			_, metricsViews, err := t.getValidCanvasAndMetricsViews(ctx, args.Canvas)
			if err != nil {
				return nil, err
			}

			for _, res := range metricsViews {
				metricsViewNames = append(metricsViewNames, res.Meta.Name.Name)
			}
		}

		// Pre-invoke the query_metrics_view tool for each metrics view tied to the explore/canvas.
		for _, mvName := range metricsViewNames {
			_, err := s.CallTool(ctx, RoleAssistant, QueryMetricsViewSummaryName, nil, &QueryMetricsViewSummaryArgs{
				MetricsView: mvName,
			})
			if err != nil {
				return nil, err
			}

			_, err = s.CallTool(ctx, RoleAssistant, GetMetricsViewName, nil, &GetMetricsViewArgs{
				MetricsView: mvName,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	// If no specific dashboard is being explored, we pre-invoke the list_metrics_views tool.
	if first && len(metricsViewNames) == 0 {
		_, err := s.CallTool(ctx, RoleAssistant, ListMetricsViewsName, nil, &ListMetricsViewsArgs{})
		if err != nil {
			return nil, err
		}
	}

	// Determine tools that can be used
	tools := []string{}
	if args.Explore == "" {
		tools = append(tools, ListMetricsViewsName, GetMetricsViewName, GetCanvasName)
	}
	tools = append(tools, QueryMetricsViewSummaryName, QueryMetricsViewName, CreateChartName)

	// Build completion messages
	systemPrompt, err := t.systemPrompt(ctx, metricsViewNames, args)
	if err != nil {
		return nil, err
	}
	messages := []*aiv1.CompletionMessage{NewTextCompletionMessage(RoleSystem, systemPrompt)}
	messages = append(messages, s.NewCompletionMessages(s.MessagesWithChildren(FilterByType(MessageTypeCall), FilterByTool(AnalystAgentName)))...)

	// If this is the first agent call in the session, re-organize messages to put the user prompt at the end (after the seeded tool calls).
	// NOTE: We should find a cleaner way to organize/prioritize message ordering.
	if first {
		for i, m := range messages {
			if m.Role == string(RoleUser) {
				messages = slices.Delete(messages, i, i+1)
				messages = append(messages, m)
				break
			}
		}
	}

	// Run an LLM tool call loop
	var response string
	err = s.Complete(ctx, "Analyst loop", &response, &CompleteOptions{
		Messages:      messages,
		Tools:         tools,
		MaxIterations: 20,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &AnalystAgentResult{Response: response}, nil
}

func (t *AnalystAgent) systemPrompt(ctx context.Context, metricsViewNames []string, args *AnalystAgentArgs) (string, error) {
	// Prepare template data.
	// NOTE: All the template properties are optional and may be empty.
	session := GetSession(ctx)
	ff, err := t.Runtime.FeatureFlags(ctx, session.InstanceID(), session.Claims())
	if err != nil {
		return "", fmt.Errorf("failed to get feature flags: %w", err)
	}

	metricsViewsQuoted := make([]string, len(metricsViewNames))
	for i, mv := range metricsViewNames {
		metricsViewsQuoted[i] = fmt.Sprintf("`%s`", mv)
	}

	dimensionsQuoted := make([]string, len(args.Dimensions))
	for i, dim := range args.Dimensions {
		dimensionsQuoted[i] = fmt.Sprintf("`%s`", dim)
	}

	measuresQuoted := make([]string, len(args.Measures))
	for i, measure := range args.Measures {
		measuresQuoted[i] = fmt.Sprintf("`%s`", measure)
	}

	data := map[string]any{
		"ai_instructions":  session.ProjectInstructions(),
		"metrics_views":    strings.Join(metricsViewsQuoted, ", "),
		"explore":          args.Explore,
		"canvas":           args.Canvas,
		"canvas_component": args.CanvasComponent,
		"dimensions":       strings.Join(dimensionsQuoted, ", "),
		"measures":         strings.Join(measuresQuoted, ", "),
		"feature_flags":    ff,
		"now":              time.Now(),
	}

	if !args.TimeStart.IsZero() && !args.TimeEnd.IsZero() {
		data["time_start"] = args.TimeStart.Format(time.RFC3339)
		data["time_end"] = args.TimeEnd.Format(time.RFC3339)
	}

	if args.Where != nil {
		data["where"], err = metricsview.ExpressionToSQL(args.Where)
		if err != nil {
			return "", err
		}
	}

	if args.WherePerMetricsView != nil {
		wherePerMetricsView := map[string]string{}
		for metricsViewName, whereExpr := range args.WherePerMetricsView {
			wherePerMetricsView[metricsViewName], err = metricsview.ExpressionToSQL(whereExpr)
			if err != nil {
				return "", err
			}
		}
		data["where_per_metrics_view"] = wherePerMetricsView
	}

	data["forked"] = session.Forked()

	// Generate the system prompt
	return executeTemplate(`<role>
You are a data analysis agent specialized in uncovering actionable business insights.
You systematically explore data using available metrics tools, then apply analytical rigor to find surprising patterns and unexpected relationships that influence decision-making.

Today's date is {{ .now.Format "Monday, January 2, 2006" }} ({{ .now.Format "2006-01-02" }}).
</role>

<communication_style>
- Be confident, clear, and intellectually curious
- Write conversationally using "I" and "you" - speak directly to the user
- Present insights with authority while remaining enthusiastic and collaborative
</communication_style>

<process>
**Phase 1: discovery (setup)**
{{ if .explore }}
Your goal is to analyze the contents of the dashboard "{{ .explore }}", which is powered by the metrics view(s) {{ .metrics_views }}.
The user is actively viewing this dashboard, and it's what you they refer to if they use expressions like "this dashboard", "the current view", etc.
The metrics view's definition and time range of available data has been provided in your tool calls.

Here is an overview of the settings the user has currently applied to the dashboard:
{{ if (and .time_start .time_end) }}Use time range: start={{.time_start}}, end={{.time_end}}{{ end }}
{{ if .where }}Use where filters: "{{ .where }}"{{ end }}
{{ if .measures }}Use measures: {{ .measures }}{{ end }}
{{ if .dimensions }}Use dimensions: {{ .dimensions }}{{ end }}

You should:
1. Carefully study the metrics view definition to understand the measures and dimensions available for analysis.
2. Remember the time range of available data and use it to inform and filter your queries.
{{ else if .canvas }}
Your goal is to analyze the contents of the canvas "{{ .canvas }}", which is powered by the metrics view(s) {{ .metrics_views }}.
The user is actively viewing this dashboard, and it's what you they refer to if they use expressions like "this dashboard", "the current view", etc.
The metrics views and canvas definitions have been provided in your tool calls.

Here is an overview of the settings the user has currently applied to the dashboard (Merge component's dimension_filters with "and"):
{{ if (and .time_start .time_end) }}Use time range: start={{.time_start}}, end={{.time_end}}{{ end }}
{{ if .where_per_metrics_view }}{{range $mv, $filter := .where_per_metrics_view}}Use where filters for metrics view "{{ $mv }}": "{{ $filter }}"
{{end}}{{ end }}

You should:
1. Carefully study the canvas and metrics view definition to understand the measures and dimensions available for analysis.
2. Remember the time range of available data and use it to inform and filter your queries.
{{ if .canvas_component }}
The user is looking at "{{ .canvas_component }}".
{{ end }}
{{ else }}
Follow these steps in order:
1. **Discover**: Use "list_metrics_views" to identify available datasets
2. **Understand**: Use "get_metrics_view" to understand measures and dimensions for the selected view  
3. **Scope**: Use "query_metrics_view_summary" to determine the span of available data
{{ end }}

{{ if .forked }}
Important instructions regarding access permissions:
This conversation has been transferred and the new owner may have different access permissions.
If you start seeing access errors like "action not allowed"", "resource not found" (for resources earlier available) etc., consider repeating metadata listings and lookups.
If you run into such issues, explicitly mention to the user that this may be due to conversation forking and that they may not have access to the data that the previous user had.
{{ end }}

**Phase 2: analysis (loop)**
In an iterative OODA loop, you should repeatedly use the "query_metrics_view" tool to query for insights.
Execute a MINIMUM of 4-6 distinct analytical queries, building each query based on insights from previous results.
Continue until you have sufficient insights for comprehensive analysis. Some analyses may require up to 20 queries.

In each iteration, you should:
- **Observe**: What data patterns emerge? What insights are surfacing? What gaps remain?
- **Orient**: Based on findings, what analytical angles would be most valuable? How do current insights shape next queries?
- **Decide**: Choose specific dimensions, filters, time periods, or comparisons to explore
- **Act**: Execute the query and evaluate results in <thinking> tags
{{ if .feature_flags.chat_charts }}

**Phase 3: visualization**
Create a chart: After running "query_metrics_view" create a chart using "create_chart" unless:
- The user explicitly requests a table-only response
- The query returns only a single scalar value

Choose the appropriate chart type based on your data:
- Time series data: line_chart or area_chart (better for cummalative trends)
- Category comparisons: bar_chart or stacked_bar
- Part-to-whole relationships: donut_chart
- Multiple dimensions: Use color encoding with bar_chart, stacked_bar or line_chart
- Two measures from the same metrics view: Use combo_chart
- Multiple measures from the same metrics view (more that 2): Use stacked bar chart with multiple measure fields
- Distribution across two dimensions: heatmap
{{ end }}
</process>

<analysis_guidelines>
**Phase 1: discovery**: 
- Briefly explain your approach before starting
- Complete each step fully before proceeding
- If any step fails, investigate and adapt

**Phase 2: analysis**:
- Start broad (overall patterns), then drill into specific segments
- Always include time-based analysis using comparison features (delta_abs, delta_rel)
- Focus on insights that are surprising, actionable, and quantified
- Never repeat identical queries - each should explore new analytical angles
- Use <thinking> tags between queries to evaluate results and plan next steps

**Quality Standards**:
- Prioritize findings that contradict expectations or reveal hidden patterns
- Quantify changes and impacts with specific numbers
- Link insights to business implications and decisions

**Data Accuracy Requirements**:
- ALL numbers and calculations must come from "query_metrics_view" tool results
- NEVER perform manual calculations or mathematical operations
- If a desired calculation cannot be achieved through the metrics tools, explicitly state this limitation
- Use only the exact numbers returned by the tools in your analysis
</analysis_guidelines>

<guardrails>
You only engage in conversation that relates to the project's data.
If a question seems unrelated, first inspect the available metrics views to see if it fits the dataset's domain.
Decline to engage if the topic is clearly outside the scope of the data (e.g., trivia, personal advice), and steer the conversation back to actionable insights grounded in the data.
</guardrails>

<thinking>
After each query in Phase 2, think through:
- What patterns or anomalies did this reveal?
- How does this connect to previous findings?
- What new questions does this raise?
- What's the most valuable next query to run?
- Are there any surprising insights worth highlighting?
</thinking>

<output_format>
**Format your analysis as follows**:
{{ backticks }}markdown
Based on the data analysis, here are the key insights:

1. ## [Headline with specific impact/number]
   [Finding with business context and implications]

2. ## [Headline with specific impact/number]  
   [Finding with business context and implications]

3. ## [Headline with specific impact/number]
   [Finding with business context and implications]

[Optional: Offer specific follow-up analysis options]
{{ backticks }}

**Citation Requirements**:
- Every 'query_metrics_view' result includes an 'open_url' field - use this as a markdown link to cite EVERY quantitative claim made to the user
- Citations must be inline at the end of a sentence or paragraph, not on a separate line
- Use descriptive text in sentence case (e.g. "This suggests Android is valuable ([Device breakdown](url))." or "Revenue increased 25%% ([Revenue by country](url)).")
- When one paragraph contains multiple insights from the same query, cite once at the end of the paragraph
</output_format>

{{ if .ai_instructions }}
<additional_user_provided_instructions>
{{ .ai_instructions }}
</additional_user_provided_instructions>
{{ end }}
`, data)
}

func (t *AnalystAgent) getValidExploreAndMetricsView(ctx context.Context, exploreName string) (*runtimev1.Resource, *runtimev1.Resource, error) {
	session := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, session.InstanceID())
	if err != nil {
		return nil, nil, err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: exploreName}, false)
	if err != nil {
		return nil, nil, err
	}

	explore, access, err := t.Runtime.ApplySecurityPolicy(ctx, session.InstanceID(), session.Claims(), r)
	if err != nil {
		return nil, nil, err
	}
	if !access {
		return nil, nil, fmt.Errorf("explore %q not found", exploreName)
	}

	exploreSpec := explore.GetExplore().State.ValidSpec
	if exploreSpec == nil {
		return nil, nil, fmt.Errorf("explore %q is not valid", exploreName)
	}

	metricsView, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: exploreSpec.MetricsView}, false)
	if err != nil {
		return nil, nil, err
	}

	metricsView, access, err = t.Runtime.ApplySecurityPolicy(ctx, session.InstanceID(), session.Claims(), metricsView)
	if err != nil {
		return nil, nil, err
	}
	if !access {
		return nil, nil, fmt.Errorf("metrics view %q not found", exploreSpec.MetricsView)
	}

	metricsViewSpec := metricsView.GetMetricsView().State.ValidSpec
	if metricsViewSpec == nil {
		return nil, nil, fmt.Errorf("metrics view %q is not valid", exploreSpec.MetricsView)
	}

	return explore, metricsView, nil
}

func (t *AnalystAgent) getValidCanvasAndMetricsViews(ctx context.Context, canvasName string) (*runtimev1.Resource, map[string]*runtimev1.Resource, error) {
	session := GetSession(ctx)

	resolvedCanvas, err := t.Runtime.ResolveCanvas(ctx, session.InstanceID(), canvasName, session.Claims())
	if err != nil {
		return nil, nil, err
	}

	if resolvedCanvas == nil || resolvedCanvas.Canvas == nil {
		return nil, nil, fmt.Errorf("canvas %q not found", canvasName)
	}

	metricsViews := map[string]*runtimev1.Resource{}
	for mv, res := range resolvedCanvas.ReferencedMetricsViews {
		metricsView := res.GetMetricsView()
		if metricsView == nil || metricsView.State.ValidSpec == nil {
			continue
		}
		metricsViews[mv] = res
	}

	return resolvedCanvas.Canvas, metricsViews, nil
}
