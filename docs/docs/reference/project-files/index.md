---
note: GENERATED. DO NOT EDIT.
title: YAML Syntax
sidebar_position: 30
---

## Overview

When you create models and dashboards, these objects are represented as object files on the file system. You can find these files in your `models` and `dashboards` folders in your project by default. 

:::info Working with resources outside their native folders

It is possible to define resources (such as [models](models.md), [metrics-views](metrics-views.md), [dashboards](explore-dashboards.md), [custom APIs](apis.md), or [themes](themes.md)) within <u>any</u> nested folder within your Rill project directory. However, for any YAML configuration file, it is then imperative that the `type` property is then appropriately defined within the underlying resource configuration or Rill will not able to resolve the resource type correctly!

:::

Projects can simply be rehydrated from Rill project files into an explorable data application as long as there is sufficient access and credentials to the source data - figuring out the dependencies, pulling down data, & validating your model queries and metrics view configurations. The result is a set of functioning exploratory dashboards.

You can see a few different example projects by visiting our [example github repository](https://github.com/rilldata/rill-examples).

:::tip

For more information about using Git or cloning projects locally, please see our page on [GitHub Basics](/deploy/deploy-dashboard/github-101).

:::


## Project files types


- [Connector YAML](connectors.md)
- [Source YAML](sources.md)
- [Models YAML](models.md)
- [Metrics View YAML](metrics-views.md)
- [Canvas Dashboard YAML](canvas-dashboards.md)
- [Explore Dashboard YAML](explore-dashboards.md)
- [Alert YAML](alerts.md)
- [API YAML](apis.md)
- [Theme YAML](themes.md)
- [Component YAML](component.md)
- [Project YAML](rill-yaml.md)