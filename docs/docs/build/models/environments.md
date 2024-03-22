---
title: Environments
description: Using environments to separate logic in Rill Developer and Cloud
sidebar_label: Environments
sidebar_position: 10
---

## Overview

Environments allow the separation of logic between different deployments of Rill, most commonly between when using a project in Rill Developer and Rill Cloud. Generally speaking, Rill Developer is meant primarily for local development purposes, where the developer may want to use a segment or sample of data for modeling purposes to help validate the model logic is correct and producing expected results. Then, once finalized and a project is ready to be deployed to Rill Cloud, the focus is on shared collaboration at scale and where most of the dashboard consumption in production will happen.

There could be a few reasons to have separate logic for sources, models, and dashboards for Rill Developer and Rill Cloud respectively:
1. As the size of data starts growing, the locally embedded OLAP engine (using DuckDB) may start to face scaling challenges, which can impact performance and the "snappiness" of models and dashboards. Furthermore, for model development, the full data is often not needed and working with a sample is sufficient. For production though, where analysts and business users are interacting with Rill dashboards to perform interactive, exploratory analysis or make decisions, it is important that these same models and dashboards are powered by the entire range of data available.
2. For certain organizations, there might be a development and production version of source data. Therefore, you can develop your models and validate the results in Rill Developer against development. When deployed to Rill Cloud, these same models and dashboards can then be powered by your production source, reflecting the most correct and up-to-date version of your business data.

## Default dev and prod environments

Rill comes with a default `dev` and `prod` property defined, corresponding to Rill Developer and Rill Cloud respectively. Any environment specific YAML overrides or custom templated SQL logic can reference these environments without any additional configuration needed.

:::tip Shortcut to specify dev and prod in YAML files

For the built-in `dev` and `prod` environments **specifically**, Rill provides a shorthand where you can specify properties for these environments directly under `dev` / `prod` without first nesting it under a parent `env` key. For example, if you had the following `rill.yaml`:
```yaml
env:
  dev:
    path: s3://path/to/bucket/Y=2024/M=01/*.parquet
  prod:
    refresh:
      cron: 0 * * * *
    models:
      materialize: true
```

This would be exactly equivalent to (within the same `rill.yaml`):
```yaml
dev:
  path: s3://path/to/bucket/Y=2024/M=01/*.parquet
prod:
  refresh:
    cron: 0 * * * *
  models:
    materialize: true
```

For other custom environments that you are defining manually, you will still need to pass them in using the standard environment YAML syntax:
```yaml
env:
  custom_env:
    property: value
```
:::

### Specifying a custom environment

When using Rill Developer, you can specify a custom environment for your local instance (instead of defaulting to `dev`) by using the following command:

```bash
rill start --env <name_of_environment>
```

## Specifying environment specific YAML overrides

Environment overrides can be applied to source properties in the [YAML configuration](/reference/project-files/sources.md) of a source. For example, let's say that you have a [S3](/reference/connectors/s3.md) source defined but you only wanted to read from a particular month partition during local development and make sure that [source refreshes](/build/connect/source-refresh.md) are only applied _in production_ (i.e. when a project is deployed on Rill Cloud). Then, in your `source.yaml` file, you can define it as:

```yaml
connector: s3
path: s3://path/to/bucket/*.parquet
env:
  dev:
    path: s3://path/to/bucket/Y=2024/M=01/*.parquet
  prod:
    refresh:
      cron: 0 * * * *
```

Similarly, if you wanted to set a project-wide default in `rill.yaml` where models are [materialized](/reference/project-files/models.md#model-materialization) only on Rill Cloud (i.e. `prod) and dashboards use a different default [theme](../dashboards/customize.md#changing-themes--colors) in production compared to locally, you could do this by:

```yaml
env:
  prod:
    models:
      materialize: true
    dashboards:
      theme: <name_of_theme>
```

:::info Hierarchy of inheritance and property overrides

As a general rule of thumb, properties that have been specified at a more _granular_ level will supercede or override higher level properties that have been inherited. Therefore, in order of inheritance, Rill will prioritize properties in the following order:
1. Individual [source](/reference/project-files/sources.md)/[model](/reference/project-files/models.md)/[dashboard](/reference/project-files/dashboards.md) object level properties (e.g. `source.yaml` or `dashboard.yaml`)
2. [Environment](/docs/build/models/environments.md) level properties (e.g. a specific property that have been set for `dev`)
3. [Project-wide defaults](/reference/project-files/rill-yaml.md#project-wide-defaults) for a specific property and resource type

:::

## Using environments to generate custom templated SQL

Environments are also useful when you wish to apply environment-specific SQL logic to your sources and models. One common use case would to apply a filter or limit for models automatically when developing locally (in Rill Developer) but not having these same conditions applied to production models deployed on Rill Cloud. These same principles could also be extended to apply more advanced logic and conditional statements based on your requirements. This is all possible by combining environments with Rill's ability to leverage [templating](/deploy/templating.md).

Similar to the example in the previous section, let's say we had a S3 source defined but this time we did not have a partitioned bucket. However, it contains an `updated_at` timestamp column that allows us to leverage DuckDB's ability to read from the S3 file directly and then apply a filter post-download (but we only want to do this locally). In production, we still want to make sure that our models and dashboards are using the full data present in the S3 source.

Now, for your `source.yaml` file (and combined with templating), you could do something like:

```yaml
connector: "duckdb"
sql: SELECT * FROM read_parquet('s3://path/to/bucket/*.parquet') {{ if dev }} where updated_at >= '2024-01-01' AND updated_at < '2024-02-01' {{ end }}
```

On the other hand, let's say we had some kind of intermediate model where we wanted to apply a limit of 10000 _(but only for local development)_, our `model.sql` file may look something like the following instead:

```sql
SELECT * FROM {{ ref "<source_name>" }}
{{ if dev }} LIMIT 10000 {{ end }}
```

:::warning When applying templated logic to model SQL, make sure to leverage the `ref` function

If you use templating in SQL models, you must replace references to tables / models created by other sources or models with `ref` tags. See this section on ["Referencing other tables or models in SQL when using templating"](../../deploy/templating.md#referencing-other-tables-or-models-in-sql-when-using-templating). This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::