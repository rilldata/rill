---
title: Customization
description: Customize Metrics views
sidebar_label: Customization
sidebar_position: 20
---

Rill provides several top-level configuration options to customize how your metrics view behaves and appears. These properties help define the context, calendar logic, and data boundaries of your metrics.

## Display Name & Description

You can provide a human-readable name and description for your metrics view. These are displayed in the UI to help users understand the purpose and content of the metrics.

```yaml
display_name: "Revenue Metrics View"
description: "Daily revenue metrics broken down by product and region"
```

## Time Logic

### Smallest Time Grain

The `smallest_time_grain` property sets the minimum granularity available for time-based analysis. This is useful when your data is captured at a high frequency (e.g., milliseconds) but business analysis only makes sense at a coarser level (e.g., hours or days).

Allowed values: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`.

```yaml
smallest_time_grain: "day"
```

### First Day of Week

You can customize the start of the week for your analysis. This affects how weeks are grouped and displayed in time series charts and tables. The value is an integer from 1 (Monday) to 7 (Sunday).

```yaml
# Set Monday as the first day of the week (Default: 1)
first_day_of_week: 1
```

### First Month of Year

For businesses that operate on a fiscal year different from the calendar year, you can set the `first_month_of_year`. This affects how "Year to Date" calculations and year groupings are handled. The value is an integer from 1 (January) to 12 (December).

```yaml
# Set February as the start of the fiscal year
first_month_of_year: 2
```

## Data freshness

### Watermark

The `watermark` field allows you to define the maximum valid timestamp for your data. This is particularly useful for filtering out incomplete data periods (e.g., the current partially complete day) or for aligning your dashboard's "current" time to the last successful ETL load.

The value should be a SQL expression that returns a timestamp.

```yaml
# Set the effective "end" of the data to midnight of the current day
watermark: "date_trunc('day', now())"
```

```yaml
# Set watermark to the maximum timestamp found in the table
watermark: "max(timestamp_col)"
```

## AI Instructions

The `ai_instructions` field allows you to provide context and guidance to AI tools that interact with your metrics view (like Rill's AI Chat or external agents). This helps ensure that AI-generated answers are accurate and aligned with your specific business logic.

```yaml
ai_instructions: "This metrics view tracks recurring revenue. When asked about 'churn', always refer to the 'churn_rate' measure, not just raw cancellations."
```

## Annotations

Annotations allow you to enrich your metrics view with external data sources or context that isn't directly part of the primary model. This is often used to overlay events, holidays, or deployment markers onto your time series charts.

Each annotation requires a reference to a `table` or `model` and a `measures` mapping to define what data should be visualized.

```yaml
annotations:
  - name: "holidays"
    model: "public_holidays"
    measures:
      holiday_name: "name"
```
