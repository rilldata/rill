---
title: "Build Dashboards"
description: Define your time dimension, measures and dimensions to create a dashboard
sidebar_label: "Build Dashboards"
sidebar_position: 10
---

In Rill, your dashboards are defined by _metric definitions_. Metric definitions are composed of:
* _**A model**_ - A data model creating a One Big Table that will power the dashboard.
* _**A timeseries**_ - A column from your model that will underlie x-axis data in the line charts. Time will be truncated into different time periods.
* _**Measures**_ - Numerical aggregates of columns from your data model shown on the y-axis of the line charts and the "big number" summaries.
* _**Dimensions**_ - Categorical columns from your data model whose values are shown in _leaderboards_ and allow you to look at segments and filter the data.

:::tip Dashboard Properties
For more details about available configurations and properties, check our [Dashboard YAML](../reference/project-files/dashboards) reference page.
:::

## Creating valid metrics

### Timeseries

Your timeseries must be a column from your data model of [type](https://duckdb.org/docs/sql/data_types/timestamp) `TIMESTAMP`, `TIME`, or `DATE`. If your source has a date in a different format, you can apply [time functions](https://duckdb.org/docs/sql/functions/timestamp) to transform these fields into valid timeseries types.

### Measures

Measures are numeric aggregates of columns from your data model. A measure must be defined with [DuckDB SQL](https://duckdb.org/docs/sql/introduction.html) aggregation functions and expressions on columns from your data model. The following operators and functions are allowed in measure expressions:

* Any DuckDB SQL [numeric](https://duckdb.org/docs/sql/functions/numeric) operators and functions
* This set of DuckDB SQL [aggregates](https://duckdb.org/docs/sql/aggregates): `AVG`, `COUNT`, `FAVG`,`FIRST`, `FSUM`, `LAST`, `MAX`, `MIN`, `PRODUCT`, `SUM`, `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, `VAR_SAMP`.
* [Filtered aggregates](https://duckdb.org/docs/sql/query_syntax/filter.html) can be used to filter set of rows fed to the aggregate functions

As an example, if you have a table of sales events with the sales price and customer id, you could calculate the following metrics with these aggregates and expressions:
* Number of sales: `COUNT(*)`
* Total revenue: `SUM(sales_price)` 
* Revenue per customer: `CAST(SUM(sales_price) AS FLOAT)/CAST(COUNT(DISTINCT customer_id) AS FLOAT)`
* Number of orders with order value more than $100 : `count(*) FILTER (WHERE order_val > 100)`

You can also add labels, descriptions, and your choice of number formatting to customize how they are shown in the dashboard.


### Dimensions

Dimensions are used for exploring segments and filtering the dashboard. Valid dimensions can be any type and are selected using the drop down menu. You can also add labels and descriptions to your dimensions to customize how they are shown in the dashboard.


## Updating dashboards

### Using the UI / code
When you add a metrics definition using the UI, a code definition will automatically be created as a .yaml file in your Rill project in the dashboards directory. However, you can also create metrics definitions more directly by creating the artifact.

In your Rill project directory, a `dashboard_name.yaml` file is created in the `dashboards` directory and its definition its definition can be adapted from the following template:

```yaml
model: model_name
title: Dashboard name
default_time_range: ""
smallest_time_grain: ""
timeseries: timestamp_column_name

dimensions:
  - column: model_column_1
    label: Column label 1
    description: ""
  # Add more dimensions here

measures:
  - label: "Count"
    name: "count"
    expression: count(*)
    description: ""
    format_preset: "humanize"
  # Add more measures here
```

For details about all available properties, see the syntax [reference](../reference/project-files/dashboards).

