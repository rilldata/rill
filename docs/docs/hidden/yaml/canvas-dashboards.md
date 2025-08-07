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

### `max_width`

_[integer]_ - Max width in pixels of the canvas 

### `gap_x`

_[integer]_ - Horizontal gap in pixels of the canvas 

### `gap_y`

_[integer]_ - Vertical gap in pixels of the canvas 

### `filters`

_[object]_ - Indicates if filters should be enabled for the canvas. 

  - **`enable`** - _[boolean]_ - Toggles filtering functionality for the canvas dashboard. 

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

### `defaults`

_[object]_ - defines the defaults YAML struct
  ```yaml
  defaults: #define all the defaults within here
    time_range: P1M
    comparison_mode: dimension #time, none
    comparison_dimension: filename
  ```
 

  - **`time_range`** - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

  - **`comparison_mode`** - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: `none` (no comparison), `time` (compares with past based on default_time_range), `dimension` (compares based on comparison_dimension values) 

  - **`comparison_dimension`** - _[string]_ - for dimension mode, specify the comparison dimension by name 

### `theme`

_[oneOf]_ - Theme configuration. Can be either a string reference to an existing theme or an inline theme configuration object. 

### `security`

_[object]_ - Defines [security rules and access control policies](/manage/security) for dashboards (without row filtering) 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 