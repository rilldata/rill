---
title: Introduction to Models
description: Create models from source data and apply SQL transformations
sidebar_label: Models 101
sidebar_position: 00
---
 
Models in Rill enable data transformation, preparation, and enrichment through SQL queries and YAML configurations. They serve as the foundation for creating clean, structured datasets that power your metrics views and dashboards. 

Data models are built using SQL SELECT statements applied to your source data. They allow you to join, transform, and clean data.
### SQL Transformations

By default, data transformations in Rill Developer are powered by DuckDB and its dialect of SQL (DuckDB SQL). Please visit the [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction) to learn how to write your queries.

You can change the default [OLAP engine](https://docs.rilldata.com/connect/olap) for [the entire project](https://docs.rilldata.com/reference/project-files/rill-yaml#configuring-the-default-olap-engine) or [a specific metrics view](https://docs.rilldata.com/reference/project-files/metrics-views). You will need to define the connector credentials within your Rill project or via environment variables.

:::tip Supported OLAP engines for modeling

We support modeling on [ClickHouse\*](/connect/olap/clickhouse), [DuckDB](/connect/olap/duckdb), and [MotherDuck\*](/connect/olap/motherduck). For more information, see each OLAP engine page for further details.

\* indicates some caveats with modeling, and we encourage you to read the documentation before getting started.

:::

For additional tips on advanced expressions (either in models or measure definitions), visit our [advanced expressions page](/build/metrics-view/advanced-expressions/advanced-expressions.md).

### Intermediate Processing

Models can also be cross-referenced with each other to produce the final output for your dashboard. This approach enables more complex, intermediate data transformations to achieve your final data source. Common modeling patterns include:

- Lookups for id/name joins
- Unnesting and merging complex data types
- Combining multiple sources with data cleansing or transformation requirements



## Data Preview and Validation

### Table Preview

Rill automatically generates a preview of your data (first 150 rows) to help verify that the output table structure is correct and identify any potential issues that need to be addressed in the SQL configuration, such as data type detection problems.

### Schema Details

The right panel displays comprehensive information about your dataset and column contents:

- **Dataset Overview**: Total row and column counts
- **Data Quality Metrics**: Number of dropped rows and columns
- **Column Analysis**: 
  - Column names and data types
  - Distinct value counts for string columns
  - Basic numeric statistics (minimum, maximum, median, etc.)

This information helps you validate your model configuration and ensure data quality before proceeding with the full data ingestion.

<img src='/img/build/model/preview.png' class='rounded-gif' />
<br />

## One Big Table and Dashboarding

The power of this approach lies in translating many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level company insights (how much revenue did we make last week?) become more actionable for employees when contextualized to their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](/build/dashboards) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

:::tip Materializing metrics-powered models

We recommend materializing the model that powers your [metrics view](/build/metrics-view). You can materialize a SQL model by adding this to the top of the file. This will greatly improve the performance of your dashboards.

```sql
-- @materialize: true
```
:::
## Other Considerations to Note

### Working with Pivots

Pivots deserve their own section, as using the [Pivot](https://duckdb.org/docs/sql/statements/pivot) statement while modeling requires special consideration. Notably, there are a few existing DuckDB limitations to consider:
- DuckDB's [SQL to JSON serializer](https://duckdb.org/docs/extensions/json.html#serializing-and-deserializing-sql-to-json-and-vice-versa) doesn't support `PIVOT` without the `IN` [filter](https://duckdb.org/docs/sql/statements/pivot#in-filter-for-on-clause)
- DuckDB doesn't support creating views based on `PIVOT` without an `IN` filter (and all models are materialized as views by default in Rill)

Fortunately, there are workarounds available to address these limitations.

#### Passing the `IN` Filter with Your `PIVOT` Statement

If you know the exact values you want to pivot on, you can specify them using an `IN` filter with your `ON` clause. For example, instead of:

```sql
PIVOT table_name ON column_name USING SUM(measure)
```

You can use the following `PIVOT` statement:

```sql
PIVOT table_name ON column_name IN (value_a, value_b, value_c) USING SUM(measure)
```

### DuckDB Model's pre-exec, post-exec SQL

While we install a set of core libraries and extensions with our embedded DuckDB, you may need additional extensions for specific use cases. Use the `pre_exec` parameter to ensure all required components are loaded before running your SQL query.

Consider the [`gsheets` community extension](https://duckdb.org/community_extensions/extensions/gsheets.html) as an example. To use this extension in Rill, you'll need to install and load the plugin, then define the secret before running your SQL.

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

Another example involves attaching a database to DuckDB, running queries against it, then detaching the database.

```yaml
pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
post_exec: DETACH DATABASE IF EXISTS postgres_db # Note: this is not mandatory but nice to have
```

The `pre_exec` and `post_exec` statements run with every model execution and should be made idempotent. A typical approach is to use `IF NOT EXISTS` clauses for CREATE statements. Refer to the DuckDB documentation for exact definitions and verify statement idempotency.



