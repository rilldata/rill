---
title: Rollups
description: Accelerate metrics view queries by routing to pre-aggregated rollup tables
sidebar_label: Rollups
sidebar_position: 25
---

Rollups let a metrics view be backed by one or more pre-aggregated tables in addition to the base table. When a query's time grain, dimensions, measures, time range, and filters all match a rollup, Rill transparently rewrites the query to read from the rollup table instead of the base.

## Defining a Rollup

A rollup is defined as a separate table in the external olap engine or as a model that pre-aggregates the base model, plus a `rollups` entry in the metrics view YAML that points to it. The rollup model must produce columns with the same names as the base model's time dimension, the dimensions you list, and the measure expression inputs.

The simplest case — a daily rollup of an hourly fact table:

```yaml
# metrics_views/events.yaml
type: metrics_view
version: 1
model: events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: domain
    column: domain
  - name: country
    column: country
measures:
  - name: total_impressions
    expression: SUM("impressions")
  - name: total_clicks
    expression: SUM("clicks")
rollups:
  - model: events_daily
    time_grain: day
    dimensions: [publisher, domain]
    measures: [total_impressions, total_clicks]
```

Note that the rollup omits `country`. Queries that group by or filter on `country` will fall back to the base table; queries on `publisher` and/or `domain` at day grain or coarser will use the rollup.

### Multiple Rollups

You can define several rollups at different grains. Rill will pick the most efficient one that can answer a given query:

```yaml
rollups:
  - model: events_daily
    time_grain: day
    dimensions: [publisher, domain]
    measures: [total_impressions, total_clicks]
  - model: events_monthly
    time_grain: month
    dimensions: [publisher, domain]
    measures: [total_impressions, total_clicks]
```

A query at month grain over the full year will be served from `events_monthly`; a query at day grain over a single month will be served from `events_daily`.

### Field Selectors

`dimensions` and `measures` accept the standard field-selector forms — an explicit list, a wildcard, a regex, or an exclusion, if not defined then all dimensions and measures are included. For example, this rollup includes all dimensions and all measures except `total_clicks` measure:

```yaml
rollups:
  - model: events_daily
    time_grain: day
    measures:
      exclude: [total_clicks] # all measures except total_clicks
```

### Configuration Reference

- **`model`** (required) — The pre-aggregated table or model.
- **`time_grain`** (required) — Grain of the rollup. One of `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`.
- **`time_zone`** (optional) — IANA timezone the rollup was bucketed in (e.g. `America/New_York`). For day and coarser grains, queries are routed to the rollup only if their timezone matches.
- **`database`**, **`database_schema`** (optional) — Override the OLAP database and schema for the rollup table.
- **`dimensions`** (optional) — Field selector for which base-view dimensions are present in the rollup. Defaults to all.
- **`measures`** (optional) — Field selector for which base-view measures are present in the rollup. Defaults to all.

