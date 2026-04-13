---
name: rill-canvas
description: Detailed instructions and examples for developing canvas dashboard resources in Rill
---

# Instructions for developing a canvas dashboard in Rill

## Introduction

Canvas dashboards are free-form dashboard resources that display custom chart and table components laid out in a grid. They enable building overview and report-style dashboards with multiple visualizations, similar to traditional business intelligence tools.

Canvas dashboards differ from explore dashboards in important ways:
- **Explore dashboards:** Best for explorative analysis, drill-down investigations, and letting users freely slice data by any dimension.
- **Canvas dashboards:** Best for fixed reports, executive summaries, or combining multiple metrics views into a single view.

Canvas dashboards are lightweight resources found downstream of metrics views in the project DAG. Each component within a canvas fetches data individually, typically from a metrics view resource.

**When to use canvas dashboards:**
- Building executive summaries with KPIs and multiple visualizations
- Creating report-style dashboards with markdown explanations
- Comparing metrics across different metrics views
- Designing custom layouts not possible with explore dashboards

## Canvas Structure

A canvas dashboard is defined in a YAML file with `type: canvas`. Here is the basic structure (most canvas dashboards work great without any of the optional properties here):

```yaml
type: canvas
display_name: "Sales Overview Dashboard"

# Optional filter settings
filters:
  enable: true
  pinned:
    - region
    - product_category

# Optional time range presets
time_ranges:
  - P7D
  - P30D
  - P90D
  - inf

# Optional maximum dashboard width
max_width: 1400

# Optional theme reference
theme: my_theme

# Default time settings for all components
defaults:
  time_range: P7D
  comparison_mode: time

# Optional security access control
security:
  access: "'{{ .user.domain }}' == 'company.com'"

# Required dashboard content organized in rows
rows:
  - height: 240px
    items:
      - width: 12
        kpi_grid:
          metrics_view: sales_metrics
          measures:
            - total_revenue
            - order_count

  - height: 400px
    items:
      - width: 6
        line_chart:
          metrics_view: sales_metrics
          title: "Revenue Trend"
          x:
            type: temporal
            field: event_time
          y:
            type: quantitative
            field: total_revenue
      - width: 6
        bar_chart:
          metrics_view: sales_metrics
          title: "Revenue by Region"
          color: primary
          x:
            type: nominal
            field: region
            limit: 10
            sort: -y
          y:
            type: quantitative
            field: total_revenue
```

## Layout System

Canvas dashboards use a 12-unit grid system for layout.

### Row Configuration

Each row defines a horizontal section with a specific height:

```yaml
rows:
  - height: 240px    # Row height in pixels
    items:
      # Components go here
```

**Recommended row heights:**
- Markdown headers: 40px - 80px
- KPI grids: 128px - 240px (depending on number of measures)
- Charts and visualizations: 300px - 500px
- Leaderboards: 300px - 450px
- Tables: 300px - 500px

### Item Widths

Items within a row share the 12-unit width:

```yaml
rows:
  # Full width (1 component per row)
  - items:
    - width: 12
      markdown:
        content: "# Dashboard Title"

  # Half width (2 components per row)
  - items:
    - width: 6     
      line_chart:
        # ...
    - width: 6
      bar_chart:
        # ...

  # Third width (3 components per row)
  - items:
    - width: 4
      donut_chart:
        # ...
    - width: 4
      bar_chart:
        # ...
    - width: 4
      area_chart:
        # ...
```

**Width guidelines:**
- `width: 12` - Full width; use for KPI grids, markdown headers, wide charts
- `width: 6` - Half width; use for side-by-side comparisons
- `width: 4` - Third width; use for three equal charts
- `width: 3` - Quarter width; use for four small components (minimum practical width)

## Dashboard Composition Best Practices

When building a new canvas dashboard, follow this recommended structure:

1. **Row 1 - Context**: Start with a markdown component providing dashboard title and overview
2. **Row 2 - Key Metrics**: Add a KPI grid with 2-4 of the most business-relevant measures
3. **Row 3 - Primary Analysis**: Split into two halves:
   - Left (width 6): A leaderboard showing top entities by a key dimension
   - Right (width 6): A time-series chart (line_chart or stacked_bar) showing trends
4. **Additional Rows**: Add 1-2 more rows with relevant charts based on the data

