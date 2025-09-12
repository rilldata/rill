---
title: Get Started with Metrics Views
description: Create metrics view using source data and models with time, dimensions, and measures
sidebar_label: What are Metrics Views?
sidebar_position: 00
---


A metrics view is a 'centralized framework' used to define and organize **key measures and dimensions** for your organization. Having a centralized layer allows an organization to easily manage and reuse calculations across various reports, dashboards, and data tools. Each metrics view is powered by a single [model or table](/build/metrics-view/underlying-model).


<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" width="100%" style={{ borderRadius: '15px', padding: '20px' }} />
</div>


In Rill, your metrics view is defined by _metric definitions_. Metric definitions are composed of:
* _**model/table**_ - A data model or underlying table created with the concept of [One Big Table](/build/models/models-101#one-big-table-and-dashboarding) that will power the metrics view.
* _**timeseries**_ - A column from your model that will underlie x-axis data in Rill's Explore dashboards and Canvas dashboards. Time can be truncated into different time periods.
* _**measures**_ - Numerical aggregates of columns from your data model shown on the y-axis of the explore charts and canvas components and the "big number" summaries.
* _**dimensions**_ - Categorical columns from your data model whose values are shown in _leaderboards_ in explore dashboard and allow you to look at segments or attributes of your data (and filter/slice accordingly) as well as selectable axis in Canvas dashboard components.

## Creating a Metrics view

Once your [model or underlying table](/build/metrics-view/underlying-model) is ready to visualize, you'll need to create a metrics view to define your measures and dimensions. This can be done in a few ways. Either create a blank YAML file, use the Add metrics view button, or "Generate Metrics with AI" from the model.

### Create a Metrics view with Code
Copy the below into a blank YAML or use the Add -> metrics view to create a blank metrics view. Here you can start to define dimensions and measures as seen below.


```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: example_model # Choose a table to underpin your metrics view
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

For more information about available metrics view properties, feel free to check our [reference YAML documentation](/reference/project-files/metrics-views).

:::

### Using the Visual Metrics Editor

When you add a metrics definition using the UI, a code definition will automatically be created as a YAML file in your Rill project within the metrics directory by default.

<img src='/img/build/metrics-view/visual-metrics-editor.png' class='rounded-gif' />
<br />



### Creating a Metrics View with AI

In order to streamline the process and get to a dashboard as quickly as possible, we've added the "Create Metrics with AI" and "Create Dashboard with AI" options! This will pass your schema to OpenAI to suggest measures and dimensions to get started with Rill. You can define your own OpenAI key by creating a [connector file](/reference/project-files/connectors#openapi). If you want to disable AI from your environment, please set the following in the `rill.yaml`:

```yaml
features:
  ai: false
```

## Creating Valid Metrics Views

### Underlying Model/Table

Before creating any measures or dimensions, you'll need to select a single model or table to power you metrics view. For a full walkthrough on DuckDB vs Live Connectors, see our [Underlying Model/Table](/build/metrics-view/underlying-model) doc.

### Time Series

Time is the most critical dimension in analytics and powers our dashboards. Understanding not just the "what," but how metrics evolve over hours, days, and months provides the narrative arc for decision-making. For a full walkthrough, see our [Time Series](/build/metrics-view/time-series) doc.

### Dimensions

Dimensions are used for exploring segments and filtering. Valid dimensions can be any type and are selected using the drop-down menu. You can also add labels and descriptions to your dimensions to customize how they are displayed. See our dedicated examples and pages for more use cases. 

- **[Clickable Dimension Links](/build/metrics-view/dimensions/dimension-uri)**
- **[Unnest Dimensions](/build/metrics-view/dimensions/unnesting)**
- **[Lookups](/build/metrics-view/dimensions/lookup)**


### Measures

Measures are numeric aggregates of columns from your data model. A measure must be defined with [DuckDB SQL](https://duckdb.org/docs/sql/introduction.html) aggregation functions and expressions on columns from your data model. The following operators and functions are allowed in measure expressions:

* Any DuckDB SQL [numeric](https://duckdb.org/docs/sql/functions/numeric) operators and functions
* This set of DuckDB SQL [aggregates](https://duckdb.org/docs/sql/aggregates): `AVG`, `COUNT`, `FAVG`, `FIRST`, `FSUM`, `LAST`, `MAX`, `MIN`, `PRODUCT`, `SUM`, `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, `VAR_SAMP`.
* [Filtered aggregates](https://duckdb.org/docs/sql/query_syntax/filter.html) can be used to filter the set of rows fed to the aggregate functions.

As an example, if you have a table of sales events with the sales price and customer ID, you could calculate the following metrics with these aggregates and expressions:
* Number of sales: `COUNT(*)`
* Total revenue: `SUM(sales_price)` 
* Revenue per customer: `CAST(SUM(sales_price) AS FLOAT)/CAST(COUNT(DISTINCT customer_id) AS FLOAT)`
* Number of orders with order value more than $100: `count(*) FILTER (WHERE order_val > 100)`

You can also add labels, descriptions, and your choice of number formatting to customize how they are shown. See our dedicated examples and pages for the following advanced measures!
- **[Measure Formatting](/build/metrics-view/measures/measures-formatting)**
- **[Case Statements and Filters](/build/metrics-view/measures/case-statements)**
- **[Referencing Measures](/build/metrics-view/measures/referencing)**
- **[Quantiles](/build/metrics-view/measures/quantiles)**
- **[Fixed Measures](/build/metrics-view/measures/fixed-measures)**
- **[Window Functions](/build/metrics-view/measures/windows)**

  

## Security

Data access is an important part of Rill that allows you to create specified views of your dashboard depending on who's viewing the page. For a dedicated guide, see [Data Access](/build/metrics-view/security) for more information.