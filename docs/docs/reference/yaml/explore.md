---
note: GENERATED. DO NOT EDIT.
title: Explore YAML
sidebar_position: 36
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `explore` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the explore dashboard 

### `description`

_[string]_ - Refers to the description of the explore dashboard 

### `banner`

_[string]_ - Refers to the custom banner displayed at the header of an explore dashboard 

### `metrics_view`

_[string]_ - Refers to the metrics view resource 

### `dimensions`

_[oneOf]_ - List of dimension names. Use '*' to select all dimensions (default)  

  - **option 1** - _[string]_ - Matches all fields

  - **option 2** - _[array of string]_ - Explicit list of fields

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select dimensions using a regular expression 

    - **`expr`** - _[string]_ - (no description) 

    - **`exclude`** - _[object]_ - Select all dimensions except those listed here 

### `measures`

_[oneOf]_ -  List of measure names. Use '*' to select all measures (default) 

  - **option 1** - _[string]_ - Matches all fields

  - **option 2** - _[array of string]_ - Explicit list of fields

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select dimensions using a regular expression 

    - **`expr`** - _[string]_ - (no description) 

    - **`exclude`** - _[object]_ - Select all dimensions except those listed here 

### `theme`

_[oneOf]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

  - **option 1** - _[string]_ - (no description)

  - **option 2** - _[object]_ - (no description)

    - **`colors`** - _[object]_ - Used to override the dashboard colors. Either primary or secondary color must be provided. 

      - **`primary`** - _[string]_ - Overrides the primary blue color in the dashboard. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. Note that the hue of the input colors is used for variants but the saturation and lightness is copied over from the [blue color palette](https://tailwindcss.com/docs/customizing-colors). 

      - **`secondary`** - _[string]_ - Overrides the secondary color in the dashboard. Applies to the loading spinner only as of now. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

### `time_ranges`

_[array of oneOf]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets' 

  - **option 1** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

  - **option 2** - _[object]_ - (no description)

    - **`range`** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection _(required)_

    - **`comparison_offsets`** - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

      - **option 1** - _[string]_ - Offset string only (range is inferred)

      - **option 2** - _[object]_ - (no description)

        - **`offset`** - _[string]_ - (no description) 

        - **`range`** - _[string]_ - (no description) 

### `time_zones`

_[array of string]_ - Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) 

### `lock_time_zone`

_[boolean]_ - When true, the dashboard will be locked to the first time provided in the time_zones list. When no time_zones are provided, the dashboard will be locked to UTC 

### `allow_custom_time_range`

_[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user. 

### `defaults`

_[object]_ - defines the defaults YAML struct 

  - **`dimensions`** - _[oneOf]_ - Provides the default dimensions to load on viewing the dashboard 

    - **option 1** - _[string]_ - Matches all fields

    - **option 2** - _[array of string]_ - Explicit list of fields

    - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

      - **`regex`** - _[string]_ - Select dimensions using a regular expression 

      - **`expr`** - _[string]_ - (no description) 

      - **`exclude`** - _[object]_ - Select all dimensions except those listed here 

  - **`measures`** - _[oneOf]_ - Provides the default measures to load on viewing the dashboard 

    - **option 1** - _[string]_ - Matches all fields

    - **option 2** - _[array of string]_ - Explicit list of fields

    - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

      - **`regex`** - _[string]_ - Select dimensions using a regular expression 

      - **`expr`** - _[string]_ - (no description) 

      - **`exclude`** - _[object]_ - Select all dimensions except those listed here 

  - **`time_range`** - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

  - **`comparison_mode`** - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: 'none' (no comparison), 'time' (compares with past based on default_time_range), 'dimension' (compares based on comparison_dimension values) 

  - **`comparison_dimension`** - _[string]_ - for dimension mode, specify the comparison dimension by name 

### `embeds`

_[object]_ - (no description) 

  - **`hide_pivot`** - _[boolean]_ - (no description) 

### `security`

_[object]_ - Security rules to apply for access to the explore dashboard 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    - **option 1** - _[string]_ - (no description)

    - **option 2** - _[boolean]_ - (no description)

  - **`row_filter`** - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`include`** - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

    - **`if`** - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - (no description)

      - **option 2** - _[string]_ - (no description)

  - **`exclude`** - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

    - **`if`** - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - (no description)

      - **option 2** - _[string]_ - (no description)

  - **`rules`** - _[array of object]_ - (no description) 

    - **`type`** - _[string]_ - (no description) _(required)_

    - **`action`** - _[string]_ - (no description) 

    - **`if`** - _[string]_ - (no description) 

    - **`names`** - _[array of string]_ - (no description) 

    - **`all`** - _[boolean]_ - (no description) 

    - **`sql`** - _[string]_ - (no description) 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 