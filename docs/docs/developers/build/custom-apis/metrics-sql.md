---
title: Metrics SQL APIs
description: Create custom APIs that query metrics views using dimension and measure names
sidebar_label: Metrics SQL APIs
sidebar_position: 30
---

Metrics SQL APIs let you query [metrics views](/developers/build/metrics-view) using the dimension and measure names you've already defined. Instead of writing raw SQL against underlying tables, you write queries using your metrics view's semantic layer — and security policies are inherited automatically.

## Basic syntax

Create a YAML file in your project's `apis/` directory:

```yaml
type: api
metrics_sql: SELECT publisher, domain, total_records FROM ad_bids_metrics
```

## How Metrics SQL works

Metrics SQL transforms your query by replacing dimension and measure names with their underlying expressions. Consider this metrics view:

```yaml
# metrics/ad_bids_metrics.yaml
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
    expression: COUNT(*)
  - name: avg_bid_price
    expression: AVG(bid_price)
```

When you write:
```sql
SELECT publisher, domain, total_records FROM ad_bids_metrics
```

Rill translates this to:
```sql
SELECT toUpper(publisher) AS publisher, domain AS domain, COUNT(*) AS total_records
FROM ad_bids
GROUP BY publisher, domain
```

This means you write simple queries using business-friendly names, while Rill handles the underlying SQL complexity.

## Why use Metrics SQL over raw SQL?

| Benefit | Description |
|---------|-------------|
| **Reuse definitions** | Query using dimension/measure names defined once in your metrics view |
| **Automatic security** | Row-level security policies from the metrics view are applied automatically |
| **Simpler queries** | No need to remember complex expressions — just use names like `total_records` |
| **Consistency** | All APIs and dashboards use the same metric definitions |

## Examples

### Filtering with WHERE

```yaml
type: api
metrics_sql: |
  SELECT publisher, total_records
  FROM ad_bids_metrics
  WHERE domain = 'google.com'
  ORDER BY total_records DESC
  LIMIT 10
```

### Aggregation with HAVING

```yaml
type: api
metrics_sql: |
  SELECT publisher, total_records, avg_bid_price
  FROM ad_bids_metrics
  HAVING total_records > 1000
  ORDER BY avg_bid_price DESC
```

### Pagination

```yaml
type: api
metrics_sql: |
  SELECT publisher, domain, total_records
  FROM ad_bids_metrics
  ORDER BY total_records DESC
  LIMIT 20 OFFSET 40
```

## Security inheritance

When you use Metrics SQL, any [security policies](/developers/build/metrics-view/security) defined on the metrics view are automatically enforced. For example, if your metrics view has:

```yaml
# In your metrics view
security:
  access: true
  row_filter: "domain = '{{ .user.domain }}'"
```

Then a Metrics SQL API querying this view will automatically filter rows based on the user's domain — no additional configuration needed in the API definition.

This is one of the key advantages of Metrics SQL over raw SQL APIs. See [Security & Access Control](/developers/build/custom-apis/security) for more on how security works with custom APIs.

## Supported SQL features

For a full reference on supported SQL syntax (SELECT, WHERE, HAVING, ORDER BY, LIMIT, OFFSET) and current limitations, see the [Metrics SQL language reference](/developers/build/metrics-view/metrics-sql).

## Adding dynamic behavior

Metrics SQL APIs support the same [templating](/developers/build/custom-apis/templating) as SQL APIs:

```yaml
type: api
metrics_sql: |
  SELECT publisher, total_records
  FROM ad_bids_metrics
  {{ if hasKey .args "domain" }}
    WHERE domain = '{{ .args.domain }}'
  {{ end }}
  ORDER BY total_records DESC
  LIMIT {{ default 25 .args.limit }}
```

See [Dynamic Queries with Templating](/developers/build/custom-apis/templating) for the full guide.
