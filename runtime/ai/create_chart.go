package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"go.uber.org/zap"
)

type CreateChart struct {
	Runtime *runtime.Runtime
}

var _ Tool[CreateChartArgs, *CreateChartResult] = (*CreateChart)(nil)

type CreateChartArgs map[string]any

type CreateChartResult struct {
	ChartType string         `json:"chart_type"`
	Spec      map[string]any `json:"spec"`
	Message   string         `json:"message"`
}

func (t *CreateChart) Spec() *mcp.Tool {
	var inputSchema *jsonschema.Schema
	err := json.Unmarshal([]byte(metricsview.ChartsJSONSchema), &inputSchema)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal input schema: %w", err))
	}

	return &mcp.Tool{
		Name:        "create_chart",
		Title:       "Create chart",
		Description: createChartDescription,
		InputSchema: inputSchema,
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating chart…",
			"openai/toolInvocation/invoked":  "Finished creating chart",
		},
	}
}

func (t *CreateChart) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)

	// Must be able to query metrics
	if !s.Claims().Can(runtime.ReadMetrics) {
		return false
	}

	// Only allow for rill user agents since it doesn't work with external MCP clients
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false
	}

	// Must have the chat_charts feature flag
	ff, err := t.Runtime.FeatureFlags(ctx, s.InstanceID(), s.Claims())
	if err != nil {
		if !errors.Is(err, ctx.Err()) {
			// TODO: Propagate error?
			s.logger.Error("failed to get feature flags", zap.Error(err))
		}
		return false
	}
	return ff["chat_charts"]
}

func (t *CreateChart) Handler(ctx context.Context, args CreateChartArgs) (*CreateChartResult, error) {
	s := GetSession(ctx)

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
		isValid := false
		for _, valid := range validTimeGrains {
			if timeGrainStr == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("'time_grain' must be one of: %v", validTimeGrains)
		}
	}

	// Validate that the metrics view exists
	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricsView}, false)
	if err != nil {
		return nil, fmt.Errorf("metrics view %q not found: %w", metricsView, err)
	}
	r, access, err := t.Runtime.ApplySecurityPolicy(ctx, s.InstanceID(), s.Claims(), r)
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, fmt.Errorf("resource not found")
	}
	if r.GetMetricsView().State.ValidSpec == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", metricsView)
	}

	// Return the chart specification in a structured format
	return &CreateChartResult{
		ChartType: chartType,
		Spec:      spec,
		Message:   fmt.Sprintf("Chart created successfully: %s", chartType),
	}, nil
}

const createChartDescription = `# Chart Visualization Tool

Create visualization charts based on metrics views. This tool generates chart specifications that will be rendered in the chat interface.

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

Example Specification: Plotting a bar chart of the top 20 advertisers by total bids
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

Example with filters: Bar chart showing top advertisers in specific countries
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

### 2. Line Chart (` + "`line_chart`" + `)
**Use for:** Showing trends over time

Example Specification: Line chart with monthly aggregation
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

### 3. Area Chart (` + "`area_chart`" + `)
**Use for:** Showing magnitude of change over time with filled areas

Example Specification
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


### 5. Normalized Stacked Bar Chart (` + "`stacked_bar_normalized`" + `)
**Use for:** Showing proportions instead of absolute values (100% stacked)

Example Specification
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

Example Specification with 1 dimension and 1 measure breakdown
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
    "stage": {
      "field": "stage",
      "limit": 15,
      "type": "nominal"
    }
  }
}
` + "```" + `

Example Specification with multiple measures breakdown
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
    "mode": "width"
  }
}
` + "```" + `

### 8. Heat Map (` + "`heatmap`" + `)
**Use for:** Visualizing data density using color intensity across two dimensions

Example Specification
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
      "field": "measures",
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
	- ` + "`\"color\"`" + ` or ` + "`\"-color\"`" + `: Sort by color field values (ascending/descending) Only used for heatmap charts
	- ` + "`\"measure\"`" + ` or ` + "`\"-measure\"`" + `: Sort by measure field values (ascending/descending) Only used for donut charts
  - Array of values for custom sort order (e.g., weekday names)
- **zeroBasedOrigin**: Start y-axis from zero (true/false)
- **showTotal**: Displays the measure total without any breakdown. Only used for donut chart to display totals in center

### Special Fields
- **__time**: Built-in time dimension field
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


## Important Notes

- The ` + "`time_range`" + ` parameter is **required** for all charts
- Time range ` + "`start`" + ` is inclusive, ` + "`end`" + ` is exclusive
- Use ` + "`time_grain`" + ` to control temporal aggregation granularity (defaults to "TIME_GRAIN_DAY")
- Use ` + "`where`" + ` to filter data displayed in charts - this applies the same filtering as query_metrics_view
- You do not always have to include a color field object for different bar chart and line charts. Use when required or when more than 1 dimensions has to be visualized.
- Ensure the metrics_view name matches exactly with available views
- Field names must match the exact field names in the metrics view
- When using ` + "`__time`" + ` field, set type to ` + "`\"temporal\"`" + `
- For multiple measures, use the ` + "`fields`" + ` array in the y-axis configuration`