A metrics view must define a `timeseries` to use rollups. The full schema is documented in the [metrics view reference](/reference/project-files/metrics-views#rollups).

## How Rollup Selection Works

For each query, Rill walks through three phases: a quick disqualification, a per-rollup eligibility check, and a selection step among the eligible rollups.

### 1. Quick Disqualification

The whole rollup system is skipped — and the base table is used — when:

- The query asks for raw rows rather than aggregates.
- The query has a comparison time range. (Time comparison queries always read the base table.)
- The query's time range is on a time dimension that isn't the metrics view's primary `timeseries`.

### 2. Eligibility

A rollup is eligible for a given query only if **all** of the following hold:

1. **Grain derivable.** The query's time grain can be aggregated up from the rollup's grain. For example, a `month` query can be derived from a `day` rollup, but a `week` query cannot be derived from a `month` rollup, and a `month` query cannot be derived from a `week` rollup. Sub-day grains form one chain (`ms → s → min → hour → day`); calendar grains form another (`day → month → quarter → year`); `week` sits on its own branch and can only be derived from `day` or finer.
2. **Timezone matches** (day grain and coarser). The query's timezone must equal the rollup's `time_zone`. UTC variants (`""`, `"UTC"`, `"Etc/UTC"`) are treated as equivalent. Sub-day grains are timezone-agnostic, so this check is skipped for them.
3. **Start aligned.** The query's time-range start must fall exactly on a rollup-grain boundary in the rollup's timezone. A query starting at `2024-01-01 12:00` cannot use a daily rollup, because the noon boundary doesn't line up with a day bucket.
4. **All queried dimensions present.** Every dimension in the query (group-by, time floor, or WHERE filter) must be in the rollup's `dimensions` list. The primary time dimension is always considered present.
5. **All queried measures present.** Every measure named in the query must be in the rollup's `measures` list. Computed measures like `COUNT(*)` or `COUNT(DISTINCT …)` are rejected outright — they would produce wrong results when applied on top of pre-aggregated rows.

### 3. Time Coverage

For each eligible rollup, Rill checks that the rollup actually contains data for the requested range:

- **With a time range.** The query range is first clamped to the base table's `[min, max]` (so a query that extends past the base data isn't penalized for the rollup also stopping there). The rollup must then cover the clamped start and end.
- **Without a time range** ("all data"). The rollup must cover the base table's full `[min, max]`.
- **End alignment.** If the base table has data beyond the query's end time, the end must also be aligned to the rollup grain. Otherwise the last rollup bucket would pull in data from outside the requested range. Queries whose end falls past the latest base data don't need to be end-aligned, because there is no extra data to pull in.

### 4. Selection

Among rollups that pass eligibility and coverage:

1. Prefer the **coarsest grain** — fewer rows to scan.
2. On a tie, prefer the rollup with the **smallest data range** (tightest coverage).

The base table is used if no rollup is eligible.

## Limitations and Edge Cases

- **Time grain must be derivable.** `week` and the calendar grains (`month`, `quarter`, `year`) live on separate branches. A weekly rollup cannot answer monthly queries, and a monthly rollup cannot answer weekly queries. Define rollups at the grain you actually query.
- **Day-and-above rollups are timezone-specific.** A rollup bucketed in UTC cannot serve a dashboard query in `America/New_York`, because the day boundaries are different. If your users query in multiple timezones, either materialize a rollup per timezone or keep the rollup at hour grain (which is timezone-agnostic).
- **Misaligned starts disqualify the rollup.** A query starting mid-bucket (e.g. `2024-01-01 12:00` against a daily rollup) silently falls back to the base table. Dashboard time-range pickers typically snap to grain boundaries; ad-hoc API queries may not.
- **Computed measures fall back.** `count` and `count_distinct` measures bypass rollups even if the rollup looks otherwise suitable, because counting pre-aggregated rows is not the same as counting raw rows. Define an explicit `SUM(...)` measure on a pre-aggregated counter column in the rollup if you want this case to route.
- **Derived measures fall back.** A measure of type `derived` (one with `requires` or `per`) cannot match a rollup's measure list — only `simple` measures can. The base table is used.
- **Rollups require a `timeseries`.** Metrics views without a primary time dimension cannot define rollups.
- **Filters on missing dimensions disqualify the rollup.** A WHERE clause on `country` will skip a rollup that doesn't include `country`, even if the query's group-by columns are all in the rollup.
- **The rollup is responsible for being correct.** Rill does not validate that the rollup's measure values are consistent with the base — it trusts the model. If the rollup model uses the wrong aggregation (e.g. `AVG` where the base measure is `SUM`), queries routed to it will return wrong numbers.
- **Rollups are assumed to be roughly caught up with the base table.** Coverage is measured against the base table's latest timestamp. A rollup that lags behind the base will be silently skipped for any query that reaches the tail of the data — including common "last 24 hours" queries and queries without a time range — even if it has the right grain, dimensions, and measures. Refresh rollups in step with the base model so selection actually happens.

:::info
The full configuration schema is in the [metrics view reference](/reference/project-files/metrics-views#rollups).
:::
