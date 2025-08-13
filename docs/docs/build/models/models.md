---
title: Create Data Models
description: Create models from source data and apply SQL transformations
sidebar_label: Create SQL or YAML Models in Rill
sidebar_position: 00
---

In Rill, there are two types of data models. 

- [SQL models](/build/models/sql-models)
- [YAML models](/build/models/yaml-models)


For most use-cases, SQL models, _the default_, are enough to transform your data to prepare for visualization. The SQL models are built using SQL `SELECT` statements applied to your source data.  Under the hood, SQL models are created as views in DuckDB, and can be [materialized](/build/models/sql-models#sql-model-materialization) as tables when needed.

For more complex modeling and [data ingestion](/build/models/yaml-models/source-models), YAML models are used. By using a YAML approach, we are able to fine-tune the model's settings to enable partitions, incremental modeling, refreshes and more.


:::tip Avoid Pre-aggregated Metrics

Rill works best for slicing and dicing data meaning keeping data closer to raw to retain that granularity for flexible analysis. When loading data, be careful with adding pre-aggregated metrics like averages as that could lead to unintended results like a sum of an average. Instead, load the two raw metrics and calculate the derived metric in your model or dashboard.

:::



## SQL transformations

By default, data transformations in Rill Developer are powered by DuckDB and its dialect of SQL (DuckDB SQL). Please visit the [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction) to learn how to write your queries.

It is possible to change the default [OLAP engine](https://docs.rilldata.com/connect/olap) for [the entire project](https://docs.rilldata.com/reference/project-files/rill-yaml#configuring-the-default-olap-engine) or [a specific metrics view](https://docs.rilldata.com/reference/project-files/metrics-views). You will need to define the connector credentials within your Rill project or via environment variables.

:::tip Support OLAP engines for modeling
We support modeling on [ClickHouse\*](/connect/olap/clickhouse), [DuckDB](/connect/olap/duckdb) and [MotherDuck\*](/connect/olap/motherduck). For more information, see each OLAP engine page for further information.

\* indicates some caveats with modeling and encourage you to read the documentation before getting started.
::: 

For additional tips on advanced expressions (either in models or measureß definitions), visit our [advanced expressions page](/build/metrics-view/advanced-expressions/advanced-expressions.md).



## How to use models

### Intermediate processing

Models can also be cross-referenced with each other to produce the final output for your dashboard. The advantage here is that more complex, intermediate data transformations can be utilized to achieve your final source for the dashboard. Example ideas for modeling:

- Lookups for id/name joins
- Unnesting and merging complex data types
- Combining multiple sources with data cleansing or transformation requirements


### Working with Pivots

Pivots deserve their own section, as using the [Pivot](https://duckdb.org/docs/sql/statements/pivot) statement while modeling requires special consideration. Notably, there are a few existing DuckDB limitations to consider:
- DuckDB's [SQL to JSON serializer](https://duckdb.org/docs/extensions/json.html#serializing-and-deserializing-sql-to-json-and-vice-versa) doesn't support `PIVOT` without the `IN` [filter](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)
- DuckDB doesn't support creating views based on `PIVOT` without an `IN` filter (and all models are materialized as views by default in Rill)

Fortunately, there are a few workarounds that we can leverage to circumvent these limitations.

#### Passing the `IN` filter with your `PIVOT` statement

If you know the _exact values_ that you are trying to pivot on, you can simply pass in these values as part of your pivot query by using an `IN` filter with your `ON` clause ([link to DuckDB documentation](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)). For example, rather than:

```sql
PIVOT table_name ON column_name USING SUM(measure)
```

You can use the following `PIVOT` statement:

```sql
PIVOT table_name ON column_name IN (value_a, value_b, value_c) USING SUM(measure)
```
## One Big Table and dashboarding

It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if it is relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](/build/dashboards) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

:::tip materializing metrics powered models

We recommend materializing the model that powers your [metrics view](/build/metrics-view). You can materialze a SQL model by adding this to the top of the file. This will greatly improve the performance of your dashboards.

```sql
-- @materialize: true
```

