---
title: Explore Dashboards
description: Explore dashboard over view
sidebar_label: Explore Dashboards
sidebar_position: 00
---

Explore dashboards are interactive, slice-and-dice interfaces that transform your metrics view data into powerful visualizations for data exploration and analysis. Built on top of a single metrics view, they provide an intuitive way for users to interact with your data through real-time filtering, drilling, and comparison capabilities.

## Creating an Explore Dashboard


### Using the Code Editor

In the Explore dashboard YAML, you can define which measures and dimensions are visible, as well as the default view when a user sees your dashboard.

* _**metrics_view**_ - A metrics view that powers the dashboard
* _**measures**_ - `*` Which measures to include or exclude from the metrics view; using a wildcard will include all.
* _**dimensions**_ - `*` Which dimensions to include or exclude from the metrics view; using a wildcard will include all.

When including dimensions and measures, only the named resources will be included.
Rill also supports the ability to exclude a set of named dimensions and measures.

```yaml
type: explore

title: Title of your Explore Dashboard
description: a description for your explore dashboard
metrics_view: my_metricsview

dimensions: '*' # can use expressions
measures: '*' # can use expressions

defaults: # define all the defaults within here, was default_* in previous dashboard YAML
    dimensions: 
    measures:
    ...
```

### Using AI

In various locations throughout the platform, you have the opportunity to create a dashboard via AI. What this does is create the underlying metrics view with AI and creates your explore.yaml similar to the above with the required components.

## Preview a Dashboard in Rill Developer

Once a dashboard is ready to preview, before [deploying to Rill Cloud](/deploy/deploy-dashboard), you can preview the dashboard in Rill Developer. Especially if you are setting up [dashboard policies](/build/dashboards/customization#define-dashboard-access), it is recommended to preview and test the dashboard before deploying.

<img src='/img/build/dashboard/preview-dashboard.png' class='rounded-gif' />
<br />