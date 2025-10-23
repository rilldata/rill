---
note: GENERATED. DO NOT EDIT.
title: Canvas Dashboard YAML
sidebar_position: 35
---

Canvas dashboards provide a flexible way to create custom dashboards with drag-and-drop components.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `canvas` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the canvas _(required)_

### `title`

_[string]_ - Deprecated: use display_name instead. Refers to the display name for the canvas 

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

### `allow_filter_add`

_[boolean]_ - Whether users can add new filters to the canvas dashboard. 

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
 

  - **option 1** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

  - **option 2** - _[object]_ - Object containing time range and comparison configuration

    - **`range`** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection _(required)_

    - **`comparison_offsets`** - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

      - **option 1** - _[string]_ - Offset string only (range is inferred)

      - **option 2** - _[object]_ - Object containing offset and range configuration for time comparison

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
    filters:
      dimensions:
        - dimension: country
          values: ["US", "CA"]
          mode: in_list
      measures:
        - measure: revenue
          operator: gt
          values: ["1000"]
  ```
 

  - **`time_range`** - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

  - **`comparison_mode`** - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: `none` (no comparison), `time` (compares with past based on default_time_range), `dimension` (compares based on comparison_dimension values) 

  - **`comparison_dimension`** - _[string]_ - for dimension mode, specify the comparison dimension by name 

  - **`filters`** - _[object]_ - Default filter configuration 

    - **`dimensions`** - _[array of object]_ - List of default dimension filters 

      - **`dimension`** - _[string]_ - Name of the dimension to filter on 

      - **`values`** - _[array of string]_ - List of values to filter by 

      - **`limit`** - _[integer]_ - Maximum number of values to show in the filter 

      - **`removable`** - _[boolean]_ - Whether the filter can be removed by the user 

      - **`locked`** - _[boolean]_ - Whether the filter is locked and cannot be modified 

      - **`hidden`** - _[boolean]_ - Whether the filter is hidden from the UI 

      - **`mode`** - _[string]_ - Filter mode - select for dropdown, in_list for multi-select, contains for text search 

      - **`exclude`** - _[boolean]_ - Whether to exclude the specified values instead of including them 

    - **`measures`** - _[array of object]_ - List of default measure filters 

      - **`measure`** - _[string]_ - Name of the measure to filter on 

      - **`by_dimension`** - _[string]_ - Dimension to group the measure filter by 

      - **`operator`** - _[string]_ - Operator for the measure filter (e.g., eq, gt, lt, gte, lte) 

      - **`values`** - _[array of string]_ - List of values to filter by 

      - **`removable`** - _[boolean]_ - Whether the filter can be removed by the user 

      - **`locked`** - _[boolean]_ - Whether the filter is locked and cannot be modified 

      - **`hidden`** - _[boolean]_ - Whether the filter is hidden from the UI 

### `theme`

_[oneOf]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

  - **option 1** - _[string]_ - Name of an existing theme to apply to the dashboard

  - **option 2** - _[object]_ - Inline theme configuration.

    - **`colors`** - _[object]_ - **DEPRECATED**: Use `light` and `dark` properties instead. Legacy color override for dashboards. Cannot be used together with `light` or `dark` properties. 

      - **`primary`** - _[string]_ - **DEPRECATED**: Overrides the primary blue color in the dashboard. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

      - **`secondary`** - _[string]_ - **DEPRECATED**: Overrides the secondary color in the dashboard. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

    - **`light`** - _[object]_ - Color customization for light mode. Supports CSS color values (hex, named colors, hsl, etc.). All properties are optional. 

      - **`primary`** - _[string]_ - Primary theme color 

      - **`secondary`** - _[string]_ - Secondary theme color 

      - **`surface`** - _[string]_ - Surface color 

      - **`background`** - _[string]_ - Background color 

      - **`color-sequential-1`** - _[string]_ - Sequential palette color 1 (lightest) 

      - **`color-sequential-2`** - _[string]_ - Sequential palette color 2 

      - **`color-sequential-3`** - _[string]_ - Sequential palette color 3 

      - **`color-sequential-4`** - _[string]_ - Sequential palette color 4 

      - **`color-sequential-5`** - _[string]_ - Sequential palette color 5 (medium) 

      - **`color-sequential-6`** - _[string]_ - Sequential palette color 6 

      - **`color-sequential-7`** - _[string]_ - Sequential palette color 7 

      - **`color-sequential-8`** - _[string]_ - Sequential palette color 8 

      - **`color-sequential-9`** - _[string]_ - Sequential palette color 9 (darkest) 

      - **`color-diverging-1`** - _[string]_ - Diverging palette color 1 

      - **`color-diverging-2`** - _[string]_ - Diverging palette color 2 

      - **`color-diverging-3`** - _[string]_ - Diverging palette color 3 

      - **`color-diverging-4`** - _[string]_ - Diverging palette color 4 

      - **`color-diverging-5`** - _[string]_ - Diverging palette color 5 

      - **`color-diverging-6`** - _[string]_ - Diverging palette color 6 (neutral) 

      - **`color-diverging-7`** - _[string]_ - Diverging palette color 7 

      - **`color-diverging-8`** - _[string]_ - Diverging palette color 8 

      - **`color-diverging-9`** - _[string]_ - Diverging palette color 9 

      - **`color-diverging-10`** - _[string]_ - Diverging palette color 10 

      - **`color-diverging-11`** - _[string]_ - Diverging palette color 11 

      - **`color-qualitative-1`** - _[string]_ - Qualitative palette color 1 

      - **`color-qualitative-2`** - _[string]_ - Qualitative palette color 2 

      - **`color-qualitative-3`** - _[string]_ - Qualitative palette color 3 

      - **`color-qualitative-4`** - _[string]_ - Qualitative palette color 4 

      - **`color-qualitative-5`** - _[string]_ - Qualitative palette color 5 

      - **`color-qualitative-6`** - _[string]_ - Qualitative palette color 6 

      - **`color-qualitative-7`** - _[string]_ - Qualitative palette color 7 

      - **`color-qualitative-8`** - _[string]_ - Qualitative palette color 8 

      - **`color-qualitative-9`** - _[string]_ - Qualitative palette color 9 

      - **`color-qualitative-10`** - _[string]_ - Qualitative palette color 10 

      - **`color-qualitative-11`** - _[string]_ - Qualitative palette color 11 

      - **`color-qualitative-12`** - _[string]_ - Qualitative palette color 12 

      - **`color-qualitative-13`** - _[string]_ - Qualitative palette color 13 

      - **`color-qualitative-14`** - _[string]_ - Qualitative palette color 14 

      - **`color-qualitative-15`** - _[string]_ - Qualitative palette color 15 

      - **`color-qualitative-16`** - _[string]_ - Qualitative palette color 16 

      - **`color-qualitative-17`** - _[string]_ - Qualitative palette color 17 

      - **`color-qualitative-18`** - _[string]_ - Qualitative palette color 18 

      - **`color-qualitative-19`** - _[string]_ - Qualitative palette color 19 

      - **`color-qualitative-20`** - _[string]_ - Qualitative palette color 20 

      - **`color-qualitative-21`** - _[string]_ - Qualitative palette color 21 

      - **`color-qualitative-22`** - _[string]_ - Qualitative palette color 22 

      - **`color-qualitative-23`** - _[string]_ - Qualitative palette color 23 

      - **`color-qualitative-24`** - _[string]_ - Qualitative palette color 24 

    - **`dark`** - _[object]_ - Color customization for dark mode. Supports CSS color values (hex, named colors, hsl, etc.). All properties are optional. 

      - **`primary`** - _[string]_ - Primary theme color 

      - **`secondary`** - _[string]_ - Secondary theme color 

      - **`surface`** - _[string]_ - Surface color 

      - **`background`** - _[string]_ - Background color 

      - **`color-sequential-1`** - _[string]_ - Sequential palette color 1 (lightest) 

      - **`color-sequential-2`** - _[string]_ - Sequential palette color 2 

      - **`color-sequential-3`** - _[string]_ - Sequential palette color 3 

      - **`color-sequential-4`** - _[string]_ - Sequential palette color 4 

      - **`color-sequential-5`** - _[string]_ - Sequential palette color 5 (medium) 

      - **`color-sequential-6`** - _[string]_ - Sequential palette color 6 

      - **`color-sequential-7`** - _[string]_ - Sequential palette color 7 

      - **`color-sequential-8`** - _[string]_ - Sequential palette color 8 

      - **`color-sequential-9`** - _[string]_ - Sequential palette color 9 (darkest) 

      - **`color-diverging-1`** - _[string]_ - Diverging palette color 1 

      - **`color-diverging-2`** - _[string]_ - Diverging palette color 2 

      - **`color-diverging-3`** - _[string]_ - Diverging palette color 3 

      - **`color-diverging-4`** - _[string]_ - Diverging palette color 4 

      - **`color-diverging-5`** - _[string]_ - Diverging palette color 5 

      - **`color-diverging-6`** - _[string]_ - Diverging palette color 6 (neutral) 

      - **`color-diverging-7`** - _[string]_ - Diverging palette color 7 

      - **`color-diverging-8`** - _[string]_ - Diverging palette color 8 

      - **`color-diverging-9`** - _[string]_ - Diverging palette color 9 

      - **`color-diverging-10`** - _[string]_ - Diverging palette color 10 

      - **`color-diverging-11`** - _[string]_ - Diverging palette color 11 

      - **`color-qualitative-1`** - _[string]_ - Qualitative palette color 1 

      - **`color-qualitative-2`** - _[string]_ - Qualitative palette color 2 

      - **`color-qualitative-3`** - _[string]_ - Qualitative palette color 3 

      - **`color-qualitative-4`** - _[string]_ - Qualitative palette color 4 

      - **`color-qualitative-5`** - _[string]_ - Qualitative palette color 5 

      - **`color-qualitative-6`** - _[string]_ - Qualitative palette color 6 

      - **`color-qualitative-7`** - _[string]_ - Qualitative palette color 7 

      - **`color-qualitative-8`** - _[string]_ - Qualitative palette color 8 

      - **`color-qualitative-9`** - _[string]_ - Qualitative palette color 9 

      - **`color-qualitative-10`** - _[string]_ - Qualitative palette color 10 

      - **`color-qualitative-11`** - _[string]_ - Qualitative palette color 11 

      - **`color-qualitative-12`** - _[string]_ - Qualitative palette color 12 

      - **`color-qualitative-13`** - _[string]_ - Qualitative palette color 13 

      - **`color-qualitative-14`** - _[string]_ - Qualitative palette color 14 

      - **`color-qualitative-15`** - _[string]_ - Qualitative palette color 15 

      - **`color-qualitative-16`** - _[string]_ - Qualitative palette color 16 

      - **`color-qualitative-17`** - _[string]_ - Qualitative palette color 17 

      - **`color-qualitative-18`** - _[string]_ - Qualitative palette color 18 

      - **`color-qualitative-19`** - _[string]_ - Qualitative palette color 19 

      - **`color-qualitative-20`** - _[string]_ - Qualitative palette color 20 

      - **`color-qualitative-21`** - _[string]_ - Qualitative palette color 21 

      - **`color-qualitative-22`** - _[string]_ - Qualitative palette color 22 

      - **`color-qualitative-23`** - _[string]_ - Qualitative palette color 23 

      - **`color-qualitative-24`** - _[string]_ - Qualitative palette color 24 

### `security`

_[object]_ - Defines [security rules and access control policies](/build/metrics-view/security) for dashboards (without row filtering) 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    - **option 1** - _[string]_ - SQL expression that evaluates to a boolean to determine access

    - **option 2** - _[boolean]_ - Direct boolean value to allow or deny access

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 