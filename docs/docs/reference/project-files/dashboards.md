---
title: Dashboard YAML
sidebar_label: Dashboard YAML
sidebar_position: 30
---

In your Rill project directory, create a `<dashboard_name>.yaml` file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

_**`model`**_ — the model name powering the dashboard with no path _(required)_

_**`title`**_ — the display name for the dashboard _(required)_

_**`timeseries`**_ — the timestamp column from your model that will underlie x-axis data in the line charts _(optional)_. If not specified, the line charts will not appear.

_**`default_time_range`**_ — the default time range shown when a user initially loads the dashboard _(optional)_. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) (default). If not specified, defaults to the full time range of the `timeseries` column.

_**`smallest_time_grain`**_ — the smallest time granularity the user is allowed to view in the dashboard _(optional)_. The valid values are: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`.

_**`first_day_of_week`**_ — the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=1 and Sunday=7 _(optional)_

_**`first_month_of_year`**_ — the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=1 and December=12 _(optional)_

_**`available_time_zones`**_ — time zones that should be pinned to the top of the time zone selector _(optional)_. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). By adding one or more time zones will make the dashboard time zone aware and allow users to change current time zone within the dashboard.

_**`default_theme`**_ — default theme to apply to the dashboard _(optional)_. A valid theme must be defined in the project. Read this [page](./themes.md) for more detailed information about themes.

_**`default_comparison`**_ - defines which should be the default comparison mode. Default: none _(optional)_
  - _**`mode`**_ - comparison mode
    - _`none`_ - no comparison
    - _`time`_ - time, will pick the comparison period depending on `default_time_range`
    - _`dimension`_ - dimension comparison mode
  - _**`dimension`**_ - for dimension mode, specify the comparison dimension by name

_**`dimensions`**_ — for exploring [segments](../../develop/metrics-dashboard#dimensions) and filtering the dashboard _(required)_
  - _**`column`**_ — a categorical column _(required)_ 
  - _**`expression`**_ a non-aggregate expression such as `string_split(domain, '.')`. One of `column` and `expression` is required but cannot have both at the same time _(required)_
  - _**`name`**_ — a stable identifier for the dimension _(optional)_
  - _**`label`**_ — a label for your dashboard dimension _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_
  - _**`unnest`**_ - if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match _(optional)_
  - _**`ignore`**_ — hides the dimension _(optional)_ 

_**`measures`**_ — numeric [aggregates](../../develop/metrics-dashboard#measures) of columns from your data model  _(required)_
  - _**`expression`**_ — a combination of operators and functions for aggregations _(required)_ 
  - _**`name`**_ — a stable identifier for the measure _(required)_
  - _**`label`**_ — a label for your dashboard measure _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 
  - _**`ignore`**_ — hides the measure _(optional)_ 
  - _**`valid_percent_of_total`**_ — a boolean indicating whether percent-of-total values should be rendered for this measure _(optional)_ 
  - _**`format_d3`**_ — controls the formatting of this measure in the dashboard using a [d3-format string](https://d3js.org/d3-format). If an invalid format string is supplied, measures will be formatted with `format_preset: humanize` (described below). Measures cannot have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - **Example**: to show a measure using fixed point formatting with 2 digits after the decimal point, your measure specification would include: `format_d3: ".2f"`.
    - **Example**: to show a measure using grouped thousands with two significant digits, your measure specification would include: `format_d3: ",.2r"`.
  - _**`format_preset`**_ — controls the formatting of this measure in the dashboard according to option specified below. Measures cannot have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - _`humanize`_ — round off numbers in an opinionated way to thousands (K), millions (M), billions (B), etc
    - _`none`_ — raw output
    - _`currency_usd`_ —  output rounded to 2 decimal points prepended with a dollar sign: `$`
    - _`currency_eur`_ —  output rounded to 2 decimal points prepended with a euro symbol: `€`
    - _`percentage`_ — output transformed from a rate to a percentage appended with a percentage sign
    - _`interval_ms`_ — time intervals given in milliseconds are transformed into human readable time units like hours (h), days (d), years (y), etc

_**`available_time_ranges`**_ — Override the list of default time range selections available in the dropdown _(optional)_. Note that `All Time` and `Custom` selections are always available.
  - _**`range`**_ — a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) for the selection _(required)_
  - _**`comparison_offsets`**_ — list of time comparison options for this time range selection _(optional)_. Must be one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions).
  - **Example**:
    ```yaml
    available_time_ranges:
    - PT15M // Simplified syntax to specify only the range
    - PT1H
    - PT6H
    - P7D
    - range: P5D // Advanced syntax to specify comparison_offsets as well
      comparison_offsets:
        - rill-PP
        - rill-PW
    - P4W
    - rill-TD // Today
    - rill-WTD // Week-To-date
    ```

_**`security`**_ - define a [security policy](../../develop/security) for the dashboard _(optional)_
  - _**`access`**_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to `false` and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. _(optional)_
  - _**`row_filter`**_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a `WHERE` clause. _(optional)_
  - _**`exclude`**_ - List of dimension or measure names to exclude from the dashboard. If `exclude` is defined all other dimensions and measures are included. _(optional)_
    - **`if`** - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean. _(required)_
    - **`names`** - List of fields to exclude. Should match the `name` of one of the dashboard's dimensions or measures. _(required)_
  - _**`include`**_ - List of dimension or measure names to include in the dashboard. If `include` is defined all other dimensions and measures are excluded. _(optional)_
    - **`if`** - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean. _(required)_
    - **`names`** - List of fields to include. Should match the `name` of one of the dashboard's dimensions or measures. _(required)_
