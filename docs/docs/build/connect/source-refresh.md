---
title: Configure Source Refresh
description: Manage refresh schedules for sources deployed to Rill Cloud
sidebar_label: Configure Source Refresh
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

When creating or updating a source in Rill Cloud, you also have the option to configure how often the underlying source is refreshed (and thus ingested into the underlying OLAP layer powering Rill models and dashboards). By default, sources are refreshed manually but this can also be automated to a predefined schedule. This is handled through the underlying [source](/reference/project-files/sources.md) and/or [project YAML](/reference/project-files/rill-yaml.md#project-wide-defaults) using standard cron or Go duration syntax.

:::tip Configuring source refreshes for Cloud deployments

It is generally strongly recommended to configure source refreshes when [deploying a project](/deploy/existing-project/existing-project.md) to Rill Cloud to ensure that your production data (and dashboards) _remains up-to-date_. The interval that you should set really depends on how often your own source data is refreshed. Furthermore, while it is technically possible to configure source refreshes for Rill Developer as well, Rill Developer is primarily used for local development and thus typically does not require working with the most up-to-date data (local source refreshes that occur too often as well could also lead to resource constraints on your local machine). For more details, please see our pages on [environments](/build/models/environments#default-dev-and-prod-environments), [templating](/deploy/templating#environments-and-rill), and [performance optimization](/deploy/performance#configure-source-refresh-schedules-in-production-only).

:::

## Configuring source refresh individually

To specify a source refresh schedule for a particular source, this can be handled using the `refresh` property in the underlying YAML file. For example, to set a daily refresh for a source, you can do the following:

```yaml
refresh:
  every: 24h
```

Similarly, if you would like to utilize cron syntax, the following example would update a source every 15 minutes:

```yaml
refresh:
  cron: '*/15 * * * *'
```

:::note Source settings

For more details about available source configurations and properties, check our [Source YAML](../../reference/project-files/sources) reference page.

:::

## Configuring a project-wide default

You can also specify a project-wide refresh schedule that will apply to all sources by default. This can be done through the `rill.yaml` file. More details can be found [here](../../reference/project-files/rill-yaml#project-wide-defaults).

Using the same example as above, the following sets a project-wide default of refreshing sources every 24 hours:
```yaml
sources:
  refresh:
    every: 24h
```

Similarly, the following would use cron syntax to set a project-wide configuration of refreshing sources by default every 15 minutes (unless overridden at the individual source level):
```yaml
sources:
  refresh:
    cron: '*/15 * * * *'
```

:::info Did you know?

If you have both a project-wide default and source specific refresh schedule _configured in the same project_, the source specific refresh will **override** the project default based on how [inheritance](/build/models/environments#specifying-environment-specific-yaml-overrides) works in Rill. Otherwise, if not specified, the project-wide default will be used instead!

:::