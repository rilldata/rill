---
title: Structure your project
sidebar_label: Organize your Code Files
sidebar_position: 00
---

After creating your initial set of sources, models, and dashboards, you may have noticed the following _native_ folders that exist in your Rill project directory:
- [Models](/reference/project-files/models)
- [Metrics Views](/reference/project-files/metrics-views)
- [Dashboards](/reference/project-files/explore-dashboards)

By default, any new sources, models, metrics views, and dashboards will be created in their respective native folders. However, this does not necessarily have to be the case, and Rill Developer allows a flexible project directory structure, including nested folders or even storing objects in non-native folders. This is a powerful feature that allows you, as a developer, to organize your project to meet your team's specific needs.

## Adding new resources or parent folders

Within Rill Developer, from the left-hand side (file explorer), you should be able to click on the `Add` button to add a new resource, such as a new source, model, or dashboard. Furthermore, you will also have the ability to add a new parent folder to store groups of resources (which can be mixed). If you choose to add a new folder, you should see the folder structure reflected when you check the project directory via the CLI. 

<img src = '/img/build/structure/adding-objects.png' class='rounded-gif' />
<br />

:::warning Make sure to include the `type` property

For backward compatibility purposes, any resource that belongs in the `sources`, `models`, and `dashboards` native folders is assumed to be a source, model, or dashboard respectively (including nested folders that belong within a native folder). 

However, if you'd like to create a resource outside one of these native folders, make sure to include the `type` property in the resource definition, or Rill will not be able to properly resolve the resource type! For more details, see our [reference documentation](/reference/project-files/rill-yaml).

:::

## Navigating Upstream / Downstream Objects

<img src = '/img/build/structure/breadcrumb.png' class='rounded-gif' />
<br />
When selecting between a source, model, metrics view, and dashboard, you can view the upstream/downstream objects to the current view. For example, if you are selecting a metrics view, you can see all of the dashboards (in a dropdown) that are built on the metrics view. Likewise, if your model references several sources, these will be available to select. 

## Moving resources within your project

From the UI, within the file explorer, you should be able to drag resources/objects around and move them to or from folders as necessary. 

:::info Using the CLI

For developers who prefer to use the CLI, the project structure can still be controlled or adjusted directly via the CLI or your preferred IDE (e.g., VS Code).

:::

## Adding a nested folder

Rather than creating a parent folder and moving it into another folder manually, there is a shortcut to create a nested folder by directly hovering over an existing folder in the file explorer, clicking on the "triple dots," and selecting `New folder`.

<img src = '/img/build/structure/adding-nested-folder.png' class='rounded-gif' />
<br />