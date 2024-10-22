---
title: "What is a Metrics View?"
sidebar_label: "What is a Metrics View?"
sidebar_position: 11
hide_table_of_contents: true
---

What is a metrics layer, and why did we decide to split the dashboard into two individual components? 


## What is a Metrics Layer?

A metrics layer is a `centralized framework` used to define and organize **key metrics** for your organization. Having a centralized layer allows an organization to easily manage and re-use calculations across various reports, dashboard, and data tools. As Rill continues to grow, we decided to separate metrics layer from the dashboard configuration.


### Historically, in Rill...
<img src = '/img/concepts/metrics-view/old-dashboard.png' class='rounded-gif' />
<br />

Historically in Rill, the metrics layer and dashboard configuration were a single file. As seen above, the metrics would be defined **inside** a dashboard YAML file along with the dashboard components and dashboard customizations. However, as we continue to create more features, we found that this was not the best approach. In order to create a metrics layer in Rill as a first class resource and not a consequence of dashboards, we found it necessary to split the two resources into their own files. Thus, the metrics-view was born.

:::tip
Starting from version 0.50, the operation of creating a dashboard via AI will create a metrics-view and dashboard separately in their own respective folders and navigate you to a preview of your dashboard. If you find that some of the metrics need to be modified, you will need to navigate to your [metrics/model_name_metrics.yaml](/build/metrics-view/) file. 


Assuming that you have the ' * ', select all, in your dashboard configurations, any changes will automatically be changed on your [explore dashboard](/build/dashboards/).
:::

## Splitting the Dashboard into two components, Metrics-view and Dashboard Configuration
Splitting the metrics view into its own component allows us more freedom to continue building Rill and adding new additional features. Instead of querying a dashboard for data, we would be querying the metrics-layer. The dashboard will directly query the metrics-view along with many new components that are currently being developed.

### New Metrics View as an independent object in Rill 

![img](/img/concepts/metrics-view/metrics-view-components.png)

### (Explore) Dashboard

With the split of metrics view, dashboard configurations experienced an overhaul. Instead of defining measure and dimensions, you will now reference the object into your dashboard. What this allows is creating customized dashboards for specific viewers and reusability of a single metrics-view in multiple dashboards!

![img](/img/concepts/metrics-view/explore-dashboard.png)

For more information on what is changed, please review the [migrations](/manage/migration) and the [reference page](/reference/project-files/explore-dashboards).

## More to come!

New features are being developed as we speak! If you're curious, feel free to [contact us](/contact)! 