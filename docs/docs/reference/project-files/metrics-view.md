---
title: Metrics View YAML
sidebar_label: Metrics View YAML
sidebar_position: 28
hide_table_of_contents: true
---

In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.


## Properties

**`version`** - Refers to the version of the metrics view _(required)_. 

**`type`** — Refers to the resource type and must be `metrics_view` _(required)_. 

**`title`** — Refers to the display name for the metrics view [deprecated, use `display_name`] _(required)_.

**`display_name`** - Refers to the display name for the metrics view _(required)_.

**`description`** - A description for the project. _(optional)_.

**`database`** - Refers to the database to use in the OLAP engine (to be used in conjunction with `table`). Otherwise, will use the default database or schema if not specified _(optional)_.

**`database_schema`** — Refers to the schema to use in the OLAP engine (to be used in conjunction with `table`). Otherwise, will use the default database or schema if not specified _(optional)_.

**`watermark`** - A SQL expression that tells us the max timestamp that the metrics are considered valid for. Usually does not need to be overwritten, _(optional)_.

**`timeseries`** — Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear _(optional)_.

**`connector`** — Refers to the OLAP engine, if you are not using DuckDB, IE: [ClickHouse OLAP engine](../olap-engines/multiple-olap.md). _(optional)_.

**`model`** — Refers to the **model** powering the dashboard with no path specified; should only be used for [Rill models](/build/models/models.md) _(either **model** or **table** is required)_.

**`table`** — Refers to the **table** powering the dashboard with no path specified; should be used instead of `model` for dashboards create from [external OLAP tables](../../concepts/OLAP.md#external-olap-tables) _(either **table** or **model** is required)_. 


**`dimensions`** — Relates to exploring segments or [dimensions](/build/metrics-view/metrics-view.md#dimensions) of your data and filtering the dashboard _(required)_.
  - **`column`** — a categorical column _(required)_ 
  - **`expression`** a non-aggregate expression such as `string_split(domain, '.')`. One of `column` and `expression` is required but cannot have both at the same time _(required)_
  - **`name`** — a stable identifier for the dimension _(optional)_
  - **`label`** — a label for your dimension _(optional)_ 
  - **`description`** — a freeform text description of the dimension  _(optional)_
  - **`unnest`** - if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match _(optional)_

**`measures`** — Used to define the numeric [aggregates](/build/metrics-view/metrics-view.md#measures) of columns from your data model  _(required)_.
  - **`expression`** — a combination of operators and functions for aggregations _(required)_ 
  - **`name`** — a stable identifier for the measure _(required)_
  - **`display_name`** - the display name of your measure._(required)_
  - **`label`** — a label for your measure, deprecated use `display_name` _(optional)_ 
  - **`description`** — a freeform text description of the dimension  _(optional)_ 
  - **`valid_percent_of_total`** — a boolean indicating whether percent-of-total values should be rendered for this measure _(optional)_ 
  - **`format_d3`** — controls the formatting of this measure  using a [d3-format string](https://d3js.org/d3-format). If an invalid format string is supplied, measures will be formatted with `format_preset: humanize` (described below). Measures <u>cannot</u> have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - **Example**: to show a measure using fixed point formatting with 2 digits after the decimal point, your measure specification would include: `format_d3: ".2f"`.
    - **Example**: to show a measure using grouped thousands with two significant digits, your measure specification would include: `format_d3: ",.2r"`.
  - **`format_d3_locale`** — locale configuration passed through to D3, enabling changing the currency symbol among other things. For details, see the docs for D3's [`formatLocale`](https://d3js.org/d3-format#formatLocale). _(optional)_
  - **`format_preset`** — controls the formatting of this measure according to option specified below. Measures <u>cannot</u> have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_
    - `humanize` — round off numbers in an opinionated way to thousands (K), millions (M), billions (B), etc.
    - `none` — raw output
    - `currency_usd` —  output rounded to 2 decimal points prepended with a dollar sign: `$`
    - `currency_eur` —  output rounded to 2 decimal points prepended with a euro symbol: `€`
    - `percentage` — output transformed from a rate to a percentage appended with a percentage sign
    - `interval_ms` — time intervals given in milliseconds are transformed into human readable time units like hours (h), days (d), years (y), etc.
  - **`window`** — can be used for [advanced window expressions](/build/metrics-view/expressions), cannot be used with simple measures _(optional)_ 
    - **`partition`** — boolean _(optional)_ 
    - **`order`** — using a value available in your metrics view to order the window _(optional)_ 
    - **`ordertime`** — boolean, sets the order only by the time dimensions _(optional)_ 
    - **`frame`** — sets the frame of your window. _(optional)_ 
  - **`requires`** — using an available measure or dimension in your metrics view to set a required parameter, cannot be used with simple measures  _(optional)_
 :::note window limitations
Rill supports window function, but only when applied post-aggregation. This means that window functions can only operate on data that has already been grouped and aggregated by the defined dimensions (dims).
 :::
```yaml
measures:
 - name: bids_1day_rolling_avg
    expression: AVG(measure)
    requires: [measure]
    window:
      order: timestamp
      frame: RANGE BETWEEN INTERVAL 1 DAY PRECEDING AND CURRENT ROW
```

**`smallest_time_grain`** — Refers to the smallest time granularity the user is allowed to view. The valid values are: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year` _(optional)_.

**`first_day_of_week`** — Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=`1` and Sunday=`7` _(optional)_.

**`first_month_of_year`** — Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=`1` and December=`12` _(optional)_.

**`security`** - Defines a [security policy](/manage/security) for the dashboard _(optional)_.
  - **`access`** - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to `false` and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean _(optional)_.
  - **`row_filter`** - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a `WHERE` clause _(optional)_.
  - **`exclude`** - List of dimension or measure names to exclude from the dashboard. If `exclude` is defined all other dimensions and measures are included _(optional)_.
    - **`if`** - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_.
    - **`names`** - List of fields to exclude. Should match the `name` of one of the dashboard's dimensions or measures _(required)_.
  - **`include`** - List of dimension or measure names to include in the dashboard. If `include` is defined all other dimensions and measures are excluded _(optional)_.
    - **`if`** - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_.
    - **`names`** - List of fields to include. Should match the `name` of one of the dashboard's dimensions or measures _(required)_.
