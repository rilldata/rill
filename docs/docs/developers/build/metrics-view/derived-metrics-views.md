---
title: Derived Metrics Views
description: Create derived metrics views that inherit dimensions, measures, and properties from a parent metrics view
sidebar_label: Derived Metrics Views
sidebar_position: 06
---

A derived metrics view inherits its dimensions, measures, and data source configuration from an existing parent metrics view. It can only expose a subset of the parent's dimensions and measures — it cannot define new ones. This lets you create focused views for specific teams or use cases without duplicating definitions. The parent metrics view should contain the superset of all dimensions and measures that any derived view might need.

## Basic Example

Given a parent metrics view `ad_bids.yaml`:

```yaml
version: 1
type: metrics_view
model: ad_bids_model
timeseries: timestamp
smallest_time_grain: hour

dimensions:
  - column: domain
  - column: city
  - column: country
  - column: device

measures:
  - name: total_bids
    expression: count(*)
  - name: total_revenue
    expression: sum(bid_price)
  - name: avg_bid
    expression: avg(bid_price)
```

A derived metrics view `ad_bids_summary.yaml` that inherits everything:

```yaml
type: metrics_view
parent: ad_bids
```

This creates a metrics view with the same dimensions, measures, timeseries, and data source as `ad_bids`. An explore dashboard is automatically emitted for it.

## Selecting Dimensions and Measures

Use `parent_dimensions` and `parent_measures` to control which fields are inherited. When omitted, all fields are inherited (equivalent to `'*'`).

### Select All (Wildcard)

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions: '*'
parent_measures: '*'
```

### Select by Name

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions:
  - domain
  - country

parent_measures:
  - total_bids
  - total_revenue
```

### Exclude Specific Fields

Use `exclude` to inherit everything except specific fields:

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions:
  exclude:
    - city

parent_measures:
  exclude:
    - avg_bid
```

### Select by Regex

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions:
  regex: "^(domain|country)$"

parent_measures:
  regex: "^total_.*"
```

### Select by DuckDB Expression

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions:
  expr: "* EXCLUDE (city)"
```

## Overriding Inherited Properties

A derived metrics view can override certain properties from the parent. Properties you set on the child take precedence; properties you omit are inherited.

### Overridable Properties

| Property | Behavior |
|---|---|
| `display_name` | Child's value if set, otherwise defaults to the resource name (not inherited from parent) |
| `description` | Child's value if set, otherwise empty (not inherited from parent) |
| `timeseries` | Child's value if set, otherwise parent's |
| `smallest_time_grain` | Child's value if set, otherwise parent's. Must be >= parent's grain |
| `first_day_of_week` | Child's value if set, otherwise parent's |
| `first_month_of_year` | Child's value if set, otherwise parent's |
| `watermark` | Child's value if set, otherwise parent's |
| `ai_instructions` | Child's value if set, otherwise parent's |

### Always Inherited (Cannot Override)

These properties are always taken from the parent:
- `model` / `table` (data source)
- `connector`, `database`, `database_schema`
- Cache settings (`cache.enabled`, `cache.key_sql`, `cache.key_ttl`)

### Example: Override Time Grain

The parent uses hourly data, but this derived view restricts to daily granularity:

```yaml
type: metrics_view
parent: ad_bids
display_name: Ad Bids Daily Summary

smallest_time_grain: day

parent_dimensions:
  - domain
  - country

parent_measures: '*'
```

:::info
When overriding `smallest_time_grain`, the value must be equal to or coarser than the parent's grain. For example, if the parent uses `hour`, the child can use `day` but not `minute`.
:::

## Dimensions and Measures

A derived metrics view can only inherit dimensions and measures from its parent — it cannot define its own `dimensions` or `measures` directly. The parent metrics view must contain the superset of all dimensions and measures that any derived view might need. Use `parent_dimensions` and `parent_measures` to select which subset to expose.

:::caution
Setting `dimensions` or `measures` on a derived metrics view will produce a validation error. All dimensions and measures must be defined on the parent.
:::

## Security Rules

All security rules from the parent are inherited by the derived metrics view. The only exception is the `access` rule: if the derived view defines its own `access` rule, the parent's `access` rule is skipped. All other parent rules (`row_filter`, `field_access`, `include`, `exclude`) are always appended to the child's rules regardless.

This means a derived view can narrow who has access (by overriding `access`), and can add additional row filters or field restrictions on top of whatever the parent already enforces.

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions: '*'
parent_measures: '*'

security:
  access: "'{{ .user.domain }}' = 'partner.com'"
  row_filter: "country = 'US'"
```

