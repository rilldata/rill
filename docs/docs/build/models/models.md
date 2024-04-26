---
title: Create Models
description: Create models from source data and apply SQL transformations
sidebar_label: Create Models
sidebar_position: 00
---

[Data models](/reference/project-files/models.md) in Rill are composed of SQL `SELECT` statements that operate on source data. They allow you to join, transform, and clean data.

## SQL transformations

Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL (DuckDB SQL). Please visit [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction) to learn about how to write your queries.

For additional tips on commonly used expressions (either in models or dashboard definitions) visit our [common expressions page](../dashboards/expressions.md).


## Adding a data model

### Using the UI
To add a new data model using the UI, click "+" by Models in the left hand navigation pane. You can now begin typing a DuckDB SQL `SELECT` query for your model in the code editor – with keystroke-by-keystroke feedback.

### Using code
When you add a data model using the UI, a code definition will automatically be created as a `<model_name>.sql` file in the `models` folder in your Rill project.

You can also create a model outside of the application and add it to Rill by placing a `<model_name>.sql` file in the `models` directory containing a DuckDB SQL `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.

:::tip

See also our [Model YAML](../../reference/project-files/models) reference page.

:::

## How to use data models

### One Big Table and dashboarding

It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if it is relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](/build/dashboards) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

### Intermediate processing

Models can also be cross-referenced between each other to produce the final output for your dashboard. The advantage here is more complex, intermediate data transformations can be utilized to achieve your final source for the dashboard. Example ideas for modeling:

- Lookups for id/name joins
- Unnesting and merging complex data types
- Combining multiple sources with data cleansing or transformation requirements

## Working with Pivots

Pivots deserve their own section as using the [Pivot](https://duckdb.org/docs/sql/statements/pivot) statement while modeling deserves its own considerations. Notably, there are a few existing DuckDB limitations to consider:
- DuckDB's [SQL to JSON serializer](https://duckdb.org/docs/extensions/json.html#serializing-and-deserializing-sql-to-json-and-vice-versa) doesn't support `PIVOT` without the `IN` [filter](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)
- DuckDB doesn't support creating views based on `PIVOT` without an `IN` filter (and all models are materialized as views by default in Rill)

Fortunately, there are a few workarounds that we can leverage to circumvent these limitations.

### Passing the `IN` filter with your `PIVOT` statement

If you know the _exact values_ that you are trying to pivot on, you can simply pass in these values as part of your pivot query by using an `IN` filter with your `ON` clause ([link to DuckDB documentation](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)). For example, rather than:

```sql
PIVOT table_name ON column_name USING SUM(measure)
```

You can use the following `PIVOT` statement:

```sql
PIVOT table_name ON column_name IN (value_a, value_b, value_c) USING SUM(measure)
```

### Materializing the model as a table and then utilizing the `ref` function

If you <u>need</u> to use the `PIVOT` statement specifically but don't want to specify an `IN` filter, then you will need to inform Rill to materialize this model as a table **and** to leverage the `ref` function as well (for proper DAG resolution). Using the same example, this would instead look something like:

```sql
-- @materialize: true

PIVOT {{ ref "table_name" }} ON column_name USING SUM(measure)
```

:::info A note on model materialization

The `-- @materialize: true` at the top of your model SQL file informs Rill to materialize the model as a table in the underlying OLAP engine (instead of the default view). More details about materializing models can be found in our [model reference page](/reference/project-files/models#model-materialization).

:::