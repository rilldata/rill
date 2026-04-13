---
name: rill-metrics-view
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

## Reference documentation

Here is a full JSON schema for the metrics view syntax:

```
allOf:
    - properties:
        ai_instructions:
            description: Extra instructions for [AI agents](/guide/ai/mcp). Used to guide natural language question answering and routing.
            type: string
        annotations:
            description: Used to define annotations that can be displayed on charts
            items:
                properties:
                    connector:
                        description: Refers to the connector to use for the annotation
                        type: string
                    database:
                        description: Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
                        type: string
                    database_schema:
                        description: Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
                        type: string
                    measures:
                        anyOf:
                            - description: Simple field name as a string.
                              type: string
                            - description: List of field selectors, each can be a string or an object with detailed configuration.
                              items:
                                anyOf:
                                    - description: Shorthand field selector, interpreted as the name.
                                      type: string
                                    - additionalProperties: false
                                      description: Detailed field selector configuration with name and optional time grain.
                                      properties:
                                        name:
                                            description: Name of the field to select.
                                            type: string
                                        time_grain:
                                            description: Time grain for time-based dimensions.
                                            enum:
                                                - ""
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
                                                - "y"
                                                - year
                                            type: string
                                      required:
                                        - name
                                      type: object
                              type: array
                        description: Specifies which measures to apply the annotation to. Applies to all measures if not specified
                    model:
                        description: Refers to the model powering the annotation (either table or model is required). The model must have 'time' and 'description' columns. Optional columns include 'time_end' for range annotations and 'grain' to specify when the annotation should appear based on dashboard grain level.
                        type: string
                    name:
                        description: A stable identifier for the annotation. Defaults to model or table names when not specified
                        type: string
                    table:
                        description: Refers to the table powering the annotation, should be used instead of model for annotations from external OLAP tables (either table or model is required)
                        type: string
                type: object
            type: array
        connector:
            description: Refers to the connector type for the metrics view, see [OLAP engines](/developers/build/connectors/olap) for more information
            type: string
        database:
            description: Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
            type: string
        database_schema:
            description: Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified
            type: string
        description:
            description: Refers to the description for the metrics view
            type: string
        dimensions:
            description: Relates to exploring segments or dimensions of your data and filtering the dashboard
            items:
                anyOf:
                    - required:
                        - column
                    - required:
                        - expression
                properties:
                    column:
                        description: a categorical column
                        type: string
                    description:
                        description: a freeform text description of the dimension
                        type: string
                    display_name:
                        description: a display name for your dimension
                        type: string
                    expression:
                        description: a non-aggregate expression such as string_split(domain, '.'). One of column and expression is required but cannot have both at the same time
                        type: string
                    name:
                        description: a stable identifier for the dimension
                        type: string
                    tags:
                        description: optional list of tags for categorizing the dimension (defaults to empty)
                        items:
                            type: string
                        type: array
                    type:
                        description: 'Dimension type: "geo" for geospatial dimensions, "time" for time dimensions or "categorical" for categorial dimensions. Default is undefined and the type will be inferred instead'
                        type: string
                    unnest:
                        description: if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match
                        type: boolean
                    uri:
                        description: enable if your dimension is a clickable URL to enable single click navigation (boolean or valid SQL expression)
                        type:
                            - string
                            - boolean
                type: object
            type: array
        display_name:
            description: Refers to the display name for the metrics view
            type: string
        explore:
            $ref: '#/definitions/explore_properties'
            description: Defines an optional inline explore view for the metrics view. If not specified a default explore will be emitted unless `skip` is set to true.
            required:
                - type
        first_day_of_week:
            description: Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=1 and Sunday=7
            type: integer
        first_month_of_year:
            description: Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=1 and December=12
            type: integer
        measures:
            description: Used to define the numeric aggregates of columns from your data model
            items:
                properties:
                    description:
                        description: a freeform text description of the measure
                        type: string
                    display_name:
                        description: the display name of your measure.
                        type: string
                    expression:
                        description: a combination of operators and functions for aggregations
                        type: string
                    format_d3:
                        description: 'Controls the formatting of this measure using a [d3-format](https://d3js.org/d3-format) string. If an invalid format string is supplied, the measure will fall back to `format_preset: humanize`. A measure cannot have both `format_preset` and `format_d3`. If neither is provided, the humanize preset is used by default. Example: `format_d3: ".2f"` formats using fixed-point notation with two decimal places. Example: `format_d3: ",.2r"` formats using grouped thousands with two significant digits. (optional)'
                        type: string
                    format_d3_locale:
                        description: |
                            locale configuration passed through to D3, enabling changing the currency symbol among other things. For details, see the docs for D3's formatLocale.
                              ```yaml
                              format_d3: "$,"
                              format_d3_locale:
                                grouping: [3, 2]
                                currency: ["₹", ""]
                              ```
                        properties:
                            currency:
                                description: the currency symbol
                                type: array
                            grouping:
                                description: the grouping of the currency symbol
                                type: array
                        type: object
                    format_preset:
                        description: |
                            Controls the formatting of this measure using a predefined preset. Measures cannot have both `format_preset` and `format_d3`. If neither is supplied, the measure will be formatted using the `humanize` preset by default.

                              Available options:
                              - `humanize`: Round numbers into thousands (K), millions(M), billions (B), etc.
                              - `none`: Raw output.
                              - `currency_usd`: Round to 2 decimal points with a dollar sign ($).
                              - `currency_eur`: Round to 2 decimal points with a euro sign (€).
                              - `percentage`: Convert a rate into a percentage with a % sign.
                              - `interval_ms`: Convert milliseconds into human-readable durations like hours (h), days (d), years (y), etc. (optional)
                        type: string
                    label:
                        description: a label for your measure, deprecated use display_name
                        type: string
                    name:
                        description: a stable identifier for the measure
                        type: string
                    per:
                        $ref: '#/definitions/field_selectors_properties'
                        description: for per dimensions
                    requires:
                        $ref: '#/definitions/field_selectors_properties'
                        description: using an available measure or dimension in your metrics view to set a required parameter, cannot be used with simple measures. See [referencing measures](/developers/build/metrics-view/measures/referencing) for more information.
                    tags:
                        description: optional list of tags for categorizing the measure (defaults to empty)
                        items:
                            type: string
                        type: array
                    treat_nulls_as:
                        description: used to configure what value to fill in for missing time buckets. This also works generally as COALESCING over non empty time buckets.
                        type: string
                    type:
                        description: 'Measure calculation type: "simple" for basic aggregations, "derived" for calculations using other measures, or "time_comparison" for period-over-period analysis. Defaults to "simple" unless dependencies exist.'
                        type: string
                    valid_percent_of_total:
                        description: a boolean indicating whether percent-of-total values should be rendered for this measure
                        type: boolean
                    window:
                        anyOf:
                            - description: 'Shorthand: `time` or `true` means time-partitioned, `all` means non-partitioned.'
                              enum:
                                - time
                                - "true"
                                - all
                              type: string
                            - additionalProperties: false
                              description: Detailed window configuration for measure calculations, allowing control over partitioning, ordering, and frame definition.
                              properties:
                                frame:
                                    description: Defines the window frame boundaries for calculations, specifying which rows are included in the window relative to the current row.
                                    type: string
                                order:
                                    $ref: '#/definitions/field_selectors_properties'
                                    description: Specifies the fields to order the window by, determining the sequence of rows within each partition.
                                    type: string
                                partition:
                                    description: Controls whether the window is partitioned. When true, calculations are performed within each partition separately.
                                    type: boolean
                              type: object
                        description: A measure window can be defined as a keyword string (e.g. 'time' or 'all') or an object with detailed window configuration. For more information, see the [window functions](/developers/build/metrics-view/measures/windows) documentation.
                required:
                    - name
                    - display_name
                    - expression
                type: object
            type: array
        model:
            description: Refers to the model powering the dashboard (either model or table is required)
            type: string
        parent:
            description: Refers to the parent metrics view from which this metrics view is derived. If specified, this will inherit properties from the parent metrics view
            type: string
        parent_dimensions:
            $ref: '#/definitions/field_selector_properties'
            description: Optional field selectors for dimensions to inherit from the parent metrics view.
        parent_measures:
            $ref: '#/definitions/field_selector_properties'
            description: Optional field selectors for measures to inherit from the parent metrics view.
        security:
            $ref: '#/definitions/security_policy_properties'
            description: Defines a security policy for the dashboard
        smallest_time_grain:
            description: 'Refers to the smallest time granularity the user is allowed to view. The valid values are: millisecond, second, minute, hour, day, week, month, quarter, year'
            type: string
        table:
            description: Refers to the table powering the dashboard, should be used instead of model for dashboards create from external OLAP tables (either table or model is required)
            type: string
        timeseries:
            description: Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear
            type: string
        type:
            const: metrics_view
            description: Refers to the resource type and must be `metrics_view`
            type: string
        version:
            description: The version of the metrics view schema
            type: string
        watermark:
            description: A SQL expression that tells us the max timestamp that the measures are considered valid for. Usually does not need to be overwritten
            type: string
      title: Properties
      type: object
    - $ref: '#/definitions/common_properties'
description: In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.
id: metrics-views
title: Metrics View YAML
type: object
```