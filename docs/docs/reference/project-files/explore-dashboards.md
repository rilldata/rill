---
note: GENERATED. DO NOT EDIT.
title: Explore Dashboard YAML
sidebar_position: 36
---

Explore dashboards provide an interactive way to explore data with predefined measures and dimensions.

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

_[oneOf]_ - List of dimension names. Use '*' to select all dimensions (default) 

  - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

  - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select fields using a regular expression 

    - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

    - **`exclude`** - _[object]_ - Select all fields except those listed here 

```yaml
# Example: Select a dimension
dimensions:
    - country
```

```yaml
# Example: Select all dimensions except one
dimensions:
    exclude:
        - country
```

```yaml
# Example: Select all dimensions that match a regex
dimensions:
    expr: "^public_.*$"
```

### `measures`

_[oneOf]_ - List of measure names. Use '*' to select all measures (default) 

  - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

  - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select fields using a regular expression 

    - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

    - **`exclude`** - _[object]_ - Select all fields except those listed here 

```yaml
# Example: Select a measure
measures:
    - sum_of_total
```

```yaml
# Example: Select all measures except one
measures:
    exclude:
        - sum_of_total
```

```yaml
# Example: Select all measures that match a regex
measures:
    expr: "^public_.*$"
```

### `theme`

_[oneOf]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

  - **option 1** - _[string]_ - Name of an existing theme to apply to the dashboard

  - **option 2** - _[object]_ - Inline theme configuration.

    - **`colors`** - _[object]_ - Used to override the dashboard colors. Either primary or secondary color must be provided. 

      - **`primary`** - _[string]_ - Overrides the primary blue color in the dashboard. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. Note that the hue of the input colors is used for variants but the saturation and lightness is copied over from the [blue color palette](https://tailwindcss.com/docs/customizing-colors). 

      - **`secondary`** - _[string]_ - Overrides the secondary color in the dashboard. Applies to the loading spinner only as of now. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

    - **`light`** - _[object]_ - Light theme color configuration 

      - **`primary`** - _[string]_ - Primary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

      - **`secondary`** - _[string]_ - Secondary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

      - **`variables`** - _[object]_ - Custom CSS variables for light theme 

    - **`dark`** - _[object]_ - Dark theme color configuration 

      - **`primary`** - _[string]_ - Primary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

      - **`secondary`** - _[string]_ - Secondary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

      - **`variables`** - _[object]_ - Custom CSS variables for dark theme 

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
 

  - **`dimensions`** - _[oneOf]_ - Provides the default dimensions to load on viewing the dashboard 

    - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

    - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

    - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

      - **`regex`** - _[string]_ - Select fields using a regular expression 

      - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

      - **`exclude`** - _[object]_ - Select all fields except those listed here 

  - **`measures`** - _[oneOf]_ - Provides the default measures to load on viewing the dashboard 

    - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

    - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

    - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

      - **`regex`** - _[string]_ - Select fields using a regular expression 

      - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

      - **`exclude`** - _[object]_ - Select all fields except those listed here 

  - **`time_range`** - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

  - **`comparison_mode`** - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: `none` (no comparison), `time` (compares with past based on default_time_range), `dimension` (compares based on comparison_dimension values) 

  - **`comparison_dimension`** - _[string]_ - for dimension mode, specify the comparison dimension by name 

### `embeds`

_[object]_ - Configuration options for embedded dashboard views 

  - **`hide_pivot`** - _[boolean]_ - When true, hides the pivot table view in embedded mode 

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