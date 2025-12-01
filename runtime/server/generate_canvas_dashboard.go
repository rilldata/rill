package server

import (
	"bytes"
	"context"
	"errors"
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
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// GenerateCanvasFile generates a canvas YAML file from a metrics view
func (s *Server) GenerateCanvasFile(ctx context.Context, req *runtimev1.GenerateCanvasFileRequest) (*runtimev1.GenerateCanvasFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metrics_view_name", req.MetricsViewName),
		attribute.String("args.path", req.Path),
		attribute.Bool("args.use_ai", req.UseAi),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	// Get the metrics view resource
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	metricsView, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: req.MetricsViewName}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Errorf(codes.NotFound, "metrics view %q not found", req.MetricsViewName)
		}
		return nil, err
	}

	// Get the metrics view spec
	spec := metricsView.GetMetricsView().State.ValidSpec
	if spec == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "metrics view %q is not valid", req.MetricsViewName)
	}

	// Extract measures, dimensions, and time dimension
	var measures []string
	for _, m := range spec.Measures {
		measures = append(measures, m.Name)
	}

	var dimensions []string
	for _, d := range spec.Dimensions {
		dimensions = append(dimensions, d.Name)
	}

	timeDimension := spec.TimeDimension

	// Try to generate the YAML with AI
	var data string
	var aiSucceeded bool
	if req.UseAi {
		// Generate
		start := time.Now()
		res, err := s.generateCanvasDashboardYAMLWithAI(ctx, req.InstanceId, req.MetricsViewName, measures, dimensions, timeDimension)
		if err != nil {
			s.logger.Warn("failed to generate canvas dashboard YAML using AI", zap.Error(err), observability.ZapCtx(ctx))
		} else {
			data = res
			aiSucceeded = true
		}

		// Emit event
		attrs := []attribute.KeyValue{
			attribute.Int("measures_count", len(measures)),
			attribute.Int("dimensions_count", len(dimensions)),
			attribute.Bool("has_time_dimension", timeDimension != ""),
		}
		attrs = append(attrs,
			attribute.Bool("succeeded", aiSucceeded),
			attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()),
		)
		if err != nil {
			attrs = append(attrs, attribute.String("error", err.Error()))
		}
		s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_canvas_dashboard_yaml", attrs...)
	}

	// If we didn't manage to generate the YAML using AI, we fall back to the simple generator
	if data == "" {
		data, err = generateCanvasDashboardYAMLSimple(req.MetricsViewName, measures, dimensions, timeDimension)
		if err != nil {
			return nil, err
		}
	}

	s.logger.Info("Generated canvas dashboard YAML",
		zap.String("metrics_view", req.MetricsViewName),
		zap.String("path", req.Path),
		zap.Bool("ai_succeeded", aiSucceeded),
		zap.Int("yaml_length", len(data)),
		observability.ZapCtx(ctx),
	)

	// Write the file to the repo
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()
	err = repo.Put(ctx, req.Path, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateCanvasFileResponse{AiSucceeded: aiSucceeded}, nil
}

// generateCanvasDashboardYAMLWithAI attempts to generate a canvas dashboard YAML definition using AI
func (s *Server) generateCanvasDashboardYAMLWithAI(ctx context.Context, instanceID, metricsViewName string, measures, dimensions []string, timeDimension string) (string, error) {
	// Build messages
	systemPrompt := canvasDashboardYAMLSystemPrompt()
	userPrompt := canvasDashboardYAMLUserPrompt(metricsViewName, measures, dimensions, timeDimension)

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
		return "", err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to generate canvas dashboard YAML
	res, err := ai.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})
	if err != nil {
		return "", err
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
			return "", fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
	}

	// The AI may produce Markdown output. Remove the code tags around the YAML.
	responseText = strings.TrimPrefix(responseText, "```yaml")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse the YAML to validate it's well-formed
	var doc canvasDashboardYAML
	if err := yaml.Unmarshal([]byte(responseText), &doc); err != nil {
		return "", fmt.Errorf("invalid canvas dashboard YAML: %w", err)
	}

	// Ensure required fields are set
	doc.Type = "canvas"
	if doc.DisplayName == "" {
		doc.DisplayName = identifierToDisplayName(metricsViewName) + " Dashboard"
	}
	if doc.Defaults == nil {
		doc.Defaults = &canvasDefaults{
			TimeRange:      "P7D",
			ComparisonMode: "time",
		}
	}

	// Render the updated YAML
	return marshalCanvasDashboardYAML(&doc, true)
}

