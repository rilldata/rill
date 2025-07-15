---
title: Environments
description: Using environments to separate logic in Rill Developer and Cloud
sidebar_label: Environments
sidebar_position: 10
---

## Overview

Environments allow the separation of logic between different deployments of Rill, most commonly when using a project in Rill Developer and Rill Cloud. Generally speaking, Rill Developer is meant primarily for local development purposes, where the developer may want to use a segment or sample of data for modeling purposes to help validate the model logic is correct and producing expected results. Then, once finalized and a project is ready to be deployed to Rill Cloud, the focus is on shared collaboration at scale and where most of the dashboard consumption in production will happen.

There could be a few reasons to have separate logic for sources, models, and dashboards for Rill Developer and Rill Cloud, respectively:
1. As the size of data grows, the locally embedded OLAP engine (using DuckDB) may start to face scaling challenges, which can impact performance and the "snappiness" of models and dashboards. Furthermore, for model development, the full data is often not needed and working with a sample is sufficient. For production, though, where analysts and business users are interacting with Rill dashboards to perform interactive, exploratory analysis or make decisions, it is important that these same models and dashboards are powered by the entire range of data available.
2. For certain organizations, there might be a development and production version of source data. Therefore, you can develop your models and validate the results in Rill Developer against development data. When deployed to Rill Cloud, these same models and dashboards can then be powered by your production source, reflecting the most correct and up-to-date version of your business data.

## Default dev and prod environments

Rill comes with default `dev` and `prod` properties defined, corresponding to Rill Developer and Rill Cloud, respectively. You can use these keys to set environment-specific YAML overrides or SQL logic.

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

Environment overrides can be applied to source properties in the [YAML configuration](/reference/project-files/sources.md) of a source. For example, let's say that you have a [S3](/reference/connectors/s3.md) source defined but you only wanted to read from a particular month partition during local development. Then, in your `source.yaml` file, you can define it as:

```yaml
type: source
connector: s3
path: s3://path/to/bucket/*.parquet
dev:
  path: s3://path/to/bucket/Y=2024/M=01/*.parquet
```

Similarly, if you wanted to set a project-wide default in `rill.yaml` where models are [materialized](/reference/project-files/models.md#model-materialization) only on Rill Cloud (i.e., `prod`) and dashboards use a different default [theme](../dashboards/customize.md#changing-themes--colors) in production compared to locally, you could do this by:

```yaml
prod:
  models:
    materialize: true
  explores:
    theme: <name_of_theme>
```

:::info Hierarchy of inheritance and property overrides

As a general rule of thumb, properties that have been specified at a more _granular_ level will supersede or override higher-level properties that have been inherited. Therefore, in order of inheritance, Rill will prioritize properties in the following order:
1. Individual [source](/reference/project-files/sources.md)/[model](/reference/project-files/models.md)/[dashboard](/reference/project-files/explore-dashboards.md) object level properties (e.g. `source.yaml` or `dashboard.yaml`)
2. [Environment](/docs/build/models/environments.md) level properties (e.g., a specific property that has been set for `dev`)
3. [Project-wide defaults](/reference/project-files/rill-yaml.md#project-wide-defaults) for a specific property and resource type

:::

## Running scheduled source refreshes in development

As an exception, scheduled source refreshes specified using `refresh:` are not applied in the `dev` environment by default. If you want to run or test scheduled refreshes in local development, you can override this behavior using the `run_in_dev` property:
```yaml
refresh:
  cron: 0 * * * *
  run_in_dev: true
```

:::tip Why are source refreshes only enabled by default for Rill Cloud?

Source refreshes are primarily meant to _help keep the data in your deployed dashboards on Rill Cloud up-to-date_ (without needing to manually trigger refreshes). For more details, see our documentation on [configuring source refreshes](/build/connect/source-refresh.md).

:::

## Using environments to generate custom templated SQL

Environments are also useful when you wish to apply environment-specific SQL logic to your sources and models. One common use case would be to apply a filter or limit for models automatically when developing locally (in Rill Developer), but not have these same conditions applied to production models deployed on Rill Cloud. These same principles could also be extended to apply more advanced logic and conditional statements based on your requirements. This is all possible by combining environments with Rill's ability to leverage [templating](/deploy/templating.md).

Similar to the example in the previous section, let's say we had a S3 source defined but this time we did not have a partitioned bucket. However, it contains an `updated_at` timestamp column that allows us to leverage DuckDB's ability to read from the S3 file directly and then apply a filter post-download (but we only want to do this locally). In production, we still want to make sure that our models and dashboards are using the full data present in the S3 source.

Now, for your `source.yaml` file (and combined with templating), you could do something like:

```yaml
type: source
connector: "duckdb"
sql: SELECT * FROM read_parquet('s3://path/to/bucket/*.parquet') {{ if dev }} where updated_at >= '2024-01-01' AND updated_at < '2024-02-01' {{ end }}
```

On the other hand, let's say we had some kind of intermediate model where we wanted to apply a limit of 10,000 _(but only for local development)_, our `model.sql` file may look something like the following instead:

```sql
SELECT * FROM {{ ref "<source_name>" }}
{{ if dev }} LIMIT 10000 {{ end }}
```

:::warning When applying templated logic to model SQL, make sure to leverage the `ref` function

If you use templating in SQL models, you must replace references to tables/models created by other sources or models with `ref` tags. See this section on ["Referencing other tables or models in SQL when using templating"](../../deploy/templating.md#referencing-other-tables-or-models-in-sql-when-using-templating). This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::