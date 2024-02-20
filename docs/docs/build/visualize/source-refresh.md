---
title: Configure Source Refresh
description: Manage refresh schedules for sources deployed to Rill Cloud
sidebar_label: Configure Source Refresh
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

After deploying a project to Rill Cloud, you can then configure how often the underlying sources are refreshed. This is handled through the underlying source or project YAML using standard cron or go duration string syntax. 

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

:::tip Source settings

For more details about available source configurations and properties, check our [Source YAML](https://docs.rilldata.com/reference/project-files/sources) reference page.

:::

## Configuring a project-wide default

You can also specify a project-wide refresh schedule that will apply to all sources by default. This can be done through the `rill.yaml` file. More details can be found [here](https://docs.rilldata.com/reference/project-files/rill-yaml#project-wide-defaults).

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

If you have both a project-wide default and source specific refresh schedule configured in the same project, the source specific refresh will override the project default. Otherwise, the schedule will be inherited!

:::