// Chart type examples in YAML format for canvas dashboards

const barChartExample = `bar_chart:
  metrics_view: "<metrics_view_name>"
  title: "Top Advertisers by Total Bids"
  color: "primary"
  x:
    field: "<dimension1>"
    limit: 20
    showNull: true
    type: "nominal"
    sort: "-y"
  y:
    field: "<measure1>"
    type: "quantitative"
    zeroBasedOrigin: true`

const lineChartExample = `line_chart:
  metrics_view: "<metrics_view_name>"
  title: "Trends Over Time"
  color:
    field: "<dimension1>"
    limit: 3
    type: "nominal"
  x:
    field: "<time_dimension>"
    type: "temporal"
  y:
    field: "<measure1>"
    type: "quantitative"
    zeroBasedOrigin: true`

const areaChartExample = `area_chart:
  metrics_view: "<metrics_view_name>"
  title: "Magnitude Over Time"
  color:
    field: "<dimension1>"
    type: "nominal"
  x:
    field: "<time_dimension>"
    limit: 20
    showNull: true
    type: "temporal"
  y:
    field: "<measure1>"
    type: "quantitative"
    zeroBasedOrigin: true`

const stackedBarExample = `stacked_bar:
  metrics_view: "<metrics_view_name>"
  title: "Stacked Metrics Over Time"
  color:
    field: "rill_measures"
    legendOrientation: "top"
    type: "value"
  x:
    field: "<time_dimension>"
    limit: 20
    type: "temporal"
  y:
    field: "<measure1>"
    fields:
      - "<measure1>"
      - "<measure2>"
      - "<measure3>"
    type: "quantitative"
    zeroBasedOrigin: true`

const stackedBarNormalizedExample = `stacked_bar_normalized:
  metrics_view: "<metrics_view_name>"
  title: "Proportional Distribution"
  color:
    field: "<dimension1>"
    limit: 3
    type: "nominal"
  x:
    field: "<time_dimension>"
    limit: 20
    type: "temporal"
  y:
    field: "<measure1>"
    type: "quantitative"
    zeroBasedOrigin: true`

const donutChartExample = `donut_chart:
  metrics_view: "<metrics_view_name>"
  title: "Distribution by Category"
  color:
    field: "<dimension1>"
    limit: 20
    type: "nominal"
  innerRadius: 50
  measure:
    field: "<measure1>"
    type: "quantitative"
    showTotal: true`

const heatmapExample = `heatmap:
  metrics_view: "<metrics_view_name>"
  title: "Density Across Two Dimensions"
  color:
    field: "<measure1>"
    type: "quantitative"
  x:
    field: "<dimension1>"
    limit: 10
    type: "nominal"
  y:
    field: "<dimension2>"
    limit: 24
    type: "nominal"
    sort: "-color"`

const comboChartExample = `combo_chart:
  metrics_view: "<metrics_view_name>"
  title: "Combined Metrics Analysis"
  color:
    field: "measures"
    legendOrientation: "top"
    type: "value"
  x:
    field: "<time_dimension>"
    limit: 20
    type: "temporal"
  y1:
    field: "<measure1>"
    mark: "bar"
    type: "quantitative"
    zeroBasedOrigin: true
  y2:
    field: "<measure2>"
    mark: "line"
    type: "quantitative"
    zeroBasedOrigin: true`

