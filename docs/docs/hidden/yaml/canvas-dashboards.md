---
note: GENERATED. DO NOT EDIT.
title: Canvas Dashboard YAML
sidebar_position: 36
---

Canvas dashboards provide a flexible way to create custom dashboards with drag-and-drop components.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `canvas` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the canvas _(required)_

### `description`

_[string]_ - Description for the canvas dashboard 

### `banner`

_[string]_ - Refers to the custom banner displayed at the header of an Canvas dashboard 

### `rows`

_[array of object]_ - Refers to all of the rows displayed on the Canvas 

  - **`height`** - _[string]_ - Height of the row in px 

  - **`items`** - _[array of object]_ - List of components to display in the row 

    - **`component`** - _[string]_ - Name of the component to display. Each component type has its own set of properties.
    Available component types:
    
        - **markdown** - Text component, uses markdown formatting
        - **kpi_grid** - KPI component, similar to TDD in Rill Explore, display quick KPI charts
        - **stacked_bar_normalized** - Bar chart normalized to 100% values
        - **line_chart** - Normal Line chart
        - **bar_chart** - Normal Bar chart
        - **stacked_bar** - Stacked Bar chart
        - **area_chart** - Line chart with area
        - **image** - Provide a URL to embed into canvas dashboard
        - **table** - Similar to Pivot table, add dimensions and measures to visualize your data
        - **heatmap** - Heat Map chart to visualize distribution of data
        - **donut_chart** - Donut or Pie chart to display sums of total
 

    - **`width`** - _[string, integer]_ - Width of the component (can be a number or string with unit) 

### `components`

_[array of object]_ - Array of components to display on the canvas 

  - **`type`** - _[string]_ - Type of component (chart, table, text, etc.) 

  - **`title`** - _[string]_ - Title for the component 

  - **`data`** - _[oneOf]_ - Data source for the component 

      - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

      - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

      - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

      - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

      - **`args`** - _[object]_ - Arguments to pass to the custom API. 

      - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

        - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

        - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

      - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

      - **`resource_status`** - _[object]_ - Based on resource status _(required)_

        - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

```yaml
resource_status:
  where_error: true
```


  - **`layout`** - _[object]_ - Layout configuration for the component 

  - **`style`** - _[object]_ - Styling configuration for the component 

### `max_width`

_[integer]_ - Max width in pixels of the canvas 

### `gap_x`

_[integer]_ - Horizontal gap in pixels of the canvas 

### `gap_y`

_[integer]_ - Vertical gap in pixels of the canvas 

### `theme`

_[oneOf]_ - Theme configuration. Can be either a string reference to an existing theme or an inline theme configuration object. 

### `allow_custom_time_range`

_[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user. 

### `time_ranges`

_[array of oneOf]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets'
  ```yaml
  time_ranges:
    - PT15M // Simplified syntax to specify only the range
    - PT1H
    - PT6H
    - P7D
    - range: P5D // Advanced syntax to specify comparison_offsets as well
    - P4W
    - rill-TD // Today
    - rill-WTD // Week-To-date
  ```
 

    - **`range`** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection _(required)_

    - **`comparison_offsets`** - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

        - **`offset`** - _[string]_ - Time offset for comparison (e.g., 'P1D' for one day ago) 

        - **`range`** - _[string]_ - Custom time range for comparison period 

### `time_zones`

_[array of string]_ - Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) 

### `filters`

_[object]_ - Indicates if filters should be enabled for the canvas. 

### `security`

_[object]_ - Defines security rules and access control policies for dashboards (without row filtering) 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 