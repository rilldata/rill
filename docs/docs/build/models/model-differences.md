---
title: When to use SQL vs YAML
description: Create models from source data and apply SQL transformations
sidebar_label: When to use SQL vs YAML
sidebar_position: 03
---

In Rill, there are two types of data models:

- [SQL models](/build/models/model-differences#sql-models)
- [YAML models](/build/models/model-differences#yaml-models)

For most use cases, SQL models, _the default_, are sufficient to transform your data to prepare for visualization. SQL models are built using SQL `SELECT` statements applied to your source data. Under the hood, SQL models are created as views in DuckDB and can be [materialized](/build/models/performance#materialization) as tables when needed.

For more complex modeling and [data ingestion](/build/models/source-models), YAML models are used. By using a YAML approach, we are able to fine-tune the model's settings to enable partitions, incremental modeling, refreshes, and more.

:::tip Avoid Pre-aggregated Metrics

Rill works best for slicing and dicing data, meaning keeping data closer to raw to retain that granularity for flexible analysis. When loading data, be careful with adding pre-aggregated metrics like averages, as that could lead to unintended results like a sum of an average. Instead, load the two raw metrics and calculate the derived metric in your model or dashboard.

:::

## SQL Models

### When to use SQL Models?

For most users working with DuckDB-backed Rill projects, SQL models provide everything needed to transform and prepare data for visualizations. These models are the default option when using the UI and offer full functionality for data transformation.

### Creating a SQL Model

When using the UI to create a new model, you'll see something similar to the below screenshot. You can also create a model directly from the connector UI in the bottom left by selecting the "...". This will create a `select * from underlying_table` as SQL model file.

<img src='/img/build/model/model.png' class='rounded-gif' />
<br />

## YAML Models

Unlike SQL models, YAML file models allow you to fine-tune a model to perform additional capabilities such as pre-exec, post-exec SQL, partitioning, and incremental modeling. This is an important addition to modeling, as it allows users to customize the model's build process. In the case of partitions and incremental modeling, this will reduce the amount of data ingested into Rill at each interval and provide insight into specific issues per partition. Another use case is when using [multiple OLAP engines](/build/connectors/olap/multiple-olap), which allows you to define where a SQL query is run.

### When to use YAML Models

For the majority of users on a DuckDB-backed Rill project, YAML models are not required. When a project grows larger and refreshing entire datasets becomes a time-consuming and costly task, we introduce incremental ingestion to help alleviate the problem. Along with incremental modeling, we use partitions to divide a dataset into smaller, more manageable partitions. After enabling partitions, you will be able to refresh individual partitions of data when required.

Another use case is when using multiple OLAP engines. This allows you to specify where your SQL query is running. When both DuckDB and ClickHouse are enabled in a single environment, you will need to define `connector: duckdb/clickhouse` in the YAML to tell Rill where to run the SQL query, as well as define the `output` location. For more information, refer to the [YAML reference](/reference/project-files/models).

### Types of YAML Models

1. [Source Models](/build/models/source-models)
2. [Incremental Models](/build/models/incremental-models)
3. [Partitioned Models](/build/models/partitioned-models)
4. [Incremental + Partitioned Models](/build/models/incremental-partitioned-models)
5. [Staging Models](/build/models/staging-models)

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

