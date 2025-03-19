---
title: "Quantiles"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Quantiles"
sidebar_position: 03
---

### Quantiles

In addition to common aggregates, you may wish to look at the value of a metric within a certain band or quantile. In the example below, we can measure the P95 query time as a benchmark.

```yaml
  - label: "P95 Query time"
    expression: QUANTILE_CONT(query_time, 0.95)
    format_preset: interval_ms
    description: "P95 time (in sec) of query time"
```