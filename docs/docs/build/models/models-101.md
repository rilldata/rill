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

You can change the default [OLAP engine](https://docs.rilldata.com/build/connectors/olap) for [the entire project](https://docs.rilldata.com/reference/project-files/rill-yaml#configuring-the-default-olap-engine) or [a specific metrics view](https://docs.rilldata.com/reference/project-files/metrics-views). You will need to define the connector credentials within your Rill project or via environment variables.

:::tip Supported OLAP engines for modeling

We support modeling on [ClickHouse\*](/build/connectors/olap/clickhouse), [DuckDB](/build/connectors/olap/duckdb), and [MotherDuck\*](/build/connectors/olap/motherduck). For more information, see each OLAP engine page for further details.

\* indicates some caveats with modeling, and we encourage you to read the documentation before getting started.

:::

For additional tips on advanced expressions (either in models or measure definitions), visit our [advanced expressions page](/build/metrics-view).

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

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable [metrics definition](/build/metrics-view) to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

:::tip Materializing metrics-powered models

We recommend materializing the model that powers your [metrics view](/build/metrics-view). You can materialize a SQL model by adding this to the top of the file. This will greatly improve the performance of your dashboards.

```sql
-- @materialize: true
```
:::
