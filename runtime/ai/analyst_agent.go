package ai

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type AnalystAgent struct {
	Runtime *runtime.Runtime
}

var _ Tool[*AnalystAgentArgs, *AnalystAgentResult] = (*AnalystAgent)(nil)

type AnalystAgentArgs struct {
	Prompt  string `json:"prompt"`
	Explore string `json:"explore,omitempty" jsonschema:"Optional explore dashboard name. If provided, the exploration will be limited to this dashboard."`
}

type AnalystAgentResult struct {
	Response string `json:"response"`
}

func (t *AnalystAgent) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "analyst_agent",
		Title:       "Analyst Agent",
		Description: "Agent that assists with data analysis tasks.",
	}
}

func (t *AnalystAgent) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *AnalystAgent) Handler(ctx context.Context, args *AnalystAgentArgs) (*AnalystAgentResult, error) {
	session := GetSession(ctx)

	// If a specific dashboard is being explored, we pre-invoke some relevant tool calls for that dashboard.
	var metricsViewName string
	if args.Explore != "" {
		_, metricsView, err := t.getValidExploreAndMetricsView(ctx, args.Explore)
		if err != nil {
			return nil, err
		}
		metricsViewName = metricsView.Meta.Name.Name

		_, err = session.CallTool(ctx, RoleAssistant, "query_metrics_view_time_range", nil, &QueryMetricsViewTimeRangeArgs{
			MetricsView: metricsViewName,
		})
		if err != nil {
			return nil, err
		}

		_, err = session.CallTool(ctx, RoleAssistant, "get_metrics_view", nil, &GetMetricsViewArgs{
			MetricsView: metricsViewName,
		})
		if err != nil {
			return nil, err
		}
	}

	// If no specific dashboard is being explored, we pre-invoke the list_metrics_views tool.
	if args.Explore == "" {
		var listRes *ListMetricsViewsResult
		_, err := session.CallTool(ctx, RoleAssistant, "list_metrics_views", &listRes, &ListMetricsViewsArgs{})
		if err != nil {
			return nil, err
		}
	}

	// Add the analyst agent system prompt, optionally tailored for the current explore.
	systemPrompt, err := t.systemPrompt(ctx, metricsViewName, args.Explore)
	if err != nil {
		return nil, err
	}
	session.AddMessage(&AddMessageOptions{
		Role:        RoleSystem,
		Type:        MessageTypePrompt,
		ContentType: MessageContentTypeText,
		Content:     systemPrompt,
	})

	// Determine tools that can be used
	tools := []string{}
	if args.Explore == "" {
		tools = append(tools, "list_metrics_views", "get_metrics_view")
	}
	tools = append(tools, "query_metrics_view_time_range", "query_metrics_view")

	// Run an LLM tool call loop
	var response string
	err = session.Complete(ctx, "Analyst loop", &response, &CompleteOptions{
		Messages:      session.DefaultCompletionMessages(),
		Tools:         tools,
		MaxIterations: 15,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &AnalystAgentResult{Response: response}, nil
}

func (t *AnalystAgent) systemPrompt(ctx context.Context, metricsView, explore string) (string, error) {
	// Prepare template data.
	// NOTE: All the template properties are optional and may be empty.
	session := GetSession(ctx)
	data := map[string]any{
		"ai_instructions": session.ProjectInstructions(),
		"metrics_view":    metricsView,
		"explore":         explore,
	}

	// Generate the system prompt
	return executeTemplate(`<role>
You are a data analysis agent specialized in uncovering actionable business insights.
You systematically explore data using available metrics tools, then apply analytical rigor to find surprising patterns and unexpected relationships that influence decision-making.
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
The metrics view's definition and time range of available data has been provided in your tool calls. You should:
1. Carefully study the metrics view definition to understand the measures and dimensions available for analysis.
2. Remember the time range of available data and use it to inform and filter your queries.
{{ else }}
Follow these steps in order:
1. **Discover**: Use "list_metrics_views" to identify available datasets
2. **Understand**: Use "get_metrics_view" to understand measures and dimensions for the selected view  
3. **Scope**: Use "query_metrics_view_time_range" to determine the span of available data
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
</process>

<analysis_guidelines>
**Phase 1: discovery**: 
- Complete each step fully before proceeding
- Explain your approach briefly before starting
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

<thinking>
After each query in Phase 2, think through:
- What patterns or anomalies did this reveal?
- How does this connect to previous findings?
- What new questions does this raise?
- What's the most valuable next query to run?
- Are there any surprising insights worth highlighting?
</thinking>

<output_format>
Format your analysis as follows:
{{ backticks }}markdown
[Brief acknowledgment and explanation of approach]

Based on my systematic analysis, here are the key insights:

1. ## [Headline with specific impact/number]
   [Finding with business context and implications]

2. ## [Headline with specific impact/number]  
   [Finding with business context and implications]

3. ## [Headline with specific impact/number]
   [Finding with business context and implications]

[Offer specific follow-up analysis options]
{{ backticks }}
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
