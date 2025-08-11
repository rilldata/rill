---
title: "Debugging Dashboard Access"
description: Getting Unexpected behavior in Dashboard Access
sidebar_label: "Debugging Dashboard Access"
sidebar_position: 00
---

# Debugging Dashboard Access

Dashboard access and data control are fundamental aspects of Rill, offering multiple configuration options for securing your dashboards. This complexity can sometimes lead to access issues that need troubleshooting.

This guide provides essential troubleshooting steps and solutions to quickly resolve dashboard access problems and get your dashboards working again.


:::note 

This is assuming that you've already set up and understand [dashboard / data access policies](/build/metrics-view/security). 

:::


## Where to start?

There are three different types of data access issues that you might encounter in Rill.

1. 403, **permission denied**
2. **Failed to load dashboard**: This dashboard currently has no data to display. This may be due to access permissions.
3. **Canvas dashboard component** failed to load.


Depending on the type of error you are experiencing, depends on how you can go about troubleshooting.

:::tip not an admin/editor?

If you are not an admin, editor or have access to the underlying GitHub repository, you will need to reach out internally to resolve the issue.

:::

## Where are my policies defined?

In Rill, there are project default settings in the rill.yaml and object level overrides.

```yaml
#rill.yaml prjoect defaults

```


```yaml
#metrics/metrics_view.yaml or dashboards/canvas_explore.yaml

```


## Override or Combine



## Map your policies


> create a mini app here



## Test changes in Rill Developer




## Push to Rill Cloud