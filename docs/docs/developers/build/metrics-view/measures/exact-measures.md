---
title: "Approximate and Exact Measures"
description: Serve fast approximate aggregates by default and let viewers switch to exact values on demand
sidebar_label: "Approximate and Exact"
sidebar_position: 45
---

Distinct counts, quantiles and other sketch-based aggregates come in two flavours in every OLAP engine: a fast approximate form and a much slower exact form. An approximate distinct count keeps a small HyperLogLog sketch per group and is usually within one or two percent of the truth; the exact form holds every distinct value per group, so on large tables it can be ten to a hundred times slower and cannot be served from [rollups](/developers/build/metrics-view/rollups).

A measure can define both. `expression` stays the default and is treated as approximate, and `exact_expression` is used only when a viewer or a query explicitly asks for exact values.

```yaml
measures:
  - name: unique_users
    display_name: Unique Users
    expression: uniqCombined64(user_id)       # fast, approximate default
    exact_expression: uniqExact(user_id)      # used only when exact values are requested
    format_preset: humanize

  - name: p95_latency
    display_name: P95 Latency
    expression: quantileTDigest(0.95)(latency_ms)
    exact_expression: quantileExact(0.95)(latency_ms)
    format_preset: interval_ms

  - name: revenue_per_user
    display_name: Revenue per User
    requires: [total_revenue, unique_users]
    expression: total_revenue / unique_users   # uses unique_users as the query computes it
    format_preset: currency_usd
```

## How it behaves

- Both expressions share every other setting of the measure: `treat_nulls_as`, formatting, tags and security policies. Only the SQL differs. Setting `exact_expression` without `expression` is an error, and only simple measures can define it: a measure with `requires` or `per` (a derived measure), a measure with a `window`, and an ad hoc measure defined in a dashboard have no exact variant.
- Dashboards show the approximate value by default and mark measures that have an exact variant with a small `≈` chip. Clicking the chip computes that measure exactly for the current filters and time range, in the leaderboard, table or chart the chip sits in; see [Exact values in Explore](/guide/dashboards/explore/exact-measures).
- A derived measure such as `revenue_per_user` above is calculated from the other measures as the query computes them: in a table where `unique_users` is exact, the ratio uses the exact count. It shows no chip of its own. Precision is resolved once per query: a measure requested exactly anywhere in a query, including inside a measure filter, is computed exactly everywhere in it.
- Measures without an `exact_expression` are unchanged. Rill does not know whether a lone `expression` is approximate, so they show no chip.
- Queries that compute a measure exactly never use rollups, because exact aggregates cannot be derived from pre-aggregated rows.
- Rill validates `exact_expression` like `expression` when the metrics view reconciles, and the exact variant must produce a numeric type.

## Engine functions

| Engine | Approximate | Exact |
|--------|-------------|-------|
| ClickHouse | `uniq(x)`, `uniqCombined64(x)`, `quantileTDigest(p)(x)` | `uniqExact(x)`, `quantileExact(p)(x)` |
| DuckDB | `approx_count_distinct(x)`, `approx_quantile(x, p)` | `count(DISTINCT x)`, `quantile_cont(x, p)` |
| Druid | `APPROX_COUNT_DISTINCT_DS_HLL(x)` | `COUNT(DISTINCT x)` |
| Pinot | `DISTINCTCOUNTHLL(x)` | `DISTINCTCOUNT(x)` |

:::caution Druid
Druid treats `COUNT(DISTINCT x)` as approximate unless the connection sets the query context `useApproximateCountDistinct=false`, and in exact mode older Druid versions allow only one distinct count per query. A query that computes two measures exactly can fail; the dashboard shows the error on the measure's chip.
:::

## Requesting exact values from the API

Every measure in a [metrics API](/developers/build/metrics-view/metrics-sql) aggregation request accepts `exact: true`:

```json
{
  "metrics_view": "events",
  "measures": [
    { "name": "unique_users", "exact": true },
    { "name": "total_revenue" }
  ]
}
```

The flag is a no-op for measures without an `exact_expression`. The [AI tools](/guide/ai/mcp) use the same flag when a user asks for exact or reconciled numbers. Scheduled reports and alerts do not set it unless the stored query does.

## Disabling exact queries

Set `rill.metrics.exact_measures: false` in the `variables` of `rill.yaml` to make any query that would compute a measure with its `exact_expression` fail with an error, including queries that request it only inside a measure filter. Use it as a kill switch when exact aggregations put too much load on the database; approximate queries keep working, and so does a flag on a measure without an `exact_expression`.