const leaderboardDetailsPrompt = `
## Leaderboard Component

The leaderboard component displays a ranked table of dimension values based on one or more measures. Leaderboards show ranked data with the top performers highlighted.

**When to use:**
- Showing top N entities by a specific measure (e.g., top customers by revenue, top products by sales)
- Ranking categories or groups
- Displaying tabular data with sorting capabilities

**Key parameters:**
- ` + "`metrics_view`" + `: The metrics view name to query data from (required)
- ` + "`dimensions`" + `: Array of dimension fields to display as columns (required, typically 1-2 dimensions)
- ` + "`measures`" + `: Array of measure fields to display as columns (required, typically 1-3 measures)
- ` + "`num_rows`" + `: Number of rows to display (default: 7, typically 5-15)

**Important notes:**
- NEVER use time dimensions in the leaderboard dimensions array - time dimensions are only for temporal charts
- Use non-time dimensions only (e.g., customer_id, product_name, category, region)
- The leaderboard automatically sorts by the first measure in descending order
- Best suited for categorical data analysis, not time-series data

**Example:**
` + "```yaml" + `
leaderboard:
  metrics_view: "<metrics_view_name>"
  dimensions:
    - "<dimension1>"
  measures:
    - "<measure1>"
  num_rows: 10
` + "```" + `
`

const markdownDetailsPrompt = `
## Markdown Component

The markdown component allows you to add rich text content, documentation, and context to your canvas dashboards. Use it to provide descriptions, insights, and guidance for dashboard users.

**When to use:**
- Adding overview and context about the dashboard's purpose
- Documenting key questions the dashboard answers
- Providing insights or analysis notes
- Adding headers, sections, or explanatory text

**Key parameters:**
- ` + "`content`" + `: The markdown text content (required, supports full markdown syntax)
- ` + "`alignment`" + `: Optional alignment settings
  - ` + "`horizontal`" + `: left, center, or right (default: left)
  - ` + "`vertical`" + `: top, middle, or bottom (default: top)

**Supported markdown features:**
- Headers (# H1, ## H2, ### H3, etc.)
- Bold (**text**) and italic (*text*)
- Lists (bulleted and numbered)
- Links [text](url)
- Horizontal rules (---)

**Best practices:**
- Place markdown components at the top of the dashboard to provide context
- Add empty new line between each markdown feature to render properly
- Use headers to organize content and create visual hierarchy
- Keep content concise and focused on key insights
- Use bullet points for easy scanning

**Example:**

Notice how empty new lines have been added in the content to render properly.
` + "```yaml" + `
markdown:
  content: |
    ## Dashboard Overview

    This dashboard provides a comprehensive overview of bidding activity, spend, win rates across your advertising inventory.

    ---
  alignment:
    horizontal: left
    vertical: top
` + "```" + `
`

