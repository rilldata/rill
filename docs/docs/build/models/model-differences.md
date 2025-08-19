---
title: Creating Models in Rill
description: Create models from source data and apply SQL transformations
sidebar_label: SQL Models vs. YAML Models
sidebar_position: 00
---

In Rill, there are two types of data models:

- [SQL models](/build/models/model-differences#sql-models)
- [YAML models](/build/models/model-differences#yaml-models)

For most use cases, SQL models, _the default_, are sufficient to transform your data to prepare for visualization. SQL models are built using SQL `SELECT` statements applied to your source data. Under the hood, SQL models are created as views in DuckDB and can be [materialized](/build/models/sql-models#sql-model-materialization) as tables when needed.

For more complex modeling and [data ingestion](/build/models/yaml-models/source-models), YAML models are used. By using a YAML approach, we are able to fine-tune the model's settings to enable partitions, incremental modeling, refreshes, and more.

:::tip Avoid Pre-aggregated Metrics

Rill works best for slicing and dicing data, meaning keeping data closer to raw to retain that granularity for flexible analysis. When loading data, be careful with adding pre-aggregated metrics like averages, as that could lead to unintended results like a sum of an average. Instead, load the two raw metrics and calculate the derived metric in your model or dashboard.

:::

## One Big Table and Dashboarding

It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if they are relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](/build/dashboards) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

:::tip Materializing metrics-powered models

We recommend materializing the model that powers your [metrics view](/build/metrics-view). You can materialize a SQL model by adding this to the top of the file. This will greatly improve the performance of your dashboards.

```sql
-- @materialize: true
```
:::

## Model Materialization



Model materialization is something to consider when creating intermediate models. Other than [source models](/build/models/source-models), intermediate models are not, by default, materialized and are views in your underlying database engine. There are some pros and cons to enabling it during the development process.

```sql
-- model.sql
-- @materialize: true
```

```yaml
# model.yaml
materialize: true
```

The pros include improved performance for downstream models and dashboards, especially with complex logic and/or large data sizes. Some cons are certain edge cases like cross joins might have a degraded keystroke-by-keystroke experience, and materialized models are billable.

If you are seeing degraded performance, the first recommendation you'll hear from us is to materialize the metrics-powered model.

### Default Model Materialization

If you want, you can change the default behavior of all models in Rill by setting the default model behavior in the rill.yaml file.

```yaml
models:
  materialize: true
```

To override this on a per-model basis, simply set the specific model.sql to false.

```sql
-- @materialize: false
```

:::info To materialize or not to materialize?

There are both pros and cons to materializing your models.
- Pros can include improved performance for downstream models and dashboards, especially when the SQL is complex and/or the data size is large. We generally recommend _materializing_ final models that power dashboards.
- Cons can include a degraded keystroke-by-keystroke modeling experience or for specific edge cases, such as when using cross joins.

If unsure, we would generally recommend leaving the defaults and/or [reaching out](/contact) for further guidance!

:::

## SQL Models

### When to use SQL Models?

For most users working with DuckDB-backed Rill projects, SQL models provide everything needed to transform and prepare data for visualizations. These models are the default option when using the UI and offer full functionality for data transformation.

### Creating a SQL Model

When using the UI to create a new model, you'll see something similar to the below screenshot. You can also create a model directly from the connector UI in the bottom left by selecting the "...". This will create a `select * from underlying_table` as SQL model file.

<img src = '/img/build/model/model.png' class='rounded-gif' />
<br />


## YAML Models

Unlike SQL models, YAML file models allow you to fine-tune a model to perform additional capabilities such as pre-exec, post-exec SQL, partitioning, and incremental modeling. This is an important addition to modeling, as it allows users to customize the model's build process. In the case of partitions and incremental modeling, this will reduce the amount of data ingested into Rill at each interval and provide insight into specific issues per partition. Another use case is when using [multiple OLAP engines](/connect/olap/multiple-olap), which allows you to define where a SQL query is run.

### When to use YAML Models

For the majority of users on a DuckDB-backed Rill project, YAML models are not required. When a project grows larger and refreshing entire datasets becomes a time-consuming and costly task, we introduce incremental ingestion to help alleviate the problem. Along with incremental modeling, we use partitions to divide a dataset into smaller, more manageable partitions. After enabling partitions, you will be able to refresh individual partitions of data when required.

Another use case is when using multiple OLAP engines. This allows you to specify where your SQL query is running. When both DuckDB and ClickHouse are enabled in a single environment, you will need to define `connector: duckdb/clickhouse` in the YAML to tell Rill where to run the SQL query, as well as define the `output` location. For more information, refer to the [YAML reference](/reference/project-files/models).

### Types of YAML Models

1. [Source Models](/build/models/source-models)
2. [Incremental Models](/build/models/incremental-models)
3. [Partitioned Models](/build/models/partitions)
4. [Staging Models](/build/models/staging)
5. [DuckDB `pre_exec`/`post_exec` Models](/build/models/model-differences#duckdb-models-pre-exec-post-exec-sql)

### Creating a YAML Model

You can get started with an advanced model using the following code block:

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
connector: duckdb

sql: select * from <source>

output:
  connector: duckdb
  table: output_name
```

Please refer to [our reference documentation](/reference/project-files/models) linked above for the available parameters to set in your model.

:::note

Currently, there isn't a UI button to start with an advanced model YAML. Creating a model in Rill via the UI will always create a model.sql file. Instead, start with a blank file, rename it to `model_name.yaml`, and add the above sample code.

:::

### DuckDB Model's pre-exec, post-exec SQL

While we install a set of core libraries and extensions with our embedded DuckDB, there might be specific use cases where you want to add a different one. In order to do this, you will need to use the pre-exec parameter to ensure that everything is loaded before running your SQL query.

Take the example of the [`gsheets` community extension](https://duckdb.org/community_extensions/extensions/gsheets.html). In order to use this extension in Rill, you'll need to install and load the plugin. Once that's done, you can define the secret and finally run the SQL.

```yaml
pre_exec: |
    INSTALL gsheets FROM community;
    LOAD gsheets;
    CREATE TEMPORARY SECRET IF NOT EXISTS secret (TYPE gsheet, PROVIDER access_token, TOKEN '<your_token>');

sql: SELECT * FROM read_gsheet('https://docs.google.com/spreadsheets/d/<your_unique_ID>', headers=false);
```

:::tip Multiple queries to run?

Like any SQL query, you can separate multiple queries with semicolons. This is available for both `pre_exec` and `post_exec`. The default `sql` parameter requires a single SELECT statement to run.

:::

Another example is attaching a database to DuckDB, running some queries against it, then detaching the database.

```yaml
pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
post_exec: DETACH DATABASE IF EXISTS postgres_db # Note: this is not mandatory but nice to have
```

The `pre_exec` and `post_exec` statements are run for every model execution and thus should be made idempotent. A typical approach is to use `IF NOT EXISTS` clauses for CREATE statements. Please refer to the DuckDB documentation for exact definitions and verify if the statements are idempotent.



## Other Considerations to Note


### SQL Transformations

By default, data transformations in Rill Developer are powered by DuckDB and its dialect of SQL (DuckDB SQL). Please visit the [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction) to learn how to write your queries.

It is possible to change the default [OLAP engine](https://docs.rilldata.com/connect/olap) for [the entire project](https://docs.rilldata.com/reference/project-files/rill-yaml#configuring-the-default-olap-engine) or [a specific metrics view](https://docs.rilldata.com/reference/project-files/metrics-views). You will need to define the connector credentials within your Rill project or via environment variables.

:::tip Supported OLAP engines for modeling

We support modeling on [ClickHouse\*](/connect/olap/clickhouse), [DuckDB](/connect/olap/duckdb), and [MotherDuck\*](/connect/olap/motherduck). For more information, see each OLAP engine page for further details.

\* indicates some caveats with modeling, and we encourage you to read the documentation before getting started.

:::

For additional tips on advanced expressions (either in models or measure definitions), visit our [advanced expressions page](/build/metrics-view/advanced-expressions/advanced-expressions.md).

### Intermediate Processing

Models can also be cross-referenced with each other to produce the final output for your dashboard. The advantage here is that more complex, intermediate data transformations can be utilized to achieve your final source for the dashboard. Example ideas for modeling:

- Lookups for id/name joins
- Unnesting and merging complex data types
- Combining multiple sources with data cleansing or transformation requirements


### Working with Pivots

Pivots deserve their own section, as using the [Pivot](https://duckdb.org/docs/sql/statements/pivot) statement while modeling requires special consideration. Notably, there are a few existing DuckDB limitations to consider:
- DuckDB's [SQL to JSON serializer](https://duckdb.org/docs/extensions/json.html#serializing-and-deserializing-sql-to-json-and-vice-versa) doesn't support `PIVOT` without the `IN` [filter](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)
- DuckDB doesn't support creating views based on `PIVOT` without an `IN` filter (and all models are materialized as views by default in Rill)

Fortunately, there are a few workarounds that we can leverage to circumvent these limitations.

#### Passing the `IN` Filter with Your `PIVOT` Statement

If you know the _exact values_ that you are trying to pivot on, you can simply pass in these values as part of your pivot query by using an `IN` filter with your `ON` clause ([link to DuckDB documentation](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)). For example, rather than:

```sql
PIVOT table_name ON column_name USING SUM(measure)
```

You can use the following `PIVOT` statement:

```sql
PIVOT table_name ON column_name IN (value_a, value_b, value_c) USING SUM(measure)
```
