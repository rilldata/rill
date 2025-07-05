---
title: Model Project Defaults
description: Using environments to separate logic in Rill Developer and Cloud
sidebar_label: Model Project Defaults
sidebar_position: 30
---

## Default dev and prod environments

Rill comes with default `dev` and `prod` properties defined, corresponding to Rill Developer and Rill Cloud respectively. You can use these keys to set environment-specific YAML overrides or SQL logic.

For example, the following `rill.yaml` file explicitly sets the default materialization setting for models to `false` in development and `true` in production:
```yaml
dev:
  models:
    materialize: false

prod:
  models:
    materialize: true
```

### Specifying a custom environment

When using Rill Developer, instead of defaulting to `dev`, you can run your project in production mode using the following command:

```bash
rill start --environment prod
```

## Specifying environment specific YAML overrides

If you wanted to set a project-wide default in `rill.yaml` where models are [materialized](/reference/project-files/models#model-materialization) only on Rill Cloud (i.e. `prod), or set a default model refresh at a specific time. 

```yaml
models:
  materialize: true

  refresh:
    cron: 0 * * * *
```

:::info Hierarchy of inheritance and property overrides

As a general rule of thumb, properties that have been specified at a more _granular_ level will supercede or override higher level properties that have been inherited. Therefore, in order of inheritance, Rill will prioritize properties in the following order:
1. Individual [model](/reference/project-files/models)/ object level properties (e.g. `source.yaml` or `dashboard.yaml`)
2. [Environment](#default-dev-and-prod-environments) level properties (e.g. a specific property that have been set for `dev`)
3. [Project-wide defaults](#specifying-environment-specific-yaml-overrides) for a specific property and resource type

:::

## Running scheduled source refreshes in development

As an exception, scheduled source refreshes specified using `refresh:` are not applied in the `dev` environment by default. If you want to run or test scheduled refreshes in local development, you can override this behavior using the `run_in_dev` property:
```yaml
models:
  refresh:
    cron: 0 * * * *
    run_in_dev: true
```

:::tip Why are source refreshes only enabled by default for Rill Cloud?

Source refreshes are primarily meant to _help keep the data in your deployed dashboards on Rill Cloud up-to-date_ (without needing to manually trigger refreshes). For more details, see our documentation on [configuring source refreshes](/ingest/connect/source-refresh.md).

:::

