---
title: Project files
sidebar_label: Project files
sidebar_position: 0
hide_table_of_contents: true
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

When you create sources, models, and dashboards, these objects are represented as object files on the file system. You can find these files in your `sources`, `models` and `dashboards` folders in your project by default. 

:::info Working with resources outside their native folders

It is possible to define [sources](sources.md), [models](models.md), and [dashboards](dashboards.md) within <u>any</u> nested folder within your Rill project directory. However, it is imperative that the `kind` property is then appropriately defined in the underlying resource configuration or Rill will not able to resolve the resource type correctly!

:::

Projects can simply be rehydrated from Rill project files into an explorable data application as long as there is sufficient access and credentials to the source data - figuring out the dependencies, pulling down data, & validating your model queries and metrics configurations. The result is a set of functioning exploratory dashboards.

You can see a few different example projects by visiting our [example github repository](https://github.com/rilldata/rill-developer-example.git).

:::tip

For more information about using Git or cloning projects locally, please see our page on [Github Basics](/deploy/existing-project/github-101).

:::
