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
model: my_external_table
connector: clickhouse  # Optional: specify if different from default
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
    display_name: Country
    column: country

  # Computed expression
  - name: device_category
    display_name: Device Category
    expression: CASE WHEN device_type IN ('phone', 'tablet') THEN 'Mobile' ELSE 'Desktop' END

  # With description
  - name: campaign_name
    display_name: Campaign
    column: campaign_name
    description: Marketing campaign that drove the traffic
```

**Naming**: Each dimension needs a `name` (stable identifier used in APIs and references) and a `display_name` (human-readable label shown in the UI). If only `column:` is provided, both default to the column name.

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
- `humanize`: Round to K, M, B (e.g., 1.2M)
- `none`: Raw number
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

### Auto-generated explore

When you create a metrics view, Rill automatically generates an explore dashboard with the same name, exposing all dimensions and measures. To customize the explore, add an `explore:` block:

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
    expression: CASE
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
  access: "'{{ .user.is_admin }}' = 'true'"
```

### Row-level security

The `row_filter:` property restricts which rows a user can see. It's a SQL expression that references user attributes via templating:

```yaml
security:
  access: true
  row_filter: >
    region IN ('{{ .user.allowed_regions | join "', '" }}')
```

Common user attributes:
- `{{ .user.email }}`: User's email address
- `{{ .user.domain }}`: Email domain (e.g., "acme.com")
- `{{ .user.name }}`: User's display name
- `{{ .user.admin }}`: Boolean admin flag
- Custom attributes configured in your identity provider

### Complex row filters

Use logical operators for sophisticated access patterns:

```yaml
security:
  access: true
  row_filter: >
    '{{ .user.is_admin }}' = 'true'
    OR '{{ .user.domain }}' = 'acme.com'
    OR client_id IN ('{{ .user.allowed_clients | join "', '" }}')
```

### Hiding dimensions and measures

The `exclude:` property conditionally hides specific dimensions or measures from certain users:

```yaml
security:
  access: true
  row_filter: client_id IN ('{{ .user.client_ids | join "', '" }}')
  exclude:
    - if: "'{{ .user.is_admin }}' != 'true'"
      names:
        - cost_per_acquisition  # Hide sensitive cost data from non-admins
        - internal_notes
```

### Security policy examples

**Multi-tenant SaaS dashboard**:
```yaml
security:
  access: "'{{ .user.subscription_tier }}' != 'free'"
  row_filter: tenant_id = '{{ .user.tenant_id }}'
```

**Publisher dashboard with regional access**:
```yaml
security:
  access: true
  row_filter: >
    '{{ .user.role }}' = 'admin'
    OR publisher_id IN ('{{ .user.publisher_ids | join "', '" }}')
  exclude:
    - if: "'{{ .user.role }}' != 'admin'"
      names:
        - revenue_share_percentage
        - internal_score
```

**Domain-restricted access with admin override**:
```yaml
security:
  access: true
  row_filter: >
    '{{ .user.domain }}' = 'mycompany.com'
    OR site IN ('{{ .user.allowed_sites | join "', '" }}')
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

When a column contains arrays, use `unnest: true` to enable "contains" filtering:

```yaml
dimensions:
  - name: tags
    display_name: Tags
    column: tags
    unnest: true  # Allows filtering by individual array elements
```

### Cache configuration

Configure caching for expensive metrics views:

```yaml
cache:
  enabled: true
  key_ttl: 5m
  key_sql: SELECT MAX(updated_at) FROM orders
```

### Time zones

Pin specific time zones to the dashboard:

```yaml
available_time_zones:
  - America/New_York
  - America/Los_Angeles
  - Europe/London
  - UTC
