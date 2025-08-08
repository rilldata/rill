---
note: GENERATED. DO NOT EDIT.
title: Metrics View YAML
sidebar_position: 37
---

In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `metrics_view` _(required)_

### `parent`

_[string]_ - Refers to the parent metrics from which this metrics view is derived. If specified, this will inherit properties from the parent metrics view 

### `display_name`

_[string]_ - Refers to the display name for the metrics view 

### `description`

_[string]_ - Refers to the description for the metrics view 

### `ai_instructions`

_[string]_ - Extra instructions for AI agents. Used to guide natural language question answering and routing. 

### `model`

_[string]_ - Refers to the model powering the dashboard (either model or table is required) 

### `database`

_[string]_ - Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified 

### `database_schema`

_[string]_ - Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified 

### `table`

_[string]_ - Refers to the table powering the dashboard, should be used instead of model for dashboards create from external OLAP tables (either table or model is required) 

### `timeseries`

_[string]_ - Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear 

### `watermark`

_[string]_ - A SQL expression that tells us the max timestamp that the metrics are considered valid for. Usually does not need to be overwritten 

### `smallest_time_grain`

_[string]_ - Refers to the smallest time granularity the user is allowed to view. The valid values are: millisecond, second, minute, hour, day, week, month, quarter, year 

### `first_day_of_week`

_[integer]_ - Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=1 and Sunday=7 

### `first_month_of_year`

_[integer]_ - Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=1 and December=12 

### `dimensions`

_[array of object]_ - Relates to exploring segments or dimensions of your data and filtering the dashboard 

  - **`name`** - _[string]_ - a stable identifier for the dimension 

  - **`display_name`** - _[string]_ - a display name for your dimension 

  - **`description`** - _[string]_ - a freeform text description of the dimension 

  - **`column`** - _[string]_ - a categorical column 

  - **`expression`** - _[string]_ - a non-aggregate expression such as string_split(domain, '.'). One of column and expression is required but cannot have both at the same time 

  - **`unnest`** - _[boolean]_ - if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match 

  - **`uri`** - _[string, boolean]_ - enable if your dimension is a clickable URL to enable single click navigation (boolean or valid SQL expression) 

### `measures`

