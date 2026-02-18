---
title: Rill ISO 8601 Extensions
description: Legacy rill- prefixed time syntax extensions
sidebar_label: Rill ISO Extensions
sidebar_position: 10
---

Rill extends the ISO 8601 standard with special prefixed keywords for common time ranges. These are primarily used for backward compatibility and in specific contexts like comparison ranges.

:::tip New Syntax Available
For new configurations, we recommend using the modern [Time Range Syntax](/reference/time-syntax) which provides more flexibility and expressiveness.
:::

## Time Range Extensions

These extensions specify time ranges relative to a reference point. The reference point varies by context:
- **Dashboards**: Uses `latest` (most recent data timestamp)
- **Alerts**: Uses `watermark` (data completeness marker)

| Rill Extension | Description | Equivalent Modern Syntax |
|----------------|-------------|--------------------------|
| `inf` | All time | `earliest to latest` |
| `rill-TD` | Today | `ref/D to ref/D+1D as of watermark` |
| `rill-WTD` | Week to Date | `ref/W to ref/D+1D as of watermark` |
| `rill-MTD` | Month to Date | `ref/M to ref/D+1D as of watermark` |
| `rill-QTD` | Quarter to Date | `ref/Q to ref/D+1D as of watermark` |
| `rill-YTD` | Year to Date | `ref/Y to ref/D+1D as of watermark` |
| `rill-PDC` | Yesterday (Previous Day Complete) | `-1D/D to ref/D as of watermark` |
| `rill-PWC` | Previous Week Complete | `-1W/W to ref/W as of watermark` |
| `rill-PMC` | Previous Month Complete | `-1M/M to ref/M as of watermark` |
| `rill-PQC` | Previous Quarter Complete | `-1Q/Q to ref/Q as of watermark` |
| `rill-PYC` | Previous Year Complete | `-1Y/Y to ref/Y as of watermark` |

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

The modern syntax provides equivalent functionality with more flexibility. Notably, the modern `DTD` syntax supports intraday ranges (e.g., `ref/D to ref/h+1h`) while `rill-TD` cannot.

| Legacy | Modern Equivalent |
|--------|-------------------|
| `rill-TD` | `DTD as of watermark/D+1D` |
| `rill-WTD` | `WTD as of watermark/D+1D` |
| `rill-MTD` | `MTD as of watermark/D+1D` |
| `rill-PDC` | `1D as of watermark/D` |
| `rill-PWC` | `1W as of watermark/W` |
| `P7D` | `7D` |
| `P1M` | `1M` or `30D` |

See [Time Range Syntax](/reference/time-syntax) for the complete modern syntax reference.
