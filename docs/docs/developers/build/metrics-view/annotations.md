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
  3. **`time_end` column (Optional)**: Timestamp - If present, defines the end time for a range-based annotation.
  4. **`duration` column (Optional)**: String/Enum - If present, defines the granularity (e.g., 'day', 'hour', 'minute').selected dashboard time grain. It also forces `time` and `time_end` in the UI to be truncated to the selected grain.

## Configuring Annotations

To add an annotation, you need to define a reference to a `table` or `model` in your metrics view YAML.

```yaml
annotations:
  - model: annotations_auction
    name: auction_annotations
    measures: ['requests']
```

### Configuration Properties

- **`model`** or **`table`**: Reference to the data source containing the annotation data.
- **`name`**: A unique identifier for the annotation set.
- **`measures`** (optional): A list of measures to display these annotations alongside. If not specified, the annotation will appear for all measures.

### Visual Appearance

Annotations appear as markers or ranges on the time series charts in your dashboard.
- **Point Annotations**: Events with a single `time` timestamp appear as point markers.
- **Range Annotations**: Events with both `time` and `time_end` timestamps appear as shaded regions spanning the duration.
- **Hover Details**: Hovering over an annotation marker reveals the text from the `description` column.

:::info
Refer to the [`annotations` section](/reference/project-files/metrics-views#annotations) in Metrics View YAML reference for more details on how to implement annotations.
:::
