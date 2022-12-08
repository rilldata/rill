---
title: Define metrics dashboard
description: Define your time dimension, measures and dimensions to create a dashboard
sidebar_label: Define metrics dashboard
sidebar_position: 30
---

In Rill, your dashboards are defined by _metrics_. Metrics are composed of:
* a _time dimension_, which will be a time stamp column from your data model. The time dimension will be the x-axis in the line charts shown on your dashboard, as well as serving as the grouping dimension over which _measure_ time grain aggregates are computed.

* _measures_ which are numerical aggregates of columns from your data model, and are shown on the y-axis of the line charts in your dashboard as well as the "big number" summaries next to the line charts. 

* _dimensions_ which are categorical columns from your data model. Dimensions are shown in _leaderboards_, which display the most frequently occurring values in each dimension, allow you to filter the data shown in your dashboard to only include data points that have the values you have selected from these leaderboards.

:::tip

To get you up and running quickly, Rill can generate a dashboard directly from a data source or a data model. These dashboards will be populated using:
* the first time stamp column from your data set as the time dimension
* the number of events per time period (based on the selected time dimension) as the default measure
* all available categorical columns as dimensions.

:::

## Editing dashboard metrics in the UI

Dashboards can be created and improved using the metrics editor. The metrics editor helps you define a time series, set of measures, and categorical dimensions that are directly tied to your dashboard.

### Time dimension

Your time dimension must be a column from your data model of type [`TIMESTAMP`](https://duckdb.org/docs/sql/data_types/timestamp), [`TIME`](https://duckdb.org/docs/sql/data_types/overview), or [`DATE`](https://duckdb.org/docs/sql/data_types/date).

:::tip

Strings representing dates are not supported, but you may be able to [`CAST`](https://duckdb.org/docs/sql/expressions/cast) such a string to one of these types while developing your data model.

:::

### Measures
Measures are numeric aggregates of columns from your data model, and power the line charts that you see in Rill.

A measure must be defined with a [DuckSQL](./sql-models.md) aggregation function over columns from your data model, or a mathematical expression built with one or more such aggregates.

For example, if you have a table of sales events with columns including a timestamp for the sales date, the sales price, and customer id, you could calculate the following metrics per time period with these expressions:
* number of sales: `COUNT(*)` (note that this would be equivalent to counting the total number of rows for any column, e.g. `COUNT(sales_date)`)
* total revenue: `SUM(sales_price)` 
* revenue per customer: `SUM(sales_price)/COUNT(DISTINCT customer_id)`

Any [DuckSQL numeric operator or function](https://duckdb.org/docs/sql/functions/numeric) is allowed in a measure expression, as are any of the following [DuckSQL aggregation expression](https://duckdb.org/docs/sql/aggregates): `AVG`, `COUNT`, `FAVG`,`FIRST`, `FSUM`, `LAST`, `MAX`, `MIN`, `PRODUCT`, `SUM`, `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, and `VAR_SAMP`.

You can also add labels, descriptions, and your choice of number formatting to your measures to customize how they are shown in the dashboard.

### Dimensions


Dimensions in Rill are used for filtering the data shown in your dashboard, and must come from "categorical" columns in your data model, which correspond to the following DuckDB data types: `BOOLEAN`, `BOOL`, `LOGICAL`, `BYTE_ARRAY`, `VARCHAR`, `CHAR`, `BPCHAR`, `TEXT`, and `STRING`.

You can also add labels and descriptions to your dimensions to customize how they are shown in the dashboard.

:::tip

Try creating categorical columns from numeric columns in your data model by using SQL [`CASE`](https://duckdb.org/docs/sql/expressions/case#:~:text=DuckDB%20%2D%20Case%20Statement&text=The%20CASE%20statement%20performs%20a,a%20%3A%20b%20) statements to convert numeric ranges into meaningful categories.

:::

## Editing dashboard metrics using code

In your Rill project directory, create a `dashboard_name.yaml` file in the `dashboards` directory and adapt its defintion from the following template:

```yaml
model: model_name
display_name: Dashboard name
description: 

timeseries: timestamp_column_name
default_timegrain: ""
timegrains:
  - day
  - month
  - year

dimensions:
  - property: model_column_1
    label: Column label 1
    description: ""
  - property: model_column_2
    label: Column label 2
    description: ""
  # Add more dimensions here

measures:
  - label: "Count"
    expression: count(*)
    description: ""
  # Add more measures here
```

Rill will ingest the dashboard definition next time you run `rill start`. For details about all available properties, see the metrics syntax [reference](../reference/metrics.md).
