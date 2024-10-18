---
title: "What is a Metrics View?"
sidebar_label: "What is a Metrics View?"
sidebar_position: 11
hide_table_of_contents: true
---

What is a metrics layer, and why did we decide to split the dashboard into two individual components? 


## What is a Metrics Layer?

A metrics layer is a `centralized framework` used to define and organize **key metrics** for your organization. Having a centralized layer allows an organization to easily manage and re-use calculations across various reports, dashboard, and data tools. As Rill continues to grow, we decided to separate metrics layer from the dashboard configuration.



:::tip
Starting from version 0.50, the operation of creating a dashboard via AI will create a metrics-view and dashboard separately in their own respective folders and navigate you to a preview of your dashboard. If you find that some of the metrics need to be modified, you will need to navigate to your [metrics/model_name_metrics.yaml](/build/metrics-view/) file. 


Assuming that you have the ' * ', select all, in your dashboard configurations, any changes will automatically be changed on your [explore dashboard](/build/dashboards/).
:::

<h1> rewrite this! </h1>