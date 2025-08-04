---
note: GENERATED. DO NOT EDIT.
title: Metrics Views YAML
sidebar_position: 35
---

In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `metrics_view` 

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

### `required`

_[no type]_ - (no description) 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 