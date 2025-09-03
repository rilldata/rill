---
title: Time Series in Metrics Views
description: "Configure time-based dimensions and aggregations for comprehensive temporal analysis"
sidebar_label: Time Series
sidebar_position: 03
---

Time is the most critical dimension in analytics and powers our dashboards. Understanding not just the "what," but how metrics evolve over hours, days, and months provides the narrative arc for decision-making.

## What is a Time Series?

A **time series** is a sequence of data points recorded at regular or irregular intervals, always tied to a timestamp. 
- **Daily** revenue trends
- **Hourly** error counts
- **Weekly** active users

Without a well-defined time series, analyses can drift, losing alignment across dimensions.

## Defining Your Time Series Column

Your time series must be a column from your data model of [type](https://duckdb.org/docs/sql/data_types/timestamp) `TIMESTAMP`, `TIME`, or `DATE`. If your source has a date in a different format, you can apply [time functions](https://duckdb.org/docs/sql/functions/timestamp) to transform these fields into valid time series types.

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

model: example_model # Choose a table to underpin your metrics
timeseries: timestamp_column # Choose a timestamp column from your table
```

### Additional Time Series Columns 

Along with your primary time series column, you can create secondary or terciarcy, etc. columns to swap your dashboards view. This is useful when you have different definitions of when your business starts its fiscal year but also want to be able to view a yearly blah blah blah.

```yaml
dimnensions:
   - name: secondary_time_dimension
     column: time_series_2
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

## Time Range Configuration

While Rill supports common time ranges such as `P1D`, `P3M`, etc. we realize that this isn't always the default time ranges that fit your business needs. Instead, we allow developers amd consumers the ability to create their own time ranges for their exact use case.

### Defining Default Time Ranges

The simplest syntax looks something like this:
```
-7d to -3d/D
7D as of now/D+1D
```

To create any time range, you'll need to gather the following requirements:
- two points in time to measure `-7d to -3d/`, or a time range from today `7D`
- how to measure your data: wall clock now, latest data, or "complete data"  `as of ...`
- the grain in which you want to see your data rolled up `/D`
- whether to view completed dates only, or partial dates (only makes sense for dates comapared to today) `+1D`

Having this information, you can customize your time range and create default time ranges for your users to select. To define a time range in dashboards, see our [explore dashboard build section](/build/dashboards/customize#time-ranges), and [canvas dashboard build section](/build/canvas/customization#time-ranges)


See our guide for a [comprehensive explanation on our time series syntax](/guides/time-series-syntax).

## Time Series Transformation

While metrics views handle time configuration, you can also transform time data at the [model level](/build/models) using SQL functions. Since both metrics views and models are powered by the underlying [OLAP engine](/connect/olap), similar functions can be applied to modify your time series dynamically.

:::note Query-time vs Model processing

There are benefits to pre-procesing the data in the model layer but for some quick processing this can be done in the metrics view.

**Query-time processing** (in metrics views):
- Flexible and dynamic
- No storage overhead
- Slightly slower for complex calculations

**Model-level processing** (in SQL models):
- Pre-computed and optimized
- Faster query performance
- Requires model refresh for updates

:::

### DuckDB Time Functions

DuckDB provides a comprehensive toolkit for temporal data manipulation:

- **`DATE_TRUNC`**: Normalize timestamps to consistent intervals (day, week, month, quarter, year)
- **`EXTRACT`**: Extract specific time components (year, quarter, month, day of week, hour)
- **`LAG/LEAD`**: Reference prior or future rows for period-over-period comparisons
- **`DATE_ADD/DATE_SUB`**: Perform date arithmetic for dynamic time ranges
- **`STRFTME`**: Extract strings from a time column.

<!-- 
```yaml
  - name: day_of_week
    expression: STRFTIME(last_login, '%A')
``` -->

For comprehensive documentation on all available time functions, see the [DuckDB time functions documentation](https://duckdb.org/docs/stable/sql/functions/timestamp.html).

### Time Aggregation (Roll-ups)

Roll-ups aggregate granular events into coarser intervals. For example, if your data arrives hourly but daily analysis suffices:

```sql
SELECT DATE_TRUNC('day', timestamp_column) AS time_series_column,
        ...  
FROM your_model
```

**Benefits of roll-ups:**
- **Storage efficiency**: Reduces data volume while preserving analytical value
- **Query performance**: Faster aggregations on pre-computed time buckets



<!-- ### Example

Heatmaps
- **Day of Week + Hour of Day**: Perfect for heat maps showing activity patterns

<img src='/img/build/metrics-view/heatmap.png' class='rounded-gif' />
<br />

Example of creating a day-of-week and hour-of-day dimension for heat maps:

```sql
SELECT 
    EXTRACT(dow FROM timestamp_column) AS day_of_week,
    EXTRACT(hour FROM timestamp_column) AS hour_of_day,
    ...
FROM your_model
``` 

```yaml
dimensions:
    - name: day_of_week
      expression: EXTRACT(dow FROM timestamp_column)
    - name: hour_of_day
      expression: EXTRACT(hour FROM timestamp_column)
```
 -->