// chartGuidelinesPrompt contains the visualization guidelines for canvas dashboard generation
const chartGuidelinesPrompt = `

# Chart configuration guidelines

### Data Types
- **nominal**: Categorical data (e.g., categories, names, labels), use for dimensions
- **temporal**: Time-based data (dates, timestamps), use for time dimensions and timestamps
- **quantitative**: Numerical data (counts, amounts, measurements), use for measures
- **value**: Special type for multiple measures (used in color field)

### Special Fields
- **rill_measures**: Special field for multiple measures in stacked charts and area charts. The field name is only used in color field object. DO NOT USE it for other keys except for "color" key in the field object.

### Common Field Properties
- **field**: The field name from the metrics view
- **type**: Data type (nominal, temporal, quantitative, value)
- **limit**: Maximum number of values to display for selected sort mode
- **showNull**: Include null values in the visualization (true/false)
- **sort**: Sorting order
  - ` + "`\"x\"`" + ` or ` + "`\"-x\"`" + `: Sort by x-axis values (ascending/descending)
  - ` + "`\"y\"`" + ` or ` + "`\"-y\"`" + `: Sort by y-axis values (ascending/descending)
	- ` + "`\"color\"`" + ` or ` + "`\"-color\"`" + `: Sort by color field values (ascending/descending) Only used for heatmap charts
	- ` + "`\"measure\"`" + ` or ` + "`\"-measure\"`" + `: Sort by measure field values (ascending/descending) Only used for donut charts
  - Array of values for custom sort order (e.g., weekday names)
- **zeroBasedOrigin**: Start y-axis from zero (true/false)
- **showTotal**: Displays the measure total without any breakdown. Only used for donut chart to display totals in center

## Color Configuration for Charts

Colors can be specified in two ways depending on the chart type and requirements:

### 1. Single Color String
For bar_chart, stacked_bar, line_chart, and area_chart types in single measure mode and only 1 dimensions is involved. That is, there is no color dimension, only the X-axis dimension is present:
- Named colors: "primary" or "secondary"
- CSS color values: "#FF5733", "rgb(255, 87, 51)", "hsl(12, 100%%, 60%%)"
- **Note**: If no color field object is provided, a color string MUST be included for the mentioned chart types

### 2. Field-Based Color Object
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

### Part-to-Whole Relationships
- **` + "`donut_chart`" + `**: Shows composition of a whole
- **` + "`stacked_bar_normalized`" + `**: Compares part-to-whole across multiple groups
- **Consideration**: Avoid when precise value comparison is needed

### Multiple Dimensions
- **` + "`combo_chart`" + `**: Combines different chart types for metrics with different scales. Used when comparing 2 measures.
- **` + "`stacked_bar`" + `**: Shows cumulative values across categories (use for 2+ measures)
- **` + "`heatmap`" + `**: Reveals patterns across two categorical dimensions along with single measure
- **Color encoding**: Add a second dimension to bar, stacked bar, line and area charts through color mapping

### Field Configuration
- **Y-axis with multiple measures**: Use the 'fields' array in y-axis configuration
- **Field names**: Must exactly match field names in the metrics view (case-sensitive)
- **Metrics view name**: Must exactly match available view names

## Chart Type Examples

### Bar Chart
` + "```yaml" + `
%s
` + "```" + `

### Line Chart
` + "```yaml" + `
%s
` + "```" + `

### Area Chart
` + "```yaml" + `
%s
` + "```" + `

### Stacked Bar Chart
` + "```yaml" + `
%s
` + "```" + `

### Normalized Stacked Bar Chart
` + "```yaml" + `
%s
` + "```" + `

### Donut Chart
` + "```yaml" + `
%s
` + "```" + `

### Heatmap
` + "```yaml" + `
%s
` + "```" + `

### Combo Chart
` + "```yaml" + `
%s
` + "```" + `
`

