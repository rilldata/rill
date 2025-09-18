---
title: Time Series in Metrics Views
description: "Configure time-based dimensions and aggregations for comprehensive temporal analysis"
sidebar_label: Time Series
sidebar_position: 07
---

Time is the most critical dimension in analytics and powers our dashboards. Understanding not just the "what," but how metrics evolve over hours, days, and months provides the narrative arc for decision-making.

## Defining Your Time Series Column

Your time series must be a column from your data model of type `TIMESTAMP`, `TIME`, or `DATE`. If your source has a date in a different format, you can apply time functions to transform these fields into valid time series types. The specific functions available depend on your OLAP engine.

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

model: example_model # Choose a table to underpin your metrics
timeseries: timestamp_column # Choose a timestamp column from your table
```

### Time Configuration Parameters

You can customize time behavior using the following parameters:

**`first_day_of_week`**
Specifies which day should be considered the start of the week. Valid values are 1 through 7, where Monday=1 and Sunday=7.

**`first_month_of_year`**
Determines which month should be treated as the beginning of the year. Valid values are 1 through 12, where January=1 and December=12.

```yaml
timeseries: transaction_date
first_month_of_year: 7  # July start
first_day_of_week: 2    # Monday start
```

These parameters enable you to define non-standard reporting periods. For example, if you set June as the starting month, "past year" calculations will span from June to May instead of January to December.

:::tip Full YAML Configurations

Please refer to our [reference page](/reference/project-files/metrics-views) for all the available parameters to define in a metrics view.
:::

### Time Grain Configuration

The `smallest_time_grain` parameter controls the minimum temporal resolution available in your dashboards. Limiting granularity provides several benefits:

- **Performance**: Reduces query complexity and improves dashboard responsiveness
- **Consistency**: Ensures all users see data at the same level of detail
- **Focus**: Prevents analysis paralysis from overly granular data


```yaml
timeseries: order_timestamp
smallest_time_grain: day
```


:::note Time Grain Availability

The time picker automatically adjusts available time grains based on your selected time range. For example, when viewing "Last 24 hours," only day, and hour grains are available, while "Last 30 days" offers day, week, and month grains. This ensures meaningful time-based analysis appropriate to your data range.

:::

### `watermark`

The `watermark` parameter defines the data freshness threshold for your metrics view. It determines the latest point in time where data is considered "complete" and reliable for analysis.

**Purpose:**
- Prevents analysis of incomplete or partial data
- Ensures consistent reporting across different time zones
- Provides a clear boundary between complete and incomplete data

**Configuration:**
```yaml
timeseries: event_time
watermark: "MAX(__TIME) - INTERVAL 3 DAYS"
```

**How it works:**
- `MAX(__TIME)` gets the latest timestamp in your data
- `INTERVAL 3 DAYS` subtracts 3 days from that timestamp
- If your latest data is September 5th, complete data extends only to September 2nd
- Queries for September 3rd-5th data will return empty or incomplete results
- This prevents misleading metrics from partial data

**Common watermark expressions:**
- `"MAX(__TIME) - INTERVAL 1 DAY"` - For daily batch processing (most common)
- `"MAX(__TIME) - INTERVAL 1 HOUR"` - For real-time data with hourly completeness
- `"MAX(__TIME) - INTERVAL 1 WEEK"` - For weekly aggregated data
- `"MAX(__TIME)"` - No watermark (use with caution)

:::tip Best practices

Set your watermark based on your data pipeline's processing time. If your ETL takes 2 hours to complete daily data, set watermark to "2 hours" or "1 day" to ensure you only analyze complete datasets.

:::