```

### Default time range

Set the initial time range when users open the dashboard:

```yaml
default_time_range: P7D  # Last 7 days (ISO 8601 duration)
```

## Dialect-Specific Notes

SQL expressions in dimensions and measures use the underlying OLAP database's dialect.

### DuckDB

DuckDB is the default OLAP engine for local development.

**Conditional aggregation with FILTER**:
```yaml
# DuckDB supports FILTER clause for conditional aggregation
expression: COUNT(*) FILTER (WHERE status = 'completed')
```

**Date extraction**:
```yaml
expression: EXTRACT(YEAR FROM order_date)
expression: DATE_TRUNC('month', order_date)
```

**String functions**:
```yaml
expression: CONCAT(first_name, ' ', last_name)
expression: SPLIT_PART(email, '@', 2)  # Get domain from email
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
expression: has(tags, 'featured')  # Check array contains value
```

### Druid

Druid is optimized for real-time analytics.

**Time functions**:
```yaml
expression: TIME_FLOOR(__time, 'P1D')
expression: TIME_EXTRACT(__time, 'YEAR')
```

**Approximate distinct counts**:
```yaml
expression: APPROX_COUNT_DISTINCT_DS_HLL(user_id)
```

## JSON Schema

```
metrics-views:
    title: Metrics View YAML
    id: metrics-views
    type: object
    description: In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.
    allOf:
      - title: Properties
        type: object
        properties:
          version:
            type: string
            description: The version of the metrics view schema
          type:
            type: string
            const: metrics_view
            description: Refers to the resource type and must be `metrics_view`
          connector:
            type: string
            description: Refers to the connector type for the metrics view, see [OLAP engines](/build/connectors/olap) for more information
          display_name:
            type: string
            description: Refers to the display name for the metrics view
          description:
            type: string
            description: Refers to the description for the metrics view
          ai_instructions:
            type: string
            description: Extra instructions for [AI agents](/explore/mcp). Used to guide natural language question answering and routing.
          parent:
            type: string
            description: Refers to the parent metrics view from which this metrics view is derived. If specified, this will inherit properties from the parent metrics view
          model:
            type: string
            description: Refers to the model powering the dashboard (either model or table is required)
          database:
            type: string
            description: Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
          database_schema:
            type: string
            description: Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
          table:
            type: string
            description: Refers to the table powering the dashboard, should be used instead of model for dashboards create from external OLAP tables (either table or model is required)
          timeseries:
            type: string
            description: Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear
          watermark:
            type: string
            description: A SQL expression that tells us the max timestamp that the measures are considered valid for. Usually does not need to be overwritten
          smallest_time_grain:
            type: string
            description: 'Refers to the smallest time granularity the user is allowed to view. The valid values are: millisecond, second, minute, hour, day, week, month, quarter, year'
          first_day_of_week:
            type: integer
            description: Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=1 and Sunday=7
          first_month_of_year:
            type: integer
            description: Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=1 and December=12
          dimensions:
            type: array
            description: Relates to exploring segments or dimensions of your data and filtering the dashboard
            items:
              type: object
              properties:
                name:
                  type: string
                  description: a stable identifier for the dimension
                display_name:
                  type: string
                  description: a display name for your dimension
                description:
                  type: string
                  description: a freeform text description of the dimension
                tags:
                  type: array
                  description: optional list of tags for categorizing the dimension (defaults to empty)
                  items:
                    type: string
                type:
                    type: string
                    description: 'Dimension type: "geo" for geospatial dimensions, "time" for time dimensions or "categorical" for categorial dimensions. Default is undefined and the type will be inferred instead'
                column:
                  type: string
                  description: a categorical column
                expression:
                  type: string
                  description: a non-aggregate expression such as string_split(domain, '.'). One of column and expression is required but cannot have both at the same time
                unnest:
                  type: boolean
                  description: if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match
                uri:
                  type:
                    - string
                    - boolean
                  description: enable if your dimension is a clickable URL to enable single click navigation (boolean or valid SQL expression)
              anyOf:
                - required:
                    - column
                - required:
                    - expression
          measures:
            type: array
            description: Used to define the numeric aggregates of columns from your data model
            items:
              type: object
              properties:
                name:
                  type: string
                  description: a stable identifier for the measure
                display_name:
                  type: string
                  description: the display name of your measure.
                label:
                  type: string
                  description: a label for your measure, deprecated use display_name
                description:
                  type: string
                  description: a freeform text description of the measure
                tags:
                  type: array
                  description: optional list of tags for categorizing the measure (defaults to empty)
                  items:
                    type: string
                type:
                  type: string
                  description: 'Measure calculation type: "simple" for basic aggregations, "derived" for calculations using other measures, or "time_comparison" for period-over-period analysis. Defaults to "simple" unless dependencies exist.'
                expression:
                  type: string
                  description: a combination of operators and functions for aggregations
                window:
                  description: A measure window can be defined as a keyword string (e.g. 'time' or 'all') or an object with detailed window configuration. For more information, see the [window functions](/build/metrics-view/measures/windows) documentation.
                  anyOf:
                    - type: string
                      enum:
                        - time
                        - 'true'
                        - all
                      description: 'Shorthand: `time` or `true` means time-partitioned, `all` means non-partitioned.'
                    - type: object
                      description: 'Detailed window configuration for measure calculations, allowing control over partitioning, ordering, and frame definition.'
                      properties:
                        partition:
                          type: boolean
                          description: 'Controls whether the window is partitioned. When true, calculations are performed within each partition separately.'
                        order:
                          type: string
                          $ref: '#/definitions/field_selectors_properties'
                          description: 'Specifies the fields to order the window by, determining the sequence of rows within each partition.'
                        frame:
                          type: string
                          description: 'Defines the window frame boundaries for calculations, specifying which rows are included in the window relative to the current row.'
                      additionalProperties: false
                per:
                  $ref: '#/definitions/field_selectors_properties'
                  description: for per dimensions
                requires:
                  $ref: '#/definitions/field_selectors_properties'
                  description: using an available measure or dimension in your metrics view to set a required parameter, cannot be used with simple measures. See [referencing measures](/build/metrics-view/measures/referencing) for more information.
                valid_percent_of_total:
                  type: boolean
                  description: a boolean indicating whether percent-of-total values should be rendered for this measure
                format_preset:
                  type: string
                  description: |
                    Controls the formatting of this measure using a predefined preset. Measures cannot have both `format_preset` and `format_d3`. If neither is supplied, the measure will be formatted using the `humanize` preset by default.

                      Available options:
                      - `humanize`: Round numbers into thousands (K), millions(M), billions (B), etc.
                      - `none`: Raw output.
                      - `currency_usd`: Round to 2 decimal points with a dollar sign ($).
                      - `currency_eur`: Round to 2 decimal points with a euro sign (€).
                      - `percentage`: Convert a rate into a percentage with a % sign.
                      - `interval_ms`: Convert milliseconds into human-readable durations like hours (h), days (d), years (y), etc. (optional)
                format_d3:
                  type: string
                  description: 'Controls the formatting of this measure using a [d3-format](https://d3js.org/d3-format) string. If an invalid format string is supplied, the measure will fall back to `format_preset: humanize`. A measure cannot have both `format_preset` and `format_d3`. If neither is provided, the humanize preset is used by default. Example: `format_d3: ".2f"` formats using fixed-point notation with two decimal places. Example: `format_d3: ",.2r"` formats using grouped thousands with two significant digits. (optional)'
                format_d3_locale:
                  type: object
                  description: |
                      locale configuration passed through to D3, enabling changing the currency symbol among other things. For details, see the docs for D3's formatLocale.
                        ```yaml
                        format_d3: "$,"
                        format_d3_locale:
                          grouping: [3, 2]
                          currency: ["₹", ""]
                        ```
                  properties:
                    grouping:
                      type: array
                      description: the grouping of the currency symbol
                    currency:
                      type: array
                      description: the currency symbol

                treat_nulls_as:
                  type: string
                  description: used to configure what value to fill in for missing time buckets. This also works generally as COALESCING over non empty time buckets.

              required:
                - name
                - display_name
                - expression

          parent_dimensions:
            description: Optional field selectors for dimensions to inherit from the parent metrics view.
            $ref: '#/definitions/field_selector_properties'
          parent_measures:
            description: Optional field selectors for measures to inherit from the parent metrics view.
            $ref: '#/definitions/field_selector_properties'
          annotations:
            type: array
            description: Used to define annotations that can be displayed on charts
            items:
              type: object
              properties:
                name:
                  type: string
                  description: A stable identifier for the annotation. Defaults to model or table names when not specified
                model:
                  type: string
                  description: Refers to the model powering the annotation (either table or model is required). The model must have 'time' and 'description' columns. Optional columns include 'time_end' for range annotations and 'grain' to specify when the annotation should appear based on dashboard grain level.
                database:
                  type: string
                  description: Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
                database_schema:
                  type: string
                  description: Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
                table:
                  type: string
                  description: Refers to the table powering the annotation, should be used instead of model for annotations from external OLAP tables (either table or model is required)
                connector:
                  type: string
                  description: Refers to the connector to use for the annotation
                measures:
                  description: Specifies which measures to apply the annotation to. Applies to all measures if not specified
                  anyOf:
                    - type: string
                      description: Simple field name as a string.
                    - type: array
                      description: List of field selectors, each can be a string or an object with detailed configuration.
                      items:
                        anyOf:
                          - type: string
                            description: Shorthand field selector, interpreted as the name.
                          - type: object
                            description: Detailed field selector configuration with name and optional time grain.
                            properties:
                              name:
                                type: string
                                description: Name of the field to select.
                              time_grain:
                                type: string
                                description: Time grain for time-based dimensions.
                                enum:
                                  - ''
                                  - ms
                                  - millisecond
                                  - s
                                  - second
                                  - min
                                  - minute
                                  - hadditionalProperties: fal
                                  - hour
                                  - d
                                  - day
                                  - w
                                  - week
                                  - month
                                  - q
                                  - quarter
                                  - 'y'
                                  - year
                            required:
                              - name
                            additionalProperties: false
          security:
              $ref: '#/definitions/security_policy_properties'
              description: Defines a security policy for the dashboard
          explore:
            $ref: '#/definitions/explore_properties'
            description: Defines an optional inline explore view for the metrics view. If not specified a default explore will be emitted unless `skip` is set to true.
            required:
              - type

      - $ref: '#/definitions/common_properties'
```
