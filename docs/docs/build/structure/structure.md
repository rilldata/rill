---
title: Structure your project
sidebar_label: Structure Project
sidebar_position: 00
---

After creating your initial set of sources, models, and dashboards, you may have noticed the following _native_ folders that exist in your Rill project directory:
- [Sources](/reference/project-files/sources)
- [Models](/reference/project-files/models)
- [Metrics View](/reference/project-files/metrics-view)
- [Dashboards](/reference/project-files/explore-dashboards)

By default, any new sources, models, metrics view and dashboards will be created in their respective native folder. However, this does not necessarily have to be the case and Rill Developer allows a flexible project directory structure, including nested folders or even storing objects in non-native folders. This is a powerful feature that allows you as a developer to organize your project to meet your specific team needs.

## Adding new resources or parent folders

Within Rill Developer, from the left-hand side (file explorer), you should be able to click on the `Add` button to add a new base resource, such as a new source, model, or dashboard. Furthermore, you will also have the ability to add a new parent folder to store groups of resources (can be mixed). If you choose to add a new folder, you should see the folder structure reflected if you check the project directory via the CLI. 

![Adding objects](/img/build/structure/adding-objects.png)

:::warning Make sure to include the `type` property

For backward-compatibility purposes, any resource that belongs in the `sources`, `models`, and `dashboards` native folders are assumed to be a source, model, or dashboard respectively (including nested folders that belong within a native folder). 

However, if you'd like to create a resource outside one of these native folders, make sure to include the `type` property in the resource definition or Rill will not be able to properly resolve the resource type! For more details, see our [reference documentation](/reference/project-files/index.md).

:::

## Navigating Upstream / Downstream Objects

<img src = '/img/build/structure/breadcrumb.png' class='rounded-gif' />
<br />
When selecting between a source, model, metrics view and dashboard, you can view the upstream/downstream objects to the current view. For example, if you are selecting a metrics view, you can see all of the dashboards (in a dropdown) that are built on the metrics view. Likewise, if your model references several sources, this will be available to select. 

## Moving resources within your project

From the UI, within the file explorer, you should be able to drag resources / objects around and move them from / to folders as necessary. 

:::info Using the CLI

For developers who prefer to use the CLI, the project structure can still be controlled or adjusted directly via the CLI and/or using your preferred IDE of choice (e.g. VSCode).

:::

## Adding a nested folder

Rather than creating a parent folder and moving it into another folder manually, there is a shortcut to create a nested folder by directly hovering over an existing folder in the file explorer > click on the "triple dots" > `New folder`.

![Adding nested folder](/img/build/structure/adding-nested-folder.png)