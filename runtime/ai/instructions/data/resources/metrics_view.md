---
description: Detailed instructions and examples for developing metrics view resources in Rill
---

# Instructions for developing a metrics view in Rill

## Introduction

Metrics views are resources that define queryable business metrics on top of a table in an OLAP database. They implement what other business intelligence tools call a "semantic layer" or "metrics layer".

Metrics views are lightweight resources that only perform validation when reconciled. They are typically found downstream of connectors and models in the project's DAG. They power many user-facing features:

- **Explore dashboards**: Interactive drill-down interfaces for data exploration
- **Canvas dashboards**: Custom chart and table components
- **Alerts**: Notifications when data meets certain criteria
- **Reports**: Scheduled data exports and summaries
- **Custom APIs**: Programmatic access to metrics

## Core Concepts

### Table source

The `model:` property specifies the underlying table that powers the metrics view. It can reference:

1. **A model in the project**: Just use the model name (e.g., `model: events`)
2. **An external table**: Specify the table name as it exists in the OLAP connector

```yaml
# Referencing a model in the project
model: events

# Referencing an external table (connector defaults to project's default OLAP)
connector: clickhouse  # Optional: specify if different from default
model: my_external_table # Note: Doesn't support dot syntax for database/schema name. Use the separate `database:` or `database_schema:` keys for that if relevant (but try without first and see if that works).
```

**Note**: The `table:` property is a legacy alias for referencing external tables. Always prefer `model:` in new metrics views.

### Timeseries

The `timeseries:` property identifies the timestamp column used for time-based filtering and line charts. This column must be a time/timestamp type in the underlying table.

```yaml
timeseries: event_time
```

If the timeseries column is not listed in `dimensions:`, Rill automatically adds it as a time dimension. You can optionally configure additional time-related settings:

```yaml
timeseries: event_time
smallest_time_grain: hour      # Minimum granularity users can select
first_day_of_week: 7           # Sunday (1=Monday, 7=Sunday)
first_month_of_year: 4         # April (fiscal year starting in April)
```

It is _strongly_ recommended that you add a primary timeseries to every metrics view you create (it makes for a much better dashboard experience).

### Dimensions

Dimensions are attributes you can group by or filter on. They are typically categorical (strings, enums) or temporal (dates, timestamps). Rill infers the dimension type from the underlying SQL data type:

- **Categorical**: String, enum, boolean columns
- **Time**: Timestamp, date, datetime columns
- **Geospatial**: Geometry or geography columns

Define dimensions using either a direct column reference or a SQL expression:

```yaml
dimensions:
  # Simple column reference
  - name: country
    column: country

  # Computed expression
  - name: device_category
    expression: CASE WHEN device_type IN ('phone', 'tablet') THEN 'Mobile' ELSE 'Desktop' END

  # With display name and description
  - name: campaign_name
    display_name: Campaign
    description: Marketing campaign that drove the traffic
    column: campaign_name
```

**Naming**: Each dimension needs a `name` (stable identifier used in APIs and references), which defaults to `column:` if provided. The `display_name:` is optional, and defaults to a humanized version of `name` if not specified.

### Measures

Measures are aggregation expressions that compute numeric values when grouped by dimensions. They must use aggregate functions like `SUM()`, `COUNT()`, `AVG()`, `MIN()`, `MAX()`.

```yaml
measures:
  - name: total_revenue
    display_name: Total Revenue
    expression: SUM(revenue)
    description: Sum of all revenue in USD
    format_preset: currency_usd

  - name: unique_users
    display_name: Unique Users
    expression: COUNT(DISTINCT user_id)
    format_preset: humanize

  - name: conversion_rate
    display_name: Conversion Rate
    expression: SUM(conversions) / NULLIF(SUM(visits), 0)
    format_preset: percentage
    valid_percent_of_total: false  # Disable % of total for ratios
```

