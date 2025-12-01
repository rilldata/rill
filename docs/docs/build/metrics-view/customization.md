---
title: Customization 
description: Configure display properties and metadata for your metrics views
sidebar_label: Customization 
sidebar_position: 20
---

Rill provides several top-level configuration options to customize how your metrics view is presented to users and enriched with external context.

## Display Name & Description

You can provide a human-readable name and description for your metrics view. These are displayed in the UI to help users understand the purpose and content of the metrics.

```yaml
display_name: "Revenue Metrics View"
description: "Daily revenue metrics broken down by product and region"
```

## Annotations

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
## Time Configuration

For time-related settings such as `smallest_time_grain`, `first_day_of_week`, `first_month_of_year`, and `watermark`, please refer to our dedicated [Time Series Configuration](/build/metrics-view/time-series) guide.

## AI Configuration

You can provide context and instructions for AI tools interacting with your metrics view using the `ai_instructions` field. This is useful for clarifying specific metrics, dimensions, or data quirks that apply only to this specific view. For project-wide instructions, see the [Project Configuration](/build/project-configuration#ai-configuration) guide.

```yaml
ai_instructions: |
  # Metric Definitions
  - "Churn Rate" excludes trial users who cancelled within 7 days.
  - "Active Users" are defined as users with at least one login in the selected period.

  # Data Context
  - Data for the "Legacy Plan" is static and will not update after Dec 2023.
  - When analyzing "Revenue", always breakdown by "Region" to see currency impacts.
```

