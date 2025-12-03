---
title: Annotations
description: Enrich your metrics view with external context using annotations
sidebar_label: Annotations
sidebar_position: 20
---

Annotations allow you to enrich your metrics view with external data sources or context that isn't directly part of the primary model. This is often used to overlay events, holidays, or deployment markers onto your time series charts.

Each annotation requires a reference to a `table` or `model` and a `measures` mapping to define what data should be visualized.

```yaml
annotations:
  - model: annotations_auction
    name: auction_annotations
    measures: ['requests']
```
<img src = '/img/build/metrics-view/annotations.png' class='rounded-gif' />
<br />

:::info
Refer to the [`annotations` section](/reference/project-files/metrics-views#annotations) in Metrics View YAML reference for more details on how to implement annotations.
:::

