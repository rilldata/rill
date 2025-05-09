---
note: GENERATED. DO NOT EDIT.
title: New Project files
sidebar_position: 30
---
## Overview

When you create models and dashboards, these objects are represented as object files on the file system. You can find these files in your `models` and `dashboards` folders in your project by default. 

:::info Working with resources outside their native folders

It is possible to define resources (such as [models](model.md), [metrics-views](metrics-view.md), [dashboards](explore-dashboard.md), [custom APIs](api.md), or [themes](theme.md)) within <u>any</u> nested folder within your Rill project directory. However, for any YAML configuration file, it is then imperative that the `type` property is then appropriately defined within the underlying resource configuration or Rill will not able to resolve the resource type correctly!

:::

Projects can simply be rehydrated from Rill project files into an explorable data application as long as there is sufficient access and credentials to the source data - figuring out the dependencies, pulling down data, & validating your model queries and metrics configurations. The result is a set of functioning exploratory dashboards.

You can see a few different example projects by visiting our [example github repository](https://github.com/rilldata/rill-examples).

:::tip

For more information about using Git or cloning projects locally, please see our page on [GitHub Basics](/deploy/deploy-dashboard/github-101).

:::

## Project files types

- [Alert YAML](alert.md)
- [API YAML](api.md)
- [Canvas Dashboard YAML](canvas-dashboard.md)
- [Component YAML](component.md)
- [Connector YAML](connector.md)
- [Explore Dashboard YAML](explore-dashboard.md)
- [Metrics View YAML](metrics-view.md)
- [Model YAML](model.md)
- [Report YAML](report.md)
- [Theme YAML](theme.md)
- [Project YAML](project.md)