**Choosing chart types:**
- **Time-series analysis**: Use `line_chart` or `area_chart` with temporal x-axis
- **Categorical comparisons**: Use `bar_chart` or `stacked_bar` with nominal x-axis
- **Part-to-whole**: Use `donut_chart` or `stacked_bar_normalized`
- **Two-dimensional patterns**: Use `heatmap`
- **Dual-metric comparison**: Use `combo_chart` for two measures with different scales
- **Funnel analysis**: Use `funnel_chart` to visualize sequential stage drop-offs


# Field guidelines
The field names are case sensitive and should match exactly to the fields present in the metrics view.

**Time dimension restrictions:**
The time dimension (timeseries field from the metrics view) is special and can ONLY be used in the x-axis field for temporal charts. Never use the time dimension in:
- Leaderboard dimensions
- Color fields
- Any other dimension configuration

## Component Types

### Markdown

Add text content, headers, and documentation:

```yaml
markdown:
  content: |
    ## Dashboard Overview

    This dashboard tracks key sales metrics across all regions.

    ---
  alignment:
    horizontal: left    # left, center, right
    vertical: middle    # top, middle, bottom
```

**Best practices:**
- Use markdown for dashboard titles, section headers, and explanatory text
- Add blank lines between markdown elements for proper rendering
- Use `---` for horizontal rules to separate sections

### KPI Grid

Display key metrics with comparison values and sparklines:

```yaml
kpi_grid:
  metrics_view: sales_metrics
  measures:
    - total_revenue
    - order_count
    - average_order_value
    - customer_count
  comparison:
    - delta           # Absolute change
    - percent_change  # Percentage change
    - previous        # Previous period value
  sparkline: right    # right, bottom, none
```

**With dimension filters:**

```yaml
kpi_grid:
  metrics_view: sales_metrics
  measures:
    - total_revenue
    - order_count
  dimension_filters: region IN ('North America', 'Europe')
  comparison:
    - percent_change
  sparkline: bottom
  hide_time_range: true
```

### Leaderboard

Display ranked dimension values by measures:

```yaml
leaderboard:
  metrics_view: sales_metrics
  title: "Top Products"
  description: "Products ranked by total revenue"
  dimensions:
    - product_category
  measures:
    - total_revenue
    - order_count
  num_rows: 10
```

**With multiple dimensions:**

```yaml
leaderboard:
  metrics_view: sales_metrics
  dimensions:
    - region
    - product_category
  measures:
    - total_revenue
    - average_order_value
    - order_count
  num_rows: 7
```

**Important:** Never use time dimensions in leaderboard dimensions. Leaderboards are for categorical ranking, not time-series analysis.

### Line Chart

Show trends over time:

```yaml
line_chart:
  metrics_view: sales_metrics
  title: "Revenue Trend"
  color: primary
  x:
    field: order_date
    type: temporal
    limit: 30
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

**With color dimension breakdown:**

```yaml
line_chart:
  metrics_view: sales_metrics
  title: "Revenue by Region"
  color:
    field: region
    type: nominal
    limit: 5
    legendOrientation: top
  x:
    field: order_date
    type: temporal
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

**With custom color mapping:**

```yaml
line_chart:
  metrics_view: sales_metrics
  title: "Performance Comparison"
  color:
    field: status
    type: nominal
    colorMapping:
      - value: "active"
        color: hsl(120, 70%, 45%)
      - value: "inactive"
        color: hsl(0, 70%, 50%)
  x:
    field: event_date
    type: temporal
  y:
    field: event_count
    type: quantitative
```

### Bar Chart

Compare values across categories:

```yaml
bar_chart:
  metrics_view: sales_metrics
  title: "Revenue by Product Category"
  color: hsl(210, 70%, 50%)
  x:
    field: product_category
    type: nominal
    limit: 10
    sort: -y
    labelAngle: 0
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

**With color dimension:**

```yaml
bar_chart:
  metrics_view: sales_metrics
  title: "Revenue by Category and Region"
  color:
    field: region
    type: nominal
    limit: 5
  x:
    field: product_category
    type: nominal
    limit: 8
    sort: -y
  y:
    field: total_revenue
    type: quantitative