**Format presets**: Control how values are displayed:
- `none`: Raw number
- `humanize`: Round to K, M, B (e.g., 1.2M)
- `currency_usd`: Dollar format with 2 decimals ($1,234.56)
- `currency_eur`: Euro format
- `percentage`: Multiply by 100 and add % sign
- `interval_ms`: Convert milliseconds to human-readable duration

For custom formatting, use `format_d3` with a [d3-format](https://d3js.org/d3-format) string:

```yaml
format_d3: "$,.2f"  # $1,234.56
format_d3: ".1%"    # 12.3%
format_d3: ",.0f"   # 1,235 (rounded, with thousands separator)
```

### Best practices for dimensions and measures

**Naming conventions:**
- Use `snake_case` for the `name` field (e.g., `total_revenue`, `unique_users`)
- Only add `display_name` and `description` if they provide meaningful context beyond what `name` conveys (display names auto-humanize from the name by default)
- Ensure measure names don't collide with column names in the underlying table

**Getting started with measures:**
- Start with a `COUNT(*)` measure as a baseline (e.g., `total_records` or `total_events`)
- Add `SUM()` measures for numeric columns that represent quantities or values
- Use `humanize` as the default format preset unless the data has a specific format requirement
- Keep initial measures simple using only `COUNT`, `SUM`, `AVG`, `MIN`, `MAX` aggregations
- Add more complex expressions (ratios, conditional aggregations) only when needed

**Dimension selection:**
- Include all categorical columns (strings, enums, booleans) that users might want to filter or group by
- Start with 5-10 dimensions; add more based on user needs

**Timeseries:**
- If there is any date/timestamp column in the underlying table, pick the primary or most interesting one and add it under `dimensions:`
- It is also _strongly_ recommended that you configure a primary time dimension using `timeseries:`

### Auto-generated explore

When you create a metrics view, Rill automatically generates an explore dashboard with the same name, exposing all dimensions and measures. To customize the explore (you usually should not need to), add an `explore:` block:

```yaml
explore:
  display_name: Sales Dashboard
  defaults:
    time_range: P7D
    measures:
      - total_revenue
      - order_count
```

**Legacy behavior**: Files with `version: 1` do NOT auto-generate an explore. Omit `version:` in new metrics views to get the auto-generated explore.

## Full Example

Here is a complete, annotated metrics view:

```yaml
# metrics/orders.yaml
type: metrics_view

# Display metadata
display_name: Orders Analytics
description: Analyze order performance by various dimensions

# Data source - references the 'orders' model in the project
model: orders

# Time column for time-series charts and filtering
timeseries: order_date
smallest_time_grain: day

# Dimensions for grouping and filtering
dimensions:
  - name: order_date
    display_name: Order Date
    column: order_date

  - name: country
    display_name: Country
    column: shipping_country

  - name: product_category
    display_name: Product Category
    column: category
    description: High-level product grouping

  - name: customer_segment
    display_name: Customer Segment
    expression: | 
      CASE
        WHEN lifetime_value > 1000 THEN 'High Value'
        WHEN lifetime_value > 100 THEN 'Medium Value'
        ELSE 'Low Value'
      END

  - name: is_repeat_customer
    display_name: Repeat Customer
    expression: CASE WHEN order_number > 1 THEN 'Yes' ELSE 'No' END

# Measures for aggregation
measures:
  - name: total_orders
    display_name: Total Orders
    expression: COUNT(*)
    format_preset: humanize

  - name: total_revenue
    display_name: Total Revenue
    expression: SUM(order_total)
    format_preset: currency_usd
    description: Gross revenue before refunds

  - name: average_order_value
    display_name: Avg Order Value
    expression: SUM(order_total) / NULLIF(COUNT(*), 0)
    format_preset: currency_usd
    valid_percent_of_total: false

  - name: unique_customers
    display_name: Unique Customers
    expression: COUNT(DISTINCT customer_id)
    format_preset: humanize

  - name: items_per_order
    display_name: Items per Order
    expression: SUM(item_count) / NULLIF(COUNT(*), 0)
    format_d3: ",.1f"
    valid_percent_of_total: false
```

## Security Policies

Security policies control who can access a metrics view and what data they can see. This is a powerful feature for multi-tenant dashboards and role-based access control.

### Basic access control

The `access:` property controls whether users can view the metrics view at all:

```yaml
security:
  # Allow access for everyone
  access: true

  # Deny access for everyone (useful for draft dashboards)
  access: false

  # Conditional access based on user attributes
  access: "'{{ .user.admin }}' = 'true'"
```

The expression syntax should be a DuckDB expression, which will be evaluated in a sandbox without access to any tables.

### Row-level security

The `row_filter:` property restricts which rows a user can see. It's a SQL expression that references user attributes via templating:

```yaml
security:
  access: true
  row_filter: domain = '{{ .user.domain }}'
```

Common user attributes:
- `{{ .user.email }}`: User's email address
- `{{ .user.domain }}`: Email domain (e.g., "acme.com")
- `{{ .user.admin }}`: Boolean admin flag
- Custom attributes configured in Rill Cloud

The row filter should use the SQL syntax of the metrics view's model, and can reference other tables in the model's connector.

### Complex row filters

Use logical operators for sophisticated access patterns:

```yaml
security:
  access: true
  row_filter: >
    {{ .user.admin }}
    OR '{{ .user.domain }}' = 'acme.com'
    {{ if hasKey .user "tenant_id" }}
    OR tenant_id = '{{ .user.tenant_id }}'
    {{ end }}
```

### Hiding dimensions and measures

The `exclude:` property conditionally hides specific dimensions or measures from certain users:

```yaml
security:
  access: true
  exclude:
    - if: "NOT {{ .user.admin }}"
      names:
        - cost_per_acquisition  # Hide sensitive cost data from non-admins
        - internal_notes
```

## Advanced Features

### Annotations

Annotations overlay contextual information (like events or milestones) on time-series charts:

```yaml
annotations:
  - name: product_launches
    model: product_launches  # Must have 'time' and 'description' columns
    measures:
      - total_revenue        # Only show on these measures

  # Optional columns in annotation model:
  # - time_end: For range annotations
  # - grain: Show only at specific time grains (day, week, etc.)
```

### Unnest for array dimensions

When a column contains arrays, use `unnest: true` to flatten it at query time:

```yaml
dimensions:
  - name: tags
    display_name: Tags
    column: tags
    unnest: true  # Allows filtering by individual array elements
```

### Cache configuration

Configure caching for slow metrics views that use external tables:

```yaml
cache:
  enabled: true
  key_ttl: 5m
  key_sql: SELECT MAX(updated_at) FROM orders
```

You should not add a `cache:` config when the metrics view references a model inside the project since Rill does automatic cache management in that case.

## Dialect-Specific Notes

SQL expressions in dimensions and measures use the underlying OLAP database's dialect.

### DuckDB

DuckDB is the default OLAP engine for local development.

**Conditional aggregation with FILTER**:
```yaml
# DuckDB supports FILTER clause for conditional aggregation
expression: COUNT(*) FILTER (WHERE status = 'completed')
```

### ClickHouse

ClickHouse is recommended for production workloads with large datasets.

**Conditional aggregation**:
```yaml
# ClickHouse uses IF or CASE inside aggregations
expression: countIf(status = 'completed')
expression: sumIf(revenue, status = 'completed')
```

**Date functions**:
```yaml
expression: toYear(order_date)
expression: toStartOfMonth(order_date)
expression: toYYYYMMDD(order_date)
```

**Array functions**:
```yaml
expression: arrayJoin(tags)  # Unnest arrays
```

### Druid

**Approximate distinct counts**:
```yaml
expression: APPROX_COUNT_DISTINCT_DS_HLL(user_id)
```

## JSON Schema

```
{% json_schema_for_resource "metrics_view" %}
```
