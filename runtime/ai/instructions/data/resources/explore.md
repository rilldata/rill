---
description: Detailed instructions and examples for developing explore dashboard resources in Rill
---

# Instructions for developing an explore dashboard in Rill

## Introduction

Explore dashboards are resources that configure an interactive, drill-down dashboard for a metrics view. They are Rill's default dashboard type, designed for explorative slice-and-dice analysis of a single metrics view.

Explore dashboards are lightweight resources that sit downstream of a metrics view in the project DAG. Their reconcile logic is fast (validation only), so they can be created and modified freely without performance concerns.

### Key characteristics

- **Lightweight:** Reconciliation only validates the configuration; no data processing occurs.
- **Single metrics view:** Each explore dashboard renders exactly one metrics view.
- **Opinionated UI:** Provides a pre-built interface for time-series analysis, dimension breakdowns, and measure comparisons.
- **Usually one per metrics view:** Most projects create one explore dashboard for each metrics view.

### When to use explores vs canvases

- **Explore dashboards:** Best for ad-hoc analysis, drill-down investigations, and letting users freely slice data by any dimension.
- **Canvas dashboards:** Best for fixed reports, executive summaries, or combining multiple metrics views into a single view.

## Development approach

Explore dashboards require minimal configuration. In most cases, you only need to:

1. Reference the metrics view
2. Select which dimensions and measures to expose (usually all)
3. Optionally configure defaults and time ranges

**Best practice:** Keep explore configurations simple. Only add advanced features (security policies, custom themes, restricted dimensions) when there is a clear requirement. The metrics view already defines the business logic; the explore just controls presentation and access.

## Inline explores in metrics views

Metrics views can optionally include an `explore:` block that configures an explore dashboard directly within the metrics view file. This creates an explore resource with the same name as the metrics view.

```yaml
# metrics/sales.yaml
type: metrics_view
title: Sales Analytics
model: sales_model
timeseries: order_date

dimensions:
  - column: region
  - column: product_category

measures:
  - name: total_revenue
    expression: SUM(revenue)

# Inline explore configuration (optional)
explore:
  time_ranges:
    - P7D
    - P30D
    - P90D
  defaults:
    time_range: P30D
```

Use inline explores for simple cases where you want to keep the metrics view and its dashboard configuration together. Use separate explore files when you need multiple explores for the same metrics view, or when the explore configuration is complex.

## Properties reference

### Required properties

| Property       | Description                        |
| -------------- | ---------------------------------- |
| `type`         | Must be `explore`                  |
| `metrics_view` | Name of the metrics view to render |

### Display properties

| Property       | Description                                                 |
| -------------- | ----------------------------------------------------------- |
| `display_name` | Human-readable name shown in the UI                         |
| `description`  | Description text for the dashboard                          |
| `banner`       | Custom banner message displayed at the top of the dashboard |

### Field selection

| Property     | Description                                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------------------------- |
| `dimensions` | Dimensions to expose. Use `'*'` for all (default), a list of names, or `exclude:` to omit specific dimensions |
| `measures`   | Measures to expose. Use `'*'` for all (default), a list of names, or `exclude:` to omit specific measures     |

### Time configuration

| Property                  | Description                                                                                                                                           |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `time_ranges`             | List of time range presets available in the dropdown. Uses ISO 8601 durations (e.g., `P7D`, `P30D`) or Rill extensions (e.g., `rill-WTD`, `rill-MTD`) |
| `time_zones`              | List of IANA time zones to pin at the top of the time zone selector                                                                                   |
| `lock_time_zone`          | When `true`, locks the dashboard to the first time zone in `time_zones` (or UTC if none specified)                                                    |
| `allow_custom_time_range` | When `false`, hides the custom time range picker. Defaults to `true`                                                                                  |

### Defaults

The `defaults:` block configures the initial dashboard state when a user first loads it:

| Property                        | Description                                                  |
| ------------------------------- | ------------------------------------------------------------ |
| `defaults.time_range`           | Initial time range selection                                 |
| `defaults.dimensions`           | Initial dimensions to display                                |
| `defaults.measures`             | Initial measures to display                                  |
| `defaults.comparison_mode`      | Initial comparison mode: `none`, `time`, or `dimension`      |
| `defaults.comparison_dimension` | For `dimension` comparison mode, the dimension to compare by |

### Theming

| Property | Description                                                                                          |
| -------- | ---------------------------------------------------------------------------------------------------- |
| `theme`  | Name of a theme resource, or an inline theme definition with `colors.primary` and `colors.secondary` |

### Security

| Property          | Description                                                                                                                               |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `security.access` | Expression controlling who can access the dashboard. Uses templating with user attributes like `{{ .user.admin }}` or `{{ .user.email }}` |

### Embedding

| Property            | Description                                              |
| ------------------- | -------------------------------------------------------- |
| `embeds.hide_pivot` | When `true`, hides the pivot table view in embedded mode |

## Common time range presets

ISO 8601 durations:
- `PT6H`, `PT24H` — Hours
- `P7D`, `P14D`, `P30D`, `P90D` — Days
- `P4W` — Weeks
- `P1M`, `P3M`, `P12M` — Months

