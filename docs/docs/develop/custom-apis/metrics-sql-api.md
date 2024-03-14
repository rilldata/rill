---
title: Metrics SQL API
sidebar_label: Metrics SQL API
sidebar_position: 30
---

Metrics SQL API allows you to query data from your predefined [metrics view](../metrics-dashboard.md) using [Metrics SQL](./metrics-sql.md) dialect. 

Example:

```yaml
kind: api
metrics_sql: SELECT dimension, AGGREGATE(measure) FROM my_metrics GROUP BY dimension
```

where `my_metrics` is your metrics view name, `measure` is a metrics that you have defined in the metrics view and `dimension` is a dimension defined in the metrics view.

Read more the dialect here: [Metrics SQL](./metrics-sql.md).

## Templating

It supports templating the same way as [SQL API Templating](./sql-api.md#sql-templating).

