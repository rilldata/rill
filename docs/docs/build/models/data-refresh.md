---
title: Schedule Your Data Refresh
description: Manage refresh schedules for models deployed to Rill Cloud
sidebar_label: Scheduled Refreshes
sidebar_position: 15
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

When creating or updating a model in Rill Cloud, you also have the option to configure how often the underlying model is refreshed (and thus ingested into the underlying OLAP layer powering Rill models and dashboards). By default, models are refreshed manually, but this can also be automated to a predefined schedule. This is handled through the underlying [model settings](/reference/project-files/models) and/or [project YAML](/reference/project-files/rill-yaml#project-wide-defaults) using standard cron or Go duration syntax.

:::tip Configuring model refreshes for Cloud deployments

It is generally strongly recommended to configure model refreshes when [deploying a project](/deploy/deploy-dashboard) to Rill Cloud to ensure that your production data (and dashboards) _remains up-to-date_. The interval that you should set really depends on how often your own data is being refreshed. Furthermore, while it is technically possible to configure model refreshes for Rill Developer as well, Rill Developer is primarily used for local development and thus typically does not require working with the most up-to-date data (local model refreshes that occur too often could also lead to resource constraints on your local machine). For more details, please see our pages on [environments](/build/connectors/credentials#variables), [templating](/build/connectors/templating), and [performance optimization](/build/models/performance).

:::

## Configuring Model Refresh Individually

To specify a model refresh schedule for a particular model, this can be handled using the `refresh` property in the underlying YAML file. For example, to set a daily refresh for a model, you can do the following:

```yaml
refresh:
  every: 24h
```

Similarly, if you would like to utilize cron syntax, the following example would update a model every 15 minutes:

```yaml
refresh:
  cron: '*/15 * * * *'
```

:::note Model settings

For more details about available model configurations and properties, check our [model YAML](/reference/project-files/models) reference page.

:::

## Configuring a Project-Wide Default

You can also specify a project-wide refresh schedule that will apply to all models by default. This can be done through the `rill.yaml` file. More details can be found [here](/reference/project-files/rill-yaml#project-wide-defaults).

:::tip Set project-wide defaults
You can set a default refresh schedule for all models in your project.
[Learn more about project defaults →](/build/project-configuration#model-refresh-schedule)
:::

Using the same example as above, the following sets a project-wide default of refreshing models every 24 hours:

```yaml
models:
  refresh:
    every: 24h
```

Similarly, the following would use cron syntax to set a project-wide configuration of refreshing models by default every 15 minutes (unless overridden at the individual model level):

```yaml
models:
  refresh:
    cron: '*/15 * * * *'
```

:::info Did you know?

If you have both a project-wide default and model-specific refresh schedule _configured in the same project_, the model-specific refresh will **override** the project default based on how inheritance works in Rill. Otherwise, if not specified, the project-wide default will be used instead!

:::

## Running Scheduled Source Refreshes in Development

As an exception, scheduled source refreshes specified using `refresh:` are not applied in the `dev` environment by default. If you want to run or test scheduled refreshes in local development, you can override this behavior using the `run_in_dev` property:

```yaml
refresh:
  cron: 0 * * * *
  run_in_dev: true
```