```

### Stacked Bar

Show cumulative values across categories or time:

```yaml
stacked_bar:
  metrics_view: sales_metrics
  title: "Revenue Over Time by Region"
  color:
    field: region
    type: nominal
    limit: 5
  x:
    field: order_date
    type: temporal
    limit: 20
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

**With multiple measures:**

```yaml
stacked_bar:
  metrics_view: sales_metrics
  title: "Cost Breakdown Over Time"
  color:
    field: rill_measures
    type: value
    legendOrientation: top
  x:
    field: order_date
    type: temporal
    limit: 20
  y:
    field: cost_of_goods
    fields:
      - cost_of_goods
      - shipping_cost
      - marketing_cost
    type: quantitative
    zeroBasedOrigin: true
```

### Stacked Bar Normalized

Show proportional distribution (100% stacked):

```yaml
stacked_bar_normalized:
  metrics_view: sales_metrics
  title: "Revenue Share by Region"
  color:
    field: region
    type: nominal
    limit: 5
  x:
    field: order_date
    type: temporal
    limit: 20
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

**With custom color mapping for measures:**

```yaml
stacked_bar_normalized:
  metrics_view: inventory_metrics
  title: "Inventory Status Distribution"
  color:
    field: rill_measures
    type: value
    legendOrientation: top
    colorMapping:
      - value: "in_stock"
        color: hsl(120, 60%, 50%)
      - value: "low_stock"
        color: hsl(45, 90%, 50%)
      - value: "out_of_stock"
        color: hsl(0, 70%, 50%)
  x:
    field: report_date
    type: temporal
    limit: 20
  y:
    field: in_stock
    fields:
      - in_stock
      - low_stock
      - out_of_stock
    type: quantitative
```

### Area Chart

Show magnitude over time with optional stacking:

```yaml
area_chart:
  metrics_view: sales_metrics
  title: "Order Volume Over Time"
  color: primary
  x:
    field: order_date
    type: temporal
    limit: 30
  y:
    field: order_count
    type: quantitative
    zeroBasedOrigin: true
```

**With color dimension:**

```yaml
area_chart:
  metrics_view: sales_metrics
  title: "Revenue by Channel"
  color:
    field: sales_channel
    type: nominal
    limit: 4
  x:
    field: order_date
    type: temporal
    limit: 20
  y:
    field: total_revenue
    type: quantitative
    zeroBasedOrigin: true
```

### Donut Chart

Show proportional breakdown:

```yaml
donut_chart:
  metrics_view: sales_metrics
  title: "Revenue by Region"
  innerRadius: 50
  color:
    field: region
    type: nominal
    limit: 8
    sort: -measure
  measure:
    field: total_revenue
    type: quantitative
    showTotal: true
```

### Heatmap

Show patterns across two dimensions:

```yaml
heatmap:
  metrics_view: activity_metrics
  title: "Activity by Day and Hour"
  color:
    field: event_count
    type: quantitative
  x:
    field: day_of_week
    type: nominal
    limit: 7
  y:
    field: hour_of_day
    type: nominal
    limit: 24
    sort: -color
```

**With custom color range:**

```yaml
heatmap:
  metrics_view: performance_metrics
  title: "Performance Score Matrix"
  color:
    field: score
    type: quantitative
    colorRange:
      mode: scheme
      scheme: sequential
  x:
    field: category
    type: nominal
    limit: 10
  y:
    field: subcategory
    type: nominal
    limit: 15
```

**With custom Vega-Lite config for colors:**

```yaml
heatmap:
  metrics_view: utilization_metrics
  title: "Resource Utilization"
  vl_config: |
    {
      "range": {
        "heatmap": ["#F4A261", "#D63946", "#457B9D"]
      }
    }
  color:
    field: utilization_rate
    type: quantitative
  x:
    field: resource_name
    type: nominal
    limit: 20
  y:
    field: time_slot
    type: nominal
    limit: 12
```

### Combo Chart

Combine bar and line on dual axes:

```yaml
combo_chart:
  metrics_view: sales_metrics
  title: "Revenue and Order Count"
  color:
    field: measures
    type: value
    legendOrientation: top
  x:
    field: order_date
    type: temporal
    limit: 20
  y1:
    field: total_revenue
    type: quantitative
    mark: bar
    zeroBasedOrigin: true
  y2:
    field: order_count
    type: quantitative
    mark: line
    zeroBasedOrigin: true