Rill extensions:
- `rill-TD` — Today
- `rill-WTD`, `rill-MTD`, `rill-QTD`, `rill-YTD` — Week/Month/Quarter/Year to date
- `rill-PDC`, `rill-PWC`, `rill-PMC`, `rill-PQC`, `rill-PYC` — Previous complete day/week/month/quarter/year
- `inf` — All time

## Example with annotations

```yaml
# dashboards/sales_explore.yaml

# Required: resource type
type: explore

# Required: the metrics view this dashboard renders
metrics_view: sales_metrics

# Optional: display name shown in the navigation and header
display_name: "Sales Performance"

# Optional: informational banner at the top of the dashboard
banner: "Data refreshes daily at 6 AM UTC"

# Optional: which dimensions to expose (use '*' for all)
dimensions: '*'

# Optional: which measures to expose (use '*' for all)
measures: '*'

# Optional: customize the time range dropdown
time_ranges:
  - P7D
  - P30D
  - P90D
  - P12M
  - rill-MTD
  - rill-YTD

# Optional: default dashboard state on first load
defaults:
  time_range: P30D
  comparison_mode: time

# Optional: pin specific time zones to the top of the selector
time_zones:
  - America/Los_Angeles
  - America/New_York

# Optional: custom theme colors
theme:
  colors:
    primary: hsl(210, 70%, 50%)
    secondary: hsl(280, 60%, 55%)

# Optional: restrict access to specific users
security:
  access: "{{ .user.admin }} OR '{{ .user.email }}' LIKE '%@example.com'"
```

## Minimal example

For most use cases, a minimal explore is sufficient:

```yaml
type: explore
metrics_view: sales_metrics
display_name: "Sales Dashboard"
dimensions: '*'
measures: '*'
```

## JSON Schema

```
  explore:
    title: Explore Dashboard YAML
    id: explore-dashboards
    type: object
    description: Explore dashboards provide an interactive way to explore data with predefined measures and dimensions.
    allOf:
      - title: Properties
        type: object
        properties:
          type:
            type: string
            const: explore
            description: Refers to the resource type and must be `explore`
          display_name:
            type: string
            description: Refers to the display name for the explore dashboard
          metrics_view:
            type: string
            description: Refers to the metrics view resource
          description:
            type: string
            description: Refers to the description of the explore dashboard
          banner:
            type: string
            description: Refers to the custom banner displayed at the header of an explore dashboard
          dimensions:
            description:  List of dimension names. Use '*' to select all dimensions (default)
            $ref: '#/definitions/field_selector_properties'
            examples:
              - # Example: Select a dimension
                dimensions:
                  - country

              - # Example: Select all dimensions except one
                dimensions:
                  exclude:
                    - country

              - # Example: Select all dimensions that match a regex
                dimensions:
                  expr: "^public_.*$"

          measures:
            description: List of measure names. Use '*' to select all measures (default)
            $ref: '#/definitions/field_selector_properties'
            examples:
              - # Example: Select a measure
                measures:
                  - sum_of_total

              - # Example: Select all measures except one
                measures:
                  exclude:
                    - sum_of_total

              - # Example: Select all measures that match a regex
                measures:
                  expr: "^public_.*$"

          theme:
            oneOf:
              - type: string
                description: Name of an existing theme to apply to the dashboard
              - $ref: '#/definitions/theme_properties'
                description: Inline theme configuration.
            description: Name of the theme to use. Only one of theme and embedded_theme can be set.
          time_ranges:
            type: array
            description: |
              Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets'
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
            items:
              $ref: '#/definitions/explore_time_range_properties'
          time_zones:
            type: array
            description: Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)
            items:
              type: string
          lock_time_zone:
            type: boolean
            description: When true, the dashboard will be locked to the first time provided in the time_zones list. When no time_zones are provided, the dashboard will be locked to UTC
          allow_custom_time_range:
            type: boolean
            description: Defaults to true, when set to false it will hide the ability to set a custom time range for the user.
          defaults:
            type: object
            description: |
              defines the defaults YAML struct
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
            properties:
              dimensions:
                description: Provides the default dimensions to load on viewing the dashboard
                $ref: '#/definitions/field_selector_properties'
              measures:
                description: Provides the default measures to load on viewing the dashboard
                $ref: '#/definitions/field_selector_properties'
              time_range:
                description: Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions)
                type: string
              comparison_mode:
                description: 'Controls how to compare current data with historical or categorical baselines. Options: `none` (no comparison), `time` (compares with past based on default_time_range), `dimension` (compares based on comparison_dimension values)'
                type: string
                enum:
                  - none
                  - time
                  - dimension
              comparison_dimension:
                description: 'for dimension mode, specify the comparison dimension by name'
                type: string
            additionalProperties: false
          embeds:
            type: object
            description: Configuration options for embedded dashboard views
            properties:
              hide_pivot:
                type: boolean
                description: When true, hides the pivot table view in embedded mode
            additionalProperties: false

          security:
            description: Security rules to apply for access to the explore dashboard
            $ref: '#/definitions/dashboard_security_policy_properties'
        required:
          - type
          - display_name
          - metrics_view
      - $ref: '#/definitions/common_properties'
```
