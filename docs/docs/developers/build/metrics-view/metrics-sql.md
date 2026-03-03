---
title: Metrics SQL 
description: Query metrics views using SQL syntax
sidebar_label: Metrics SQL
---

You can write a SQL query referring to metrics definitions and dimensions defined in a metrics view.
It should have the following structure:
    
```yaml
type: api
metrics_sql: SELECT publisher, domain, total_records FROM ad_bids_metrics
```

## Querying Fundamentals

Metrics SQL transforms queries that reference `dimensions` and `measures` within a `metrics view` into their corresponding database columns or expressions. This transformation is based on the mappings defined in a metrics view YAML configuration, enabling reuse of dimension or measure definitions. Additionally, any security policies defined in the metrics view are also inherited.

## Example: Crafting a Metrics SQL Query

Consider a metrics view configured as follows:
```yaml
#metrics/ad_bids_metrics.yaml
type: metrics_view
title: Ad Bids
model: ad_bids
timeseries: timestamp
dimensions:
  - name: publisher
    expression: toUpper(publisher)
  - name: domain
    column: domain
measures:
  - name: total_records
    display_name: Total records
    expression: COUNT(*)
```

To query this view, a user might write a Metrics SQL query like:
```sql
SELECT publisher, domain, total_records FROM ad_bids_metrics
```

This Metrics SQL is internally translated to a standard SQL query as follows:
```sql
SELECT toUpper(publisher) AS publisher, domain AS domain, COUNT(*) AS total_records FROM ad_bids_metrics GROUP BY publisher, domain
```

## Security and Compliance

Queries executed via Metrics SQL are subject to the security policies and access controls defined in the metrics view YAML configuration, ensuring data security and compliance.

## Limitations

Metrics SQL is specifically designed for querying metrics views and may not support all features found in standard SQL. Its primary focus is on providing an efficient and easy way to extract data within the constraints of metrics view configurations.

## Supported SQL Features

- **SELECT** statements with plain `dimension` and `measure` references.
- A single **FROM** clause referencing a `metrics view`.
- **WHERE** clause that can reference selected `dimensions` only.
- Operators in **WHERE** and **HAVING** clauses include `=`, `!=`, `>`, `>=`, `<`, `<=`, IN, LIKE, AND, OR, and parentheses for structuring the expression.
- **HAVING** clause for filtering on aggregated results, referencing selected dimension and measure names. Supports the same expression capabilities as the WHERE clause.
- **ORDER BY** clause for sorting the results.
- **LIMIT** and **OFFSET** clauses for controlling the result set size and pagination.

:::warning
 The Metrics SQL feature is currently evolving. We are dedicated to enhancing the syntax by introducing additional SQL features, while striving to maintain support for existing syntax. However, please be advised that backward compatibility cannot be guaranteed at all times. Additionally, users should be aware that there may be untested edge cases in the current implementation. We appreciate your understanding as we work to refine and improve this feature.
:::

## Using Metrics SQL in custom APIs

To expose Metrics SQL queries as HTTP API endpoints, see the [Metrics SQL APIs](/developers/build/custom-apis/metrics-sql) guide. You can also add [dynamic templating](/developers/build/custom-apis/templating), [security rules](/developers/build/custom-apis/security), and [OpenAPI documentation](/developers/build/custom-apis/openapi) to your Metrics SQL APIs.