```

**With custom color mapping:**

```yaml
combo_chart:
  metrics_view: funnel_metrics
  title: "Conversions and Conversion Rate"
  color:
    field: measures
    type: value
    legendOrientation: top
    colorMapping:
      - value: "Conversions"
        color: hsl(210, 100%, 73%)
      - value: "Conversion Rate"
        color: hsl(280, 70%, 55%)
  x:
    field: event_date
    type: temporal
    limit: 30
  y1:
    field: conversions
    type: quantitative
    mark: bar
  y2:
    field: conversion_rate
    type: quantitative
    mark: line
```

### Funnel Chart

Show flow through stages or conversion processes:

```yaml
funnel_chart:
  metrics_view: conversion_metrics
  title: "Conversion Funnel"
  breakdownMode: dimension
  color: stage
  mode: width
  stage:
    field: funnel_stage
    type: nominal
    limit: 10
  measure:
    field: user_count
    type: quantitative
```

**With multiple measures breakdown:**

```yaml
funnel_chart:
  metrics_view: engagement_metrics
  title: "Engagement Funnel"
  breakdownMode: measures
  color: value
  mode: width
  measure:
    field: impressions
    type: quantitative
    fields:
      - impressions
      - clicks
      - signups
      - purchases
```

**Breakdown modes and color options:**
- `breakdownMode: dimension` with `color: stage` (different colors per stage) or `color: measure` (similar colors by value)
- `breakdownMode: measures` with `color: name` (different colors per measure) or `color: value` (similar colors by value)

### Pivot

Create pivot tables with row and column dimensions:

```yaml
pivot:
  metrics_view: sales_metrics
  title: "Sales by Region and Category"
  row_dimensions:
    - region
    - product_category
  col_dimensions:
    - quarter
  measures:
    - total_revenue
    - order_count
    - average_order_value
```

**Simple pivot (rows only):**

```yaml
pivot:
  metrics_view: sales_metrics
  row_dimensions:
    - region
  col_dimensions: []
  measures:
    - total_revenue
    - order_count
    - margin_rate
```

### Table

Display tabular data with specified columns:

```yaml
table:
  metrics_view: sales_metrics
  title: "Product Performance"
  description: "Detailed breakdown of product metrics"
  columns:
    - product_name
    - product_category
    - total_revenue
    - order_count
    - average_price
```

**With dimension filters:**

```yaml
table:
  metrics_view: sales_metrics
  title: "North America Sales"
  columns:
    - product_name
    - total_revenue
    - order_count
  dimension_filters: region IN ('North America')
```

### Image

Display external images:

```yaml
image:
  url: https://example.com/logo.png
  alignment:
    horizontal: center
    vertical: middle
```

## Field Configuration

### Data Types

- **`nominal`**: Categorical data (strings, categories). Use for dimensions.
- **`temporal`**: Time-based data (dates, timestamps). Use for time dimensions.
- **`quantitative`**: Numerical data (counts, amounts). Use for measures.
- **`value`**: Special type for multiple measures. Use only in color field with `rill_measures`.

### Axis Properties

```yaml
x:
  field: category_name       # Field name from metrics view
  type: nominal              # Data type
  limit: 10                  # Max values to display
  sort: -y                   # Sort order (see below)
  showNull: true             # Include null values
  labelAngle: 45             # Label rotation angle
```

### Sort Options

- `"x"` or `"-x"`: Sort by x-axis values (ascending/descending)
- `"y"` or `"-y"`: Sort by y-axis values (ascending/descending)
- `"color"` or `"-color"`: Sort by color field (heatmaps)
- `"measure"` or `"-measure"`: Sort by measure (donut charts)
- Array of values: Custom sort order (e.g., `["Mon", "Tue", "Wed"]`)

### Y-Axis Properties

```yaml
y:
  field: total_revenue
  type: quantitative
  zeroBasedOrigin: true      # Start y-axis at zero
```

**Multiple measures:**

```yaml
y:
  field: revenue
  fields:
    - revenue
    - cost
    - profit
  type: quantitative