In this example, the child overrides the parent's `access` rule with its own, but the parent's `row_filter`, `include`, `exclude`, and `field_access` rules are still inherited and appended to the child's rules. The child also adds its own `row_filter` on top of the parent's.

## Inline Explore Configuration

By default, a derived metrics view automatically emits an explore dashboard. You can customize it using the `explore` key, or disable it entirely.

:::note
Auto-emission of the explore dashboard only happens when the YAML file does not set `version` (or sets `version: 0`). If `version: 1` is set, no explore is emitted automatically — you must either define one inline with the `explore` key or create a separate explore file.
:::

### Default Behavior

When no `explore` key is specified and `version` is not set, an explore dashboard is created with all inherited dimensions and measures.

### Custom Explore

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions: '*'
parent_measures: '*'

explore:
  display_name: Ad Bids Partner View
  time_ranges:
    - P7D
    - P30D
    - range: P90D
      comparison_offsets:
        - P90D
  defaults:
    dimensions:
      - domain
    measures:
      - total_bids
      - total_revenue
    time_range: P7D
    comparison_mode: time
```

### Disable Explore

```yaml
type: metrics_view
parent: ad_bids

parent_dimensions: '*'
parent_measures: '*'

explore:
  skip: true
```

### Explore Options

The inline `explore` key supports the same configuration as a standalone [explore dashboard](/reference/project-files/explore-dashboards):

| Property | Description |
|---|---|
| `skip` | Set to `true` to disable the explore dashboard |
| `name` | Custom name for the explore resource |
| `display_name` | Display name shown in the UI |
| `description` | Description for the explore |
| `banner` | Custom banner at the header of the explore |
| `theme` | Theme name or inline theme object |
| `time_ranges` | Available time range selections |
| `time_zones` | Pinned time zones (IANA identifiers) |
| `lock_time_zone` | Lock to the first time zone in `time_zones` |
| `allow_custom_time_range` | Allow custom time range selection (default: `true`) |
| `defaults` | Default UI state: `dimensions`, `measures`, `time_range`, `comparison_mode`, `comparison_dimension` |
| `embeds` | Embed configuration, e.g. `hide_pivot: true` |

## Complete Example

A parent metrics view and two derived views for different teams:

**`metrics/ad_bids.yaml`** — the shared parent:

```yaml
version: 1
type: metrics_view
model: ad_bids_model
timeseries: timestamp
smallest_time_grain: hour

dimensions:
  - column: domain
  - column: city
  - column: country
  - column: device
  - column: publisher

measures:
  - name: total_bids
    expression: count(*)
  - name: total_revenue
    expression: sum(bid_price)
  - name: avg_bid
    expression: avg(bid_price)
  - name: unique_domains
    expression: count(distinct domain)

security:
  access: true
  row_filter: "bid_price > 0"
```

**`metrics/ad_bids_sales.yaml`** — for the sales team:

```yaml
type: metrics_view
parent: ad_bids
display_name: Ad Bids - Sales

parent_dimensions:
  - domain
  - country

parent_measures:
  - total_revenue
  - avg_bid

smallest_time_grain: day

explore:
  display_name: Sales Dashboard
  defaults:
    time_range: P30D
    measures:
      - total_revenue
```

**`metrics/ad_bids_engineering.yaml`** — for the engineering team:

```yaml
type: metrics_view
parent: ad_bids
display_name: Ad Bids - Engineering

parent_dimensions: '*'

parent_measures:
  exclude:
    - avg_bid

security:
  access: "'{{ .user.department }}' = 'engineering'"

explore:
  defaults:
    time_range: P7D
    comparison_mode: time
```

## Validation Rules

The parser enforces these constraints on derived metrics views:

- `parent` must reference an existing metrics view with a valid state.
- `dimensions` and `measures` cannot be defined on a derived view — use `parent_dimensions` and `parent_measures` to select from the parent instead.
- `model`, `table`, `database`, `database_schema`, and cache settings cannot be set on a derived view (they come from the parent).
- `parent_dimensions` and `parent_measures` can only be used when `parent` is set.
- `smallest_time_grain`, if specified, must be coarser than or equal to the parent's grain.
- Deprecated top-level explore fields (`default_time_range`, `available_time_zones`, etc.) cannot be used; use the `explore` key instead.
