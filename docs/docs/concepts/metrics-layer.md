---
title: "What is a Metrics Layer?"
sidebar_label: "What is a Metrics View?"
sidebar_position: 11
hide_table_of_contents: true
---

## What is a Metrics Layer?

A metrics layer is a `centralized framework` used to define and organize **key metrics** for your organization. Having a centralized layer allows an organization to easily manage and reuse calculations across various reports, dashboards, and data tools. As Rill continues to grow, we decided to separate the metrics layer from the dashboard configuration.

:::tip
Starting from version 0.50, the operation of creating a dashboard via AI will create a metrics view and dashboard separately in their own respective folders and navigate you to a preview of your dashboard. If you find that some metrics need to be modified, you will need to navigate to your [metrics/model_name_metrics.yaml](/build/metrics-view) file. 


Assuming that you have the '*' (select all) in your dashboard configurations, any changes will automatically be reflected on your [explore dashboard](/build/dashboards).
:::


## Introducing Metrics View
Within Rill, we refer to metrics layers as a metrics view. It's a single view or file that contains all of your measures and dimensions that will be used to display the data in various ways. The metrics view also contains some configuration settings that are required to ensure that the data being displayed is as accurate as you need it to be. 


<img src = '/img/tutorials/rill-basics/new-viz-editor.png' class='rounded-gif' />
<br />

:::tip
It is possible to develop the metrics layer in a traditional BI-as-code manner as well as via the UI. To switch between the two, select the toggle in the top right corner.
:::

## Interfacing with a Metrics view

A metrics view on its own does not have any visualization capabilities. Instead, the metrics view is the building block for all the visualizations and components within Rill. Please see below for the currently available and soon-to-be-released features of Rill!

<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" width="100%" style={{ borderRadius: '15px', padding: '20px' }} />
</div>


## Next Steps

- [Learn about Rill's Architecture](/concepts/architecture)
- [Get started with Rill](/home/install)
- [Explore the Reference](/connect)
- [Step-by-step Tutorial](/guides)
