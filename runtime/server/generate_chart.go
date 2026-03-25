package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GenerateRenderer(ctx context.Context, req *runtimev1.GenerateRendererRequest) (*runtimev1.GenerateRendererResponse, error) {
	rp, err := json.Marshal(req.ResolverProperties.AsMap())
	if err != nil {
		return nil, err
	}
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.resolver", req.Resolver),
		attribute.String("args.resolver_property", string(rp)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	res, _, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         req.InstanceId,
		Resolver:           req.Resolver,
		ResolverProperties: req.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             claims,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	start := time.Now()
	renderer, props, err := s.generateRendererWithAI(ctx, req.InstanceId, req.Prompt, res.Schema())

	var propsPB *structpb.Struct
	if err == nil && props != nil {
		propsPB, err = structpb.NewStruct(props)
	}

	s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_renderer",
		attribute.Int("table_column_count", len(res.Schema().Fields)),
		attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()),
		attribute.Bool("succeeded", err == nil),
	)

	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateRendererResponse{
		Renderer:           renderer,
		RendererProperties: propsPB,
	}, nil
}

// generateRendererWithAI attempts to generate a component renderer based on a user-provided prompt and a data schema.
// It currently only supports generating a Vega lite render.
func (s *Server) generateRendererWithAI(ctx context.Context, instanceID, userPrompt string, schema *runtimev1.StructType) (string, map[string]any, error) {
	// Build messages
	systemPrompt := vegaSpecSystemPrompt()
	userPrompt = vegaSpecUserPrompt(userPrompt, schema)

	msgs := []*aiv1.CompletionMessage{
		{
			Role: "system",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: systemPrompt,
					},
				},
			},
		},
		{
			Role: "user",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: userPrompt,
					},
				},
			},
		},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})
	if err != nil {
		return "", nil, err
	}

	// Extract text from content blocks
	var responseText string
	for _, block := range res.Message.Content {
		switch blockType := block.GetBlockType().(type) {
		case *aiv1.ContentBlock_Text:
			if text := blockType.Text; text != "" {
				responseText += text
			}
		default:
			// For chart generation, we only expect text responses
			return "", nil, fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
	}

	// The AI may produce Markdown output. Remove the code tags around the JSON.
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")

	return "vega_lite", map[string]any{"spec": responseText}, nil
}

// vegaSpecSystemPrompt returns the static system prompt for the Vega spec generation AI.
func vegaSpecSystemPrompt() string {
	// `{ "name": "table" }` is our format to add data in the UI. To retain the formatting of the json it is better to ask AI to keep this as the `data` config.
	return `
You are an agent whose only task is to suggest relevant chart based on a table schema.
Replace the data field in vega lite json with,
{ "name": "table" }

Your output should consist of valid JSON in the format below:

<vega lite json in the format: https://vega.github.io/schema/vega-lite/v5.json >
`
}

