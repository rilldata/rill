---
title: Create Dashboards
description: Create dashboards using source data and models with time, dimensions, and measures
sidebar_label: Create Dashboards
sidebar_position: 00
---

In Rill, dashboards are one of many components that access the metrics layer. Currently, it is possible to create an explore dashboard but more features are on the way!

![img](/img/build/dashboard/explore-dashboard.png)

* _**metrics_view**_ - A metrics view that powers the dashboard
* _**measures**_ - `*` Which measures to include or exclude from the metrics view, using a wildcard will include all.
* _**dimensions**_ -  `*` Which dimensions to include or exclude from the metrics view, using a wildcard will include all.

When including dimensions and measures only the named resources will be included. 
Rill also supports the ability to exclude a set of named dimensions and measures.

```yaml
metrics_view: my_metrics_view

dimensions: [country, region, product_category] # Only these three dimensions will be included
measures:
  exclude: [profit] # All measures except profit will be included
```

:::tip
Starting in version 0.50, metrics view has been separated from dashboard. This allows for a cleaner, more accessible metrics layer and the ability to build various dashboards and components on top of a single metrics layer. For more information on why we decided to do this, please refer to the following: [Why separate the dashboard and metrics layer](/concepts/metrics-layer)

For migration steps, see [Migrations](/manage/migration#v049---v050).
:::


:::note Dashboard Properties
For more details about available configurations and properties, check our [Dashboard YAML](/reference/project-files/explore-dashboards) reference page.
:::

### Preview a Dashboard in Rill Developer
Once a dashboard is ready to preview, before [deploying to Rill Cloud](/deploy/deploy-dashboard/), you can preview the dashboard in Rill Developer. Especially if you are setting up [dashboard policies](/manage/security), it is recommended to preview and test the dashboard before deploying.

![preview](/img/build/dashboard/preview-dashboard.png)


### Multi-editor and external IDE support

Rill Developer is meant to be developer friendly and has been built around the idea of keystroke-by-keystroke feedback when modeling your data, allowing live interactivity and a real-time feedback loop to iterate quickly (or make adjustments as necessary) with your models and dashboards. Additionally, Rill Developer has support for the concept of "hot reloading", which means that you can keep two windows of Rill open at the same time and/or use a preferred editor of choice, such as VSCode, side-by-side with the dashboard that you're actively developing!

![hot-reload-0-36](https://cdn.rilldata.com/docs/release-notes/36_hot_reload.gif)
