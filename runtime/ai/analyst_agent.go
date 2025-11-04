package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Prompt     string                  `json:"prompt"`
	Explore    string                  `json:"explore" yaml:"explore" jsonschema:"Optional explore dashboard name. If provided, the exploration will be limited to this dashboard."`
	Dimensions []string                `json:"dimensions" yaml:"dimensions" jsonschema:"Optional list of dimensions for queries. If provided, the queries will be limited to these dimensions."`
	Measures   []string                `json:"measures" yaml:"measures" jsonschema:"Optional list of measures for queries. If provided, the queries will be limited to these measures."`
	Where      *metricsview.Expression `json:"filters" yaml:"filters" jsonschema:"Optional filter for queries. If provided, this filter will be applied to all queries."`
	TimeStart  time.Time               `json:"time_start" yaml:"time_start" jsonschema:"Optional start time for queries. time_end must be provided if time_start is provided."`
	TimeEnd    time.Time               `json:"time_end" yaml:"time_end" jsonschema:"Optional end time for queries. time_start must be provided if time_end is provided."`
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
	return &mcp.Tool{
		Name:        AnalystAgentName,
		Title:       "Analyst Agent",
		Description: "Agent that assists with data analysis tasks.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Analyzing…",
			"openai/toolInvocation/invoked":  "Completed analysis",
		},
	}
}

func (t *AnalystAgent) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)

	// Must be allowed to use AI features
	if !s.Claims().Can(runtime.UseAI) {
		return false
	}

	// Only allow for rill user agents since it's not useful in MCP contexts.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false
	}

	return true
}

func (t *AnalystAgent) Handler(ctx context.Context, args *AnalystAgentArgs) (*AnalystAgentResult, error) {
	s := GetSession(ctx)

	// Determine if it's the first invocation of the agent in this session.
	first := len(s.Messages(FilterByType(MessageTypeCall), FilterByTool(AnalystAgentName))) == 1

	// If a specific dashboard is being explored, we pre-invoke some relevant tool calls for that dashboard.
	// TODO: This uses `first`, but that may not be safe if the user has navigated to another dashboard. We probably need some more sophisticated de-duplication here.
	var metricsViewName string
	if first && args.Explore != "" {
		_, metricsView, err := t.getValidExploreAndMetricsView(ctx, args.Explore)
		if err != nil {
			return nil, err
		}
		metricsViewName = metricsView.Meta.Name.Name

		_, err = s.CallTool(ctx, RoleAssistant, "query_metrics_view_summary", nil, &QueryMetricsViewSummaryArgs{
			MetricsView: metricsViewName,
		})
		if err != nil {
			return nil, err
		}

		_, err = s.CallTool(ctx, RoleAssistant, "get_metrics_view", nil, &GetMetricsViewArgs{
			MetricsView: metricsViewName,
		})
		if err != nil {
			return nil, err
		}
	}

	// If no specific dashboard is being explored, we pre-invoke the list_metrics_views tool.
	if first && args.Explore == "" {
		_, err := s.CallTool(ctx, RoleAssistant, "list_metrics_views", nil, &ListMetricsViewsArgs{})
		if err != nil {
			return nil, err
		}
	}

	// Determine tools that can be used
	tools := []string{}
	if args.Explore == "" {
		tools = append(tools, "list_metrics_views", "get_metrics_view")
	}
	tools = append(tools, "query_metrics_view_summary", "query_metrics_view", "create_chart")

	// Build completion messages
	systemPrompt, err := t.systemPrompt(ctx, metricsViewName, args)
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

func (t *AnalystAgent) systemPrompt(ctx context.Context, metricsViewName string, args *AnalystAgentArgs) (string, error) {
	// Prepare template data.
	// NOTE: All the template properties are optional and may be empty.
	session := GetSession(ctx)
	ff, err := t.Runtime.FeatureFlags(ctx, session.InstanceID(), session.Claims())
	if err != nil {
		return "", fmt.Errorf("failed to get feature flags: %w", err)
	}
	data := map[string]any{
		"ai_instructions": session.ProjectInstructions(),
		"metrics_view":    metricsViewName,
		"explore":         args.Explore,
		"dimensions":      strings.Join(args.Dimensions, ", "),
		"measures":        strings.Join(args.Measures, ", "),
		"feature_flags":   ff,
		"now":             time.Now(),
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
Your goal is to analyze the contents of the dashboard "{{ .explore }}", which is powered by the metrics view "{{ .metrics_view }}".
The user is actively viewing this dashboard, and it's what you they refer to if they use expressions like "this dashboard", "the current view", etc.
The metrics view's definition and time range of available data has been provided in your tool calls.

Here is an overview of the settings the user has currently applied to the dashboard:
{{ if (and .time_start .time_end) }}Use time range: start={{.time_start}}, end={{.time_end}}{{ end }}
{{ if .where }}Use where filters: "{{ .where }}"{{ end }}
{{ if .measures }}Use measures: "{{ .measures }}"{{ end }}
{{ if .dimensions }}Use dimensions: "{{ .dimensions }}"{{ end }}

You should:
1. Carefully study the metrics view definition to understand the measures and dimensions available for analysis.
2. Remember the time range of available data and use it to inform and filter your queries.
{{ else }}
Follow these steps in order:
1. **Discover**: Use "list_metrics_views" to identify available datasets
2. **Understand**: Use "get_metrics_view" to understand measures and dimensions for the selected view  
3. **Scope**: Use "query_metrics_view_summary" to determine the span of available data
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
