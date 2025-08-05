---
title: Create Explore Dashboards
description: Create dashboards using source data and models with time, dimensions, and measures
sidebar_label: Create Explore Dashboards
sidebar_position: 00
---

<img src = '/img/build/dashboard/explore-dashboard.gif' class='rounded-gif' />
<br />

In Rill, Explore dashboards are used to visually understand your data with real-time filtering, based on the dimensions and measures defined in your metrics view. In the Explore dashboard YAML, you can define which measures and dimensions are visible, as well as the default view when a user sees your dashboard. 

:::tip
Starting in version 0.50, metrics views have been separated from Explore dashboards. This allows for a cleaner, more accessible metrics layer and the ability to build various dashboards and components on top of a single metrics view. For more information on what a metrics view is please see: [What is a Metrics View?](/concepts/metrics-layer)

For migration steps, see [Migrations](/other/v50-dashboard-changes#how-to-migrate-your-current-dashboards).
:::

* _**metrics_view**_ - A metrics view that powers the dashboard
* _**measures**_ - `*` Which measures to include or exclude from the metrics view; using a wildcard will include all.
* _**dimensions**_ - `*` Which dimensions to include or exclude from the metrics view; using a wildcard will include all.

When including dimensions and measures only the named resources will be included. 
Rill also supports the ability to exclude a set of named dimensions and measures.

```yaml
type: explore

title: Title of your Explore Dashboard
description: a description for your explore dashboard
metrics_view: my_metricsview

dimensions: '*' #can use expressions
measures: '*' #can use expressions

defaults: #define all the defaults within here, was default_* in previous dashboard YAML
    dimensions: 
    measures:
    ...
security:
    access: #only dashboard access can be defined here, other security policies must be set on the metrics view
```


:::note Dashboard Properties
For more details about available configurations and properties, check our [Dashboard YAML](/reference/project-files/explore-dashboards) reference page.
:::

### Preview a Dashboard in Rill Developer
Once a dashboard is ready to preview, before [deploying to Rill Cloud](/deploy/deploy-dashboard), you can preview the dashboard in Rill Developer. Especially if you are setting up [dashboard policies](/manage/security), it is recommended to preview and test the dashboard before deploying.

<img src = '/img/build/dashboard/preview-dashboard.png' class='rounded-gif' />
<br />


### Clickable Dimension Links 
Adding an additional parameter to your dimension in the [metrics view](/build/metrics-view) can allow for clickable links directly from the dashboard.

```yaml
dimensions:
  - label: Company Url
    column: Company URL
    uri: true #if already set to the URL, also accepts SQL expressions
```
 <img src = '/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />


### Multi-editor and external IDE support

Rill Developer is meant to be developer friendly and has been built around the idea of keystroke-by-keystroke feedback when modeling your data, allowing live interactivity and a real-time feedback loop to iterate quickly (or make adjustments as necessary) with your models and dashboards. Additionally, Rill Developer supports "hot reloading", which means that you can keep two windows of Rill open at the same time and/or use a preferred editor, such as VSCode, side-by-side with the dashboard that you're actively developing!

<img src = 'https://cdn.rilldata.com/docs/release-notes/36_hot_reload.gif' class='rounded-gif' />
<br />