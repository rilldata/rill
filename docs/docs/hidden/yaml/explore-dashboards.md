---
note: GENERATED. DO NOT EDIT.
title: Explore Dashboard YAML
sidebar_position: 37
---

Explore dashboards provide an interactive way to explore data with predefined metrics and dimensions.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `explore` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the explore dashboard _(required)_

### `metrics_view`

_[string]_ - Refers to the metrics view resource _(required)_

### `description`

_[string]_ - Refers to the description of the explore dashboard 

### `banner`

_[string]_ - Refers to the custom banner displayed at the header of an explore dashboard 

### `dimensions`

_[anyOf]_ - List of dimension names. Use '*' to select all dimensions (default)
 ```yaml
 # Example: Select a dimension
 dimensions:
   - country

 # Example: Select all dimensions except one
 dimensions:
   exclude:
     - country

 # Example: Select all dimensions that match a regex
 dimensions:
 regex: "^public_.*$"
 ```
 

  - **option 1** - _[string]_ - Simple field name as a string.

  - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

    - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

    - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

      - **`name`** - _[string]_ - Name of the field to select. _(required)_

      - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

### `measures`

_[anyOf]_ - List of measure names. Use '*' to select all measures (default) 

  - **option 1** - _[string]_ - Simple field name as a string.

  - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

    - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

    - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

      - **`name`** - _[string]_ - Name of the field to select. _(required)_

      - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

### `theme`

_[oneOf]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

### `time_range`

_[oneOf]_ - Default time range for the dashboard 

    - **`range`** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection _(required)_

    - **`comparison_offsets`** - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

        - **`offset`** - _[string]_ - Time offset for comparison (e.g., 'P1D' for one day ago) 

        - **`range`** - _[string]_ - Custom time range for comparison period 

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

### `lock_time_zone`

_[boolean]_ - When true, the dashboard will be locked to the first time provided in the time_zones list. When no time_zones are provided, the dashboard will be locked to UTC 

### `allow_custom_time_range`

_[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user. 

### `defaults`

_[object]_ - defines the defaults YAML struct
  ```yaml
  defaults: #define all the defaults within here
    dimensions:
      - dim_1
      - dim_2
    measures:
      - measure_1
      - measure_2
    time_range: P1M
    comparison_mode: dimension #time, none
    comparison_dimension: filename
  ```
 

  - **`dimensions`** - _[anyOf]_ - Provides the default dimensions to load on viewing the dashboard 

    - **option 1** - _[string]_ - Simple field name as a string.

    - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

      - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

      - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

        - **`name`** - _[string]_ - Name of the field to select. _(required)_

        - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

  - **`measures`** - _[anyOf]_ - Provides the default measures to load on viewing the dashboard 

    - **option 1** - _[string]_ - Simple field name as a string.

    - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

      - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

      - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

        - **`name`** - _[string]_ - Name of the field to select. _(required)_

        - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

  - **`time_range`** - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

  - **`comparison_mode`** - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: `none` (no comparison), `time` (compares with past based on default_time_range), `dimension` (compares based on comparison_dimension values) 

  - **`comparison_dimension`** - _[string]_ - for dimension mode, specify the comparison dimension by name 

### `embeds`

_[object]_ - Configuration options for embedded dashboard views 

  - **`hide_pivot`** - _[boolean]_ - When true, hides the pivot table view in embedded mode 

### `filters`

_[array of object]_ - Default filters to apply to the dashboard 

  - **`dimension`** - _[string]_ - Dimension to filter on 

  - **`value`** - _[no type]_ - Value to filter by 

  - **`operator`** - _[string]_ - Filter operator (eq, ne, in, etc.) 

### `security`

_[object]_ - Defines security rules and access control policies for dashboards (without row filtering) 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 