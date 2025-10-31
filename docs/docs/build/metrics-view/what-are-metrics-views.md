---
title: Get Started with Metrics Views
description: Create metrics view using source data and models with time, dimensions, and measures
sidebar_label: What are Metrics Views?
sidebar_position: 00
---


A metrics view is a 'centralized framework' used to define and organize **key measures and dimensions** for your organization. Having a centralized layer allows an organization to easily manage and reuse calculations across various reports, dashboards, and data tools. Each metrics view is powered by a single [model or table](/build/metrics-view/underlying-model).

Rill metrics views are analogous to the **semantic layer** and **metrics layer** concepts found in other business intelligence and analytics tools. It provides a consistent, reusable abstraction over your underlying data, enabling standardized definitions of business metrics that can be shared across different dashboards and reports.


<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" width="100%" style={{ borderRadius: '15px', padding: '20px' }} />
</div>


In Rill, your metrics view is defined by _metric definitions_. Metric definitions are composed of:
* [_**model/table**_](/build/metrics-view/underlying-model) - A data model or underlying table created with the concept of [One Big Table](/build/models/models-101#one-big-table-and-dashboarding) that will power the metrics view.
* [_**timeseries**_](/build/metrics-view/time-series) - A column from your model that will underlie x-axis data in Rill's Explore dashboards and Canvas dashboards. Time can be truncated into different time periods.
* [_**dimensions**_](/build/metrics-view/dimensions) - Categorical columns from your data model whose values are shown in _leaderboards_ in explore dashboard and allow you to look at segments or attributes of your data (and filter/slice accordingly) as well as selectable axis in Canvas dashboard components.
* _[**measures**](/build/metrics-view/measures)_ - Numerical aggregates of columns from your data model shown on the y-axis of the explore charts and canvas components and the "big number" summaries.

## Creating a Metrics view

Once your [model or underlying table](/build/metrics-view/underlying-model) is ready to visualize, you'll need to create a metrics view to define your measures and dimensions. This can be done in a few ways. Either create a blank YAML file, use the Add metrics view button, or "Generate Metrics with AI" from the model.



### Create a Metrics view with Code
Copy the below into a blank YAML or use the Add -> metrics view to create a blank metrics view. Here you can start to define dimensions and measures as seen below.


```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: example_model # Choose a model to underpin your metrics view
timeseries: timestamp_column # Choose a timestamp column (if any) from your table

dimensions:
  - column: category
    display_name: "Category"
    description: "Description of the dimension"

measures:
  - expression: "SUM(revenue)"
    display_name: "Total Revenue"
    description: "Total revenue generated"
```

:::tip Set project-wide time defaults
Configure default time modifiers like first day of week and smallest time grain for all metrics views.
[Learn more about metrics view defaults →](/build/project-configuration#metrics-views-time-modifiers)
:::
### Using the Visual Metrics Editor

When you add a metrics definition using the UI, a code definition will automatically be created as a YAML file in your Rill project within the metrics directory by default.

<img src='/img/build/metrics-view/visual-metrics-editor.png' class='rounded-gif' />
<br />



### Creating a Metrics View with AI


In order to streamline the process and get to a dashboard as quickly as possible, we've added the "Create Metrics with AI" and "Create Dashboard with AI" options! This will pass your schema to OpenAI to suggest measures and dimensions to get started with Rill.

<img src='/img/build/metrics-view/create-with-ai.png' class='rounded-gif' />
<br />


:::tip Skipped creating a model?

You can now create dashboards directly for the connector panel. This will create a model, metrics view and dashboard all in one step.

<img src='/img/build/dashboard/explorable-metrics.png' class='rounded-gif' />
<br />

:::

You can define your own OpenAI key by creating a [connector file](/reference/project-files/connectors#openai). If you want to disable AI from your environment, please set the following in the `rill.yaml`:


```yaml
features:
  ai: false
```