// canvasDashboardYAMLSystemPrompt returns the static system prompt for the canvas dashboard generation AI
func canvasDashboardYAMLSystemPrompt() string {
	template := canvasDashboardYAML{
		DisplayName: "<human-friendly display name for the dashboard>",
		Defaults: &canvasDefaults{
			TimeRange:      "P7D",
			ComparisonMode: "time",
		},
		Rows: []*canvasRow{
			{
				Items: []*canvasItem{
					{
						Width: 12,
						Component: map[string]interface{}{
							"markdown": map[string]interface{}{
								"content": "<markdown content>",
								"alignment": map[string]interface{}{
									"horizontal": "left",
									"vertical":   "top",
								},
							},
						},
					},
				},
				Height: "180px",
			},
			{
				Items: []*canvasItem{
					{
						Width: 12,
						Component: map[string]interface{}{
							"kpi_grid": map[string]interface{}{
								"measures":     []string{"<measure1>", "<measure2>", "<measure3>", "<measure4>"},
								"metrics_view": "<metrics_view_name>",
								"comparison":   []string{"percent_change", "delta", "previous"},
							},
						},
					},
				},
				Height: "240px",
			},
			{
				Items: []*canvasItem{
					{
						Width: 6,
						Component: map[string]interface{}{
							"leaderboard": map[string]interface{}{
								"measures":     []string{"<measure1>"},
								"metrics_view": "<metrics_view_name>",
								"num_rows":     10,
								"dimensions":   []string{"<dimension1>"},
							},
						},
					},
					{
						Width: 6,
						Component: map[string]interface{}{
							"line_chart": map[string]interface{}{
								"metrics_view": "<metrics_view_name>",
								"title":        "<descriptive chart title>",
								"x": map[string]interface{}{
									"field": "<time_dimension>",
									"type":  "temporal",
								},
								"y": map[string]interface{}{
									"field":           "<measure1>",
									"type":            "quantitative",
									"zeroBasedOrigin": true,
								},
								"color": "primary",
							},
						},
					},
				},
				Height: "400px",
			},
			{
				Items: []*canvasItem{
					{
						Width: 4,
						Component: map[string]interface{}{
							"bar_chart": map[string]interface{}{
								"metrics_view": "<metrics_view_name>",
								"title":        "<chart title>",
								"color":        "secondary",
								"x": map[string]interface{}{
									"field":    "<dimension2>",
									"limit":    10,
									"showNull": true,
									"type":     "nominal",
									"sort":     "-y",
								},
								"y": map[string]interface{}{
									"field":           "<measure2>",
									"type":            "quantitative",
									"zeroBasedOrigin": true,
								},
							},
						},
					},
					{
						Width: 4,
						Component: map[string]interface{}{
							"donut_chart": map[string]interface{}{
								"metrics_view": "<metrics_view_name>",
								"title":        "<chart title>",
								"color": map[string]interface{}{
									"field": "<dimension1>",
									"limit": 8,
									"type":  "nominal",
								},
								"innerRadius": 50,
								"measure": map[string]interface{}{
									"field":     "<measure1>",
									"type":      "quantitative",
									"showTotal": true,
								},
							},
						},
					},
					{
						Width: 4,
						Component: map[string]interface{}{
							"area_chart": map[string]interface{}{
								"metrics_view": "<metrics_view_name>",
								"title":        "<chart title>",
								"color": map[string]interface{}{
									"field": "<dimension1>",
									"type":  "nominal",
									"limit": 3,
								},
								"x": map[string]interface{}{
									"field": "<time_dimension>",
									"type":  "temporal",
									"limit": 20,
								},
								"y": map[string]interface{}{
									"field":           "<measure1>",
									"type":            "quantitative",
									"zeroBasedOrigin": true,
								},
							},
						},
					},
				},
				Height: "400px",
			},
		},
	}
	out, err := yaml.Marshal(template)
	if err != nil {
		panic(err)
	}

	// Format the chart guidelines with all the examples
	formattedChartGuidelines := fmt.Sprintf(chartGuidelinesPrompt,
		barChartExample,
		lineChartExample,
		areaChartExample,
		stackedBarExample,
		stackedBarNormalizedExample,
		donutChartExample,
		heatmapExample,
		comboChartExample,
	)

	prompt := fmt.Sprintf(`
You are an agent whose only task is to create a Canvas dashboard YAML configuration based on a metrics view.
The canvas should include business-relevant components and visualizations that help users understand their key metrics.

Your output should only consist of valid YAML in the format below:

%s

# Layout Guidelines:

## Row and Item Structure
- Canvas dashboards contain multiple rows, each with an 'items' array containing widgets/components
- Each row has a total span of **12 units**
- Components can have widths from **3 (minimum)** to **12 (maximum)** units
- Maximum of **4 components** can fit in a single row (4 × 3 = 12)
- You can add multiple rows, but keep the dashboard balanced and not overwhelming

## Width Best Practices
- Full width (12): Use for KPI grids, markdown overviews, or primary visualizations
- Half width (6): Use for side-by-side comparisons (leaderboard + chart, chart + chart)
- Third width (4): Use for three equal components in a row
- Quarter width (3): Use for four equal components in a row (e.g., small KPI cards, small charts)

## Row Height Recommendations
- Markdown/Text: 120px-200px (depending on content)
- KPI Grid: 250px-240px
- Charts: 400px-500px (standard visualization height)
- Leaderboards: 300px-450px (depending on num_rows)

# Content Guidelines:
1. Row 1 should have a markdown component at the top to provide dashboard context and overview
2. Row 2 should contain a KPI grid with 2-4 of the most business-relevant measures
3. Row 3 should contain:
   - Left (width 6): A leaderboard with the most important NON-TIME dimension and a relevant measures
   - Right (width 6): A stacked_bar or line_chart showing trends over time (if time dimension exists)
4. You may add Row 4 or Row 5 with additional relevant charts. Pick the right number of component per row as needed.
5. All components must reference the provided metrics_view name
6. Choose dimensions and measures that would provide the most business value
7. Only use valid and available measures and dimensions names provided by the user
8. For charts with time dimension, use the timestamp from the metrics view as the field name
9. Use descriptive titles for charts
10. IMPORTANT: The time dimension is SPECIAL - it can ONLY be used in the x-axis field for temporal charts. NEVER use the time dimension in:
   - Leaderboard dimensions
   - Color fields
   - Any other dimension fields

# Component types available:
- markdown: For adding text, context, and documentation
- kpi_grid: For key metrics summary
- leaderboard: For top rankings
- stacked_bar: For temporal or categorical comparisons
- line_chart: For time series trends
- bar_chart: For categorical comparisons
- donut_chart: For proportional breakdowns
- heatmap: For two-dimensional distribution

%s

%s

%s
`, string(out), markdownDetailsPrompt, leaderboardDetailsPrompt, formattedChartGuidelines)

	return prompt
}

