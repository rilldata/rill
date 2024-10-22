---
title: Create Metrics Views
description: Create metrics-view using source data and models with time, dimensions, and measures
sidebar_label: Create Metrics Views
sidebar_position: 00
---

<img src = '/img/build/metrics-view/visual-metrics-editor.gif' class='rounded-gif' />
<br />

In Rill, your metrics view is defined by _metric definitions_. Metric definitions are composed of:
* _**model**_ - A data model creating a One Big Table that will power the metrics view.
* _**timeseries**_ - A column from your model that will underlie x-axis data in the line charts. Time will be truncated into different time periods.
* _**measures**_ - Numerical aggregates of columns from your data model shown on the y-axis of the line charts and the "big number" summaries.
* _**dimensions**_ - Categorical columns from your data model whose values are shown in _leaderboards_ and allows you to look at segments or attributes of your data (and filter / slice accordingly).


:::tip
Starting in version 0.50, metrics view has been separated from dashboard. This allows for a cleaner, more accessible metrics layer and the ability to build various dashboards and components on top of a single metrics layer. For more information on why we decided to do this, please refer to the following: [Why separate the dashboard and metrics layer](/concepts/metrics-layer)

For migration steps, see [Migrations](/manage/migration#v049---v050).
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

You can also add labels, descriptions, and your choice of number formatting to customize how they are shown.


### Dimensions

Dimensions are used for exploring segments and filtering. Valid dimensions can be any type and are selected using the drop down menu. You can also add labels and descriptions to your dimensions to customize how they are shown.


## Updating the Metrics View

Whether you prefer the UI or YAML artifacts, Rill supports both methods for updating your metrics view.

### Using the Visual Metrics Editor

![visual-metric-editor](/img/build/metrics-view/visual-metrics-editor.png)

When you add a metrics definition using the UI, a code definition will automatically be created as a YAML file in your Rill project within the metrics directory by default. 

### Using YAML
You can also create metrics definitions more directly by creating the artifact.

In your Rill project directory, after the `metrics-view.yaml` file is created in the `metrics` directory, its configuration or definition can be updated as needed by updating the YAML file directly, using the following template as an example:

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: example_table # Choose a table to underpin your metrics
timeseries: timestamp_column # Choose a timestamp column (if any) from your table

dimensions:
  - column: category
    label: "Category"
    description: "Description of the dimension"

measures:
  - expression: "SUM(revenue)"
    label: "Total Revenue"
    description: "Total revenue generated"

```
:::info Check our reference documentation

For more information about available metrics view properties, feel free to check our [reference YAML documentation](/reference/project-files/metrics-view.

:::


### Multi-editor and external IDE support

Rill Developer is meant to be developer friendly and has been built around the idea of keystroke-by-keystroke feedback when modeling your data, allowing live interactivity and a real-time feedback loop to iterate quickly (or make adjustments as necessary) with your models and dashboards. Additionally, Rill Developer has support for the concept of "hot reloading", which means that you can keep two windows of Rill open at the same time and/or use a preferred editor of choice, such as VSCode, side-by-side with the dashboard that you're actively developing!

![hot-reload-0-36](https://cdn.rilldata.com/docs/release-notes/36_hot_reload.gif)