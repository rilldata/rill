---
title: Rill ISO 8601 Extensions
description: Legacy rill- prefixed time syntax extensions
sidebar_label: Rill ISO Extensions
sidebar_position: 10
---

Rill supports a set of legacy `rill-` prefixed keywords for common time ranges. These are retained for backward compatibility with existing configurations.

:::tip New Syntax Available
For new configurations, use the modern [Time Range Syntax](/reference/time-syntax), which is more expressive and consistent across all contexts.
:::

## Time Range Extensions

| Rill Extension | Description |
|----------------|-------------|
| `inf` | All time |
| `rill-TD` | Today |
| `rill-WTD` | Week to Date |
| `rill-MTD` | Month to Date |
| `rill-QTD` | Quarter to Date |
| `rill-YTD` | Year to Date |
| `rill-PDC` | Yesterday (Previous Day Complete) |
| `rill-PWC` | Previous Week Complete |
| `rill-PMC` | Previous Month Complete |
| `rill-PQC` | Previous Quarter Complete |
| `rill-PYC` | Previous Year Complete |

:::note Reference point behavior
In a dashboard context, the reference point for these expressions is `latest` (most recent data timestamp). In alert contexts, the reference point is `watermark` (data completeness marker).
:::

## Time Comparison Extensions

These extensions are used specifically in comparison contexts (the "Comparing" feature in dashboards).

| Rill Extension | Description | Usage |
|----------------|-------------|-------|
| `rill-PP` | Previous Period | Compares against the immediately preceding period of same duration |
| `rill-PD` | Previous Day | Compares against the same time yesterday |
| `rill-PW` | Previous Week | Compares against the same time last week |
| `rill-PM` | Previous Month | Compares against the same time last month |
| `rill-PQ` | Previous Quarter | Compares against the same time last quarter |
| `rill-PY` | Previous Year | Compares against the same time last year |

## Usage Context

### As Time Range
Extensions ending in `TD` (to-date) or `C` (complete) are valid as primary time ranges:

```yaml
# In metrics view or explore configuration
default_time_range: "rill-MTD"  # Month to date
```

### As Comparison
Extensions starting with `rill-P` (previous) are typically used for comparisons:

```yaml
# In explore configuration
default_comparison:
  dimension: ""  # No dimension comparison
  mode: time
```

Then select "Previous Period", "Previous Day", etc. in the dashboard UI.

## ISO 8601 Duration Support

Rill also supports standard ISO 8601 duration format:

| Format | Description | Example |
|--------|-------------|---------|
| `P<n>Y` | n years | `P1Y` = 1 year |
| `P<n>M` | n months | `P6M` = 6 months |
| `P<n>W` | n weeks | `P2W` = 2 weeks |
| `P<n>D` | n days | `P7D` = 7 days |
| `PT<n>H` | n hours | `PT24H` = 24 hours |
| `PT<n>M` | n minutes | `PT30M` = 30 minutes |
| `PT<n>S` | n seconds | `PT60S` = 60 seconds |

Combined durations:
- `P1Y6M` = 1 year and 6 months
- `P1DT12H` = 1 day and 12 hours
- `PT1H30M` = 1 hour and 30 minutes

## Migration to Modern Syntax

The modern syntax provides equivalent functionality with more flexibility. One important distinction: `DTD` supports intraday ranges (e.g., `ref/D to ref/h+1h`) while `rill-TD` does not.

| Legacy | Modern Equivalent |
|--------|-------------------|
| `rill-TD` | `DTD` |
| `rill-WTD` | `WTD` |
| `rill-MTD` | `MTD` |
| `rill-QTD` | `QTD` |
| `rill-YTD` | `YTD` |
| `rill-PDC` | `1D as of watermark/D` |
| `rill-PWC` | `1W as of watermark/W` |
| `rill-PMC` | `1M as of watermark/M` |
| `rill-PQC` | `1Q as of watermark/Q` |
| `rill-PYC` | `1Y as of watermark/Y` |
| `P7D` | `7D` |
| `P1M` | `1M` or `30D` |

See [Time Range Syntax](/reference/time-syntax) for the complete modern syntax reference.