_[array of object]_ - Used to define the numeric aggregates of columns from your data model 

  - **`name`** - _[string]_ - a stable identifier for the measure 

  - **`display_name`** - _[string]_ - the display name of your measure. 

  - **`description`** - _[string]_ - a freeform text description of the dimension 

  - **`type`** - _[string]_ - Measure calculation type: "simple" for basic aggregations, "derived" for calculations using other measures, or "time_comparison" for period-over-period analysis. Defaults to "simple" unless dependencies exist. 

  - **`expression`** - _[string]_ - a combination of operators and functions for aggregations 

  - **`window`** - _[anyOf]_ - A measure window can be defined as a keyword string (e.g. 'time' or 'all') or an object with detailed window configuration. 

    - **option 1** - _[string]_ - Shorthand: `time` or `true` means time-partitioned, `all` means non-partitioned.

    - **option 2** - _[object]_ - Detailed window configuration for measure calculations, allowing control over partitioning, ordering, and frame definition.

      - **`partition`** - _[boolean]_ - Controls whether the window is partitioned. When true, calculations are performed within each partition separately. 

      - **`order`** - _[anyOf]_ - Specifies the fields to order the window by, determining the sequence of rows within each partition. 

        - **option 1** - _[string]_ - Simple field name as a string.

        - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

          - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

          - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

            - **`name`** - _[string]_ - Name of the field to select. _(required)_

            - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

      - **`frame`** - _[string]_ - Defines the window frame boundaries for calculations, specifying which rows are included in the window relative to the current row. 

  - **`per`** - _[anyOf]_ - for per dimensions 

    - **option 1** - _[string]_ - Simple field name as a string.

    - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

      - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

      - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

        - **`name`** - _[string]_ - Name of the field to select. _(required)_

        - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

  - **`requires`** - _[anyOf]_ - using an available measure or dimension in your metrics view to set a required parameter, cannot be used with simple measures 

    - **option 1** - _[string]_ - Simple field name as a string.

    - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

      - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

      - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

        - **`name`** - _[string]_ - Name of the field to select. _(required)_

        - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

  - **`format_preset`** - _[string]_ - Controls the formatting of this measure using a predefined preset. Measures cannot have both `format_preset` and `format_d3`. If neither is supplied, the measure will be formatted using the `humanize` preset by default.
  
    Available options:
    - `humanize`: Round numbers into thousands (K), millions(M), billions (B), etc.
    - `none`: Raw output.
    - `currency_usd`: Round to 2 decimal points with a dollar sign ($).
    - `currency_eur`: Round to 2 decimal points with a euro sign (€).
    - `percentage`: Convert a rate into a percentage with a % sign.
    - `interval_ms`: Convert milliseconds into human-readable durations like hours (h), days (d), years (y), etc. (optional)
 

  - **`format_d3`** - _[string]_ - Controls the formatting of this measure using a [d3-format](https://d3js.org/d3-format) string. If an invalid format string is supplied, the measure will fall back to `format_preset: humanize`. A measure cannot have both `format_preset` and `format_d3`. If neither is provided, the humanize preset is used by default. Example: `format_d3: ".2f"` formats using fixed-point notation with two decimal places. Example: `format_d3: ",.2r"` formats using grouped thousands with two significant digits. (optional) 

  - **`format_d3_locale`** - _[object]_ - locale configuration passed through to D3, enabling changing the currency symbol among other things. For details, see the docs for D3's [formatLocale](https://d3js.org/d3-format#formatLocale) 

  - **`valid_percent_of_total`** - _[boolean]_ - a boolean indicating whether percent-of-total values should be rendered for this measure 

  - **`treat_nulls_as`** - _[string]_ - used to configure what value to fill in for missing time buckets. This also works generally as COALESCING over non empty time buckets. 

### `parent_dimensions`

_[oneOf]_ - Optional field selectors for dimensions to inherit from the parent metrics view. 

  - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

  - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select fields using a regular expression 

    - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

    - **`exclude`** - _[object]_ - Select all fields except those listed here 

### `parent_measures`

_[oneOf]_ - Optional field selectors for measures to inherit from the parent metrics view. 

  - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

  - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

  - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`regex`** - _[string]_ - Select fields using a regular expression 

    - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

    - **`exclude`** - _[object]_ - Select all fields except those listed here 

### `annotations`

_[array of object]_ - Used to define annotations that can be displayed on charts 

  - **`name`** - _[string]_ - A stable identifier for the annotation. Defaults to model or table names when not specified 

  - **`model`** - _[string]_ - Refers to the model powering the annotation (either table or model is required). The model must have 'time' and 'description' columns. Optional columns include 'time_end' for range annotations and 'duration' to specify when the annotation should appear based on dashboard grain level. 

  - **`database`** - _[string]_ - Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified 

  - **`database_schema`** - _[string]_ - Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified 

  - **`table`** - _[string]_ - Refers to the table powering the annotation, should be used instead of model for annotations from external OLAP tables (either table or model is required) 

  - **`connector`** - _[string]_ - Refers to the connector to use for the annotation 

  - **`measures`** - _[anyOf]_ - Specifies which measures to apply the annotation to. Applies to all measures if not specified 

    - **option 1** - _[string]_ - Simple field name as a string.

    - **option 2** - _[array of anyOf]_ - List of field selectors, each can be a string or an object with detailed configuration.

      - **option 1** - _[string]_ - Shorthand field selector, interpreted as the name.

      - **option 2** - _[object]_ - Detailed field selector configuration with name and optional time grain.

        - **`name`** - _[string]_ - Name of the field to select. _(required)_

        - **`time_grain`** - _[string]_ - Time grain for time-based dimensions. 

### `security`

_[object]_ - Defines security rules and access control policies for resources 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    - **option 1** - _[string]_ - SQL expression that evaluates to a boolean to determine access

    - **option 2** - _[boolean]_ - Direct boolean value to allow or deny access

  - **`row_filter`** - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`include`** - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

    - **`if`** - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - List of specific field names to include

      - **option 2** - _[string]_ - Wildcard '*' to include all fields

  - **`exclude`** - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

    - **`if`** - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - List of specific field names to exclude

      - **option 2** - _[string]_ - Wildcard '*' to exclude all fields

  - **`rules`** - _[array of object]_ - List of detailed security rules that can be used to define complex access control policies 

    - **`type`** - _[string]_ - Type of security rule - access (overall access), field_access (field-level access), or row_filter (row-level filtering) _(required)_

    - **`action`** - _[string]_ - Whether to allow or deny access for this rule 

    - **`if`** - _[string]_ - Conditional expression that determines when this rule applies. Must be a valid SQL expression that evaluates to a boolean 

    - **`names`** - _[array of string]_ - List of field names this rule applies to (for field_access type rules) 

    - **`all`** - _[boolean]_ - When true, applies the rule to all fields (for field_access type rules) 

    - **`sql`** - _[string]_ - SQL expression for row filtering (for row_filter type rules) 

### `explore`

_[object]_ - Defines an optional inline explore view for the metrics view. If not specified a default explore will be emitted unless `skip` is set to true. 

  - **`skip`** - _[boolean]_ - If true, disables the explore view for this metrics view. 

  - **`name`** - _[string]_ - Name of the explore view. 

  - **`display_name`** - _[string]_ - Display name for the explore view. 

  - **`description`** - _[string]_ - Description for the explore view. 

  - **`banner`** - _[string]_ - Custom banner displayed at the header of the explore view. 

  - **`theme`** - _[oneOf]_ - Name of the theme to use or define a theme inline. Either theme name or inline theme can be set. 

    - **option 1** - _[string]_ - Name of an existing theme to apply to the explore view.

    - **option 2** - _[object]_ - Inline theme configuration.

      - **`colors`** - _[object]_ - Used to override the dashboard colors. Either primary or secondary color must be provided. 

        - **`primary`** - _[string]_ - Overrides the primary blue color in the dashboard. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. Note that the hue of the input colors is used for variants but the saturation and lightness is copied over from the [blue color palette](https://tailwindcss.com/docs/customizing-colors). 

        - **`secondary`** - _[string]_ - Overrides the secondary color in the dashboard. Applies to the loading spinner only as of now. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

  - **`time_ranges`** - _[array of oneOf]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets'. 

    - **option 1** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

    - **option 2** - _[object]_ - Object containing time range and comparison configuration

      - **`range`** - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection _(required)_

      - **`comparison_offsets`** - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

        - **option 1** - _[string]_ - Offset string only (range is inferred)

        - **option 2** - _[object]_ - Object containing offset and range configuration for time comparison

          - **`offset`** - _[string]_ - Time offset for comparison (e.g., 'P1D' for one day ago) 

          - **`range`** - _[string]_ - Custom time range for comparison period 

  - **`time_zones`** - _[array of string]_ - List of time zones to pin to the top of the time zone selector. Should be a list of IANA time zone identifiers. 

  - **`lock_time_zone`** - _[boolean]_ - When true, the explore view will be locked to the first time zone provided in the time_zones list. If no time_zones are provided, it will be locked to UTC. 

  - **`allow_custom_time_range`** - _[boolean]_ - Defaults to true. When set to false, hides the ability to set a custom time range for the user. 

  - **`defaults`** - _[object]_ - Preset UI state to show by default. 

    - **`dimensions`** - _[oneOf]_ - Default dimensions to load on viewing the explore view. 

      - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

      - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

      - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

        - **`regex`** - _[string]_ - Select fields using a regular expression 

        - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

        - **`exclude`** - _[object]_ - Select all fields except those listed here 

    - **`measures`** - _[oneOf]_ - Default measures to load on viewing the explore view. 

      - **option 1** - _[string]_ - Wildcard(*) selector that includes all available fields in the selection

      - **option 2** - _[array of string]_ - Explicit list of fields to include in the selection

      - **option 3** - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

        - **`regex`** - _[string]_ - Select fields using a regular expression 

        - **`expr`** - _[string]_ - DuckDB SQL expression to select fields based on custom logic 

        - **`exclude`** - _[object]_ - Select all fields except those listed here 

    - **`time_range`** - _[string]_ - Default time range to display when the explore view loads. 

    - **`comparison_mode`** - _[string]_ - Default comparison mode for metrics (none, time, or dimension). 

    - **`comparison_dimension`** - _[string]_ - Default dimension to use for comparison when comparison_mode is 'dimension'. 

  - **`embeds`** - _[object]_ - Configuration options for embedded explore views. 

    - **`hide_pivot`** - _[boolean]_ - When true, hides the pivot table view in embedded mode. 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 