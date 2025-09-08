---
title: Time Series in Metrics Views
description: "Configure time-based dimensions and aggregations for comprehensive temporal analysis"
sidebar_label: Time Series
sidebar_position: 03
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


<!-- Valid until new time picker -->
It is not possible to set a limit for largest time grain and based on your selection of time range, certain grains will be unavailable to select. IE: minute or second grain at "Last 24 hours".
