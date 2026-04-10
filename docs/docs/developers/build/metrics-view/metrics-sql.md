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

## Supported SQL Features

### SELECT

Reference dimensions and measures by name. The `date_trunc` function can be used to group a time dimension by a specific grain (and optionally aliased with `AS`):

```sql
SELECT date_trunc('MONTH', timestamp) AS month, publisher, total_records FROM ad_bids_metrics
```

Supported grains: `SECOND`, `MINUTE`, `HOUR`, `DAY`, `WEEK`, `MONTH`, `QUARTER`, `YEAR`.

### FROM

A single metrics view name. Joins and subqueries in the FROM clause are not supported.

### WHERE and HAVING

`WHERE` filters on dimensions; `HAVING` filters on aggregated measures. Both support the same operators and functions.

### ORDER BY, LIMIT, OFFSET

Standard SQL sorting and pagination clauses are supported:

```sql
SELECT publisher, total_records FROM ad_bids_metrics
ORDER BY total_records DESC
LIMIT 20 OFFSET 40
```

## Operators

The following operators are supported in `WHERE` and `HAVING` clauses:

| Operator |
|----------|
| `=`, `!=`, `<`, `<=`, `>`, `>=` |
| `AND`, `OR`, `(` `)` |
| `IN (...)`, `NOT IN (...)` |
| `LIKE`, `NOT LIKE` (case-insensitive, `%` wildcard) |
| `BETWEEN ... AND ...` |
| `IS NULL`, `IS NOT NULL` |
| `IS TRUE`, `IS FALSE`, `IS NOT TRUE`, `IS NOT FALSE` |

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE (publisher IS NOT NULL AND domain LIKE '%google%')
   OR publisher IN ('Yahoo', 'Microsoft')
```

## Functions

### Time range functions

`time_range_start` and `time_range_end` resolve a [Rill time expression](/developers/build/metrics-view/time-series/rill-time) against the metrics view's watermark and time range. They must be compared against the time dimension:

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE timestamp > time_range_start('7D as of watermark/D+1D')
  AND timestamp <= time_range_end('7D as of watermark/D+1D')
```

### Interval arithmetic

Add or subtract an interval from a timestamp literal using `INTERVAL amount UNIT` syntax. Supported units: `SECOND`, `MINUTE`, `HOUR`, `DAY`, `WEEK`, `MONTH`, `YEAR`.

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE timestamp > '2024-07-30' - INTERVAL 90 DAY
```

### now()

Returns the current timestamp:

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE timestamp > now() - INTERVAL 7 DAY
```

### CAST

Casting to `DATETIME` or `TIMESTAMP` is supported (other target types are not):

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE timestamp > CAST('2024-01-01' AS TIMESTAMP)
```

## Subqueries

Subqueries are supported inside `IN` expressions. The subquery must select exactly one dimension from the same metrics view and can include its own `WHERE` and `HAVING` clauses:

```sql
SELECT publisher, total_records FROM ad_bids_metrics
WHERE publisher IN (
  SELECT publisher FROM ad_bids_metrics
  HAVING total_records > 100
)
```

Subqueries do not support `ORDER BY`, `LIMIT`, `DISTINCT`, window functions, joins, or CTEs.

## Limitations

- Only one metrics view can be queried per statement (no joins).
- `SELECT *` is not supported; list dimensions and measures explicitly.
- `GROUP BY` is implicit based on selected dimensions and cannot be specified manually.
- Aggregate functions like `COUNT()` or `SUM()` cannot be used directly; reference predefined measures instead.
- Set operations (`UNION`, `INTERSECT`, `EXCEPT`) and CTEs (`WITH`) are not supported.

:::warning
 The Metrics SQL feature is currently evolving. We are dedicated to enhancing the syntax by introducing additional SQL features, while striving to maintain support for existing syntax. However, please be advised that backward compatibility cannot be guaranteed at all times. Additionally, users should be aware that there may be untested edge cases in the current implementation. We appreciate your understanding as we work to refine and improve this feature.
:::

## Using Metrics SQL in custom APIs

To expose Metrics SQL queries as HTTP API endpoints, see the [Metrics SQL APIs](/developers/build/custom-apis/metrics-sql) guide. You can also add [dynamic templating](/developers/build/custom-apis/templating), [security rules](/developers/build/custom-apis/security), and [OpenAPI documentation](/developers/build/custom-apis/openapi) to your Metrics SQL APIs.