func vegaSpecUserPrompt(userPrompt string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
Prompt provided by the user: %s:

Based on a table with schema:
`, userPrompt)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}

// GenerateChart generates metrics_sql queries and a Vega-Lite spec for a canvas custom chart from a natural language prompt.
// It loads all available metrics views so the AI can query any of them.
func (s *Server) GenerateChart(ctx context.Context, req *runtimev1.GenerateChartRequest) (*runtimev1.GenerateChartResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Access check: require AI usage permission
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// Load all metrics views
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	resources, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}

	// Build schema descriptions for all valid metrics views
	var schemasDesc strings.Builder
	var metricsViewNames []string
	for _, r := range resources {
		mvSpec := r.GetMetricsView().State.ValidSpec
		if mvSpec == nil {
			continue
		}
		name := r.Meta.Name.Name
		metricsViewNames = append(metricsViewNames, name)
		schemasDesc.WriteString(buildMetricsViewSchemaDescription(name, mvSpec))
		schemasDesc.WriteString("\n---\n\n")
	}

	if len(metricsViewNames) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no valid metrics views found in this instance")
	}

	// Build messages
	systemPrompt := generateChartSystemPrompt(metricsViewNames, schemasDesc.String())
	userPrompt := buildGenerateChartUserPrompt(req)

	msgs := []*aiv1.CompletionMessage{
		{
			Role:    "system",
			Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: systemPrompt}}},
		},
		{
			Role:    "user",
			Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: userPrompt}}},
		},
	}

	// Connect to the AI service
	aiSvc, release, err := s.runtime.AI(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	start := time.Now()

	// Call AI service
	res, err := aiSvc.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})

	s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_custom_chart",
		attribute.Int("metrics_view_count", len(metricsViewNames)),
		attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()),
		attribute.Bool("succeeded", err == nil),
	)

	if err != nil {
		return nil, fmt.Errorf("AI completion failed: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range res.Message.Content {
		if textBlock, ok := block.GetBlockType().(*aiv1.ContentBlock_Text); ok {
			responseText += textBlock.Text
		}
	}

	// Parse the structured response
	metricsSQL, vegaSpec, err := parseGenerateChartResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &runtimev1.GenerateChartResponse{
		MetricsSql: metricsSQL,
		VegaSpec:   vegaSpec,
	}, nil
}

// buildMetricsViewSchemaDescription builds a human-readable description of a metrics view's schema for the AI.
func buildMetricsViewSchemaDescription(metricsViewName string, spec *runtimev1.MetricsViewSpec) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Metrics View: %s\n\n", metricsViewName))

	if spec.DisplayName != "" {
		b.WriteString(fmt.Sprintf("Display Name: %s\n", spec.DisplayName))
	}
	if spec.Description != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", spec.Description))
	}

	// Time dimension
	if spec.TimeDimension != "" {
		b.WriteString(fmt.Sprintf("\n### Time Dimension\n- Column: %s\n", spec.TimeDimension))
	}

	// Dimensions
	b.WriteString("\n### Dimensions\n")
	for _, dim := range spec.Dimensions {
		displayName := dim.DisplayName
		if displayName == "" {
			displayName = dim.Name
		}
		b.WriteString(fmt.Sprintf("- `%s` (display_name: %q", dim.Name, displayName))
		if dim.Description != "" {
			b.WriteString(fmt.Sprintf(", description: %q", dim.Description))
		}
		b.WriteString(")\n")
	}

	// Measures
	b.WriteString("\n### Measures\n")
	for _, m := range spec.Measures {
		displayName := m.DisplayName
		if displayName == "" {
			displayName = m.Name
		}
		b.WriteString(fmt.Sprintf("- `%s` (display_name: %q", m.Name, displayName))
		if m.Description != "" {
			b.WriteString(fmt.Sprintf(", description: %q", m.Description))
		}
		if m.FormatPreset != "" {
			b.WriteString(fmt.Sprintf(", format_preset: %q", m.FormatPreset))
		}
		if m.FormatD3 != "" {
			b.WriteString(fmt.Sprintf(", format_d3: %q", m.FormatD3))
		}
		b.WriteString(")\n")
	}

	return b.String()
}

// generateChartSystemPrompt returns the system prompt for the GenerateChart AI call.
func generateChartSystemPrompt(metricsViewNames []string, schemasDesc string) string {
	return fmt.Sprintf(`You are a data visualization expert working with Rill's metrics_sql query language. Your job is to generate SQL queries and a Vega-Lite chart spec from a natural language description.

# Available Metrics Views

%s

# metrics_sql Query Language

metrics_sql lets you write SELECT queries against metrics views as virtual tables. Each metrics view exposes its dimensions and measures as columns.

## Query rules
- Table names: use any of the available metrics views listed above: %s
- Columns: only reference dimension and measure names defined in the schema for that view
- Measures are pre-aggregated; never wrap them in SUM(), COUNT(), AVG(), or other aggregate functions
- Grouping is implicit by selected dimensions; you do not need GROUP BY unless combining with expressions like date_trunc()
- Use date_trunc('<grain>', <time_dimension>) for time bucketing (grain: minute, hour, day, week, month, quarter, year)
- Always include ORDER BY for deterministic results
- Use LIMIT to keep result sets reasonable (default to 50 for top-N queries, 500 for time series)
- Do not alias column names unless strictly required for disambiguation in multi-query specs
- Canvas-level time and dimension filters are injected automatically at runtime; do not add WHERE clauses for them
- You may write multiple queries against different metrics views; results are bound as "query1", "query2", etc.

## Example queries
Single view:
  SELECT publisher, total_bids, bid_price FROM bids_metrics ORDER BY total_bids DESC LIMIT 20
Time series:
  SELECT date_trunc('day', __time) as day, impressions, revenue FROM ad_metrics ORDER BY day
Cross-view (two queries):
  query1: SELECT campaign, spend FROM spend_metrics ORDER BY spend DESC LIMIT 10
  query2: SELECT campaign, conversions FROM conversion_metrics ORDER BY conversions DESC LIMIT 10

# Vega-Lite Specification

## Construction rules
- Generate a valid Vega-Lite v5 JSON specification
- Bind data with {"name": "query1"}, {"name": "query2"}, etc.
- Set "width": "container" and "height": "container" so the chart fills its parent
- Always include "autosize": {"type": "fit"} at the top level of the spec
- Use display_name values from the schema for axis titles, legend titles, and tooltip labels
- Apply format_d3 or format_preset from measure metadata to axis and tooltip format strings
- Pick the best mark type for the data: bar, line, area, point, rect (heatmap), arc (pie/donut), etc.
- Include tooltips with all relevant fields and human-readable formatting
- Use a clean, professional color scheme; prefer Rill's categorical palette when possible
- For temporal axes: set "type": "temporal" and choose an appropriate timeUnit
- For categorical axes: sort by the primary measure descending unless the user specifies otherwise
- For layered or multi-view charts, use the "layer" or "concat" composition operators
- Avoid unnecessary chart junk: remove gridlines on categorical axes, use concise axis labels

# Output Format

Return a single JSON object with no markdown fencing:

{"metrics_sql": ["SELECT ..."], "vega_spec": { ... }}

metrics_sql: array of SQL query strings
vega_spec: a complete Vega-Lite v5 spec object

Do NOT wrap the response in markdown code fences. Return raw JSON only.`, schemasDesc, strings.Join(metricsViewNames, ", "))
}

// buildGenerateChartUserPrompt builds the user prompt including the request and any previous context for refinement.
func buildGenerateChartUserPrompt(req *runtimev1.GenerateChartRequest) string {
	var b strings.Builder
	b.WriteString(req.Prompt)

	if len(req.PreviousSql) > 0 || req.PreviousSpec != "" {
		b.WriteString("\n\n---\nThe chart below is the current version. Apply the instruction above as a targeted edit; do not regenerate from scratch unless the change requires it.\n\n")
		if len(req.PreviousSql) > 0 {
			b.WriteString("Current metrics_sql:\n")
			for i, sql := range req.PreviousSql {
				b.WriteString(fmt.Sprintf("  query%d: %s\n", i+1, sql))
			}
		}
		if req.PreviousSpec != "" {
			b.WriteString(fmt.Sprintf("\nCurrent vega_spec:\n%s\n", req.PreviousSpec))
		}
	}

	return b.String()
}

// parseGenerateChartResponse parses the AI response to extract metrics_sql and vega_spec.
func parseGenerateChartResponse(responseText string) ([]string, string, error) {
	// Strip markdown code fences if present
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse as JSON
	var result struct {
		MetricsSQL json.RawMessage `json:"metrics_sql"`
		VegaSpec   json.RawMessage `json:"vega_spec"`
	}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, "", fmt.Errorf("invalid JSON response from AI: %w\nResponse: %s", err, responseText[:min(len(responseText), 500)])
	}

	// Parse metrics_sql: could be a single string or array of strings
	var metricsSQL []string
	if err := json.Unmarshal(result.MetricsSQL, &metricsSQL); err != nil {
		// Try single string
		var single string
		if err2 := json.Unmarshal(result.MetricsSQL, &single); err2 != nil {
			return nil, "", fmt.Errorf("metrics_sql must be a string or array of strings: %w", err)
		}
		metricsSQL = []string{single}
	}

	if len(metricsSQL) == 0 {
		return nil, "", fmt.Errorf("AI returned empty metrics_sql")
	}

	// Validate vega_spec is valid JSON
	if !json.Valid(result.VegaSpec) {
		return nil, "", fmt.Errorf("AI returned invalid vega_spec JSON")
	}

	// Pretty-print the vega spec for readability
	var prettySpec json.RawMessage
	if err := json.Unmarshal(result.VegaSpec, &prettySpec); err != nil {
		return nil, "", err
	}
	vegaSpecBytes, err := json.MarshalIndent(prettySpec, "", "  ")
	if err != nil {
		return nil, "", err
	}

	return metricsSQL, string(vegaSpecBytes), nil
}