// canvasDashboardYAMLUserPrompt returns the dynamic user prompt for the canvas dashboard generation AI
func canvasDashboardYAMLUserPrompt(metricsViewName string, measures, dimensions []string, timeDimension string) string {
	prompt := fmt.Sprintf(`
Create a Canvas dashboard based on the metrics view named %q.

Available measures:
`, metricsViewName)
	for _, m := range measures {
		prompt += fmt.Sprintf("- %s\n", m)
	}

	prompt += "\nAvailable dimensions:\n"
	for _, d := range dimensions {
		prompt += fmt.Sprintf("- %s\n", d)
	}

	if timeDimension != "" {
		prompt += fmt.Sprintf("\nTime dimension: %s\n", timeDimension)
	} else {
		prompt += "\nNo time dimension available.\n"
	}

	return prompt
}

// generateCanvasDashboardYAMLSimple generates a simple canvas dashboard YAML definition
func generateCanvasDashboardYAMLSimple(metricsViewName string, measures, dimensions []string, timeDimension string) (string, error) {
	doc := &canvasDashboardYAML{
		Type:        "canvas",
		DisplayName: identifierToDisplayName(metricsViewName) + " Dashboard",
		Defaults: &canvasDefaults{
			TimeRange:      "P7D",
			ComparisonMode: "time",
		},
	}

	// Filter out time dimension from regular dimensions (time dimension is special and can't be used in leaderboards or color fields)
	var nonTimeDimensions []string
	for _, d := range dimensions {
		if timeDimension == "" || d != timeDimension {
			nonTimeDimensions = append(nonTimeDimensions, d)
		}
	}

	// Row 1: KPI Grid with up to 4 measures
	kpiMeasures := measures
	if len(kpiMeasures) > 4 {
		kpiMeasures = measures[:4]
	}

	if len(kpiMeasures) > 0 {
		row1 := &canvasRow{
			Height: "240px",
			Items: []*canvasItem{
				{
					Width: 12,
					Component: map[string]interface{}{
						"kpi_grid": map[string]interface{}{
							"measures":     kpiMeasures,
							"metrics_view": metricsViewName,
							"comparison":   []string{"percent_change", "delta", "previous"},
						},
					},
				},
			},
		}
		doc.Rows = append(doc.Rows, row1)
	}

	// Row 2: Leaderboard (left) and Chart (right)
	if len(measures) > 0 {
		row2Items := []*canvasItem{}

		// Left: Leaderboard with first non-time dimension and first measure
		if len(nonTimeDimensions) > 0 {
			row2Items = append(row2Items, &canvasItem{
				Width: 6,
				Component: map[string]interface{}{
					"leaderboard": map[string]interface{}{
						"measures":     []string{measures[0]},
						"metrics_view": metricsViewName,
						"num_rows":     7,
						"dimensions":   []string{nonTimeDimensions[0]},
					},
				},
			})
		}

		// Right: Chart with time dimension if available
		if timeDimension != "" {
			chartType := "stacked_bar"
			chartComponent := map[string]interface{}{
				"metrics_view": metricsViewName,
				"title":        fmt.Sprintf("%s over time", identifierToDisplayName(measures[0])),
				"x": map[string]interface{}{
					"field":    timeDimension,
					"type":     "temporal",
					"limit":    10,
					"showNull": true,
				},
				"y": map[string]interface{}{
					"field":           measures[0],
					"type":            "quantitative",
					"zeroBasedOrigin": true,
				},
			}

			// Add color dimension if we have non-time dimensions (can't use time dimension in color field)
			if len(nonTimeDimensions) > 0 {
				chartComponent["color"] = map[string]interface{}{
					"field": nonTimeDimensions[0],
					"type":  "nominal",
					"limit": 10,
				}
			}

			row2Items = append(row2Items, &canvasItem{
				Width: 6,
				Component: map[string]interface{}{
					chartType: chartComponent,
				},
			})
		}

		if len(row2Items) > 0 {
			row2 := &canvasRow{
				Items:  row2Items,
				Height: "400px",
			}
			doc.Rows = append(doc.Rows, row2)
		}
	}

	return marshalCanvasDashboardYAML(doc, false)
}

