---
title: "Quantiles"
description: Tips & Tricks for Measure Quantiles
sidebar_label: "Quantiles"
sidebar_position: 30
---

In addition to common aggregates, you may wish to look at the value of a measure within a certain band or quantile. In the example below, we can measure the P95 of a given measure using `QUANTILE_CONT`.

<img src = '/img/build/metrics-view/examples/percentile-visual.png' class='rounded-gif' />
<br />


Using [DuckDB aggregate function](https://duckdb.org/docs/stable/sql/functions/aggregates.html#quantile_contx-pos), you can easily calculate various quantiles.

:::tip Not on DuckDB?
If you are using a different OLAP engine to power your dashboard, simply use the correct function for quantile calculation. 

E.g.: [ClickHouse quantile](https://clickhouse.com/docs/sql-reference/aggregate-functions/reference/quantile), [Pinot percentile](https://docs.pinot.apache.org/configuration-reference/functions/percentile)
:::
Please review the reference documentation, [here.](/reference/project-files/metrics-views)

## Examples

<img src = '/img/build/metrics-view/examples/percentile-example.png' class='rounded-gif' />
<br />

In this example we see the values of P95 and P99 are calculated using the following expressions:

```yaml
  - name: p95_quantile_global_intensity
    expression: QUANTILE_CONT(Global_intensity, 0.95)
    format_d3: ".3f"
    description: P95 of Global Intensity
  - name: p99_quantile_global_intensity
    expression: QUANTILE_CONT(Global_intensity, 0.99)
    format_d3: ".4f"
    description: P99 of Global Intensity
```

## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/rill-kaggle-elec-consumption/explore/household_power_consumption_metrics_explore)