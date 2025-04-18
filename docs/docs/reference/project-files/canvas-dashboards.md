---
title: Canvas Dashboard YAML
sidebar_label: Canvas Dashboard YAML
sidebar_position: 30
hide_table_of_contents: true
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

**`type`** — Refers to the resource type and must be `explore` _(required)_. 

**`title`** — Refers to the display name for the dashboard [deprecated, use `display_name`] _(required)_.

**`display_name`** - Refers to the display name for the metrics view _(required)_.

**`banner`** - Refers to the custom banner displayed at the header of an Canvas dashboard  _(optional)_.

**`rows`** - Refers to all of the rows displayed on the Canvas dashboard _(required)_.

    - **`items`** - Refers to each row of components, mulitple items can be listed in a single `items` . _(required)_.
      - **`markdown`** - text component, uses markdown formatting.
      - **`kpi_grid`** - KPI component, similar to TDD in Rill Explore, display quick KPI charts.
      - **`stacked_bar_normalized`** - Bar chart normalized to 100% values.
      - **`line_chart`** - Normal Line chart
      - **`bar_chart`** - Normal Bar chart
      - **`stacked_bar`** - Stacked Bar chart
      - **`area_chart`** - Line chart with area 
      - **`image`** - provide a `url` to embed into canvas dashboard
      - **`table`** - similar to Pivot table, add dimensions and measures to visual your data
  
  :::tip 
        Each component varies slightly on what keys are required. For the most part, each component will require a `metrics_view` (except for text and image.) The charts will require a `x` and `y` valeu To build a successful component via code, take a look at what gets generated in the YAML file when select various features in the visual canvas editor.
  :::

```yaml
  - items:
      - stacked_bar:
          metrics_view: <metrics_view>
          title: ""
          description: ""
          color: hsl(240,100%,67%)
          x:
            field: <x_field>
            limit: 20
            sort: -y
            type: temporal
          y:
            field: <y_field>
            type: quantitative
            zeroBasedOrigin: true
        width: 6
      - kpi_grid:
          metrics_view: <metrics_view>
          measures:
            - <measure_1>
            - <measure_2>
            - <measure_3>
          comparison:
            - delta
            - percent_change
        width: 12
    height: 128px

```

**`max_width`**: Max width of the Canvas dashboard. Defaults to 1200 _(optional)_.

**`defaults`** - defines the defaults YAML struct

    - **`time_range`** — Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) (default). If not specified, defaults to the full time range of the `timeseries` column _(optional)_.


    - **`comparison_mode`** - comparison mode _(optional)_.
      - `none` - no comparison
      - `time` - time, will pick the comparison period depending on `default_time_range`


    **Default Example:**
    ```yaml
    defaults: #define all the defaults within here
        time_range: P1M 
        comparison_mode: time
```


**`time_ranges`** — Overrides the list of default time range selections available in the dropdown. Note that `All Time` and `Custom` selections are always available _(optional)_.
  - **`range`** — a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions) for the selection _(required)_
  - **`comparison_offsets`** — list of time comparison options for this time range selection _(optional)_. Must be one of the [Rill ISO 8601 extensions](../rill-iso-extensions.md#extensions).
  
  **Example**:
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

**`time_zones`** — Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). By adding one or more time zones will make the dashboard time zone aware and allow users to change current time zone within the dashboard _(optional)_.

**`theme`** — Refers to the default theme to apply to the dashboard. A valid theme must be defined in the project. Read this [page](./themes.md) for more detailed information about themes _(optional)_.
```yaml
theme:
  colors:
    primary: hsl(180, 100%, 50%)
    secondary: lightgreen
```

**`security`** - Defines a [security policy](/manage/security) for the dashboard _(optional)_.
  - **`access`** - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to `false` and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean _(optional)_.