// canvasDashboardYAML is a struct for generating a canvas dashboard YAML file
type canvasDashboardYAML struct {
	Type        string          `yaml:"type,omitempty"`
	DisplayName string          `yaml:"display_name,omitempty"`
	Defaults    *canvasDefaults `yaml:"defaults,omitempty"`
	Rows        []*canvasRow    `yaml:"rows,omitempty"`
}

type canvasDefaults struct {
	TimeRange      string `yaml:"time_range,omitempty"`
	ComparisonMode string `yaml:"comparison_mode,omitempty"`
}

type canvasRow struct {
	Height string        `yaml:"height,omitempty"`
	Items  []*canvasItem `yaml:"items,omitempty"`
}

type canvasItem struct {
	Width     interface{}            `yaml:"width,omitempty"`
	Component map[string]interface{} `yaml:",inline"`
}

func marshalCanvasDashboardYAML(doc *canvasDashboardYAML, aiPowered bool) (string, error) {
	buf := new(bytes.Buffer)

	buf.WriteString("# Canvas Dashboard YAML\n")
	buf.WriteString("# Reference documentation: https://docs.rilldata.com/reference/project-files/canvas-dashboards\n")
	if aiPowered {
		buf.WriteString("# This file was generated using AI.\n")
	}
	buf.WriteString("\n")

	yamlBytes, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &rootNode); err != nil {
		return "", err
	}

	insertEmptyLinesInCanvasYaml(&rootNode)

	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(&rootNode); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func insertEmptyLinesInCanvasYaml(node *yaml.Node) {
	for i := 0; i < len(node.Content); i++ {
		if node.Content[i].Kind == yaml.MappingNode {
			for j := 0; j < len(node.Content[i].Content); j += 2 {
				keyNode := node.Content[i].Content[j]
				valueNode := node.Content[i].Content[j+1]

				if keyNode.Value == "rows" {
					keyNode.HeadComment = "\n"
				}
				insertEmptyLinesInCanvasYaml(valueNode)
			}
		} else if node.Content[i].Kind == yaml.SequenceNode {
			for j := 0; j < len(node.Content[i].Content); j++ {
				if node.Content[i].Content[j].Kind == yaml.MappingNode {
					node.Content[i].Content[j].HeadComment = "\n"
				}
			}
		}
	}
}
