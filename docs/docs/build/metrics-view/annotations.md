---
title: Annotations
description: Enrich your metrics view with external context using annotations
sidebar_label: Annotations
sidebar_position: 20
---

Annotations allow you to enrich your metrics view with external data sources or context that isn't directly part of the primary model. This is often used to overlay events, holidays, or deployment markers onto your time series charts.

<img src = '/img/build/metrics-view/annotations.png' class='rounded-gif' />
<br />

## Requirements

The underlying table or model used for annotations must strictly follow this schema:

1.  **`time` column (Required)**: Used to position the annotation on the time series chart.
2.  **`description` column (Required)**: The text displayed when hovering over the annotation.
3.  **`time_end` column (Optional)**: If present, defines the end time for a range-based annotation.
4.  **`duration` column (Optional)**: If present, this is used to filter out annotations that are more granular than the currently selected dashboard time grain. It also forces `time` and `time_end` in the UI to be truncated to the selected grain.

## Configuring Annotations

To add an annotation, you need to define a reference to a `table` or `model` in your metrics view YAML. 

```yaml
annotations:
  - model: annotations_auction
    name: auction_annotations
    measures: ['requests']
```

:::info
Refer to the [`annotations` section](/reference/project-files/metrics-views#annotations) in Metrics View YAML reference for more details on how to implement annotations.
:::
