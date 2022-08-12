---
title: Defining Metrics
---

In Rill, Your dashboards are powered by _metrics_. Metrics are composed of:
* a _time dimension_, which will be a time stamp column from your data model. The time dimension will be the x-axis in the line charts shown on your dashboard, as well as serving as the grouping dimension over which _measure_ aggregates are computed.

* _measures_ which are numerical aggregates of columns from your data model, and are shown on the y-axis of the line charts in your dashboard as well as the "big number" summaries next to the line charts. 

* _dimensions_ which are categorical columns from your data model. Dimensions are shown in _leaderboards_, and allow you to filter the data shown in your dashboard.

:::tip

To get you up and running quickly, Rill can automatically generate a set of metrics directly from a data source or a data model. These metrics will be populated using:
* the first time stamp column from your data set as the time dimension
* the number of events per time period (based on the selected time dimension) as the default measure
* all available categorical columns as dimensions.

:::

# Time Dimension

Your time dimension must be a column from your data model of type `TIMESTAMP`, `TIME`, `DATETIME`, or `DATE`.

:::tip

Strings representing dates are not supported, but you may be able to `CAST` such a string to one of these types while developing your data model.

:::

# Measures
measures are numeric aggregates of columns from your data model, and power the line charts that you see in Rill.

A measure must be defined with a [DuckSQL](./sqldialects/duck-sql.md) aggregation function over columns from your data model, or a mathematical expression built with one or more such aggregates.

For example, if you have a table of sales events with columns including a timestamp for the sales date, the sales price, and customer id, you could calculate the following metrics per time period with these expressions:
* number of sales: `COUNT(*)` (note that this would be equivalent to counting any column, e.g. `COUNT(sales_date)`)
* total revenue: `SUM(sales_price)` 
* revenue per customer: `SUM(sales_price)/COUNT(customer_id)`

Any [DuckSQL numeric operator or function](https://duckdb.org/docs/sql/functions/numeric) is allowed in a measure expression, as are any of the following [DuckSQL aggregation expression](https://duckdb.org/docs/sql/aggregates): `AVG`, `COUNT`, `FAVG`,`FIRST`, `FSUM`, `LAST`, `MAX`, `MIN`, `PRODUCT`, `SUM`, `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, and `VAR_SAMP`.

You can also add labels, descriptions, and your choice of number formatting to your measures to customize how they are shown in the dashboard.

# Dimensions


Dimensions in Rill are used for filtering the data shown in your dashboard, and must come from "categorical" columns in your data model, which correspond to the following DuckDB data types: `BOOLEAN`, `BOOL`, `LOGICAL`, `BYTE_ARRAY`, `VARCHAR`, `CHAR`, `BPCHAR`, `TEXT`, and `STRING`.

You can also add labels and descriptions to your dimensions to customize how they are shown in the dashboard.

:::tip

Try creating categorical columns from numeric columns in your data model by using SQL `CASE` statements to convert numeric ranges into meaningful categories.

:::

