---
title: Explore Dashboards
description: Explore dashboard overview
sidebar_label: Explore Dashboards
sidebar_position: 00
---

Explore dashboards are interactive, slice-and-dice interfaces that transform your metrics view data into powerful visualizations for data exploration and analysis. Built on top of a single metrics view, they provide an intuitive way for users to interact with your data through real-time filtering, drilling, and comparison capabilities.

## Creating an Explore Dashboard in Rill Developer

### Using the Code Editor

In the Explore dashboard YAML, you can define dashboard level parameters to customize the capabilities. For a full list, see our [explore dashboard reference](/reference/project-files/explore-dashboards) doc.

* _**metrics_view**_ - A single metrics view that powers the dashboard
* _**measures**_ - `*` Which measures to include or exclude from the metrics view; using a wildcard will include all.
* _**dimensions**_ - `*` Which dimensions to include or exclude from the metrics view; using a wildcard will include all.

In some cases, a specific dashboard will not need to include all of the underlying metrics view's measures and/or dimensions. In this case, you can use the `measures` and `dimensions` parameters to filter these out. Rill supports providing a single value, list, or regex to filter out unnecessary measures and dimensions.

```yaml
type: explore

title: Title of your Explore Dashboard
description: a description for your explore dashboard
metrics_view: my_metricsview

dimensions:
    expr: "^public_.*$"
measures:
  - total_downloads
  - total_impressions 

defaults:
    comparison_mode: time
    time_range: P3M
  measures:
    - total_downloads
    - total_impressions 
  dimensions:
    - show_name
    - season
    - program_name
```
:::tip Customize default time ranges
Set project-wide default time ranges and available options for all explore dashboards.
[Learn more about dashboard defaults →](/build/project-configuration#explore-defaults)
:::

### Using AI

In various locations throughout the platform, you have the opportunity to fast-track your dashboard creation via AI. This feature [creates the underlying metrics view with AI](/build/metrics-view/what-are-metrics-views#creating-a-metrics-view-with-ai) and generates your explore.yaml similar to the example above with the required components.

**Directly from your connector's tables:**

<img src='/img/build/dashboard/explorable-metrics.png' class='rounded-gif' />
<br />

**From an ingested model:**


<img src='/img/build/metrics-view/create-with-ai.png' class='rounded-gif' />
<br />

## Preview a Dashboard in Rill Developer

Once a dashboard is ready to preview, before [deploying to Rill Cloud](/deploy/deploy-dashboard), you can preview the dashboard in Rill Developer.

<img src='/img/build/dashboard/preview.png' class='rounded-gif' />
<br />

### Setting Up Dashboard Access

If you are setting up [dashboard policies](/build/dashboards/customization#define-dashboard-access), it is recommended to preview and test the dashboard before deploying. This option will be available for testing if you have set up access policies at the [project level](/build/project-configuration#testing-security), [metrics view level](/build/metrics-view/security), or [dashboard level](/build/dashboards/customization#define-dashboard-access).



<img src='/img/build/dashboard/preview-dashboard.png' class='rounded-gif' />