```

### Color Configuration

**Simple color string:**

```yaml
color: primary              # Named color
color: secondary
color: "#FF5733"            # Hex color
color: hsl(210, 70%, 50%)   # HSL color
```

**Field-based color:**

```yaml
color:
  field: region
  type: nominal
  limit: 10
  legendOrientation: top    # top, bottom, left, right, none
```

**Custom color mapping:**

```yaml
color:
  field: status
  type: nominal
  colorMapping:
    - value: "success"
      color: hsl(120, 70%, 45%)
    - value: "warning"
      color: hsl(45, 90%, 50%)
    - value: "error"
      color: hsl(0, 70%, 50%)
```

**Color scheme:**

```yaml
color:
  field: score
  type: quantitative
  colorRange:
    mode: scheme
    scheme: sequential
```

### Special Field: rill_measures

Use `rill_measures` in the color field when displaying multiple measures in stacked charts:

```yaml
color:
  field: rill_measures
  type: value
  legendOrientation: top
y:
  field: revenue
  fields:
    - revenue
    - cost
    - profit
  type: quantitative
```

## Advanced Features

### Dimension Filters

Filter component data without affecting other components:

```yaml
kpi_grid:
  metrics_view: sales_metrics
  measures:
    - total_revenue
  dimension_filters: region IN ('North America') AND status IN ('active')
```

### Time Range Override

Override the default time range for a specific component:

```yaml
heatmap:
  metrics_view: activity_metrics
  time_range:
    preset: last_7_days
  # ... other config
```

### Time Filters

Override time settings with detailed control:

```yaml
stacked_bar:
  metrics_view: sales_metrics
  time_filters: tr=P12M&compare_tr=rill-PY&grain=week
  # ... other config
```

### Vega-Lite Configuration

Customize chart appearance with Vega-Lite config:

```yaml
bar_chart:
  metrics_view: sales_metrics
  vl_config: |
    {
      "axisX": {
        "grid": true,
        "labelAngle": 45
      },
      "range": {
        "category": ["#D63946", "#457B9D", "#F4A261", "#2A9D8F"]
      }
    }
  # ... other config
```

## Complete Example

```yaml
type: canvas
display_name: "Monthly Business Report"

defaults:
  time_range: P30D
  comparison_mode: time

max_width: 1400
theme: corporate_theme

rows:
  - height: 100px
    items:
      - width: 12
        markdown:
          content: |
            # Monthly Business Report

            Comprehensive overview of business performance metrics.

            ---
          alignment:
            horizontal: center
            vertical: middle

  - height: 50px
    items:
      - width: 12
        markdown:
          content: "## Key Metrics"
          alignment:
            horizontal: left
            vertical: middle

  - height: 200px
    items:
      - width: 12
        kpi_grid:
          metrics_view: business_metrics
          measures:
            - revenue
            - profit
            - customers
            - orders
          comparison:
            - percent_change
            - previous
          sparkline: right

  - height: 50px
    items:
      - width: 12
        markdown:
          content: "## Revenue Analysis"
          alignment:
            horizontal: left
            vertical: middle

  - height: 400px
    items:
      - width: 8
        combo_chart:
          metrics_view: business_metrics
          title: "Revenue and Profit Margin"
          color:
            field: measures
            type: value
            legendOrientation: top
          x:
            field: report_date
            type: temporal
            limit: 30
          y1:
            field: revenue
            type: quantitative
            mark: bar
          y2:
            field: profit_margin
            type: quantitative
            mark: line

      - width: 4
        donut_chart:
          metrics_view: business_metrics
          title: "Revenue by Segment"
          innerRadius: 50
          color:
            field: customer_segment
            type: nominal
            limit: 5
          measure:
            field: revenue
            type: quantitative
            showTotal: true

  - height: 50px
    items:
      - width: 12
        markdown:
          content: "## Regional Performance"
          alignment:
            horizontal: left
            vertical: middle

  - height: 350px
    items:
      - width: 6
        leaderboard:
          metrics_view: business_metrics
          dimensions:
            - region
          measures:
            - revenue
            - profit
            - order_count
          num_rows: 8

      - width: 6
        heatmap:
          metrics_view: business_metrics
          title: "Revenue by Region and Product"
          color:
            field: revenue
            type: quantitative
          x:
            field: product_category
            type: nominal
            limit: 8
          y:
            field: region
            type: nominal
            limit: 6
```