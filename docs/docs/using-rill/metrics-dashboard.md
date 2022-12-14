---
title: Define metrics dashboard
description: Define your time dimension, measures and dimensions to create a dashboard
sidebar_label: Define metrics dashboard
sidebar_position: 30
---

In Rill, your dashboards are defined by _metrics_. Metrics are composed of:
* _**a model**_ - A [data model](./sql-models.md) creates One Big Table that will power the dashboard.
* _**a timeseries**_ - A column from your model that will underlie x-axis data in the line charts. Time will be truncated into different time perioids
* _**measures**_ - Numerical aggregates of columns from your data model shown on the y-axis of the line charts and the "big number" summaries.
* _**dimensions**_ - Categorical columns from your data model whose values are shown in _leaderboards_ and allow you to look at segments and filter the data.


## Creating valid metrics

### Timeseries

Your timeseries must be a column from your data model of [type](https://duckdb.org/docs/sql/data_types/timestamp) `TIMESTAMP`, `TIME`, or `DATE`. If your source has a date in a different format, you can apply [time functions](https://duckdb.org/docs/sql/functions/timestamp) to transform these fields into valid timeseries types.


### Measures

Measures are numeric aggregates of columns from your data model. A measure must be defined with [DuckSQL](./sql-models.md) aggregation functions and expressions on columns from your data model. The following operators and function are allowed in a measure expression:

* any DuckSQL [numeric](https://duckdb.org/docs/sql/functions/numeric) operators and functions
* this set of  DuckSQL [aggregates](https://duckdb.org/docs/sql/aggregates): `AVG`, `COUNT`, `FAVG`,`FIRST`, `FSUM`, `LAST`, `MAX`, `MIN`, `PRODUCT`, `SUM`, `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, `VAR_SAMP`.

As an example, if you have a table of sales events with the sales price and customer id, you could calculate the following metrics with these aggregates and expressions:
* number of sales: `COUNT(*)`
* total revenue: `SUM(sales_price)` 
* revenue per customer: `CAST(SUM(sales_price) AS FLOAT)/CAST(COUNT(DISTINCT customer_id) AS FLOAT)`

You can also add labels, descriptions, and your choice of number formatting to customize how they are shown in the dashboard.


### Dimensions

Dimensions are used for exploring segments and filtering the dashboard. Valid dimensions must be "categorical" columns of type `BOOLEAN`, `BOOL`, `LOGICAL`, `BYTE_ARRAY`, `VARCHAR`, `CHAR`, `BPCHAR`, `TEXT`, or `STRING`. If you have a column that is not formatted correctly to be a dimension, try creating categorical columns in the data model using SQL [`CASE`](https://duckdb.org/docs/sql/expressions/case#:~:text=DuckDB%20%2D%20Case%20Statement&text=The%20CASE%20statement%20performs%20a,a%20%3A%20b%20) statements.

You can also add labels and descriptions to your dimensions to customize how they are shown in the dashboard.


## Using the UI

Dashboards can be created and improved on using the metrics editor. The metrics editor helps you define a model, a timeseries, set of measures, and categorical dimensions that are directly tied to your dashboard. 

To create a new dashboard from scratch, click "+" by Dashboards in the left hand navigation pane to open the metrics editor.

In addition, you can quickly generate a dashboard with opinionated defaults using the "Create Dashboard" button in the upper right hand corner of the source or model views or "Quick Start" button in the metrics editor itself. These dashboards will be populated using:

- the first timestamp column from your model set as the timeseries
- the number of records as the default measure (`COUNT(*)`)
- all available categorical columns as dimensions

If you want to revisit these configurations, you can open the metrics editor by clicking on the "Edit Metrics" button in the upper right hand corner of the dashboard view.


## Using the CLI

We do not currently support adding or editing metrics from the CLI.

## Using code
When you add a metrics definition using the UI, a code definition will automatically be created as a .yaml file in your Rill project in the dashboards directory. However, you can also create metrics definitions more directly by creating the artifact.

In your Rill project directory, create a `dashboard_name.yaml` file in the `dashboards` directory and adapt its definition from the following template:

```yaml
model: model_name
display_name: Dashboard name

timeseries: timestamp_column_name

dimensions:
  - property: model_column_1
    label: Column label 1
    description: ""
  # Add more dimensions here

measures:
  - label: "Count"
    expression: count(*)
    description: ""
    format_preset: "humanize"
  # Add more measures here
```

Rill will ingest the dashboard definition next time you run `rill start`. For details about all available properties, see the syntax [reference](../references/project-files.md#dashboard-metrics).

