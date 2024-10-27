---
title: "What is a Metrics Layer?"
sidebar_label: "What is a Metrics View?"
sidebar_position: 11
hide_table_of_contents: true
---

## What is a Metrics Layer?

A metrics layer is a `centralized framework` used to define and organize **key metrics** for your organization. Having a centralized layer allows an organization to easily manage and re-use calculations across various reports, dashboard, and data tools. As Rill continues to grow, we decided to separate metrics layer from the dashboard configuration.

:::tip
Starting from version 0.50, the operation of creating a dashboard via AI will create a metrics view and dashboard separately in their own respective folders and navigate you to a preview of your dashboard. If you find that some of the metrics need to be modified, you will need to navigate to your [metrics/model_name_metrics.yaml](/build/metrics-view/) file. 


Assuming that you have the ' * ', select all, in your dashboard configurations, any changes will automatically be changed on your [explore dashboard](/build/dashboards/).
:::


## Introducing Metrics View
Within Rill, we refer to Metrics layers as a metrics view. It's a single view or file that contains all of your measures and dimensions that will be used to display the data in various ways. The metrics view also contains some configurations settings that are required to ensure that the data being displayed is as accurate as you need. 

![img](/img/tutorials/102/new-viz-editor.png)

:::tip
It is possible to develop the metrics layer in a traditional BI-as-code manner as well as via the UI. To switch between the two, select the toggle in the top right corner.
:::

## Interfacing with a Metrics view

A metrics view on its own does not have any visualization capabilities. Instead, the metrics view is the building block for all of the visualization and components within Rill. Please see below for the currently available and soon to be released features of Rill!

![img](/img/concepts/metrics-view/metrics-view-components.png)

